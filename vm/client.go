package vm

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"testrand-vm/compile"
	"testrand-vm/config"
	"testrand-vm/util"
	"testrand-vm/vm/iface"
)

type TaskAddRequest struct {
	Body              *string `json:"sexp_body"`
	From              *string `json:"from"`
	GlobalNamespaceId *string `json:"global_namespace_id"`
}

func NewSupervisor() *Supervisor {
	return &Supervisor{}
}

type GroupTaskId string
type TaskId string

type Supervisor struct {
	CompileEnv  *compile.CompilerEnvironment
	GlobalEnvId string
	SelfNetwork struct {
		Host string
		Port string
	}
	Config       config.Value
	Mutex        *sync.RWMutex
	Tasks        map[TaskId]*Closure
	RefGroupTask map[TaskId]GroupTaskId
	GroupTask    map[GroupTaskId]*struct {
		Count              uint
		Complete           uint
		Tasks              []TaskId
		CompletedRawResult [][]byte
		OnComplete         *Closure
	}
}

func (s *Supervisor) StartCallbackReceiveServer(conf config.Value) {
	router := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	router.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})
	router.Post("/receive/:id", func(c *fiber.Ctx) error {
		var req struct {
			Result string `json:"result"`
		}
		reqId := c.Params("id")
		parseErr := c.BodyParser(&req)
		if parseErr != nil {
			fmt.Println(parseErr.Error())

			return parseErr
		}

		closure, hasCallback := s.CompleteTask(reqId)

		if !hasCallback {
			return nil
		}

		sample := strings.NewReader(fmt.Sprintf("%s\n", req.Result))
		read := compile.NewReader(s.CompileEnv, bufio.NewReader(sample))
		result, parseErr := read.Read()

		vm := NewVM(s.CompileEnv)
		vm.Stack.Push(result)
		vm.Stack.Push(closure)
		vm.Code = []compile.Instr{
			compile.CreateCallInstr(1),
			compile.CreateEndCodeInstr(),
		}

		VMRun(vm)

		c.Status(http.StatusOK)
		return nil
	})
	if err := router.Listen(fmt.Sprintf(":%s", conf.SelfOnCompletePort)); err != nil {
		panic(err)
	}
}

func (s *Supervisor) sendSingleSexpToServer(taskId TaskId, sendTask compile.SExpression) {
	conf := config.Get()
	reqAddr := fmt.Sprintf("%s:%s", s.SelfNetwork.Host, s.SelfNetwork.Port)
	b := sendTask.String(s.CompileEnv)
	values, err := json.Marshal(TaskAddRequest{
		From:              &reqAddr,
		GlobalNamespaceId: &s.GlobalEnvId,
		Body:              &b,
	})

	if err != nil {
		panic(err)
	}

	transport := http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := &net.Dialer{}
			return dialer.DialContext(ctx, "tcp4", addr)
		},
	}

	client := http.Client{
		Transport: &transport,
	}

	sendReqBody := map[string]string{
		"sexp_body":           b,
		"from":                fmt.Sprintf("%s:%s", s.SelfNetwork.Host, s.SelfNetwork.Port),
		"global_namespace_id": s.GlobalEnvId,
	}
	sendReqBodyByte, _ := json.Marshal(sendReqBody)
	fmt.Println(string(sendReqBodyByte))
	send, err := http.Post(fmt.Sprintf("http://%s:%s/send-request", conf.ProxyHost, conf.ProxyPort), "application/json", bytes.NewBuffer(sendReqBodyByte))
	if err != nil {
		log.Fatal(err)
		return
	}
	sendTargetResult := struct {
		Addr string `json:"addr"`
	}{}
	sendTargetResultByte, err := io.ReadAll(send.Body)
	if err := json.Unmarshal(sendTargetResultByte, &sendTargetResult); err != nil {
		log.Fatal(err)
		return
	}

	res, err := client.Post(fmt.Sprintf("%s/add-task/%s", sendTargetResult.Addr, taskId), "application/json", bytes.NewBuffer(values))
	if err != nil {
		log.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(res.Body)
}

/**
@sendTask: 送信するタスク
@onComplete: タスクが完了した時に呼び出す関数
@return: タスクID
*/

func (s *Supervisor) AddTaskWithCallback(sendTask compile.SExpression, onComplete *Closure) TaskId {
	taskId := TaskId(uuid.NewString())
	s.sendSingleSexpToServer(taskId, sendTask)
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.Tasks[taskId] = onComplete
	return taskId
}

