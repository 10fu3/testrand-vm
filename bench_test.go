package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"testrand-vm/compile"
	"testrand-vm/vm"
)

//func BenchmarkRead(b *testing.B) {
//	sample := strings.NewReader(`
//(begin
//(define a 0)
//(loop (< a 50000000) (begin
//(set a (+ a 1))
//))
//a
//)
//	`)
//	r := bufio.NewReader(sample)
//	machine := vm.NewVM()
//	sexp, err := reader.NewReader(r).Read()
//	stack, _, err := compile.GenerateOpCode(sexp, machine.Pc)
//	machine.AddCode(stack)
//	if err != nil {
//		panic(err)
//	}
//	b.StartTimer()
//	vm._vMRun(machine)
//	b.StopTimer()
//	if err != nil {
//		panic(err)
//	}
//}

func BenchmarkIO(b *testing.B) {

	compileEnv := compile.NewCompileEnvironment()
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

	sample := strings.NewReader(`
(begin
  (define before (get-time-nano))
  (define local-word-count (hashmap))
  (define i 0)
  (loop (< i 1000) (begin
    (foreach-array (string-split (read-file "sample1.txt") " ")
      (lambda (word)
        (hashmap-set local-word-count word (+ (hashmap-get local-word-count word 0) 1))
      )
    )
    (set i (+ i 1))
  ))
  (println (/ (- (get-time-nano) before) 1000))
)
	`)
	r := bufio.NewReader(sample)
	sexp, err := compile.NewReader(compileEnv, r).Read()

	if err != nil {
		fmt.Println(err)
	}

	if err := compileEnv.Compile(sexp); err != nil {
		panic(err)
	}
	runtime.GC()
	b.StartTimer()
	vm.VMRunFromEntryPoint(runner)
	b.StopTimer()
	if err != nil {
		panic(err)
	}
}
