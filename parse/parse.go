package parse

import (
	"fmt"
	"log"
	"strconv"
	"unicode"

	"github.com/pdk/gosh/lexer"
	"github.com/pdk/gosh/token"
)

// Parser processes a list of lexed nodes.
type Parser struct {
	lexer *lexer.Lexer
}

// New makes a new Parser given a Lexer.
func New(lexer *lexer.Lexer) *Parser {
	return &Parser{
		lexer: lexer,
	}
}

// Parse returns the result of parsing the input.
func (p *Parser) Parse() *Node {
	ast := p.expression(0)
	ast = ast.applyTransforms()

	return ast
}

// Node is the result of a parse.
type Node struct {
	lexeme   *lexer.Lexeme
	children []*Node
}

// Print will print a parse tree with indentation.
func (n *Node) Print() {
	n.print(0)
}

// Token returns the token of the lexeme of the node.
func (n *Node) Token() token.Token {
	return n.lexeme.Token()
}

func (n *Node) applyTransforms() *Node {

	var newChildren []*Node

	if n.Token() == token.LPAREN {

		first := n.firstChild()

		// transform method invocation
		// ("(" (. o m) ...) ==> (m-apply o m ...)
		if first != nil && first.Token() == token.PERIOD && len(first.children) == 2 {

			l := n.lexeme.WithToken(token.METHAPPLY).WithLiteral("m-apply")
			n.lexeme = &l

			newChildren = append(newChildren, n.children[0].children[0])
			newChildren = append(newChildren, n.children[0].children[1])
			n.children = n.children[1:]

		} else {

			// transform function invocation
			// ("(" f ...) ==> (f-apply f ...)
			l := n.lexeme.WithToken(token.FUNCAPPLY).WithLiteral("f-apply")
			n.lexeme = &l
		}
	}

	for _, x := range n.children {
		newChildren = append(newChildren, x.applyTransforms())
	}

	n.children = newChildren

	return n
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

// Children returns the children of a parsed node.
func (n *Node) Children() []*Node {
	return n.children
}

func containsWhitespace(s string) bool {
	for _, c := range s {
		if unicode.IsSpace(c) {
			return true
		}
	}
	return false
}

func sexprQuote(s string) string {
	if containsWhitespace(s) || s == "(" || s == ")" {
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

type nudFunc func(*Node, *Parser) *Node

type ledFunc func(*Node, *Parser, *Node) *Node

type stdFunc func(*Node, *Parser) *Node

// tdopEntry contains the information required to drive the parser.
type tdopEntry struct {
	bindingPower int
	nud          nudFunc
	led          ledFunc
	std          stdFunc
}

var tdopRegistry [token.KeywordEnd]tdopEntry

// Ordered precedence values. (Even nubmers cuz we need -1 on right-associative
// infix operators.)
const (
	P_EOF = iota * 2
	P_UNEXPECTED
	P_SELF

	P_PIPE
	P_ASSIGN
	P_LOGIC
	P_RETURN
	P_COMMA
	P_COMPARE
	P_PLUSMINUS
	P_MULTDIV
	P_PREFIX
	P_FUNC
	P_PERIOD
)

func init() {
	tdopRegistry[token.ILLEGAL] = tdopEntry{}

	tdopRegistry[token.PERIOD] = infix(P_PERIOD)

	tdopRegistry[token.NOT] = prefix(P_PREFIX)

	tdopRegistry[token.MULT] = infix(P_MULTDIV)
	tdopRegistry[token.DIV] = infix(P_MULTDIV)
	tdopRegistry[token.MODULO] = infix(P_MULTDIV)

	tdopRegistry[token.PLUS] = infix(P_PLUSMINUS)
	tdopRegistry[token.MINUS] = prefixInfix(P_PREFIX, P_PLUSMINUS)

	tdopRegistry[token.EQUAL] = infix(P_COMPARE)
	tdopRegistry[token.LESS] = infix(P_COMPARE)
	tdopRegistry[token.GRTR] = infix(P_COMPARE)
	tdopRegistry[token.NOT_EQUAL] = infix(P_COMPARE)
	tdopRegistry[token.LESS_EQUAL] = infix(P_COMPARE)
	tdopRegistry[token.GRTR_EQUAL] = infix(P_COMPARE)
	tdopRegistry[token.ISA] = infix(P_COMPARE)
	tdopRegistry[token.HASA] = infix(P_COMPARE)

	tdopRegistry[token.COMMA] = infix(P_COMMA)

	tdopRegistry[token.RETURN] = prefix(P_RETURN)

	tdopRegistry[token.LOG_AND] = infix(P_LOGIC)
	tdopRegistry[token.LOG_OR] = infix(P_LOGIC)

	tdopRegistry[token.BREAK] = self()
	tdopRegistry[token.CONTINUE] = self()

	tdopRegistry[token.ASSIGN] = rinfix(P_ASSIGN)
	tdopRegistry[token.ACCUM] = rinfix(P_ASSIGN)
	tdopRegistry[token.QASSIGN] = rinfix(P_ASSIGN)

	tdopRegistry[token.LPIPE] = infix(P_PIPE)
	tdopRegistry[token.RPIPE] = rinfix(P_PIPE)

	tdopRegistry[token.EOF] = consumable()
	tdopRegistry[token.COMMENT] = consumable()

	tdopRegistry[token.IDENT] = self()
	tdopRegistry[token.INT] = self()
	tdopRegistry[token.FLOAT] = self()
	tdopRegistry[token.CHAR] = self()
	tdopRegistry[token.STRING] = self()
	tdopRegistry[token.NIL] = self()
	tdopRegistry[token.TRUE] = self()
	tdopRegistry[token.FALSE] = self()

	tdopRegistry[token.SEMI] = consumable()

	tdopRegistry[token.LPAREN] = leftParen()
	tdopRegistry[token.LSQR] = tdopEntry{}
	tdopRegistry[token.LBRACE] = tdopEntry{}

	tdopRegistry[token.RPAREN] = unbalanced()
	// tdopRegistry[token.RSQR] = tdopEntry{}
	// tdopRegistry[token.RBRACE] = tdopEntry{}

	tdopRegistry[token.COLON] = tdopEntry{}
	tdopRegistry[token.DOLLAR] = tdopEntry{}
	tdopRegistry[token.DDOLLAR] = tdopEntry{}

	tdopRegistry[token.ELSE] = tdopEntry{}
	tdopRegistry[token.FOR] = tdopEntry{}
	tdopRegistry[token.IN] = tdopEntry{}
	tdopRegistry[token.FUNC] = tdopEntry{}
	tdopRegistry[token.IF] = tdopEntry{}
	tdopRegistry[token.IMPORT] = tdopEntry{}
	tdopRegistry[token.PKG] = tdopEntry{}
	tdopRegistry[token.STRUCT] = tdopEntry{}
	tdopRegistry[token.SWITCH] = tdopEntry{}
	tdopRegistry[token.WHILE] = tdopEntry{}
	tdopRegistry[token.ENUM] = tdopEntry{}
	tdopRegistry[token.SYS] = tdopEntry{}
}

func consumable() tdopEntry {
	return tdopEntry{
		bindingPower: 0,
	}
}

func newNode(lex *lexer.Lexeme) *Node {
	return &Node{
		lexeme: lex,
	}
}

func bindPowerOf(lex *lexer.Lexeme) int {
	return tdopRegistry[lex.Token()].bindingPower
}

func leftParen() tdopEntry {
	return tdopEntry{
		bindingPower: P_FUNC,
		nud: func(node *Node, p *Parser) *Node {
			n := p.expression(P_UNEXPECTED)
			p.advance(token.RPAREN)
			return n
		},
		led: func(node *Node, p *Parser, left *Node) *Node {
			node.children = append(node.children, left)
			if p.lexer.Peek().Token() != token.RPAREN {
				node.children = append(node.children, p.expression(P_COMMA))
				for p.lexer.Peek().Token() == token.COMMA {
					p.advance(token.COMMA)
					node.children = append(node.children, p.expression(P_COMMA))
				}
			}
			p.advance(token.RPAREN)
			return node
		},
	}
}

func unbalanced() tdopEntry {
	return tdopEntry{
		bindingPower: P_UNEXPECTED,
	}
}

func prefix(rightBP int) tdopEntry {
	return tdopEntry{
		bindingPower: 0,
		nud: func(node *Node, p *Parser) *Node {
			node.children = append(node.children, p.expression(rightBP))
			return node
		},
	}
}

func infix(bindingPower int) tdopEntry {
	return tdopEntry{
		bindingPower: bindingPower,
		led: func(node *Node, p *Parser, left *Node) *Node {
			node.children = append(node.children, left)
			node.children = append(node.children, p.expression(bindingPower))
			return node
		},
	}
}

func prefixInfix(prefixBP, infixBP int) tdopEntry {
	return tdopEntry{
		bindingPower: infixBP,
		nud: func(node *Node, p *Parser) *Node {
			node.children = append(node.children, p.expression(prefixBP))
			return node
		},
		led: func(node *Node, p *Parser, left *Node) *Node {
			node.children = append(node.children, left)
			node.children = append(node.children, p.expression(infixBP))
			return node
		},
	}
}

func rinfix(bindingPower int) tdopEntry {
	return tdopEntry{
		bindingPower: bindingPower,
		led: func(node *Node, p *Parser, left *Node) *Node {
			node.children = append(node.children, left)
			node.children = append(node.children, p.expression(bindingPower-1))
			return node
		},
	}
}

func self() tdopEntry {
	return tdopEntry{
		bindingPower: P_SELF,
		nud: func(node *Node, p *Parser) *Node {
			return node
		},
	}
}

func (n *Node) nud() nudFunc {
	return tdopRegistry[n.Token()].nud
}

func (n *Node) led() ledFunc {
	return tdopRegistry[n.Token()].led
}

func (n *Node) std() stdFunc {
	return tdopRegistry[n.Token()].std
}

func (n *Node) bind() int {
	return tdopRegistry[n.Token()].bindingPower
}

func (p *Parser) expression(rbp int) *Node {

	var left *Node
	node := newNode(p.lexer.Next())

	if node.nud() != nil {
		left = node.nud()(node, p)
	} else {
		log.Fatalf("parser error: unexpected left token %s %s at %d,%d",
			node.Token().String(),
			strconv.Quote(node.lexeme.Literal()),
			node.lexeme.LineNo(), node.lexeme.CharNo())
	}

	for rbp < bindPowerOf(p.lexer.Peek()) {
		node := newNode(p.lexer.Next())
		if node.led() != nil {
			left = node.led()(node, p, left)
		} else {
			log.Fatalf("parser error: unexpected right token %s %s at %d,%d",
				node.Token().String(),
				strconv.Quote(node.lexeme.Literal()),
				node.lexeme.LineNo(), node.lexeme.CharNo())
		}
	}

	return left
}

func (p *Parser) advance(match token.Token) *Node {

	next := p.lexer.Next()
	if next.Token() != match {
		log.Fatalf("failed to process (3): expecting %s", match.String())
	}

	return newNode(next)
}
