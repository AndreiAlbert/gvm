package vm

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

type StackFrame struct {
	Locals        map[uint16]int32
	ReturnAddress uint
	LocalStack    []int32
}

type VM struct {
	Ip        uint
	Bytecode  []byte
	Running   bool
	CallStack []StackFrame
}

func NewVm(bytecode []byte) *VM {
	vm := &VM{
		Ip:       0,
		Bytecode: bytecode,
		Running:  true,
	}
	vm.PushFrame(0)
	return vm
}

func (v *VM) PushFrame(returnAddress uint) {
	frame := StackFrame{
		Locals:        make(map[uint16]int32),
		ReturnAddress: returnAddress,
	}
	v.CallStack = append(v.CallStack, frame)
}

func (v *VM) getCurrentFrame() *StackFrame {
	currentFrameIdx := len(v.CallStack) - 1
	return &v.CallStack[currentFrameIdx]
}

func (v *VM) extractInt32() int32 {
	value := binary.BigEndian.Uint32(v.Bytecode[v.Ip : v.Ip+4])
	v.Ip += 4
	return int32(value)
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

func (v *VM) push(value int32) {
	if len(v.CallStack) == 0 {
		log.Fatal("call stack empty")
	}
	currentFrameIdx := len(v.CallStack) - 1
	currentFrame := &v.CallStack[currentFrameIdx]
	currentFrame.LocalStack = append(currentFrame.LocalStack, value)
}

func (v *VM) pop() int32 {
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
		value := v.extractInt32()
		v.push(value)
	case POP:
		v.pop()
	case ADD:
		v1 := v.pop()
		v2 := v.pop()
		v.push(v1 + v2)
	case SUB:
		v1 := v.pop()
		v2 := v.pop()
		v.push(v2 - v1)
	case MUL:
		v1 := v.pop()
		v2 := v.pop()
		v.push(v1 * v2)
	case JMP:
		addr := uint(v.extractUInt16())
		v.Ip = uint(addr)
	// jump to an addr if top of stack not equal to value
	case JNE:
		addr := uint(v.extractUInt16())
		if addr >= uint(len(v.Bytecode)) {
			log.Fatalf("Invalid address: %d", addr)
		}
		value := v.extractInt32()
		topOfStack := v.pop()
		if value != topOfStack {
			v.Ip = addr
		}
	case JE:
		addr := uint(v.extractUInt16())
		if addr >= uint(len(v.Bytecode)) {
			log.Fatalf("Invalid address: %d", addr)
		}
		value := v.extractInt32()
		topOfStack := v.pop()
		if value == topOfStack {
			v.Ip = addr
		}
	case DIV:
		v1 := v.pop()
		v2 := v.pop()
		if v1 == 0 {
			panic("Division by zero")
		}
		v.push(v2 / v1)
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
		var args []int32
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
		if v1 == v2 {
			v.push(1)
		} else {
			v.push(0)
		}
	case NE:
		v1 := v.pop()
		v2 := v.pop()
		if v1 != v2 {
			v.push(1)
		} else {
			v.push(0)
		}
	case LT:
		v1 := v.pop()
		v2 := v.pop()
		if v2 < v1 {
			v.push(1)
		} else {
			v.push(0)
		}
	case LE:
		v1 := v.pop()
		v2 := v.pop()
		if v2 <= v1 {
			v.push(1)
		} else {
			v.push(0)
		}

	case GT:
		v1 := v.pop()
		v2 := v.pop()
		if v2 > v1 {
			v.push(1)
		} else {
			v.push(0)
		}
	case GE:
		v1 := v.pop()
		v2 := v.pop()
		if v2 >= v1 {
			v.push(1)
		} else {
			v.push(0)
		}
	}
}
