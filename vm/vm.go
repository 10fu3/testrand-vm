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

type StackSexp struct {
	nil      reader.Nil
	num      reader.Number
	str      reader.Str
	sym      reader.Symbol
	boo      reader.Bool
	cell     reader.ConsCell
	arr      *reader.NativeArray
	hm       *reader.NativeHashMap
	closure  *Closure
	env      *Env
	sexpType reader.SExpressionType
}

func NewStackNil() *StackSexp {
	return &StackSexp{
		nil:      reader.NewNil(),
		sexpType: reader.SExpressionTypeNil,
	}
}

func NewStackNum(num reader.Number) *StackSexp {
	return &StackSexp{
		num:      num,
		sexpType: reader.SExpressionTypeNumber,
	}
}

func NewStackStr(str reader.Str) *StackSexp {
	return &StackSexp{
		str:      str,
		sexpType: reader.SExpressionTypeString,
	}
}

func NewStackSym(sym reader.Symbol) *StackSexp {
	return &StackSexp{
		sym:      sym,
		sexpType: reader.SExpressionTypeSymbol,
	}
}

func NewStackBool(boo reader.Bool) *StackSexp {
	return &StackSexp{
		boo:      boo,
		sexpType: reader.SExpressionTypeBool,
	}
}

func NewStackCell(cell reader.ConsCell) *StackSexp {
	return &StackSexp{
		cell:     cell,
		sexpType: reader.SExpressionTypeConsCell,
	}
}

func NewStackArr(arr *reader.NativeArray) *StackSexp {
	return &StackSexp{
		arr:      arr,
		sexpType: reader.SExpressionTypeNativeArray,
	}
}

func NewStackHashMap(hm *reader.NativeHashMap) *StackSexp {
	return &StackSexp{
		hm:       hm,
		sexpType: reader.SExpressionTypeNativeHashmap,
	}
}

func NewStackClosure(closure *Closure) *StackSexp {
	return &StackSexp{
		closure:  closure,
		sexpType: reader.SExpressionTypeClosure,
	}
}

func NewStackEnv(env *Env) *StackSexp {
	return &StackSexp{
		env:      env,
		sexpType: reader.SExpressionTypeEnvironment,
	}
}

func (s *StackSexp) ToSExpression() reader.SExpression {
	switch s.sexpType {
	case reader.SExpressionTypeNil:
		return s.nil
	case reader.SExpressionTypeNumber:
		return s.num
	case reader.SExpressionTypeString:
		return s.str
	case reader.SExpressionTypeSymbol:
		return s.sym
	case reader.SExpressionTypeBool:
		return s.boo
	case reader.SExpressionTypeConsCell:
		return s.cell
	case reader.SExpressionTypeNativeArray:
		return s.arr
	case reader.SExpressionTypeNativeHashmap:
		return s.hm
	case reader.SExpressionTypeClosure:
		return s.closure
	case reader.SExpressionTypeEnvironment:
		return &Env{}
	default:
		panic("unknown type")
	}
}

