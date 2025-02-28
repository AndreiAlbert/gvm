package assembler

import (
	"fmt"
	. "stack_vm/common"
	"stack_vm/vm"
	"strings"
)

type NodeType int

const (
	NODE_PROGRAM NodeType = iota
	NODE_STRUCT
	NODE_FUNCTION
	NODE_INSTRUCTION
	NODE_LABEL
)

type Program struct {
	Structs   []StructType
	Functions []ParsedFunction
}

type ParsedFunction struct {
	Name             string
	Params           []ParsedParam
	ReturnType       ValueKind
	Body             []Instruction
	Labels           map[string]int
	ReturnStructName string
}

type ParsedParam struct {
	Name string
	Type ValueKind
}

type Instruction struct {
	Token    Token
	Opcode   vm.Opcode
	Operands []Token
	isLabel  bool
	Label    string
}

type Parser struct {
	lexer        *Lexer
	currentToken Token
	peekToken    Token
	errors       []string
}

// String method for Program
func (p *Program) String() string {
	var sb strings.Builder
	sb.WriteString("Program {\n")

	sb.WriteString("  Structs: [\n")
	for _, s := range p.Structs {
		lines := strings.Split(s.String(), "\n")
		for _, line := range lines {
			sb.WriteString("    " + line + "\n")
		}
	}
	sb.WriteString("  ]\n")

	sb.WriteString("  Functions: [\n")
	for _, f := range p.Functions {
		lines := strings.Split(f.String(), "\n")
		for _, line := range lines {
			sb.WriteString("    " + line + "\n")
		}
	}
	sb.WriteString("  ]\n")
	sb.WriteString("}")
	return sb.String()
}

func (p *ParsedParam) String() string {
	return fmt.Sprintf("%s: %v", p.Name, p.Type)
}

func (f *ParsedFunction) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Function %s(\n", f.Name))

	// Parameters
	for i, param := range f.Params {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(param.String())
	}
	sb.WriteString(fmt.Sprintf(") -> %v {\n", f.ReturnType))

	// Labels
	if len(f.Labels) > 0 {
		sb.WriteString("  Labels: {\n")
		for label, pos := range f.Labels {
			sb.WriteString(fmt.Sprintf("    %s: %d\n", label, pos))
		}
		sb.WriteString("  }\n")
	}

	// Instructions
	sb.WriteString("  Body: [\n")
	for _, instr := range f.Body {
		lines := strings.Split(instr.String(), "\n")
		for _, line := range lines {
			sb.WriteString("    " + line + "\n")
		}
	}
	sb.WriteString("  ]\n")
	sb.WriteString("}")
	return sb.String()
}

// String method for ParsedParam

// String method for Instruction
func (i *Instruction) String() string {
	var sb strings.Builder
	if i.isLabel {
		sb.WriteString(fmt.Sprintf("Label(%s)", i.Label))
	} else {
		sb.WriteString(fmt.Sprintf("Instruction(%v", i.Opcode))
		if len(i.Operands) > 0 {
			sb.WriteString(" [")
			for j, op := range i.Operands {
				if j > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("%v", op))
			}
			sb.WriteString("]")
		}
		sb.WriteString(")")
	}
	return sb.String()
}

func NewParser(l *Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		errors: []string{},
	}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) expectToken(t TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	return false
}

func TokenTypeToValueKind(t TokenType) ValueKind {
	switch t {
	case INT32:
		return ValueInt32
	case FLOAT32:
		return ValueFloat32
	case VOID:
		return ValueVoid
	case STRING_TYPE:
		return ValueString
	default:
		return ValueVoid
	}
}

