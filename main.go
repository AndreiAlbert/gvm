package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"

	"stack_vm/assembler"
	// "stack_vm/common"
	// "stack_vm/vm"
)

func to32bits(f float32) [4]byte {
	var buffer [4]byte
	binary.BigEndian.PutUint32(buffer[:], math.Float32bits(f))
	return buffer
}

func main() {
	lexer := assembler.NewLexer(`.structs 
		struct Point {
			x: int32 
			y: int32
		}
		.text
		func pula(a: int32, b:int32) -> int32 {
			label1:
			push  420	
			ret
		}	
		`)
	parser := assembler.NewParser(lexer)
	prog, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v", prog)
}
