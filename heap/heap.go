package heap

import (
	"errors"
	"fmt"
	"log"
	. "stack_vm/common"
	"syscall"
	"unsafe"
)

type Heap struct {
	Memory map[uintptr][]byte
}

func NewHeap() *Heap {
	return &Heap{
		Memory: make(map[uintptr][]byte),
	}
}

func (heap *Heap) Allocate(size uintptr) (uintptr, error) {
	pageSize := syscall.Getpagesize()
	pagesRequired := (int(size) + pageSize - 1) / pageSize

	mem, err := syscall.Mmap(
		-1, 0,
		pageSize*pagesRequired,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_ANON|syscall.MAP_PRIVATE,
	)
	if err != nil {
		return 0, fmt.Errorf("mmap failed: %w\n", err)
	}
	ptr := uintptr(unsafe.Pointer(&mem[0]))
	heap.Memory[ptr] = mem
	return ptr, nil
}

func (heap *Heap) Free(ptr uintptr) error {
	mem, exists := heap.Memory[ptr]
	if !exists {
		return fmt.Errorf(`Failed to free memory at address: %d`, ptr)
	}
	if err := syscall.Munmap(mem); err != nil {
		return fmt.Errorf("freeing memory failed: %w", err)
	}
	delete(heap.Memory, ptr)
	return nil
}

func (heap *Heap) StoreValue(ptr uintptr, value Value) error {
	mem, exists := heap.Memory[ptr]
	if !exists {
		return errors.New("invalid memory address")
	}
	mem[0] = byte(value.Kind)
	var requiredSize uintptr
	switch value.Kind {
	case ValueInt32, ValueFloat32:
		requiredSize = 5 // type tag + 4 bytes
	case ValuePtr:
		requiredSize = 1 + unsafe.Sizeof(uintptr(0))
	}
	if uintptr(len(mem)) < requiredSize {
		return errors.New("Memory access out of bounds")
	}
	switch value.Kind {
	case ValueInt32:
		*(*int32)(unsafe.Pointer(ptr + 1)) = value.AsInt32()
	case ValueFloat32:
		*(*float32)(unsafe.Pointer(ptr + 1)) = value.AsFloat32()
	case ValuePtr:
		*(*uintptr)(unsafe.Pointer(ptr + 1)) = value.AsPtr()
	}
	return nil
}

func (heap *Heap) LoadValue(ptr uintptr) (*Value, error) {
	mem, exists := heap.Memory[ptr]
	if !exists {
		return nil, errors.New("Invalid memory address")
	}

	if len(mem) < 1 {
		return nil, errors.New("memory access out of bounds")
	}

	kind := ValueKind(mem[0])
	var value Value
	value.Kind = kind
	switch kind {
	case ValueInt32:
		if len(mem) < 5 { // tag + int32
			return nil, errors.New("memory access out of bounds")
		}
		value.Raw = uint32(*(*int32)(unsafe.Pointer(ptr + 1)))
	case ValueFloat32:
		if len(mem) < 5 { // tag + float32
			return nil, errors.New("memory access out of bounds")
		}
		value.Raw = *(*uint32)(unsafe.Pointer(ptr + 1))
	case ValuePtr:
		ptrSize := unsafe.Sizeof(uintptr(0))
		if uintptr(len(mem)) < 1+ptrSize {
			return nil, errors.New("memory access out of bounds")
		}
		value.Ptr = *(*uintptr)(unsafe.Pointer(ptr + 1))
	}
	return &value, nil
}

func (heap *Heap) AllocateString(s string) (uintptr, error) {
	// type tag + length + actual string
	totalSize := uintptr(5 + len(s))
	ptr, err := heap.Allocate(totalSize)
	if err != nil {
		return 0, err
	}
	heap.Memory[ptr][0] = byte(ValueString)
	*(*int32)(unsafe.Pointer(ptr + 1)) = int32(len(s))
	copy(heap.Memory[ptr][5:], s)
	return ptr, nil
}

func (heap *Heap) LoadString(ptr uintptr) (string, error) {
	mem, exists := heap.Memory[ptr]
	if !exists {
		return "", errors.New("Invalid memory address")
	}
	if ValueKind(mem[0]) != ValueString {
		return "", errors.New("Not a string value")
	}
	length := *(*int32)(unsafe.Pointer(ptr + 1))
	if len(mem) < 5+int(length) {
		return "", errors.New("memory access out of bounds")
	}
	return string(mem[5 : 5+length]), nil
}

