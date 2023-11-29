package vm

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	"strconv"
	"strings"
	"sync"
	"testrand-vm/reader"
)

type Env struct {
	Frame map[string]reader.SExpression
}

func (e *Env) TypeId() string {
	return "environment"
}

func (e *Env) SExpressionTypeId() reader.SExpressionType {
	return reader.SExpressionTypeEnvironment
}

func (e *Env) String() string {
	return "environment"
}

func (e *Env) IsList() bool {
	return false
}

func (e *Env) Equals(sexp reader.SExpression) bool {
	panic("implement me")
}

type VM struct {
	Mutex         *sync.RWMutex
	Stack         []reader.SExpression
	Code          []reader.SExpression
	Pc            int64
	Env           *Env
	Cont          *VM
	ContPc        int64
	TemporaryArgs []reader.Symbol
}

func (vm *VM) TypeId() string {
	return "vm"
}

func (vm *VM) SExpressionTypeId() reader.SExpressionType {
	return reader.SExpressionTypeVM
}

func (vm *VM) String() string {
	return "vm"
}

func (vm *VM) IsList() bool {
	return false
}

func (vm *VM) Equals(sexp reader.SExpression) bool {
	//TODO implement me
	panic("implement me")
}

func NewVM() *VM {
	return &VM{
		Stack: make([]reader.SExpression, 0),
		Pc:    0,
		Env:   &Env{Frame: make(map[string]reader.SExpression)},
		Cont:  nil,
		Mutex: &sync.RWMutex{},
	}
}

