package test

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
	eval "testrand-vm/reader"
)

func TestRead(t *testing.T) {
	//expected,actual
	tests := []struct {
		name   string
		input  string
		actual eval.SExpression
	}{
		{"input int", "1", eval.NewInt(1)},
		{"input string", `"hello"`, eval.NewString("hello")},
		{"input list", "(1 2 3)", eval.NewConsCell(eval.NewInt(1), eval.NewConsCell(eval.NewInt(2), eval.NewConsCell(eval.NewInt(3), eval.NewConsCell(eval.NewNil(), eval.NewNil()))))},
		{"input nested list", "(1 (2 3) 4)", eval.NewConsCell(eval.NewInt(1), eval.NewConsCell(eval.NewConsCell(eval.NewInt(2), eval.NewConsCell(eval.NewInt(3), eval.NewNil())), eval.NewConsCell(eval.NewInt(4), eval.NewConsCell(eval.NewNil(), eval.NewNil()))))},
		{"input dotted list", "(1 2 . 3)", eval.NewConsCell(eval.NewInt(1), eval.NewConsCell(eval.NewInt(2), eval.NewInt(3)))},
		{"input nested dotted list", "(1 (2 . 3) 4)", eval.NewConsCell(eval.NewInt(1), eval.NewConsCell(eval.NewConsCell(eval.NewInt(2), eval.NewInt(3)), eval.NewConsCell(eval.NewInt(4), eval.NewConsCell(eval.NewNil(), eval.NewNil()))))},
		{"input symbol", "hello", eval.NewSymbol("hello")},
		{"input quoted list", "'(1 2 3)", eval.NewConsCell(eval.NewSymbol("quote"), eval.NewConsCell(eval.NewInt(1), eval.NewConsCell(eval.NewInt(2), eval.NewConsCell(eval.NewInt(3), eval.NewConsCell(eval.NewNil(), eval.NewNil())))))},
	}

	t.Parallel()

	for _, tt := range tests {

		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			stdin := bufio.NewReader(strings.NewReader(fmt.Sprintf("%s\n", tt.input)))
			r := eval.NewReader(stdin)
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
