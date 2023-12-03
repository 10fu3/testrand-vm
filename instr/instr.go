package instr

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"strings"
	"testrand-vm/reader"
)

// instruction data structure
// all byte array
// 1st byte: length of instruction
// 2nd byte: instruction type
// 3rd... byte: instruction data

type Instr struct {
	Length uint64
	Type   uint8
	Data   []byte
}

func (i Instr) String() string {
	switch i.Type {
	case OPCODE_POP:
		return "Instr{Type: POP}"
	case OPCODE_PUSH_NUM:
		return fmt.Sprintf("Instr{Type: PUSH_NUM, Data: %d}", DeserializePushNumberInstr(i))
	case OPCODE_PUSH_STR:
		return fmt.Sprintf("Instr{Type: PUSH_STR, Data: %s}", DeserializePushStringInstr(i))
	case OPCODE_PUSH_SYM:
		return fmt.Sprintf("Instr{Type: PUSH_SYM, Data: %s}", DeserializePushSymbolInstr(i))
	case OPCODE_PUSH_TRUE:
		return "Instr{Type: PUSH_TRUE}"
	case OPCODE_PUSH_FALSE:
		return "Instr{Type: PUSH_FALSE}"
	case OPCODE_PUSH_NIL:
		return "Instr{Type: PUSH_NIL}"
	case OPCODE_PUSH_SEXP:
		sexp, err := DeserializeSexpressionInstr(i)
		if err != nil {
			panic(err)
		}
		return fmt.Sprintf("Instr{Type: PUSH_SEXP, Data: %s}", sexp)
	case OPCODE_JMP:
		return fmt.Sprintf("Instr{Type: JMP, Data: %d}", DeserializeJmpInstr(i))
	case OPCODE_JMP_IF:
		return fmt.Sprintf("Instr{Type: JMP_IF, Data: %d}", DeserializeJmpIfInstr(i))
	case OPCODE_JMP_ELSE:
		return fmt.Sprintf("Instr{Type: JMP_ELSE, Data: %d}", DeserializeJmpElseInstr(i))
	case OPCODE_LOAD:
		return "Instr{Type: LOAD}"
	case OPCODE_DEFINE:
		return fmt.Sprintf("Instr{Type: DEFINE, Data: %s}", DeserializeDefineInstr(i))
	case OPCODE_DEFINE_ARGS:
		return fmt.Sprintf("Instr{Type: DEFINE_ARGS, Data: %s}", DeserializeDefineArgsInstr(i))
	case OPCODE_SET:
		return fmt.Sprintf("Instr{Type: SET, Data: %s}", DeserializeSetInstr(i))
	case OPCODE_NEW_ENV:
		return "Instr{Type: NEW_ENV}"
	case OPCODE_CREATE_CLOSURE:
		varslen, funcOpAffectedCode := DeserializeCreateClosureInstr(i)
		return fmt.Sprintf("Instr{Type: CREATE_CLOSURE, Data: %d, %d}", varslen, funcOpAffectedCode)
	case OPCODE_CALL:
		return fmt.Sprintf("Instr{Type: CALL, Data: %d}", DeserializeCallInstr(i))
	case OPCODE_RETURN:
		return "Instr{Type: RETURN}"
	case OPCODE_AND:
		return fmt.Sprintf("Instr{Type: AND, ArgsSize: %d}", DeserializeAndInstr(i))
	case OPCODE_OR:
		return fmt.Sprintf("Instr{Type: OR, ArgsSize: %d}", DeserializeOrInstr(i))
	case OPCODE_PRINT:
		return fmt.Sprintf("Instr{Type: PRINT, ArgsSize: %d}", DeserializePrintInstr(i))
	case OPCODE_PRINTLN:
		return fmt.Sprintf("Instr{Type: PRINTLN, ArgsSize: %d}", DeserializePrintlnInstr(i))
	case OPCODE_PLUS_NUM:
		return fmt.Sprintf("Instr{Type: PLUS_NUM, ArgsSize: %d}", DeserializePlusNumInstr(i))
	case OPCODE_MINUS_NUM:
		return fmt.Sprintf("Instr{Type: MINUS_NUM, ArgsSize: %d}", DeserializeMinusNumInstr(i))
	case OPCODE_MULTIPLY_NUM:
		return fmt.Sprintf("Instr{Type: MULTIPLY_NUM, ArgsSize: %d}", DeserializeMultiplyNumInstr(i))
	case OPCODE_DIVIDE_NUM:
		return fmt.Sprintf("Instr{Type: DIVIDE_NUM, ArgsSize: %d}", DeserializeDivideNumInstr(i))
	case OPCODE_MODULO_NUM:
		return fmt.Sprintf("Instr{Type: MODULO_NUM, ArgsSize: %d}", DeserializeModuloNumInstr(i))
	case OPCODE_EQUAL_NUM:
		return fmt.Sprintf("Instr{Type: EQUAL_NUM, ArgsSize: %d}", DeserializeEqualNumInstr(i))
	case OPCODE_NOT_EQUAL_NUM:
		return fmt.Sprintf("Instr{Type: NOT_EQUAL_NUM, ArgsSize: %d}", DeserializeNotEqualNumInstr(i))
	case OPCODE_GREATER_THAN_NUM:
		return fmt.Sprintf("Instr{Type: GREATER_THAN_NUM, ArgsSize: %d}", DeserializeGreaterThanNumInstr(i))
	case OPCODE_GREATER_THAN_OR_EQUAL_NUM:
		return fmt.Sprintf("Instr{Type: GREATER_THAN_OR_EQUAL_NUM, ArgsSize: %d}", DeserializeGreaterThanOrEqualNumInstr(i))
	case OPCODE_LESS_THAN_NUM:
		return fmt.Sprintf("Instr{Type: LESS_THAN_NUM, ArgsSize: %d}", DeserializeLessThanNumInstr(i))
	case OPCODE_LESS_THAN_OR_EQUAL_NUM:
		return fmt.Sprintf("Instr{Type: LESS_THAN_OR_EQUAL_NUM, ArgsSize: %d}", DeserializeLessThanOrEqualNumInstr(i))
	case OPCODE_CAR:
		return "Instr{Type: CAR}"
	case OPCODE_CDR:
		return "Instr{Type: CDR}"
	case OPCODE_RANDOM_ID:
		return "Instr{Type: RANDOM_ID}"
	case OPCODE_NEW_ARRAY:
		return "Instr{Type: NEW_ARRAY}"
	case OPCODE_ARRAY_GET:
		return fmt.Sprintf("Instr{Type: ARRAY_GET}")
	case OPCODE_ARRAY_SET:
		return fmt.Sprintf("Instr{Type: ARRAY_SET}")
	case OPCODE_ARRAY_LENGTH:
		return fmt.Sprintf("Instr{Type: ARRAY_LENGTH}")
	case OPCODE_ARRAY_PUSH:
		return fmt.Sprintf("Instr{Type: ARRAY_PUSH}")
	case OPCODE_NEW_MAP:
		return fmt.Sprintf("Instr{Type: NEW_MAP}")
	case OPCODE_MAP_GET:
		return fmt.Sprintf("Instr{Type: MAP_GET}")
	case OPCODE_MAP_SET:
		return fmt.Sprintf("Instr{Type: MAP_SET}")
	case OPCODE_MAP_LENGTH:
		return fmt.Sprintf("Instr{Type: MAP_LENGTH}")
	case OPCODE_MAP_KEYS:
		return fmt.Sprintf("Instr{Type: MAP_KEYS}")
	case OPCODE_MAP_DELETE:
		return fmt.Sprintf("Instr{Type: MAP_DELETE}")
	case OPCODE_CALL_CC:
		return fmt.Sprintf("Instr{Type: CALL_CC,ArgsSize: %d}", DeserializeCallCCInstr(i))
	case OPCODE_END_CODE:
		return "Instr{Type: END_CODE}"
	case OPCODE_NOP:
		return "Instr{Type: NOP}"
	}
	return fmt.Sprintf("Instr{Length: %d, Type: %s, Data: %s}", i.Length, OpCodeMap[i.Type], string(i.Data))
}

