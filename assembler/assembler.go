package assembler

// Assembler is the main struct that handles assembling source code to bytecode
type Assembler struct {
	source     string
	lexer      *Lexer
	parser     *Parser
	generator  *CodeGenerator
	program    *Program
	bytecode   []byte
	debugMode  bool
	outputFile string
}

// NewAssembler creates a new assembler for the given source code
func NewAssembler(source string) *Assembler {
	return &Assembler{
		source:    source,
		debugMode: false,
	}
}
