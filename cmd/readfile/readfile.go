package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pdk/gosh/lexer"
	"github.com/pdk/gosh/reader"
)

func main() {

	// result, err := reader.ReadFile(os.Args[1])
	// if err != nil {
	// 	log.Fatalf("%s", err)
	// }

	// for i, l := range result {
	// 	fmt.Printf("%3d. %s\n", i, string(l))
	// }

	result, err := reader.ReadLines(os.Args[1])
	if err != nil {
		log.Fatalf("%s", err)
	}

	for i, l := range result {
		fmt.Printf("%3d. %s\n", i+1, l)
	}

	lexed := lexer.New(result).Lexemes()
	for _, t := range lexed {
		fmt.Printf("%s\n", t.String())
	}
}
