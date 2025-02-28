package assembler

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math"
	"strings"
	"testing"

	. "stack_vm/common"
	"stack_vm/vm"
)

// Helper function to create a parsed program for testing
func createTestProgram() *Program {
	return &Program{
		Structs:   []StructType{},
		Functions: []ParsedFunction{},
	}
}

// Helper function to add a struct to a test program
func addTestStruct(prog *Program, name string, fields ...StructField) {
	structType := StructType{
		Name:    name,
		Fields:  fields,
		Methods: make(map[string]uint),
	}
	prog.Structs = append(prog.Structs, structType)
}

// Helper function to add a function to a test program
func addTestFunction(prog *Program, name string, returnType ValueKind, params []ParsedParam,
	instructions []Instruction, labels map[string]int) {
	function := ParsedFunction{
		Name:       name,
		ReturnType: returnType,
		Params:     params,
		Body:       instructions,
		Labels:     labels,
	}
	prog.Functions = append(prog.Functions, function)
}

// Helper to extract bytes from bytecode at a specific position
func bytesAt(bytecode []byte, start, length int) []byte {
	if start+length > len(bytecode) {
		return bytecode[start:]
	}
	return bytecode[start : start+length]
}

// Helper to create a token for testing
func createToken(tokenType TokenType, literal string) Token {
	return Token{
		Type:    tokenType,
		Literal: literal,
		Line:    1,
		Column:  1,
	}
}

// Helper to create an instruction for testing
func createInstruction(opcode vm.Opcode, operands ...Token) Instruction {
	return Instruction{
		Token:    Token{Type: IDENT, Literal: "test"},
		Opcode:   opcode,
		Operands: operands,
	}
}

// TestEmptyProgram tests generating bytecode for an empty program
func TestEmptyProgram(t *testing.T) {
	prog := createTestProgram()

	generator := NewCodeGenerator(prog)
	bytecode, err := generator.Generate()

	if err != nil {
		t.Fatalf("Expected no error for empty program, got: %v", err)
	}

	if len(bytecode) != 1 {
		t.Fatalf("Expected empty bytecode for empty program, got %d bytes", len(bytecode))
	}
}

// TestBasicStructDefinition tests generating bytecode for a struct definition
func TestBasicStructDefinition(t *testing.T) {
	prog := createTestProgram()

	// Add a simple Point struct
	addTestStruct(prog, "Point",
		StructField{Name: "x", Type: ValueInt32, Offset: 0},
		StructField{Name: "y", Type: ValueInt32, Offset: 4},
	)

	generator := NewCodeGenerator(prog)
	bytecode, err := generator.Generate()

	if err != nil {
		t.Fatalf("Failed to generate bytecode: %v", err)
	}

	// Check that the bytecode begins with a DEFSTRUCT opcode
	if len(bytecode) == 0 || bytecode[0] != byte(vm.DEFSTRUCT) {
		t.Fatal("Expected bytecode to start with DEFSTRUCT opcode")
	}

	// Check that the struct name is encoded correctly
	nameBytes := []byte("Point\x00") // Name followed by null terminator
	if !bytes.Contains(bytecode, nameBytes) {
		t.Fatalf("Expected struct name 'Point' in bytecode, not found")
	}
}

// TestEmptyFunction tests generating bytecode for an empty function
func TestEmptyFunction(t *testing.T) {
	prog := createTestProgram()

	// Add an empty main function
	addTestFunction(prog, "main", ValueVoid, []ParsedParam{}, []Instruction{}, map[string]int{})

	generator := NewCodeGenerator(prog)
	bytecode, err := generator.Generate()

	if err != nil {
		t.Fatalf("Failed to generate bytecode: %v", err)
	}

	// Check that the bytecode begins with a FUNC opcode
	if len(bytecode) == 0 || bytecode[0] != byte(vm.FUNC) {
		t.Fatal("Expected bytecode to start with FUNC opcode")
	}

	// Check that the function is marked as MAIN
	if len(bytecode) < 2 || bytecode[1] != byte(vm.FUNC_MAIN) {
		t.Fatal("Expected main function marker")
	}
}

