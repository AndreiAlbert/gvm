package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"

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
	fmt.Printf("size is: %d\n", unsafe.Sizeof(common.Value{}))
	bytecode := []byte{
		byte(vm.STRALLOC),
		0x00, 0x05,
		'h', 'e', 'l', 'l', 'o',
		byte(vm.HALT),
	}
	v := vm.NewVm(bytecode)
	v.Run()
}
