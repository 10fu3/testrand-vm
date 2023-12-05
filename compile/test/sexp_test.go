package test

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
	"testrand-vm/compile"
)

func TestRead(t *testing.T) {
	//expected,actual
	tests := []struct {
		name   string
		input  string
		actual compile.SExpression
	}{
		{"input int", "1", compile.NewInt(1)},
		{"input string", `"hello"`, compile.NewString("hello")},
		{"input list", "(1 2 3)", compile.NewConsCell(compile.NewInt(1), compile.NewConsCell(compile.NewInt(2), compile.NewConsCell(compile.NewInt(3), compile.NewConsCell(compile.NewNil(), compile.NewNil()))))},
		{"input nested list", "(1 (2 3) 4)", compile.NewConsCell(compile.NewInt(1), compile.NewConsCell(compile.NewConsCell(compile.NewInt(2), compile.NewConsCell(compile.NewInt(3), compile.NewNil())), compile.NewConsCell(compile.NewInt(4), compile.NewConsCell(compile.NewNil(), compile.NewNil()))))},
		{"input dotted list", "(1 2 . 3)", compile.NewConsCell(compile.NewInt(1), compile.NewConsCell(compile.NewInt(2), compile.NewInt(3)))},
		{"input nested dotted list", "(1 (2 . 3) 4)", compile.NewConsCell(compile.NewInt(1), compile.NewConsCell(compile.NewConsCell(compile.NewInt(2), compile.NewInt(3)), compile.NewConsCell(compile.NewInt(4), compile.NewConsCell(compile.NewNil(), compile.NewNil()))))},
		{"input symbol", "hello", compile.NewSymbol("hello")},
		{"input quoted list", "'(1 2 3)", compile.NewConsCell(compile.NewSymbol("quote"), compile.NewConsCell(compile.NewInt(1), compile.NewConsCell(compile.NewInt(2), compile.NewConsCell(compile.NewInt(3), compile.NewConsCell(compile.NewNil(), compile.NewNil())))))},
	}

	t.Parallel()

	for _, tt := range tests {

		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			stdin := bufio.NewReader(strings.NewReader(fmt.Sprintf("%s\n", tt.input)))
			r := compile.NewReader(stdin)
			actual, err := r.Read()
			if err != nil {
				t.Errorf("Read() error = %v", err)
				return
			}
			if !actual.Equals(tt.actual) {
				t.Errorf("Read() actual = %v, expected %v", actual, tt.actual)
			}
		})
	}
}