type Closure struct {
	Mutex         *sync.RWMutex
	Stack         []*StackSexp
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
	stack := make([]*StackSexp, len(vm.Stack))
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
		Stack: make([]*StackSexp, 0),
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
			selfVm.Push(NewStackNil())
			selfVm.Pc++
		//case "push-sym":
		case instr.OPCODE_PUSH_SYM:
			//selfVm.Push(reader.NewSymbol(opCodeAndArgs[1]))
			selfVm.Push(NewStackSym(instr.DeserializePushSymbolInstr(code)))
			selfVm.Pc++

		//case "push-num":
		case instr.OPCODE_PUSH_NUM:
			selfVm.Push(NewStackNum(reader.NewInt(instr.DeserializePushNumberInstr(code))))
			selfVm.Pc++
		//case "push-boo":
		case instr.OPCODE_PUSH_TRUE:
			//val := opCodeAndArgs[1] == "#t"
			//
			//if opCodeAndArgs[1] != "#f" && opCodeAndArgs[1] != "#t" {
			//	fmt.Println("not a bool")
			//	goto ESCAPE
			//}
			selfVm.Push(NewStackBool(reader.NewBool(true)))
			selfVm.Pc++
		case instr.OPCODE_PUSH_FALSE:
			selfVm.Push(NewStackBool(reader.NewBool(false)))
			selfVm.Pc++
		//case "push-str":
		case instr.OPCODE_PUSH_STR:
			//selfVm.Push(reader.NewString(opCodeAndArgs[1]))
			selfVm.Push(NewStackStr(reader.NewString(instr.DeserializePushStringInstr(code))))
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
			if val.sexpType != reader.SExpressionTypeBool {
				fmt.Println("not a bool")
				goto ESCAPE
			}
			if val.boo.GetValue() {
				selfVm.Pc = jumpTo
			} else {
				selfVm.Pc++
			}
		//case "jmp-else":
		case instr.OPCODE_JMP_ELSE:
			//jumpTo, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			jumpTo := instr.DeserializeJmpElseInstr(code)
			val := selfVm.Pop()
			if val.sexpType != reader.SExpressionTypeBool {
				fmt.Println("not a bool")
				goto ESCAPE
			}

			if !val.boo.GetValue() {
				selfVm.Pc = jumpTo
			} else {
				selfVm.Pc++
			}

		//case "load":
		case instr.OPCODE_LOAD:
			sym := selfVm.Pop().sym

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
				selfVm.PushSexp(*meVm.Env.Frame[sym.GetValue()])
				selfVm.Pc++
			} else {
				fmt.Println("Symbol not found: " + sym.GetValue())
				goto ESCAPE
			}
		//case "define":
		case instr.OPCODE_DEFINE:
			//sym := reader.NewSymbol(opCodeAndArgs[1])
			deserialize := instr.DeserializeDefineInstr(code)
			val := selfVm.Pop().ToSExpression()
			//selfVm.Env.Frame[opCodeAndArgs[1]] = &val
			selfVm.Env.Frame[deserialize] = &val
			//selfVm.Push(sym)
			selfVm.PushSym(reader.NewSymbol(deserialize))
			selfVm.Pc++
		//case "define-args":
		case instr.OPCODE_DEFINE_ARGS:
			//sym := reader.NewSymbol(opCodeAndArgs[1])
			deserialize := instr.DeserializeDefineArgsInstr(code)
			selfVm.PushSym(reader.NewSymbol(deserialize))
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
			selfVm.PushSexp(deserialize)
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
			sexp := val.ToSExpression()
			thisVm.Env.Frame[deserialize] = &sexp
			selfVm.Push(val)
			selfVm.Pc++
		//case "new-env":
		case instr.OPCODE_NEW_ENV:
			env := &Env{
				Frame: make(map[string]*reader.SExpression),
			}
			selfVm.PushEnv(env)
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
				sym := selfVm.Pop().sym
				newVm.TemporaryArgs = append(newVm.TemporaryArgs, sym)
			}

			newVm.Cont = selfVm
			newVm.Env = selfVm.Pop().env
			newVm.Pc = 0
			selfVm.PushClosure(newVm)
			selfVm.Pc++
		//case "call":
		case instr.OPCODE_CALL:
			rawClosure := selfVm.Pop()

			if rawClosure.sexpType != reader.SExpressionTypeClosure {
				fmt.Println("not a closure")
				goto ESCAPE
			}

			nextVm := rawClosure.closure
			env := nextVm.Env

			argsSize := instr.DeserializeCallInstr(code)

			if argsSize != int64(len(nextVm.TemporaryArgs)) {
				fmt.Println("args size not match")
				goto ESCAPE
			}

			for _, sym := range nextVm.TemporaryArgs {
				val := selfVm.Pop().ToSExpression()
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
			selfVm.Stack = []*StackSexp{}
			selfVm.Pc = 0
			selfVm = selfVm.ReturnCont
			selfVm.Pc = retPc
			selfVm.Push(val)
			selfVm.Pc++
		//case "and":
		case instr.OPCODE_AND:
			//argsSize, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			argsSize := instr.DeserializeAndInstr(code)
			val := selfVm.Pop().boo.GetValue()
			var tmp reader.SExpression
			flag := true
			for i := int64(1); i < argsSize; i++ {
				if selfVm.Peek().sexpType != reader.SExpressionTypeBool {
					fmt.Println("not a bool")
					goto ESCAPE
				}
				tmp = selfVm.Pop().boo
				if flag == false {
					continue
				}
				if val != tmp.(reader.Bool).GetValue() {
					flag = false
				}
			}
			if flag {
				selfVm.PushBool(reader.NewBool(true))
			} else {
				selfVm.PushBool(reader.NewBool(false))
			}
			selfVm.Pc++
		//case "or":
		case instr.OPCODE_OR:
			//argsSize, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			argsSize := instr.DeserializeOrInstr(code)
			var tmp = selfVm.Pop().boo
			flag := false
			for i := int64(0); i < argsSize; i++ {
				if tmp.(reader.Bool).GetValue() {
					flag = true
				}
				if flag == true {
					continue
				}
				tmp = selfVm.Pop().boo
			}
			if flag {
				selfVm.PushBool(reader.NewBool(true))
			} else {
				selfVm.PushBool(reader.NewBool(false))
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
				line += selfVm.Pop().str.GetValue()
			}
			fmt.Print(line)
			selfVm.PushNil()
			selfVm.Pc++
		//case "println":
		case instr.OPCODE_PRINTLN:
			argLen := instr.DeserializePrintlnInstr(code)
			line := ""
			for i := int64(0); i < argLen; i++ {
				line += selfVm.Pop().str.GetValue()
			}
			fmt.Println(line)
			selfVm.PushNil()
			selfVm.Pc++
		//case "+":
		case instr.OPCODE_PLUS_NUM:
			argLen := instr.DeserializePlusNumInstr(code)
			sum := int64(0)
			for i := int64(0); i < argLen; i++ {
				if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
					fmt.Println("not a number")
					goto ESCAPE
				}
				sum += selfVm.Pop().num.GetValue()
			}
			selfVm.PushNum(reader.NewInt(sum))
			selfVm.Pc++
		//case "-":
		case instr.OPCODE_MINUS_NUM:
			argLen := instr.DeserializeMinusNumInstr(code)
			if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
				fmt.Println("not a number")
				goto ESCAPE
			}
			sum := selfVm.Pop().num.GetValue()
			for i := int64(1); i < argLen; i++ {
				if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
					fmt.Println("not a number")
					goto ESCAPE
				}
				sum -= selfVm.Pop().num.GetValue()
			}
			selfVm.PushNum(reader.NewInt(sum))
			selfVm.Pc++
		//case "*":
		case instr.OPCODE_MULTIPLY_NUM:
			argLen := instr.DeserializeMultiplyNumInstr(code)
			sum := int64(1)
			for i := int64(0); i < argLen; i++ {
				if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
					fmt.Println("not a number")
					goto ESCAPE
				}
				sum *= selfVm.Pop().num.GetValue()
			}
			selfVm.PushNum(reader.NewInt(sum))
			selfVm.Pc++
		//case "/":
		case instr.OPCODE_DIVIDE_NUM:
			argLen := instr.DeserializeDivideNumInstr(code)
			if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
				fmt.Println("not a number")
				goto ESCAPE
			}
			sum := selfVm.Pop().num.GetValue()
			for i := int64(1); i < argLen; i++ {
				if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
					fmt.Println("not a number")
					goto ESCAPE
				}
				sum /= selfVm.Pop().num.GetValue()
			}
			selfVm.PushNum(reader.NewInt(sum))
			selfVm.Pc++
		//case "mod":
		case instr.OPCODE_MODULO_NUM:
			argLen := instr.DeserializeModuloNumInstr(code)
			if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
				fmt.Println("not a number")
				goto ESCAPE
			}
			sum := selfVm.Pop().num.GetValue()
			for i := int64(1); i < argLen; i++ {
				if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
					fmt.Println("not a number")
					goto ESCAPE
				}
				sum %= selfVm.Pop().num.GetValue()
			}
			selfVm.PushNum(reader.NewInt(sum))
			selfVm.Pc++
		//case "=":
		case instr.OPCODE_EQUAL_NUM:
			argLen := instr.DeserializeEqualNumInstr(code)
			val := selfVm.Pop()
			var tmp reader.Number
			var result = true
			for i := int64(1); i < argLen; i++ {
				if result == false {
					continue
				}
				if reader.SExpressionTypeNumber != selfVm.Peek().sexpType {
					fmt.Println("arg is not number")
				}
				tmp = selfVm.Pop().num
				if val.num.GetValue() != tmp.(reader.Number).GetValue() {
					result = false
				}
			}
			if result {
				selfVm.PushBool(reader.NewBool(true))
			} else {
				selfVm.PushBool(reader.NewBool(false))
			}
			selfVm.Pc++
		//case "!=":
		case instr.OPCODE_NOT_EQUAL_NUM:
			argLen := instr.DeserializeNotEqualNumInstr(code)
			val := selfVm.Pop()
			var tmp reader.Bool
			var result = true
			for i := int64(1); i < argLen; i++ {
				if result == false {
					continue
				}
				tmp = selfVm.Pop().boo
				if val.boo.GetValue() != tmp.GetValue() {
					result = false
				}
			}
			if result {
				selfVm.PushBool(reader.NewBool(true))
			} else {
				selfVm.PushBool(reader.NewBool(false))
			}
			selfVm.Pc++
		//case ">":
		case instr.OPCODE_GREATER_THAN_NUM:
			argLen := instr.DeserializeGreaterThanNumInstr(code)
			if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
				fmt.Println("not a number")
				goto ESCAPE
			}
			val := selfVm.Pop().num.GetValue()
			var tmp reader.Number
			flag := true
			for i := int64(1); i < argLen; i++ {
				if flag == false {
					continue
				}
				if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
					fmt.Println("not a number")
					goto ESCAPE
				}
				tmp = selfVm.Pop().num
				if val >= tmp.GetValue() {
					flag = false
				}
			}
			if flag {
				selfVm.PushBool(reader.NewBool(true))
			} else {
				selfVm.PushBool(reader.NewBool(false))
			}
			selfVm.Pc++
		//case "<":
		case instr.OPCODE_LESS_THAN_NUM:
			argLen := instr.DeserializeLessThanNumInstr(code)

			if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
				fmt.Println("not a number")
				goto ESCAPE
			}

			val := selfVm.Pop().num.GetValue()
			var tmp reader.Number
			flag := true
			for i := int64(1); i < argLen; i++ {
				if flag == false {
					continue
				}
				if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
					fmt.Println("not a number")
					goto ESCAPE
				}
				tmp = selfVm.Pop().num
				if val <= tmp.GetValue() {
					flag = false
				}
			}
			if flag {
				selfVm.PushBool(reader.NewBool(true))
			} else {
				selfVm.PushBool(reader.NewBool(false))
			}
			selfVm.Pc++
		//case ">=":
		case instr.OPCODE_GREATER_THAN_OR_EQUAL_NUM:
			argLen := instr.DeserializeGreaterThanOrEqualNumInstr(code)
			if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
				fmt.Println("not a number")
				goto ESCAPE
			}
			val := selfVm.Pop().num.GetValue()
			var tmp reader.Number
			flag := true
			for i := int64(1); i < argLen; i++ {
				if flag == false {
					continue
				}
				tmp = selfVm.Pop().num
				if val > tmp.(reader.Number).GetValue() {
					flag = false
				}
			}
			if flag {
				selfVm.PushBool(reader.NewBool(true))
			} else {
				selfVm.PushBool(reader.NewBool(false))
			}
			selfVm.Pc++

		//case "<=":
		case instr.OPCODE_LESS_THAN_OR_EQUAL_NUM:
			argLen := instr.DeserializeLessThanOrEqualNumInstr(code)
			if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
				fmt.Println("not a number")
				goto ESCAPE
			}
			val := selfVm.Pop().num.GetValue()
			var tmp reader.Number
			flag := true
			for i := int64(1); i < argLen; i++ {
				if flag == false {
					continue
				}
				if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
					fmt.Println("not a number")
					goto ESCAPE
				}
				tmp = selfVm.Pop().num
				if val < tmp.(reader.Number).GetValue() {
					flag = false
				}
			}
			if flag {
				selfVm.PushBool(reader.NewBool(true))
			} else {
				selfVm.PushBool(reader.NewBool(false))
			}
			selfVm.Pc++
		//case "car":
		case instr.OPCODE_CAR:
			if selfVm.Peek().sexpType != reader.SExpressionTypeConsCell {
				fmt.Println("car target is not cons cell")
				goto ESCAPE
			}
			target := selfVm.Pop().cell
			selfVm.PushSexp(target.GetCar())
			selfVm.Pc++
		//case "cdr":
		case instr.OPCODE_CDR:
			if selfVm.Peek().sexpType != reader.SExpressionTypeConsCell {
				fmt.Println("cdr target is not cons cell")
				goto ESCAPE
			}
			target := selfVm.Pop().cell
			selfVm.PushSexp(target.GetCdr())
			selfVm.Pc++
		//case "random-id":
		case instr.OPCODE_RANDOM_ID:
			id := uuid.New()
			selfVm.PushStr(reader.NewString(id.String()))
			selfVm.Pc++
		case instr.OPCODE_NEW_ARRAY:
			selfVm.PushArr(reader.NewNativeArray(nil))
			selfVm.Pc++
		case instr.OPCODE_ARRAY_GET:
			if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
				fmt.Println("index is not number")
				goto ESCAPE
			}
			index := selfVm.Pop().num.GetValue()
			if selfVm.Peek().sexpType != reader.SExpressionTypeNativeArray {
				fmt.Println("not an array")
				goto ESCAPE
			}
			target := selfVm.Pop().arr
			selfVm.PushSexp(target.Get(index))
			selfVm.Pc++
		case instr.OPCODE_ARRAY_SET:
			if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
				fmt.Println("elem is not number")
				goto ESCAPE
			}
			elem := selfVm.Pop()
			if selfVm.Peek().sexpType != reader.SExpressionTypeNumber {
				fmt.Println("index is not number")
				goto ESCAPE
			}
			index := selfVm.Pop().num.GetValue()
			if selfVm.Peek().sexpType != reader.SExpressionTypeNativeArray {
				fmt.Println("not an array")
				goto ESCAPE
			}
			target := selfVm.Pop().arr
			if err := target.Set(index, elem.ToSExpression()); err != nil {
				fmt.Println(err)
				goto ESCAPE
			}
			selfVm.PushArr(target)
			selfVm.Pc++
		case instr.OPCODE_ARRAY_LENGTH:
			if selfVm.Peek().sexpType != reader.SExpressionTypeNativeArray {
				fmt.Println("not an array")
				goto ESCAPE
			}
			target := selfVm.Pop().arr
			selfVm.PushNum(reader.NewInt(target.Length()))
			selfVm.Pc++
		case instr.OPCODE_ARRAY_PUSH:
			elem := selfVm.Pop()
			if selfVm.Peek().sexpType != reader.SExpressionTypeNativeArray {
				fmt.Println("not an array")
				goto ESCAPE
			}
			targetRaw := selfVm.Pop()
			target := targetRaw.arr
			target.Push(elem.ToSExpression())
			selfVm.PushArr(target)
			selfVm.Pc++
		case instr.OPCODE_NEW_MAP:
			selfVm.PushHashMap(reader.NewNativeHashmap(nil))
			selfVm.Pc++
		case instr.OPCODE_MAP_GET:
			if selfVm.Peek().sexpType != reader.SExpressionTypeString {
				fmt.Println("key is not string")
				goto ESCAPE
			}
			key := selfVm.Pop().str
			if selfVm.Peek().sexpType != reader.SExpressionTypeNativeHashmap {
				fmt.Println("not an hashmap")
				goto ESCAPE
			}
			target := selfVm.Pop().hm
			selfVm.PushSexp(target.Get(key.GetValue()))
			selfVm.Pc++
		case instr.OPCODE_MAP_SET:
			val := selfVm.Pop()
			if selfVm.Peek().sexpType != reader.SExpressionTypeString {
				fmt.Println("key is not string")
				goto ESCAPE
			}
			key := selfVm.Pop().str
			if selfVm.Peek().sexpType != reader.SExpressionTypeNativeHashmap {
				fmt.Println("not an hashmap")
				goto ESCAPE
			}
			target := selfVm.Pop().hm
			target.Set(key.GetValue(), val.ToSExpression())
			selfVm.PushHashMap(target)
			selfVm.Pc++
		case instr.OPCODE_MAP_LENGTH:
			if selfVm.Peek().sexpType != reader.SExpressionTypeNativeHashmap {
				fmt.Println("not an hashmap")
				goto ESCAPE
			}
			target := selfVm.Pop().hm
			selfVm.PushNum(reader.NewInt(target.Length()))
			selfVm.Pc++
		case instr.OPCODE_MAP_KEYS:
			if selfVm.Peek().sexpType != reader.SExpressionTypeNativeHashmap {
				fmt.Println("not an hashmap")
				goto ESCAPE
			}
			target := selfVm.Pop().hm
			selfVm.PushHashMap(target)
			selfVm.Pc++
		case instr.OPCODE_MAP_DELETE:
			if selfVm.Peek().sexpType != reader.SExpressionTypeString {
				fmt.Println("key is not string")
				goto ESCAPE
			}
			key := selfVm.Pop().str
			if selfVm.Peek().sexpType != reader.SExpressionTypeNativeHashmap {
				fmt.Println("not an hashmap")
				goto ESCAPE
			}
			target := selfVm.Pop().hm
			target.Delete(key.GetValue())
			selfVm.PushHashMap(target)
			selfVm.Pc++
		}
	}
