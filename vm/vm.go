package vm

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"os"
	"strings"
	"sync/atomic"
	"testrand-vm/compile"
	"time"
)

type Env struct {
	SelfIndex uint64
	Frame     map[uint64]*compile.SExpression
	IsLocked  uint32
}

func (e *Env) TypeId() string {
	return "environment"
}

func (e *Env) SExpressionTypeId() compile.SExpressionType {
	return compile.SExpressionTypeEnvironment
}

func (e *Env) String(env *compile.CompilerEnvironment) string {
	return "environment"
}

func (e *Env) IsList() bool {
	return false
}

func (e *Env) Equals(sexp compile.SExpression) bool {
	panic("implement me")
}

type Closure struct {
	GlobalEnvId   string
	EnvId         uint64
	CompilerEnv   *compile.CompilerEnvironment
	Stack         SexpStack
	Code          []compile.Instr
	Pc            int64
	Cont          *Closure
	ReturnCont    *Closure
	ReturnPc      int64
	TemporaryArgs []compile.Symbol
	Result        compile.SExpression
}

type SexpStack struct {
	stack []compile.SExpression
	Size  int
}

func (stk *SexpStack) Push(data compile.SExpression) {

	if stk.Size < len(stk.stack) {
		stk.stack[stk.Size] = data
		stk.Size++
		return
	}

	stk.stack = append(stk.stack, data, nil, nil, nil, nil, nil)
	stk.Size++
}

func (stk *SexpStack) Pop() compile.SExpression {
	if stk.Size == 0 {
		return nil
	}

	r := stk.stack[stk.Size-1]

	if len(stk.stack)/2 > stk.Size && len(stk.stack) > 31 {
		stk.stack = stk.stack[:(len(stk.stack) * 3 / 4)]
	}

	stk.Size--
	return r
}

func (stk *SexpStack) Peek() compile.SExpression {
	if stk.Size == 0 {
		return nil
	}
	return stk.stack[stk.Size-1]
}

func NewSexpStack() SexpStack {
	return SexpStack{
		stack: make([]compile.SExpression, 0, 6),
	}
}

func (vm *Closure) TypeId() string {
	return "closure"
}

func (vm *Closure) SExpressionTypeId() compile.SExpressionType {
	return compile.SExpressionTypeClosure
}

func (vm *Closure) String(compEnv *compile.CompilerEnvironment) string {
	return "closure"
}

func (vm *Closure) IsList() bool {
	return false
}

func (vm *Closure) Equals(sexp compile.SExpression) bool {
	//TODO implement me
	panic("implement me")
}

var globalEnv = map[uint64]Env{
	0: {
		SelfIndex: 0,
		Frame:     map[uint64]*compile.SExpression{},
	},
}

var globalEnvMutex = uint32(0)

func AtomicControlEnv(f func(map[uint64]Env) error) error {
	for !atomic.CompareAndSwapUint32(&globalEnvMutex, 0, 1) {
	}
	defer atomic.StoreUint32(&globalEnvMutex, 0)
	return f(globalEnv)
}

func NewVM(compEnv *compile.CompilerEnvironment) *Closure {
	return &Closure{
		CompilerEnv: compEnv,
		Stack:       NewSexpStack(),
		Pc:          0,
		Cont:        nil,
	}
}

func NewVMWithGlobalEnvId(compEnv *compile.CompilerEnvironment, globalEnvId string) *Closure {
	return &Closure{
		GlobalEnvId: globalEnvId,
		CompilerEnv: compEnv,
		Stack:       NewSexpStack(),
		Pc:          0,
		Cont:        nil,
	}
}

func VMRunFromEntryPoint(vm *Closure) {
	vm.Pc = 0
	vm.Code = vm.CompilerEnv.GetInstr()
	VMRun(vm)
	vm.Code = nil
}

