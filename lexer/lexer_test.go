package lexer_test

import (
	"strings"
	"testing"

	"github.com/pdk/gosh/lexer"
	"github.com/pdk/gosh/reader"
	"github.com/pdk/gosh/token"
)

func TestIdent(t *testing.T) {

	lines := reader.ReadLinesToStrings(strings.NewReader("blah"))
	l := lexer.New("testing", lines)

	if len(l.Lexemes()) != 3 {
		t.Errorf("expected 3 tokens, got %d", len(l.Lexemes()))
	}

	checkTokens(t, l, token.IDENT, token.SEMI, token.EOF)
}

func TestParens(t *testing.T) {

	checkLexed(t, "()", token.LPAREN, token.RPAREN, token.SEMI, token.EOF)
	checkLexed(t, "(a)", token.LPAREN, token.IDENT, token.RPAREN, token.SEMI, token.EOF)
	checkLexed(t, "(a,b)", token.LPAREN, token.IDENT, token.COMMA, token.IDENT, token.RPAREN, token.SEMI, token.EOF)
	checkLexed(t, "(,a,)", token.LPAREN, token.COMMA, token.IDENT, token.COMMA, token.RPAREN, token.SEMI, token.EOF)
}

func TestBrace(t *testing.T) {

	checkLexed(t, "{}", token.LBRACE, token.RBRACE, token.SEMI, token.EOF)
	checkLexed(t, "{\n\n}", token.LBRACE, token.RBRACE, token.SEMI, token.EOF)
	checkLexed(t, "{\n# comment\n}", token.LBRACE, token.COMMENT, token.RBRACE, token.SEMI, token.EOF)
	checkLexed(t, "{a}", token.LBRACE, token.IDENT, token.RBRACE, token.SEMI, token.EOF)
	checkLexed(t, "{a,b}", token.LBRACE, token.IDENT, token.COMMA, token.IDENT, token.RBRACE, token.SEMI, token.EOF)
	checkLexed(t, "{,a,}", token.LBRACE, token.COMMA, token.IDENT, token.COMMA, token.RBRACE, token.SEMI, token.EOF)

	checkNext(t, "{\n\n}", token.LBRACE, token.RBRACE, token.SEMI, token.EOF)
	checkNext(t, "{\n# comment\n}", token.LBRACE, token.RBRACE, token.SEMI, token.EOF)
	checkNext(t, "{\n# comment\na\n}", token.LBRACE, token.IDENT, token.RBRACE, token.SEMI, token.EOF)
	checkNext(t, "{a}", token.LBRACE, token.IDENT, token.RBRACE, token.SEMI, token.EOF)
	checkNext(t, "{a\nb\n}", token.LBRACE, token.IDENT, token.SEMI, token.IDENT, token.RBRACE, token.SEMI, token.EOF)
	checkNext(t, "{a,b}", token.LBRACE, token.IDENT, token.COMMA, token.IDENT, token.RBRACE, token.SEMI, token.EOF)
	checkNext(t, "{,a,}", token.LBRACE, token.COMMA, token.IDENT, token.COMMA, token.RBRACE, token.SEMI, token.EOF)
}

func TestIf(t *testing.T) {

	checkLexed(t, "if p {}", token.IF, token.IDENT, token.LBRACE, token.RBRACE, token.SEMI, token.EOF)
	checkNext(t, "if p {}", token.IF, token.IDENT, token.LBRACE, token.RBRACE, token.SEMI, token.EOF)
	checkNext(t, "if a!=b{\na:=b # cmt\n}", token.IF, token.IDENT, token.NOT_EQUAL, token.IDENT,
		token.LBRACE, token.IDENT, token.ASSIGN, token.IDENT, token.RBRACE, token.SEMI, token.EOF)
}

func checkLexed(t *testing.T, input string, expected ...token.Token) {

	lines := reader.ReadLinesToStrings(strings.NewReader(input))
	l := lexer.New("testing", lines)

	checkTokens(t, l, expected...)
}

func toksOfLexed(lexed []lexer.Lexeme) []token.Token {
	var x []token.Token
	for _, y := range lexed {
		x = append(x, y.Token())
	}
	return x
}

// checkNext makes sure that the .Next() function returns the right series of tokens.
func checkNext(t *testing.T, input string, expected ...token.Token) {

	lines := reader.ReadLinesToStrings(strings.NewReader(input))
	lex := lexer.New("testing", lines)

	match := true
	var fromNext []token.Token

	for _, e := range expected {
		t := lex.Next()
		if t != nil {
			if t.Token() != e {
				match = false
			}
			fromNext = append(fromNext, t.Token())
		}
	}

	if match {
		return
	}

	lex.LogDump()

	t.Errorf("tokens from .Next() did not match: expected %s, got %s", expected, fromNext)
}

// checkTokens makes sure the lexer got all the tokens.
func checkTokens(t *testing.T, lex *lexer.Lexer, expected ...token.Token) {

	toks := toksOfLexed(lex.Lexemes())

	match := true
	if len(toks) != len(expected) {
		match = false
	} else {
		for i, t := range expected {
			if toks[i] != t {
				match = false
				break
			}
		}
	}
	if match {
		return
	}

	lex.LogDump()

	t.Errorf("tokens did not match: expected %s, got %s", expected, toks)
}

func TestStrings(t *testing.T) {

	checkLexed(t, `"hello"`, token.STRING, token.SEMI, token.EOF)
	checkLexed(t, `"hello" "world"`, token.STRING, token.STRING, token.SEMI, token.EOF)
	checkLexed(t, `"hell   o   " "  wor   ld"`, token.STRING, token.STRING, token.SEMI, token.EOF)
	checkLexed(t, `"\"hello\" \"world\""`, token.STRING, token.SEMI, token.EOF)
}