func TokenTypeToOpcode(t TokenType) (vm.Opcode, error) {
	switch t {
	case HALT:
		return vm.HALT, nil
	case PUSH:
		return vm.PUSH, nil
	case POP:
		return vm.POP, nil
	case IADD:
		return vm.IADD, nil
	case ISUB:
		return vm.ISUB, nil
	case IMUL:
		return vm.IMUL, nil
	case IDIV:
		return vm.IDIV, nil
	case FADD:
		return vm.FADD, nil
	case FSUB:
		return vm.FSUB, nil
	case FMUL:
		return vm.FMUL, nil
	case FDIV:
		return vm.FDIV, nil
	case JMP:
		return vm.JMP, nil
	case IJNE:
		return vm.IJNE, nil
	case IJE:
		return vm.IJE, nil
	case FJNE:
		return vm.FJNE, nil
	case FJE:
		return vm.FJE, nil
	case EQ:
		return vm.EQ, nil
	case NE:
		return vm.NE, nil
	case LT:
		return vm.LT, nil
	case GT:
		return vm.GT, nil
	case GE:
		return vm.GE, nil
	case LE:
		return vm.LE, nil
	case LOAD:
		return vm.LOAD, nil
	case STORE:
		return vm.STORE, nil
	case CALL:
		return vm.CALL, nil
	case RET:
		return vm.RET, nil
	case ALLOC:
		return vm.ALLOC, nil
	case FREE:
		return vm.FREE, nil
	case LOADH:
		return vm.LOADH, nil
	case STOREH:
		return vm.STOREH, nil
	case DUP:
		return vm.DUP, nil
	case STRALLOC:
		return vm.STRALLOC, nil
	case SYSCALL:
		return vm.SYSCALL, nil
	case NEWARR:
		return vm.NEWARR, nil
	case LDELEM:
		return vm.LDELEM, nil
	case STELEM:
		return vm.STELEM, nil
	case NEWSTRUCT:
		return vm.NEWSTRUCT, nil
	case FLDGET:
		return vm.FLDGET, nil
	case STFIELD:
		return vm.STFIELD, nil
	default:
		return 0, fmt.Errorf("unknown opcode for token type: %v", t)
	}
}

func (p *Parser) Parse() (*Program, error) {
	program := &Program{}
	for p.currentToken.Type != EOF {
		switch p.currentToken.Type {
		case SECTION_STRUCTS:
			p.nextToken()
			for p.currentToken.Type == STRUCT {
				if structDef := p.parseStructDef(); structDef != nil {
					program.Structs = append(program.Structs, *structDef)
				}
			}
		case SECTION_TEXT:
			p.nextToken()
			for p.currentToken.Type == FUNC {
				if function := p.parseFunction(); function != nil {
					program.Functions = append(program.Functions, *function)
				}
			}
		default:
			p.nextToken()
		}
	}
	if len(p.errors) > 0 {
		var errMsg strings.Builder
		errMsg.WriteString("parser encountered the following errors:\n")
		for i, err := range p.errors {
			errMsg.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err))
		}
		return nil, fmt.Errorf(errMsg.String())
	}
	return program, nil
}

