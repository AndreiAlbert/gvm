package vm

type Opcode byte

const (
	HALT Opcode = iota
	PUSH
	POP
	ADD
	SUB
	MUL
	DIV
	JMP
	JNE
	JE
	EQ
	NE
	LT
	GT
	GE
	LE
	LOAD
	STORE
	CALL
	RET
	RETV
)
