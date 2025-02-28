package assembler

import (
	"encoding/binary"
	"fmt"
	"math"
	. "stack_vm/common"
	"stack_vm/vm"
	"strconv"
)

type CodeGenerator struct {
	program         *Program
	bytecode        []byte
	functionTable   map[string]uint
	structTable     map[string]StructType
	currentFunction *ParsedFunction
}

func NewCodeGenerator(program *Program) *CodeGenerator {
	return &CodeGenerator{
		program:       program,
		bytecode:      []byte{},
		functionTable: make(map[string]uint),
		structTable:   make(map[string]StructType),
	}
}

func (g *CodeGenerator) Generate() ([]byte, error) {
	if err := g.defineStructs(); err != nil {
		return nil, err
	}
	for _, function := range g.program.Functions {
		g.emitByte(byte(vm.FUNC))
		if function.Name == "main" {
			g.emitByte(byte(vm.FUNC_MAIN))
		} else {
			g.emitByte(byte(vm.FUNC_NORMAL))
		}
		paramCountByte := make([]byte, 2)
		binary.BigEndian.PutUint16(paramCountByte, uint16(len(function.Params)))
		g.emitBytes(paramCountByte)
		g.emitByte(byte(function.ReturnType))
		if function.ReturnType == ValueStruct {
			if _, exists := g.structTable[function.ReturnStructName]; !exists {
				return nil, fmt.Errorf("undefined struct return type: %s in function %s", function.ReturnStructName, function.Name)
			}
			g.emitString(function.ReturnStructName)
		}
		bodyStart := uint(len(g.bytecode))
		g.functionTable[function.Name] = bodyStart
		g.currentFunction = &function
		for _, instruction := range function.Body {
			if err := g.generateInstruction(instruction); err != nil {
				return nil, fmt.Errorf("error generating instruction %v: %w", instruction, err)
			}
		}
	}
	g.emitByte(byte(vm.HALT))
	return g.bytecode, nil
}

func (g *CodeGenerator) defineStructs() error {
	for _, structDef := range g.program.Structs {
		g.structTable[structDef.Name] = structDef
		g.emitByte(byte(vm.DEFSTRUCT))
		g.emitString(structDef.Name)
		g.emitByte(byte(len(structDef.Fields)))
		for _, field := range structDef.Fields {
			g.emitString(field.Name)
			g.emitByte(byte(field.Type))
			if field.ArrayType != nil {
				g.emitByte(byte(ValueArray))
				g.emitByte(byte(*field.ArrayType))
			}
		}
	}
	return nil
}

func (g *CodeGenerator) generateFunctionMetadata() error {
	for _, function := range g.program.Functions {
		// Record function header start position
		// Emit FUNC opcode
		g.emitByte(byte(vm.FUNC))

		// Emit function type (main or normal)
		if function.Name == "main" {
			g.emitByte(byte(vm.FUNC_MAIN))
		} else {
			g.emitByte(byte(vm.FUNC_NORMAL))
		}
		// Emit parameter count
		paramCountBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(paramCountBytes, uint16(len(function.Params)))
		g.emitBytes(paramCountBytes)
		// Emit return type
		g.emitByte(byte(function.ReturnType))
		// For struct returns, emit the struct name
		if function.ReturnType == ValueStruct {
			if _, exists := g.structTable[function.ReturnStructName]; !exists {
				return fmt.Errorf("undefined struct return type: %s in function %s",
					function.ReturnStructName, function.Name)
			}
			g.emitString(function.ReturnStructName)
		}
		// Record where this function's body will begin
		bodyStart := uint(len(g.bytecode))
		g.functionTable[function.Name] = bodyStart
	}
	return nil
}

