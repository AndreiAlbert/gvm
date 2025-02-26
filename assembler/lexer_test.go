package assembler

import "testing"

func TestNextToken(t *testing.T) {
	input := `.structs
	struct Point {
		x: int32
		y: float32
	}

.text
	; This is a comment
	func main() -> void {
		push int32 42
		push float32 3.14
		iadd
		dup
		pop
		halt
	}

	func calc(a: int32, b: float32) -> int32 {
		; Test all arithmetic operations
		iadd
		isub
		imul
		idiv
		fadd
		fsub
		fmul
		fdiv

		; Test memory operations
		store 0
		load 1
		alloc
		free
		loadh
		storeh

		; Test comparisons
		eq
		ne
		lt
		le
		gt
		ge

		; Test control flow
		jmp label1
		ije label2 42
		ijne label3 10
		fje label4 3.14
		fjne label5 2.5
		call func1
		ret

		; Test arrays
		newarr int32
		ldelem
		stelem

		; Test strings
		stralloc "hello world"

		; Test structs
		newstruct Point
		fldget "x"
		stfield "y"

		syscall
	}`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{SECTION_STRUCTS, ".structs"},
		{STRUCT, "struct"},
		{IDENT, "Point"},
		{LBRACE, "{"},
		{IDENT, "x"},
		{COLON, ":"},
		{INT32, "int32"},
		{IDENT, "y"},
		{COLON, ":"},
		{FLOAT32, "float32"},
		{RBRACE, "}"},

		{SECTION_TEXT, ".text"},
		{FUNC, "func"},
		{IDENT, "main"},
		{LPAREN, "("},
		{RPAREN, ")"},
		{ARROW, "->"},
		{VOID, "void"},
		{LBRACE, "{"},
		{PUSH, "push"},
		{INT32, "int32"},
		{INT, "42"},
		{PUSH, "push"},
		{FLOAT32, "float32"},
		{FLOAT, "3.14"},
		{IADD, "iadd"},
		{DUP, "dup"},
		{POP, "pop"},
		{HALT, "halt"},
		{RBRACE, "}"},

		{FUNC, "func"},
		{IDENT, "calc"},
		{LPAREN, "("},
		{IDENT, "a"},
		{COLON, ":"},
		{INT32, "int32"},
		{COMMA, ","},
		{IDENT, "b"},
		{COLON, ":"},
		{FLOAT32, "float32"},
		{RPAREN, ")"},
		{ARROW, "->"},
		{INT32, "int32"},
		{LBRACE, "{"},

		{IADD, "iadd"},
		{ISUB, "isub"},
		{IMUL, "imul"},
		{IDIV, "idiv"},
		{FADD, "fadd"},
		{FSUB, "fsub"},
		{FMUL, "fmul"},
		{FDIV, "fdiv"},

		{STORE, "store"},
		{INT, "0"},
		{LOAD, "load"},
		{INT, "1"},
		{ALLOC, "alloc"},
		{FREE, "free"},
		{LOADH, "loadh"},
		{STOREH, "storeh"},

		{EQ, "eq"},
		{NE, "ne"},
		{LT, "lt"},
		{LE, "le"},
		{GT, "gt"},
		{GE, "ge"},

		{JMP, "jmp"},
		{IDENT, "label1"},
		{IJE, "ije"},
		{IDENT, "label2"},
		{INT, "42"},
		{IJNE, "ijne"},
		{IDENT, "label3"},
		{INT, "10"},
		{FJE, "fje"},
		{IDENT, "label4"},
		{FLOAT, "3.14"},
		{FJNE, "fjne"},
		{IDENT, "label5"},
		{FLOAT, "2.5"},
		{CALL, "call"},
		{IDENT, "func1"},
		{RET, "ret"},

		{NEWARR, "newarr"},
		{INT32, "int32"},
		{LDELEM, "ldelem"},
		{STELEM, "stelem"},

		{STRALLOC, "stralloc"},
		{STRING, "hello world"},

		{NEWSTRUCT, "newstruct"},
		{IDENT, "Point"},
		{FLDGET, "fldget"},
		{STRING, "x"},
		{STFIELD, "stfield"},
		{STRING, "y"},

		{SYSCALL, "syscall"},
		{RBRACE, "}"},
		{EOF, ""},
	}

	l := NewLexer(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Errorf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Errorf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestInvalidNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"42.42.42", "Invalid number format"},
		{"3.14.15", "Invalid number format"},
		{"999999999999999", "Invalid integer format"},
	}

	for i, tt := range tests {
		l := NewLexer(tt.input)
		tok := l.NextToken()

		if tok.Type != ILLEGAL {
			t.Errorf("tests[%d] - expected illegal token, got=%q", i, tok.Type)
		}

		if tok.Literal != tt.expected {
			t.Errorf("tests[%d] - expected error message=%q, got=%q",
				i, tt.expected, tok.Literal)
		}
	}
}

func TestComments(t *testing.T) {
	input := `push int32 42 ; this is a comment
; this is a full line comment
pop ; another comment`

	expected := []TokenType{
		PUSH,
		INT32,
		INT,
		POP,
		EOF,
	}

	l := NewLexer(input)

	for i, expectedType := range expected {
		tok := l.NextToken()
		if tok.Type != expectedType {
			t.Errorf("tests[%d] - expected=%q, got=%q",
				i, expectedType, tok.Type)
		}
	}
}

func TestStringLiterals(t *testing.T) {
	input := `stralloc "hello world"
stralloc "test with spaces"
stralloc "special chars: \n \t \\"`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{STRALLOC, "stralloc"},
		{STRING, "hello world"},
		{STRALLOC, "stralloc"},
		{STRING, "test with spaces"},
		{STRALLOC, "stralloc"},
		{STRING, "special chars: \n \t \\"},
		{EOF, ""},
	}

	l := NewLexer(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Errorf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Errorf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
