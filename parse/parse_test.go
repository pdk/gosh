package parse_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pdk/gosh/lexer"
	"github.com/pdk/gosh/parse"
	"github.com/pdk/gosh/reader"
)

func TestBinaries(t *testing.T) {

	checkParse(t, "1+2", tr("+", "1", "2"), "simple add")
	checkParse(t, "3*4", tr("*", "3", "4"), "simple mult")
	checkParse(t, "5-6", tr("-", "5", "6"), "simple subtract")
	checkParse(t, "7/8", tr("/", "7", "8"), "simple div")
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

func checkSexpr(t *testing.T, input, expected, preface string) {

	result := parseInput(input).Sexpr()

	if result != expected {
		t.Errorf("%s: expected %s but got %s", preface, expected, result)
	}
}

func parseInput(input string) *parse.Node {

	lines := reader.ReadLinesToStrings(strings.NewReader(input))
	lxr := lexer.New(lines)
	parser := parse.New(lxr)

	return parser.Parse()
}

func checkParse(t *testing.T, input string, exp tree, preface string) {
	checkMatch(t, parseInput(input), exp, preface+": "+input)
}

func checkMatch(t *testing.T, ast *parse.Node, exp tree, preface string) {

	f := ast.Value().Literal()
	e := exp.value

	if f != e {
		t.Errorf("%s: expected %s but found %s", preface, e, f)
	}

	for i, c := range exp.children {
		checkMatch(t, ast.Children()[i], c, preface)
	}
}

type tree struct {
	value    string
	children []tree
}

func tr(v string, kids ...interface{}) tree {

	n := tree{
		value: v,
	}

	for _, k := range kids {

		sval, ok := k.(string)
		if ok {
			n.children = append(n.children, tr(sval))
			continue
		}

		tval, ok := k.(tree)
		if ok {
			n.children = append(n.children, tval)
			continue
		}

		panic(fmt.Sprintf("unhandled type %T in tr()", k))
	}

	return n
}