func GetElementSize(kind ValueKind) uintptr {
	switch kind {
	case ValueFloat32, ValueInt32:
		return 4
	case ValuePtr, ValueString, ValueArray, ValueStruct:
		return unsafe.Sizeof(uintptr(0))
	default:
		log.Fatalf("Unsupported array element type: %v\n", kind)
		return 0
	}
}

func (heap *Heap) AllocateArray(elementKind ValueKind, length int32) (uintptr, error) {
	elementSize := GetElementSize(elementKind)
	// type tag(1) + element type (1) + size (4) + array elements
	totalSize := uintptr(1 + 1 + 4 + (elementSize * uintptr(length)))
	ptr, err := heap.Allocate(totalSize)
	if err != nil {
		return 0, err
	}
	mem := heap.Memory[ptr]
	mem[0] = byte(ValueArray)
	mem[1] = byte(elementKind)
	*(*int32)(unsafe.Pointer(ptr + 2)) = length
	return ptr, nil
}

func (heap *Heap) SetArrayElement(arrayPtr uintptr, index int32, value Value) error {
	mem, exists := heap.Memory[arrayPtr]
	if !exists {
		return errors.New("Invalid memory access")
	}
	if ValueKind(mem[0]) != ValueArray {
		return errors.New("Not an array")
	}
	elementKind := ValueKind(mem[1])
	length := *(*int32)(unsafe.Pointer(arrayPtr + 2))
	if index < 0 || index >= length {
		return errors.New("Array index out of bounds!")
	}
	if elementKind != value.Kind {
		return fmt.Errorf("Type mismatch: expected %v, got %v\n", elementKind, value.Kind)
	}
	elementSize := GetElementSize(elementKind)
	elementPtr := arrayPtr + 6 + uintptr(index)*elementSize
	switch elementKind {
	case ValueInt32, ValueFloat32:
		*(*uint32)(unsafe.Pointer(elementPtr)) = value.Raw
	case ValuePtr, ValueString:
		*(*uintptr)(unsafe.Pointer(elementPtr)) = value.Ptr
	default:
		return fmt.Errorf("Unsupported element type: %v\n", elementKind)
	}
	return nil
}

func (heap *Heap) GetArrayElement(arrayPtr uintptr, index int32) (*Value, error) {
	mem, exists := heap.Memory[arrayPtr]
	if !exists {
		return nil, errors.New("Invalid memory access")
	}
	if ValueKind(mem[0]) != ValueArray {
		return nil, errors.New("Not an array")
	}
	elementKind := ValueKind(mem[1])
	length := *(*int32)(unsafe.Pointer(arrayPtr + 2))
	if index < 0 || index >= length {
		return nil, errors.New("Array index out of bounds")
	}
	elementSize := GetElementSize(elementKind)
	elementPtr := arrayPtr + 6 + uintptr(index)*elementSize
	value := &Value{
		Kind: elementKind,
	}
	switch elementKind {
	case ValueInt32, ValueFloat32:
		value.Raw = *(*uint32)(unsafe.Pointer(elementPtr))
	case ValuePtr, ValueString:
		value.Ptr = *(*uintptr)(unsafe.Pointer(elementPtr))
	default:
		return nil, fmt.Errorf("Unsupported element type: %v\n", elementKind)
	}
	return value, nil
}

func (heap *Heap) AllocateStruct(str StructType) (uintptr, error) {
	// kind struct + struct itself
	totalSize := uintptr(1 + str.Size)
	ptr, err := heap.Allocate(totalSize)
	if err != nil {
		return 0, err
	}
	mem, ok := heap.Memory[ptr]
	if !ok {
		return 0, errors.New("Could not allocate memory properly")
	}
	mem[0] = byte(ValueStruct)
	*(*StructType)(unsafe.Pointer(ptr + 1)) = str
	return ptr, nil
}

