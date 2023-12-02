package vm

import (
	"fmt"
	"github.com/google/uuid"
	"sync"
	"testrand-vm/instr"
	"testrand-vm/reader"
)

type Env struct {
	Frame  map[string]*reader.SExpression
	Parent *Env
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

type Stack struct {
	Stack []reader.SExpression
}

func (st *Stack) Push(sexp reader.SExpression) {
	st.Stack = append(st.Stack, sexp)
}

func (vm *Stack) Pop() reader.SExpression {
	if len(vm.Stack) == 0 {
		return nil
	}

	ret := vm.Stack[len(vm.Stack)-1]
	vm.Stack = vm.Stack[:len(vm.Stack)-1]
	return ret
}

func (vm *Stack) Peek() reader.SExpression {
	if len(vm.Stack) == 0 {
		return nil
	}

	return vm.Stack[len(vm.Stack)-1]
}

func (vm *Stack) PeekIndex(index int) (reader.SExpression, error) {

	if index < 0 || len(vm.Stack) <= index {
		return nil, fmt.Errorf("index out of range")
	}

	return vm.Stack[index], nil
}

type Continuation struct {
	Stack
	Env    *Env
	Code   []instr.Instr
	Pc     int64
	Parent *Continuation
}

func (c *Continuation) TypeId() string {
	return "continuation"
}

func (c *Continuation) SExpressionTypeId() reader.SExpressionType {
	return reader.SExpressionTypeContinuation
}

func (c *Continuation) String() string {
	return "continuation"
}

func (c *Continuation) IsList() bool {
	return false
}

func (c *Continuation) Equals(sexp reader.SExpression) bool {
	if sexp.TypeId() != "continuation" {
		return false
	}
	return sexp.(*Continuation) == c
}

func NewContinuation() *Continuation {
	return &Continuation{
		Stack: Stack{Stack: make([]reader.SExpression, 0)},
		Pc:    0,
		Env:   &Env{Frame: make(map[string]*reader.SExpression)},
	}
}

type Closure struct {
	Cont          *Continuation
	TemporaryArgs []reader.Symbol
}

func (c *Closure) TypeId() string {
	return "closure"
}

func (c *Closure) SExpressionTypeId() reader.SExpressionType {
	return reader.SExpressionTypeClosure
}

func (c *Closure) String() string {
	return "closure"
}

func (c *Closure) IsList() bool {
	return false
}

func (c *Closure) Equals(sexp reader.SExpression) bool {
	if sexp.TypeId() != "closure" {
		return false
	}
	return sexp.(*Closure) == c
}

type Machine struct {
	Mutex *sync.RWMutex
	Cont  *Continuation
}

func (vm *Machine) TypeId() string {
	return "closure"
}

func (vm *Machine) SExpressionTypeId() reader.SExpressionType {
	return reader.SExpressionTypeClosure
}

func (vm *Machine) String() string {
	return "closure"
}

func (vm *Machine) IsList() bool {
	return false
}

func (vm *Machine) Equals(sexp reader.SExpression) bool {
	//TODO implement me
	panic("implement me")
}

func NewVM() *Machine {
	return &Machine{
		Mutex: &sync.RWMutex{},
		Cont: &Continuation{
			Stack: Stack{Stack: make([]reader.SExpression, 0)},
			Pc:    0,
			Env:   &Env{Frame: make(map[string]*reader.SExpression)},
			Code:  make([]instr.Instr, 0),
		},
	}
}

func VMRun(vm *Machine) {

	selfVm := vm

	for {

		//rawCode := selfVm.Cont.Code[selfVm.Cont.Pc].(reader.Symbol).GetValue()
		code := selfVm.Cont.Code[selfVm.Cont.Pc]

		//switch opCodeAndArgs[0] {
		switch code.Type {
		case instr.OPCODE_PUSH_NIL:
			selfVm.Cont.Push(reader.NewNil())
			selfVm.Cont.Pc++
		//case "push-sym":
		case instr.OPCODE_PUSH_SYM:
			//selfVm.Cont.Push(reader.NewSymbol(opCodeAndArgs[1]))
			selfVm.Cont.Push(instr.DeserializePushSymbolInstr(code))
			selfVm.Cont.Pc++

		//case "push-num":
		case instr.OPCODE_PUSH_NUM:
			selfVm.Cont.Push(reader.NewInt(instr.DeserializePushNumberInstr(code)))
			selfVm.Cont.Pc++
		//case "push-boo":
		case instr.OPCODE_PUSH_TRUE:
			//val := opCodeAndArgs[1] == "#t"
			//
			//if opCodeAndArgs[1] != "#f" && opCodeAndArgs[1] != "#t" {
			//	fmt.Println("not a bool")
			//	goto ESCAPE
			//}
			selfVm.Cont.Push(reader.NewBool(true))
			selfVm.Cont.Pc++
		case instr.OPCODE_PUSH_FALSE:
			selfVm.Cont.Push(reader.NewBool(false))
			selfVm.Cont.Pc++
		//case "push-str":
		case instr.OPCODE_PUSH_STR:
			//selfVm.Cont.Push(reader.NewString(opCodeAndArgs[1]))
			selfVm.Cont.Push(reader.NewString(instr.DeserializePushStringInstr(code)))
			selfVm.Cont.Pc++
		//case "pop":
		case instr.OPCODE_POP:
			selfVm.Cont.Pop()
			selfVm.Cont.Pc++
		//case "jmp":
		case instr.OPCODE_JMP:
			//jumpTo, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			jumpTo := instr.DeserializeJmpInstr(code)
			selfVm.Cont.Pc = jumpTo
		//case "jmp-if":
		case instr.OPCODE_JMP_IF:
			//jumpTo, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			jumpTo := instr.DeserializeJmpIfInstr(code)
			val := selfVm.Cont.Pop()
			if val.SExpressionTypeId() != reader.SExpressionTypeBool {
				fmt.Println("not a bool")
				goto ESCAPE
			}
			if val.(reader.Bool).GetValue() {
				selfVm.Cont.Pc = jumpTo
			} else {
				selfVm.Cont.Pc++
			}
		//case "jmp-else":
		case instr.OPCODE_JMP_ELSE:
			//jumpTo, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			jumpTo := instr.DeserializeJmpElseInstr(code)
			val := selfVm.Cont.Pop()
			if val.SExpressionTypeId() != reader.SExpressionTypeBool {
				fmt.Println("not a bool")
				goto ESCAPE
			}

			if !val.(reader.Bool).GetValue() {
				selfVm.Cont.Pc = jumpTo
			} else {
				selfVm.Cont.Pc++
			}

		//case "load":
		case instr.OPCODE_LOAD:
			sym := selfVm.Cont.Pop().(reader.Symbol)

			meCont := selfVm.Cont
			found := false
			for {
				if meCont.Env.Frame[sym.GetValue()] != nil {
					found = true
					break
				}
				if meCont.Parent == nil {
					break
				}
				meCont = meCont.Parent
			}
			if found {
				selfVm.Cont.Push(*meCont.Env.Frame[sym.GetValue()])
				selfVm.Cont.Pc++
			} else {
				fmt.Println("Symbol not found: " + sym.GetValue())
				goto ESCAPE
			}
		//case "define":
		case instr.OPCODE_DEFINE:
			//sym := reader.NewSymbol(opCodeAndArgs[1])
			deserialize := instr.DeserializeDefineInstr(code)
			val := selfVm.Cont.Pop()
			//selfVm.Cont.Env.Frame[opCodeAndArgs[1]] = &val
			selfVm.Cont.Env.Frame[deserialize] = &val
			//selfVm.Cont.Push(sym)
			selfVm.Cont.Push(reader.NewSymbol(deserialize))
			selfVm.Cont.Pc++
		//case "define-args":
		case instr.OPCODE_DEFINE_ARGS:
			//sym := reader.NewSymbol(opCodeAndArgs[1])
			deserialize := instr.DeserializeDefineArgsInstr(code)
			selfVm.Cont.Push(reader.NewSymbol(deserialize))
			selfVm.Cont.Pc++
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
			selfVm.Cont.Push(deserialize)
			selfVm.Cont.Pc++
		//case "set":
		case instr.OPCODE_SET:
			//sym := reader.NewSymbol(opCodeAndArgs[1])
			deserialize := instr.DeserializeSetInstr(code)

			thisEnv := selfVm.Cont.Env
			for {
				//if thisVm.Env.Frame[sym.GetValue()] != nil {
				if thisEnv.Frame[deserialize] == nil {
					if thisEnv.Parent == nil {
						fmt.Println("Symbol not found: " + deserialize)
						goto ESCAPE
					}
					thisEnv = thisEnv.Parent
					continue
				}
				break
			}

			val := selfVm.Cont.Pop()
			thisEnv.Frame[deserialize] = &val
			selfVm.Cont.Push(val)
			selfVm.Cont.Pc++
		//case "new-env":
		case instr.OPCODE_NEW_ENV:
			env := &Env{
				Frame: make(map[string]*reader.SExpression),
			}
			selfVm.Cont.Push(env)
			selfVm.Cont.Pc++
		//case "create-lambda":
		case instr.OPCODE_CREATE_CLOSURE:
			//argsSizeAndCodeLen := strings.SplitN(opCodeAndArgs[1], " ", 2)
			//argsSize, _ := strconv.ParseInt(argsSizeAndCodeLen[0], 10, 64)
			//codeLen, _ := strconv.ParseInt(argsSizeAndCodeLen[1], 10, 64)

			argsSize, codeLen := instr.DeserializeCreateClosureInstr(code)

			pc := selfVm.Cont.Pc

			//newVm := NewVM()
			//
			//for i := int64(1); i <= codeLen; i++ {
			//	newVm.Code = append(newVm.Code, selfVm.Cont.Code[pc+i])
			//	selfVm.Cont.Pc++
			//}
			//
			//for i := int64(0); i < argsSize; i++ {
			//	sym := selfVm.Cont.Pop().(reader.Symbol)
			//	newVm.TemporaryArgs = append(newVm.TemporaryArgs, sym)
			//}
			//
			//newVm.Continuation = selfVm
			//newVm.Env = selfVm.Cont.Pop().(*Env)
			//newVm.Pc = 0
			//selfVm.Cont.Push(newVm)
			//selfVm.Cont.Pc++

			newClosure := &Closure{
				Cont: &Continuation{
					Stack: Stack{},
					Env: &Env{
						Frame: make(map[string]*reader.SExpression),
					},
					Code:   nil,
					Pc:     0,
					Parent: selfVm.Cont,
				},
			}

			for i := int64(1); i <= codeLen; i++ {
				newClosure.Cont.Code = append(newClosure.Cont.Code, selfVm.Cont.Code[pc+i])
				selfVm.Cont.Pc++
			}

			for i := int64(0); i < argsSize; i++ {
				sym := selfVm.Cont.Pop().(reader.Symbol)
				newClosure.TemporaryArgs = append(newClosure.TemporaryArgs, sym)
			}
			newClosure.Cont.Env = selfVm.Cont.Pop().(*Env)
			newClosure.Cont.Pc = 0
			selfVm.Cont.Push(newClosure)
			selfVm.Cont.Pc++

		//case "call":
		case instr.OPCODE_CALL:
			rawClosure := selfVm.Cont.Pop()

			if rawClosure.SExpressionTypeId() != reader.SExpressionTypeClosure {
				fmt.Println("not a closure")
				goto ESCAPE
			}

			nextVm := rawClosure.(*Closure)
			env := nextVm.Cont.Env

			argsSize := instr.DeserializeCallInstr(code)

			if argsSize != int64(len(nextVm.TemporaryArgs)) {
				fmt.Println("args size not match")
				goto ESCAPE
			}

			for _, sym := range nextVm.TemporaryArgs {
				val := selfVm.Cont.Pop()
				env.Frame[sym.GetValue()] = &val
			}
			nextVm.Cont.Env = env
			nextVm.Cont.Stack.Push(selfVm.Cont)

			selfVm.Cont = nextVm.Cont
		//case "ret":
		case instr.OPCODE_RETURN:
			val := selfVm.Cont.Pop()
			retCont, err := selfVm.Cont.Stack.PeekIndex(0)
			if err != nil {
				panic(err)
			}

			selfVm.Cont.Stack.Stack = []reader.SExpression{}
			selfVm.Cont.Pc = 0
			selfVm.Cont = retCont.(*Continuation)
			selfVm.Cont.Push(val)
			selfVm.Cont.Pc++
		//case "and":
		case instr.OPCODE_AND:
			//argsSize, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			argsSize := instr.DeserializeAndInstr(code)
			val := selfVm.Cont.Pop().(reader.Bool).GetValue()
			var tmp reader.SExpression
			flag := true
			for i := int64(1); i < argsSize; i++ {
				tmp = selfVm.Cont.Pop()
				if flag == false {
					continue
				}
				if val != tmp.(reader.Bool).GetValue() {
					flag = false
				}
			}
			if flag {
				selfVm.Cont.Push(reader.NewBool(true))
			} else {
				selfVm.Cont.Push(reader.NewBool(false))
			}
			selfVm.Cont.Pc++
		//case "or":
		case instr.OPCODE_OR:
			//argsSize, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			argsSize := instr.DeserializeOrInstr(code)
			var tmp reader.SExpression = selfVm.Cont.Pop()
			flag := false
			for i := int64(1); i < argsSize; i++ {
				if tmp.(reader.Bool).GetValue() {
					flag = true
				}
				if flag == true {
					continue
				}
				tmp = selfVm.Cont.Pop()
			}
			if flag {
				selfVm.Cont.Push(reader.NewBool(true))
			} else {
				selfVm.Cont.Push(reader.NewBool(false))
			}
			selfVm.Cont.Pc++
		//case "end-code":
		case instr.OPCODE_END_CODE:
			fmt.Println(selfVm.Cont.Pop())
			goto ESCAPE
		case instr.OPCODE_NOP:
			selfVm.Cont.Pc++

		//case "print":
		case instr.OPCODE_PRINT:
			argLen := instr.DeserializePrintInstr(code)
			line := ""
			for i := int64(0); i < argLen; i++ {
				line += selfVm.Cont.Pop().String()
			}
			fmt.Print(line)
			selfVm.Cont.Push(reader.NewNil())
			selfVm.Cont.Pc++
		//case "println":
		case instr.OPCODE_PRINTLN:
			argLen := instr.DeserializePrintlnInstr(code)
			line := ""
			for i := int64(0); i < argLen; i++ {
				line += selfVm.Cont.Pop().String()
			}
			fmt.Println(line)
			selfVm.Cont.Push(reader.NewNil())
			selfVm.Cont.Pc++
		//case "+":
		case instr.OPCODE_PLUS_NUM:
			argLen := instr.DeserializePlusNumInstr(code)
			sum := int64(0)
			for i := int64(0); i < argLen; i++ {
				sum += selfVm.Cont.Pop().(reader.Number).GetValue()
			}
			selfVm.Cont.Push(reader.NewInt(sum))
			selfVm.Cont.Pc++
		//case "-":
		case instr.OPCODE_MINUS_NUM:
			argLen := instr.DeserializeMinusNumInstr(code)
			sum := selfVm.Cont.Pop().(reader.Number).GetValue()
			for i := int64(1); i < argLen; i++ {
				sum -= selfVm.Cont.Pop().(reader.Number).GetValue()
			}
			selfVm.Cont.Push(reader.NewInt(sum))
			selfVm.Cont.Pc++
		//case "*":
		case instr.OPCODE_MULTIPLY_NUM:
			argLen := instr.DeserializeMultiplyNumInstr(code)
			sum := int64(1)
			for i := int64(0); i < argLen; i++ {
				sum *= selfVm.Cont.Pop().(reader.Number).GetValue()
			}
			selfVm.Cont.Push(reader.NewInt(sum))
			selfVm.Cont.Pc++
		//case "/":
		case instr.OPCODE_DIVIDE_NUM:
			argLen := instr.DeserializeDivideNumInstr(code)
			sum := selfVm.Cont.Pop().(reader.Number).GetValue()
			for i := int64(1); i < argLen; i++ {
				sum /= selfVm.Cont.Pop().(reader.Number).GetValue()
			}
			selfVm.Cont.Push(reader.NewInt(sum))
			selfVm.Cont.Pc++
		//case "mod":
		case instr.OPCODE_MODULO_NUM:
			argLen := instr.DeserializeModuloNumInstr(code)
			sum := selfVm.Cont.Pop().(reader.Number).GetValue()
			for i := int64(1); i < argLen; i++ {
				sum %= selfVm.Cont.Pop().(reader.Number).GetValue()
			}
			selfVm.Cont.Push(reader.NewInt(sum))
			selfVm.Cont.Pc++
		//case "=":
		case instr.OPCODE_EQUAL_NUM:
			argLen := instr.DeserializeEqualNumInstr(code)
			val := selfVm.Cont.Pop()
			var tmp reader.SExpression
			var result = true
			for i := int64(1); i < argLen; i++ {
				tmp = selfVm.Cont.Pop()
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
				selfVm.Cont.Push(reader.NewBool(true))
			} else {
				selfVm.Cont.Push(reader.NewBool(false))
			}
			selfVm.Cont.Pc++
		//case "!=":
		case instr.OPCODE_NOT_EQUAL_NUM:
			argLen := instr.DeserializeNotEqualNumInstr(code)
			val := selfVm.Cont.Pop()
			var tmp reader.SExpression
			var result = true
			for i := int64(1); i < argLen; i++ {
				tmp = selfVm.Cont.Pop()
				if result == false {
					continue
				}
				if val.Equals(tmp) {
					result = false
				}
			}
			if result {
				selfVm.Cont.Push(reader.NewBool(true))
			} else {
				selfVm.Cont.Push(reader.NewBool(false))
			}
			selfVm.Cont.Pc++
		//case ">":
		case instr.OPCODE_GREATER_THAN_NUM:
			argLen := instr.DeserializeGreaterThanNumInstr(code)
			val := selfVm.Cont.Pop().(reader.Number).GetValue()
			var tmp reader.SExpression
			flag := true
			for i := int64(1); i < argLen; i++ {
				tmp = selfVm.Cont.Pop()
				if flag == false {
					continue
				}
				if val >= tmp.(reader.Number).GetValue() {
					flag = false
				}
			}
			if flag {
				selfVm.Cont.Push(reader.NewBool(true))
			} else {
				selfVm.Cont.Push(reader.NewBool(false))
			}
			selfVm.Cont.Pc++
		//case "<":
		case instr.OPCODE_LESS_THAN_NUM:
			argLen := instr.DeserializeLessThanNumInstr(code)
			val := selfVm.Cont.Pop().(reader.Number).GetValue()
			var tmp reader.SExpression
			flag := true
			for i := int64(1); i < argLen; i++ {
				tmp = selfVm.Cont.Pop()
				if flag == false {
					continue
				}
				if val <= tmp.(reader.Number).GetValue() {
					flag = false
				}
			}
			if flag {
				selfVm.Cont.Push(reader.NewBool(true))
			} else {
				selfVm.Cont.Push(reader.NewBool(false))
			}
			selfVm.Cont.Pc++
		//case ">=":
		case instr.OPCODE_GREATER_THAN_OR_EQUAL_NUM:
			argLen := instr.DeserializeGreaterThanOrEqualNumInstr(code)
			val := selfVm.Cont.Pop().(reader.Number).GetValue()
			var tmp reader.SExpression
			flag := true
			for i := int64(1); i < argLen; i++ {
				tmp = selfVm.Cont.Pop()
				if flag == false {
					continue
				}
				if val > tmp.(reader.Number).GetValue() {
					flag = false
				}
			}
			if flag {
				selfVm.Cont.Push(reader.NewBool(true))
			} else {
				selfVm.Cont.Push(reader.NewBool(false))
			}
			selfVm.Cont.Pc++

		//case "<=":
		case instr.OPCODE_LESS_THAN_OR_EQUAL_NUM:
			argLen := instr.DeserializeLessThanOrEqualNumInstr(code)
			val := selfVm.Cont.Pop().(reader.Number).GetValue()
			var tmp reader.SExpression
			flag := true
			for i := int64(1); i < argLen; i++ {
				tmp = selfVm.Cont.Pop()
				if flag == false {
					continue
				}
				if val < tmp.(reader.Number).GetValue() {
					flag = false
				}
			}
			if flag {
				selfVm.Cont.Push(reader.NewBool(true))
			} else {
				selfVm.Cont.Push(reader.NewBool(false))
			}
			selfVm.Cont.Pc++
		//case "car":
		case instr.OPCODE_CAR:
			target := selfVm.Cont.Pop()
			if target.SExpressionTypeId() != reader.SExpressionTypeConsCell {
				fmt.Println("car target is not cons cell")
			}
			selfVm.Cont.Push(target.(reader.ConsCell).GetCar())
			selfVm.Cont.Pc++
		//case "cdr":
		case instr.OPCODE_CDR:
			target := selfVm.Cont.Pop()
			if target.SExpressionTypeId() != reader.SExpressionTypeConsCell {
				fmt.Println("cdr target is not cons cell")
			}
			selfVm.Cont.Push(target.(reader.ConsCell).GetCdr())
			selfVm.Cont.Pc++
		//case "random-id":
		case instr.OPCODE_RANDOM_ID:
			id := uuid.New()
			selfVm.Cont.Push(reader.NewString(id.String()))
			selfVm.Cont.Pc++
		case instr.OPCODE_NEW_ARRAY:
			selfVm.Cont.Push(reader.NewNativeArray(nil))
			selfVm.Cont.Pc++
		case instr.OPCODE_ARRAY_GET:
			if selfVm.Cont.Peek().SExpressionTypeId() != reader.SExpressionTypeNumber {
				fmt.Println("index is not number")
				goto ESCAPE
			}
			index := selfVm.Cont.Pop().(reader.Number).GetValue()
			if selfVm.Cont.Peek().SExpressionTypeId() != reader.SExpressionTypeNativeArray {
				fmt.Println("not an array")
				goto ESCAPE
			}
			target := selfVm.Cont.Pop().(*reader.NativeArray)
			selfVm.Cont.Push(target.Get(index))
			selfVm.Cont.Pc++
		case instr.OPCODE_ARRAY_SET:
			if selfVm.Cont.Peek().SExpressionTypeId() != reader.SExpressionTypeNumber {
				fmt.Println("elem is not number")
				goto ESCAPE
			}
			elem := selfVm.Cont.Pop()
			if selfVm.Cont.Peek().SExpressionTypeId() != reader.SExpressionTypeNumber {
				fmt.Println("index is not number")
				goto ESCAPE
			}
			index := selfVm.Cont.Pop().(reader.Number).GetValue()
			if selfVm.Cont.Peek().SExpressionTypeId() != reader.SExpressionTypeNativeArray {
				fmt.Println("not an array")
				goto ESCAPE
			}
			target := selfVm.Cont.Pop().(*reader.NativeArray)
			if err := target.Set(index, elem); err != nil {
				fmt.Println(err)
				goto ESCAPE
			}
			selfVm.Cont.Push(target)
			selfVm.Cont.Pc++
		case instr.OPCODE_ARRAY_LENGTH:
			if selfVm.Cont.Peek().SExpressionTypeId() != reader.SExpressionTypeNativeArray {
				fmt.Println("not an array")
				goto ESCAPE
			}
			targetRaw := selfVm.Cont.Pop()
			target := targetRaw.(*reader.NativeArray)
			selfVm.Cont.Push(reader.NewInt(target.Length()))
			selfVm.Cont.Pc++
		case instr.OPCODE_ARRAY_PUSH:
			elem := selfVm.Cont.Pop()
			if selfVm.Cont.Peek().SExpressionTypeId() != reader.SExpressionTypeNativeArray {
				fmt.Println("not an array")
				goto ESCAPE
			}
			targetRaw := selfVm.Cont.Pop()
			target := targetRaw.(*reader.NativeArray)
			target.Push(elem)
			selfVm.Cont.Push(target)
			selfVm.Cont.Pc++
		case instr.OPCODE_NEW_MAP:
			selfVm.Cont.Push(reader.NewNativeHashmap(nil))
			selfVm.Cont.Pc++
		case instr.OPCODE_MAP_GET:
			if selfVm.Cont.Peek().SExpressionTypeId() != reader.SExpressionTypeString {
				fmt.Println("key is not string")
				goto ESCAPE
			}
			key := selfVm.Cont.Pop()
			if selfVm.Cont.Peek().SExpressionTypeId() != reader.SExpressionTypeNativeHashmap {
				fmt.Println("not an hashmap")
				goto ESCAPE
			}
			target := selfVm.Cont.Pop().(*reader.NativeHashMap)
			selfVm.Cont.Push(target.Get(key.(reader.Str).GetValue()))
			selfVm.Cont.Pc++
		case instr.OPCODE_MAP_SET:
			val := selfVm.Cont.Pop()
			if selfVm.Cont.Peek().SExpressionTypeId() != reader.SExpressionTypeString {
				fmt.Println("key is not string")
				goto ESCAPE
			}
			key := selfVm.Cont.Pop()
			if selfVm.Cont.Peek().SExpressionTypeId() != reader.SExpressionTypeNativeHashmap {
				fmt.Println("not an hashmap")
				goto ESCAPE
			}
			target := selfVm.Cont.Pop().(*reader.NativeHashMap)
			target.Set(key.(reader.Str).GetValue(), val)
			selfVm.Cont.Push(target)
			selfVm.Cont.Pc++
		case instr.OPCODE_MAP_LENGTH:
			if selfVm.Cont.Peek().SExpressionTypeId() != reader.SExpressionTypeNativeHashmap {
				fmt.Println("not an hashmap")
				goto ESCAPE
			}
			target := selfVm.Cont.Pop().(*reader.NativeHashMap)
			selfVm.Cont.Push(reader.NewInt(target.Length()))
			selfVm.Cont.Pc++
		case instr.OPCODE_MAP_KEYS:
			if selfVm.Cont.Peek().SExpressionTypeId() != reader.SExpressionTypeNativeHashmap {
				fmt.Println("not an hashmap")
				goto ESCAPE
			}
			target := selfVm.Cont.Pop().(*reader.NativeHashMap)
			selfVm.Cont.Push(target)
			selfVm.Cont.Pc++
		case instr.OPCODE_MAP_DELETE:
			if selfVm.Cont.Peek().SExpressionTypeId() != reader.SExpressionTypeString {
				fmt.Println("key is not string")
				goto ESCAPE
			}
			key := selfVm.Cont.Pop()
			if selfVm.Cont.Peek().SExpressionTypeId() != reader.SExpressionTypeNativeHashmap {
				fmt.Println("not an hashmap")
				goto ESCAPE
			}
			target := selfVm.Cont.Pop().(*reader.NativeHashMap)
			target.Delete(key.(reader.Str).GetValue())
			selfVm.Cont.Push(target)
			selfVm.Cont.Pc++
		}
	}
ESCAPE:
	{
		cont := selfVm.Cont
		for {
			cont.Stack.Stack = []reader.SExpression{}
			cont.Code = []instr.Instr{}
			cont.Pc = 0
			callParent, err := cont.Stack.PeekIndex(0)
			if err != nil {
				break
			}
			if callParent.SExpressionTypeId() != reader.SExpressionTypeContinuation {
				break
			}
			cont = callParent.(*Continuation)
		}
	}
}