func VMRun(vm *Closure) compile.SExpression {

	selfVm := vm

	for {

		//rawCode := selfVm.Code[selfVm.Pc].(reader.Symbol).GetSymbolIndex()
		code := selfVm.Code[selfVm.Pc]
		//switch opCodeAndArgs[0] {
		switch code.Type {
		case compile.OPCODE_PUSH_NIL:
			selfVm.Push(compile.NewNil())
			selfVm.Pc++
		//case "push-sym":
		case compile.OPCODE_PUSH_SYM:
			//selfVm.Push(reader.NewSymbol(opCodeAndArgs[1]))
			selfVm.Push(compile.DeserializePushSymbolInstr(code))
			selfVm.Pc++

		//case "push-num":
		case compile.OPCODE_PUSH_NUM:
			selfVm.Push(compile.Number(compile.DeserializePushNumberInstr(vm.CompilerEnv, code)))
			selfVm.Pc++
		//case "push-boo":
		case compile.OPCODE_PUSH_TRUE:
			//val := opCodeAndArgs[1] == "#t"
			//
			//if opCodeAndArgs[1] != "#f" && opCodeAndArgs[1] != "#t" {
			//	fmt.Println("not a bool")
			//	goto ESCAPE
			//}
			selfVm.Push(compile.Bool(true))
			selfVm.Pc++
		case compile.OPCODE_PUSH_FALSE:
			selfVm.Push(compile.Bool(false))
			selfVm.Pc++
		//case "push-str":
		case compile.OPCODE_PUSH_STR:
			//selfVm.Push(reader.NewString(opCodeAndArgs[1]))
			selfVm.Push(compile.DeserializePushStringInstr(vm.CompilerEnv, code))
			selfVm.Pc++
		//case "pop":
		case compile.OPCODE_POP:
			selfVm.Pop()
			selfVm.Pc++
		//case "jmp":
		case compile.OPCODE_JMP:
			//jumpTo, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			jumpTo := compile.DeserializeJmpInstr(vm.CompilerEnv, code)
			selfVm.Pc = jumpTo
		//case "jmp-if":
		case compile.OPCODE_JMP_IF:
			//jumpTo, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			jumpTo := compile.DeserializeJmpIfInstr(vm.CompilerEnv, code)
			val := selfVm.Pop()
			if val.SExpressionTypeId() != compile.SExpressionTypeBool {
				fmt.Println("not a bool")
				goto ESCAPE
			}
			if val.(compile.Bool) {
				selfVm.Pc = jumpTo
			} else {
				selfVm.Pc++
			}
		//case "jmp-else":
		case compile.OPCODE_JMP_ELSE:
			//jumpTo, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			jumpTo := compile.DeserializeJmpElseInstr(vm.CompilerEnv, code)
			val := selfVm.Pop()
			if val.SExpressionTypeId() != compile.SExpressionTypeBool {
				fmt.Println("not a bool")
				goto ESCAPE
			}

			if !val.(compile.Bool) {
				selfVm.Pc = jumpTo
			} else {
				selfVm.Pc++
			}

		//case "load":
		case compile.OPCODE_LOAD:

			envId, symId := compile.DeserializeLoadInstr(vm.CompilerEnv, code)

			err := AtomicControlEnv(func(env map[uint64]Env) error {
				loaded, ok := env[envId]
				if !ok {
					return fmt.Errorf("not found")
				}
				selfVm.Push(*loaded.Frame[symId])
				return nil
			})

			if err != nil {
				fmt.Println(err)
				goto ESCAPE
			}

			selfVm.Pc++

		//case "define":
		case compile.OPCODE_DEFINE:
			//sym := reader.NewSymbol(opCodeAndArgs[1])
			envId, symIndexId := compile.DeserializeDefineInstr(vm.CompilerEnv, code)
			val := selfVm.Pop()

			AtomicControlEnv(func(env map[uint64]Env) error {
				env[envId].Frame[symIndexId] = &val
				return nil
			})
			symId, err := vm.CompilerEnv.FindSymbolInEnvironment(envId, symIndexId)
			if err != nil {
				fmt.Println(err)
				goto ESCAPE
			}
			selfVm.Push(compile.NewSymbol(symId))
			selfVm.Pc++
		//case "define-args":
		case compile.OPCODE_DEFINE_ARGS:
			//sym := reader.NewSymbol(opCodeAndArgs[1])
			_, symId := compile.DeserializeDefineArgsInstr(vm.CompilerEnv, code)
			selfVm.Push(compile.NewSymbol(symId))
			selfVm.Pc++
		//case "load-sexp":
		case compile.OPCODE_PUSH_SEXP:
			//r := bufio.NewReader(strings.NewReader(opCodeAndArgs[1]))
			//sexp, err := reader.NewReader(r).Read()
			//if err != nil {
			//	panic(err)
			//}
			deserialize, err := compile.DeserializeSexpressionInstr(vm.CompilerEnv, code)

			if err != nil {
				panic(err)
			}
			selfVm.Push(deserialize)
			selfVm.Pc++
		//case "set":
		case compile.OPCODE_SET:
			//sym := reader.NewSymbol(opCodeAndArgs[1])
			envId, symId := compile.DeserializeSetInstr(vm.CompilerEnv, code)
			val := selfVm.Pop()
			AtomicControlEnv(func(env map[uint64]Env) error {
				env[envId].Frame[symId] = &val
				return nil
			})
			selfVm.Push(val)
			selfVm.Pc++
		//case "new-env":
		case compile.OPCODE_NEW_ENV:

			_, envId := compile.DeserializeNewEnvInstr(vm.CompilerEnv, code)

			newEnv := Env{
				SelfIndex: envId,
				Frame:     make(map[uint64]*compile.SExpression),
				IsLocked:  0,
			}

			AtomicControlEnv(func(env map[uint64]Env) error {
				env[envId] = newEnv
				return nil
			})

			selfVm.Push(&newEnv)
			selfVm.Pc++
		//case "create-lambda":
		case compile.OPCODE_CREATE_CLOSURE:
			//argsSizeAndCodeLen := strings.SplitN(opCodeAndArgs[1], " ", 2)
			//argsSize, _ := strconv.ParseInt(argsSizeAndCodeLen[0], 10, 64)
			//codeLen, _ := strconv.ParseInt(argsSizeAndCodeLen[1], 10, 64)

			argsSize, codeLen := compile.DeserializeCreateClosureInstr(vm.CompilerEnv, code)

			pc := selfVm.Pc

			newVm := NewVM(vm.CompilerEnv)

			for i := int64(1); i <= codeLen; i++ {
				newVm.Code = append(newVm.Code, selfVm.Code[pc+i])
				selfVm.Pc++
			}

			for i := int64(0); i < argsSize; i++ {
				sym := selfVm.Pop().(compile.Symbol)
				newVm.TemporaryArgs = append(newVm.TemporaryArgs, sym)
			}

			newVm.Cont = selfVm
			newVm.EnvId = selfVm.Pop().(*Env).SelfIndex
			newVm.Pc = 0
			selfVm.Push(newVm)
			selfVm.Pc++
		//case "call":
		case compile.OPCODE_CALL:
			rawClosure := selfVm.Pop()

			if rawClosure.SExpressionTypeId() != compile.SExpressionTypeClosure {
				fmt.Println("not a closure")
				goto ESCAPE
			}

			nextVm := rawClosure.(*Closure)
			nextEnvId := nextVm.EnvId

			if err := AtomicControlEnv(func(env map[uint64]Env) error {
				newEnv, ok := env[nextEnvId]

				if !ok {
					return errors.New("not found")
				}

				argsSize := compile.DeserializeCallInstr(vm.CompilerEnv, code)

				if argsSize != int64(len(nextVm.TemporaryArgs)) {
					return errors.New("args size not match")
				}
				for _, sym := range nextVm.TemporaryArgs {
					val := selfVm.Pop()
					newEnv.Frame[sym.GetSymbolIndex()] = &val
				}
				return nil
			}); err != nil {
				fmt.Println(err)
				goto ESCAPE
			}

			nextVm.ReturnCont = selfVm
			nextVm.ReturnPc = selfVm.Pc
			selfVm = nextVm
		//case "ret":
		case compile.OPCODE_RETURN:
			val := selfVm.Pop()
			retPc := selfVm.ReturnPc
			selfVm.Stack = NewSexpStack()
			selfVm.Pc = 0
			selfVm = selfVm.ReturnCont
			selfVm.Pc = retPc
			selfVm.Push(val)
			selfVm.Pc++
		//case "and":
		case compile.OPCODE_AND:
			//argsSize, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			argsSize := compile.DeserializeAndInstr(vm.CompilerEnv, code)
			val := selfVm.Pop().(compile.Bool)
			var tmp compile.SExpression
			flag := true
			for i := int64(1); i < argsSize; i++ {
				tmp = selfVm.Pop()
				if flag == false {
					continue
				}
				if val != tmp.(compile.Bool) {
					flag = false
				}
			}
			if flag {
				selfVm.Push(compile.Bool(true))
			} else {
				selfVm.Push(compile.Bool(false))
			}
			selfVm.Pc++
		//case "or":
		case compile.OPCODE_OR:
			//argsSize, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			argsSize := compile.DeserializeOrInstr(vm.CompilerEnv, code)
			var tmp compile.SExpression = selfVm.Pop()
			flag := false
			for i := int64(0); i < argsSize; i++ {
				if tmp.(compile.Bool) {
					flag = true
				}
				if flag == true {
					continue
				}
				tmp = selfVm.Pop()
			}
			if flag {
				selfVm.Push(compile.Bool(true))
			} else {
				selfVm.Push(compile.Bool(false))
			}
			selfVm.Pc++
		//case "end-code":
		case compile.OPCODE_END_CODE:
			val := selfVm.Pop()
			disp := val.String(vm.CompilerEnv)
			fmt.Println(disp)
			goto ESCAPE
		case compile.OPCODE_NOP:
			selfVm.Pc++

		//case "print":
		case compile.OPCODE_PRINT:
			argLen := compile.DeserializePrintInstr(vm.CompilerEnv, code)
			line := ""
			for i := int64(0); i < argLen; i++ {
				line += selfVm.Pop().String(vm.CompilerEnv)
			}
			fmt.Print(line)
			selfVm.Push(compile.NewNil())
			selfVm.Pc++
		//case "println":
		case compile.OPCODE_PRINTLN:
			argLen := compile.DeserializePrintlnInstr(vm.CompilerEnv, code)
			line := ""
			for i := int64(0); i < argLen; i++ {
				line += selfVm.Pop().String(vm.CompilerEnv)
			}
			fmt.Println(line)
			selfVm.Push(compile.NewNil())
			selfVm.Pc++
		//case "+":
		case compile.OPCODE_PLUS_NUM:
			argLen := compile.DeserializePlusNumInstr(vm.CompilerEnv, code)
			sum := int64(0)
			for i := int64(0); i < argLen; i++ {
				sum += int64(selfVm.Pop().(compile.Number))
			}
			selfVm.Push(compile.Number(sum))
			selfVm.Pc++
		//case "-":
		case compile.OPCODE_MINUS_NUM:
			argLen := compile.DeserializeMinusNumInstr(vm.CompilerEnv, code)
			minus := int64(0)
			for i := int64(0); i < argLen-1; i++ {
				minus += int64(selfVm.Pop().(compile.Number))
			}
			selfVm.Push(compile.Number(int64(selfVm.Pop().(compile.Number)) - minus))
			selfVm.Pc++
		//case "*":
		case compile.OPCODE_MULTIPLY_NUM:
			argLen := compile.DeserializeMultiplyNumInstr(vm.CompilerEnv, code)
			sum := int64(1)
			for i := int64(0); i < argLen; i++ {
				sum *= int64(selfVm.Pop().(compile.Number))
			}
			selfVm.Push(compile.Number(sum))
			selfVm.Pc++
		//case "/":
		case compile.OPCODE_DIVIDE_NUM:
			argLen := compile.DeserializeDivideNumInstr(vm.CompilerEnv, code)
			sum := int64(1)
			for i := int64(0); i < argLen-1; i++ {
				sum *= int64(selfVm.Pop().(compile.Number))
			}
			if sum == 0 {
				fmt.Println("divide by zero")
				goto ESCAPE
			}
			selfVm.Push(compile.Number(int64(selfVm.Pop().(compile.Number)) / sum))
			selfVm.Pc++
		//case "mod":
		case compile.OPCODE_MODULO_NUM:
			argLen := compile.DeserializeModuloNumInstr(vm.CompilerEnv, code)

			args := make([]int64, argLen)

			for i := argLen - 1; 0 <= i; i-- {
				args[i] = int64(selfVm.Pop().(compile.Number))
			}

			sum := args[0]

			for i := int64(1); i < argLen; i++ {
				sum %= args[i]
			}

			selfVm.Push(compile.Number(sum))
			selfVm.Pc++
		//case "=":
		case compile.OPCODE_EQUAL_NUM:
			argLen := compile.DeserializeEqualNumInstr(vm.CompilerEnv, code)
			val := selfVm.Pop()
			var tmp compile.SExpression
			var result = true
			for i := int64(1); i < argLen; i++ {
				tmp = selfVm.Pop()
				if result == false {
					continue
				}
				if compile.SExpressionTypeNumber != tmp.SExpressionTypeId() {
					fmt.Println("arg is not number")
				}
				if !val.Equals(tmp) {
					result = false
				}
			}
			if result {
				selfVm.Push(compile.Bool(true))
			} else {
				selfVm.Push(compile.Bool(false))
			}
			selfVm.Pc++
		//case "!=":
		case compile.OPCODE_NOT_EQUAL_NUM:
			argLen := compile.DeserializeNotEqualNumInstr(vm.CompilerEnv, code)
			val := selfVm.Pop()
			var tmp compile.SExpression
			var result = true
			for i := int64(1); i < argLen; i++ {
				tmp = selfVm.Pop()
				if result == false {
					continue
				}
				if val.SExpressionTypeId() != compile.SExpressionTypeNumber {
					fmt.Println("arg is not number")
					goto ESCAPE
				}
				if val.Equals(tmp) {
					result = false
				}
			}
			if result {
				selfVm.Push(compile.Bool(true))
			} else {
				selfVm.Push(compile.Bool(false))
			}
			selfVm.Pc++
		//case ">":
		case compile.OPCODE_GREATER_THAN_NUM:
			argLen := compile.DeserializeGreaterThanNumInstr(vm.CompilerEnv, code)
			val := selfVm.Pop().(compile.Number)
			var tmp compile.SExpression
			flag := true
			for i := int64(1); i < argLen; i++ {
				tmp = selfVm.Pop()
				if flag == false {
					continue
				}
				if val.SExpressionTypeId() != compile.SExpressionTypeNumber {
					fmt.Println("arg is not number")
					goto ESCAPE
				}
				if val >= tmp.(compile.Number) {
					flag = false
				}
			}
			if flag {
				selfVm.Push(compile.Bool(true))
			} else {
				selfVm.Push(compile.Bool(false))
			}
			selfVm.Pc++
		//case "<":
		case compile.OPCODE_LESS_THAN_NUM:
			argLen := compile.DeserializeLessThanNumInstr(vm.CompilerEnv, code)
			val := selfVm.Pop().(compile.Number)
			var tmp compile.SExpression
			flag := true
			for i := int64(1); i < argLen; i++ {
				tmp = selfVm.Pop()
				if flag == false {
					continue
				}
				if val.SExpressionTypeId() != compile.SExpressionTypeNumber {
					fmt.Println("arg is not number")
					goto ESCAPE
				}
				if val <= tmp.(compile.Number) {
					flag = false
				}
			}
			if flag {
				selfVm.Push(compile.Bool(true))
			} else {
				selfVm.Push(compile.Bool(false))
			}
			selfVm.Pc++
		//case ">=":
		case compile.OPCODE_GREATER_THAN_OR_EQUAL_NUM:
			argLen := compile.DeserializeGreaterThanOrEqualNumInstr(vm.CompilerEnv, code)
			val := selfVm.Pop().(compile.Number)
			var tmp compile.SExpression
			flag := true
			for i := int64(1); i < argLen; i++ {
				tmp = selfVm.Pop()
				if flag == false {
					continue
				}
				if val.SExpressionTypeId() != compile.SExpressionTypeNumber {
					fmt.Println("arg is not number")
					goto ESCAPE
				}
				if val > tmp.(compile.Number) {
					flag = false
				}
			}
			if flag {
				selfVm.Push(compile.Bool(true))
			} else {
				selfVm.Push(compile.Bool(false))
			}
			selfVm.Pc++

		//case "<=":
		case compile.OPCODE_LESS_THAN_OR_EQUAL_NUM:
			argLen := compile.DeserializeLessThanOrEqualNumInstr(vm.CompilerEnv, code)
			val := selfVm.Pop().(compile.Number)
			var tmp compile.SExpression
			flag := true
			for i := int64(1); i < argLen; i++ {
				tmp = selfVm.Pop()
				if flag == false {
					continue
				}
				if val.SExpressionTypeId() != compile.SExpressionTypeNumber {
					fmt.Println("arg is not number")
					goto ESCAPE
				}
				if val < tmp.(compile.Number) {
					flag = false
				}
			}
			if flag {
				selfVm.Push(compile.Bool(true))
			} else {
				selfVm.Push(compile.Bool(false))
			}
			selfVm.Pc++
		//case "car":
		case compile.OPCODE_CAR:
			target := selfVm.Pop()
			if target.SExpressionTypeId() != compile.SExpressionTypeConsCell {
				fmt.Println("car target is not cons cell")
			}
			selfVm.Push(target.(compile.ConsCell).GetCar())
			selfVm.Pc++
		//case "cdr":
		case compile.OPCODE_CDR:
			target := selfVm.Pop()
			if target.SExpressionTypeId() != compile.SExpressionTypeConsCell {
				fmt.Println("cdr target is not cons cell")
			}
			selfVm.Push(target.(compile.ConsCell).GetCdr())
			selfVm.Pc++
		//case "random-id":
		case compile.OPCODE_RANDOM_ID:
			id := uuid.New()
			selfVm.Push(compile.Str(vm.CompilerEnv.GetCompilerSymbol(id.String())))
			selfVm.Pc++
		case compile.OPCODE_NEW_ARRAY:
			selfVm.Push(compile.NewNativeArray(vm.CompilerEnv, nil))
			selfVm.Pc++
		case compile.OPCODE_ARRAY_GET:
			arrArgSize := compile.DeserializeArrayGetInstr(vm.CompilerEnv, code)
			if arrArgSize != 2 {
				fmt.Println("array get arg size is not 2")
				goto ESCAPE
			}
			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeNumber {
				fmt.Println("index is not number")
				goto ESCAPE
			}
			index := int64(selfVm.Pop().(compile.Number))
			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeNativeArray {
				fmt.Println("not an array")
				goto ESCAPE
			}
			target := selfVm.Pop().(*compile.NativeArray)
			selfVm.Push(target.Get(index))
			selfVm.Pc++
		case compile.OPCODE_ARRAY_SET:
			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeNumber {
				fmt.Println("elem is not number")
				goto ESCAPE
			}
			elem := selfVm.Pop()
			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeNumber {
				fmt.Println("index is not number")
				goto ESCAPE
			}
			index := int64(selfVm.Pop().(compile.Number))
			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeNativeArray {
				fmt.Println("not an array")
				goto ESCAPE
			}
			target := selfVm.Pop().(*compile.NativeArray)
			if err := target.Set(index, elem); err != nil {
				fmt.Println(err)
				goto ESCAPE
			}
			selfVm.Push(target)
			selfVm.Pc++
		case compile.OPCODE_ARRAY_LENGTH:
			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeNativeArray {
				fmt.Println("not an array")
				goto ESCAPE
			}
			targetRaw := selfVm.Pop()
			target := targetRaw.(*compile.NativeArray)
			selfVm.Push(compile.Number(target.Length()))
			selfVm.Pc++
		case compile.OPCODE_ARRAY_PUSH:
			elem := selfVm.Pop()
			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeNativeArray {
				fmt.Println("not an array")
				goto ESCAPE
			}
			targetRaw := selfVm.Pop()
			target := targetRaw.(*compile.NativeArray)
			target.Push(elem)
			selfVm.Push(target)
			selfVm.Pc++
		case compile.OPCODE_NEW_MAP:
			selfVm.Push(compile.NewNativeHashmap(vm.CompilerEnv, map[uint64]compile.SExpression{}))
			selfVm.Pc++
		case compile.OPCODE_MAP_GET:
			arrArgSize := compile.DeserializeMapGetInstr(vm.CompilerEnv, code)

			if arrArgSize != 2 && arrArgSize != 3 {
				fmt.Println("map get arg size is not 2 or 3")
				goto ESCAPE
			}

			var defaultVal compile.SExpression = compile.NewNil()
			if arrArgSize == 3 {
				defaultVal = selfVm.Pop()
			}

			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeString {
				fmt.Println("key is not string")
				goto ESCAPE
			}
			key := selfVm.Pop()
			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeNativeHashmap {
				fmt.Println("not an hashmap")
				goto ESCAPE
			}
			target := selfVm.Pop().(*compile.NativeHashMap)

			val, ok := target.Get(uint64(key.(compile.Str)))
			if !ok {
				selfVm.Push(defaultVal)
			} else {
				selfVm.Push(val)
			}

			selfVm.Pc++
		case compile.OPCODE_MAP_SET:
			val := selfVm.Pop()
			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeString {
				fmt.Println("key is not string")
				goto ESCAPE
			}
			key := selfVm.Pop()
			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeNativeHashmap {
				fmt.Println("not an hashmap")
				goto ESCAPE
			}
			target := selfVm.Pop().(*compile.NativeHashMap)
			target.Set(uint64(key.(compile.Str)), val)
			selfVm.Push(target)
			selfVm.Pc++
		case compile.OPCODE_MAP_LENGTH:
			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeNativeHashmap {
				fmt.Println("not an hashmap")
				goto ESCAPE
			}
			target := selfVm.Pop().(*compile.NativeHashMap)
			selfVm.Push(compile.Number(target.Length()))
			selfVm.Pc++
		case compile.OPCODE_MAP_KEYS:
			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeNativeHashmap {
				fmt.Println("not an hashmap")
				goto ESCAPE
			}
			target := selfVm.Pop().(*compile.NativeHashMap)
			selfVm.Push(target)
			selfVm.Pc++
		case compile.OPCODE_MAP_DELETE:
			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeString {
				fmt.Println("key is not string")
				goto ESCAPE
			}
			key := selfVm.Pop()
			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeNativeHashmap {
				fmt.Println("not an hashmap")
				goto ESCAPE
			}
			target := selfVm.Pop().(*compile.NativeHashMap)
			target.Delete(uint64(key.(compile.Str)))
			selfVm.Push(target)
			selfVm.Pc++

		case compile.OPCODE_HEAVY:
			argsLen := compile.DeserializeHeavyInstr(vm.CompilerEnv, code)
			if argsLen <= 0 || argsLen > 2 {
				fmt.Println("invalid heavy instr")
				goto ESCAPE
			}
			if argsLen == 2 {
				callBackRaw := selfVm.Pop()
				if callBackRaw.SExpressionTypeId() != compile.SExpressionTypeClosure {
					fmt.Println("not a closure")
					goto ESCAPE
				}
				callBack := callBackRaw.(*Closure)
				sendBody := selfVm.Pop()
				//send to heavy
				GetSupervisor().AddTaskTaskWithCallback(vm.CompilerEnv, sendBody, callBack)
			}
			if argsLen == 1 {
				sendBody := selfVm.Pop()
				//send to heavy
				GetSupervisor().AddTask(vm.CompilerEnv, sendBody)
			}
			//selfVm.Push(compile.NewString(vm.CompilerEnv, "somethin-uuid"))
			selfVm.Pc++
		case compile.OPCODE_STRING_SPLIT:

			argsLen := compile.DeserializeStringSplitInstr(vm.CompilerEnv, code)

			if argsLen != 2 {
				fmt.Println("invalid string split instr")
				goto ESCAPE
			}

			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeString {
				fmt.Println("not a string")
				goto ESCAPE
			}

			sep := selfVm.Pop().(compile.Str).GetValue(vm.CompilerEnv)

			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeString {
				fmt.Println("not a string")
				goto ESCAPE
			}

			target := selfVm.Pop().(compile.Str).GetValue(vm.CompilerEnv)

			splitted := strings.Split(target, sep)

			var convArr = make([]compile.SExpression, len(splitted))

			for i := 0; i < len(splitted); i++ {
				convArr[i] = compile.Str(vm.CompilerEnv.GetCompilerSymbol(splitted[i]))
			}

			arr := compile.NewNativeArray(vm.CompilerEnv, convArr)

			selfVm.Push(arr)
			selfVm.Pc++
		case compile.OPCODE_STRING_JOIN:
			argsLen := compile.DeserializeStringJoinInstr(vm.CompilerEnv, code)

			if argsLen != 2 {
				fmt.Println("invalid string join instr")
				goto ESCAPE
			}

			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeString {
				fmt.Println("not a string")
				goto ESCAPE
			}

			sep := selfVm.Pop().(compile.Str).GetValue(vm.CompilerEnv)

			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeNativeArray {
				fmt.Println("not an array")
				goto ESCAPE
			}

			target := selfVm.Pop().(*compile.NativeArray)
			conv := make([]string, target.Length())

			for i := int64(0); i < target.Length(); i++ {
				conv[i] = target.Get(i).(compile.Str).GetValue(vm.CompilerEnv)
			}

			joined := strings.Join(conv, sep)

			selfVm.Push(compile.Str(vm.CompilerEnv.GetCompilerSymbol(joined)))
			selfVm.Pc++
		case compile.OPCODE_GET_NOW_TIME_NANO:
			selfVm.Push(compile.Number(time.Now().UnixNano()))
			selfVm.Pc++
		case compile.OPCODE_READ_FILE:
			pathRaw := selfVm.Pop()
			if pathRaw.SExpressionTypeId() != compile.SExpressionTypeString {
				fmt.Println("not a string")
				goto ESCAPE
			}
			filePath := pathRaw.(compile.Str).GetValue(vm.CompilerEnv)

			file, err := os.Open(filePath)
			if err != nil {
				fmt.Println(err)
				goto ESCAPE
			}
			defer file.Close()

			//read file content
			fileInfo, err := file.Stat()
			if err != nil {
				fmt.Println(err)
				goto ESCAPE
			}
			fileSize := fileInfo.Size()
			fileContent := make([]byte, fileSize)
			_, err = file.Read(fileContent)
			if err != nil {
				fmt.Println("file read err: ", err)
				goto ESCAPE
			}
			selfVm.Push(compile.Str(vm.CompilerEnv.GetCompilerSymbol(string(fileContent))))
			selfVm.Pc++
		default:
			fmt.Println("unknown opcode: ", code)
			goto ESCAPE
		}
	}
ESCAPE:
	{
		for {
			selfVm.Stack = NewSexpStack()
			selfVm.Code = []compile.Instr{}
			selfVm.Pc = 0
			if selfVm.ReturnCont == nil {
				break
			}
			selfVm = selfVm.ReturnCont
		}
	}
	return vm.Result
}

func (vm *Closure) Push(sexp compile.SExpression) {
	vm.Stack.Push(sexp)
}

func (vm *Closure) Pop() compile.SExpression {
	return vm.Stack.Pop()
}

func (vm *Closure) Peek() compile.SExpression {
	return vm.Stack.Peek()
}

func (vm *Closure) SetCont(cont *Closure) {
	vm.Cont = cont
}

func (vm *Closure) GetCont() *Closure {
	return vm.Cont
}

func (vm *Closure) SetCode(code []compile.Instr) {
	vm.Code = code
}

func (vm *Closure) AddCode(code []compile.Instr) {
	vm.Code = append(vm.Code, code...)
}

func (vm *Closure) GetCode() []compile.Instr {
	return vm.Code
}

func (vm *Closure) SetPc(pc int64) {
	vm.Pc = pc
}

func (vm *Closure) GetPc() int64 {
	return vm.Pc
}
