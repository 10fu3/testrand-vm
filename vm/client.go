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
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"testrand-vm/compile"
	"testrand-vm/config"
	"testrand-vm/util"
)

type TaskAddRequest struct {
	Body              *string `json:"body"`
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

func (s *Supervisor) StartCallbackReceiveServer() {
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
		vm.Stack.Push(closure)
		vm.Stack.Push(result)
		vm.Code = []compile.Instr{
			compile.CreateCallInstr(1),
		}

		VMRun(vm)

		c.Status(http.StatusOK)
		return nil
	})
}

func (s *Supervisor) sendSingleSexpToServer(taskId TaskId, sendTask compile.SExpression) {
	conf := config.Get()
	reqAddr := fmt.Sprintf("%s:%s", s.SelfNetwork.Host, s.SelfNetwork.Port)
	values, err := json.Marshal(TaskAddRequest{
		From:              &reqAddr,
		GlobalNamespaceId: &s.GlobalEnvId,
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

	b := sendTask.String(s.CompileEnv)

	sendReqBody := map[string]string{
		"body": b,
		"from": fmt.Sprintf("%s:%s", s.SelfNetwork.Host, s.SelfNetwork.Port),
	}
	sendReqBodyByte, err := json.Marshal(sendReqBody)
	send, err := http.Post(fmt.Sprintf("http://%s:%s/send-request", conf.ProxyHost, conf.ProxyPort), "application/json", bytes.NewBuffer(sendReqBodyByte))
	sendTargetResult := struct {
		Addr string `json:"addr"`
	}{}
	sendTargetResultByte, err := ioutil.ReadAll(send.Body)
	if err := json.Unmarshal(sendTargetResultByte, &sendTargetResult); err != nil {
		fmt.Println(err)
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
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
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

//func (s *Supervisor) StartCallbackReceiveServer() {
//	engine := fiber.New(fiber.Config{
//		JSONEncoder: json.Marshal,
//		JSONDecoder: json.Unmarshal,
//	})
//
//	engine.Get("/", func(c *fiber.Ctx) error {
//		return c.JSON(struct {
//			Message string `json:"message"`
//		}{Message: "OK"})
//	})
//	engine.Get("/routine-count", func(c *fiber.Ctx) error {
//		fmt.Printf("health check: %d\n", runtime.NumGoroutine())
//		return c.JSON(struct {
//			Count int `json:"count"`
//		}{Count: runtime.NumGoroutine()})
//	})
//	engine.Get("/health", func(c *fiber.Ctx) error {
//		fmt.Println("health check")
//		return c.JSON(struct {
//			Status string `json:"status"`
//		}{Status: "OK"})
//	})
//	engine.Post("/add-task/:id", func(c *fiber.Ctx) error {
//		requestId := c.Params("id")
//		var req TaskAddRequest
//		err := c.BodyParser(&req)
//		if requestId == "" {
//			return c.JSON(fiber.Map{
//				"status":  "ng",
//				"message": "not allowed empty id",
//			})
//		}
//		if req.From == nil {
//			return c.JSON(fiber.Map{
//				"status":  "ng",
//				"message": "not allowed empty port",
//			})
//		}
//		if req.Body == nil {
//			return c.JSON(fiber.Map{
//				"status":  "ng",
//				"message": "not allowed empty body",
//			})
//		}
//		if req.GlobalNamespaceId == nil {
//			return c.JSON(fiber.Map{
//				"status":  "ng",
//				"message": "not allowed empty session_id",
//			})
//		}
//		go func() {
//			if err != nil {
//				fmt.Println("req err: " + err.Error())
//				return
//			}
//			env, err := NewGlobalEnvironmentById(*req.GlobalNamespaceId)
//
//			if err != nil {
//				panic(err)
//			}
//
//			input := strings.NewReader(fmt.Sprintf("%s\n", *req.Body))
//			read := New(bufio.NewReader(input))
//			readSexp, err := read.Read()
//			if err != nil {
//				fmt.Println("read err: " + err.Error())
//				return
//			}
//			result, err := Eval(ctx, readSexp, env)
//			TopLevelEnvDelete(*req.GlobalNamespaceId)
//			env = nil
//			if err != nil {
//				fmt.Println(err)
//				return
//			}
//			sendBody := struct {
//				Result string `json:"result"`
//			}{
//				Result: result.String(),
//			}
//			sendBodyBytes, err := json.Marshal(&sendBody)
//			sendBodyBuff := bytes.NewBuffer(sendBodyBytes)
//			result = nil
//			sendAddr := fmt.Sprintf("http://%s/receive/%s", *req.From, requestId)
//
//			fmt.Println("sendAddr:", sendAddr)
//
//			_, err = http.Post(sendAddr, "application/json", sendBodyBuff)
//
//			if err != nil {
//				fmt.Println(err)
//			}
//			for i := 0; i < 5; i++ {
//				if err == nil {
//					break
//				}
//				time.Sleep(time.Second * 3)
//				_, err = http.Post(sendAddr, "application/json", sendBodyBuff)
//			}
//		}()
//		return c.JSON(fiber.Map{
//			"status": "ok",
//			"id":     requestId,
//		})
//	})
//	if err := engine.Listen(fmt.Sprintf(":%s", randomPort)); err != nil {
//		panic(err)
//	}
//}

var superV *Supervisor

func GetSupervisor() *Supervisor {
	return superV
}

func StartSupervisorForClient(compileEnv *compile.CompilerEnvironment, config config.Value) *Supervisor {
	if superV != nil {
		return superV
	}
	supervisor := NewSupervisor()
	supervisor.CompileEnv = compileEnv
	supervisor.GlobalEnvId = uuid.NewString()
	supervisor.Config = config
	ip, err := util.GetLocalIP()

	if err != nil {
		panic(err)
	}

	supervisor.SelfNetwork = struct {
		Host string
		Port string
	}{Host: ip, Port: config.SelfOnCompletePort}
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
	return supervisor
}