func NewInstr(instrType uint8, data []byte) Instr {
	return Instr{
		Length: uint64(len(data)) + 2,
		Type:   instrType,
		Data:   data,
	}
}

func CreatePopInstr() Instr {
	return NewInstr(OPCODE_POP, []byte{})
}

func CreatePushNumberInstr(number int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(number))
	return NewInstr(OPCODE_PUSH_NUM, b)
}

func CreatePushStringInstr(str string) Instr {
	return NewInstr(OPCODE_PUSH_STR, []byte(str))
}

func CreatePushSymbolInstr(symbol string) Instr {
	return NewInstr(OPCODE_PUSH_SYM, []byte(symbol))
}

func CreatePushBoolInstr(boolean bool) Instr {
	if boolean {
		return NewInstr(OPCODE_PUSH_TRUE, []byte{})
	} else {
		return NewInstr(OPCODE_PUSH_FALSE, []byte{})
	}
}

func CreatePushNilInstr() Instr {
	return NewInstr(OPCODE_PUSH_NIL, []byte{})
}

func CreatePushSExpressionInstr(sexp reader.SExpression) Instr {
	return NewInstr(OPCODE_PUSH_SEXP, []byte(fmt.Sprintf("%s\n", sexp)))
}

func CreateJmpInstr(jmpTo int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(jmpTo))
	return NewInstr(OPCODE_JMP, b)
}

