package lexer

import (
	"fmt"
	"os"
)

// PrintError prints an error message, and the offending line of input.
func (lex *Lexeme) PrintError(message string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, message+"\n", args...)
	fmt.Fprintf(os.Stderr, "%s\n", lex.lexer.input[lex.lineNumber-1])
	fmt.Fprintf(os.Stderr, "%*s\n", lex.charNumber, "^")
}