func (p *Parser) parseStructDef() *StructType {
	structType := &StructType{
		Methods: make(map[string]uint),
	}
	if !p.expectToken(IDENT) {
		return nil
	}
	structType.Name = p.currentToken.Literal
	if !p.expectToken(LBRACE) {
		return nil
	}
	p.nextToken()
	for p.currentToken.Type != RBRACE && p.currentToken.Type != EOF {
		field := StructField{}
		if p.currentToken.Type != IDENT {
			p.errors = append(p.errors, fmt.Sprintf("expected field name, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
			return nil
		}
		field.Name = p.currentToken.Literal
		if !p.expectToken(COLON) {
			return nil
		}

		// Parse the field type
		if !p.expectToken(INT32) && !p.expectToken(FLOAT32) && !p.expectToken(STRING_TYPE) {
			p.errors = append(p.errors, fmt.Sprintf("expected type, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
			return nil
		}

		// Save the type we just parsed
		baseType := p.currentToken.Type
		valueType := TokenTypeToValueKind(baseType)

		// Check if the next token is a left bracket (array type)
		if p.peekToken.Type == LBRACKET {
			// This is an array type
			field.Type = ValueArray
			elementType := valueType
			field.ArrayType = &elementType

			// Consume the brackets
			if !p.expectToken(LBRACKET) {
				p.errors = append(p.errors, fmt.Sprintf("expected [, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
				return nil
			}
			if !p.expectToken(RBRACKET) {
				p.errors = append(p.errors, fmt.Sprintf("expected ], got %v at line %d", p.currentToken.Type, p.currentToken.Line))
				return nil
			}
		} else {
			// Regular non-array type
			field.Type = valueType
		}

		structType.Fields = append(structType.Fields, field)
		p.nextToken()
	}
	if p.currentToken.Type != RBRACE {
		p.errors = append(p.errors, fmt.Sprintf("expected }, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
		return nil
	}
	p.nextToken()
	return structType
}

func (p *Parser) parseFunction() *ParsedFunction {
	function := &ParsedFunction{
		Labels:           make(map[string]int),
		ReturnStructName: "", // Initialize the new field
	}
	if !p.expectToken(IDENT) {
		return nil
	}
	function.Name = p.currentToken.Literal
	if !p.expectToken(LPAREN) {
		return nil
	}
	p.nextToken()
	for p.currentToken.Type != RPAREN && p.currentToken.Type != EOF {
		param := ParsedParam{}
		if p.currentToken.Type != IDENT {
			p.errors = append(p.errors, fmt.Sprintf("expected parameter name, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
			return nil
		}
		param.Name = p.currentToken.Literal
		if !p.expectToken(COLON) {
			p.errors = append(p.errors, fmt.Sprintf("expected :, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
			return nil
		}
		if !p.expectToken(INT32) && !p.expectToken(FLOAT32) {
			p.errors = append(p.errors, fmt.Sprintf("expected type, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
			return nil
		}
		param.Type = TokenTypeToValueKind(p.currentToken.Type)
		function.Params = append(function.Params, param)
		p.nextToken()
		if p.currentToken.Type == COMMA {
			p.nextToken()
		}
	}
	if !p.expectToken(ARROW) {
		return nil
	}

	// Handle struct return types
	// Modified return type parsing section
	if p.expectToken(INT32) || p.expectToken(FLOAT32) || p.expectToken(VOID) || p.expectToken(STRING_TYPE) {
		function.ReturnType = TokenTypeToValueKind(p.currentToken.Type)
	} else if p.expectToken(IDENT) {
		structName := p.currentToken.Literal
		function.ReturnType = ValueStruct
		function.ReturnStructName = structName
	} else {
		p.errors = append(p.errors, fmt.Sprintf("expected return type, got %v at line %d",
			p.currentToken.Type, p.currentToken.Line))
		return nil
	}

	if !p.expectToken(LBRACE) {
		p.errors = append(p.errors, fmt.Sprintf("expected {, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
		return nil
	}
	p.nextToken()
	instIndex := 0
	for p.currentToken.Type != RBRACE && p.currentToken.Type != EOF {
		if p.peekToken.Type == COLON {
			labelName := p.currentToken.Literal
			function.Labels[labelName] = instIndex
			p.nextToken()
			p.nextToken()
			continue
		}
		if instr := p.parseInstruction(); instr != nil {
			function.Body = append(function.Body, *instr)
			instIndex++
		} else {
			return nil
		}
	}
	if p.currentToken.Type != RBRACE {
		p.errors = append(p.errors, fmt.Sprintf("expected }, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
		return nil
	}
	p.nextToken()
	return function
}

func (p *Parser) parseInstruction() *Instruction {
	opcode, err := TokenTypeToOpcode(p.currentToken.Type)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("at line %d: %v", p.currentToken.Line, p.currentToken.Type))
		return nil
	}
	instr := &Instruction{
		Token:  p.currentToken,
		Opcode: opcode,
	}
	p.nextToken()
	switch opcode {
	case vm.PUSH:
		if p.currentToken.Type != INT32 && p.currentToken.Type != FLOAT32 {
			p.errors = append(p.errors, fmt.Sprintf("push requires operand type first, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
			p.nextToken()
			return nil
		}
		instr.Operands = append(instr.Operands, p.currentToken)
		p.nextToken()
		if p.currentToken.Type != INT && p.currentToken.Type != FLOAT {
			p.errors = append(p.errors, fmt.Sprintf("push requires value operand, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
			p.nextToken()
			return nil
		}
		instr.Operands = append(instr.Operands, p.currentToken)
		p.nextToken()
	case vm.STORE, vm.LOAD:
		if p.currentToken.Type != INT {
			p.errors = append(p.errors, fmt.Sprintf("store/load requires integer operand, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
			p.nextToken()
			return nil
		}
		instr.Operands = append(instr.Operands, p.currentToken)
		p.nextToken()
	case vm.JMP, vm.IJE, vm.IJNE, vm.FJNE, vm.FJE:
		if p.currentToken.Type != IDENT {
			p.errors = append(p.errors, fmt.Sprintf("jump requires label operand, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
			p.nextToken()
			return nil
		}
		instr.Operands = append(instr.Operands, p.currentToken)
		p.nextToken()
		if opcode != vm.JMP {
			if p.currentToken.Type != INT && p.currentToken.Type != FLOAT {
				p.errors = append(p.errors, fmt.Sprintf("conditional jump requires value operand, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
				return nil
			}
			instr.Operands = append(instr.Operands, p.currentToken)
			p.nextToken()
		}
	case vm.CALL:
		if p.currentToken.Type != IDENT {
			p.errors = append(p.errors, fmt.Sprintf("call requires function name, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
			p.nextToken()
			return nil
		}
		instr.Operands = append(instr.Operands, p.currentToken)
		p.nextToken()
	case vm.STRALLOC:
		if p.currentToken.Type != STRING {
			p.errors = append(p.errors, fmt.Sprintf("stralloc requires string literla, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
			p.nextToken()
			return nil
		}
		instr.Operands = append(instr.Operands, p.currentToken)
		p.nextToken()
	case vm.NEWSTRUCT:
		if p.currentToken.Type != IDENT {
			p.errors = append(p.errors, fmt.Sprintf("new struct requires struct name, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
			p.nextToken()
			return nil
		}
		instr.Operands = append(instr.Operands, p.currentToken)
		p.nextToken()
	case vm.STFIELD, vm.FLDGET:
		if p.currentToken.Type != STRING {
			p.errors = append(p.errors, fmt.Sprintf("field access requires field name, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
			p.nextToken()
			return nil
		}
		instr.Operands = append(instr.Operands, p.currentToken)
		p.nextToken()
	case vm.NEWARR:
		if p.currentToken.Type != FLOAT32 && p.currentToken.Type != INT32 {
			p.errors = append(p.errors, fmt.Sprintf("newarr requires type operand, got %v at line %d", p.currentToken.Type, p.currentToken.Line))
			p.nextToken()
			return nil
		}
		instr.Operands = append(instr.Operands, p.currentToken)
		p.nextToken()
	case vm.POP, vm.DUP, vm.IADD, vm.ISUB, vm.IMUL, vm.IDIV, vm.FADD, vm.FSUB, vm.FMUL, vm.FDIV, vm.FREE, vm.LOADH, vm.STOREH, vm.EQ, vm.NE, vm.LT, vm.LE, vm.GT, vm.GE, vm.LDELEM, vm.STELEM:
		return instr
	case vm.RET:
		return instr
	default:
		p.errors = append(p.errors, fmt.Sprintf("unkown instruction %v at line %d", opcode, p.currentToken.Line))
		p.nextToken()
	}
	return instr
}
