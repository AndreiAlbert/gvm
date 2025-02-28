package vm

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	. "stack_vm/common"
	"stack_vm/heap"
	"strings"
)

type FunctionSignature struct {
	Address          uint
	ParamCount       uint16
	ReturnType       ValueKind
	isMain           bool
	ReturnStructName string
}

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
	Functions map[uint]FunctionSignature
	Structs   map[string]StructType
}

func (f FunctionSignature) String() string {
	var str strings.Builder
	if f.isMain {
		str.WriteString("Found main function\n")
	}
	fmt.Fprintf(&str, "Number of arguments: %d\n", f.ParamCount)
	fmt.Fprintf(&str, "Address of the body: %d\n", f.Address)
	fmt.Fprintf(&str, "Return type: %v\n", f.ReturnType)
	return str.String()
}

func NewVm(bytecode []byte) *VM {
	vm := &VM{
		Ip:        0,
		Bytecode:  bytecode,
		Running:   true,
		Heap:      heap.NewHeap(),
		Functions: make(map[uint]FunctionSignature),
		Structs:   make(map[string]StructType),
	}
	vm.buildFunctionTable()
	hasStrucs := false
	if len(bytecode) > 0 {
		if bytecode[0] == byte(DEFSTRUCT) {
			hasStrucs = true
		}
	}
	if hasStrucs {
		vm.buildStructsTable()
	}
	vm.PushFrame(0xFFFFFFFF)
	return vm
}

func (v *VM) extractString() string {
	start := v.Ip
	for v.Bytecode[v.Ip] != 0 {
		v.Ip++
	}
	str := string(v.Bytecode[start:v.Ip])
	v.Ip++
	return str
}

func (v *VM) buildFunctionTable() {
	ip := uint(0)
	mainAddr := uint(0)
	foundMain := false
	for ip < uint(len(v.Bytecode)) {
		if v.Bytecode[ip] == byte(FUNC) {
			ip++
			isMain := v.Bytecode[ip] == byte(FUNC_MAIN)
			ip++

			signature := FunctionSignature{
				ParamCount: binary.BigEndian.Uint16(v.Bytecode[ip : ip+2]),
				ReturnType: ValueKind(v.Bytecode[ip+2]),
			}
			ip += 3 // Skip param count and return type
			// If it's a struct return type, extract the struct name
			if signature.ReturnType == ValueStruct {
				startPos := ip
				for ip < uint(len(v.Bytecode)) && v.Bytecode[ip] != 0 {
					ip++
				}
				signature.ReturnStructName = string(v.Bytecode[startPos:ip])
				ip++                   // Skip the null terminator
				signature.Address = ip // Address points after struct name
			} else {
				signature.Address = ip // Standard address after metadata
			}

			if isMain && foundMain {
				log.Fatal("Multiple main functions")
			} else if isMain && signature.ReturnType != ValueVoid {
				log.Fatal("Main function should always be void")
			} else if isMain {
				mainAddr = signature.Address
				foundMain = true
				signature.isMain = true
			}
			v.Functions[signature.Address] = signature
		} else {
			ip++
		}
	}
	if !foundMain {
		log.Fatal("No main function found")
	}
	v.Ip = mainAddr
}

