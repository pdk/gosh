package lexer

import (
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/pdk/gosh/token"
)

// Lexer reads an input string identifying tokens.
type Lexer struct {
	input  []string
	tokens []Token
}

// Toks returns the tokenized result
func (lex *Lexer) Toks() []Token {
	return lex.tokens
}

// Token contains a lex'd token and the literal value.
type Token struct {
	token      token.Token
	literal    string
	lineNumber int
	charNumber int
}

func (tok Token) at(lineNo, charNo int) Token {

	tok.lineNumber = lineNo
	tok.charNumber = charNo

	return tok
}

func (tok Token) String() string {
	lit := tok.literal
	if tok.token == token.STRING {
		lit = strconv.Quote(tok.literal)
	}
	return fmt.Sprintf("%3d, %3d %-10s %s", tok.lineNumber, tok.charNumber, tok.token.String(), lit)
}

// New returns a new Lexer.
func New(input []string) *Lexer {

	l := &Lexer{input: input}

	for lineOffset, line := range input {
		l.tokens = append(l.tokens, lex(line, lineOffset+1)...)
	}

	eof := newToken(token.EOF, "").at(len(input)+1, 0)
	l.tokens = append(l.tokens, eof)

	return l
}

func lex(line string, lineNo int) []Token {

	var toks []Token

	chars := stringRunes(line)
	l := len(chars)

	i := 0
	for i < l {
		nt, c := nextToken(chars[i:])
		if nt.token != token.NADA {
			toks = append(toks, nt.at(lineNo, i+1))
		}
		i += c
	}

	if len(toks) == 0 {
		return toks
	}

	comment := []Token{}
	if toks[len(toks)-1].token == token.COMMENT {
		comment = append(comment, toks[len(toks)-1])
		toks = toks[0 : len(toks)-1]
	}

	if len(toks) == 0 {
		return append(toks, comment...)
	}

	lastTok := toks[len(toks)-1].token
	if doAddSemiAfter(lastTok) {
		toks = append(toks, newToken(token.SEMI, ";").at(lineNo, i+1))
	}

	return append(toks, comment...)
}

func doAddSemiAfter(lastTok token.Token) bool {

	// https://medium.com/golangspec/automatic-semicolon-insertion-in-go-1990338f2649

	if lastTok == token.IDENT ||
		lastTok == token.INT ||
		lastTok == token.FLOAT ||
		lastTok == token.CHAR ||
		lastTok == token.STRING ||
		lastTok == token.BREAK ||
		lastTok == token.CONTINUE ||
		lastTok == token.RETURN ||
		lastTok == token.RPAREN ||
		lastTok == token.RSQR ||
		lastTok == token.RBRACE ||
		lastTok == token.DOLLAR ||
		lastTok == token.DDOLLAR {

		return true
	}

	return false
}

func stringRunes(line string) []rune {

	len := utf8.RuneCount([]byte(line))
	chars := make([]rune, len, len)

	i := 0
	for _, ch := range line {
		chars[i] = ch
		i++
	}

	return chars
}

