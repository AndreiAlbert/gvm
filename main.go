package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"

	"stack_vm/assembler"
	"stack_vm/vm"
)

func to32bits(f float32) [4]byte {
	var buffer [4]byte
	binary.BigEndian.PutUint32(buffer[:], math.Float32bits(f))
	return buffer
}

func main() {
	lexer := assembler.NewLexer(` .structs
		struct person {
			age: int32
			name: string
		}
		.text
		func main() -> void {
			newstruct person
			dup	
			push int32 42
			stfield "age"
			fldget "age"	
		}
		`)
	parser := assembler.NewParser(lexer)
	var prog *assembler.Program
	var err error
	if prog, err = parser.Parse(); err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
	fmt.Printf("%+v\n", prog)
	generator := assembler.NewCodeGenerator(prog)
	bytecode, err := generator.Generate()
	if err != nil {
		log.Fatal(err)
	}
	vm := vm.NewVm(bytecode)
	vm.Run()
}
