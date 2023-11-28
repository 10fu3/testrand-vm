package compile

import (
	"fmt"
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

func GenerateOpCode(sexp reader.SExpression, nowStartLine int64) ([]reader.SExpression, int64) {
	switch sexp.SExpressionTypeId() {
	case reader.SExpressionTypeSymbol:
		return []reader.SExpression{
				reader.NewSymbol(fmt.Sprintf("push-sym %s", sexp.String())),
				reader.NewSymbol("load"),
			},
			2
	case reader.SExpressionTypeNumber:
		return []reader.SExpression{reader.NewSymbol(fmt.Sprintf("push-num %s", sexp.String()))}, 1
	case reader.SExpressionTypeBool:
		return []reader.SExpression{reader.NewSymbol(fmt.Sprintf("push-boo %s", sexp.String()))}, 1
	case reader.SExpressionTypeString:
		return []reader.SExpression{reader.NewSymbol(fmt.Sprintf("push-str %s", sexp.String()))}, 1
	}

	cell := sexp.(reader.ConsCell)

	label := cell.GetCar()

	if reader.SExpressionTypeSymbol != label.SExpressionTypeId() {
		carOpCode, carAffectedCode := GenerateOpCode(cell.GetCar(), nowStartLine)
		if reader.IsEmptyList(cell.GetCdr()) {
			return carOpCode, carAffectedCode
		}
		cdrOpCode, cdrAffectedCode := GenerateOpCode(cell.GetCdr(), nowStartLine+carAffectedCode)

		return append(cdrOpCode, carOpCode...), carAffectedCode + cdrAffectedCode
	}

	cellContent := cell.GetCdr().(reader.ConsCell)
	cellArr, cellArrLen := ToArraySexp(cellContent)

	switch label.(reader.Symbol).GetValue() {
	case "loop":
		if 2 != cellArrLen {
			panic("Invalid syntax 2")
		}
		// cond-opcode(?)|jump-(1)|loop-body-opcode(?)|jump-lable(1)
		loopCond := cellArr[0]
		condOpCode, condAffectedCode := GenerateOpCode(loopCond, nowStartLine)
		loopBody := cellArr[1]
		bodyOpCode, bodyAffectedCode := GenerateOpCode(loopBody, nowStartLine+condAffectedCode+1)
		return append(append(condOpCode, reader.NewSymbol(fmt.Sprintf("jump %d", nowStartLine+condAffectedCode+bodyAffectedCode+2))), append(bodyOpCode, reader.NewSymbol(fmt.Sprintf("jump-%d", nowStartLine)))...), condAffectedCode + bodyAffectedCode + 2
	case "begin":
		bodies, bodiesSize := ToArraySexp(cellContent)
		var result []reader.SExpression
		var lineNum = nowStartLine
		var addedRows = int64(0)
		for i := int64(0); i < bodiesSize; i++ {
			bodiesOpCodes, affectedOpCodeLine := GenerateOpCode(bodies[i], lineNum)
			lineNum += affectedOpCodeLine
			addedRows += affectedOpCodeLine
			result = append(result, bodiesOpCodes...)
			if i != bodiesSize-1 {
				lineNum += 1
				addedRows += 1
				result = append(result, reader.NewSymbol("pop"))
			}
		}
		result = append(result, reader.NewSymbol("ret"))
		return result, addedRows + 1
	case "cond":
		condAndBody, condAndBodySize := ToArraySexp(cellContent)

		if 0 == condAndBodySize {
			panic("Invalid syntax 3")
		}

		var opCodes = []reader.SExpression{}

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

			condOpCodes, condAffectedCode := GenerateOpCode(condSexp, nowLine)

			bodyOpCodes, bodyAffectedCode := GenerateOpCode(bodySexp, nowLine+condAffectedCode+1)

			opCodes = append(opCodes, make([]reader.SExpression, condAffectedCode+bodyAffectedCode+2)...)

			for j := int64(0); j < condAffectedCode; j++ {
				opCodes[j+indexesIndex] = condOpCodes[j]
			}

			indexesIndex += condAffectedCode

			opCodes[indexesIndex] = reader.NewSymbol(fmt.Sprintf("jump-else %d", nowLine+condAffectedCode+bodyAffectedCode+2))

			indexesIndex += 1

			for j := int64(0); j < bodyAffectedCode; j++ {
				opCodes[j+indexesIndex] = bodyOpCodes[j]
			}

			indexesIndex += bodyAffectedCode

			opCodes[indexesIndex] = reader.NewSymbol("temporary jump")
			lastIndexes[i] = indexesIndex

			indexesIndex += 1

			nowLine += condAffectedCode + bodyAffectedCode + 1
		}

		for i := int64(0); i < condAndBodySize; i++ {
			opCodes[lastIndexes[i]] = reader.NewSymbol(fmt.Sprintf("jump-%d", nowLine+int64(len(opCodes))))
		}

		return opCodes, int64(len(opCodes))

	case "and":
		cond, condLen := ToArraySexp(cellContent)

		if 0 == condLen {
			panic("Invalid syntax 3")
		}

		var opCodes = []reader.SExpression{}

		affectedCode := nowStartLine

		for i := int64(0); i < condLen; i++ {
			condOpCodes, condAffectedCode := GenerateOpCode(cond[i], affectedCode)
			affectedCode += condAffectedCode
			opCodes = append(opCodes, condOpCodes...)
		}

		opCodes = append(opCodes, reader.NewSymbol(fmt.Sprintf("args-len %d", condLen)))
		opCodes = append(opCodes, reader.NewSymbol("and"))

		return opCodes, affectedCode - nowStartLine + 2

	case "or":
		cond, condLen := ToArraySexp(cellContent)

		if 0 == condLen {
			panic("Invalid syntax 3")
		}

		var opCodes = []reader.SExpression{}

		affectedCode := nowStartLine

		for i := int64(0); i < condLen; i++ {
			condOpCodes, condAffectedCode := GenerateOpCode(cond[i], affectedCode)
			affectedCode += condAffectedCode
			opCodes = append(opCodes, condOpCodes...)
		}

		opCodes = append(opCodes, reader.NewSymbol(fmt.Sprintf("args-len %d", condLen)))
		opCodes = append(opCodes, reader.NewSymbol("or"))

		return opCodes, affectedCode - nowStartLine + 2

	case "set":
		if 2 != len(cellArr) {
			panic("Invalid syntax 4")
		}
	// set-opcode(?)|jump-(1)

	case "define":
		if 2 != cellArrLen {
			panic("Invalid syntax 4")
		}
		symbol := cellArr[0]
		value := cellArr[1]

		opCodes, affectedCode := GenerateOpCode(value, nowStartLine)

		if symbol.SExpressionTypeId() != reader.SExpressionTypeSymbol {
			panic("Invalid syntax 4")
		}

		opCodes = append(opCodes, reader.NewSymbol(fmt.Sprintf("define %s", symbol.(reader.Symbol).GetValue())))

		return opCodes, affectedCode + 1
	}

	var carOpCode []reader.SExpression

	if reader.IsEmptyList(cell.GetCdr()) {
		var carAffectedCode int64
		carOpCode, carAffectedCode = GenerateOpCode(cell.GetCar(), nowStartLine)
		return carOpCode, carAffectedCode
	}

	args, argsLen := ToArraySexp(cell.GetCdr())
	var cdrOpCode []reader.SExpression
	affectedCdrOpeCodeRowCount := nowStartLine
	for i := int64(0); i < argsLen; i++ {
		argsOpCode, argsOpCodeAffectedRowCount := GenerateOpCode(args[i], affectedCdrOpeCodeRowCount)
		cdrOpCode = append(cdrOpCode, argsOpCode...)
		affectedCdrOpeCodeRowCount += argsOpCodeAffectedRowCount
	}

	carOpCode, carAffectedCode := GenerateOpCode(cell.GetCar(), affectedCdrOpeCodeRowCount)

	cdrAffectedCode := affectedCdrOpeCodeRowCount - nowStartLine
	return append(append(cdrOpCode, carOpCode...), reader.NewSymbol("call")), carAffectedCode + cdrAffectedCode + 1
}
