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

	{
		//load file
		file, err := os.Open("./lib-lisp/lib.t-lisp")
		if err != nil {
			panic(err)
		}
		defer file.Close()
		r := eval.NewReader(bufio.NewReader(file))
		libSexp, err := r.Read()
		if err != nil {
			panic(err)
		}
		libStack, _, err := compile.GenerateOpCode(libSexp, machine.Pc)
		if err != nil {
			panic(err)
		}
		machine.AddCode(libStack)
		vm.VMRun(machine)
	}

	for {
		machine.Code = nil
		sexp, err := read.Read()
		if err != nil {
			break
		}
		//stack, stacklen := compile.GenerateOpCode(sexp, machine.Pc)
		stack, _, err := compile.GenerateOpCode(sexp, machine.Pc)

		if err != nil {
			fmt.Println(err)
			continue
		}

		machine.AddCode(stack)

		for i := 0; i < len(stack); i++ {
			fmt.Println(i, stack[i])
		}

		vm.VMRun(machine)
		//fmt.Println("pc:", machine.Pc, "len:", stacklen)
	}
}
