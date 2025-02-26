package assembler

import (
	. "stack_vm/common"
	"strings"
	"testing"
)

func TestParseStructDefinition(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid struct",
			input: `.structs
                    struct Point {
                        x: int32
                        y: int32
                    }`,
			wantErr: false,
		},
		{
			name: "missing field type",
			input: `.structs
                    struct Point {
                        x:
                    }`,
			wantErr: true,
			errMsg:  "expected type",
		},
		{
			name: "invalid field type",
			input: `.structs
                    struct Point {
                        x: string
                    }`,
			wantErr: true,
			errMsg:  "expected type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			p := NewParser(l)
			program, err := p.Parse()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got no error", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(program.Structs) != 1 {
					t.Errorf("expected 1 struct, got %d", len(program.Structs))
				}
			}
		})
	}
}

func TestParseFunctionDefinition(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid function",
			input: `.text
                    func add(a: int32, b: int32) -> int32 {
                        push int32 42
                        ret
                    }`,
			wantErr: false,
		},
		{
			name: "missing parameter type",
			input: `.text
                    func add(a, b: int32) -> int32 {
                        ret
                    }`,
			wantErr: true,
			errMsg:  "parser encountered the following errors:\n  1. expeted :, got IDENT at lien 2\n",
		},
		{
			name: "invalid return type",
			input: `.text
                    func add(a: int32) -> string {
                        ret
                    }`,
			wantErr: true,
			errMsg:  "expected return type",
		},
		{
			name: "missing function body",
			input: `.text
                    func add(a: int32) -> int32`,
			wantErr: true,
			errMsg:  "expected {",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			p := NewParser(l)
			program, err := p.Parse()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got no error", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(program.Functions) != 1 {
					t.Errorf("expected 1 function, got %d", len(program.Functions))
				}
			}
		})
	}
}

func TestParseInstructions(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid push int",
			input: `.text
                    func test() -> void {
                        push int32 42
                        ret
                    }`,
			wantErr: false,
		},
		{
			name: "invalid push without type",
			input: `.text
                    func test() -> void {
                        push 42
                        ret
                    }`,
			wantErr: true,
			errMsg:  "push requires operand type first",
		},
		{
			name: "valid store",
			input: `.text
                    func test() -> void {
                        store 0
                    }`,
			wantErr: false,
		},
		{
			name: "invalid store operand",
			input: `.text
                    func test() -> void {
                        store x
                    }`,
			wantErr: true,
			errMsg:  "store/load requires integer operand",
		},
		{
			name: "valid jump",
			input: `.text
                    func test() -> void {
                        label1:
                        jmp label1
                    }`,
			wantErr: false,
		},
		{
			name: "invalid conditional jump",
			input: `.text
                    func test() -> void {
                        ije label1 x
                    }`,
			wantErr: true,
			errMsg:  "conditional jump requires value operand",
		},
		{
			name: "valid function call",
			input: `.text
                    func test() -> void {
                        call other_func
                    }`,
			wantErr: false,
		},
		{
			name: "valid string allocation",
			input: `.text
                    func test() -> void {
                        stralloc "hello"
                    }`,
			wantErr: false,
		},
		{
			name: "valid array creation",
			input: `.text
                    func test() -> void {
                        newarr int32
                    }`,
			wantErr: false,
		},
		{
			name: "valid field access",
			input: `.text
                    func test() -> void {
                        fldget "x"
                    }`,
			wantErr: false,
		},
		{
			name: "valid arithmetic",
			input: `.text
                    func test() -> void {
                        iadd
                        isub
                        imul
                        idiv
                    }`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			p := NewParser(l)
			program, err := p.Parse()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got no error", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(program.Functions) != 1 {
					t.Errorf("expected 1 function, got %d", len(program.Functions))
				}
			}
		})
	}
}

func TestCompleteProgram(t *testing.T) {
	input := `.structs
        struct Point {
            x: int32
            y: int32
        }

        struct Rectangle {
            top_left: int32
            width: int32
            height: int32
        }

    .text
        func calculate_area(rect: int32) -> int32 {
            ; Load width and height
            push int32 10
            push int32 20
            ; Multiply them
            imul
            ret
        }

        func main() -> void {
            ; Create rectangle
            push int32 5
            store 0
            
            ; Calculate area
            load 0
            call calculate_area
            
            ; Compare result
            push int32 200
            eq
            
            ; Jump based on comparison
            label1:
            ije label1 1
            
			ret
        }`

	l := NewLexer(input)
	p := NewParser(l)
	program, err := p.Parse()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(program.Structs) != 2 {
		t.Errorf("expected 2 structs, got %d", len(program.Structs))
	}

	if len(program.Functions) != 2 {
		t.Errorf("expected 2 functions, got %d", len(program.Functions))
	}

	var mainFunc *ParsedFunction
	for _, f := range program.Functions {
		if f.Name == "main" {
			mainFunc = &f
			break
		}
	}

	if mainFunc == nil {
		t.Fatal("main function not found")
	}

	if mainFunc.ReturnType != ValueVoid {
		t.Errorf("expected main return type void, got %v", mainFunc.ReturnType)
	}

	if len(mainFunc.Labels) != 1 {
		t.Errorf("expected 1 label in main, got %d", len(mainFunc.Labels))
	}

	if _, exists := mainFunc.Labels["label1"]; !exists {
		t.Error("label1 not found in main function")
	}
}
