package compile

import (
	"errors"
)

func ToArraySexp(sexp SExpression) ([]SExpression, int64) {
	var result []SExpression
	var count int64 = 0
	for {
		if SExpressionTypeConsCell != sexp.SExpressionTypeId() {
			break
		}
		if IsEmptyList(sexp) {
			break
		}
		cell := sexp.(ConsCell)
		result = append(result, cell.GetCar())
		count++
		sexp = cell.GetCdr()
	}
	return result, count
}

func IsNativeFunc(compEnv *CompilerEnvironment, sexp SExpression) bool {
	if SExpressionTypeSymbol != sexp.SExpressionTypeId() {
		return false
	}
	if NativeFuncNameToOpCodeMap[sexp.(Symbol).String(compEnv)] == nil {
		return false
	}
	return true
}

func GenerateOpCode(compileEnv *CompilerEnvironment, sexp SExpression, nowStartLine int64) ([]Instr, int64, error) {
	codes, leng, err := _generateOpCode(compileEnv, sexp, nowStartLine)
	//return append(codes, NewSymbol("end-code")), leng + 1
	return append(codes, CreateEndCodeInstr()), leng + 1, err
}

func _generateOpCode(compileEnv *CompilerEnvironment, sexp SExpression, nowStartLine int64) ([]Instr, int64, error) {
	switch sexp.SExpressionTypeId() {
	case SExpressionTypeSymbol:
		symId := compileEnv.GetCompilerSymbol(sexp.(Symbol).String(compileEnv))
		return []Instr{CreateLoadInstr(symId)}, 1, nil
	case SExpressionTypeNumber:
		//return []SExpression{NewSymbol(fmt.Sprintf("push-num %s", sexp.String()))}, 1
		return []Instr{CreatePushNumberInstr(sexp.(Number).GetValue())}, 1, nil
	case SExpressionTypeBool:
		//return []SExpression{NewSymbol(fmt.Sprintf("push-boo %s", sexp.String()))}, 1
		return []Instr{CreatePushBoolInstr(sexp.(Bool).GetValue())}, 1, nil
	case SExpressionTypeString:
		i := compileEnv.GetCompilerSymbol(sexp.(Str).GetValue(compileEnv))
		return []Instr{CreatePushStringInstr(i)}, 1, nil
	case SExpressionTypeNil:
		return []Instr{CreatePushNilInstr()}, 1, nil
	}

	cell := sexp.(ConsCell)

	label := cell.GetCar()

	if SExpressionTypeSymbol != label.SExpressionTypeId() {
		carOpCode, carAffectedCode, err := _generateOpCode(compileEnv, cell.GetCar(), nowStartLine)
		if err != nil {
			return nil, 0, err
		}
		if IsEmptyList(cell.GetCdr()) {
			return carOpCode, carAffectedCode, nil
		}

		_, argsLen := ToArraySexp(cell.GetCdr())
		cdrOpCode, cdrAffectedCode, err := _generateOpCode(compileEnv, cell.GetCdr(), nowStartLine+carAffectedCode)
		//return append(append(cdrOpCode, carOpCode...), NewSymbol(fmt.Sprintf("call %d", argsLen))), carAffectedCode + cdrAffectedCode + 1
		return append(append(cdrOpCode, carOpCode...), CreateCallInstr(argsLen)), carAffectedCode + cdrAffectedCode + 1, nil
	}

	cellContent := cell.GetCdr().(ConsCell)
	cellArr, cellArrLen := ToArraySexp(cellContent)

	switch label.(Symbol).String(compileEnv) {
	case "quote":
		if cellArrLen != 1 {
			return nil, 0, errors.New("Invalid Syntax Quote")
		}
		//return []SExpression{NewSymbol(fmt.Sprintf("load-sexp %s\n", cellArr[0]))}, 1
		i := compileEnv.GetCompilerSymbol(cellArr[0].String(compileEnv))
		return []Instr{CreatePushSExpressionInstr(i)}, 1, nil

	//case "loop":
	//	if 2 != cellArrLen {
	//		panic("Invalid syntax 2")
	//	}
	//	// cond-opcode(?)|jump-(1)|loop-body-opcode(?)|jump-lable(1)
	//	loopCond := cellArr[0]
	//	condOpCode, condAffectedCode := _generateOpCode(loopCond, nowStartLine)
	//	loopBody := cellArr[1]
	//	bodyOpCode, bodyAffectedCode := _generateOpCode(loopBody, nowStartLine+condAffectedCode+1)
	//	return append(append(condOpCode, NewSymbol(fmt.Sprintf("jump %d", nowStartLine+condAffectedCode+bodyAffectedCode+2))), append(bodyOpCode, NewSymbol(fmt.Sprintf("jump-%d", nowStartLine)))...), condAffectedCode + bodyAffectedCode + 2
	case "begin":
		bodies, bodiesSize := ToArraySexp(cellContent)
		var result []Instr
		var lineNum = nowStartLine
		var addedRows = int64(0)
		for i := int64(0); i < bodiesSize; i++ {
			bodiesOpCodes, affectedOpCodeLine, err := _generateOpCode(compileEnv, bodies[i], lineNum)
			if err != nil {
				return nil, 0, err
			}
			lineNum += affectedOpCodeLine
			addedRows += affectedOpCodeLine
			result = append(result, bodiesOpCodes...)
			if i != bodiesSize-1 {
				lineNum += 1
				addedRows += 1
				//result = append(result, NewSymbol("pop"))
				result = append(result, CreatePopInstr())
			}
		}
		return result, int64(len(result)), nil
	case "cond":
		condAndBody, condAndBodySize := ToArraySexp(cellContent)

		if 0 == condAndBodySize {
			return nil, 0, errors.New("Invalid Syntax 3")
		}

		var opCodes = []Instr{}

		var lastIndexes = make([]int64, condAndBodySize)
		var indexesIndex = int64(0)

		nowLine := nowStartLine

		for i := int64(0); i < condAndBodySize; i++ {
			condAndBodyCell := condAndBody[i].(ConsCell)
			condAndBodyCellArr, _ := ToArraySexp(condAndBodyCell)

			if 2 != len(condAndBodyCellArr) {
				return nil, 0, errors.New("Invalid Syntax 4")
			}
			condSexp := condAndBodyCellArr[0]
			bodySexp := condAndBodyCellArr[1]

			condOpCodes, condAffectedCode, err := _generateOpCode(compileEnv, condSexp, nowLine)

			if err != nil {
				return nil, 0, err
			}

			bodyOpCodes, bodyAffectedCode, err := _generateOpCode(compileEnv, bodySexp, nowLine+condAffectedCode+1)

			if err != nil {
				return nil, 0, err
			}

			opCodes = append(opCodes, make([]Instr, condAffectedCode+bodyAffectedCode+2)...)

			for j := int64(0); j < condAffectedCode; j++ {
				opCodes[j+indexesIndex] = condOpCodes[j]
			}

			indexesIndex += condAffectedCode

			//opCodes[indexesIndex] = NewSymbol(fmt.Sprintf("jmp-else %d", nowLine+condAffectedCode+bodyAffectedCode+2))
			opCodes[indexesIndex] = CreateJmpElseInstr(nowLine + condAffectedCode + bodyAffectedCode + 2)

			indexesIndex += 1

			for j := int64(0); j < bodyAffectedCode; j++ {
				opCodes[j+indexesIndex] = bodyOpCodes[j]
			}

			indexesIndex += bodyAffectedCode

			//opCodes[indexesIndex] = NewSymbol("temporary jump")
			opCodes[indexesIndex] = CreateDummyInstr()
			lastIndexes[i] = indexesIndex

			indexesIndex += 1

			nowLine += condAffectedCode + bodyAffectedCode + 2
		}

		for i := int64(0); i < condAndBodySize; i++ {
			//opCodes[lastIndexes[i]] = NewSymbol(fmt.Sprintf("jmp %d", nowLine))
			opCodes[lastIndexes[i]] = CreateJmpInstr(nowLine)
		}

		return opCodes, int64(len(opCodes)), nil

	case "and":
		cond, condLen := ToArraySexp(cellContent)

		if 0 == condLen {
			return nil, 0, errors.New("Invalid syntax 3")
		}

		var opCodes = []Instr{}

		affectedCode := nowStartLine

		for i := int64(0); i < condLen; i++ {
			condOpCodes, condAffectedCode, err := _generateOpCode(compileEnv, cond[i], affectedCode)
			if err != nil {
				return nil, 0, err
			}
			affectedCode += condAffectedCode
			opCodes = append(opCodes, condOpCodes...)
		}

		//opCodes = append(opCodes, NewSymbol(fmt.Sprintf("and %d", condLen)))
		opCodes = append(opCodes, CreateAndInstr(condLen))

		return opCodes, affectedCode - nowStartLine + 1, nil

	case "or":
		cond, condLen := ToArraySexp(cellContent)

		if 0 == condLen {
			return nil, 0, errors.New("Invalid syntax 3")
		}

		var opCodes = []Instr{}

		affectedCode := nowStartLine

		for i := int64(0); i < condLen; i++ {
			condOpCodes, condAffectedCode, err := _generateOpCode(compileEnv, cond[i], affectedCode)
			if err != nil {
				return nil, 0, err
			}
			affectedCode += condAffectedCode
			opCodes = append(opCodes, condOpCodes...)
		}

		//opCodes = append(opCodes, NewSymbol(fmt.Sprintf("or %d", condLen)))
		opCodes = append(opCodes, CreateOrInstr(condLen))

		return opCodes, affectedCode - nowStartLine + 1, nil

	case "set":
		if 2 != len(cellArr) {
			return nil, 0, errors.New("Invalid syntax 4")
		}
		symbol, ok := cellArr[0].(Symbol)
		if !ok {
			return nil, 0, errors.New("Invalid syntax 4")
		}
		value := cellArr[1]
		opCodes, affectedCode, err := _generateOpCode(compileEnv, value, nowStartLine)

		if err != nil {
			return nil, 0, err
		}

		//opCodes = append(opCodes, NewSymbol(fmt.Sprintf("set %s", symbol.(Symbol).GetSymbolIndex())))

		if err != nil {
			return nil, 0, err
		}

		opCodes = append(opCodes, CreateSetInstr(symbol.GetSymbolIndex()))

		return opCodes, affectedCode + 1, nil
	case "define":
		if 2 != cellArrLen {
			return nil, 0, errors.New("Invalid syntax 4")
		}
		symbol, ok := cellArr[0].(Symbol)

		if !ok {
			return nil, 0, errors.New("Invalid syntax 4")
		}

		value := cellArr[1]
		opCodes, affectedCode, err := _generateOpCode(compileEnv, value, nowStartLine)

		if err != nil {
			return nil, 0, err
		}

		//opCodes = append(opCodes, NewSymbol(fmt.Sprintf("define %s", symbol.(Symbol).GetSymbolIndex())))
		opCodes = append(opCodes, CreateDefineInstr(uint64(symbol)))

		return opCodes, affectedCode + 1, nil

	case "lambda":
		if 2 != cellArrLen {
			return nil, 0, errors.New("Invalid syntax 4")
		}

		//opCode := []SExpression{NewSymbol("new-env")}
		opCode := []Instr{CreateNewEnvInstr()}
		opCodeLine := nowStartLine + 1

		//(a b c)
		vars, varslen := ToArraySexp(cellArr[0])

		for i := int64(0); i < varslen; i++ {
			//opCode = append(opCode, NewSymbol(fmt.Sprintf("define-args %s", vars[i].(Symbol).GetSymbolIndex())))
			opCode = append(opCode, CreateDefineArgsInstr(uint64(vars[i].(Symbol))))
			opCodeLine += 1
		}

		rawBody := cellArr[1]

		//opCode = append(opCode, NewSymbol("create-lambda-dummy arg-len this-stack-instr func-instr-size"))
		opCode = append(opCode, CreateDummyInstr())

		createFuncOpCodeLine := opCodeLine
		opCodeLine += 1

		funcOpCode, funcOpCodeAffectLow, err := _generateOpCode(compileEnv, rawBody, 0)
		if err != nil {
			return nil, 0, err
		}

		opCode = append(opCode, funcOpCode...)
		opCodeLine += funcOpCodeAffectLow

		//opCode[createFuncOpCodeLine-nowStartLine] = NewSymbol(fmt.Sprintf("create-lambda %d %d", varslen, funcOpCodeAffectLow+1))
		opCode[createFuncOpCodeLine-nowStartLine] = CreateCreateLambdaInstr(varslen, funcOpCodeAffectLow+1)

		//opCode = append(opCode, NewSymbol("ret"))
		opCode = append(opCode, CreateRetInstr())

		return opCode, opCodeLine - nowStartLine + 1, nil //+1 is return instr count

	case "loop":
		if 2 != cellArrLen {
			return nil, 0, errors.New("Invalid syntax 4")
		}

		cond := cellArr[0]
		body := cellArr[1]

		//cond|jmp-else|body|jmp|...

		startIndex := nowStartLine
		condOpCode, condAffectedCode, err := _generateOpCode(compileEnv, cond, nowStartLine)
		if err != nil {
			return nil, 0, err
		}

		bodyOpCode, bodyAffectedCode, err := _generateOpCode(compileEnv, body, nowStartLine+condAffectedCode+1)
		if err != nil {
			return nil, 0, err
		}

		//opCode := append(condOpCode, NewSymbol(fmt.Sprintf("jmp-else-dummy %d", nowStartLine+condAffectedCode+1+bodyAffectedCode)))
		opCode := append(condOpCode, CreateDummyInstr())

		dummyIndex := condAffectedCode
		opCode = append(opCode, bodyOpCode...)

		//opCode = append(opCode, NewSymbol(fmt.Sprintf("jmp %d", startIndex)))
		opCode = append(opCode, CreateJmpInstr(startIndex))

		//opCode[dummyIndex] = NewSymbol(fmt.Sprintf("jmp-else %d", nowStartLine+condAffectedCode+1+bodyAffectedCode+1))
		opCode[dummyIndex] = CreateJmpElseInstr(nowStartLine + condAffectedCode + 1 + bodyAffectedCode + 1)

		return opCode, condAffectedCode + 1 + bodyAffectedCode + 1, nil
	}

	//if IsEmptyList(cell.GetCdr()) {
	//	var carAffectedCode int64
	//	carOpCode, carAffectedCode = _generateOpCode(cell.GetCar(), nowStartLine)
	//	return carOpCode, carAffectedCode
	//}
	var carOpCode []Instr
	args, argsLen := ToArraySexp(cell.GetCdr())
	var cdrOpCode []Instr
	affectedCdrOpeCodeRowCount := nowStartLine
	for i := int64(0); i < argsLen; i++ {
		argsOpCode, argsOpCodeAffectedRowCount, err := _generateOpCode(compileEnv, args[i], affectedCdrOpeCodeRowCount)
		if err != nil {
			return nil, 0, err
		}

		cdrOpCode = append(cdrOpCode, argsOpCode...)
		affectedCdrOpeCodeRowCount += argsOpCodeAffectedRowCount
	}

	var carAffectedCode int64

	if IsNativeFunc(compileEnv, cell.GetCar()) {
		//carOpCode = []SExpression{NewSymbol(fmt.Sprintf("%s %d", cell.GetCar(), argsLen))}
		funcName := cell.GetCar().(Symbol).String(compileEnv)
		tartgetFunc := NativeFuncNameToOpCodeMap[funcName]
		if nil == tartgetFunc {
			panic("Invalid syntax 7")
		}
		carOpCode = []Instr{tartgetFunc(argsLen)}

		carAffectedCode = 1
		cdrAffectedCode := affectedCdrOpeCodeRowCount - nowStartLine
		return append(cdrOpCode, carOpCode...), carAffectedCode + cdrAffectedCode, nil
	} else {
		var err error
		carOpCode, carAffectedCode, err = _generateOpCode(compileEnv, cell.GetCar(), affectedCdrOpeCodeRowCount)
		if err != nil {
			return nil, 0, err
		}
		cdrAffectedCode := affectedCdrOpeCodeRowCount - nowStartLine

		//return append(append(cdrOpCode, carOpCode...), NewSymbol(fmt.Sprintf("call %d", argsLen))), carAffectedCode + cdrAffectedCode + 1
		return append(append(cdrOpCode, carOpCode...), CreateCallInstr(argsLen)), carAffectedCode + cdrAffectedCode + 1, nil
	}
}
