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
	Frame map[string]*reader.SExpression
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

type Closure struct {
	Mutex         *sync.RWMutex
	Stack         []reader.SExpression
	Code          []reader.SExpression
	Pc            int64
	Env           *Env
	Cont          *Closure
	ReturnCont    *Closure
	ReturnPc      int64
	TemporaryArgs []reader.Symbol
}

func (vm *Closure) TypeId() string {
	return "closure"
}

func (vm *Closure) SExpressionTypeId() reader.SExpressionType {
	return reader.SExpressionTypeClosure
}

func (vm *Closure) String() string {
	return "closure"
}

func (vm *Closure) IsList() bool {
	return false
}

func (vm *Closure) Equals(sexp reader.SExpression) bool {
	//TODO implement me
	panic("implement me")
}

func NewVM() *Closure {
	return &Closure{
		Stack: make([]reader.SExpression, 0),
		Pc:    0,
		Env:   &Env{Frame: make(map[string]*reader.SExpression)},
		Cont:  nil,
		Mutex: &sync.RWMutex{},
	}
}

func VMRun(vm *Closure) {

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
			val := opCodeAndArgs[1] == "#t"

			if opCodeAndArgs[1] != "#f" && opCodeAndArgs[1] != "#t" {
				fmt.Println("not a bool")
				goto ESCAPE
			}

			selfVm.Push(reader.NewBool(val))
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

		case "jump-if":
			jumpTo, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			val := selfVm.Pop()
			if val.SExpressionTypeId() != reader.SExpressionTypeBool {
				fmt.Println("not a bool")
				goto ESCAPE
			}
			if val.(reader.Bool).GetValue() {
				selfVm.Pc = jumpTo
			} else {
				selfVm.Pc++
			}
		case "jump-else":
			jumpTo, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			val := selfVm.Pop()
			if val.SExpressionTypeId() != reader.SExpressionTypeBool {
				fmt.Println("not a bool")
				goto ESCAPE
			}

			if !val.(reader.Bool).GetValue() {
				selfVm.Pc = jumpTo
			} else {
				selfVm.Pc++
			}

		case "load":
			sym := selfVm.Pop().(reader.Symbol)

			meVm := selfVm
			found := false
			for {
				if meVm.Env.Frame[sym.GetValue()] != nil {
					found = true
					break
				}
				if meVm.Cont == nil {
					break
				}
				meVm = meVm.Cont
			}
			if found {
				selfVm.Push(*meVm.Env.Frame[sym.GetValue()])
				selfVm.Pc++
			} else {
				fmt.Println("Symbol not found: " + sym.GetValue())
				goto ESCAPE
			}
		case "define":
			sym := reader.NewSymbol(opCodeAndArgs[1])
			val := selfVm.Pop()
			selfVm.Env.Frame[opCodeAndArgs[1]] = &val
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
			val := selfVm.Pop()
			thisVm.Env.Frame[sym.GetValue()] = &val
			thisVm.Mutex.Unlock()
			selfVm.Push(val)
			selfVm.Pc++
		case "new-env":
			env := &Env{
				Frame: make(map[string]*reader.SExpression),
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
			selfVm.Push(newVm)
			selfVm.Pc++
		case "call":
			rawClosure := selfVm.Pop()

			if rawClosure.SExpressionTypeId() != reader.SExpressionTypeClosure {
				fmt.Println("not a closure")
				goto ESCAPE
			}

			nextVm := rawClosure.(*Closure)
			env := nextVm.Env

			argsSize, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)

			if argsSize != int64(len(nextVm.TemporaryArgs)) {
				fmt.Println("args size not match")
				goto ESCAPE
			}

			for _, sym := range nextVm.TemporaryArgs {
				val := selfVm.Pop()
				env.Frame[sym.GetValue()] = &val
			}
			nextVm.Env = env
			nextVm.ReturnCont = selfVm
			nextVm.ReturnPc = selfVm.Pc
			selfVm = nextVm
		case "ret":
			val := selfVm.Pop()
			retPc := selfVm.ReturnPc
			selfVm.Stack = []reader.SExpression{}
			selfVm.Pc = 0
			selfVm = selfVm.ReturnCont
			selfVm.Pc = retPc
			selfVm.Push(val)
			selfVm.Pc++
		case "and":
			argsSize, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			val := selfVm.Pop().(reader.Bool).GetValue()
			var tmp reader.SExpression
			flag := true
			for i := int64(1); i < argsSize; i++ {
				tmp = selfVm.Pop()
				if flag == false {
					continue
				}
				if val != tmp.(reader.Bool).GetValue() {
					flag = false
				}
			}
			if flag {
				selfVm.Push(reader.NewBool(true))
			} else {
				selfVm.Push(reader.NewBool(false))
			}
			selfVm.Pc++
		case "or":
			argsSize, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			var tmp reader.SExpression
			flag := false
			for i := int64(1); i < argsSize; i++ {
				tmp = selfVm.Pop()
				if flag == true {
					continue
				}
				if tmp.(reader.Bool).GetValue() {
					flag = true
				}
			}
			if flag {
				selfVm.Push(reader.NewBool(true))
			} else {
				selfVm.Push(reader.NewBool(false))
			}
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
			case "<":
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
			case ">=":
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

			case "<=":
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
			case "car":
				target := selfVm.Pop()
				if target.SExpressionTypeId() != reader.SExpressionTypeConsCell {
					fmt.Println("car target is not cons cell")
				}
				selfVm.Push(target.(reader.ConsCell).GetCar())
				selfVm.Pc++
			case "cdr":
				target := selfVm.Pop()
				if target.SExpressionTypeId() != reader.SExpressionTypeConsCell {
					fmt.Println("cdr target is not cons cell")
				}
				selfVm.Push(target.(reader.ConsCell).GetCdr())
				selfVm.Pc++
			case "ramdom-id":
				id := uuid.New()
				selfVm.Push(reader.NewString(id.String()))
				selfVm.Pc++
			}
		case "end-code":
			fmt.Println(selfVm.Pop())
			goto ESCAPE
		}
	}
