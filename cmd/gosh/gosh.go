package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/pdk/gosh/compile"
	"github.com/pdk/gosh/reader"
	"github.com/pdk/gosh/repl"
)

func main() {

	if len(os.Args) > 1 {
		inputName := os.Args[1]
		input, err := reader.ReadLines(inputName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
		}

		execute(inputName, input)
		return
	}

	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		input := reader.ReadLinesToStrings(os.Stdin)
		execute("stdin", input)
		return
	}

	fmt.Println("gosh 0.0.x")
	repl.Start(os.Stdin, os.Stdout, os.Stderr)
}

func execute(inputName string, input []string) {

	topContext := compile.GlobalScope()

	vals, err := repl.Evaluate(inputName, input, topContext)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}

	if len(vals) > 0 {
		printable := []string{}
		for _, v := range vals {
			printable = append(printable, v.String())
		}
		fmt.Printf("%s\n", strings.Join(printable, ", "))
	}
}
