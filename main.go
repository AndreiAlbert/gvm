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
	floatBites := to32bits(5.0)
	// floatBites2 := to32bits(2.0)
	bytecode := []byte{
		byte(vm.DEFSTRUCT),
		'p', 'o', 'i', 'n', 't', 0,
		0x02,

		'x', 0,
		byte(common.ValueFloat32),

		'y', 0,
		byte(common.ValueFloat32),

		byte(vm.FUNC),
		byte(vm.FUNC_MAIN),
		0x00, 0x00,
		byte(common.ValueVoid),

		byte(vm.NEWSTRUCT),
		'p', 'o', 'i', 'n', 't', 0,
		byte(vm.DUP),

		byte(vm.PUSH),
		byte(common.ValueFloat32),
		floatBites[0], floatBites[1], floatBites[2], floatBites[3],

		byte(vm.STFIELD),
		'x', 0,

		byte(vm.FLDGET),
		'x', 0,

		byte(vm.HALT),
	}
	v := vm.NewVm(bytecode)
	v.Run()
}
