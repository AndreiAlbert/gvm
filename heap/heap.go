package heap

import (
	"errors"
	"fmt"
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
