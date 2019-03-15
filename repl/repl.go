package repl

import (
	"bufio"
	"fmt"
	"io"
	"log"

	"github.com/pdk/gosh/compile"
	"github.com/pdk/gosh/lexer"
	"github.com/pdk/gosh/parse"
)

// Prompt is show when waiting for input.
var Prompt = ">>> "

// PrintAnalysis does an analysis of a parse result and prints the result.
func PrintAnalysis(ast *parse.Node) {

	bits := compile.NewAnalysis()

	ctree := compile.ConvertParseToCompile(ast)
	ctree.ScopeAnalysis(bits)

	bits.Print()

	for _, f := range ctree.AllFuncs() {
		f.Analysis().Print()
	}
}

// Start begins reading expressions. Stops when no more input.
func Start(in io.Reader, out, errout io.Writer) {

	scanner := bufio.NewScanner(in)

	var input []string

	for {
		fmt.Fprintf(out, Prompt)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		nextLine := scanner.Text()
		if nextLine != "." {
			input = append(input, nextLine)
			continue
		}

		l := lexer.New(input)
		// l.LogDump()
		p := parse.New(l)

		result, err := p.Parse()
		if err != nil {
			_, err = fmt.Fprintf(errout, "%s\n", err)
			if err != nil {
				log.Fatalf("%s", err)
			}
		}

		_, err = fmt.Fprintf(out, "%s\n", result.Sexpr())
		if err != nil {
			log.Fatalf("%s", err)
		}

		PrintAnalysis(result)

		input = []string{}
	}
}
