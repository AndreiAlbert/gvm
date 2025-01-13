package main

import (
	"encoding/binary"
	"math"
	"stack_vm/vm"
)

func to32bits(f float32) [4]byte {
	var buffer [4]byte
	binary.BigEndian.PutUint32(buffer[:], math.Float32bits(f))
	return buffer
}

func main() {
	// floatBites := to32bits(5.0)
	// floatBites2 := to32bits(2.0)
	bytecode := []byte{
		byte(vm.PUSH),
		byte(vm.ValueInt32),
		0x00, 0x00, 0x00, 0x01,
		byte(vm.PUSH),
		byte(vm.ValueInt32),
		0x00, 0x00, 0x00, 0x01,
		byte(vm.LE),
		byte(vm.HALT),
	}
	v := vm.NewVm(bytecode)
	v.Run()
}
