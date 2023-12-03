package instr

const (
	OPCODE_PUSH_SYM                  = uint8(iota) // 0
	OPCODE_PUSH_NUM                                // 1
	OPCODE_PUSH_TRUE                               // 2
	OPCODE_PUSH_FALSE                              // 3
	OPCODE_PUSH_STR                                // 4
	OPCODE_PUSH_NIL                                // 5
	OPCODE_PUSH_SEXP                               // 6
	OPCODE_POP                                     // 7
	OPCODE_JMP                                     // 8
	OPCODE_JMP_IF                                  // 9
	OPCODE_JMP_ELSE                                // 10
	OPCODE_LOAD                                    // 11
	OPCODE_DEFINE                                  // 12
	OPCODE_DEFINE_ARGS                             // 13
	OPCODE_SET                                     // 14
	OPCODE_NEW_ENV                                 // 15
	OPCODE_CREATE_CLOSURE                          // 16
	OPCODE_CALL                                    // 17
	OPCODE_RETURN                                  // 18
	OPCODE_AND                                     // 19
	OPCODE_OR                                      // 20
	OPCODE_PRINT                                   // 21
	OPCODE_PRINTLN                                 // 22
	OPCODE_PLUS_NUM                                // 23
	OPCODE_MINUS_NUM                               // 24
	OPCODE_MULTIPLY_NUM                            // 25
	OPCODE_DIVIDE_NUM                              // 26
	OPCODE_MODULO_NUM                              // 27
	OPCODE_EQUAL_NUM                               // 28
	OPCODE_NOT_EQUAL_NUM                           // 29
	OPCODE_GREATER_THAN_NUM                        // 30
	OPCODE_GREATER_THAN_OR_EQUAL_NUM               // 31
	OPCODE_LESS_THAN_NUM                           // 32
	OPCODE_LESS_THAN_OR_EQUAL_NUM                  // 33
	OPCODE_CAR                                     // 34
	OPCODE_CDR                                     // 35
	OPCODE_RANDOM_ID                               // 36
	OPCODE_NEW_ARRAY                               // 37
	OPCODE_ARRAY_GET                               // 38
	OPCODE_ARRAY_SET                               // 39
	OPCODE_ARRAY_LENGTH                            // 40
	OPCODE_ARRAY_PUSH                              // 41
	OPCODE_NEW_MAP                                 // 42
	OPCODE_MAP_GET                                 // 43
	OPCODE_MAP_SET                                 // 44
	OPCODE_MAP_LENGTH                              // 45
	OPCODE_MAP_KEYS                                // 46
	OPCODE_MAP_DELETE                              // 47
	OPCODE_END_CODE                                // 48
	OPCODE_NOP                                     // 49
	OPCODE_CALL_CC                                 // 50
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
		"call/cc":      OPCODE_CALL_CC,
	}
}
