package lexer

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/pdk/gosh/token"
)

// Lexer reads an input string identifying tokens.
type Lexer struct {
	input []string // input is a slice of strings. each string is a line from the input.
	lexed []Lexeme // the result of lexing an input file is a slice of Lexemes.
	pos   int      // track position for Next(), Peek()
}

// Lexemes returns the tokenized result
func (lex *Lexer) Lexemes() []Lexeme {
	return lex.lexed
}

func (lex *Lexer) LogDump() {
	for _, x := range lex.lexed {
		log.Printf("%s", x.String())
	}
}

// skipForward skips over some things that are challenging for the parser.
func (lex *Lexer) skipForward() {

	for lex.isComment() || lex.isSemiBrace() {
		lex.pos++
	}
}

// isSemiBrace returns true if there is nothing between here and the next right
// brace but comments and semicolons.
func (lex *Lexer) isSemiBrace() bool {

	if lex.pos < len(lex.lexed) && lex.lexed[lex.pos].Token() == token.SEMI {

		// make sure there's something between here and the next RBRACE.
		for i := lex.pos; i < len(lex.lexed); i++ {
			if lex.lexed[i].Token() == token.RBRACE {
				return true
			}

			if lex.lexed[i].Token() != token.SEMI && lex.lexed[i].Token() != token.COMMENT {
				break
			}
		}
	}

	return false
}

func (lex *Lexer) isComment() bool {
	if lex.pos < len(lex.lexed) &&
		lex.lexed[lex.pos].Token() == token.COMMENT {

		return true
	}

	return false
}

// Next returns the next Lexeme, and increments our position.
func (lex *Lexer) Next() *Lexeme {
	lex.skipForward()

	if lex.pos >= len(lex.lexed) {
		return nil
	}

	l := lex.lexed[lex.pos]
	lex.pos++

	return &l
}

// Peek returns the next Lexeme, but does not increment our position.
func (lex *Lexer) Peek() *Lexeme {
	lex.skipForward()

	if lex.pos >= len(lex.lexed) {
		return nil
	}

	l := lex.lexed[lex.pos]

	return &l
}

// Lexeme contains a lex'd token and the literal value.
type Lexeme struct {
	token      token.Token
	literal    string
	lineNumber int
	charNumber int
}

// newLexeme makes a new Lexeme, with literal.
func newLexeme(tok token.Token, lit string) Lexeme {
	return Lexeme{
		token:   tok,
		literal: lit,
	}
}

// WithToken will override the existing token with a new value.
func (lex Lexeme) WithToken(tok token.Token) Lexeme {
	lex.token = tok
	return lex
}

// WithLiteral will override the existing literal.
func (lex Lexeme) WithLiteral(lit string) Lexeme {
	lex.literal = lit
	return lex
}

// Rewrite creates a new Lexeme based on an existing Lexeme.
func (lex Lexeme) Rewrite(tok token.Token, lit string) *Lexeme {

	lex.token = tok
	lex.literal = lit

	return &lex
}

// at sets a Lexeme's location
func (lex Lexeme) at(lineNo, charNo int) Lexeme {

	lex.lineNumber = lineNo
	lex.charNumber = charNo

	return lex
}

// Token returns the token.Token of the Lexeme.
func (lex Lexeme) Token() token.Token {
	return lex.token
}

// LineNo returns the line number the token was found in the input.
func (lex Lexeme) LineNo() int {
	return lex.lineNumber
}

// CharNo returns the character number the token was found in the line.
func (lex Lexeme) CharNo() int {
	return lex.charNumber
}

// Literal returns the string of the actual value found in the input.
func (lex Lexeme) Literal() string {
	return lex.literal
}

// String returns a string representation of a Lexeme for user-friendly viewing.
func (lex Lexeme) String() string {
	lit := lex.literal
	if lex.token == token.STRING {
		lit = strconv.Quote(lex.literal)
	}
	return fmt.Sprintf("%3d, %3d %-10s %s", lex.lineNumber, lex.charNumber, lex.token.String(), lit)
}

// IndentString useful for printing trees of Lexemes.
func (lex Lexeme) IndentString(n int) string {
	lit := lex.literal
	if lex.token == token.STRING {
		lit = strconv.Quote(lex.literal)
	}
	return fmt.Sprintf("%3d, %3d %-10s %s%s", lex.lineNumber, lex.charNumber, lex.token.String(), strings.Repeat(" ", n), lit)
}

// New returns a new Lexer. Actually does all the work of lexing the input.
func New(input []string) *Lexer {

	l := &Lexer{
		input: input,
	}

	for lineOffset, line := range input {
		l.lexed = append(l.lexed, lex(line, lineOffset+1)...)
	}

	eof := newLexeme(token.EOF, "").at(len(input)+1, 0)
	l.lexed = append(l.lexed, eof)

	return l
}