ESCAPE:
	{
		for {
			selfVm.Stack = []*StackSexp{}
			selfVm.Code = []instr.Instr{}
			selfVm.Pc = 0
			if selfVm.ReturnCont == nil {
				break
			}
			selfVm = selfVm.ReturnCont
		}
	}
}

func (vm *Closure) Push(sexp *StackSexp) {
	vm.Stack = append(vm.Stack, sexp)
}

func (vm *Closure) PushSexp(sexp reader.SExpression) {
	switch sexp.SExpressionTypeId() {
	case reader.SExpressionTypeNil:
		vm.Push(NewStackNil())
	case reader.SExpressionTypeBool:
		vm.Push(NewStackBool(sexp.(reader.Bool)))
	case reader.SExpressionTypeNumber:
		vm.Push(NewStackNum(sexp.(reader.Number)))
	case reader.SExpressionTypeString:
		vm.Push(NewStackStr(sexp.(reader.Str)))
	case reader.SExpressionTypeSymbol:
		vm.Push(NewStackSym(sexp.(reader.Symbol)))
	case reader.SExpressionTypeConsCell:
		vm.Push(NewStackCell(sexp.(reader.ConsCell)))
	case reader.SExpressionTypeClosure:
		vm.Push(NewStackClosure(sexp.(*Closure)))
	case reader.SExpressionTypeNativeArray:
		vm.Push(NewStackArr(sexp.(*reader.NativeArray)))
	case reader.SExpressionTypeNativeHashmap:
		vm.Push(NewStackHashMap(sexp.(*reader.NativeHashMap)))
	case reader.SExpressionTypeEnvironment:
		vm.Push(NewStackEnv(sexp.(*Env)))
	default:
		panic("not implemented")
	}
}