/*
*
@sendTask: 送信するタスク
@onComplete: タスクが完了した時に呼び出す関数
@return: タスクID
*/
func (s *Supervisor) AddGroupTasks(sendTask []compile.SExpression, onComplete *Closure) GroupTaskId {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	groupId := GroupTaskId(uuid.NewString())
	sendTaskSize := len(sendTask)
	for i := 0; i < sendTaskSize; i++ {
		taskId := TaskId(uuid.NewString())
		s.RefGroupTask[taskId] = groupId
		s.GroupTask[groupId].Tasks = append(s.GroupTask[groupId].Tasks, taskId)
	}
	s.GroupTask[groupId].OnComplete = onComplete
	s.GroupTask[groupId].Count = uint(sendTaskSize)
	return groupId
}

func (s *Supervisor) AddTask(comp *compile.CompilerEnvironment, sendTask compile.SExpression) TaskId {
	taskId := TaskId(uuid.NewString())
	s.sendSingleSexpToServer(taskId, sendTask)
	return taskId
}

/*
*
@taskId: タスクID
*/
func (s *Supervisor) CompleteTask(taskId string) (*Closure, bool) {
	tId := TaskId(taskId)
	s.Mutex.RLock()
	if groupId, ok := s.RefGroupTask[tId]; ok {
		s.Mutex.RUnlock()
		s.Mutex.Lock()
		s.GroupTask[groupId].Complete++
		if s.GroupTask[groupId].Complete == s.GroupTask[groupId].Count {
			defer delete(s.GroupTask, groupId)
			defer delete(s.RefGroupTask, tId)
			s.Mutex.Unlock()
			return s.GroupTask[groupId].OnComplete, true
		}
		s.Mutex.Unlock()
	} else {
		s.Mutex.RUnlock()
		s.Mutex.Lock()
		closure, ok := s.Tasks[tId]
		if !ok {
			s.Mutex.Unlock()
			return nil, false
		}
		delete(s.Tasks, tId)
		s.Mutex.Unlock()
		return closure, true
	}

	return nil, false
}

var superV *Supervisor

func GetSupervisor() *Supervisor {
	return superV
}

func LoadBalancingRegisterForClient(conf iface.LoadBalancingRegisterConfig) error {

	if conf.Self.Host == conf.LoadBalancer.Host {
		conf.LoadBalancer.Host = "localhost"
	}

	jsonContent := map[string]string{
		"machine_type":        "client",
		"from":                fmt.Sprintf("http://%s:%s", conf.Self.Host, conf.Self.Port),
		"global_namespace_id": conf.Self.EnvId,
	}
	jsonByte, err := json.Marshal(jsonContent)
	if err != nil {
		panic(err)
	}
	sendBodyBuff := bytes.NewBuffer(jsonByte)
	post, err := http.Post(fmt.Sprintf("http://%s:%s/register-client", conf.LoadBalancer.Host, conf.LoadBalancer.Port), "application/json", sendBodyBuff)
	if err != nil {
		return err
	}
	fmt.Printf("regist result: %d\n", post.StatusCode)
	return nil
}

func StartSupervisorForClient(compileEnv *compile.CompilerEnvironment, conf config.Value) *Supervisor {
	if superV != nil {
		return superV
	}
	supervisor := NewSupervisor()
	supervisor.CompileEnv = compileEnv
	supervisor.GlobalEnvId = uuid.NewString()
	supervisor.Config = conf
	ip, err := util.GetLocalIP()

	if err != nil {
		panic(err)
	}

	supervisor.SelfNetwork = struct {
		Host string
		Port string
	}{Host: ip, Port: conf.SelfOnCompletePort}
	supervisor.Mutex = &sync.RWMutex{}
	supervisor.Tasks = make(map[TaskId]*Closure)
	supervisor.RefGroupTask = make(map[TaskId]GroupTaskId)
	supervisor.GroupTask = make(map[GroupTaskId]*struct {
		Count              uint
		Complete           uint
		Tasks              []TaskId
		CompletedRawResult [][]byte
		OnComplete         *Closure
	})
	superV = supervisor
	err = LoadBalancingRegisterForClient(iface.LoadBalancingRegisterConfig{
		Self: iface.LoadBalancingRegisterSelfConfig{
			Host: ip,
			Port: conf.SelfOnCompletePort,
		},
		LoadBalancer: iface.LoadBalancingRegisterBalancerConfig{
			Host: conf.ProxyHost,
			Port: conf.ProxyPort,
		},
	})
	if err != nil {
		panic(err)
		return nil
	}
	go func() {
		supervisor.StartCallbackReceiveServer(conf)
	}()
	return supervisor
}
