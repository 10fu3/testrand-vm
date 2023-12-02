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
(loop (< a 50000000) (begin
(set a (+ a 1))
))
a
)
	`)
	r := bufio.NewReader(sample)
	machine := vm.NewVM()
	sexp, err := reader.NewReader(r).Read()
	cont := vm.NewContinuation()

	stack, _ := compile.GenerateOpCode(sexp, cont.Pc)
	cont.Code = stack
	machine.Cont = cont
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