// Improved method to generate all function bodies with accurate address tracking
func (g *CodeGenerator) generateAllFunctionBodies() error {
	for _, function := range g.program.Functions {
		bodyAddr := g.functionTable[function.Name]
		currentPos := uint(len(g.bytecode))
		if currentPos != bodyAddr {
			g.functionTable[function.Name] = currentPos
		}
		g.currentFunction = &function
		for _, instruction := range function.Body {

			if err := g.generateInstruction(instruction); err != nil {
				return fmt.Errorf("error generating instruction %v: %w", instruction, err)
			}
		}
	}
	return nil
}

func (g *CodeGenerator) generateInstruction(inst Instruction) error {
	g.emitByte(byte(inst.Opcode))
	switch inst.Opcode {
	case vm.PUSH:
		if len(inst.Operands) != 2 {
			return fmt.Errorf("push requires two operands, got %d", len(inst.Operands))
		}
		typeToken := inst.Operands[0]
		valueToken := inst.Operands[1]
		if typeToken.Type == INT32 {
			g.emitByte(byte(ValueInt32))
			value, err := parseInt32(valueToken.Literal)
			if err != nil {
				return err
			}
			g.emitInt32(value)
		} else if typeToken.Type == FLOAT32 {
			g.emitByte(byte(ValueFloat32))
			value, err := parseFloat32(valueToken.Literal)
			if err != nil {
				return err
			}
			g.emitFloat32(value)
		} else if typeToken.Type == BYTE_TYPE {
			g.emitByte(byte(ValueByte))
			value, err := parseInt32(valueToken.Literal)
			if err != nil {
				return err
			}
			if value < 0 || value > 255 {
				return fmt.Errorf("byte value out of range: %d", value)
			}
			g.emitByte(byte(value))
		} else {
			return fmt.Errorf("unsupported type in push: %v", typeToken.Type)
		}
	case vm.STORE, vm.LOAD:
		if len(inst.Operands) != 1 {
			return fmt.Errorf("%v requires one operand, got %d", inst.Opcode, len(inst.Operands))
		}
		addrToken := inst.Operands[0]
		addr, err := parseInt32(addrToken.Literal)
		if err != nil {
			return err
		}
		g.emitUint16(uint16(addr))
	case vm.CALL:
		if len(inst.Operands) != 1 {
			return fmt.Errorf("call requires one operand, got %d", len(inst.Operands))
		}
		funcName := inst.Operands[0].Literal
		funcAddr, exists := g.functionTable[funcName]
		if !exists {
			return fmt.Errorf("undefined functionn: %d", funcAddr)
		}
		g.emitUint16(uint16(funcAddr))
	case vm.JMP, vm.IJE, vm.IJNE, vm.FJE, vm.FJNE:
		if len(inst.Operands) < 1 {
			return fmt.Errorf("jump requires at least one operand, got %d", len(inst.Operands))
		}
		labelName := inst.Operands[0].Literal
		labelPos, exists := g.currentFunction.Labels[labelName]
		if !exists {
			return fmt.Errorf("undefined label: %s", labelName)
		}
		g.emitUint16(uint16(labelPos))
		if inst.Opcode != vm.JMP {
			if len(inst.Operands) != 2 {
				return fmt.Errorf("conditional jump requires two operands, got %d", len(inst.Operands))
			}
			valueToken := inst.Operands[1]
			if valueToken.Type == INT {
				value, err := parseInt32(valueToken.Literal)
				if err != nil {
					return err
				}
				g.emitInt32(value)
			} else if valueToken.Type == FLOAT {
				value, err := parseFloat32(valueToken.Literal)
				if err != nil {
					return err
				}
				g.emitFloat32(value)
			} else {
				return fmt.Errorf("unsupported type in conditional jump: %v", valueToken.Type)
			}
		}
	case vm.STRALLOC:
		if len(inst.Operands) != 1 {
			return fmt.Errorf("stralloc requires one operand, got %d", len(inst.Operands))
		}
		str := inst.Operands[0]
		g.emitUint16(uint16(len(str.Literal)))
		g.bytecode = append(g.bytecode, []byte(str.Literal)...)
	case vm.SYSCALL:
		if len(inst.Operands) != 1 {
			return fmt.Errorf("syscall requires one operand(syscall number or name), got %d", len(inst.Operands))
		}
		syscall, err := strconv.ParseUint(inst.Operands[0].Literal, 10, 16)
		if err != nil {
			return fmt.Errorf("invalid syscall number: %s", inst.Operands[0].Literal)
		}
		g.emitUint16(uint16(syscall))
	case vm.NEWARR:
		if len(inst.Operands) != 1 {
			return fmt.Errorf("newarr requires one operand, got %d", len(inst.Operands))
		}
		typeToken := inst.Operands[0]
		if typeToken.Type == INT32 {
			g.emitByte(byte(ValueInt32))
		} else if typeToken.Type == FLOAT32 {
			g.emitByte(byte(ValueFloat32))
		} else {
			return fmt.Errorf("unsupported type in newarr: %v", typeToken.Type)
		}
	case vm.NEWSTRUCT:
		if len(inst.Operands) != 1 {
			return fmt.Errorf("newstruct requires one operand, got %d", len(inst.Operands))
		}
		structName := inst.Operands[0].Literal
		if _, exists := g.structTable[structName]; !exists {
			return fmt.Errorf("undefined struct: %s", structName)
		}
		g.emitString(structName)
	case vm.STFIELD, vm.FLDGET:
		if len(inst.Operands) != 1 {
			return fmt.Errorf("field access requires one operand, got %d", len(inst.Operands))
		}
		fieldName := inst.Operands[0].Literal
		g.emitString(fieldName)
	case vm.LDELEM, vm.STELEM:
	case vm.ALLOC:
		// ALLOC takes no explicit operands - it uses the value on top of the stack
	case vm.FREE:
		// FREE takes no explicit operands - it uses the value on top of the stack
	case vm.LOADH:
		// LOADH takes no explicit operands - it uses the value on top of the stack
	case vm.STOREH:
		// STOREH takes no explicit operands - it uses values on top of the stack
	case vm.POP:
		// POP takes no operands
	case vm.DUP:
		// DUP takes no operands
	case vm.IADD, vm.ISUB, vm.IMUL, vm.IDIV,
		vm.FADD, vm.FSUB, vm.FMUL, vm.FDIV:
		// Arithmetic operations take no operands
	case vm.EQ, vm.NE, vm.LT, vm.LE, vm.GT, vm.GE:
		// Comparison operations take no operands
	}
	return nil
}

