package vm

import "fmt"

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
	ALLOC
	FREE
	LOADH
	STOREH
	DUP
	STRALLOC
	SYSCALL
	NEWARR
	LDELEM // load element from array
	STELEM // store element to array
	FUNC
	FUNC_NORMAL
	FUNC_MAIN
	DEFSTRUCT
	NEWSTRUCT
	FLDGET
	STFIELD
)

func (op Opcode) String() string {
	switch op {
	case HALT:
		return "HALT"
	case PUSH:
		return "PUSH"
	case POP:
		return "POP"
	case IADD:
		return "IADD"
	case ISUB:
		return "ISUB"
	case IMUL:
		return "IMUL"
	case IDIV:
		return "IDIV"
	case FADD:
		return "FADD"
	case FSUB:
		return "FSUB"
	case FMUL:
		return "FMUL"
	case FDIV:
		return "FDIV"
	case JMP:
		return "JMP"
	case IJNE:
		return "IJNE"
	case IJE:
		return "IJE"
	case FJNE:
		return "FJNE"
	case FJE:
		return "FJE"
	case EQ:
		return "EQ"
	case NE:
		return "NE"
	case LT:
		return "LT"
	case GT:
		return "GT"
	case GE:
		return "GE"
	case LE:
		return "LE"
	case LOAD:
		return "LOAD"
	case STORE:
		return "STORE"
	case CALL:
		return "CALL"
	case RET:
		return "RET"
	case RETV:
		return "RETV"
	case ALLOC:
		return "ALLOC"
	case FREE:
		return "FREE"
	case LOADH:
		return "LOADH"
	case STOREH:
		return "STOREH"
	case DUP:
		return "DUP"
	case STRALLOC:
		return "STRALLOC"
	case SYSCALL:
		return "SYSCALL"
	case NEWARR:
		return "NEWARR"
	case LDELEM:
		return "LDELEM"
	case STELEM:
		return "STELEM"
	case FUNC:
		return "FUNC"
	case FUNC_NORMAL:
		return "FUNC_NORMAL"
	case FUNC_MAIN:
		return "FUNC_MAIN"
	case DEFSTRUCT:
		return "DEFSTRUCT"
	case NEWSTRUCT:
		return "NEWSTRUCT"
	case FLDGET:
		return "FLDGET"
	case STFIELD:
		return "STFIELD"
	default:
		return fmt.Sprintf("UNKNOWN_OPCODE(%d)", op)
	}
}
