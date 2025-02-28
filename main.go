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
	lexer := assembler.NewLexer(`.structs
		struct person {
			age: int32
		}
		.text
		func newperson(age: int32) -> person {
			store 0			
			newstruct person 
			dup 
			load 0		
			stfield "age"
			ret
		}
		func main() -> void {
			push int32 42
			call newperson
			fldget "age"
			ret
		}
		`)
	parser := assembler.NewParser(lexer)
	var prog *assembler.Program
	var err error
	if prog, err = parser.Parse(); err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
	generator := assembler.NewCodeGenerator(prog)
	bytecode, err := generator.Generate()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(bytecode)
	vm := vm.NewVm(bytecode)
	vm.Run()
}
