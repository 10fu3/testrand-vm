package main

import (
	"bufio"
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
(loop (< a 10000) (begin
(set a (+ a 1))
))
)
	`)
	r := bufio.NewReader(sample)
	machine := vm.NewVM()
	sexp, err := reader.NewReader(r).Read()
	stack, _ := compile.GenerateOpCode(sexp, machine.Pc)
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