func (vm *Closure) PushNil() {
	vm.Push(NewStackNil())
}

func (vm *Closure) PushBool(boo reader.Bool) {
	vm.Push(NewStackBool(boo))
}

func (vm *Closure) PushNum(num reader.Number) {
	vm.Push(NewStackNum(num))
}

func (vm *Closure) PushStr(str reader.Str) {
	vm.Push(NewStackStr(str))
}

func (vm *Closure) PushSym(sym reader.Symbol) {
	vm.Push(NewStackSym(sym))
}

func (vm *Closure) PushCell(cell reader.ConsCell) {
	vm.Push(NewStackCell(cell))
}

func (vm *Closure) PushClosure(closure *Closure) {
	vm.Push(NewStackClosure(closure))
}

func (vm *Closure) PushArr(arr *reader.NativeArray) {
	vm.Push(NewStackArr(arr))
}

func (vm *Closure) PushHashMap(hashMap *reader.NativeHashMap) {
	vm.Push(NewStackHashMap(hashMap))
}

func (vm *Closure) PushEnv(env *Env) {
	vm.Push(NewStackEnv(env))
}

func (vm *Closure) Pop() *StackSexp {
	if len(vm.Stack) == 0 {
		return nil
	}

	ret := vm.Stack[len(vm.Stack)-1]
	vm.Stack = vm.Stack[:len(vm.Stack)-1]
	return ret
}

func (vm *Closure) Peek() *StackSexp {
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
