package vm

import (
	"log"
	"os"
	"stack_vm/common"
)

type Systemcall byte

const (
	STR_LEN Systemcall = iota
	STR_CAT
	STR_EQUALS
	WRITE_BYTE
	READ_BYTE
)

func (v *VM) executeSystemCall(call Systemcall) {
	switch call {
	case STR_LEN:
		strPtr := v.pop().AsPtr()
		str, err := v.Heap.LoadString(strPtr)
		if err != nil {
			log.Fatal(err)
		}
		v.push(common.Int32Value(int32(len(str))))
	case STR_CAT:
		str1Ptr := v.pop().AsPtr()
		str2Ptr := v.pop().AsPtr()
		str1, err1 := v.Heap.LoadString(str1Ptr)
		str2, err2 := v.Heap.LoadString(str2Ptr)
		if err1 != nil {
			log.Fatal(err1)
		}
		if err2 != nil {
			log.Fatal(err2)
		}
		ptr, err := v.Heap.AllocateString(str1 + str2)
		if err != nil {
			log.Fatal(err)
		}
		v.push(common.PtrValue(ptr))
	case STR_EQUALS:
		str1Ptr := v.pop().AsPtr()
		str2Ptr := v.pop().AsPtr()
		str1, err1 := v.Heap.LoadString(str1Ptr)
		str2, err2 := v.Heap.LoadString(str2Ptr)
		if err1 != nil {
			log.Fatal(err1)
		}
		if err2 != nil {
			log.Fatal(err2)
		}
		if str1 == str2 {
			v.push(common.Int32Value(1))
		} else {
			v.push(common.Int32Value(0))
		}
	case WRITE_BYTE:
		value := v.pop()
		var byteValue byte
		if value.Kind == common.ValueByte {
			byteValue = value.AsByte()
		} else if value.Kind == common.ValueInt32 {
			byteValue = byte(value.AsInt32() & 0xFF)
		} else {
			log.Fatal("WRITE_BYTE expects a byte or int32 value")
		}
		_, err := os.Stdout.Write([]byte{byteValue})
		if err != nil {
			log.Fatal(err)
		}
	case READ_BYTE:
		var buffer [1]byte
		_, err := os.Stdin.Read(buffer[:])
		if err != nil {
			log.Fatal(err)
		}
		v.push(common.ByteValue(buffer[0]))
	}
}