func (v *VM) buildStructsTable() {
	ip := uint(0)
	for ip < uint(len(v.Bytecode)) {
		if v.Bytecode[ip] == byte(DEFSTRUCT) {
			ip++
			start := ip
			for v.Bytecode[ip] != 0 {
				ip++
			}
			currentOffset := uint(0)
			structName := string(v.Bytecode[start:ip])
			ip++
			fieldsNumber := uint8(v.Bytecode[ip])
			ip++
			fields := make([]StructField, fieldsNumber)
			for i := 0; i < int(fieldsNumber); i++ {
				start := ip
				for v.Bytecode[ip] != 0 {
					ip++
				}
				fieldName := string(v.Bytecode[start:ip])
				ip++
				fieldType := ValueKind(v.Bytecode[ip])
				ip++

				var arrayType *ValueKind

				if fieldType == ValueArray {
					elemType := ValueKind(v.Bytecode[ip])
					arrayType = &elemType
					ip++
				}

				fields[i] = StructField{
					Name:      fieldName,
					Type:      fieldType,
					Offset:    currentOffset,
					ArrayType: arrayType,
				}
				currentOffset += uint(heap.GetElementSize(fieldType))
			}
			structType := StructType{
				Name:    structName,
				Fields:  fields,
				Size:    currentOffset,
				Methods: make(map[string]uint),
			}
			v.Structs[structName] = structType
		} else {
			ip++
		}
	}
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
	// reader := bufio.NewReader(os.Stdin)
	for v.Running {
		// reader.ReadString('\n')
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
	fmt.Println("========================================")
	fmt.Printf("Functions table:\n")
	for _, f := range v.Functions {
		fmt.Println(f)
	}
	fmt.Println("========================================")
	fmt.Println()
	fmt.Printf("Structs table:\n")
	for name, s := range v.Structs {
		fmt.Printf("Struct %s:\n", name)
		fmt.Printf("  Size: %d bytes\n", s.Size)
		fmt.Printf("  Fields:\n")
		for _, field := range s.Fields {
			fmt.Printf("    %s: type=%v offset=%d\n",
				field.Name, field.Type, field.Offset)
		}
		fmt.Printf("  Methods:\n")
		for methodName, addr := range s.Methods {
			fmt.Printf("    %s: address=%d\n", methodName, addr)
		}
		fmt.Println()
	}
	fmt.Println("========================================")
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
	// v.debugState(opcode)
	// v.Heap.Debug()
	switch opcode {
	case HALT:
		v.Running = false
	case PUSH:
		typeTag := v.getByte()
		var val Value
		switch ValueKind(typeTag) {
		case ValueInt32:
			bits := v.extractUInt32()
			val = Value{Kind: ValueInt32, Raw: bits}
		case ValueFloat32:
			bits := v.extractUInt32()
			val = Value{Kind: ValueFloat32, Raw: bits}
		case ValueByte:
			bits := uint32(v.getByte())
			val = Value{Kind: ValueByte, Raw: bits}
		default:
			log.Fatalf("Unsupported type in PUSH: %v", ValueKind(typeTag))
		}
		v.push(val)
	case POP:
		fmt.Printf("aparent apelam pop din switch DE CE: %d\n", v.Ip)
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
		calleAddr := uint(v.extractUInt16())
		signature, exists := v.Functions[calleAddr]
		if !exists {
			log.Fatalf("function not found at address: %d\n", calleAddr)
		}
		var args []Value
		for i := 0; i < int(signature.ParamCount); i++ {
			args = append(args, v.pop())
		}
		returnAddress := v.Ip
		frame := StackFrame{
			Locals:        make(map[uint16]Value),
			ReturnAddress: returnAddress,
		}
		v.CallStack = append(v.CallStack, frame)
		for i := len(args) - 1; i >= 0; i-- {
			v.push(args[i])
		}
		v.Ip = calleAddr
	case RET:
		if len(v.CallStack) == 0 {
			log.Fatal("Cannot RET: callstack empty")
		}
		calleeFrame := v.getCurrentFrame()
		if len(calleeFrame.LocalStack) == 0 {
			log.Fatal("Cannot RET: local stack empty")
		}
		returnValue := calleeFrame.LocalStack[len(calleeFrame.LocalStack)-1]
		// Check for sentinel value (program termination)
		if calleeFrame.ReturnAddress == 0xFFFFFFFF {
			fmt.Println("Program execution complete - returning from main")
			v.Running = false
			return
		}
		var foundCallee bool
		var calleeReturnType ValueKind
		var calleeReturnStructName string
		currentFuncStart := v.Ip - 1
		for addr, sig := range v.Functions {
			if currentFuncStart >= addr && currentFuncStart < uint(len(v.Bytecode)) {
				calleeReturnType = sig.ReturnType
				calleeReturnStructName = sig.ReturnStructName
				foundCallee = true
				break
			}
		}
		// --- Return type checking ---
		if foundCallee {
			// Check if return value matches expected type
			if calleeReturnType != returnValue.Kind {
				// Special case for struct returns
				if calleeReturnType == ValueStruct && returnValue.Kind == ValuePtr {
					// Verify the struct type matches
					mem, exists := v.Heap.Memory[returnValue.AsPtr()]
					if !exists || ValueKind(mem[0]) != ValueStruct {
						log.Fatalf("Return type mismatch: expected struct %s, got %v",
							calleeReturnStructName, returnValue.Kind)
					}
				} else {
					// Regular type mismatch
					log.Fatalf("Return type mismatch: function has return type %v, but returning %v",
						calleeReturnType, returnValue.Kind)
				}
			} else {
				fmt.Printf("Return value type check passed: %v\n", returnValue.Kind)
			}
		} else {
			fmt.Printf("Warning: Could not determine current function for return type checking\n")
		}
		// This is to find the function we are returning TO (the caller)
		v.CallStack = v.CallStack[:len(v.CallStack)-1]
		if len(v.CallStack) == 0 {
			log.Fatal("Cannot RET: callstack empty after popping frame")
		}
		// Push return value onto caller's stack
		callerFrame := v.getCurrentFrame()
		callerFrame.LocalStack = append(callerFrame.LocalStack, returnValue)
		// Set IP to return address
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
		if err != nil {
			log.Fatal(err)
		}
		v.push(PtrValue(ptr))
	case FREE:
		ptr := v.pop().AsPtr()
		err := v.Heap.Free(ptr)
		if err != nil {
			log.Fatal(err)
		}
	case LOADH:
		ptr := v.pop().AsPtr()
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
	case STRALLOC:
		length := v.extractUInt16()
		data := string(v.Bytecode[v.Ip : v.Ip+uint(length)])
		v.Ip += uint(length)
		ptr, err := v.Heap.AllocateString(data)
		if err != nil {
			log.Fatal(err)
		}
		v.push(PtrValue(ptr))
	case NEWARR:
		elementKind := ValueKind(v.getByte())
		length := v.pop().AsInt32()
		ptr, err := v.Heap.AllocateArray(elementKind, length)
		if err != nil {
			log.Fatal(err)
		}
		v.push(PtrValue(ptr))
	case LDELEM:
		index := v.pop().AsInt32()
		arrayPtr := v.pop().AsPtr()
		value, err := v.Heap.GetArrayElement(arrayPtr, index)
		if err != nil {
			log.Fatal(err)
		}
		v.push(*value)
	case STELEM:
		value := v.pop()
		index := v.pop().AsInt32()
		arrayPtr := v.pop().AsPtr()
		err := v.Heap.SetArrayElement(arrayPtr, index, value)
		if err != nil {
			log.Fatal(err)
		}
	case SYSCALL:
		call := Systemcall(v.extractUInt16())
		v.executeSystemCall(call)
	case NEWSTRUCT:
		typeName := v.extractString()
		structType, ok := v.Structs[typeName]
		if !ok {
			log.Fatalf("Unkown struct type: %s", typeName)
		}
		ptr, err := v.Heap.AllocateStruct(structType)
		if err != nil {
			log.Fatal(err)
		}
		v.push(PtrValue(ptr))
	case FLDGET:
		fieldName := v.extractString()
		structPtr := v.pop().AsPtr()
		value, err := v.Heap.GetStructField(structPtr, fieldName)
		if err != nil {
			log.Fatal(err)
		}
		v.push(*value)
	case STFIELD:
		value := v.pop()
		structPtr := v.pop().AsPtr()
		fieldName := v.extractString()
		err := v.Heap.SetStructureField(structPtr, fieldName, value)
		if err != nil {
			log.Fatal(err)
		}
	case FUNC:
		v.Ip += 5
	}
}
