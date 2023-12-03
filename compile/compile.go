package compile

import (
	"testrand-vm/instr"
	"testrand-vm/reader"
)

func ToArraySexp(sexp reader.SExpression) ([]reader.SExpression, int64) {
	var result []reader.SExpression
	var count int64 = 0
	for {
		if reader.SExpressionTypeConsCell != sexp.SExpressionTypeId() {
			break
		}
		if reader.IsEmptyList(sexp) {
			break
		}
		cell := sexp.(reader.ConsCell)
		result = append(result, cell.GetCar())
		count++
		sexp = cell.GetCdr()
	}
	return result, count
}

func IsNativeFunc(sexp reader.SExpression) bool {
	if reader.SExpressionTypeSymbol != sexp.SExpressionTypeId() {
		return false
	}
	switch sexp.(reader.Symbol).GetValue() {
	case
		"+",
		"-",
		"*",
		"/",
		"%",
		">",
		"<",
		">=",
		"<=",
		"=",
		"and",
		"or",
		"not",
		"eq?",
		"println",
		"print",
		"car",
		"cdr",
		"cons",
		"random-id",
		"array",
		"array-get",
		"array-set",
		"array-len",
		"array-push",
		"map",
		"map-get",
		"map-set",
		"map-len",
		"map-keys",
		"call/cc":
		return true
	}
	return false
}

func GenerateOpCode(sexp reader.SExpression, nowStartLine int64) ([]instr.Instr, int64) {
	codes, leng := _generateOpCode(sexp, nowStartLine)
	//return append(codes, reader.NewSymbol("end-code")), leng + 1
	return append(codes, instr.CreateEndCodeInstr()), leng + 1
}

