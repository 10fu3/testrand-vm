package vm

import (
	"bufio"
	"fmt"
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
	Mutex *sync.RWMutex
	Stack []reader.SExpression
	Code  []reader.SExpression
	Pc    int64
	Env   *Env
	Cont  *VM
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
	for {
		if vm.Pc >= int64(len(vm.Code)) {
			break
		}

		rawCode := vm.Code[vm.Pc].(reader.Symbol).GetValue()
		var opCodeAndArgs = strings.SplitN(rawCode, " ", 2)

		switch opCodeAndArgs[0] {
		case "push-sym":
			vm.Mutex.Lock()
			vm.Push(reader.NewSymbol(opCodeAndArgs[1]))
			vm.Mutex.Unlock()
		case "push-num":
			convertedStrInt64, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			vm.Mutex.Lock()
			vm.Push(reader.NewInt(convertedStrInt64))
			vm.Mutex.Unlock()
		case "push-boo":
			convertedStrBool, _ := strconv.ParseBool(opCodeAndArgs[1])
			vm.Mutex.Lock()
			vm.Push(reader.NewBool(convertedStrBool))
			vm.Mutex.Unlock()
		case "push-str":
			vm.Mutex.Lock()
			vm.Push(reader.NewString(opCodeAndArgs[1]))
			vm.Mutex.Unlock()
		case "pop":
			vm.Mutex.Lock()
			vm.Pop()
			vm.Mutex.Unlock()
		case "jump":
			jumpTo, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			vm.Mutex.Lock()
			vm.Pc = jumpTo - 1
			vm.Mutex.Unlock()
		case "load":
			vm.Mutex.Lock()
			sym := vm.Pop().(reader.Symbol)
			vm.Push(vm.Env.Frame[sym.GetValue()])
			vm.Mutex.Unlock()
		case "define":
			vm.Mutex.Lock()
			sym := vm.Pop().(reader.Symbol)
			vm.Env.Frame[sym.GetValue()] = vm.Pop()
			vm.Mutex.Unlock()
		case "load-sexp":
			r := bufio.NewReader(strings.NewReader(opCodeAndArgs[1]))
			sexp, err := reader.NewReader(r).Read()
			if err != nil {
				panic(err)
			}
			vm.Mutex.Lock()
			vm.Push(sexp)
			vm.Mutex.Unlock()
		case "print":
			vm.Mutex.Lock()
			sexp := vm.Pop()
			vm.Mutex.Unlock()
			fmt.Println(sexp.String())
		case "set":
			sym := reader.NewSymbol(opCodeAndArgs[1])

			thisVm := vm
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
			thisVm.Env.Frame[sym.GetValue()] = vm.Pop()
			thisVm.Mutex.Unlock()

		case "new-env":
			vm.Mutex.Lock()
			vm.Push(&Env{Frame: make(map[string]reader.SExpression)})
			vm.Mutex.Lock()
		case "create-lambda":
			codeLen, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)

			pc := vm.Pc

			newVm := NewVM()

			for i := int64(1); i <= codeLen; i++ {
				newVm.Code = append(newVm.Code, vm.Code[pc+i])
			}

			newVm.Cont = vm
			newVm.Env = vm.Pop().(*Env)
			newVm.Pc = 0
			vm.Mutex.Lock()
			vm.Push(newVm)
			vm.Mutex.Unlock()
		case "call":
			vm.Mutex.Lock()
			vm = vm.Pop().(*VM)
			vm.Mutex.Unlock()
		case "ret":
			vm.Mutex.Lock()
			val := vm.Pop()
			vm = vm.Cont
			vm.Push(val)
			vm.Mutex.Unlock()
		}
		vm.Pc++
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

func (vm *VM) GetCode() []reader.SExpression {
	return vm.Code
}

func (vm *VM) SetPc(pc int64) {
	vm.Pc = pc
}

func (vm *VM) GetPc() int64 {
	return vm.Pc
}