func VMRun(vm *VM) {

	selfVm := vm

	for {

		rawCode := selfVm.Code[selfVm.Pc].(reader.Symbol).GetValue()
		var opCodeAndArgs = strings.SplitN(rawCode, " ", 2)

		switch opCodeAndArgs[0] {
		case "push-sym":
			selfVm.Push(reader.NewSymbol(opCodeAndArgs[1]))
			selfVm.Pc++

		case "push-num":
			convertedStrInt64, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)

			selfVm.Push(reader.NewInt(convertedStrInt64))
			selfVm.Pc++
		case "push-boo":
			convertedStrBool, _ := strconv.ParseBool(opCodeAndArgs[1])

			selfVm.Push(reader.NewBool(convertedStrBool))
			selfVm.Pc++
		case "push-str":

			selfVm.Push(reader.NewString(opCodeAndArgs[1]))
			selfVm.Pc++
		case "pop":

			selfVm.Pop()
			selfVm.Pc++
		case "jump":
			jumpTo, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)

			selfVm.Pc = jumpTo
		case "load":

			sym := selfVm.Pop().(reader.Symbol)
			selfVm.Push(selfVm.Env.Frame[sym.GetValue()])
			selfVm.Pc++
		case "define":
			sym := reader.NewSymbol(opCodeAndArgs[1])
			selfVm.Env.Frame[opCodeAndArgs[1]] = selfVm.Pop()
			selfVm.Push(sym)
			selfVm.Pc++
		case "define-args":
			sym := reader.NewSymbol(opCodeAndArgs[1])
			selfVm.Push(sym)
			selfVm.Pc++
		case "load-sexp":
			r := bufio.NewReader(strings.NewReader(opCodeAndArgs[1]))
			sexp, err := reader.NewReader(r).Read()
			if err != nil {
				panic(err)
			}

			selfVm.Push(sexp)
			selfVm.Pc++
		case "set":
			sym := reader.NewSymbol(opCodeAndArgs[1])

			thisVm := selfVm
			for {
				thisVm.Mutex.RLock()
				if thisVm.Env.Frame[sym.GetValue()] != nil {
					thisVm.Mutex.RUnlock()
					break
				}
				thisVm = thisVm.Cont
				thisVm.Mutex.RUnlock()
			}

			thisVm.Mutex.Lock()
			thisVm.Env.Frame[sym.GetValue()] = selfVm.Pop()
			thisVm.Mutex.Unlock()
			selfVm.Pc++
		case "new-env":
			env := &Env{
				Frame: make(map[string]reader.SExpression),
			}
			selfVm.Push(env)
			selfVm.Pc++
		case "create-lambda":
			argsSizeAndCodeLen := strings.SplitN(opCodeAndArgs[1], " ", 2)
			argsSize, _ := strconv.ParseInt(argsSizeAndCodeLen[0], 10, 64)
			codeLen, _ := strconv.ParseInt(argsSizeAndCodeLen[1], 10, 64)

			pc := selfVm.Pc

			newVm := NewVM()

			for i := int64(1); i <= codeLen; i++ {
				newVm.Code = append(newVm.Code, selfVm.Code[pc+i])
				selfVm.Pc++
			}

			for i := int64(0); i < argsSize; i++ {
				sym := selfVm.Pop().(reader.Symbol)
				newVm.TemporaryArgs = append(newVm.TemporaryArgs, sym)
			}

			newVm.Cont = selfVm
			newVm.Env = selfVm.Pop().(*Env)
			newVm.Pc = 0
			newVm.ContPc = selfVm.Pc + 1
			selfVm.Push(newVm)
			selfVm.Pc++
		case "call":
			nextVm := selfVm.Pop().(*VM)
			env := nextVm.Env
			for _, sym := range nextVm.TemporaryArgs {
				env.Frame[sym.GetValue()] = selfVm.Pop()
			}
			nextVm.Env = env
			selfVm = nextVm
		case "ret":
			val := selfVm.Pop()
			retPc := selfVm.ContPc
			selfVm.Pc = 0
			selfVm = selfVm.Cont
			selfVm.Pc = retPc
			selfVm.Push(val)
			selfVm.Pc++
		case "ramdom-id":
			id := uuid.New()
			selfVm.Push(reader.NewString(id.String()))
			selfVm.Pc++
		case "call-native":
			funcNameAndArgLen := strings.SplitN(opCodeAndArgs[1], " ", 2)
			argLen, _ := strconv.ParseInt(funcNameAndArgLen[1], 10, 64)
			switch funcNameAndArgLen[0] {
			case "print":
				line := ""
				for i := int64(0); i < argLen; i++ {
					line += selfVm.Pop().String()
				}
				fmt.Print(line)
				selfVm.Push(reader.NewNil())
				selfVm.Pc++
			case "println":
				line := ""
				for i := int64(0); i < argLen; i++ {
					line += selfVm.Pop().String()
				}
				fmt.Println(line)
				selfVm.Push(reader.NewNil())
				selfVm.Pc++
			case "+":
				sum := int64(0)
				for i := int64(0); i < argLen; i++ {
					sum += selfVm.Pop().(reader.Number).GetValue()
				}
				selfVm.Push(reader.NewInt(sum))
				selfVm.Pc++
			case "-":
				sum := selfVm.Pop().(reader.Number).GetValue()
				for i := int64(1); i < argLen; i++ {
					sum -= selfVm.Pop().(reader.Number).GetValue()
				}
				selfVm.Push(reader.NewInt(sum))
				selfVm.Pc++
			case "*":
				sum := int64(1)
				for i := int64(0); i < argLen; i++ {
					sum *= selfVm.Pop().(reader.Number).GetValue()
				}
				selfVm.Push(reader.NewInt(sum))
				selfVm.Pc++
			case "/":
				sum := selfVm.Pop().(reader.Number).GetValue()
				for i := int64(1); i < argLen; i++ {
					sum /= selfVm.Pop().(reader.Number).GetValue()
				}
				selfVm.Push(reader.NewInt(sum))
				selfVm.Pc++
			case "mod":
				sum := selfVm.Pop().(reader.Number).GetValue()
				for i := int64(1); i < argLen; i++ {
					sum %= selfVm.Pop().(reader.Number).GetValue()
				}
				selfVm.Push(reader.NewInt(sum))
				selfVm.Pc++
			case "=":
				val := selfVm.Pop()
				var tmp reader.SExpression
				var result = true
				for i := int64(1); i < argLen; i++ {
					tmp = selfVm.Pop()
					if result == false {
						continue
					}
					if reader.SExpressionTypeNumber != tmp.SExpressionTypeId() {
						fmt.Println("arg is not number")
					}
					if !val.Equals(tmp) {
						result = false
					}
				}
				if result {
					selfVm.Push(reader.NewBool(true))
				} else {
					selfVm.Push(reader.NewBool(false))
				}
				selfVm.Pc++
			case "!=":
				val := selfVm.Pop()
				var tmp reader.SExpression
				var result = true
				for i := int64(1); i < argLen; i++ {
					tmp = selfVm.Pop()
					if result == false {
						continue
					}
					if val.Equals(tmp) {
						result = false
					}
				}
				if result {
					selfVm.Push(reader.NewBool(true))
				} else {
					selfVm.Push(reader.NewBool(false))
				}
				selfVm.Pc++
			case ">":
				val := selfVm.Pop().(reader.Number).GetValue()
				var tmp reader.SExpression
				flag := true
				for i := int64(1); i < argLen; i++ {
					tmp = selfVm.Pop()
					if flag == false {
						continue
					}
					if val <= tmp.(reader.Number).GetValue() {
						flag = false
					}
				}
				if flag {
					selfVm.Push(reader.NewBool(true))
				} else {
					selfVm.Push(reader.NewBool(false))
				}
				selfVm.Pc++
			case "<":
				val := selfVm.Pop().(reader.Number).GetValue()
				var tmp reader.SExpression
				flag := true
				for i := int64(1); i < argLen; i++ {
					tmp = selfVm.Pop()
					if flag == false {
						continue
					}
					if val >= tmp.(reader.Number).GetValue() {
						flag = false
					}
				}
				if flag {
					selfVm.Push(reader.NewBool(true))
				} else {
					selfVm.Push(reader.NewBool(false))
				}
				selfVm.Pc++
			case ">=":
				val := selfVm.Pop().(reader.Number).GetValue()
				var tmp reader.SExpression
				flag := true
				for i := int64(1); i < argLen; i++ {
					tmp = selfVm.Pop()
					if flag == false {
						continue
					}
					if val < tmp.(reader.Number).GetValue() {
						flag = false
					}
				}
				if flag {
					selfVm.Push(reader.NewBool(true))
				} else {
					selfVm.Push(reader.NewBool(false))
				}
				selfVm.Pc++

			case "<=":
				val := selfVm.Pop().(reader.Number).GetValue()
				var tmp reader.SExpression
				flag := true
				for i := int64(1); i < argLen; i++ {
					tmp = selfVm.Pop()
					if flag == false {
						continue
					}
					if val > tmp.(reader.Number).GetValue() {
						flag = false
					}
				}
				if flag {
					selfVm.Push(reader.NewBool(true))
				} else {
					selfVm.Push(reader.NewBool(false))
				}
				selfVm.Pc++
			}
		case "end-code":
			fmt.Println(selfVm.Pop())
			selfVm.Stack = []reader.SExpression{}
			selfVm.Pc++
			goto ESCAPE
		}
	}
