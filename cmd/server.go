package main

import (
	"testrand-vm/config"
	"testrand-vm/vm"
)

func main() {
	conf := config.Get()
	vm.StartServer(conf)
}
