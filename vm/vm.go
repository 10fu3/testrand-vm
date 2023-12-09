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

type Closure struct {
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
	ResultErr     error
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
		stack: make([]compile.SExpression, 0, 8),
	}
}

func (vm Closure) TypeId() string {
	return "closure"
}

func (vm Closure) SExpressionTypeId() compile.SExpressionType {
	return compile.SExpressionTypeClosure
}

func (vm Closure) String(compEnv *compile.CompilerEnvironment) string {
	return "closure"
}

func (vm Closure) IsList() bool {
	return false
}

func (vm Closure) Equals(sexp compile.SExpression) bool {
	//TODO implement me
	panic("implement me")
}

var globalEnvMutex = uint32(0)

func NewVM(compEnv *compile.CompilerEnvironment) *Closure {
	return &Closure{
		CompilerEnv: compEnv,
		Stack: SexpStack{
			stack: make([]compile.SExpression, 0, 8),
		},
		Pc:   0,
		Cont: nil,
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
			selfVm.Stack.Push(compile.NewNil())
			selfVm.Pc++
		//case "push-sym":
		case compile.OPCODE_PUSH_SYM:
			// selfVm.Stack.Push(reader.NewSymbol(opCodeAndArgs[1]))
			selfVm.Stack.Push(compile.DeserializePushSymbolInstr(code))
			selfVm.Pc++

		//case "push-num":
		case compile.OPCODE_PUSH_NUM:
			selfVm.Stack.Push(compile.Number(compile.DeserializePushNumberInstr(vm.CompilerEnv, code)))
			selfVm.Pc++
		//case "push-boo":
		case compile.OPCODE_PUSH_TRUE:
			//val := opCodeAndArgs[1] == "#t"
			//
			//if opCodeAndArgs[1] != "#f" && opCodeAndArgs[1] != "#t" {
			//	fmt.Println("not a bool")
			//	goto ESCAPE
			//}
			selfVm.Stack.Push(compile.Bool(true))
			selfVm.Pc++
		case compile.OPCODE_PUSH_FALSE:
			selfVm.Stack.Push(compile.Bool(false))
			selfVm.Pc++
		//case "push-str":
		case compile.OPCODE_PUSH_STR:
			// selfVm.Stack.Push(reader.NewString(opCodeAndArgs[1]))
			selfVm.Stack.Push(compile.DeserializePushStringInstr(vm.CompilerEnv, code))
			selfVm.Pc++
		//case "pop":
		case compile.OPCODE_POP:
			selfVm.Stack.Pop()
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
			val, ok := selfVm.Stack.Pop().(compile.Bool)
			if !ok {
				selfVm.ResultErr = errors.New("not a bool")
				goto ESCAPE
			}
			if val {
				selfVm.Pc = jumpTo
			} else {
				selfVm.Pc++
			}
		//case "jmp-else":
		case compile.OPCODE_JMP_ELSE:
			//jumpTo, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			jumpTo := compile.DeserializeJmpElseInstr(vm.CompilerEnv, code)
			val, ok := selfVm.Stack.Pop().(compile.Bool)
			if !ok {
				selfVm.ResultErr = errors.New("not a bool")
				goto ESCAPE
			}

			if !val {
				selfVm.Pc = jumpTo
			} else {
				selfVm.Pc++
			}

		//case "load":
		case compile.OPCODE_LOAD:

			envId, symId := compile.DeserializeLoadInstr(vm.CompilerEnv, code)
			for !atomic.CompareAndSwapUint32(&globalEnvMutex, 0, 1) {
			}
			loaded := vm.CompilerEnv.GlobalEnv[envId]
			selfVm.Stack.Push(loaded.Frame[symId])
			atomic.StoreUint32(&globalEnvMutex, 0)
			selfVm.Pc++

		//case "define":
		case compile.OPCODE_DEFINE:
			//sym := reader.NewSymbol(opCodeAndArgs[1])
			envId, symIndexId := compile.DeserializeDefineInstr(vm.CompilerEnv, code)
			val := selfVm.Stack.Pop()
			for !atomic.CompareAndSwapUint32(&globalEnvMutex, 0, 1) {
			}
			vm.CompilerEnv.GlobalEnv[envId].Frame[symIndexId] = val
			atomic.StoreUint32(&globalEnvMutex, 0)
			symId, err := vm.CompilerEnv.FindSymbolInEnvironment(envId, symIndexId)
			if err != nil {
				selfVm.ResultErr = err
				goto ESCAPE
			}
			selfVm.Stack.Push(compile.NewSymbol(symId))
			selfVm.Pc++
		//case "define-args":
		case compile.OPCODE_DEFINE_ARGS:
			//sym := reader.NewSymbol(opCodeAndArgs[1])
			_, symId := compile.DeserializeDefineArgsInstr(vm.CompilerEnv, code)
			selfVm.Stack.Push(compile.NewSymbol(symId))
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
				selfVm.ResultErr = err
			}
			selfVm.Stack.Push(deserialize)
			selfVm.Pc++
		//case "set":
		case compile.OPCODE_SET:
			//sym := reader.NewSymbol(opCodeAndArgs[1])
			envId, symId := compile.DeserializeSetInstr(vm.CompilerEnv, code)
			for !atomic.CompareAndSwapUint32(&globalEnvMutex, 0, 1) {
			}
			vm.CompilerEnv.GlobalEnv[envId].Frame[symId] = selfVm.Stack.Peek()
			atomic.StoreUint32(&globalEnvMutex, 0)
			selfVm.Pc++
		//case "new-env":
		case compile.OPCODE_NEW_ENV:

			_, envId := compile.DeserializeNewEnvInstr(vm.CompilerEnv, code)

			newEnv := compile.RuntimeEnv{
				SelfIndex: envId,
				Frame:     make(map[uint64]compile.SExpression),
			}
			for !atomic.CompareAndSwapUint32(&globalEnvMutex, 0, 1) {
			}

			vm.CompilerEnv.GlobalEnv = append(vm.CompilerEnv.GlobalEnv, newEnv)
			atomic.StoreUint32(&globalEnvMutex, 0)

			selfVm.Stack.Push(newEnv)
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
				sym := selfVm.Stack.Pop().(compile.Symbol)
				newVm.TemporaryArgs = append(newVm.TemporaryArgs, sym)
			}

			newVm.Cont = selfVm
			newVm.EnvId = selfVm.Stack.Pop().(compile.RuntimeEnv).SelfIndex
			newVm.Pc = 0
			selfVm.Stack.Push(newVm)
			selfVm.Pc++
		//case "call":
		case compile.OPCODE_CALL:
			nextVm, ok := selfVm.Stack.Pop().(*Closure)

			if !ok {
				selfVm.ResultErr = errors.New("not a closure")
				goto ESCAPE
			}

			nextEnvId := nextVm.EnvId
			for !atomic.CompareAndSwapUint32(&globalEnvMutex, 0, 1) {
			}
			newEnv := vm.CompilerEnv.GlobalEnv[nextEnvId]
			atomic.StoreUint32(&globalEnvMutex, 0)

			argsSize := compile.DeserializeCallInstr(vm.CompilerEnv, code)

			if argsSize != int64(len(nextVm.TemporaryArgs)) {
				selfVm.ResultErr = errors.New("args size not match")
				goto ESCAPE
			}

			for _, sym := range nextVm.TemporaryArgs {
				val := selfVm.Stack.Pop()
				newEnv.Frame[uint64(sym)] = val
			}

			nextVm.ReturnCont = selfVm
			nextVm.ReturnPc = selfVm.Pc
			selfVm = nextVm
		//case "ret":
		case compile.OPCODE_RETURN:
			val := selfVm.Stack.Pop()
			retPc := selfVm.ReturnPc
			selfVm.Pc = 0
			selfVm = selfVm.ReturnCont
			selfVm.Pc = retPc
			selfVm.Stack.Push(val)
			selfVm.Pc++
		//case "and":
		case compile.OPCODE_AND:
			//argsSize, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			argsSize := compile.DeserializeAndInstr(vm.CompilerEnv, code)
			val, ok := selfVm.Stack.Pop().(compile.Bool)
			var tmp compile.Bool
			flag := true
			for i := int64(1); i < argsSize; i++ {
				if !ok {
					selfVm.ResultErr = errors.New("arg is not bool")
					goto ESCAPE
				}
				tmp, ok = selfVm.Stack.Pop().(compile.Bool)
				if flag == false {
					continue
				}
				if val != tmp {
					flag = false
				}
			}
			if flag {
				selfVm.Stack.Push(compile.Bool(true))
			} else {
				selfVm.Stack.Push(compile.Bool(false))
			}
			selfVm.Pc++
		//case "or":
		case compile.OPCODE_OR:
			//argsSize, _ := strconv.ParseInt(opCodeAndArgs[1], 10, 64)
			argsSize := compile.DeserializeOrInstr(vm.CompilerEnv, code)
			tmp, ok := selfVm.Stack.Pop().(compile.Bool)
			flag := false
			for i := int64(0); i < argsSize; i++ {
				if !ok {
					selfVm.ResultErr = errors.New("arg is not bool")
					goto ESCAPE
				}
				if tmp {
					flag = true
				}
				if flag == true {
					continue
				}
				tmp, ok = selfVm.Stack.Pop().(compile.Bool)
			}
			if flag {
				selfVm.Stack.Push(compile.Bool(true))
			} else {
				selfVm.Stack.Push(compile.Bool(false))
			}
			selfVm.Pc++
		//case "end-code":
		case compile.OPCODE_END_CODE:
			val := selfVm.Stack.Pop()
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
				line += selfVm.Stack.Pop().String(vm.CompilerEnv)
			}
			fmt.Print(line)
			selfVm.Stack.Push(compile.NewNil())
			selfVm.Pc++
		//case "println":
		case compile.OPCODE_PRINTLN:
			argLen := compile.DeserializePrintlnInstr(vm.CompilerEnv, code)
			line := ""
			for i := int64(0); i < argLen; i++ {
				line += selfVm.Stack.Pop().String(vm.CompilerEnv)
			}
			fmt.Println(line)
			selfVm.Stack.Push(compile.NewNil())
			selfVm.Pc++
		//case "+":
		case compile.OPCODE_PLUS_NUM:
			argLen := compile.DeserializePlusNumInstr(vm.CompilerEnv, code)
			sum := int64(0)
			tmp := compile.Number(0)
			ok := false
			for i := int64(0); i < argLen; i++ {
				tmp, ok = selfVm.Stack.Pop().(compile.Number)
				if !ok {
					selfVm.ResultErr = errors.New("arg is not number")
					goto ESCAPE
				}
				sum += int64(tmp)
			}
			selfVm.Stack.Push(compile.Number(sum))
			selfVm.Pc++
		//case "-":
		case compile.OPCODE_MINUS_NUM:
			argLen := compile.DeserializeMinusNumInstr(vm.CompilerEnv, code)
			minus := int64(0)
			tmp := compile.Number(0)
			ok := false
			for i := int64(0); i < argLen-1; i++ {
				tmp, ok = selfVm.Stack.Pop().(compile.Number)
				if !ok {
					selfVm.ResultErr = errors.New("arg is not number")
					goto ESCAPE
				}
				minus += int64(tmp)
			}
			selfVm.Stack.Push(compile.Number(int64(selfVm.Stack.Pop().(compile.Number)) - minus))
			selfVm.Pc++
		//case "*":
		case compile.OPCODE_MULTIPLY_NUM:
			argLen := compile.DeserializeMultiplyNumInstr(vm.CompilerEnv, code)
			sum := int64(1)
			tmp := compile.Number(0)
			ok := false
			for i := int64(0); i < argLen; i++ {
				tmp, ok = selfVm.Stack.Pop().(compile.Number)
				if !ok {
					selfVm.ResultErr = errors.New("arg is not number")
					goto ESCAPE
				}
				sum *= int64(tmp)
			}
			selfVm.Stack.Push(compile.Number(sum))
			selfVm.Pc++
		//case "/":
		case compile.OPCODE_DIVIDE_NUM:
			argLen := compile.DeserializeDivideNumInstr(vm.CompilerEnv, code)
			sum := int64(1)
			tmp := compile.Number(0)
			ok := false

			for i := int64(0); i < argLen-1; i++ {
				tmp, ok = selfVm.Stack.Pop().(compile.Number)
				if !ok {
					selfVm.ResultErr = errors.New("arg is not number")
					goto ESCAPE
				}
				if tmp == 0 {
					selfVm.ResultErr = errors.New("divide by zero")
					goto ESCAPE
				}
				sum *= int64(tmp)
			}

			if sum == 0 {
				selfVm.ResultErr = errors.New("divide by zero")
				goto ESCAPE
			}
			selfVm.Stack.Push(compile.Number(int64(selfVm.Stack.Pop().(compile.Number)) / sum))
			selfVm.Pc++
		//case "mod":
		case compile.OPCODE_MODULO_NUM:
			argLen := compile.DeserializeModuloNumInstr(vm.CompilerEnv, code)

			args := make([]int64, argLen)
			tmp := compile.Number(0)
			ok := false
			for i := argLen - 1; 0 <= i; i-- {
				tmp, ok = selfVm.Stack.Pop().(compile.Number)
				if !ok {
					selfVm.ResultErr = errors.New("arg is not number")
					goto ESCAPE
				}
				args[i] = int64(tmp)
			}

			sum := args[0]

			for i := int64(1); i < argLen; i++ {
				sum %= args[i]
			}

			selfVm.Stack.Push(compile.Number(sum))
			selfVm.Pc++
		//case "=":
		case compile.OPCODE_EQUAL_NUM:
			argLen := compile.DeserializeEqualNumInstr(vm.CompilerEnv, code)
			val := selfVm.Stack.Pop()
			var tmp compile.Number
			var ok bool
			var result = true
			for i := int64(1); i < argLen; i++ {
				tmp, ok = selfVm.Stack.Pop().(compile.Number)
				if result == false {
					continue
				}
				if !ok {
					selfVm.ResultErr = errors.New("arg is not number")
				}
				if val != tmp {
					result = false
				}
			}
			if result {
				selfVm.Stack.Push(compile.Bool(true))
			} else {
				selfVm.Stack.Push(compile.Bool(false))
			}
			selfVm.Pc++
		//case "!=":
		case compile.OPCODE_NOT_EQUAL_NUM:
			argLen := compile.DeserializeNotEqualNumInstr(vm.CompilerEnv, code)
			val := selfVm.Stack.Pop()
			var tmp compile.Number
			var ok bool
			var result = true
			for i := int64(1); i < argLen; i++ {
				tmp, ok = selfVm.Stack.Pop().(compile.Number)
				if result == false {
					continue
				}
				if !ok {
					selfVm.ResultErr = errors.New("arg is not number")
				}
				if val == tmp {
					result = false
				}
			}
			if result {
				selfVm.Stack.Push(compile.Bool(true))
			} else {
				selfVm.Stack.Push(compile.Bool(false))
			}
			selfVm.Pc++
		//case ">":
		case compile.OPCODE_GREATER_THAN_NUM:
			argLen := compile.DeserializeGreaterThanNumInstr(vm.CompilerEnv, code)
			val, ok := selfVm.Stack.Pop().(compile.Number)
			var tmp compile.Number
			flag := true
			for i := int64(1); i < argLen; i++ {
				if !ok {
					selfVm.ResultErr = errors.New("arg is not number")
					goto ESCAPE
				}
				tmp, ok = selfVm.Stack.Pop().(compile.Number)
				if flag == false {
					continue
				}
				if !ok {
					selfVm.ResultErr = errors.New("arg is not number")
					goto ESCAPE
				}
				if val >= tmp {
					flag = false
				}
			}
			if flag {
				selfVm.Stack.Push(compile.Bool(true))
			} else {
				selfVm.Stack.Push(compile.Bool(false))
			}
			selfVm.Pc++
		//case "<":
		case compile.OPCODE_LESS_THAN_NUM:
			argLen := compile.DeserializeLessThanNumInstr(vm.CompilerEnv, code)
			val, ok := selfVm.Stack.Pop().(compile.Number)
			var tmp compile.Number
			flag := true
			for i := int64(1); i < argLen; i++ {
				if !ok {
					selfVm.ResultErr = errors.New("arg is not number")
					goto ESCAPE
				}
				tmp, ok = selfVm.Stack.Pop().(compile.Number)
				if flag == false {
					continue
				}
				if !ok {
					selfVm.ResultErr = errors.New("arg is not number")
					goto ESCAPE
				}
				if val <= tmp {
					flag = false
				}
			}
			if flag {
				selfVm.Stack.Push(compile.Bool(true))
			} else {
				selfVm.Stack.Push(compile.Bool(false))
			}
			selfVm.Pc++
		//case ">=":
		case compile.OPCODE_GREATER_THAN_OR_EQUAL_NUM:
			argLen := compile.DeserializeGreaterThanOrEqualNumInstr(vm.CompilerEnv, code)
			val, ok := selfVm.Stack.Pop().(compile.Number)
			var tmp compile.Number
			flag := true
			for i := int64(1); i < argLen; i++ {
				if !ok {
					selfVm.ResultErr = errors.New("arg is not number")
					goto ESCAPE
				}
				tmp, ok = selfVm.Stack.Pop().(compile.Number)
				if !ok {
					selfVm.ResultErr = errors.New("arg is not number")
					goto ESCAPE
				}
				if flag == false {
					continue
				}
				if val > tmp {
					flag = false
				}
			}
			if flag {
				selfVm.Stack.Push(compile.Bool(true))
			} else {
				selfVm.Stack.Push(compile.Bool(false))
			}
			selfVm.Pc++

		//case "<=":
		case compile.OPCODE_LESS_THAN_OR_EQUAL_NUM:
			argLen := compile.DeserializeLessThanOrEqualNumInstr(vm.CompilerEnv, code)
			val, ok := selfVm.Stack.Pop().(compile.Number)
			var tmp compile.Number
			flag := true
			for i := int64(1); i < argLen; i++ {
				if !ok {
					selfVm.ResultErr = errors.New("arg is not number")
					goto ESCAPE
				}
				tmp, ok = selfVm.Stack.Pop().(compile.Number)
				if flag == false {
					continue
				}
				if !ok {
					selfVm.ResultErr = errors.New("arg is not number")
					goto ESCAPE
				}
				if val < tmp {
					flag = false
				}
			}
			if flag {
				selfVm.Stack.Push(compile.Bool(true))
			} else {
				selfVm.Stack.Push(compile.Bool(false))
			}
			selfVm.Pc++
		//case "car":
		case compile.OPCODE_CAR:
			target, ok := selfVm.Stack.Pop().(compile.ConsCell)
			if !ok {
				selfVm.ResultErr = errors.New("car target is not cons cell")
				goto ESCAPE
			}
			selfVm.Stack.Push(target.GetCar())
			selfVm.Pc++
		//case "cdr":
		case compile.OPCODE_CDR:
			target, ok := selfVm.Stack.Pop().(compile.ConsCell)
			if !ok {
				selfVm.ResultErr = errors.New("cdr target is not cons cell")
				goto ESCAPE
			}
			selfVm.Stack.Push(target.GetCdr())
			selfVm.Pc++
		//case "random-id":
		case compile.OPCODE_RANDOM_ID:
			id := uuid.New()
			selfVm.Stack.Push(compile.Str(vm.CompilerEnv.GetCompilerSymbol(id.String())))
			selfVm.Pc++
		case compile.OPCODE_NEW_ARRAY:
			selfVm.Stack.Push(compile.NewNativeArray(vm.CompilerEnv, nil))
			selfVm.Pc++
		case compile.OPCODE_ARRAY_GET:
			arrArgSize := compile.DeserializeArrayGetInstr(vm.CompilerEnv, code)
			if arrArgSize != 2 {
				selfVm.ResultErr = errors.New("array get arg size is not 2")
				goto ESCAPE
			}
			index, ok := selfVm.Stack.Pop().(compile.Number)
			if !ok {
				selfVm.ResultErr = errors.New("index is not number")
				goto ESCAPE
			}
			target, ok := selfVm.Stack.Pop().(*compile.NativeArray)
			if !ok {
				selfVm.ResultErr = errors.New("not an array")
				goto ESCAPE
			}
			selfVm.Stack.Push(target.Get(int64(index)))
			selfVm.Pc++
		case compile.OPCODE_ARRAY_SET:
			elem, ok := selfVm.Stack.Pop().(compile.Number)
			if !ok {
				selfVm.ResultErr = errors.New("elem is not number")
				goto ESCAPE
			}
			rawIndex, ok := selfVm.Stack.Pop().(compile.Number)
			if !ok {
				selfVm.ResultErr = errors.New("index is not number")
				goto ESCAPE
			}
			target, ok := selfVm.Stack.Pop().(*compile.NativeArray)
			if !ok {
				selfVm.ResultErr = errors.New("not an array")
				goto ESCAPE
			}
			if err := target.Set(int64(rawIndex), elem); err != nil {
				selfVm.ResultErr = err
				goto ESCAPE
			}
			selfVm.Stack.Push(target)
			selfVm.Pc++
		case compile.OPCODE_ARRAY_LENGTH:
			targetRaw := selfVm.Stack.Pop()
			target, ok := targetRaw.(*compile.NativeArray)
			if !ok {
				selfVm.ResultErr = errors.New("not an array")
				goto ESCAPE
			}
			selfVm.Stack.Push(compile.Number(target.Length()))
			selfVm.Pc++
		case compile.OPCODE_ARRAY_PUSH:
			elem := selfVm.Stack.Pop()
			targetRaw := selfVm.Stack.Pop()
			target, ok := targetRaw.(*compile.NativeArray)
			if !ok {
				fmt.Println("not an array")
				goto ESCAPE
			}
			target.Push(elem)
			selfVm.Stack.Push(target)
			selfVm.Pc++
		case compile.OPCODE_NEW_MAP:
			selfVm.Stack.Push(compile.NewNativeHashmap(vm.CompilerEnv, map[uint64]compile.SExpression{}))
			selfVm.Pc++
		case compile.OPCODE_MAP_GET:
			arrArgSize := compile.DeserializeMapGetInstr(vm.CompilerEnv, code)

			if arrArgSize != 2 && arrArgSize != 3 {
				selfVm.ResultErr = errors.New("map get arg size is not 2 or 3")
				goto ESCAPE
			}

			var defaultVal compile.SExpression = compile.NewNil()
			if arrArgSize == 3 {
				defaultVal = selfVm.Stack.Pop()
			}

			key, ok := selfVm.Stack.Pop().(compile.Str)
			if !ok {
				selfVm.ResultErr = errors.New("key is not string")
				goto ESCAPE
			}
			target, ok := selfVm.Stack.Pop().(*compile.NativeHashMap)
			if !ok {
				selfVm.ResultErr = errors.New("not an hashmap")
				goto ESCAPE
			}

			val, ok := target.Get(uint64(key))
			if !ok {
				selfVm.Stack.Push(defaultVal)
			} else {
				selfVm.Stack.Push(val)
			}

			selfVm.Pc++
		case compile.OPCODE_MAP_SET:
			val := selfVm.Stack.Pop()
			key, ok := selfVm.Stack.Pop().(compile.Str)
			if !ok {
				selfVm.ResultErr = errors.New("key is not string")
				goto ESCAPE
			}
			target, ok := selfVm.Stack.Pop().(*compile.NativeHashMap)
			if !ok {
				selfVm.ResultErr = errors.New("not an hashmap")
				goto ESCAPE
			}
			target.Set(uint64(key), val)
			selfVm.Stack.Push(target)
			selfVm.Pc++
		case compile.OPCODE_MAP_LENGTH:
			target, ok := selfVm.Stack.Pop().(*compile.NativeHashMap)
			if !ok {
				selfVm.ResultErr = errors.New("not an hashmap")
				goto ESCAPE
			}
			selfVm.Stack.Push(compile.Number(target.Length()))
			selfVm.Pc++
		case compile.OPCODE_MAP_KEYS:
			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeNativeHashmap {
				selfVm.ResultErr = errors.New("not an hashmap")
				goto ESCAPE
			}
			target := selfVm.Stack.Pop().(*compile.NativeHashMap)
			selfVm.Stack.Push(target)
			selfVm.Pc++
		case compile.OPCODE_MAP_DELETE:
			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeString {
				selfVm.ResultErr = errors.New("key is not string")
				goto ESCAPE
			}
			key := selfVm.Stack.Pop()
			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeNativeHashmap {
				selfVm.ResultErr = errors.New("not an hashmap")
				goto ESCAPE
			}
			target := selfVm.Stack.Pop().(*compile.NativeHashMap)
			target.Delete(uint64(key.(compile.Str)))
			selfVm.Stack.Push(target)
			selfVm.Pc++

		case compile.OPCODE_HEAVY:
			argsLen := compile.DeserializeHeavyInstr(vm.CompilerEnv, code)
			if argsLen <= 0 || argsLen > 2 {
				selfVm.ResultErr = errors.New("invalid heavy instr")
				goto ESCAPE
			}
			if argsLen == 2 {
				callBackRaw := selfVm.Stack.Pop()
				if callBackRaw.SExpressionTypeId() != compile.SExpressionTypeClosure {
					selfVm.ResultErr = errors.New("not a closure")
					goto ESCAPE
				}
				callBack := callBackRaw.(*Closure)
				sendBody := selfVm.Stack.Pop()
				//send to heavy
				GetSupervisor().AddTaskWithCallback(sendBody, callBack)
			}
			if argsLen == 1 {
				sendBody := selfVm.Stack.Pop()
				//send to heavy
				GetSupervisor().AddTask(vm.CompilerEnv, sendBody)
			}
			// selfVm.Stack.Push(compile.NewString(vm.CompilerEnv, "somethin-uuid"))
			selfVm.Pc++
		case compile.OPCODE_STRING_SPLIT:

			argsLen := compile.DeserializeStringSplitInstr(vm.CompilerEnv, code)

			if argsLen != 2 {
				selfVm.ResultErr = errors.New("invalid string split instr")
				goto ESCAPE
			}

			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeString {
				selfVm.ResultErr = errors.New("not a string")
				goto ESCAPE
			}

			s, ok := selfVm.Stack.Pop().(compile.Str)
			if !ok {
				selfVm.ResultErr = errors.New("not a string")
				goto ESCAPE
			}
			sep := s.GetValue(vm.CompilerEnv)

			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeString {
				selfVm.ResultErr = errors.New("not a string")
				goto ESCAPE
			}

			t, ok := selfVm.Stack.Pop().(compile.Str)
			if !ok {
				selfVm.ResultErr = errors.New("not a string")
				goto ESCAPE
			}
			target := t.GetValue(vm.CompilerEnv)

			splitted := strings.Split(target, sep)

			var convArr = make([]compile.SExpression, len(splitted))

			for i := 0; i < len(splitted); i++ {
				convArr[i] = compile.Str(vm.CompilerEnv.GetCompilerSymbol(splitted[i]))
			}

			arr := compile.NewNativeArray(vm.CompilerEnv, convArr)

			selfVm.Stack.Push(arr)
			selfVm.Pc++
		case compile.OPCODE_STRING_JOIN:
			argsLen := compile.DeserializeStringJoinInstr(vm.CompilerEnv, code)

			if argsLen != 2 {
				selfVm.ResultErr = errors.New("invalid string join instr")
				goto ESCAPE
			}

			if selfVm.Peek().SExpressionTypeId() != compile.SExpressionTypeString {
				selfVm.ResultErr = errors.New("not a string")
				goto ESCAPE
			}

			s, ok := selfVm.Stack.Pop().(compile.Str)
			if !ok {
				selfVm.ResultErr = errors.New("not a string")
				goto ESCAPE
			}
			sep := s.GetValue(vm.CompilerEnv)

			target, ok := selfVm.Stack.Pop().(*compile.NativeArray)
			if !ok {
				selfVm.ResultErr = errors.New("not an array")
				goto ESCAPE
			}
			conv := make([]string, target.Length())

			for i := int64(0); i < target.Length(); i++ {
				conv[i] = target.Get(i).(compile.Str).GetValue(vm.CompilerEnv)
			}

			joined := strings.Join(conv, sep)

			selfVm.Stack.Push(compile.Str(vm.CompilerEnv.GetCompilerSymbol(joined)))
			selfVm.Pc++
		case compile.OPCODE_GET_NOW_TIME_NANO:
			selfVm.Stack.Push(compile.Number(time.Now().UnixNano()))
			selfVm.Pc++
		case compile.OPCODE_READ_FILE:
			pathRaw, ok := selfVm.Stack.Pop().(compile.Str)
			if !ok {
				selfVm.ResultErr = errors.New("not a string")
				goto ESCAPE
			}
			filePath := pathRaw.GetValue(vm.CompilerEnv)

			file, err := os.Open(filePath)
			if err != nil {
				selfVm.ResultErr = err
				goto ESCAPE
			}
			defer file.Close()

			//read file content
			fileInfo, err := file.Stat()
			if err != nil {
				selfVm.ResultErr = err
				goto ESCAPE
			}
			fileSize := fileInfo.Size()
			fileContent := make([]byte, fileSize)
			_, err = file.Read(fileContent)
			if err != nil {
				selfVm.ResultErr = err
				goto ESCAPE
			}
			selfVm.Stack.Push(compile.Str(vm.CompilerEnv.GetCompilerSymbol(string(fileContent))))
			selfVm.Pc++
		default:
			selfVm.ResultErr = errors.New(fmt.Sprintf("unknown opcode: %s", code.String()))
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
		atomic.StoreUint32(&globalEnvMutex, 0)
	}
	return vm.Result
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