func _generateOpCode(sexp reader.SExpression, nowStartLine int64) ([]instr.Instr, int64) {
	switch sexp.SExpressionTypeId() {
	case reader.SExpressionTypeSymbol:
		//return []reader.SExpression{
		//		reader.NewSymbol(fmt.Sprintf("push-sym %s", sexp.String())),
		//		reader.NewSymbol("load"),
		//	},
		//	2
		return []instr.Instr{instr.CreatePushSymbolInstr(sexp.(reader.Symbol).GetValue()), instr.CreateLoadInstr()}, 2
	case reader.SExpressionTypeNumber:
		//return []reader.SExpression{reader.NewSymbol(fmt.Sprintf("push-num %s", sexp.String()))}, 1
		return []instr.Instr{instr.CreatePushNumberInstr(sexp.(reader.Number).GetValue())}, 1
	case reader.SExpressionTypeBool:
		//return []reader.SExpression{reader.NewSymbol(fmt.Sprintf("push-boo %s", sexp.String()))}, 1
		return []instr.Instr{instr.CreatePushBoolInstr(sexp.(reader.Bool).GetValue())}, 1
	case reader.SExpressionTypeString:
		//return []reader.SExpression{reader.NewSymbol(fmt.Sprintf("push-str %s", sexp.(reader.Str).GetValue()))}, 1
		return []instr.Instr{instr.CreatePushStringInstr(sexp.(reader.Str).GetValue())}, 1
	}

	cell := sexp.(reader.ConsCell)

	label := cell.GetCar()

	if reader.SExpressionTypeSymbol != label.SExpressionTypeId() {
		carOpCode, carAffectedCode := _generateOpCode(cell.GetCar(), nowStartLine)
		if reader.IsEmptyList(cell.GetCdr()) {
			return carOpCode, carAffectedCode
		}

		_, argsLen := ToArraySexp(cell.GetCdr())
		cdrOpCode, cdrAffectedCode := _generateOpCode(cell.GetCdr(), nowStartLine+carAffectedCode)
		//return append(append(cdrOpCode, carOpCode...), reader.NewSymbol(fmt.Sprintf("call %d", argsLen))), carAffectedCode + cdrAffectedCode + 1
		return append(append(cdrOpCode, carOpCode...), instr.CreateCallInstr(argsLen)), carAffectedCode + cdrAffectedCode + 1
	}

	cellContent := cell.GetCdr().(reader.ConsCell)
	cellArr, cellArrLen := ToArraySexp(cellContent)

	switch label.(reader.Symbol).GetValue() {
	case "quote":
		if cellArrLen != 1 {
			panic("Invalid Syntax Quote")
		}
		//return []reader.SExpression{reader.NewSymbol(fmt.Sprintf("load-sexp %s\n", cellArr[0]))}, 1
		return []instr.Instr{instr.CreatePushSExpressionInstr(cellArr[0])}, 1

	//case "loop":
	//	if 2 != cellArrLen {
	//		panic("Invalid syntax 2")
	//	}
	//	// cond-opcode(?)|jump-(1)|loop-body-opcode(?)|jump-lable(1)
	//	loopCond := cellArr[0]
	//	condOpCode, condAffectedCode := _generateOpCode(loopCond, nowStartLine)
	//	loopBody := cellArr[1]
	//	bodyOpCode, bodyAffectedCode := _generateOpCode(loopBody, nowStartLine+condAffectedCode+1)
	//	return append(append(condOpCode, reader.NewSymbol(fmt.Sprintf("jump %d", nowStartLine+condAffectedCode+bodyAffectedCode+2))), append(bodyOpCode, reader.NewSymbol(fmt.Sprintf("jump-%d", nowStartLine)))...), condAffectedCode + bodyAffectedCode + 2
	case "begin":
		bodies, bodiesSize := ToArraySexp(cellContent)
		var result []instr.Instr
		var lineNum = nowStartLine
		var addedRows = int64(0)
		for i := int64(0); i < bodiesSize; i++ {
			bodiesOpCodes, affectedOpCodeLine := _generateOpCode(bodies[i], lineNum)
			lineNum += affectedOpCodeLine
			addedRows += affectedOpCodeLine
			result = append(result, bodiesOpCodes...)
			if i != bodiesSize-1 {
				lineNum += 1
				addedRows += 1
				//result = append(result, reader.NewSymbol("pop"))
				result = append(result, instr.CreatePopInstr())
			}
		}
		return result, int64(len(result))
	case "cond":
		condAndBody, condAndBodySize := ToArraySexp(cellContent)

		if 0 == condAndBodySize {
			panic("Invalid syntax 3")
		}

		var opCodes = []instr.Instr{}

		var lastIndexes = make([]int64, condAndBodySize)
		var indexesIndex = int64(0)

		nowLine := nowStartLine

		for i := int64(0); i < condAndBodySize; i++ {
			condAndBodyCell := condAndBody[i].(reader.ConsCell)
			condAndBodyCellArr, _ := ToArraySexp(condAndBodyCell)

			if 2 != len(condAndBodyCellArr) {
				panic("Invalid syntax 3")
			}
			condSexp := condAndBodyCellArr[0]
			bodySexp := condAndBodyCellArr[1]

			condOpCodes, condAffectedCode := _generateOpCode(condSexp, nowLine)

			bodyOpCodes, bodyAffectedCode := _generateOpCode(bodySexp, nowLine+condAffectedCode+1)

			opCodes = append(opCodes, make([]instr.Instr, condAffectedCode+bodyAffectedCode+2)...)

			for j := int64(0); j < condAffectedCode; j++ {
				opCodes[j+indexesIndex] = condOpCodes[j]
			}

			indexesIndex += condAffectedCode

			//opCodes[indexesIndex] = reader.NewSymbol(fmt.Sprintf("jmp-else %d", nowLine+condAffectedCode+bodyAffectedCode+2))
			opCodes[indexesIndex] = instr.CreateJmpElseInstr(nowLine + condAffectedCode + bodyAffectedCode + 2)

			indexesIndex += 1

			for j := int64(0); j < bodyAffectedCode; j++ {
				opCodes[j+indexesIndex] = bodyOpCodes[j]
			}

			indexesIndex += bodyAffectedCode

			//opCodes[indexesIndex] = reader.NewSymbol("temporary jump")
			opCodes[indexesIndex] = instr.CreateDummyInstr()
			lastIndexes[i] = indexesIndex

			indexesIndex += 1

			nowLine += condAffectedCode + bodyAffectedCode + 2
		}

		for i := int64(0); i < condAndBodySize; i++ {
			//opCodes[lastIndexes[i]] = reader.NewSymbol(fmt.Sprintf("jmp %d", nowLine))
			opCodes[lastIndexes[i]] = instr.CreateJmpInstr(nowLine)
		}

		return opCodes, int64(len(opCodes))

	case "and":
		cond, condLen := ToArraySexp(cellContent)

		if 0 == condLen {
			panic("Invalid syntax 3")
		}

		var opCodes = []instr.Instr{}

		affectedCode := nowStartLine

		for i := int64(0); i < condLen; i++ {
			condOpCodes, condAffectedCode := _generateOpCode(cond[i], affectedCode)
			affectedCode += condAffectedCode
			opCodes = append(opCodes, condOpCodes...)
		}

		//opCodes = append(opCodes, reader.NewSymbol(fmt.Sprintf("and %d", condLen)))
		opCodes = append(opCodes, instr.CreateAndInstr(condLen))

		return opCodes, affectedCode - nowStartLine + 1

	case "or":
		cond, condLen := ToArraySexp(cellContent)

		if 0 == condLen {
			panic("Invalid syntax 3")
		}

		var opCodes = []instr.Instr{}

		affectedCode := nowStartLine

		for i := int64(0); i < condLen; i++ {
			condOpCodes, condAffectedCode := _generateOpCode(cond[i], affectedCode)
			affectedCode += condAffectedCode
			opCodes = append(opCodes, condOpCodes...)
		}

		//opCodes = append(opCodes, reader.NewSymbol(fmt.Sprintf("or %d", condLen)))
		opCodes = append(opCodes, instr.CreateOrInstr(condLen))

		return opCodes, affectedCode - nowStartLine + 1

	case "set":
		if 2 != len(cellArr) {
			panic("Invalid syntax 4")
		}
		symbol := cellArr[0]
		value := cellArr[1]

		opCodes, affectedCode := _generateOpCode(value, nowStartLine)

		if symbol.SExpressionTypeId() != reader.SExpressionTypeSymbol {
			panic("Invalid syntax 4")
		}

		//opCodes = append(opCodes, reader.NewSymbol(fmt.Sprintf("set %s", symbol.(reader.Symbol).GetValue())))
		opCodes = append(opCodes, instr.CreateSetInstr(symbol.(reader.Symbol).GetValue()))

		return opCodes, affectedCode + 1
	case "define":
		if 2 != cellArrLen {
			panic("Invalid syntax 4")
		}
		symbol := cellArr[0]
		value := cellArr[1]

		opCodes, affectedCode := _generateOpCode(value, nowStartLine)

		if symbol.SExpressionTypeId() != reader.SExpressionTypeSymbol {
			panic("Invalid syntax 4")
		}

		//opCodes = append(opCodes, reader.NewSymbol(fmt.Sprintf("define %s", symbol.(reader.Symbol).GetValue())))
		opCodes = append(opCodes, instr.CreateDefineInstr(symbol.(reader.Symbol).GetValue()))

		return opCodes, affectedCode + 1

	case "lambda":
		if 2 != cellArrLen {
			panic("Invalid syntax 5")
		}

		//opCode := []reader.SExpression{reader.NewSymbol("new-env")}
		opCode := []instr.Instr{instr.CreateNewEnvInstr()}
		opCodeLine := nowStartLine + 1

		//(a b c)
		vars, varslen := ToArraySexp(cellArr[0])

		for i := int64(0); i < varslen; i++ {
			//opCode = append(opCode, reader.NewSymbol(fmt.Sprintf("define-args %s", vars[i].(reader.Symbol).GetValue())))
			opCode = append(opCode, instr.CreateDefineArgsInstr(vars[i].(reader.Symbol).GetValue()))
			opCodeLine += 1
		}

		rawBody := cellArr[1]

		//opCode = append(opCode, reader.NewSymbol("create-lambda-dummy arg-len this-stack-instr func-instr-size"))
		opCode = append(opCode, instr.CreateDummyInstr())

		createFuncOpCodeLine := opCodeLine
		opCodeLine += 1

		funcOpCode, funcOpCodeAffectLow := _generateOpCode(rawBody, 0)
		opCode = append(opCode, funcOpCode...)
		opCodeLine += funcOpCodeAffectLow

		//opCode[createFuncOpCodeLine-nowStartLine] = reader.NewSymbol(fmt.Sprintf("create-lambda %d %d", varslen, funcOpCodeAffectLow+1))
		opCode[createFuncOpCodeLine-nowStartLine] = instr.CreateCreateLambdaInstr(varslen, funcOpCodeAffectLow+1)

		//opCode = append(opCode, reader.NewSymbol("ret"))
		opCode = append(opCode, instr.CreateRetInstr())

		return opCode, opCodeLine - nowStartLine + 1 //+1 is return instr count

	case "loop":
		if 2 != cellArrLen {
			panic("Invalid syntax 6")
		}

		cond := cellArr[0]
		body := cellArr[1]

		//cond|jmp-else|body|jmp|...

		startIndex := nowStartLine
		condOpCode, condAffectedCode := _generateOpCode(cond, nowStartLine)
		bodyOpCode, bodyAffectedCode := _generateOpCode(body, nowStartLine+condAffectedCode+1)

		//opCode := append(condOpCode, reader.NewSymbol(fmt.Sprintf("jmp-else-dummy %d", nowStartLine+condAffectedCode+1+bodyAffectedCode)))
		opCode := append(condOpCode, instr.CreateDummyInstr())

		dummyIndex := condAffectedCode
		opCode = append(opCode, bodyOpCode...)

		//opCode = append(opCode, reader.NewSymbol(fmt.Sprintf("jmp %d", startIndex)))
		opCode = append(opCode, instr.CreateJmpInstr(startIndex))

		//opCode[dummyIndex] = reader.NewSymbol(fmt.Sprintf("jmp-else %d", nowStartLine+condAffectedCode+1+bodyAffectedCode+1))
		opCode[dummyIndex] = instr.CreateJmpElseInstr(nowStartLine + condAffectedCode + 1 + bodyAffectedCode + 1)

		return opCode, condAffectedCode + 1 + bodyAffectedCode + 1
	}

	//if reader.IsEmptyList(cell.GetCdr()) {
	//	var carAffectedCode int64
	//	carOpCode, carAffectedCode = _generateOpCode(cell.GetCar(), nowStartLine)
	//	return carOpCode, carAffectedCode
	//}
	var carOpCode []instr.Instr
	args, argsLen := ToArraySexp(cell.GetCdr())
	var cdrOpCode []instr.Instr
	affectedCdrOpeCodeRowCount := nowStartLine
	for i := int64(0); i < argsLen; i++ {
		argsOpCode, argsOpCodeAffectedRowCount := _generateOpCode(args[i], affectedCdrOpeCodeRowCount)
		cdrOpCode = append(cdrOpCode, argsOpCode...)
		affectedCdrOpeCodeRowCount += argsOpCodeAffectedRowCount
	}

	var carAffectedCode int64

	if IsNativeFunc(cell.GetCar()) {
		//carOpCode = []reader.SExpression{reader.NewSymbol(fmt.Sprintf("%s %d", cell.GetCar(), argsLen))}
		funcName := cell.GetCar().(reader.Symbol).GetValue()
		tartgetFunc := instr.NativeFuncNameToOpCodeMap[funcName]
		if nil == tartgetFunc {
			panic("Invalid syntax 7")
		}
		carOpCode = []instr.Instr{tartgetFunc(argsLen)}

		carAffectedCode = 1
		cdrAffectedCode := affectedCdrOpeCodeRowCount - nowStartLine
		return append(cdrOpCode, carOpCode...), carAffectedCode + cdrAffectedCode
	} else {
		carOpCode, carAffectedCode = _generateOpCode(cell.GetCar(), affectedCdrOpeCodeRowCount)
		cdrAffectedCode := affectedCdrOpeCodeRowCount - nowStartLine

		//return append(append(cdrOpCode, carOpCode...), reader.NewSymbol(fmt.Sprintf("call %d", argsLen))), carAffectedCode + cdrAffectedCode + 1
		return append(append(cdrOpCode, carOpCode...), instr.CreateCallInstr(argsLen)), carAffectedCode + cdrAffectedCode + 1
	}
}
