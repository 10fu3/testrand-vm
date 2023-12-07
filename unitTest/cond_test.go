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

func TestCond(t *testing.T) {
	compileEnv := compile.NewCompileEnvironment()
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

	input := []string{
		"(define age 0)",
		"(cond ((and (<= age 3) (>= age 65)) 0) (#t 200))",
		"(set age 3)",
		"(cond ((and (<= age 3) (>= age 65)) 0) (#t 200))",
		"(set age 4)",
		"(cond ((and (<= age 3) (>= age 65)) 0) (#t 200))",
		"(set age 6)",
		"(cond ((and (<= age 3) (>= age 65)) 0) (#t 200))",
		"(set age 64)",
		"(cond ((and (<= age 3) (>= age 65)) 0) (#t 200))",
		"(set age 65)",
		"(cond ((and (<= age 3) (>= age 65)) 0) (#t 200))",
	}

	actuallyCases := []string{
		"age",
		"200",
		"3",
		"200",
		"4",
		"0",
		"6",
		"0",
		"64",
		"0",
		"65",
		"200",
	}

	for i, v := range input {
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

		except := test_util.CaptureStdout(func() {
			vm.VMRunFromEntryPoint(runner)
		})
		if actuallyCases[i]+"\n" != except {
			t.Errorf("lambda test expect:%s actual: %s", except, actuallyCases[i])
		}
	}
}
