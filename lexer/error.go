package lexer

import (
	"errors"
	"fmt"
)

// PrintError prints an error message, and the offending line of input.
// func (lex *Lexeme) PrintError(message string, args ...interface{}) {
// 	fmt.Fprintf(os.Stderr, message+"\n", args...)
// 	fmt.Fprintf(os.Stderr, "%s\n", lex.lexer.input[lex.lineNumber-1])
// 	fmt.Fprintf(os.Stderr, "%*s\n", lex.charNumber, "^")
// }

func (lex *Lexeme) Error(message string, args ...interface{}) error {

	message = fmt.Sprintf(message, args...)

	offender := lex.lexer.input[lex.lineNumber-1]
	// callout := fmt.Sprintf("%s【%s】%s",
	// 	offender[0:lex.charNumber-1],
	// 	lex.literal,
	// 	offender[lex.charNumber+len(lex.literal)-1:])

	message = fmt.Sprintf("%s:%d:%d: %s: %s",
		lex.lexer.inputName,
		lex.lineNumber, lex.charNumber,
		offender,
		message)

	return errors.New(message)
}
