package unitTest

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
	"testrand-vm/compile"
	test_util "testrand-vm/test-util"
	"testrand-vm/vm"
)

func TestClosure(t *testing.T) {
	input := []string{
		"(define b '())",
		"(define f (lambda () (begin (define a 0)  (set b (lambda () (set a (+ a 1)))))))",
		"(f)",
		"(b)",
		"(b)",
		"(b)",
	}
	actuallyCases := []string{
		"b",
		"f",
		"closure",
		"1",
		"2",
		"3",
	}

	compileEnv := compile.NewCompileEnvironment("test", nil)
	runner := vm.NewVM(compileEnv)
	{
		//load file
		file, err := os.Open("../lib-lisp/lib.t-lisp")
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

	for i, v := range input {
		actually := test_util.CaptureStdout(func() {
			sample := strings.NewReader(v + "\n")
			r := bufio.NewReader(sample)
			sexp, err := compile.NewReader(compileEnv, r).Read()

			if err != nil {
				fmt.Println(err)
				t.Errorf("reader failed %s", err)
			}

			if compErr := compileEnv.Compile(sexp); compErr != nil {
				t.Errorf("reader failed %s", compErr)
			}

			vm.VMRunFromEntryPoint(runner)
		})
		if actually != actuallyCases[i]+"\n" {
			t.Errorf("expect %s, but actually %s", actuallyCases[i], actually)
		}
	}
}