ESCAPE:
	{
	}
}

func (vm *VM) Push(sexp reader.SExpression) {
	vm.Stack = append(vm.Stack, sexp)
}

func (vm *VM) Pop() reader.SExpression {
	if len(vm.Stack) == 0 {
		return nil
	}

	ret := vm.Stack[len(vm.Stack)-1]
	vm.Stack = vm.Stack[:len(vm.Stack)-1]
	return ret
}

func (vm *VM) Peek() reader.SExpression {
	if len(vm.Stack) == 0 {
		return nil
	}

	return vm.Stack[len(vm.Stack)-1]
}

func (vm *VM) SetEnv(env *Env) {
	vm.Env = env
}

func (vm *VM) GetEnv() *Env {
	return vm.Env
}

func (vm *VM) SetCont(cont *VM) {
	vm.Cont = cont
}

func (vm *VM) GetCont() *VM {
	return vm.Cont
}

func (vm *VM) SetCode(code []reader.SExpression) {
	vm.Code = code
}

func (vm *VM) AddCode(code []reader.SExpression) {
	vm.Code = append(vm.Code, code...)
}

func (vm *VM) GetCode() []reader.SExpression {
	return vm.Code
}

func (vm *VM) SetPc(pc int64) {
	vm.Pc = pc
}

func (vm *VM) GetPc() int64 {
	return vm.Pc
}