ESCAPE:
	{
		for {
			selfVm.Stack = []reader.SExpression{}
			selfVm.Code = []reader.SExpression{}
			selfVm.Pc = 0
			if selfVm.ReturnCont == nil {
				break
			}
			selfVm = selfVm.ReturnCont
		}
	}
}

func (vm *Closure) Push(sexp reader.SExpression) {
	vm.Stack = append(vm.Stack, sexp)
}

func (vm *Closure) Pop() reader.SExpression {
	if len(vm.Stack) == 0 {
		return nil
	}

	ret := vm.Stack[len(vm.Stack)-1]
	vm.Stack = vm.Stack[:len(vm.Stack)-1]
	return ret
}

func (vm *Closure) Peek() reader.SExpression {
	if len(vm.Stack) == 0 {
		return nil
	}

	return vm.Stack[len(vm.Stack)-1]
}

func (vm *Closure) SetEnv(env *Env) {
	vm.Env = env
}

func (vm *Closure) GetEnv() *Env {
	return vm.Env
}

func (vm *Closure) SetCont(cont *Closure) {
	vm.Cont = cont
}

func (vm *Closure) GetCont() *Closure {
	return vm.Cont
}

func (vm *Closure) SetCode(code []reader.SExpression) {
	vm.Code = code
}

func (vm *Closure) AddCode(code []reader.SExpression) {
	vm.Code = append(vm.Code, code...)
}

func (vm *Closure) GetCode() []reader.SExpression {
	return vm.Code
}

func (vm *Closure) SetPc(pc int64) {
	vm.Pc = pc
}

func (vm *Closure) GetPc() int64 {
	return vm.Pc
}
