package main

import (
	"stack_vm/vm"
)

func main() {
	bytecode := []byte{
		byte(vm.PUSH),
		0x00, 0x00, 0x00, 0x02,
		byte(vm.PUSH),
		0x00, 0x00, 0x00, 0x01,
		byte(vm.GT),
		byte(vm.HALT),
	}
	v := vm.NewVm(bytecode)
	v.Run()
}
