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
	cont := vm.NewContinuation()
	machine.Cont = cont

	for {
		sexp, err := read.Read()
		if err != nil {
			break
		}
		//stack, stacklen := compile.GenerateOpCode(sexp, machine.Pc)
		stack, _ := compile.GenerateOpCode(sexp, 0)
		machine.Cont.Code = stack

		for i := 0; i < len(stack); i++ {
			fmt.Println(i, stack[i])
		}

		vm.VMRun(machine)

		fmt.Println("pc:", machine.Cont.Pc)
	}
}
