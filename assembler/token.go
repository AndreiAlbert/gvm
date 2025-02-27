package assembler

import "fmt"

type TokenType int

const (
	ILLEGAL TokenType = iota
	EOF

	// Single-character tokens
	COLON     // :
	SEMICOLON // ;
	LPAREN    // (
	RPAREN    // )
	LBRACE    // {
	RBRACE    // }
	COMMA     // ,
	ARROW     // ->
	LBRACKET  // [
	RBRACKET  // ]

	//Keywords
	FUNC
	STRUCT
	INT32
	FLOAT32
	STRING_TYPE
	VOID
	RETURN

	// System instructions
	HALT
	SYSCALL

	// Stack and memory instructions
	PUSH
	POP
	DUP
	STORE
	LOAD
	ALLOC
	FREE
	LOADH
	STOREH

	// Arithmetic instructions
	IADD
	ISUB
	IMUL
	IDIV
	FADD
	FSUB
	FMUL
	FDIV

	// Comparison instructions
	EQ
	NE
	LT
	LE
	GT
	GE

	// Control flow instructions
	JMP
	IJE
	IJNE
	FJE
	FJNE
	CALL
	RET

	// Array instructions
	NEWARR
	LDELEM
	STELEM

	// String instructions
	STRALLOC

	// Struct instructions
	NEWSTRUCT
	FLDGET
	STFIELD

	// Identifiers and literals
	IDENT  // variables, labels
	INT    // 123
	FLOAT  // 123.45
	STRING // "abc"

	// Sections
	SECTION_TEXT
	SECTION_STRUCTS // Only need text and structs sections
)

type Token struct {
	Type    TokenType
	Literal string
	Line    uint
	Column  uint
}

// creates new token
func newToken(tokType TokenType, literal string, line uint, column uint) Token {
	return Token{Type: tokType, Literal: literal, Line: line, Column: column}
}

var keywords = map[string]TokenType{
	"func":     FUNC,
	"struct":   STRUCT,
	"int32":    INT32,
	"float32":  FLOAT32,
	"void":     VOID,
	"return":   RETURN,
	".text":    SECTION_TEXT,
	".structs": SECTION_STRUCTS,
	"string":   STRING_TYPE,
}

var instructions = map[string]TokenType{
	// System
	"halt":    HALT,
	"syscall": SYSCALL,

	// Stack operations
	"push": PUSH,
	"pop":  POP,
	"dup":  DUP,

	// Memory operations
	"store":  STORE,
	"load":   LOAD,
	"alloc":  ALLOC,
	"free":   FREE,
	"loadh":  LOADH,
	"storeh": STOREH,

	// Integer arithmetic
	"iadd": IADD,
	"isub": ISUB,
	"imul": IMUL,
	"idiv": IDIV,

	// Float arithmetic
	"fadd": FADD,
	"fsub": FSUB,
	"fmul": FMUL,
	"fdiv": FDIV,

	// Comparisons
	"eq": EQ,
	"ne": NE,
	"lt": LT,
	"le": LE,
	"gt": GT,
	"ge": GE,

	// Control flow
	"jmp":  JMP,
	"ije":  IJE,
	"ijne": IJNE,
	"fje":  FJE,
	"fjne": FJNE,
	"call": CALL,
	"ret":  RET,

	// Arrays
	"newarr": NEWARR,
	"ldelem": LDELEM,
	"stelem": STELEM,

	// Strings
	"stralloc": STRALLOC,

	// Structs
	"newstruct": NEWSTRUCT,
	"fldget":    FLDGET,
	"stfield":   STFIELD,
}

func (t TokenType) String() string {
	var reverseInstructions map[TokenType]string
	reverseInstructions = make(map[TokenType]string)
	for name, tokType := range instructions {
		reverseInstructions[tokType] = name
	}
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case IDENT:
		return "IDENT"
	case INT:
		return "INT"
	case FLOAT:
		return "FLOAT"
	case STRING:
		return "STRING"
	case COLON:
		return "COLON"
	case SEMICOLON:
		return "SEMICOLON"
	case LPAREN:
		return "LPAREN"
	case RPAREN:
		return "RPAREN"
	case LBRACE:
		return "LBRACE"
	case RBRACE:
		return "RBRACE"
	case COMMA:
		return "COMMA"
	case ARROW:
		return "ARROW"
	case FUNC:
		return "FUNC"
	case STRUCT:
		return "STRUCT"
	case INT32:
		return "INT32"
	case FLOAT32:
		return "FLOAT32"
	case VOID:
		return "VOID"
	case RETURN:
		return "RETURN"
	case SECTION_TEXT:
		return "SECTION_TEXT"
	case SECTION_STRUCTS:
		return "SECTION_STRUCTS"
	case STRING_TYPE:
		return "STRING_TYPE"
	case LBRACKET:
		return "LBRACKET"
	case RBRACKET:
		return "RBRACKET"
	default:
		if instr, exists := reverseInstructions[t]; exists {
			return instr
		}
		return fmt.Sprintf("TokenType(%d)", t)
	}
}

func (t Token) String() string {
	return fmt.Sprintf("{Type: %v, Literal: %q, Line: %d, Column: %d}",
		t.Type, t.Literal, t.Line, t.Column)
}
