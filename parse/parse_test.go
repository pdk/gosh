package parse_test

import (
	"strings"
	"testing"

	"github.com/pdk/gosh/lexer"
	"github.com/pdk/gosh/parse"
	"github.com/pdk/gosh/reader"
)

func TestBinaries(t *testing.T) {

	checkSexpr(t, "1+2", "(+ 1 2)", "simple add")
	checkSexpr(t, "3*4", "(* 3 4)", "simple mult")
	checkSexpr(t, "5-6", "(- 5 6)", "simple subtract")
	checkSexpr(t, "7/8", "(/ 7 8)", "simple div")
}

func TestSimpleExpr(t *testing.T) {

	checkSexpr(t, "1+2*3-4/5", "(- (+ 1 (* 2 3)) (/ 4 5))", "simple expression")
	checkSexpr(t, "(1+2)*3-4/5", "(- (* (+ 1 2) 3) (/ 4 5))", "simple expression")
	checkSexpr(t, "1+2*(3-4)/5", "(+ 1 (/ (* 2 (- 3 4)) 5))", "simple expression")
}

func TestLogics(t *testing.T) {

	checkSexpr(t, "x := a == b && \"yes\" || \"no\"", "(:= x (|| (&& (== a b) yes) no))", "and/or")
	checkSexpr(t, "x := a == b && (c > d || e != 5)", "(:= x (&& (== a b) (|| (> c d) (!= e 5))))", "and/or")

	checkSexpr(t, "!true", "(! true)", "not")
	checkSexpr(t, "1+-2", "(+ 1 (- 2))", "prefix/infix")
}

func TestDot(t *testing.T) {

	checkSexpr(t, "thing.field", "(. thing field)", "basic dot")
	checkSexpr(t, "thing.field(1)", "(m-apply thing field 1)", "method call")
	checkSexpr(t, "!thing.field(1)", "(! (m-apply thing field 1))", "not dot")
}

func TestFuncInvoke(t *testing.T) {

	checkSexpr(t, "f(1)", "(f-apply f 1)", "invoke function")
	checkSexpr(t, "g ( ) ", "(f-apply g)", "invoke func, no args")
	checkSexpr(t, "h(1,2,3)", "(f-apply h 1 2 3)", "invoke with mult args")
	checkSexpr(t, "h(1,2+3*4,5)", "(f-apply h 1 (+ 2 (* 3 4)) 5)", "invoke with expr args")
	checkSexpr(t, "h(1,2+3*(4-5),6)", "(f-apply h 1 (+ 2 (* 3 (- 4 5))) 6)", "invoke with paren expr args")
	checkSexpr(t, "f(g(h(i(1,2)),3))", "(f-apply f (f-apply g (f-apply h (f-apply i 1 2)) 3))", "embed func calls")
}

func TestMethodInvoke(t *testing.T) {

	checkSexpr(t, "o.f(1)", "(m-apply o f 1)", "invoke method")
	checkSexpr(t, "o . g ( ) ", "(m-apply o g)", "invoke method, no args")
	checkSexpr(t, "o.h(1,2,3)", "(m-apply o h 1 2 3)", "invoke with mult args")
	checkSexpr(t, "o.h(1,2+3*4,5)", "(m-apply o h 1 (+ 2 (* 3 4)) 5)", "invoke with expr args")
	checkSexpr(t, "o.h(1,2+3*(4-5),6)", "(m-apply o h 1 (+ 2 (* 3 (- 4 5))) 6)", "invoke with paren expr args")
	checkSexpr(t, "o.f(g(m.h(i(1,2)),3))", "(m-apply o f (f-apply g (m-apply m h (f-apply i 1 2)) 3))", "embed method calls")
}

func TestString(t *testing.T) {

	checkSexpr(t, `x := "hello" + ", world\n"`, `(:= x (+ hello ", world\n"))`, "string expression")
}

func TestErg(t *testing.T) {

	checkSexpr(t, "x := 1*(2+3)/doit(3)", "(:= x (/ (* 1 (+ 2 3)) (f-apply doit 3)))", "mixed")
}

func TestStatments(t *testing.T) {

	checkSexpr(t, "a := 1\nb := 2\nc:=3", "(stmts (:= a 1) (:= b 2) (:= c 3))", "multiple statements")
	checkSexpr(t, "a==b &&\ndoThis() ||\n doThat()", "(|| (&& (== a b) (f-apply doThis)) (f-apply doThat))", "multiline conditional")
	checkSexpr(t, "a==b &&\n(doThis() ||\n doThat())", "(&& (== a b) (|| (f-apply doThis) (f-apply doThat)))", "multiline conditional")
	checkSexpr(t, "a;b;c;d;e", "(stmts a b c d e)", "many")
	checkSexpr(t, "", "", "none")
	checkSexpr(t, "a", "a", "none")
	checkSexpr(t, "a;b", "(stmts a b)", "inner semi")
	checkSexpr(t, "a;b;", "(stmts a b)", "inner, right semi")
	checkSexpr(t, ";a;b", "(stmts a b)", "left, inner semi")
	checkSexpr(t, ";a;b;", "(stmts a b)", "semi all around")
	checkSexpr(t, ";(((1+2)));", "(+ 1 2)", "parens and semis")
}

func TestIf(t *testing.T) {

	checkSexpr(t, "if true { x:=true }", "(if true (:= x true))", "basic if")
	checkSexpr(t, "if true { x:=true } else { x := false }", "(if true (:= x true) (:= x false))", "basic if-else")
	checkSexpr(t, `
		if true {
			x:=true
		} else if a<b {
			x := false
		}`, "(if true (:= x true) (< a b) (:= x false))", "if-else-if")
	checkSexpr(t, `
		if true {
			x:=true
		} else if a<b {
			x := false
		} else if !true {
			f(g(x))
			y := "fred"
		} else {
			dotThat()
		}`, "(if true (:= x true) (< a b) (:= x false) (! true) (; (f-apply f (f-apply g x)) (:= y fred)) (f-apply dotThat))", "if-else-if")
}

//
// Helpers below
//

func checkSexpr(t *testing.T, input, expected, preface string) {

	result, err := parseInput(input)
	if err != nil {
		t.Errorf("did not expect error for input \"%s\", got: %s", input, err)
		return
	}

	sexpr := result.Sexpr()
	if sexpr != expected {
		t.Errorf("%s: expected %s but got %s", preface, expected, sexpr)
	}
}

func parseInput(input string) (*parse.Node, error) {

	lines := reader.ReadLinesToStrings(strings.NewReader(input))
	lxr := lexer.New(lines)
	parser := parse.New(lxr)

	return parser.Parse()
}
