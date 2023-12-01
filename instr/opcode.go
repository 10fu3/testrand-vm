package instr

const (
	OPCODE_PUSH_SYM = uint8(iota)
	OPCODE_PUSH_NUM
	OPCODE_PUSH_TRUE
	OPCODE_PUSH_FALSE
	OPCODE_PUSH_STR
	OPCODE_PUSH_NIL
	OPCODE_PUSH_SEXP
	OPCODE_POP
	OPCODE_JMP
	OPCODE_JMP_IF
	OPCODE_JMP_ELSE
	OPCODE_LOAD
	OPCODE_DEFINE
	OPCODE_DEFINE_ARGS
	OPCODE_SET
	OPCODE_NEW_ENV
	OPCODE_CREATE_CLOSURE
	OPCODE_CALL
	OPCODE_RETURN
	OPCODE_AND
	OPCODE_OR
	OPCODE_PRINT
	OPCODE_PRINTLN
	OPCODE_PLUS_NUM
	OPCODE_MINUS_NUM
	OPCODE_MULTIPLY_NUM
	OPCODE_DIVIDE_NUM
	OPCODE_MODULO_NUM
	OPCODE_EQUAL_NUM
	OPCODE_NOT_EQUAL_NUM
	OPCODE_GREATER_THAN_NUM
	OPCODE_GREATER_THAN_OR_EQUAL_NUM
	OPCODE_LESS_THAN_NUM
	OPCODE_LESS_THAN_OR_EQUAL_NUM
	OPCODE_CAR
	OPCODE_CDR
	OPCODE_RANDOM_ID
	OPCODE_NEW_ARRAY
	OPCODE_ARRAY_GET
	OPCODE_ARRAY_SET
	OPCODE_ARRAY_LENGTH
	OPCODE_ARRAY_PUSH
	OPCODE_NEW_MAP
	OPCODE_MAP_GET
	OPCODE_MAP_SET
	OPCODE_MAP_LENGTH
	OPCODE_MAP_KEYS
	OPCODE_MAP_DELETE
	OPCODE_END_CODE
	OPCODE_NOP
)

var OpCodeMap = map[uint8]string{
	OPCODE_PUSH_SYM:                  "PUSH_SYM",
	OPCODE_PUSH_NUM:                  "PUSH_NUM",
	OPCODE_PUSH_TRUE:                 "PUSH_TRUE",
	OPCODE_PUSH_FALSE:                "PUSH_FALSE",
	OPCODE_PUSH_STR:                  "PUSH_STR",
	OPCODE_PUSH_NIL:                  "PUSH_NIL",
	OPCODE_PUSH_SEXP:                 "PUSH_SEXP",
	OPCODE_POP:                       "POP",
	OPCODE_JMP:                       "JMP",
	OPCODE_JMP_IF:                    "JMP_IF",
	OPCODE_JMP_ELSE:                  "JMP_ELSE",
	OPCODE_LOAD:                      "LOAD",
	OPCODE_DEFINE:                    "DEFINE",
	OPCODE_DEFINE_ARGS:               "DEFINE_ARGS",
	OPCODE_SET:                       "SET",
	OPCODE_NEW_ENV:                   "NEW_ENV",
	OPCODE_CREATE_CLOSURE:            "CREATE_CLOSURE",
	OPCODE_CALL:                      "CALL",
	OPCODE_RETURN:                    "RETURN",
	OPCODE_AND:                       "AND",
	OPCODE_OR:                        "OR",
	OPCODE_PRINT:                     "PRINT",
	OPCODE_PRINTLN:                   "PRINTLN",
	OPCODE_PLUS_NUM:                  "PLUS_NUM",
	OPCODE_MINUS_NUM:                 "MINUS_NUM",
	OPCODE_MULTIPLY_NUM:              "MULTIPLY_NUM",
	OPCODE_DIVIDE_NUM:                "DIVIDE_NUM",
	OPCODE_MODULO_NUM:                "MODULO_NUM",
	OPCODE_EQUAL_NUM:                 "EQUAL_NUM",
	OPCODE_NOT_EQUAL_NUM:             "NOT_EQUAL_NUM",
	OPCODE_GREATER_THAN_NUM:          "GREATER_THAN_NUM",
	OPCODE_GREATER_THAN_OR_EQUAL_NUM: "GREATER_THAN_OR_EQUAL_NUM",
	OPCODE_LESS_THAN_NUM:             "LESS_THAN_NUM",
	OPCODE_LESS_THAN_OR_EQUAL_NUM:    "LESS_THAN_OR_EQUAL_NUM",
	OPCODE_CAR:                       "CAR",
	OPCODE_CDR:                       "CDR",
	OPCODE_RANDOM_ID:                 "RANDOM_ID",
	OPCODE_NEW_ARRAY:                 "NEW_ARRAY",
	OPCODE_ARRAY_GET:                 "ARRAY_GET",
	OPCODE_ARRAY_SET:                 "ARRAY_SET",
	OPCODE_ARRAY_LENGTH:              "ARRAY_LENGTH",
	OPCODE_ARRAY_PUSH:                "ARRAY_PUSH",
	OPCODE_NEW_MAP:                   "NEW_MAP",
	OPCODE_MAP_GET:                   "MAP_GET",
	OPCODE_MAP_SET:                   "MAP_SET",
	OPCODE_MAP_LENGTH:                "MAP_LENGTH",
	OPCODE_MAP_KEYS:                  "MAP_KEYS",
	OPCODE_MAP_DELETE:                "MAP_DELETE",
	OPCODE_END_CODE:                  "END_CODE",
	OPCODE_NOP:                       "NOP",
}

func GetNativeFuncNameToOpCodeMap() map[string]uint8 {
	return map[string]uint8{
		"print":        OPCODE_PRINT,
		"println":      OPCODE_PRINTLN,
		"+":            OPCODE_PLUS_NUM,
		"-":            OPCODE_MINUS_NUM,
		"*":            OPCODE_MULTIPLY_NUM,
		"/":            OPCODE_DIVIDE_NUM,
		"%":            OPCODE_MODULO_NUM,
		"=":            OPCODE_EQUAL_NUM,
		"!=":           OPCODE_NOT_EQUAL_NUM,
		">":            OPCODE_GREATER_THAN_NUM,
		">=":           OPCODE_GREATER_THAN_OR_EQUAL_NUM,
		"<":            OPCODE_LESS_THAN_NUM,
		"<=":           OPCODE_LESS_THAN_OR_EQUAL_NUM,
		"car":          OPCODE_CAR,
		"cdr":          OPCODE_CDR,
		"random-id":    OPCODE_RANDOM_ID,
		"array":        OPCODE_NEW_ARRAY,
		"array-get":    OPCODE_ARRAY_GET,
		"array-set":    OPCODE_ARRAY_SET,
		"array-length": OPCODE_ARRAY_LENGTH,
		"array-push":   OPCODE_ARRAY_PUSH,
		"map":          OPCODE_NEW_MAP,
		"map-get":      OPCODE_MAP_GET,
		"map-set":      OPCODE_MAP_SET,
		"map-length":   OPCODE_MAP_LENGTH,
		"map-keys":     OPCODE_MAP_KEYS,
		"map-delete":   OPCODE_MAP_DELETE,
	}
}