package main

import (
	"testrand-vm/compile"
	"testrand-vm/config"
	"testrand-vm/vm"
)

func main() {
	compilerEnv := compile.NewCompileEnvironment()
	conf := config.Get()
	vm.StartServer(compilerEnv, conf)
}
