package main

import (
	"bufio"
	"fmt"
	"os"
	"testrand-vm/compile"
	eval "testrand-vm/reader"
)

func main() {
	stdin := bufio.NewReader(os.Stdin)
	read := eval.NewReader(stdin)

	for {
		sexp, err := read.Read()
		if err != nil {
			break
		}
		stack, lows := compile.GenerateOpCode(sexp, 0)

		for i := 0; i < len(stack); i++ {
			fmt.Println(i, lows, stack[i].String())
		}
	}
}
