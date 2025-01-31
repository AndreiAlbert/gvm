package main

import (
	"encoding/binary"
	"math"

	"stack_vm/common"
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
		byte(vm.FUNC),
		byte(vm.FUNC_MAIN),
		0x00, 0x00,
		byte(common.VoidValue),
		byte(vm.CALL),
		0x00, 0x0E,
		byte(vm.HALT),

		byte(vm.FUNC),
		byte(vm.FUNC_NORMAL),
		0x00, 0x00,
		byte(common.ValueInt32),
		byte(vm.PUSH),
		byte(common.ValueInt32),
		0x00, 0x00, 0x00, 0x45,
		byte(vm.RET),
	}
	v := vm.NewVm(bytecode)
	v.Run()
}
