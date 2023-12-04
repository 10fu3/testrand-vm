package main

import (
	"bufio"
	"os"
	"strings"
	"testing"
	"testrand-vm/compile"
	"testrand-vm/reader"
	"testrand-vm/vm"
)

func BenchmarkRead(b *testing.B) {
	sample := strings.NewReader(`
(begin
(define a 0)
(loop (< a 50000000) (begin
(set a (+ a 1))
))
a
)
	`)
	r := bufio.NewReader(sample)
	machine := vm.NewVM()
	sexp, err := reader.NewReader(r).Read()
	stack, _, err := compile.GenerateOpCode(sexp, machine.Pc)
	machine.AddCode(stack)
	if err != nil {
		panic(err)
	}
	b.StartTimer()
	vm.VMRun(machine)
	b.StopTimer()
	if err != nil {
		panic(err)
	}
}

func BenchmarkIO(b *testing.B) {
	machine := vm.NewVM()
	{
		//load file
		file, err := os.Open("./lib-lisp/lib.t-lisp")
		if err != nil {
			panic(err)
		}
		defer file.Close()
		r := reader.NewReader(bufio.NewReader(file))
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
	sexp, err := reader.NewReader(r).Read()
	stack, _, err := compile.GenerateOpCode(sexp, machine.Pc)
	machine.AddCode(stack)
	if err != nil {
		panic(err)
	}
	b.StartTimer()
	vm.VMRun(machine)
	b.StopTimer()
	if err != nil {
		panic(err)
	}
}
