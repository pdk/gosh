package parse

import (
	"fmt"

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

	if p == nil || p.lexer == nil {
		return false
	}

	l := p.lexer.Peek()
	if l == nil {
		return false
	}

	return l.Token() == tok
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
	P_RETURN
	P_COMMA
	P_LOGIC
	P_COMPARE
	P_PLUSMINUS
	P_MULTDIV
	P_PREFIX
	P_BRACKET
	P_CONTROL
	P_PERIOD
)

func init() {
	tdopRegistry[token.ILLEGAL] = tdopEntry{}

	tdopRegistry[token.IDENT] = self()
	tdopRegistry[token.INT] = self()
	tdopRegistry[token.FLOAT] = self()
	tdopRegistry[token.CHAR] = self()
	tdopRegistry[token.STRING] = self()
	tdopRegistry[token.NIL] = self()
	tdopRegistry[token.TRUE] = self()
	tdopRegistry[token.FALSE] = self()
	tdopRegistry[token.BREAK] = self()
	tdopRegistry[token.CONTINUE] = self()

	tdopRegistry[token.PKG] = prefix(P_PREFIX)
	tdopRegistry[token.NOT] = prefix(P_PREFIX)

	tdopRegistry[token.MINUS] = prefixInfix(P_PREFIX, P_PLUSMINUS)
	tdopRegistry[token.RETURN] = prefixOrNaught(P_RETURN)

	tdopRegistry[token.PERIOD] = infix(P_PERIOD)
	tdopRegistry[token.MULT] = infix(P_MULTDIV)
	tdopRegistry[token.DIV] = infix(P_MULTDIV)
	tdopRegistry[token.MODULO] = infix(P_MULTDIV)
	tdopRegistry[token.PLUS] = infix(P_PLUSMINUS)
	tdopRegistry[token.EQUAL] = infix(P_COMPARE)
	tdopRegistry[token.LESS] = infix(P_COMPARE)
	tdopRegistry[token.GRTR] = infix(P_COMPARE)
	tdopRegistry[token.NOT_EQUAL] = infix(P_COMPARE)
	tdopRegistry[token.LESS_EQUAL] = infix(P_COMPARE)
	tdopRegistry[token.GRTR_EQUAL] = infix(P_COMPARE)
	tdopRegistry[token.ISA] = infix(P_COMPARE)
	tdopRegistry[token.HASA] = infix(P_COMPARE)
	tdopRegistry[token.COMMA] = infix(P_COMMA)
	tdopRegistry[token.LOG_AND] = infix(P_LOGIC)
	tdopRegistry[token.LOG_OR] = infix(P_LOGIC)
	tdopRegistry[token.LPIPE] = infix(P_PIPE)

	tdopRegistry[token.SEMI] = infixOrNaught(P_SEPARATOR)

	tdopRegistry[token.RPIPE] = rinfix(P_PIPE)
	tdopRegistry[token.ASSIGN] = rinfix(P_ASSIGN)
	tdopRegistry[token.ACCUM] = rinfix(P_ASSIGN)
	tdopRegistry[token.QASSIGN] = rinfix(P_ASSIGN)

	tdopRegistry[token.LPAREN] = leftBracket(P_BRACKET, token.RPAREN)
	tdopRegistry[token.LSQR] = leftBracket(P_BRACKET, token.RSQR)

	tdopRegistry[token.WHILE] = whileExpr(P_CONTROL)
	tdopRegistry[token.FUNC] = funcExpr(P_CONTROL)
	tdopRegistry[token.IF] = ifExpr(P_CONTROL)
	tdopRegistry[token.EXTERN] = externExpr(P_CONTROL)

	// TODO
	tdopRegistry[token.COLON] = tdopEntry{}   // named parameters on function invocation
	tdopRegistry[token.DOLLAR] = tdopEntry{}  // execute system command, return pipe
	tdopRegistry[token.DDOLLAR] = tdopEntry{} // execute bash command, return pipe
	tdopRegistry[token.FOR] = tdopEntry{}     // pipe consumption loop
	tdopRegistry[token.IMPORT] = tdopEntry{}  // load another file
	tdopRegistry[token.STRUCT] = tdopEntry{}  // define a compound type
	tdopRegistry[token.SWITCH] = tdopEntry{}  // multibranch conditional
	tdopRegistry[token.ENUM] = tdopEntry{}    // define an enumeration
	tdopRegistry[token.SYS] = tdopEntry{}     // synonym for $, $$, but take expression
}