func lex(line string, lineNo int) []Lexeme {

	var xems []Lexeme

	chars := stringRunes(line)
	l := len(chars)

	i := 0
	for i < l {
		nt, c := nextLexeme(chars[i:])
		if nt.token != token.NADA {
			xems = append(xems, nt.at(lineNo, i+1))
		}
		i += c
	}

	if len(xems) == 0 {
		return xems
	}

	// Need to check last token on the line to see if we should add a semicolon.
	// Before doing that, pull off the last item, IFF it's a comment. Later
	// we'll stick the comment back on, following the semicolon.
	comment := []Lexeme{}
	if xems[len(xems)-1].token == token.COMMENT {
		comment = append(comment, xems[len(xems)-1])
		xems = xems[0 : len(xems)-1]
	}

	if len(xems) == 0 {
		// comment was the only thing on the line
		return comment
	}

	lastTok := xems[len(xems)-1].token
	if doAddSemiAfter(lastTok) {
		xems = append(xems, newLexeme(token.SEMI, ";").at(lineNo, i+1))
	}

	// reattach comment (if any) and done
	return append(xems, comment...)
}

// doAddSemiAfter returns true if we should append a semicolon to the end of the
// line. We just use the same logic as go (basically).
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

func nextLexeme(chars []rune) (Lexeme, int) {

	i := countWhitespace(chars)
	chars = chars[i:]

	if len(chars) == 0 {
		return newLexeme(token.NADA, ""), i
	}

	ch := chars[0]

	if ch == '#' {
		return newLexeme(token.COMMENT, string(chars)), i + len(chars)
	}

	peek := ' '
	if len(chars) > 1 {
		peek = chars[1]
	}

	switch ch {

	case ':':
		if peek == '=' {
			return newLexeme(token.ASSIGN, ":="), i + 2
		}
		return newLexeme(token.COLON, ":"), i + 1

	case '!':
		if peek == '=' {
			return newLexeme(token.NOT_EQUAL, "!="), i + 2
		}
		return newLexeme(token.NOT, "!"), i + 1

	case '+':
		if peek == '=' {
			return newLexeme(token.ACCUM, "+="), i + 2
		}
		return newLexeme(token.PLUS, "+"), i + 1

	case '>':
		if peek == '>' {
			return newLexeme(token.RPIPE, ">>"), i + 2
		}
		if peek == '=' {
			return newLexeme(token.GRTR_EQUAL, ">="), i + 2
		}
		return newLexeme(token.GRTR, ">"), i + 1

	case '<':
		if peek == '<' {
			return newLexeme(token.LPIPE, "<<"), i + 2
		}
		if peek == '=' {
			return newLexeme(token.LESS_EQUAL, "<="), i + 2
		}
		return newLexeme(token.LESS, "<"), i + 1

	case '&':
		if peek == '&' {
			return newLexeme(token.LOG_AND, "&&"), i + 2
		}
		return newLexeme(token.ILLEGAL, "&"), i + 1

	case '=':
		if peek == '=' {
			return newLexeme(token.EQUAL, "=="), i + 2
		}
		return newLexeme(token.ILLEGAL, "="), i + 1

	case '|':
		if peek == '|' {
			return newLexeme(token.LOG_OR, "||"), i + 2
		}
		return newLexeme(token.ILLEGAL, "|"), i + 1

	case '?':
		if peek == '=' {
			return newLexeme(token.QASSIGN, "?="), i + 2
		}
		return newLexeme(token.ILLEGAL, "?"), i + 1

	case '-':
		return newLexeme(token.MINUS, "-"), i + 1
	case ',':
		return newLexeme(token.COMMA, ","), i + 1
	case ';':
		return newLexeme(token.SEMI, ";"), i + 1
	case '.':
		return newLexeme(token.PERIOD, "."), i + 1
	case '(':
		return newLexeme(token.LPAREN, "("), i + 1
	case ')':
		return newLexeme(token.RPAREN, ")"), i + 1
	case '[':
		return newLexeme(token.LSQR, "["), i + 1
	case ']':
		return newLexeme(token.RSQR, "]"), i + 1
	case '{':
		return newLexeme(token.LBRACE, "{"), i + 1
	case '}':
		return newLexeme(token.RBRACE, "}"), i + 1
	case '*':
		return newLexeme(token.MULT, "*"), i + 1
	case '/':
		return newLexeme(token.DIV, "/"), i + 1
	case '%':
		return newLexeme(token.MODULO, "%"), i + 1

	case '$':
		tok := token.DOLLAR
		if peek == '$' {
			tok = token.DDOLLAR
		}
		command, l := scanCommand(chars)
		return newLexeme(tok, string(command)), i + l
	}

	if unicode.IsDigit(ch) {
		number, isFloat := scanNumeric(chars)
		if isFloat {
			return newLexeme(token.FLOAT, string(number)), i + len(number)
		}
		return newLexeme(token.INT, string(number)), i + len(number)
	}

	if unicode.IsLetter(ch) || ch == '_' {
		ident := string(scanIdent(chars))
		which := token.CheckIdent(ident)
		return newLexeme(which, ident), i + len(ident)
	}

	if ch == '"' {
		str0 := string(scanString(chars))
		str, err := strconv.Unquote(str0)
		if err != nil {
			return newLexeme(token.ILLEGAL, str0), i + len(str0)
		}
		return newLexeme(token.STRING, str), i + len(str0)
	}

	return newLexeme(token.ILLEGAL, string(ch)), i + 1
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

// scanCommand reads in a "command" which is stuff after a "$" or a "$$". It
// might be a single symbol, or it might be a complex string in braces, kind of
// like a quoted string.
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
	for i < len(chars) && unicode.IsSpace(chars[i]) {
		i++
	}

	return i
}
