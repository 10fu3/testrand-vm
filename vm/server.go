package vm

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"testrand-vm/compile"
	"testrand-vm/config"
	"testrand-vm/infra"
	"testrand-vm/util"
	"testrand-vm/vm/iface"
	"time"
)

func LoadBalancingRegisterForServer(conf *iface.LoadBalancingRegisterConfig) error {

	if conf.Self.Host == conf.LoadBalancer.Host {
		conf.LoadBalancer.Host = "localhost"
	}

	jsonContent := map[string]string{
		"machine_type": "heavy",
		"from":         fmt.Sprintf("http://%s:%s", conf.Self.Host, conf.Self.Port),
	}
	jsonByte, err := json.Marshal(jsonContent)
	if err != nil {
		return err
	}
	sendBodyBuff := bytes.NewBuffer(jsonByte)
	url := fmt.Sprintf("http://%s:%s/register-heavy", conf.LoadBalancer.Host, conf.LoadBalancer.Port)
	if conf.LoadBalancer.Host == "localhost" {
		url = fmt.Sprintf("http://127.0.0.1:%s/register-heavy", conf.LoadBalancer.Port)
	}
	post, err := http.Post(url, "application/json", sendBodyBuff)
	if err != nil {
		return err
	}
	fmt.Printf("regist result: %d\n", post.StatusCode)
	return nil
}

var runningVmCount = atomic.Int64{}

func StartServer(config config.Value) {

	ramdomListener, _close := util.CreateListener()
	randomPort := fmt.Sprintf("%d", ramdomListener.Addr().(*net.TCPAddr).Port)
	_close()

	engine := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	engine.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(struct {
			Message string `json:"message"`
		}{Message: "OK"})
	})
	engine.Get("/routine-count", func(c *fiber.Ctx) error {
		fmt.Printf("health check: %d\n", runningVmCount.Load())
		return c.JSON(struct {
			Count int `json:"count"`
		}{Count: runtime.NumGoroutine()})
	})
	engine.Get("/health", func(c *fiber.Ctx) error {
		fmt.Println("health check")
		return c.JSON(struct {
			Status string `json:"status"`
		}{Status: "OK"})
	})
	var requestCount uint64
	engine.Post("/add-task/:id", func(c *fiber.Ctx) error {
		requestId := c.Params("id")
		var req TaskAddRequest
		//fmt.Println(string(c.Body()))
		err := c.BodyParser(&req)
		if err != nil {
			fmt.Println("req readErr: " + err.Error())
			return err
		}
		if requestId == "" {
			return c.JSON(fiber.Map{
				"status":  "ng",
				"message": "not allowed empty id",
			})
		}
		if req.From == nil {
			return c.JSON(fiber.Map{
				"status":  "ng",
				"message": "not allowed empty port",
			})
		}
		if req.Body == nil {
			return c.JSON(fiber.Map{
				"status":  "ng",
				"message": "not allowed empty body",
			})
		}
		if req.GlobalNamespaceId == nil {
			return c.JSON(fiber.Map{
				"status":  "ng",
				"message": "not allowed empty session_id",
			})
		}
		runningVmCount.Add(1)
		go func() {
			fmt.Println("other thread start ", uuid.NewString())
			client, err := infra.SetupEtcd(*req.GlobalNamespaceId)

			if err != nil {
				fmt.Println("etcd setup err: " + err.Error())
				return
			}

			defer runningVmCount.Add(-1)
			if err != nil {
				fmt.Println("req readErr: " + err.Error())
				return
			}
			compileEnv := compile.NewCompileEnvironmentBySharedEnvId(*req.GlobalNamespaceId, client)

			//load file
			file, err := os.Open("./lib-lisp/lib.t-lisp")
			if err != nil {
				panic(err)
			}
			defer file.Close()
			r := compile.NewReader(compileEnv, bufio.NewReader(file))
			libSexp, err := r.Read()
			if err != nil {
				panic(err)
			}
			if libCompileErr := compileEnv.Compile(libSexp); libCompileErr != nil {
				fmt.Println(libCompileErr)
				os.Exit(1)
			}

			vm := NewVM(compileEnv)
			VMRunFromEntryPoint(vm)

			input := strings.NewReader(fmt.Sprintf("%s\n", *req.Body))
			read := compile.NewReader(compileEnv, bufio.NewReader(input))
			readSexp, readErr := read.Read()
			if readErr != nil {
				fmt.Println("read readErr: " + readErr.Error())
				return
			}

			if compileEnv.Compile(readSexp) != nil {
				fmt.Println("compile readErr: " + readErr.Error())
				return
			}

			VMRunFromEntryPoint(vm)

			if vm.ResultErr != nil {
				fmt.Println(vm.ResultErr)
				fmt.Println("completed 1")
				return
			}

			sendBody := struct {
				Result string `json:"result"`
			}{
				Result: vm.Result.String(compileEnv),
			}
			sendBodyBytes, readErr := json.Marshal(&sendBody)
			sendBodyBuff := bytes.NewBuffer(sendBodyBytes)
			sendAddr := fmt.Sprintf("http://%s/receive/%s", *req.From, requestId)

			fmt.Println("sendAddr:", sendAddr)

			_, readErr = http.Post(sendAddr, "application/json", sendBodyBuff)

			if readErr != nil {
				fmt.Println(readErr)
			}
			for i := 0; i < 5; i++ {
				if readErr == nil {
					break
				}
				time.Sleep(time.Second * 3)
				_, readErr = http.Post(sendAddr, "application/json", sendBodyBuff)
			}
			vm = nil
			compileEnv = nil
			atomic.AddUint64(&requestCount, 1)
			fmt.Printf("completed %d\n", requestCount)
		}()
		return c.JSON(fiber.Map{
			"status": "ok",
			"id":     requestId,
		})
	})

	go func() {
		if err := engine.Listen(fmt.Sprintf(":%s", randomPort)); err != nil {
			panic(err)
		}
	}()

	ip, err := util.GetLocalIP()
	if err != nil {
		panic(err)
	}
	err = LoadBalancingRegisterForServer(&iface.LoadBalancingRegisterConfig{
		Self: iface.LoadBalancingRegisterSelfConfig{
			Host: ip,
			Port: randomPort,
		},
		LoadBalancer: iface.LoadBalancingRegisterBalancerConfig{
			Host: config.ProxyHost,
			Port: config.ProxyPort,
		},
	})
}
