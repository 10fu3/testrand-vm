package vm

//
//import (
//	"bufio"
//	"bytes"
//	"fmt"
//	"github.com/goccy/go-json"
//	"github.com/gofiber/fiber/v2"
//	"net/http"
//	"runtime"
//	"strings"
//	"testrand-vm/compile"
//	"time"
//)
//
//func StartServer(comp *compile.CompilerEnvironment) {
//
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
//				fmt.Println("req readErr: " + err.Error())
//				return
//			}
//
//			input := strings.NewReader(fmt.Sprintf("%s\n", *req.Body))
//			read := compile.NewReader(comp, bufio.NewReader(input))
//			readSexp, readErr := read.Read()
//			if readErr != nil {
//				fmt.Println("read readErr: " + readErr.Error())
//				return
//			}
//
//			vm := NewVMWithGlobalEnvId(comp, *req.GlobalNamespaceId)
//
//			TopLevelEnvDelete(*req.GlobalNamespaceId)
//			env = nil
//			if readErr != nil {
//				fmt.Println(readErr)
//				return
//			}
//			sendBody := struct {
//				Result string `json:"result"`
//			}{
//				Result: result.String(),
//			}
//			sendBodyBytes, readErr := json.Marshal(&sendBody)
//			sendBodyBuff := bytes.NewBuffer(sendBodyBytes)
//			result = nil
//			sendAddr := fmt.Sprintf("http://%s/receive/%s", *req.From, requestId)
//
//			fmt.Println("sendAddr:", sendAddr)
//
//			_, readErr = http.Post(sendAddr, "application/json", sendBodyBuff)
//
//			if readErr != nil {
//				fmt.Println(readErr)
//			}
//			for i := 0; i < 5; i++ {
//				if readErr == nil {
//					break
//				}
//				time.Sleep(time.Second * 3)
//				_, readErr = http.Post(sendAddr, "application/json", sendBodyBuff)
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
