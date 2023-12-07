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

func TestLambdaUnit(t *testing.T) {

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
		"(define check (lambda (age)" +
			"(cond" +
			"((or (<= age 3) (>= age 65)) 0)" +
			"((and (<= 4 age) (<= age 6)) 50)" +
			"((and (<= 7 age) (<= age 12)) 100)" +
			"((and (<= 13 age) (<= age 15)) 150)" +
			"((and (<= 16 age) (<= age 18)) 180)" +
			"(#t 200))))",

		"(check 0)",   // 0
		"(check 3)",   // 0
		"(check 4)",   // 50
		"(check 6)",   // 50
		"(check 7)",   // 100
		"(check 12)",  // 100
		"(check 13)",  // 150
		"(check 15)",  // 150
		"(check 16)",  // 180
		"(check 18)",  // 180
		"(check 19)",  // 200
		"(check 64)",  // 200
		"(check 65)",  // 0
		"(check 120)", // 0
	}

	actuallyCases := []string{
		"check",
		"0",
		"0",
		"50",
		"50",
		"100",
		"100",
		"150",
		"150",
		"180",
		"180",
		"200",
		"200",
		"0",
		"0",
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