func (g *CodeGenerator) emitByte(b byte) {
	g.bytecode = append(g.bytecode, b)
}

func (g *CodeGenerator) emitBytes(b []byte) {
	g.bytecode = append(g.bytecode, b...)
}

func (g *CodeGenerator) emitString(s string) {
	g.bytecode = append(g.bytecode, []byte(s)...)
	g.emitByte(0)
}

func (g *CodeGenerator) emitRawString(s string) {
	g.bytecode = append(g.bytecode, []byte(s)...)
}

func (g *CodeGenerator) emitInt32(value int32) {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, uint32(value))
	g.emitBytes(bytes)
}

func (g *CodeGenerator) emitFloat32(value float32) {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, math.Float32bits(value))
	g.emitBytes(bytes)
}

func (g *CodeGenerator) emitUint16(value uint16) {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, value)
	g.emitBytes(bytes)
}

func parseInt32(literal string) (int32, error) {
	var i int64
	var err error
	if i, err = strconv.ParseInt(literal, 10, 32); err != nil {
		return 0, fmt.Errorf("invalid integer: %s", literal)
	}
	return int32(i), nil
}

func parseFloat32(literal string) (float32, error) {
	var value float64
	_, err := fmt.Sscanf(literal, "%f", &value)
	if err != nil {
		return 0, fmt.Errorf("invalid float: %s", literal)
	}
	if value < -math.MaxFloat32 || value > math.MaxFloat32 {
		return 0, fmt.Errorf("float out of range: %f", value)
	}
	return float32(value), nil
}
