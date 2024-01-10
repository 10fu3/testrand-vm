package main

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	"os"
	"testrand-vm/compile"
	"testrand-vm/vm"
)

func main() {
	stdin := bufio.NewReader(os.Stdin)
	compileEnv := compile.NewCompileEnvironment(uuid.New().String(), nil)
	read := compile.NewReader(compileEnv, stdin)
	runner := vm.NewVM(compileEnv)
	{
		//load file
		file, err := os.Open("./lib-lisp/lib.t-lisp")
		if err != nil {
			panic(err)
		}
		defer file.Close()
		r := compile.NewReader(compileEnv, bufio.NewReader(file))
		libSexp, err := r.Read()
		if err != nil {
			panic(err)
		}
		if libCompileErr := compileEnv.Compile(libSexp); libCompileErr != nil {
			fmt.Println(libCompileErr)
			os.Exit(1)
		}
		vm.VMRunFromEntryPoint(runner)
	}

	for {
		sexp, err := read.Read()
		if err != nil {
			break
		}

		runtimeErr := compileEnv.Compile(sexp)

		if runtimeErr != nil {
			fmt.Println("Runtime Error: ", runtimeErr)
			continue
		}
		vm.VMRunFromEntryPoint(runner)
		if runner.ResultErr != nil {
			fmt.Println("Runtime Error: ", runner.ResultErr)
		}
	}
}
