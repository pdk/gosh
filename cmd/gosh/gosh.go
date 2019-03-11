package main

import (
	"os"

	"github.com/pdk/gosh/repl"
)

func main() {
	repl.Start(os.Stdin, os.Stdout, os.Stderr)
}
