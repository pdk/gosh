package lexer

import (
	"github.com/pdk/gosh/token"
)

// Lexer reads an input string identifying tokens.
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

// New returns a new Lexer.
func New(input string) *Lexer {
	l := &Lexer{input: input}
	return l
}

// checkTwoByteToken checks if this we're at a 2-byte token, or if it's just a
// lone single-byte.
func (l *Lexer) checkTwoByteToken(nextByte byte, typeIfMatch token.TokenType, typeIfNotMatch token.TokenType) token.Token {

	if l.peekChar() != nextByte {
		return newToken(typeIfNotMatch, l.ch)
	}

	// match! it's the double-char token
	ch := l.ch
	l.readChar()
	lit := string(ch) + string(l.ch)
	return token.Token{Type: typeIfMatch, Literal: lit}
}

// NextToken returns the next token from the input stream.
func (l *Lexer) NextToken() token.Token {

	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case '=':
		switch l.peekChar() {
		case '=':
			l.readChar()
			tok = newToken2(token.EQ, "==")
		default:
			tok = newToken(token.ILLEGAL, "=")
		}
	case '!':
		tok = l.checkTwoByteToken('=', token.NOT_EQ, token.BANG)
	case ':':
		tok = l.checkTwoByteToken('=', token.ASSIGN, token.COLON)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '<':
		tok = l.checkTwoByteToken('=', token.LT_EQ, token.LT)
	case '>':
		switch l.peekChar() {
		case '=':
			l.readChar()
			tok = newToken2(token.GT_EQ, ">=")
		case '>':
			l.readChar()
			tok = newToken2(token.PIPE, ">>")
		default:
			tok = newToken(token.GT, '>')
		}
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case ']':
		tok = newToken(token.RSQR, l.ch)
	case '[':
		tok = newToken(token.LSQR, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {

	for {
		// # is comment marker
		if l.ch == '#' {
			l.skipToEndOfLine()
		}

		if l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
			l.readChar()
			continue
		}

		break
	}
}

func (l *Lexer) skipToEndOfLine() {
	for l.ch != '\n' || l.ch != '\r' {
		l.readChar()
	}
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: string(ch),
	}
}

func newToken2(tokenType token.TokenType, lit string) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: lit,
	}
}
