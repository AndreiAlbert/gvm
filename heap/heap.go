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
	size := unsafe.Sizeof(value)
	mem, exists := heap.Memory[ptr]
	if !exists {
		return errors.New("Invalid memory address")
	}
	if uintptr(len(mem)) < size {
		return errors.New("memory access out of bounds")
	}
	*(*Value)(unsafe.Pointer(ptr)) = value
	fmt.Printf("Allocated value: %v at address: %d\n", value, ptr)
	return nil
}

func (heap *Heap) LoadValue(ptr uintptr) (*Value, error) {
	size := unsafe.Sizeof(Value{})
	mem, exists := heap.Memory[ptr]
	if !exists {
		return nil, errors.New("Invalid memory address")
	}
	if uintptr(len(mem)) < size {
		return nil, errors.New("memory access out of bounds")
	}
	return (*Value)(unsafe.Pointer(ptr)), nil
}