// newNode makes a parse.Node out of a lexeme.
func newNode(lex *lexer.Lexeme) *Node {
	return &Node{
		lexeme: lex,
	}
}

// lefty sets a node to be a lefty.
func (n *Node) lefty() *Node {
	n.arity = Lefty
	return n
}

// righty sets a node to be a righty.
func (n *Node) righty() *Node {
	n.arity = Righty
	return n
}

// IsLefty returns true IFF the node has arity Lefty.
func (n *Node) IsLefty() bool {
	return n.arity == Lefty
}

// IsRighty returns true IFF the node has arity Righty.
func (n *Node) IsRighty() bool {
	return n.arity == Righty

}

// bindPowerOf looks up the binding power in the registry.
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

func externExpr(bp int) tdopEntry {
	return tdopEntry{
		bindingPower: bp,
		nud: func(node *Node, p *Parser) (*Node, error) {

			for {
				if p.peekIs(token.IDENT) {
					ident, err := p.advance(token.IDENT)
					node.children = append(node.children, ident)
					if err != nil {
						return node, err
					}
					continue
				}
				if p.peekIs(token.COMMA) {
					_, err := p.advance(token.COMMA)
					if err != nil {
						return node, err
					}
					continue
				}
				if p.peekIs(token.SEMI) || p.peekIs(token.RBRACE) {
					return node, nil
				}

				node := newNode(p.next())
				return node, parseError(node, "extern expecting identifier")
			}
		},
	}
}

// whileExpr parses a while expression. This is probably the simplest control
// structure, syntactically.
// while ... { ... }
func whileExpr(bp int) tdopEntry {
	return tdopEntry{
		bindingPower: bp,
		nud: func(node *Node, p *Parser) (*Node, error) {

			exp, err := p.expression(0)
			node.children = append(node.children, exp)
			if err != nil {
				return node, err
			}

			// Use standard block parser to add function body.
			return parseBlockToChild(node, p)
		},
	}
}

// funcExpr parses a function definition, which are of these forms:
// func(...) {...}
// func(...) [...] {...}
// That is, keyword `func` followed by param list in parens,
// optionally a list of output channel names, and then the body of
// the function in curly braces.
func funcExpr(bp int) tdopEntry {
	return tdopEntry{
		bindingPower: bp,
		nud: func(node *Node, p *Parser) (*Node, error) {

			paramNode, err := p.advance(token.LPAREN)
			if err != nil {
				return node, err
			}

			paramNode, err = parseCommaListUntil(paramNode, p, token.RPAREN)
			node.children = append(node.children, paramNode)
			if err != nil {
				return node, err
			}

			if p.peekIs(token.LSQR) {
				channelNode, err := p.advance(token.LSQR)
				if err != nil {
					return node, err
				}

				channelNode, err = parseCommaListUntil(channelNode, p, token.RSQR)
				node.children = append(node.children, channelNode)
				if err != nil {
					return node, err
				}
			} else {
				// Make sure there is an empty channel list as child position 2,
				// even if no channels were named.
				xeme := node.Lexeme().Lexer().NewLexeme(token.LSQR, "[")
				channelNode := newNode(&xeme)
				node.children = append(node.children, channelNode)
			}

			// Use standard block parser to add function body.
			return parseBlockToChild(node, p)
		},
	}
}

// parseBlockToChild is used for various command structures that expect a
// brace-bounded collection of statements. Used for if, for, while, func, ...
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

// leftBracket is used for both left parenthesis and left square bracket. Both
// of thsee may be infix or prefix operators.
// a[x] map/list lookup
// [a] a list
// (a,b) a pair of values
// f() invoking a function
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

// infix: basic left-to-right binary operators.
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

func parseError(at *Node, mesg string) error {
	if at == nil || at.lexeme == nil {
		return fmt.Errorf("parse error: %s", mesg)
	}

	return fmt.Errorf("parse error on line %d, pos %d, token %s: %s",
		at.lexeme.LineNo(), at.lexeme.CharNo(), at.lexeme.Literal(), mesg)
}

// expression is the magical driver of the top down operator precedence parser.
// this is the beating heart of the parser.
// see https://www.youtube.com/watch?v=Nlqv6NtBXcA
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

// advance asserts that the next token is a particular token, and consumes it,
// return the parse node thus created.
func (p *Parser) advance(match token.Token) (*Node, error) {

	l := p.next()
	if l == nil {
		return nil, parseError(nil, "ran out of input: no more tokens")
	}

	next := newNode(l)

	if next.Token() != match {
		return next, parseError(next, fmt.Sprintf("expecting %s", match.String()))
	}

	return next, nil
}
