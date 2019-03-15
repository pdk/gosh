package parse

import (
	"fmt"
	"strconv"
	"unicode"

	"github.com/pdk/gosh/lexer"
	"github.com/pdk/gosh/token"
)

// Arity designates if a node is Lefty/Righty
type Arity int

// Flavours of Arity
const (
	NotSpecified Arity = iota
	Lefty              // has an arg to the left
	Righty             // only has args (if any) to the right
)

// Node is the result of a parse.
type Node struct {
	lexeme   *lexer.Lexeme
	children []*Node
	arity    Arity
}

// Lexeme returns the lexeme of the node.
func (n *Node) Lexeme() *lexer.Lexeme {
	return n.lexeme
}

// Children returns the children of the node.
func (n *Node) Children() []*Node {
	return n.children
}

// Arity returns the arity of the node.
func (n *Node) Arity() Arity {
	return n.arity
}

// Print will print a parse tree with indentation.
func (n *Node) Print() {
	n.print(0)
}

// Literal returns the literal value of the lexeme.
func (n *Node) Literal() string {
	return n.lexeme.Literal()
}

// Token returns the token of the lexeme of the node.
func (n *Node) Token() token.Token {
	if n.lexeme == nil {
		return token.NADA
	}
	return n.lexeme.Token()
}

func (n *Node) firstChild() *Node {
	if len(n.children) == 0 {
		return nil
	}

	return n.children[0]
}

func (n *Node) print(depth int) {
	fmt.Printf("%s\n", n.lexeme.IndentString(3*depth))
	for _, c := range n.children {
		c.print(depth + 1)
	}
}

// Value returns the Lexeme of the parsed node.
func (n *Node) Value() *lexer.Lexeme {
	return n.lexeme
}

func containsQuotable(s string) bool {
	for _, c := range s {
		if c == '(' || c == ')' || unicode.IsSpace(c) {
			return true
		}
	}
	return false
}

func sexprQuote(s string) string {
	if containsQuotable(s) {
		return strconv.Quote(s)
	}
	return s
}

// Sexpr returns an s-expression of a parse, e.g. "1+2" => "(+ 1 2)".
func (n *Node) Sexpr() string {

	if len(n.children) == 0 {
		return sexprQuote(n.lexeme.Literal())
	}

	s := "("
	s += sexprQuote(n.lexeme.Literal())

	for _, c := range n.children {
		s += " " + c.Sexpr()
	}

	s += ")"

	return s
}
