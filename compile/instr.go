package compile

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"strings"
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

func CreatePushStringInstr(symbolIndex uint64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, symbolIndex)
	return NewInstr(OPCODE_PUSH_STR, b)
}

func CreatePushSymbolInstr(symbolIndex uint64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, symbolIndex)
	return NewInstr(OPCODE_PUSH_SYM, b)
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

func CreatePushSExpressionInstr(symbolIndex uint64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, symbolIndex)
	return NewInstr(OPCODE_PUSH_SEXP, b)
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

func CreateLoadInstr(symbolIndex uint64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, symbolIndex)
	return NewInstr(OPCODE_LOAD, b)
}

func CreateDefineInstr(symbolIndex uint64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, symbolIndex)
	return NewInstr(OPCODE_DEFINE, b)
}

func CreateDefineArgsInstr(symbolIndex uint64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, symbolIndex)
	return NewInstr(OPCODE_DEFINE_ARGS, b)
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

func CreateSetInstr(symbolIndex uint64) Instr {
	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, symbolIndex)
	return NewInstr(OPCODE_SET, b)
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

func CreateHeavyTaskInstr(argsSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(argsSize))
	return NewInstr(OPCODE_HEAVY, b)
}

func CreateReadFileInstr(argsSize int64) Instr {
	return NewInstr(OPCODE_READ_FILE, []byte{})
}

func CreateStringSplit(instrSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(instrSize))
	return NewInstr(OPCODE_STRING_SPLIT, b)
}

func CreateStringJoin(instrSize int64) Instr {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(instrSize))
	return NewInstr(OPCODE_STRING_JOIN, b)
}

func CreateGetTimeNanos(instrSize int64) Instr {
	return NewInstr(OPCODE_GET_NOW_TIME_NANO, []byte{})
}

var NativeFuncNameToOpCodeMap = map[string]FunctionGenerateInstr{
	"print":          CreatePrintInstr,
	"println":        CreatePrintlnInstr,
	"+":              CreatePlusNumInstr,
	"-":              CreateMinusNumInstr,
	"*":              CreateMultiplyNumInstr,
	"/":              CreateDivideNumInstr,
	"%":              CreateModuloNumInstr,
	"=":              CreateEqualNumInstr,
	"!=":             CreateNotEqualNumInstr,
	">":              CreateGreaterThanNumInstr,
	">=":             CreateGreaterThanOrEqualNumInstr,
	"<":              CreateLessThanNumInstr,
	"<=":             CreateLessThanOrEqualNumInstr,
	"car":            CreateCarInstr,
	"cdr":            CreateCdrInstr,
	"random-id":      CreateRandomIdInstr,
	"array":          CreateNewArrayInstr,
	"array-get":      CreateArrayGetInstr,
	"array-set":      CreateArraySetInstr,
	"array-len":      CreateArrayLengthInstr,
	"array-push":     CreateArrayPushInstr,
	"hashmap":        CreateNewMapInstr,
	"hashmap-get":    CreateMapGetInstr,
	"hashmap-set":    CreateMapSetInstr,
	"hashmap-len":    CreateMapLengthInstr,
	"hashmap-keys":   CreateMapKeysInstr,
	"hashmap-delete": CreateMapDeleteInstr,
	"heavy":          CreateHeavyTaskInstr,
	"read-file":      CreateReadFileInstr,
	"string-split":   CreateStringSplit,
	"string-join":    CreateStringJoin,
	"get-time-nano":  CreateGetTimeNanos,
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

func DeserializeInstructions(data []byte) []Instr {
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

func DeserializeLoadInstr(compEnv *CompilerEnvironment, data Instr) uint64 {
	symbolI := binary.LittleEndian.Uint64(data.Data)
	return symbolI
}

func DeserializePushNumberInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializePushStringInstr(compEnv *CompilerEnvironment, data Instr) Str {
	index := binary.LittleEndian.Uint64(data.Data)
	return NewString(index)
}

func DeserializePushSymbolInstr(data Instr) Symbol {
	return NewSymbol(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeSexpressionInstr(compEnv *CompilerEnvironment, data Instr) (SExpression, error) {
	index := binary.LittleEndian.Uint64(data.Data)
	sample := strings.NewReader(fmt.Sprintf("%s\n", compEnv.GetCompilerSymbolString(index)))
	r := bufio.NewReader(sample)
	sexp, err := NewReader(compEnv, r).Read()

	if err != nil {
		return nil, err
	}

	return sexp, nil
}

func DeserializeJmpInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeJmpIfInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeJmpElseInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

//func DeserializeNewEnvInstr(compEnv *CompilerEnvironment, data Instr) (uint64, uint64) {
//	return parentIndex, envIndex
//}

func DeserializeDefineInstr(compEnv *CompilerEnvironment, data Instr) uint64 {
	symbolI := binary.LittleEndian.Uint64(data.Data)

	return symbolI
}

func DeserializeDefineArgsInstr(compEnv *CompilerEnvironment, data Instr) uint64 {
	symbolI := binary.LittleEndian.Uint64(data.Data)

	return symbolI
}

func DeserializeCreateClosureInstr(compEnv *CompilerEnvironment, data Instr) (int64, int64) {
	varslen := int64(binary.LittleEndian.Uint64(data.Data))
	funcOpAffectedCode := int64(binary.LittleEndian.Uint64(data.Data[8:]))
	return varslen, funcOpAffectedCode
}

func DeserializeSetInstr(compEnv *CompilerEnvironment, data Instr) uint64 {
	symbolI := binary.LittleEndian.Uint64(data.Data)

	return symbolI

}

func DeserializeCallInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeAndInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeOrInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializePrintInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializePrintlnInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializePlusNumInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeMinusNumInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeMultiplyNumInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeDivideNumInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeModuloNumInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeEqualNumInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeNotEqualNumInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeGreaterThanNumInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeGreaterThanOrEqualNumInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeLessThanNumInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeLessThanOrEqualNumInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeNewArrayInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeArrayGetInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeArraySetInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeArrayLengthInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeArrayPushInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeNewMapInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeMapGetInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeMapSetInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeMapLengthInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeMapKeysInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeHeavyInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeReadFileInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeStringSplitInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeStringJoinInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeArrayForEachInstr(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}

func DeserializeGetTimeNano(compEnv *CompilerEnvironment, data Instr) int64 {
	return int64(binary.LittleEndian.Uint64(data.Data))
}
