package main

import (
	"log"
	"os"

	"github.com/pdk/gosh/lexer"
	"github.com/pdk/gosh/parse"
	"github.com/pdk/gosh/reader"
)

func main() {

	result, err := reader.ReadLines(os.Args[1])
	if err != nil {
		log.Fatalf("%s", err)
	}

	// for i, l := range result {
	// 	fmt.Printf("%3d. %s\n", i+1, l)
	// }

	lex := lexer.New(result)

	// for t := lex.Next(); t != nil; t = lex.Next() {
	// 	fmt.Printf("%s\n", t.String())
	// }

	// lex = lexer.New(result)
	ast := parse.New(lex).Parse()
	ast.Print()
}