func CreateJmpIfInstr(jmpTo int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(jmpTo))
	return NewInstr(OPCODE_JMP_IF, b)
}

func CreateJmpElseInstr(jmpTo int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(jmpTo))
	return NewInstr(OPCODE_JMP_ELSE, b)
}

func CreateLoadInstr() Instr {
	return NewInstr(OPCODE_LOAD, []byte{})
}

func CreateDefineInstr(symbol string) Instr {
	return NewInstr(OPCODE_DEFINE, []byte(symbol))
}

func CreateDefineArgsInstr(symbol string) Instr {
	return NewInstr(OPCODE_DEFINE_ARGS, []byte(symbol))
}

func CreateDummyInstr() Instr {
	return NewInstr(OPCODE_NOP, []byte{})
}

func CreateCreateLambdaInstr(varslen, funcOpAffectedCode int64) Instr {
	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, uint64(varslen))
	binary.LittleEndian.PutUint64(b[8:], uint64(funcOpAffectedCode))
	return NewInstr(OPCODE_CREATE_CLOSURE, b)
}

func CreateRetInstr() Instr {
	return NewInstr(OPCODE_RETURN, []byte{})
}

func CreateSetInstr(symbol string) Instr {
	return NewInstr(OPCODE_SET, []byte(symbol))
}

func CreateNewEnvInstr() Instr {
	return NewInstr(OPCODE_NEW_ENV, []byte{})
}

type FunctionGenerateInstr func(argsSize int64) Instr

func CreateCallInstr(argslen int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argslen))
	return NewInstr(OPCODE_CALL, b)
}

func CreateAndInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_AND, b)
}

func CreateOrInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_OR, b)
}

func CreatePrintInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_PRINT, b)
}

func CreatePrintlnInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_PRINTLN, b)
}

func CreatePlusNumInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_PLUS_NUM, b)
}

func CreateMinusNumInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_MINUS_NUM, b)
}

func CreateMultiplyNumInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_MULTIPLY_NUM, b)
}

func CreateDivideNumInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_DIVIDE_NUM, b)
}

func CreateModuloNumInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_MODULO_NUM, b)
}

func CreateEqualNumInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_EQUAL_NUM, b)
}

func CreateNotEqualNumInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_NOT_EQUAL_NUM, b)
}

/*
* >
 */
func CreateGreaterThanNumInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_GREATER_THAN_NUM, b)
}

/*
* >=
 */
func CreateGreaterThanOrEqualNumInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_GREATER_THAN_OR_EQUAL_NUM, b)
}

/*
* <
 */
func CreateLessThanNumInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_LESS_THAN_NUM, b)
}

/*
* <=
 */
func CreateLessThanOrEqualNumInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_LESS_THAN_OR_EQUAL_NUM, b)
}

func CreateCarInstr(argsSize int64) Instr {
	return NewInstr(OPCODE_CAR, []byte{})
}

func CreateCdrInstr(argsSize int64) Instr {
	return NewInstr(OPCODE_CDR, []byte{})
}

func CreateRandomIdInstr(argsSize int64) Instr {
	return NewInstr(OPCODE_RANDOM_ID, []byte{})
}

func CreateNewArrayInstr(argsSize int64) Instr {
	return NewInstr(OPCODE_NEW_ARRAY, []byte{})
}

func CreateArrayGetInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_ARRAY_GET, b)
}

func CreateArraySetInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_ARRAY_SET, b)
}

func CreateArrayLengthInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_ARRAY_LENGTH, b)
}

func CreateArrayPushInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_ARRAY_PUSH, b)
}

func CreateNewMapInstr(argsSize int64) Instr {
	return NewInstr(OPCODE_NEW_MAP, []byte{})
}

func CreateMapGetInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_MAP_GET, b)
}

func CreateMapSetInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_MAP_SET, b)
}

func CreateMapLengthInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_MAP_LENGTH, b)
}

func CreateMapKeysInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_MAP_KEYS, b)
}

func CreateMapDeleteInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_MAP_DELETE, b)
}

func CreateCallCCInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_CALL_CC, b)
}