// TestPushInstruction tests generating bytecode for push instructions
func TestPushInstruction(t *testing.T) {
	prog := createTestProgram()

	// Create a function with push instructions
	instructions := []Instruction{
		createInstruction(vm.PUSH,
			createToken(INT32, "int32"),
			createToken(INT, "42")),
		createInstruction(vm.PUSH,
			createToken(FLOAT32, "float32"),
			createToken(FLOAT, "3.14")),
	}

	addTestFunction(prog, "main", ValueVoid, []ParsedParam{}, instructions, map[string]int{})

	generator := NewCodeGenerator(prog)
	bytecode, err := generator.Generate()

	if err != nil {
		t.Fatalf("Failed to generate bytecode: %v", err)
	}

	// Find position after function header
	funcHeaderSize := 5 // FUNC(1) + FUNC_TYPE(1) + PARAM_COUNT(2) + RETURN_TYPE(1)
	codeStart := funcHeaderSize

	// Check first push instruction
	if bytecode[codeStart] != byte(vm.PUSH) {
		t.Fatalf("Expected PUSH opcode at position %d, got %d", codeStart, bytecode[codeStart])
	}

	// Check that the type is int32
	if bytecode[codeStart+1] != byte(ValueInt32) {
		t.Fatalf("Expected int32 type after PUSH, got %d", bytecode[codeStart+1])
	}

	// Check the value 42
	intBytes := bytesAt(bytecode, codeStart+2, 4)
	value := binary.BigEndian.Uint32(intBytes)
	if int32(value) != 42 {
		t.Fatalf("Expected value 42, got %d", int32(value))
	}

	// Check second push instruction
	pushFloatPos := codeStart + 6 // PUSH(1) + TYPE(1) + VALUE(4)
	if bytecode[pushFloatPos] != byte(vm.PUSH) {
		t.Fatalf("Expected second PUSH opcode at position %d, got %d",
			pushFloatPos, bytecode[pushFloatPos])
	}

	// Check that the type is float32
	if bytecode[pushFloatPos+1] != byte(ValueFloat32) {
		t.Fatalf("Expected float32 type, got %d", bytecode[pushFloatPos+1])
	}

	// Check the value 3.14
	floatBytes := bytesAt(bytecode, pushFloatPos+2, 4)
	floatBits := binary.BigEndian.Uint32(floatBytes)
	floatValue := math.Float32frombits(floatBits)
	if floatValue != 3.14 {
		t.Fatalf("Expected value 3.14, got %f", floatValue)
	}
}

// TestArithmeticInstructions tests generating bytecode for arithmetic instructions
func TestArithmeticInstructions(t *testing.T) {
	prog := createTestProgram()

	// Create a function with arithmetic instructions
	instructions := []Instruction{
		createInstruction(vm.IADD),
		createInstruction(vm.ISUB),
		createInstruction(vm.IMUL),
		createInstruction(vm.IDIV),
		createInstruction(vm.FADD),
		createInstruction(vm.FSUB),
		createInstruction(vm.FMUL),
		createInstruction(vm.FDIV),
	}

	addTestFunction(prog, "main", ValueVoid, []ParsedParam{}, instructions, map[string]int{})

	generator := NewCodeGenerator(prog)
	bytecode, err := generator.Generate()

	if err != nil {
		t.Fatalf("Failed to generate bytecode: %v", err)
	}

	// Find position after function header
	funcHeaderSize := 5 // FUNC(1) + FUNC_TYPE(1) + PARAM_COUNT(2) + RETURN_TYPE(1)
	codeStart := funcHeaderSize

	// Check that each arithmetic instruction is encoded correctly
	expectedOpcodes := []vm.Opcode{
		vm.IADD, vm.ISUB, vm.IMUL, vm.IDIV,
		vm.FADD, vm.FSUB, vm.FMUL, vm.FDIV,
	}

	for i, opcode := range expectedOpcodes {
		pos := codeStart + i
		if pos >= len(bytecode) {
			t.Fatalf("Bytecode too short, expected opcode at position %d", pos)
		}
		if bytecode[pos] != byte(opcode) {
			t.Errorf("Expected opcode %v at position %d, got %v",
				opcode, pos, vm.Opcode(bytecode[pos]))
		}
	}
}

