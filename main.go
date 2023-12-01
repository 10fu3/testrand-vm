package main

import (
	"bufio"
	"fmt"
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
		stack, stacklen := compile.GenerateOpCode(sexp, machine.Pc)

		machine.AddCode(stack)

		vm.VMRun(machine)

		fmt.Println("pc:", machine.Pc, "len:", stacklen)
	}
}