var NativeFuncNameToOpCodeMap = map[string]FunctionGenerateInstr{
	"print":      CreatePrintInstr,
	"println":    CreatePrintlnInstr,
	"+":          CreatePlusNumInstr,
	"-":          CreateMinusNumInstr,
	"*":          CreateMultiplyNumInstr,
	"/":          CreateDivideNumInstr,
	"%":          CreateModuloNumInstr,
	"=":          CreateEqualNumInstr,
	"!=":         CreateNotEqualNumInstr,
	">":          CreateGreaterThanNumInstr,
	">=":         CreateGreaterThanOrEqualNumInstr,
	"<":          CreateLessThanNumInstr,
	"<=":         CreateLessThanOrEqualNumInstr,
	"car":        CreateCarInstr,
	"cdr":        CreateCdrInstr,
	"random-id":  CreateRandomIdInstr,
	"array":      CreateNewArrayInstr,
	"array-get":  CreateArrayGetInstr,
	"array-set":  CreateArraySetInstr,
	"array-len":  CreateArrayLengthInstr,
	"array-push": CreateArrayPushInstr,
	"map":        CreateNewMapInstr,
	"map-get":    CreateMapGetInstr,
	"map-set":    CreateMapSetInstr,
	"map-len":    CreateMapLengthInstr,
	"map-keys":   CreateMapKeysInstr,
	"map-delete": CreateMapDeleteInstr,
	"call/cc":    CreateCallCCInstr,
}

func CreateEndCodeInstr() Instr {
	return NewInstr(OPCODE_END_CODE, []byte{})
}

func Serialize(instr []Instr) []byte {
	dataLen := uint64(0)
	for i := 0; i < len(instr); i++ {
		dataLen += instr[i].Length + 8
	}
	// instrLen + instrLen * (length + type + data)
	data := make([]byte, dataLen+8)

	instrLen := make([]byte, 8)
	binary.LittleEndian.PutUint64(instrLen, uint64(len(instr)))
	data[0] = instrLen[0]
	data[1] = instrLen[1]
	data[2] = instrLen[2]
	data[3] = instrLen[3]
	data[4] = instrLen[4]
	data[5] = instrLen[5]
	data[6] = instrLen[6]
	data[7] = instrLen[7]

	offset := uint64(8)
	for i := 0; i < len(instr); i++ {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, instr[i].Length)
		data[offset] = b[0]
		data[offset+1] = b[1]
		data[offset+2] = b[2]
		data[offset+3] = b[3]
		data[offset+4] = b[4]
		data[offset+5] = b[5]
		data[offset+6] = b[6]
		data[offset+7] = b[7]
		offset += 8
		data[offset] = instr[i].Type
		offset++
		for j := 0; j < len(instr[i].Data); j++ {
			data[offset] = instr[i].Data[j]
			offset++
		}
	}
	return data
}

func Deserialize(data []byte) []Instr {
	instrLen := binary.LittleEndian.Uint64(data[0:8])
	instr := make([]Instr, instrLen)
	offset := uint64(8)
	for i := uint64(0); i < instrLen; i++ {
		instr[i].Length = binary.LittleEndian.Uint64(data[offset : offset+8])
		offset += 8
		instr[i].Type = data[offset]
		offset++
		instr[i].Data = make([]byte, instr[i].Length-2)
		for j := uint64(0); j < instr[i].Length-2; j++ {
			instr[i].Data[j] = data[offset]
			offset++
		}
	}
	return instr
}

func DeserializePushNumberInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializePushStringInstr(data Instr) string {
	return string(data.Data)
}

func DeserializePushSymbolInstr(data Instr) reader.Symbol {
	return reader.NewSymbol(string(data.Data))
}

func DeserializeSexpressionInstr(data Instr) (reader.SExpression, error) {
	sample := strings.NewReader(string(data.Data))
	r := bufio.NewReader(sample)
	sexp, err := reader.NewReader(r).Read()

	if err != nil {
		return nil, err
	}

	return sexp, nil
}

func DeserializeJmpInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeJmpIfInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeJmpElseInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeDefineInstr(data Instr) string {
	return string(data.Data)
}

func DeserializeDefineArgsInstr(data Instr) string {
	return string(data.Data)
}

func DeserializeCreateClosureInstr(data Instr) (int64, int64) {
	varslen := int64(binary.LittleEndian.Uint64(data.Data))
	funcOpAffectedCode := int64(binary.LittleEndian.Uint64(data.Data[8:]))
	return varslen, funcOpAffectedCode
}

func DeserializeSetInstr(data Instr) string {
	return string(data.Data)
}

func DeserializeCallInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeAndInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeOrInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializePrintInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializePrintlnInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializePlusNumInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeMinusNumInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeMultiplyNumInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeDivideNumInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeModuloNumInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeEqualNumInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeNotEqualNumInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeGreaterThanNumInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeGreaterThanOrEqualNumInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeLessThanNumInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeLessThanOrEqualNumInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeNewArrayInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeArrayGetInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeArraySetInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeArrayLengthInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeArrayPushInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeNewMapInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeMapGetInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeMapSetInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeMapLengthInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeMapKeysInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeCallCCInstr(data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}
