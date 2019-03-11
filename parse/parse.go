package parse

import (
	"fmt"
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
func (p *Parser) Parse() (*Node, error) {

	ast, err := p.expression(0)
	if err != nil {
		return ast, err
	}

	ast = ast.applyTransforms()

	return ast, nil
}

// peek return the upcoming token from the lexer.
func (p *Parser) peek() *lexer.Lexeme {
	return p.lexer.Peek()
}

// next returns the next Lexeme from the lexer.
func (p *Parser) next() *lexer.Lexeme {
	return p.lexer.Next()
}

// pToken returns the token of the peek.
func (p *Parser) peekIs(tok token.Token) bool {
	return p.lexer.Peek().Token() == tok
}

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

// Children returns the children of a parsed node.
func (n *Node) Children() []*Node {
	return n.children
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

type nudFunc func(*Node, *Parser) (*Node, error)

type ledFunc func(*Node, *Parser, *Node) (*Node, error)

// tdopEntry contains the information required to drive the parser.
type tdopEntry struct {
	bindingPower int
	nud          nudFunc
	led          ledFunc
}

var tdopRegistry [token.KeywordEnd]tdopEntry

// Ordered precedence values. (Even nubmers cuz we need -1 on right-associative
// infix operators.)
const (
	P_EOF = iota * 2
	P_UNEXPECTED

	P_SEPARATOR
	P_SELF

	P_PIPE
	P_ASSIGN
	P_COMMA
	P_LOGIC
	P_RETURN
	P_COMPARE
	P_PLUSMINUS
	P_MULTDIV
	P_PREFIX
	P_CONTROL
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

	tdopRegistry[token.RETURN] = prefixOrNaught(P_RETURN)

	tdopRegistry[token.LOG_AND] = infix(P_LOGIC)
	tdopRegistry[token.LOG_OR] = infix(P_LOGIC)

	tdopRegistry[token.BREAK] = self()
	tdopRegistry[token.CONTINUE] = self()

	tdopRegistry[token.ASSIGN] = rinfix(P_ASSIGN)
	tdopRegistry[token.ACCUM] = rinfix(P_ASSIGN)
	tdopRegistry[token.QASSIGN] = rinfix(P_ASSIGN)

	tdopRegistry[token.LPIPE] = infix(P_PIPE)
	tdopRegistry[token.RPIPE] = rinfix(P_PIPE)

	tdopRegistry[token.EOF] = eof(P_EOF)
	tdopRegistry[token.COMMENT] = consumable()

	tdopRegistry[token.IDENT] = self()
	tdopRegistry[token.INT] = self()
	tdopRegistry[token.FLOAT] = self()
	tdopRegistry[token.CHAR] = self()
	tdopRegistry[token.STRING] = self()
	tdopRegistry[token.NIL] = self()
	tdopRegistry[token.TRUE] = self()
	tdopRegistry[token.FALSE] = self()

	tdopRegistry[token.SEMI] = infixOrNaught(P_SEPARATOR)

	tdopRegistry[token.LPAREN] = leftBracket(P_FUNC, token.RPAREN)
	tdopRegistry[token.LSQR] = leftBracket(P_FUNC, token.RSQR)
	tdopRegistry[token.LBRACE] = tdopEntry{}

	tdopRegistry[token.RPAREN] = tdopEntry{} // unbalanced()
	tdopRegistry[token.RSQR] = tdopEntry{}   // unbalanced()
	tdopRegistry[token.RBRACE] = tdopEntry{} // unbalanced()

	tdopRegistry[token.COLON] = tdopEntry{}
	tdopRegistry[token.DOLLAR] = tdopEntry{}
	tdopRegistry[token.DDOLLAR] = tdopEntry{}

	tdopRegistry[token.ELSE] = tdopEntry{}
	tdopRegistry[token.FOR] = tdopEntry{}
	tdopRegistry[token.IN] = tdopEntry{}
	tdopRegistry[token.FUNC] = tdopEntry{}
	tdopRegistry[token.IF] = ifExpr(P_CONTROL)
	tdopRegistry[token.IMPORT] = tdopEntry{}
	tdopRegistry[token.PKG] = prefix(P_PREFIX)
	tdopRegistry[token.STRUCT] = tdopEntry{}
	tdopRegistry[token.SWITCH] = tdopEntry{}
	tdopRegistry[token.WHILE] = tdopEntry{}
	tdopRegistry[token.ENUM] = tdopEntry{}
	tdopRegistry[token.SYS] = tdopEntry{}
}

func eof(bp int) tdopEntry {
	return tdopEntry{
		bindingPower: bp,
		nud: func(node *Node, p *Parser) (*Node, error) {
			return node, nil
		},
	}
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

func (n *Node) lefty() *Node {
	n.arity = Lefty
	return n
}

func (n *Node) righty() *Node {
	n.arity = Righty
	return n
}

func (n *Node) IsLefty() bool {
	return n.arity == Lefty
}

func (n *Node) IsRighty() bool {
	return n.arity == Righty

}

func bindPowerOf(lex *lexer.Lexeme) int {
	return tdopRegistry[lex.Token()].bindingPower
}

func ifExprNud(node *Node, p *Parser) (*Node, error) {

	exp, err := p.expression(0)
	node.children = append(node.children, exp)
	if err != nil {
		return node, err
	}

	node, err = parseBlockToChild(node, p)
	if err != nil {
		return node, err
	}

	if !p.peekIs(token.ELSE) {
		return node, nil
	}

	_, err = p.advance(token.ELSE)
	if err != nil {
		return node, err
	}

	switch p.peek().Token() {
	case token.LBRACE:
		// simple else condition

		node, err = parseBlockToChild(node, p)
		if err != nil {
			return node, err
		}

	case token.IF:
		// chained if

		_, err = p.advance(token.IF)
		if err != nil {
			return node, err
		}

		return ifExprNud(node, p)

	case token.SEMI:
		break

	default:
		return node, parseError(newNode(p.peek()), "expecting either { or if")
	}

	return node, err
}

func ifExpr(bp int) tdopEntry {
	return tdopEntry{
		bindingPower: bp,
		nud:          ifExprNud,
	}
}

func parseBlockToChild(node *Node, p *Parser) (*Node, error) {

	_, err := p.advance(token.LBRACE)
	if err != nil {
		return node, err
	}

	if p.peekIs(token.RBRACE) {
		brace := p.next()

		empty := brace.WithToken(token.SEMI).WithLiteral(";")

		node.children = append(node.children, newNode(&empty))
		return node, nil
	}

	exp, err := p.expression(P_UNEXPECTED)
	node.children = append(node.children, exp)
	if err != nil {
		return node, err
	}

	_, err = p.advance(token.RBRACE)
	return node, err
}

func leftBrace(bindPower int) tdopEntry {
	return tdopEntry{
		bindingPower: bindPower,
		nud: func(node *Node, p *Parser) (*Node, error) {

			n, err := p.expression(P_UNEXPECTED)
			if err != nil {
				return n, err
			}

			_, err = p.advance(token.RBRACE)
			return n, err
		},
		led: func(node *Node, p *Parser, left *Node) (*Node, error) {

			node.children = append(node.children, left)
			if !p.peekIs(token.RBRACE) {

				exp, err := p.expression(P_UNEXPECTED)
				// exp, err := p.expression(P_SEPARATOR)
				if err != nil {
					return node, err
				}

				node.children = append(node.children, exp)

				for p.peekIs(token.SEMI) {
					_, err := p.advance(token.SEMI)
					if err != nil {
						return node, err
					}

					exp, err := p.expression(P_SEPARATOR)
					if err != nil {
						return node, err
					}

					node.children = append(node.children, exp)
				}
			}

			_, err := p.advance(token.RBRACE)
			return node, err
		},
	}
}

func parseCommaListUntil(node *Node, p *Parser, rightBracket token.Token) (*Node, error) {
	for {
		if p.peekIs(rightBracket) {
			_, err := p.advance(rightBracket)
			return node, err
		}

		exp, err := p.expression(P_COMMA)
		node.children = append(node.children, exp)
		if err != nil {
			return node, err
		}

		for p.peekIs(token.COMMA) {
			_, err := p.advance(token.COMMA)
			if err != nil {
				return node, err
			}
		}
	}
}

func leftBracket(bindPower int, rightBracket token.Token) tdopEntry {
	return tdopEntry{
		bindingPower: bindPower,
		nud: func(node *Node, p *Parser) (*Node, error) {
			return parseCommaListUntil(node, p, rightBracket)
		},
		led: func(node *Node, p *Parser, left *Node) (*Node, error) {
			node.children = append(node.children, left)
			return parseCommaListUntil(node, p, rightBracket)
		},
	}
}

// func unbalanced() tdopEntry {
// 	return tdopEntry{
// 		bindingPower: P_UNEXPECTED,
// 	}
// }

func prefix(rightBP int) tdopEntry {
	return tdopEntry{
		bindingPower: 0,
		nud: func(node *Node, p *Parser) (*Node, error) {
			exp, err := p.expression(rightBP)
			if err != nil {
				return node, err
			}

			node.children = append(node.children, exp)
			return node, nil
		},
	}
}

func prefixOrNaught(rightBP int) tdopEntry {
	return tdopEntry{
		bindingPower: 0,
		nud: func(node *Node, p *Parser) (*Node, error) {

			if p.peekIs(token.SEMI) {
				return node, nil
			}

			exp, err := p.expression(rightBP)
			if err != nil {
				return node, err
			}

			node.children = append(node.children, exp)
			return node, nil
		},
	}
}

func infix(bindingPower int) tdopEntry {
	return tdopEntry{
		bindingPower: bindingPower,
		led: func(node *Node, p *Parser, left *Node) (*Node, error) {
			node.children = append(node.children, left)

			exp, err := p.expression(bindingPower)
			if err != nil {
				return node, err
			}

			node.children = append(node.children, exp)

			return node, nil
		},
	}
}

// rinfix for infix operators that eval from right to left instead of left to
// right. (e.g. assignment)
func rinfix(bindingPower int) tdopEntry {
	return tdopEntry{
		bindingPower: bindingPower,
		led: func(node *Node, p *Parser, left *Node) (*Node, error) {
			node.children = append(node.children, left)

			exp, err := p.expression(bindingPower - 1)
			if err != nil {
				return node, err
			}

			node.children = append(node.children, exp)

			return node, nil
		},
	}
}

// infixOrNaught for an operator (";") which may or may not have something
// following.
func infixOrNaught(bindingPower int) tdopEntry {
	return tdopEntry{
		bindingPower: bindingPower,
		nud: func(node *Node, p *Parser) (*Node, error) {
			exp, err := p.expression(bindingPower)
			if err != nil {
				return node, err
			}

			node.children = append(node.children, exp)

			return node, nil
		},
		led: func(node *Node, p *Parser, left *Node) (*Node, error) {
			node.children = append(node.children, left)
			if !p.peekIs(token.EOF) {
				exp, err := p.expression(bindingPower)
				if err != nil {
					return node, err
				}

				node.children = append(node.children, exp)
			}
			return node, nil
		},
	}
}

func prefixInfix(prefixBP, infixBP int) tdopEntry {
	return tdopEntry{
		bindingPower: infixBP,
		nud: func(node *Node, p *Parser) (*Node, error) {
			exp, err := p.expression(prefixBP)
			if err != nil {
				return node, err
			}

			node.children = append(node.children, exp)
			return node, nil
		},
		led: func(node *Node, p *Parser, left *Node) (*Node, error) {
			node.children = append(node.children, left)
			exp, err := p.expression(infixBP)
			if err != nil {
				return node, err
			}

			node.children = append(node.children, exp)
			return node, nil
		},
	}
}

func self() tdopEntry {
	return tdopEntry{
		bindingPower: P_SELF,
		nud: func(node *Node, p *Parser) (*Node, error) {
			return node, nil
		},
	}
}

func (n *Node) nud() nudFunc {
	return tdopRegistry[n.Token()].nud
}

func (n *Node) led() ledFunc {
	return tdopRegistry[n.Token()].led
}

func (n *Node) bind() int {
	return tdopRegistry[n.Token()].bindingPower
}

func parseError(at *Node, mesg string) error {
	return fmt.Errorf("parse error on line %d, pos %d, token %s: %s",
		at.lexeme.LineNo(), at.lexeme.CharNo(), at.lexeme.Literal(), mesg)
}

func (p *Parser) expression(rbp int) (*Node, error) {

	var err error
	var left *Node
	node := newNode(p.next()).righty()

	if node.Token() == token.EOF {
		return node, nil
	}

	if node.nud() == nil {
		return node, parseError(node, "unexpected token (left)")
	}

	left, err = node.nud()(node, p)
	if err != nil {
		return node, err
	}

	for rbp < bindPowerOf(p.peek()) {

		node := newNode(p.next()).lefty()
		if node.led() == nil {
			return node, parseError(node, "unexpected token (right)")
		}

		left, err = node.led()(node, p, left)
		if err != nil {
			return node, err
		}
	}

	return left, nil
}

func (p *Parser) advance(match token.Token) (*Node, error) {

	next := newNode(p.next())

	if next.Token() != match {
		return next, parseError(next, fmt.Sprintf("expecting %s", match.String()))
	}

	return next, nil
}
