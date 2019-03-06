package repl

import (
    "bufio"
    "fmt"
    "io"
    "github.com/pdk/gosh/lexer"
    "github.com/pdk/gosh/token"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
    scanner := bufio.NewScanner(in)

    for {
        fmt.Printf(PROMPT)
        scanned := scanner.Scan()
        if !scanned {
            return
        }

        line := scanner.Text()
        l := lexer.New(line)

        for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
            fmt.Printf("%+v\n", tok)
        }
    }
}

Ball, Thorsten. Writing An Interpreter In Go (p. 34). Kindle Edition.
