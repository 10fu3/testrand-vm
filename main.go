package main

import (
	"bufio"
	"os"
	"testrand-vm/compile"
	eval "testrand-vm/reader"
	"testrand-vm/vm"
)

func main() {
	stdin := bufio.NewReader(os.Stdin)
	read := eval.NewReader(stdin)

	nowLine := int64(0)

	for {
		sexp, err := read.Read()
		if err != nil {
			break
		}
		stack, rows := compile.GenerateOpCode(sexp, nowLine)

		nowLine += rows

		machine := vm.NewVM()

		machine.SetCode(stack)

		vm.VMRun(machine)
	}
}