func (heap *Heap) GetStructField(structPtr uintptr, fieldName string) (*Value, error) {
	mem, exists := heap.Memory[structPtr]
	if !exists {
		return nil, errors.New("Invalid memory address")
	}
	if ValueKind(mem[0]) != ValueStruct {
		return nil, errors.New("Not a struct")
	}
	structType := *(*StructType)(unsafe.Pointer(structPtr + 1))
	for _, field := range structType.Fields {
		if field.Name == fieldName {
			fieldPtr := structPtr + 1 + uintptr(unsafe.Sizeof(StructType{})) + uintptr(field.Offset)
			switch field.Type {
			case ValueFloat32, ValueInt32:
				value := &Value{
					Kind: field.Type,
					Raw:  *(*uint32)(unsafe.Pointer(fieldPtr)),
				}
				return value, nil
			case ValuePtr, ValueString, ValueStruct, ValueArray:
				value := &Value{
					Kind: field.Type,
					Ptr:  *(*uintptr)(unsafe.Pointer(fieldPtr)),
				}
				return value, nil
			default:
				return nil, fmt.Errorf("Unsupported field type: %v\n", field.Type)
			}
		}
	}
	return nil, fmt.Errorf("Field %s is not found on struct\n", fieldName)
}

func (heap *Heap) SetStructureField(structPtr uintptr, fieldName string, value Value) error {
	mem, exists := heap.Memory[structPtr]
	if !exists {
		return errors.New("Invalid memory address")
	}
	if ValueKind(mem[0]) != ValueStruct {
		return errors.New("Not a struct")
	}
	structType := *(*StructType)(unsafe.Pointer(structPtr + 1))
	for _, field := range structType.Fields {
		if fieldName == field.Name {
			if field.Type != value.Kind {
				return fmt.Errorf("Type mismatch: expected %v, got %v", field.Type, value.Kind)
			}
			fieldPtr := structPtr + 1 + uintptr(unsafe.Sizeof(StructType{})) + uintptr(field.Offset)
			switch field.Type {
			case ValueFloat32, ValueInt32:
				*(*uint32)(unsafe.Pointer(fieldPtr)) = value.Raw
			case ValuePtr, ValueString, ValueStruct, ValueArray:
				*(*uintptr)(unsafe.Pointer(fieldPtr)) = value.Ptr
			default:
				return fmt.Errorf("Unsupported type: %v\n", field.Type)
			}
			return nil
		}
	}
	return fmt.Errorf("Field %s not found in struct\n", fieldName)
}

func (heap *Heap) Debug() {
	log.Printf("[HEAP DEBUG] Current memory map:")
	if len(heap.Memory) == 0 {
		log.Println("No memory allocated on the heap")
		return
	}
	for ptr, mem := range heap.Memory {
		log.Printf("Address: %d, Size: %d bytes", ptr, len(mem))
		if len(mem) <= 0 {
			continue
		}
		kind := ValueKind(mem[0])
		switch kind {
		case ValueInt32:
			if len(mem) >= 5 {
				value := *(*int32)(unsafe.Pointer(ptr + 1))
				log.Printf("Decoded int32: %d\n", value)
			}
		case ValueFloat32:
			if len(mem) >= 5 {
				value := *(*float32)(unsafe.Pointer(ptr + 1))
				log.Printf("Decoded float32: %f\n", value)
			}
		case ValueString:
			if len(mem) >= 5 {
				length := *(*int32)(unsafe.Pointer(ptr + 1))
				if len(mem) >= 5+int(length) {
					log.Printf("Decoded string: %s\n", string(mem[5:5+length]))
				}
			}
		case ValuePtr:
			if len(mem) > 1 {
				ptrValue := *(*uintptr)(unsafe.Pointer(ptr + 1))
				log.Printf("Decoded pointer: %d\n", ptrValue)
			}
		case ValueArray:
			if len(mem) >= 6 {
				elementKind := ValueKind(mem[1])
				length := *(*int32)(unsafe.Pointer(ptr + 2))
				log.Printf("Decoded array: type=%v, length=%d\n", elementKind, length)
				elementSize := GetElementSize(elementKind)
				for i := int32(0); i < length; i++ {
					elementPtr := ptr + 6 + uintptr(i)*elementSize
					switch elementKind {
					case ValueInt32:
						value := *(*int32)(unsafe.Pointer(elementPtr))
						log.Printf("  [%d] = %d\n", i, value)
					case ValueFloat32:
						value := *(*float32)(unsafe.Pointer(elementPtr))
						log.Printf("  [%d] = %f\n", i, value)
					case ValuePtr, ValueString:
						value := *(*uintptr)(unsafe.Pointer(elementPtr))
						log.Printf("  [%d] = ptr(%d)\n", i, value)
					}
				}
			}
		case ValueStruct:
			structType := *(*StructType)(unsafe.Pointer(ptr + 1))
			fmt.Printf("%v\n", structType)
		default:
			log.Printf("Unkown value: %v\n", kind)
		}
	}
	log.Println()
}
