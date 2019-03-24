package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/pdk/gosh/compile"
	"github.com/pdk/gosh/lexer"
	"github.com/pdk/gosh/parse"
)

// Prompt is show when waiting for input.
var Prompt = ">>> "

// Analyze does an analysis of a parse result and prints the result.
func Analyze(ast *parse.Node) (*compile.Node, error) {

	bits := compile.NewAnalysis()

	ctree := compile.ConvertParseToCompile(ast)
	ctree.ScopeAnalysis(bits)

	return ctree, nil
}

// Start begins reading expressions. Stops when no more input.
func Start(in io.Reader, out, errout io.Writer) {

	topContext := compile.GlobalScope()

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

		vals, err := Evaluate("REPL", input, topContext)
		if err != nil {
			fmt.Fprintf(errout, "%s\n", err)
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

// Evaluate lexes, parses, analyzes, compiles, and then evaluates the input.
func Evaluate(inputName string, input []string, env *compile.Variables) ([]compile.Value, error) {

	l := lexer.New(inputName, input)
	p := parse.New(l)

	result, err := p.Parse()
	if err != nil {
		return compile.Values(), err
	}

	ast, err := Analyze(result)
	if err != nil {
		return compile.Values(), err
	}

	eval, err := ast.Evaluator()
	if err != nil {
		return compile.Values(), err
	}

	return eval(env)
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
