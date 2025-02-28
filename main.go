package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"

	"stack_vm/assembler"
	"stack_vm/vm"
)

func to32bits(f float32) [4]byte {
	var buffer [4]byte
	binary.BigEndian.PutUint32(buffer[:], math.Float32bits(f))
	return buffer
}

func runFile(filename string) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}
	content, err := os.ReadFile(absPath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	fmt.Printf("Content: %s", content)
	lexer := assembler.NewLexer(string(content))
	parser := assembler.NewParser(lexer)
	prog, err := parser.Parse()
	if err != nil {
		log.Fatalf("Failed to parse progrmam: %v", err)
	}
	generator := assembler.NewCodeGenerator(prog)
	bytecode, err := generator.Generate()
	fmt.Println(bytecode)
	if err != nil {
		log.Fatalf("Failed to generate bytecode: %v", err)
	}
	vm := vm.NewVm(bytecode)
	vm.Run()
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Not enough arguments")
	}
	runFile(os.Args[1])
}
