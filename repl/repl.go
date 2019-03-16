package repl

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/pdk/gosh/compile"
	"github.com/pdk/gosh/lexer"
	"github.com/pdk/gosh/parse"
)

// Prompt is show when waiting for input.
var Prompt = ">>> "

// Analyze does an analysis of a parse result and prints the result.
func Analyze(ast *parse.Node) (*compile.Node, *compile.Analysis) {

	bits := compile.NewAnalysis()

	ctree := compile.ConvertParseToCompile(ast)
	ctree.ScopeAnalysis(bits)

	return ctree, bits
}

// Start begins reading expressions. Stops when no more input.
func Start(in io.Reader, out, errout io.Writer) {

	scanner := bufio.NewScanner(in)

	var input []string

	var parenCount, bracketCount int

	for {
		fmt.Fprintf(out, Prompt)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		nextLine := scanner.Text()

		if nextLine != "." {
			input = append(input, nextLine)
			parenCount, bracketCount = countBrackets(nextLine, parenCount, bracketCount)
		}

		if nextLine != "." && parenCount+bracketCount > 0 {
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

		ast, _ := Analyze(result)

		// ast, bits := Analyze(result)

		// bits.Print()
		// for _, f := range ast.AllFuncs() {
		// 	f.Analysis().Print()
		// }

		eval := ast.Evaluator()

		vals, err := eval()

		if err != nil {
			log.Printf("Error: %s", err)
		}

		if len(vals) > 0 {
			printable := []string{}
			for _, v := range vals {
				printable = append(printable, v.String())
			}
			fmt.Printf("%s\n", strings.Join(printable, ", "))
		}

		input = []string{}
		parenCount, bracketCount = 0, 0
	}
}

func countBrackets(line string, parenCount, bracketCount int) (int, int) {
	for _, c := range line {
		switch c {
		case '(':
			parenCount++
		case ')':
			parenCount--
		case '{':
			bracketCount++
		case '}':
			bracketCount--
		}
	}

	return parenCount, bracketCount
}
