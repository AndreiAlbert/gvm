package vm

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	. "stack_vm/common"
	"stack_vm/heap"
)

type StackFrame struct {
	Locals        map[uint16]Value
	ReturnAddress uint
	LocalStack    []Value
}

type VM struct {
	Ip        uint
	Bytecode  []byte
	Running   bool
	CallStack []StackFrame
	Heap      *heap.Heap
}

func NewVm(bytecode []byte) *VM {
	vm := &VM{
		Ip:       0,
		Bytecode: bytecode,
		Running:  true,
		Heap:     heap.NewHeap(),
	}
	vm.PushFrame(0)
	return vm
}

func (v *VM) PushFrame(returnAddress uint) {
	frame := StackFrame{
		Locals:        make(map[uint16]Value),
		ReturnAddress: returnAddress,
	}
	v.CallStack = append(v.CallStack, frame)
}

func (v *VM) getCurrentFrame() *StackFrame {
	currentFrameIdx := len(v.CallStack) - 1
	return &v.CallStack[currentFrameIdx]
}

func (v *VM) extractUInt32() uint32 {
	value := binary.BigEndian.Uint32(v.Bytecode[v.Ip : v.Ip+4])
	v.Ip += 4
	return value
}

func (v *VM) extractUInt16() uint16 {
	value := binary.BigEndian.Uint16(v.Bytecode[v.Ip : v.Ip+2])
	v.Ip += 2
	return value
}

func (v *VM) getByte() byte {
	b := v.Bytecode[v.Ip]
	v.Ip++
	return b
}

func (v *VM) Run() {
	reader := bufio.NewReader(os.Stdin)
	for v.Running {
		reader.ReadString('\n')
		opcode := v.getByte()
		v.execute(Opcode(opcode))
	}
}

func (v *VM) push(value Value) {
	if len(v.CallStack) == 0 {
		log.Fatal("call stack empty")
	}
	currentFrameIdx := len(v.CallStack) - 1
	currentFrame := &v.CallStack[currentFrameIdx]
	currentFrame.LocalStack = append(currentFrame.LocalStack, value)
}

func (v *VM) pop() Value {
	if len(v.CallStack) == 0 {
		log.Fatalf("call stack empty")
	}
	currentFrameIdx := len(v.CallStack) - 1
	currentFrame := &v.CallStack[currentFrameIdx]
	if len(currentFrame.LocalStack) == 0 {
		log.Fatal("local stack empty")
	}
	value := currentFrame.LocalStack[len(currentFrame.LocalStack)-1]
	currentFrame.LocalStack = currentFrame.LocalStack[:len(currentFrame.LocalStack)-1]
	return value
}

func (v *VM) debugState(opcode Opcode) {
	fmt.Println("========================================")
	fmt.Printf(" DEBUG STATE\n")
	fmt.Println("========================================")
	fmt.Printf("  IP:     %d\n", v.Ip)
	fmt.Printf("  Opcode: %v\n\n", opcode)
	if len(v.CallStack) == 0 {
		fmt.Println("  Call Stack: [empty]")
	} else {
		fmt.Println("  Call Stack (top first):")
		for i := len(v.CallStack) - 1; i >= 0; i-- {
			frame := v.CallStack[i]
			fmt.Printf("    Frame #%d:\n", i)
			fmt.Printf("      ReturnAddress: %d\n", frame.ReturnAddress)
			if len(frame.LocalStack) > 0 {
				fmt.Printf("      LocalStack   : %v\n", frame.LocalStack)
			} else {
				fmt.Printf("      LocalStack   : [empty]\n")
			}
			if len(frame.Locals) > 0 {
				fmt.Printf("      Locals       : %v\n", frame.Locals)
			} else {
				fmt.Printf("      Locals       : [none]\n")
			}
		}
	}
	fmt.Println()
	windowSize := 8
	start := int(v.Ip)
	end := start + windowSize
	if start < 0 {
		start = 0
	}
	if end > len(v.Bytecode) {
		end = len(v.Bytecode)
	}
	fmt.Printf("  Bytecode (showing %d bytes from IP):\n", windowSize)
	if start >= len(v.Bytecode) {
		fmt.Printf("    [IP out of range]\n")
	} else {
		snippet := v.Bytecode[start:end]
		fmt.Printf("    %v\n", snippet)
	}
	fmt.Println("========================================")
}