// TestJumpInstructions tests generating bytecode for jump instructions
func TestJumpInstructions(t *testing.T) {
	prog := createTestProgram()

	// Create a function with labels and jumps
	labels := map[string]int{
		"label1": 2, // Position of the third instruction
	}

	instructions := []Instruction{
		createInstruction(vm.PUSH, createToken(INT32, "int32"), createToken(INT, "1")),
		createInstruction(vm.JMP, createToken(IDENT, "label1")),
		// This is where label1 points to:
		createInstruction(vm.PUSH, createToken(INT32, "int32"), createToken(INT, "2")),
		createInstruction(vm.IJE, createToken(IDENT, "label1"), createToken(INT, "42")),
	}

	addTestFunction(prog, "main", ValueVoid, []ParsedParam{}, instructions, labels)

	generator := NewCodeGenerator(prog)
	bytecode, err := generator.Generate()

	if err != nil {
		t.Fatalf("Failed to generate bytecode: %v", err)
	}

	// Find position after function header
	funcHeaderSize := 5 // FUNC(1) + FUNC_TYPE(1) + PARAM_COUNT(2) + RETURN_TYPE(1)

	// Skip first PUSH instruction
	pushSize := 6 // PUSH(1) + TYPE(1) + VALUE(4)
	jumpPos := funcHeaderSize + pushSize

	// Check JMP instruction
	if bytecode[jumpPos] != byte(vm.JMP) {
		t.Fatalf("Expected JMP opcode, got %v", vm.Opcode(bytecode[jumpPos]))
	}

	// Check jump target (should be 2 for label1)
	labelBytes := bytesAt(bytecode, jumpPos+1, 2)
	labelIndex := binary.BigEndian.Uint16(labelBytes)
	if int(labelIndex) != 2 {
		t.Fatalf("Expected jump to label index 2, got %d", labelIndex)
	}

	// Find IJE instruction
	ijePos := jumpPos + 3 + pushSize // JMP(1) + TARGET(2) + next PUSH(6)

	// Check IJE instruction
	if bytecode[ijePos] != byte(vm.IJE) {
		t.Fatalf("Expected IJE opcode, got %v", vm.Opcode(bytecode[ijePos]))
	}

	// Check conditional jump value (should be 42)
	valueBytes := bytesAt(bytecode, ijePos+3, 4) // IJE(1) + TARGET(2) + VALUE(4)
	value := binary.BigEndian.Uint32(valueBytes)
	if int32(value) != 42 {
		t.Fatalf("Expected conditional value 42, got %d", int32(value))
	}
}

// TestStringAllocation tests generating bytecode for string allocation
func TestStringAllocation(t *testing.T) {
	prog := createTestProgram()

	// Create a function with string allocation
	instructions := []Instruction{
		createInstruction(vm.STRALLOC, createToken(STRING, "Hello, World!")),
	}

	addTestFunction(prog, "main", ValueVoid, []ParsedParam{}, instructions, map[string]int{})

	generator := NewCodeGenerator(prog)
	bytecode, err := generator.Generate()

	if err != nil {
		t.Fatalf("Failed to generate bytecode: %v", err)
	}

	// Find position after function header
	funcHeaderSize := 5 // FUNC(1) + FUNC_TYPE(1) + PARAM_COUNT(2) + RETURN_TYPE(1)

	// Check STRALLOC instruction
	if bytecode[funcHeaderSize] != byte(vm.STRALLOC) {
		t.Fatalf("Expected STRALLOC opcode, got %v", vm.Opcode(bytecode[funcHeaderSize]))
	}

	// Check string length (should be 13 for "Hello, World!")
	lengthBytes := bytesAt(bytecode, funcHeaderSize+1, 2)
	length := binary.BigEndian.Uint16(lengthBytes)
	if int(length) != 13 {
		t.Fatalf("Expected string length 13, got %d", length)
	}

	// Check string content
	strStart := funcHeaderSize + 3 // STRALLOC(1) + LENGTH(2)
	strBytes := bytesAt(bytecode, strStart, 13)
	str := string(strBytes)
	if str != "Hello, World!" {
		t.Fatalf("Expected string 'Hello, World!', got '%s'", str)
	}
}

