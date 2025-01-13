package vm

type Opcode byte

const (
	HALT Opcode = iota
	PUSH
	POP
	IADD
	ISUB
	IMUL
	IDIV
	FADD
	FSUB
	FMUL
	FDIV
	JMP
	IJNE
	IJE
	FJNE
	FJE
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
