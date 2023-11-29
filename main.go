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

	machine := vm.NewVM()

	for {
		sexp, err := read.Read()
		if err != nil {
			break
		}
		stack, _ := compile.GenerateOpCode(sexp, 0)

		machine.AddCode(stack)

		vm.VMRun(machine)
	}
}