// TestStructOperations tests generating bytecode for struct operations
func TestStructOperations(t *testing.T) {
	prog := createTestProgram()

	// Add a Point struct
	addTestStruct(prog, "Point",
		StructField{Name: "x", Type: ValueInt32, Offset: 0},
		StructField{Name: "y", Type: ValueInt32, Offset: 4},
	)

	// Create a function with struct operations
	instructions := []Instruction{
		createInstruction(vm.NEWSTRUCT, createToken(IDENT, "Point")),
		createInstruction(vm.FLDGET, createToken(STRING, "x")),
		createInstruction(vm.STFIELD, createToken(STRING, "y")),
	}

	addTestFunction(prog, "main", ValueVoid, []ParsedParam{}, instructions, map[string]int{})

	generator := NewCodeGenerator(prog)
	bytecode, err := generator.Generate()

	if err != nil {
		t.Fatalf("Failed to generate bytecode: %v", err)
	}

	// Skip struct definition and function header to find the code
	structDefSize := 1 + len("Point") + 1 + 1 + // DEFSTRUCT(1) + "Point\0" + FIELD_COUNT(1)
		len("x") + 1 + 1 + // "x\0" + TYPE(1)
		len("y") + 1 + 1 // "y\0" + TYPE(1)

	funcHeaderSize := 5 // FUNC(1) + FUNC_TYPE(1) + PARAM_COUNT(2) + RETURN_TYPE(1)
	codeStart := structDefSize + funcHeaderSize

	// Check NEWSTRUCT instruction
	if bytecode[codeStart] != byte(vm.NEWSTRUCT) {
		t.Fatalf("Expected NEWSTRUCT opcode, got %v", vm.Opcode(bytecode[codeStart]))
	}

	// Check struct name
	pos := codeStart + 1
	if !bytes.Contains(bytecode[pos:], []byte("Point\x00")) {
		t.Fatalf("Expected struct name 'Point' in bytecode")
	}

	// Find remaining instructions (positions depend on string lengths)
	bytecodeHex := hex.EncodeToString(bytecode)

	// Check for FLDGET
	fldgetHex := hex.EncodeToString([]byte{byte(vm.FLDGET)})
	if !strings.Contains(bytecodeHex, fldgetHex+"78"+hex.EncodeToString([]byte{0})) {
		// "78" is ASCII 'x' in hex, followed by null terminator
		t.Fatalf("FLDGET instruction or field name not found in bytecode")
	}

	// Check for STFIELD
	stfieldHex := hex.EncodeToString([]byte{byte(vm.STFIELD)})
	if !strings.Contains(bytecodeHex, stfieldHex+"79"+hex.EncodeToString([]byte{0})) {
		// "79" is ASCII 'y' in hex, followed by null terminator
		t.Fatalf("STFIELD instruction or field name not found in bytecode")
	}
}

// TestCompleteProgram tests generating bytecode for a complete program
func TestCompleteProgram(t *testing.T) {
	source := `.structs
		struct Point {
			x: int32
			y: int32
		}
	.text
		func calculateSum(a: int32, b: int32) -> int32 {
			iadd
			ret
		}
		
		func main() -> void {
			push int32 10
			push int32 20
			call calculateSum
			ret
		}`

	lexer := NewLexer(source)
	parser := NewParser(lexer)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	generator := NewCodeGenerator(program)
	bytecode, err := generator.Generate()

	if err != nil {
		t.Fatalf("Failed to generate bytecode: %v", err)
	}

	// Basic validation: ensure we got some bytecode
	if len(bytecode) == 0 {
		t.Fatal("Generated bytecode is empty")
	}

	// Verify DEFSTRUCT is present
	if !bytes.Contains(bytecode, []byte{byte(vm.DEFSTRUCT)}) {
		t.Error("DEFSTRUCT opcode not found in bytecode")
	}

	// Verify both functions are present - look for their headers
	funcCount := bytes.Count(bytecode, []byte{byte(vm.FUNC)})
	if funcCount != 2 {
		t.Errorf("Expected 2 function definitions, found %d", funcCount)
	}

	// Verify arithmetic, push, call, and ret instructions are present
	opcodes := []vm.Opcode{vm.IADD, vm.PUSH, vm.CALL, vm.RET}
	for _, op := range opcodes {
		if !bytes.Contains(bytecode, []byte{byte(op)}) {
			t.Errorf("Opcode %v not found in bytecode", op)
		}
	}
}

// TestInvalidInstructions tests error handling for invalid instructions
func TestInvalidInstructions(t *testing.T) {
	tests := []struct {
		name        string
		instruction Instruction
		errSubstr   string
	}{
		{
			name:        "push without operands",
			instruction: createInstruction(vm.PUSH), // Missing type and value
			errSubstr:   "two operands",
		},
		{
			name:        "call with unknown function",
			instruction: createInstruction(vm.CALL, createToken(IDENT, "nonexistent")),
			errSubstr:   "undefined function",
		},
		{
			name:        "jump to unknown label",
			instruction: createInstruction(vm.JMP, createToken(IDENT, "nonexistent")),
			errSubstr:   "undefined label",
		},
		{
			name:        "newstruct with unknown struct",
			instruction: createInstruction(vm.NEWSTRUCT, createToken(IDENT, "NonexistentStruct")),
			errSubstr:   "undefined struct",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			prog := createTestProgram()
			// Add a test function with the invalid instruction
			instructions := []Instruction{test.instruction}
			addTestFunction(prog, "main", ValueVoid, []ParsedParam{}, instructions, map[string]int{})

			generator := NewCodeGenerator(prog)
			_, err := generator.Generate()

			// Should get an error
			if err == nil {
				t.Fatal("Expected error for invalid instruction, got nil")
			}

			// Error message should contain the expected substring
			if !strings.Contains(err.Error(), test.errSubstr) {
				t.Fatalf("Expected error containing '%s', got: %v", test.errSubstr, err)
			}
		})
	}
}
