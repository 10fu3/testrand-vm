package vm

import (
	"fmt"
	"github.com/google/uuid"
	"sync"
	"testrand-vm/instr"
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
	Code          []instr.Instr
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

func (vm *Closure) Clone() *Closure {

	//stack clone
	stack := make([]reader.SExpression, len(vm.Stack))
	for i, v := range vm.Stack {
		stack[i] = v
	}

	//code clone
	code := make([]instr.Instr, len(vm.Code))
	for i, v := range vm.Code {
		code[i] = v
	}

	return &Closure{
		Mutex:         &sync.RWMutex{},
		Stack:         stack,
		Code:          code,
		Pc:            vm.Pc,
		Env:           vm.Env,
		Cont:          vm.Cont,
		ReturnCont:    vm.ReturnCont,
		ReturnPc:      vm.ReturnPc,
		TemporaryArgs: vm.TemporaryArgs,
	}
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

		//rawCode := selfVm.Code[selfVm.Pc].(reader.Symbol).GetValue()
		code := selfVm.Code[selfVm.Pc]

		//switch opCodeAndArgs[0] {
		switch code.Type {
		case instr.OPCODE_PUSH_NIL:
			selfVm.Push(reader.NewNil())
			selfVm.Pc++
		//case "push-sym":
		case instr.OPCODE_PUSH_SYM:
			//selfVm.Push(reader.NewSymbol(opCodeAndArgs[1]))
			selfVm.Push(instr.DeserializePushSymbolInstr(code))
			selfVm.Pc++

		//case "push-num":
		case instr.OPCODE_PUSH_NUM:
			selfVm.Push(reader.NewInt(instr.DeserializePushNumberInstr(code)))
			selfVm.Pc++
		//case "push-boo":
		case instr.OPCODE_PUSH_TRUE:
			//val := opCodeAndArgs[1] == "#t"
			//
			//if opCodeAndArgs[1] != "#f" && opCodeAndArgs[1] != "#t" {
			//	fmt.Println("not a bool")
			//	goto ESCAPE
			//}
			selfVm.Push(reader.NewBool(true))
			selfVm.Pc++
		case instr.OPCODE_PUSH_FALSE:
			selfVm.Push(reader.NewBool(false))
			selfVm.Pc++
		//case "push-str":
		case instr.OPCODE_PUSH_STR:
			//selfVm.Push(reader.NewString(opCodeAndArgs[1]))
			selfVm.Push(reader.NewString(instr.DeserializePushStringInstr(code)))
			selfVm.Pc++
		//case "pop":
		case instr.OPCODE_POP:
			selfVm.Pop()
			selfVm.Pc++
		//case "jmp":
		case instr.OPCODE_JMP:
			//jumpTo, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			jumpTo := instr.DeserializeJmpInstr(code)
			selfVm.Pc = jumpTo
		//case "jmp-if":
		case instr.OPCODE_JMP_IF:
			//jumpTo, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			jumpTo := instr.DeserializeJmpIfInstr(code)
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
		//case "jmp-else":
		case instr.OPCODE_JMP_ELSE:
			//jumpTo, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			jumpTo := instr.DeserializeJmpElseInstr(code)
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

		//case "load":
		case instr.OPCODE_LOAD:
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
		//case "define":
		case instr.OPCODE_DEFINE:
			//sym := reader.NewSymbol(opCodeAndArgs[1])
			deserialize := instr.DeserializeDefineInstr(code)
			val := selfVm.Pop()
			//selfVm.Env.Frame[opCodeAndArgs[1]] = &val
			selfVm.Env.Frame[deserialize] = &val
			//selfVm.Push(sym)
			selfVm.Push(reader.NewSymbol(deserialize))
			selfVm.Pc++
		//case "define-args":
		case instr.OPCODE_DEFINE_ARGS:
			//sym := reader.NewSymbol(opCodeAndArgs[1])
			deserialize := instr.DeserializeDefineArgsInstr(code)
			selfVm.Push(reader.NewSymbol(deserialize))
			selfVm.Pc++
		//case "load-sexp":
		case instr.OPCODE_PUSH_SEXP:
			//r := bufio.NewReader(strings.NewReader(opCodeAndArgs[1]))
			//sexp, err := reader.NewReader(r).Read()
			//if err != nil {
			//	panic(err)
			//}
			deserialize, err := instr.DeserializeSexpressionInstr(code)

			if err != nil {
				panic(err)
			}
			selfVm.Push(deserialize)
			selfVm.Pc++
		//case "set":
		case instr.OPCODE_SET:
			//sym := reader.NewSymbol(opCodeAndArgs[1])
			deserialize := instr.DeserializeSetInstr(code)

			thisVm := selfVm
			for {
				//if thisVm.Env.Frame[sym.GetValue()] != nil {
				if thisVm.Env.Frame[deserialize] != nil {
					break
				}
				if thisVm.Cont == nil {
					break
				}
				thisVm = thisVm.Cont
			}

			val := selfVm.Pop()
			thisVm.Env.Frame[deserialize] = &val
			selfVm.Push(val)
			selfVm.Pc++
		//case "new-env":
		case instr.OPCODE_NEW_ENV:
			env := &Env{
				Frame: make(map[string]*reader.SExpression),
			}
			selfVm.Push(env)
			selfVm.Pc++
		//case "create-lambda":
		case instr.OPCODE_CREATE_CLOSURE:
			//argsSizeAndCodeLen := strings.SplitN(opCodeAndArgs[1], " ", 2)
			//argsSize, _ := strconv.ParseInt(argsSizeAndCodeLen[0], 10, 64)
			//codeLen, _ := strconv.ParseInt(argsSizeAndCodeLen[1], 10, 64)

			argsSize, codeLen := instr.DeserializeCreateClosureInstr(code)

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
		//case "call":
		case instr.OPCODE_CALL:
			rawClosure := selfVm.Pop()

			if rawClosure.SExpressionTypeId() != reader.SExpressionTypeClosure {
				fmt.Println("not a closure")
				goto ESCAPE
			}

			nextVm := rawClosure.(*Closure)
			env := nextVm.Env

			argsSize := instr.DeserializeCallInstr(code)

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
		//case "ret":
		case instr.OPCODE_RETURN:
			val := selfVm.Pop()
			retPc := selfVm.ReturnPc
			selfVm.Stack = []reader.SExpression{}
			selfVm.Pc = 0
			selfVm = selfVm.ReturnCont
			selfVm.Pc = retPc
			selfVm.Push(val)
			selfVm.Pc++
		//case "and":
		case instr.OPCODE_AND:
			//argsSize, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			argsSize := instr.DeserializeAndInstr(code)
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
		//case "or":
		case instr.OPCODE_OR:
			//argsSize, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			argsSize := instr.DeserializeOrInstr(code)
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
		//case "end-code":
		case instr.OPCODE_END_CODE:
			fmt.Println(selfVm.Pop())
			goto ESCAPE
		case instr.OPCODE_NOP:
			selfVm.Pc++

		//case "print":
		case instr.OPCODE_PRINT:
			argLen := instr.DeserializePrintInstr(code)
			line := ""
			for i := int64(0); i < argLen; i++ {
				line += selfVm.Pop().String()
			}
			fmt.Print(line)
			selfVm.Push(reader.NewNil())
			selfVm.Pc++
		//case "println":
		case instr.OPCODE_PRINTLN:
			argLen := instr.DeserializePrintlnInstr(code)
			line := ""
			for i := int64(0); i < argLen; i++ {
				line += selfVm.Pop().String()
			}
			fmt.Println(line)
			selfVm.Push(reader.NewNil())
			selfVm.Pc++
		//case "+":
		case instr.OPCODE_PLUS_NUM:
			argLen := instr.DeserializePlusNumInstr(code)
			sum := int64(0)
			for i := int64(0); i < argLen; i++ {
				sum += selfVm.Pop().(reader.Number).GetValue()
			}
			selfVm.Push(reader.NewInt(sum))
			selfVm.Pc++
		//case "-":
		case instr.OPCODE_MINUS_NUM:
			argLen := instr.DeserializeMinusNumInstr(code)
			sum := selfVm.Pop().(reader.Number).GetValue()
			for i := int64(1); i < argLen; i++ {
				sum -= selfVm.Pop().(reader.Number).GetValue()
			}
			selfVm.Push(reader.NewInt(sum))
			selfVm.Pc++
		//case "*":
		case instr.OPCODE_MULTIPLY_NUM:
			argLen := instr.DeserializeMultiplyNumInstr(code)
			sum := int64(1)
			for i := int64(0); i < argLen; i++ {
				sum *= selfVm.Pop().(reader.Number).GetValue()
			}
			selfVm.Push(reader.NewInt(sum))
			selfVm.Pc++
		//case "/":
		case instr.OPCODE_DIVIDE_NUM:
			argLen := instr.DeserializeDivideNumInstr(code)
			sum := selfVm.Pop().(reader.Number).GetValue()
			for i := int64(1); i < argLen; i++ {
				sum /= selfVm.Pop().(reader.Number).GetValue()
			}
			selfVm.Push(reader.NewInt(sum))
			selfVm.Pc++
		//case "mod":
		case instr.OPCODE_MODULO_NUM:
			argLen := instr.DeserializeModuloNumInstr(code)
			sum := selfVm.Pop().(reader.Number).GetValue()
			for i := int64(1); i < argLen; i++ {
				sum %= selfVm.Pop().(reader.Number).GetValue()
			}
			selfVm.Push(reader.NewInt(sum))
			selfVm.Pc++
		//case "=":
		case instr.OPCODE_EQUAL_NUM:
			argLen := instr.DeserializeEqualNumInstr(code)
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
		//case "!=":
		case instr.OPCODE_NOT_EQUAL_NUM:
			argLen := instr.DeserializeNotEqualNumInstr(code)
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
		//case ">":
		case instr.OPCODE_GREATER_THAN_NUM:
			argLen := instr.DeserializeGreaterThanNumInstr(code)
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
		//case "<":
		case instr.OPCODE_LESS_THAN_NUM:
			argLen := instr.DeserializeLessThanNumInstr(code)
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
		//case ">=":
		case instr.OPCODE_GREATER_THAN_OR_EQUAL_NUM:
			argLen := instr.DeserializeGreaterThanOrEqualNumInstr(code)
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

		//case "<=":
		case instr.OPCODE_LESS_THAN_OR_EQUAL_NUM:
			argLen := instr.DeserializeLessThanOrEqualNumInstr(code)
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
		//case "car":
		case instr.OPCODE_CAR:
			target := selfVm.Pop()
			if target.SExpressionTypeId() != reader.SExpressionTypeConsCell {
				fmt.Println("car target is not cons cell")
			}
			selfVm.Push(target.(reader.ConsCell).GetCar())
			selfVm.Pc++
		//case "cdr":
		case instr.OPCODE_CDR:
			target := selfVm.Pop()
			if target.SExpressionTypeId() != reader.SExpressionTypeConsCell {
				fmt.Println("cdr target is not cons cell")
			}
			selfVm.Push(target.(reader.ConsCell).GetCdr())
			selfVm.Pc++
		//case "random-id":
		case instr.OPCODE_RANDOM_ID:
			id := uuid.New()
			selfVm.Push(reader.NewString(id.String()))
			selfVm.Pc++
		}
	}
ESCAPE:
	{
		for {
			selfVm.Stack = []reader.SExpression{}
			selfVm.Code = []instr.Instr{}
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

func (vm *Closure) SetCode(code []instr.Instr) {
	vm.Code = code
}

func (vm *Closure) AddCode(code []instr.Instr) {
	vm.Code = append(vm.Code, code...)
}

func (vm *Closure) GetCode() []instr.Instr {
	return vm.Code
}

func (vm *Closure) SetPc(pc int64) {
	vm.Pc = pc
}

func (vm *Closure) GetPc() int64 {
	return vm.Pc
}
