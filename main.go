package main

import (
	"bufio"
	"fmt"
	"os"
	eval "testrand-vm/reader"
)

func main() {
	stdin := bufio.NewReader(os.Stdin)
	read := eval.NewReader(stdin)
start:
	{
		result, err := read.Read()
		if err != nil {
			fmt.Println(err.Error())
			goto start
		}
		fmt.Println(result)
		goto start
	}
}