func nextToken(chars []rune) (Token, int) {

	i := countWhitespace(chars)
	chars = chars[i:]

	if len(chars) == 0 {
		return newToken(token.NADA, ""), i
	}

	ch := chars[0]

	if ch == '#' {
		return Token{
			token:   token.COMMENT,
			literal: string(chars),
		}, i + len(chars)
	}

	peek := ' '
	if len(chars) > 1 {
		peek = chars[1]
	}

	switch ch {

	case ':':
		if peek == '=' {
			return newToken(token.ASSIGN, ":="), i + 2
		}
		return newToken(token.COLON, ":"), i + 1

	case '!':
		if peek == '=' {
			return newToken(token.NOT_EQUAL, "!="), i + 2
		}
		return newToken(token.NOT, "!"), i + 1

	case '+':
		if peek == '=' {
			return newToken(token.ACCUM, "+="), i + 2
		}
		return newToken(token.PLUS, "+"), i + 1

	case '>':
		if peek == '>' {
			return newToken(token.RPIPE, ">>"), i + 2
		}
		if peek == '=' {
			return newToken(token.GRTR_EQUAL, ">="), i + 2
		}
		return newToken(token.GRTR, ">"), i + 1

	case '<':
		if peek == '<' {
			return newToken(token.LPIPE, "<<"), i + 2
		}
		if peek == '=' {
			return newToken(token.LESS_EQUAL, "<="), i + 2
		}
		return newToken(token.LESS, "<"), i + 1

	case '&':
		if peek == '&' {
			return newToken(token.LOG_AND, "&&"), i + 2
		}
		return newToken(token.ILLEGAL, "&"), i + 1

	case '=':
		if peek == '=' {
			return newToken(token.EQUAL, "=="), i + 2
		}
		return newToken(token.ILLEGAL, "="), i + 1

	case '|':
		if peek == '|' {
			return newToken(token.LOG_OR, "||"), i + 2
		}
		return newToken(token.ILLEGAL, "|"), i + 1

	case '-':
		return newToken(token.MINUS, "-"), i + 1
	case ',':
		return newToken(token.COMMA, ","), i + 1
	case ';':
		return newToken(token.SEMI, ";"), i + 1
	case '.':
		return newToken(token.PERIOD, "."), i + 1
	case '(':
		return newToken(token.LPAREN, "("), i + 1
	case ')':
		return newToken(token.RPAREN, ")"), i + 1
	case '[':
		return newToken(token.LSQR, "["), i + 1
	case ']':
		return newToken(token.RSQR, "]"), i + 1
	case '{':
		return newToken(token.LBRACE, "{"), i + 1
	case '}':
		return newToken(token.RBRACE, "}"), i + 1
	case '*':
		return newToken(token.MULT, "*"), i + 1
	case '/':
		return newToken(token.DIV, "/"), i + 1
	case '%':
		return newToken(token.MODULO, "%"), i + 1

	case '$':
		tok := token.DOLLAR
		if peek == '$' {
			tok = token.DDOLLAR
		}
		command, l := scanCommand(chars)
		return newToken(tok, string(command)), i + l
	}

	if unicode.IsDigit(ch) {
		number, isFloat := scanNumeric(chars)
		if isFloat {
			return newToken(token.FLOAT, string(number)), i + len(number)
		}
		return newToken(token.INT, string(number)), i + len(number)
	}

	if unicode.IsLetter(ch) || ch == '_' {
		ident := string(scanIdent(chars))
		which := token.CheckIdent(ident)
		return newToken(which, ident), i + len(ident)
	}

	if ch == '"' {
		str0 := string(scanString(chars))
		str, err := strconv.Unquote(str0)
		if err != nil {
			return newToken(token.ILLEGAL, str0), i + len(str0)
		}
		return newToken(token.STRING, str), i + len(str0)
	}

	return newToken(token.ILLEGAL, string(ch)), i + 1
}

func scanString(chars []rune) []rune {
	var r []rune

	lastC := '\\'
	for _, c := range chars {
		if lastC != '\\' && c == '"' {
			r = append(r, c)
			return r
		}
		r = append(r, c)
		lastC = c
	}

	return r
}

func scanCommand(chars []rune) ([]rune, int) {
	var r []rune
	c := 0
	if chars[0] == '$' {
		c++
		chars = chars[1:]
	}
	if len(chars) == 0 {
		return r, c
	}
	if chars[0] == '$' {
		c++
		chars = chars[1:]
	}
	if len(chars) == 0 {
		return r, c
	}

	n := countWhitespace(chars)
	chars = chars[n:]
	c += n
	if len(chars) == 0 {
		return r, c
	}

	if unicode.IsLetter(chars[0]) || chars[0] == '_' {
		ident := scanIdent(chars)
		return ident, c + len(ident)
	}

	if chars[0] != '{' {
		return r, c
	}

	chars = chars[1:]
	c++

	for _, ch := range chars {
		if ch == '}' {
			return r, c + 1
		}
		r = append(r, ch)
		c++
	}

	return r, c
}

func scanNumeric(chars []rune) ([]rune, bool) {
	var r []rune
	gotDot := false
	for _, c := range chars {
		if c == '.' {
			if gotDot {
				return r, true
			}
			gotDot = true
		}
		if unicode.IsDigit(c) || c == '.' {
			r = append(r, c)
		} else {
			return r, gotDot
		}
	}

	return r, gotDot
}

func scanIdent(chars []rune) []rune {
	var r []rune
	for _, c := range chars {
		if unicode.IsLetter(c) || c == '_' {
			r = append(r, c)
		} else {
			return r
		}
	}

	return r
}

func countWhitespace(chars []rune) int {

	i := 0
	for unicode.IsSpace(chars[i]) {
		i++
	}

	return i
}

func newToken(tok token.Token, lit string) Token {
	return Token{
		token:   tok,
		literal: lit,
	}
}
