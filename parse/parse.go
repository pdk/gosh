package parse

import (
	"fmt"
	"log"

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
	return p.expression(0)
}

// Print will print a parse tree with indentation.
func (n *Node) Print() {
	n.print(0)
}

func (n *Node) print(depth int) {
	fmt.Printf("%s\n", n.lexeme.IndentString(3*depth))
	for _, c := range n.children {
		c.print(depth + 1)
	}
}

// Node is the result of a parse.
type Node struct {
	lexeme   *lexer.Lexeme
	children []*Node
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

func init() {
	tdopRegistry[token.ILLEGAL] = tdopEntry{}

	tdopRegistry[token.PERIOD] = infix(110)

	tdopRegistry[token.NOT] = prefix()

	tdopRegistry[token.MULT] = infix(60)
	tdopRegistry[token.DIV] = infix(60)
	tdopRegistry[token.MODULO] = infix(60)

	tdopRegistry[token.PLUS] = infix(50)
	tdopRegistry[token.MINUS] = infix(50)

	tdopRegistry[token.EQUAL] = infix(45)
	tdopRegistry[token.LESS] = infix(45)
	tdopRegistry[token.GRTR] = infix(45)
	tdopRegistry[token.NOT_EQUAL] = infix(45)
	tdopRegistry[token.LESS_EQUAL] = infix(45)
	tdopRegistry[token.GRTR_EQUAL] = infix(45)
	tdopRegistry[token.ISA] = infix(45)
	tdopRegistry[token.HASA] = infix(45)

	tdopRegistry[token.LOG_AND] = infix(40)
	tdopRegistry[token.LOG_OR] = infix(40)

	tdopRegistry[token.COMMA] = infix(35)

	tdopRegistry[token.ACCUM] = rinfix(30)
	tdopRegistry[token.ASSIGN] = rinfix(30)
	tdopRegistry[token.QASSIGN] = rinfix(30)

	tdopRegistry[token.LPIPE] = infix(10)
	tdopRegistry[token.RPIPE] = rinfix(10)

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

	tdopRegistry[token.LPAREN] = tdopEntry{}
	tdopRegistry[token.LSQR] = tdopEntry{}
	tdopRegistry[token.LBRACE] = tdopEntry{}
	tdopRegistry[token.RPAREN] = tdopEntry{}
	tdopRegistry[token.RSQR] = tdopEntry{}
	tdopRegistry[token.RBRACE] = tdopEntry{}
	tdopRegistry[token.SEMI] = tdopEntry{}
	tdopRegistry[token.COLON] = tdopEntry{}
	tdopRegistry[token.DOLLAR] = tdopEntry{}
	tdopRegistry[token.DDOLLAR] = tdopEntry{}

	tdopRegistry[token.BREAK] = tdopEntry{}
	tdopRegistry[token.CONTINUE] = tdopEntry{}
	tdopRegistry[token.ELSE] = tdopEntry{}
	tdopRegistry[token.FOR] = tdopEntry{}
	tdopRegistry[token.IN] = tdopEntry{}
	tdopRegistry[token.FUNC] = tdopEntry{}
	tdopRegistry[token.IF] = tdopEntry{}
	tdopRegistry[token.IMPORT] = tdopEntry{}
	tdopRegistry[token.PKG] = tdopEntry{}
	tdopRegistry[token.RETURN] = tdopEntry{}
	tdopRegistry[token.STRUCT] = tdopEntry{}
	tdopRegistry[token.SWITCH] = tdopEntry{}
	tdopRegistry[token.WHILE] = tdopEntry{}
	tdopRegistry[token.ENUM] = tdopEntry{}
	tdopRegistry[token.SYS] = tdopEntry{}
}

// var tdopRegistry = [...]tdopEntry{

// }

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

func prefix() tdopEntry {
	return tdopEntry{
		bindingPower: 0,
		nud: func(node *Node, p *Parser) *Node {
			node.children = append(node.children, p.expression(100))
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
		bindingPower: 0,
		nud: func(node *Node, p *Parser) *Node {
			return node
		},
	}
}

func (n *Node) nud() nudFunc {
	return tdopRegistry[n.lexeme.Token()].nud
}

func (n *Node) led() ledFunc {
	return tdopRegistry[n.lexeme.Token()].led
}

func (n *Node) std() stdFunc {
	return tdopRegistry[n.lexeme.Token()].std
}

func (n *Node) bind() int {
	return tdopRegistry[n.lexeme.Token()].bindingPower
}

var exprLevel = 0

func (p *Parser) expression(rbp int) *Node {

	exprLevel++
	e := exprLevel
	log.Printf("INN expression (%d) %d", e, rbp)

	var left *Node
	node := newNode(p.lexer.Next())

	if node.nud() != nil {
		left = node.nud()(node, p)
	} else {
		log.Printf("failed to process (1):")
		log.Fatalf("%s", node.lexeme.String())
	}

	for rbp < bindPowerOf(p.lexer.Peek()) {
		node := newNode(p.lexer.Next())
		if node.led() != nil {
			left = node.led()(node, p, left)
		} else {
			log.Printf("failed to process (2):")
			log.Fatalf("%s", node.lexeme.String())
		}
	}

	exprLevel--
	log.Printf("OUT expression (%d) %d", e, rbp)

	return left
}

func (p *Parser) advance(match token.Token) *Node {

	next := p.lexer.Next()
	if next.Token() != match {
		log.Fatalf("failed to process (3): expecting %s", match.String())
	}

	return newNode(next)
}