func (v *VM) execute(opcode Opcode) {
	v.debugState(opcode)
	switch opcode {
	case HALT:
		v.Running = false
	case PUSH:
		typeTag := v.getByte()
		bits := v.extractUInt32()
		var val Value
		switch ValueKind(typeTag) {
		case ValueInt32:
			val = Value{Kind: ValueInt32, Raw: bits}
		case ValueFloat32:
			val = Value{Kind: ValueFloat32, Raw: bits}
		}
		v.push(val)
	case POP:
		v.pop()
	case IADD:
		v1 := v.pop()
		v2 := v.pop()
		if v1.Kind != ValueInt32 || v2.Kind != ValueInt32 {
			log.Fatalf("Values need to be int32")
		}
		result := v1.AsInt32() + v2.AsInt32()
		value := Int32Value(result)
		v.push(value)
	case ISUB:
		v1 := v.pop()
		v2 := v.pop()
		if v1.Kind != ValueInt32 || v2.Kind != ValueInt32 {
			log.Fatalf("Values need to be int32")
		}
		result := v1.AsInt32() - v2.AsInt32()
		value := Int32Value(result)
		v.push(value)
	case IMUL:
		v1 := v.pop()
		v2 := v.pop()
		if v1.Kind != ValueInt32 || v2.Kind != ValueInt32 {
			log.Fatalf("Values need to be int32")
		}
		result := v1.AsInt32() * v2.AsInt32()
		value := Int32Value(result)
		v.push(value)
	case IDIV:
		v1 := v.pop()
		v2 := v.pop()
		if v1.Kind != ValueInt32 || v2.Kind != ValueInt32 {
			log.Fatalf("Values need to be int32")
		}
		if v1.AsInt32() == 0 {
			log.Fatal("Division by zero")
		}
		result := v2.AsInt32() / v1.AsInt32()
		value := Int32Value(result)
		v.push(value)
	case FADD:
		v1 := v.pop()
		v2 := v.pop()
		if v1.Kind != ValueFloat32 || v2.Kind != ValueFloat32 {
			log.Fatal("Values need to be float32")
		}
		result := v1.AsFloat32() + v2.AsFloat32()
		value := Float32Value(result)
		v.push(value)
	case FSUB:
		v1 := v.pop()
		v2 := v.pop()
		if v1.Kind != ValueFloat32 || v2.Kind != ValueFloat32 {
			log.Fatal("Values need to be float32")
		}
		result := v2.AsFloat32() - v1.AsFloat32()
		value := Float32Value(result)
		v.push(value)
	case FMUL:
		v1 := v.pop()
		v2 := v.pop()
		if v1.Kind != ValueFloat32 || v2.Kind != ValueFloat32 {
			log.Fatal("Values need to be float32")
		}
		result := v1.AsFloat32() * v2.AsFloat32()
		value := Float32Value(result)
		v.push(value)
	case FDIV:
		v1 := v.pop()
		v2 := v.pop()
		if v1.Kind != ValueFloat32 || v2.Kind != ValueFloat32 {
			log.Fatal("Values need to be float32")
		}
		if v1.AsFloat32() == 0 {
			log.Fatal("Division by zero")
		}
		result := v2.AsFloat32() / v1.AsFloat32()
		value := Float32Value(result)
		v.push(value)
	case JMP:
		addr := uint(v.extractUInt16())
		v.Ip = uint(addr)
	// jump to an addr if top of stack not equal to value
	case IJNE:
		addr := uint(v.extractUInt16())
		if addr >= uint(len(v.Bytecode)) {
			log.Fatalf("Invalid address: %d", addr)
		}
		value := int32(v.extractUInt32())
		topOfStack := v.pop()
		if topOfStack.Kind != ValueInt32 {
			log.Fatal("should be an int32")
		}
		if value != topOfStack.AsInt32() {
			v.Ip = addr
		}
	case IJE:
		addr := uint(v.extractUInt16())
		if addr >= uint(len(v.Bytecode)) {
			log.Fatalf("Invalid address: %d", addr)
		}
		value := int32(v.extractUInt32())
		topOfStack := v.pop()
		if topOfStack.Kind != ValueInt32 {
			log.Fatal("Should be an int32")
		}
		if value == topOfStack.AsInt32() {
			v.Ip = addr
		}
	case FJNE:
		addr := uint(v.extractUInt16())
		if addr >= uint(len(v.Bytecode)) {
			log.Fatalf("Invalid address: %d\n", addr)
		}
		value := math.Float32frombits(v.extractUInt32())
		topOfStack := v.pop()
		if topOfStack.Kind != ValueFloat32 {
			log.Fatal("Should be a float32")
		}
		if value != topOfStack.AsFloat32() {
			v.Ip = addr
		}
	case FJE:
		addr := uint(v.extractUInt16())
		if addr >= uint(len(v.Bytecode)) {
			log.Fatalf("Invalid address: %d\n", addr)
		}
		value := math.Float32frombits(v.extractUInt32())
		topOfStack := v.pop()
		if topOfStack.Kind != ValueFloat32 {
			log.Fatal("Should be a float32")
		}
		if value == topOfStack.AsFloat32() {
			v.Ip = addr
		}
	// store top of the stack to an address
	case STORE:
		addr := v.extractUInt16()
		topOfStack := v.pop()
		currentFrame := v.getCurrentFrame()
		currentFrame.Locals[addr] = topOfStack
	// load value from addr on top of the stack
	case LOAD:
		addr := v.extractUInt16()
		currentFrame := v.getCurrentFrame()
		value, ok := currentFrame.Locals[addr]
		if !ok {
			log.Fatalf("Local variable at address %d not found", addr)
		}
		v.push(value)
	//call to an address
	case CALL:
		calleeAddr := uint(v.extractUInt16())
		nrOfArgs := uint(v.getByte())
		var args []Value
		for i := 0; i < int(nrOfArgs); i++ {
			arg := v.pop()
			args = append(args, arg)
		}
		returnAddress := v.Ip
		v.PushFrame(returnAddress)
		for i := len(args) - 1; i >= 0; i-- {
			v.push(args[i])
		}
		v.Ip = calleeAddr
	case RET:
		if len(v.CallStack) == 0 {
			log.Fatal("Cannot RET: callstack empty")
		}
		calleeFrame := v.getCurrentFrame()
		if len(calleeFrame.LocalStack) == 0 {
			log.Fatal("Cannot RET: local stack empty")
		}
		returnValue := calleeFrame.LocalStack[len(calleeFrame.LocalStack)-1]
		v.CallStack = v.CallStack[:len(v.CallStack)-1]
		if len(v.CallStack) == 0 {
			log.Fatal("Cannot RET: callstack empty")
		}
		callerFrame := v.getCurrentFrame()
		callerFrame.LocalStack = append(callerFrame.LocalStack, returnValue)
		v.Ip = calleeFrame.ReturnAddress
	// return to callee frame without a return value (return void)
	case RETV:
		if len(v.CallStack) == 0 {
			log.Fatal("CANNOT RETV: callstack empty")
		}
		calleeFrame := v.getCurrentFrame()
		v.CallStack = v.CallStack[:len(v.CallStack)-1]
		v.Ip = calleeFrame.ReturnAddress
	case EQ:
		v1 := v.pop()
		v2 := v.pop()
		if Equals(v1, v2) {
			v.push(Int32Value(1))
		} else {
			v.push(Int32Value(0))
		}
	case NE:
		v1 := v.pop()
		v2 := v.pop()
		if !Equals(v1, v2) {
			v.push(Int32Value(1))
		} else {
			v.push(Int32Value(0))
		}
	case LT:
		v1 := v.pop()
		v2 := v.pop()
		if v2.Lesser(v1) {
			v.push(Int32Value(1))
		} else {
			v.push(Int32Value(0))
		}
	case LE:
		v1 := v.pop()
		v2 := v.pop()
		if v2.LesserOrEqual(v1) {
			v.push(Int32Value(1))
		} else {
			v.push(Int32Value(0))
		}
	case GT:
		v1 := v.pop()
		v2 := v.pop()
		if !v2.Lesser(v1) {
			v.push(Int32Value(1))
		} else {
			v.push(Int32Value(0))
		}
	case GE:
		v1 := v.pop()
		v2 := v.pop()
		if v1.Lesser(v2) || Equals(v1, v2) {
			v.push(Int32Value(1))
		} else {
			v.push(Int32Value(0))
		}
	case ALLOC:
		topOfStack := v.pop()
		if topOfStack.Kind != ValueInt32 {
			log.Fatalf("size should be an integer")
		}
		bytes := uintptr(topOfStack.AsInt32())
		ptr, err := v.Heap.Allocate(bytes)
		fmt.Printf("allocating %d\n", ptr)
		if err != nil {
			log.Fatal(err)
		}
		v.push(PtrValue(ptr))
	case FREE:
		ptr := v.pop().AsPtr()
		fmt.Printf("freeing %d\n", ptr)
		err := v.Heap.Free(ptr)
		if err != nil {
			log.Fatal(err)
		}
	case LOADH:
		ptr := v.pop().AsPtr()
		fmt.Printf("loading %d\n", ptr)
		value, err := v.Heap.LoadValue(ptr)
		if err != nil {
			log.Fatal(err)
		}
		v.push(*value)
	case STOREH:
		value := v.pop()
		ptr := v.pop().AsPtr()
		err := v.Heap.StoreValue(ptr, value)
		if err != nil {
			log.Fatal(err)
		}
	case DUP:
		topOfStack := v.pop()
		v.push(topOfStack)
		v.push(topOfStack)
	}
}
