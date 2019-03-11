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

	checkSexpr(t, `x := a == 1 && "one" || "something else"`,
		`(:= x (|| (&& (== a 1) one) "something else"))`, "alternatives")

	checkSexpr(t, `x := a == b && "yes" || "no"`, "(:= x (|| (&& (== a b) yes) no))", "and/or")
	checkSexpr(t, "x := a == b && (c > d || e != 5)", "(:= x (&& (== a b) (|| (> c d) (!= e 5))))", "and/or")

	checkSexpr(t, "!true", "(! true)", "not")
	checkSexpr(t, "1+-2", "(+ 1 (- 2))", "prefix/infix")

	checkSexpr(t, "true && return 42", `(&& true (return 42))`, "and return 42")
	checkSexpr(t, "err == nil && return err", `(&& (== err nil) (return err))`, "and return err")

	checkSexpr(t, "true || return 42", `(|| true (return 42))`, "and return 42")
	checkSexpr(t, "err == nil || return err", `(|| (== err nil) (return err))`, "and return err")

	checkSexpr(t, `err == nil && "dandy" || "not good"`, `(|| (&& (== err nil) dandy) "not good")`, "and+or")
	checkSexpr(t, `return (err == nil && "dandy" || "not good")`, `(return (|| (&& (== err nil) dandy) "not good"))`, "return and+or")

}

func TestReturn(t *testing.T) {

	checkSexpr(t, "return nil", "(return nil)", "return nil")
	checkSexpr(t, "return", "return", "bare return")
	checkSexpr(t, "return a+b", "(return (+ a b))", "return w/expr")
	checkSexpr(t, "return aleph(23/4)", "(return (f-apply aleph (/ 23 4)))", "return func")

	checkSexpr(t, "return 1,2", `(return 1 2)`, "two-val return")
	checkSexpr(t, "return (1,2)", `(return 1 2)`, "two-val return")
	checkSexpr(t, "return true || false", `(return (|| true false))`, "return or")

	// confusing expression will always return "dandy"
	checkSexpr(t, `err == nil && return "dandy" || return "failure"`,
		`(&& (== err nil) (return (|| dandy (return failure))))`, "and return or return")

	// this does the expected thing
	checkSexpr(t, `return err == nil && "dandy" || "failure"`,
		`(return (|| (&& (== err nil) dandy) failure))`, "return and or")

	checkSexpr(t, `return err, err == nil && "dandy" || "failure"`,
		`(return err (|| (&& (== err nil) dandy) failure))`, "return err, and or")

	checkSexpr(t, `return (err, err == nil && "dandy" || "failure")`,
		`(return err (|| (&& (== err nil) dandy) failure))`, "return paren err, and or")
}

func TestIsa(t *testing.T) {

	checkSexpr(t, `if x isa grape { return true }`, `(if (isa x grape) (return true))`, "isa grape")
	checkSexpr(t, `if x hasa grape { return true }`, `(if (hasa x grape) (return true))`, "isa grape")

	checkSexpr(t, `if x.f(23)+5 isa g("x") { return true }`,
		`(if (isa (+ (m-apply x f 23) 5) (f-apply g x)) (return true))`, "isa grape")
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

	checkSexpr(t, "a\nb\n", "(stmts a b)", "inner, right semi")
	checkSexpr(t, ";a\nb", "(stmts a b)", "left, inner semi")
	checkSexpr(t, ";a;b;", "(stmts a b)", "semi all around")
}

func TestIf(t *testing.T) {

	checkSexpr(t, "if p {}", "(if p stmts)", "empty if")
	checkSexpr(t, "if a!=b{\na:=b # cmt\n}", "(if (!= a b) (:= a b))", "if w/comment")
	checkSexpr(t, "if true { x:=true }", "(if true (:= x true))", "basic if")
	checkSexpr(t, "if true { x:=true } else { x := false }", "(if true (:= x true) (:= x false))", "basic if-else")
	checkSexpr(t, `
		if true {
			x:=true
		} else if a<b {
			x := false
		}`,
		"(if true (:= x true) (< a b) (:= x false))", "if-else-if")
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
		}`,
		"(if true (:= x true) (< a b) (:= x false) (! true) (stmts (f-apply f (f-apply g x)) (:= y fred)) (f-apply dotThat))",
		"if-else-if")
	checkSexpr(t, `
		if true {
			x:=true
			# howdy
		} else if a<b {
			x := false#comments abound
		} else if !true {
			# why
			f(g(x)) # are there
			y := "fred"
			# so many?
		} else {
			dotThat()
			# i wonder
		}`,
		"(if true (:= x true) (< a b) (:= x false) (! true) (stmts (f-apply f (f-apply g x)) (:= y fred)) (f-apply dotThat))",
		"if-else-if + comments")

	checkSexpr(t, `
		v := if a == b {
			f(1)
			#
			g(12)
			"first branch"
		} else if (b>c ){
			blip()
			biff.glop(23, "arf")
			"second branch"
		} else {
			# oink
			a := 23 # ack

			# bo
			bo.go()+1
			"third branch"
		}

		# the value of v will be either "first branch", "second branch" or "third
		# branch" as the value of the if expression is the value of the last evaluated
		# expression.`,
		`(:= v (if (== a b) (stmts (f-apply f 1) (f-apply g 12) "first branch") (> b c) (stmts (f-apply blip) (m-apply biff glop 23 arf) "second branch") (stmts (:= a 23) (+ (m-apply bo go) 1) "third branch")))`,
		"big if")

	checkParseErr(t, `if true {} else 1+2`, "expecting either { or if")
	checkParseErr(t, `if true return`, "expecting LBRACE")
}

func TestEmbedIf(t *testing.T) {

	checkSexpr(t, "if p { if q { if r { oink() }}}", "(if p (if q (if r (f-apply oink))))", "embedded if")
	checkSexpr(t, "if p { if q { if r { oink() }} else { boink() }}",
		"(if p (if q (if r (f-apply oink)) (f-apply boink)))", "embedded if")
}

func TestMultilineIf(t *testing.T) {

	// make sure the if and else blocks don't get collapsed by the stmts-collapser transform.
	checkSexpr(t, "if p {a;b;c} else {d;e;f}", "(if p (stmts a b c) (stmts d e f))", "multi-stmt if-else")
	checkSexpr(t, "if p {a;b;c} else if q {;d;e;f;} else if r {} else {g()}",
		"(if p (stmts a b c) q (stmts d e f) r stmts (f-apply g))", "complex multi-branch if")
}

func TestAssign(t *testing.T) {

	checkSexpr(t, "a := 1", `(:= a 1)`, "basic assign 1")
	checkSexpr(t, "a := 1+2", `(:= a (+ 1 2))`, "basic assign 2")
	checkSexpr(t, "a := a && b || c", `(:= a (|| (&& a b) c))`, "assign logic")

	checkSexpr(t, "a ?= 1", `(?= a 1)`, "if-nil assign")
	checkSexpr(t, "a += 1", `(+= a 1)`, "accum assign")
}

func TestArity(t *testing.T) {

	checkLefty(t, "1+2")
	checkRighty(t, "-1")
	checkLefty(t, "1-2")
	checkLefty(t, "f()")
	checkLefty(t, "f(1,2)")

	checkSexpr(t, "(1,2)", `("(" 1 2)`, "tuple")
	checkRighty(t, "(1,2)")

	checkSexpr(t, "(1,2,3,4)", `("(" 1 2 3 4)`, "fourple")
	checkRighty(t, "(1,2,3,4)")
	checkSexpr(t, "asdf(1,2,3,4)", `(f-apply asdf 1 2 3 4)`, "fourple")
	checkLefty(t, "asdf(1,2,3,4)")
	checkSexpr(t, "as.df(1,2,3,4)", `(m-apply as df 1 2 3 4)`, "fourple")
	checkLefty(t, "as.df(1,2,3,4)")

}

func TestSquareBracket(t *testing.T) {

	checkSexpr(t, "()", `"("`, "empty list")
	checkSexpr(t, "[]", `[`, "empty list")
	checkSexpr(t, "f()", `(f-apply f)`, "empty list")

	checkSexpr(t, "a[]", `([ a)`, "empty list")
	checkLefty(t, "a[]")
	checkRighty(t, "[a]")
	checkLefty(t, "a[b,c]")
	checkRighty(t, "[a,b,c]")

	checkSexpr(t, "a[b,c]", `([ a b c)`, "lefty bracket 3 item")
	checkSexpr(t, "[a,b,c]", `([ a b c)`, "righty bracket 3 item")

	checkSexpr(t, "[a]", `([ a)`, "list of a")
}

func TestFunc(t *testing.T) {

	checkSexpr(t, `func(a,b){return a,b}`, `(func ("(" a b) [ (return a b))`, "ab func")
	checkSexpr(t, `func(a,b) [i,j] {return a,b}`, `(func ("(" a b) ([ i j) (return a b))`, "ab ij func")
	checkSexpr(t, `func(){return 1+2}`, `(func "(" [ (return (+ 1 2)))`, "no arg func")
	checkSexpr(t, `func(){}`, `(func "(" [ stmts)`, "no arg, empty")
	checkSexpr(t, `func()[]{}`, `(func "(" [ stmts)`, "no arg, empty")

	checkParseErr(t, "func{}", "expecting LPAREN")
	checkParseErr(t, "func()", "expecting LBRACE")
	checkParseErr(t, "func foo()", "expecting LPAREN")

	checkSexpr(t, "f := func(a) {\nreturn a+1\n}",
		`(:= f (func a [ (return (+ a 1))))`, "assign func")

	checkSexpr(t, "f := func(a) {\nreturn a+1\n}(33)",
		`(:= f (f-apply (func a [ (return (+ a 1))) 33))`, "assign func")

	checkSexpr(t, "i, j := func(a) {\nreturn a+1,a-1\n}(33)",
		`(:= (, i j) (f-apply (func a [ (return (+ a 1) (- a 1))) 33))`, "assign func")

	checkSexpr(t, "i, j := func(a) {\nreturn (a+1,a-1)\n}(33)",
		`(:= (, i j) (f-apply (func a [ (return (+ a 1) (- a 1))) 33))`, "assign func")

	checkSexpr(t, "1 + func(){return 3}()",
		`(+ 1 (f-apply (func "(" [ (return 3))))`, "1 + func applied")

	// This will cause a runtime error: can't add a function to an integer, but
	// syntactically valid.
	checkSexpr(t, "1 + func(){return 3}",
		`(+ 1 (func "(" [ (return 3)))`, "1 + func")
}

func TestWhile(t *testing.T) {

	checkSexpr(t, "while true {}", `(while true stmts)`, `1 infinite lp`)

	checkSexpr(t, "while a > b && c != d || !e {}",
		`(while (|| (&& (> a b) (!= c d)) (! e)) stmts)`, `while w/logic`)

	checkSexpr(t, `while true { while a+b<10 { if x==3 { return 27 } else { while i-j>0 { return "ack"}}}}`,
		`(while true (while (< (+ a b) 10) (if (== x 3) (return 27) (while (> (- i j) 0) (return ack)))))`,
		`while complicated`)

	checkParseErr(t, `while true if`, "token if: unexpected token")

	checkSexpr(t, `while if true {false} {}`, `(while (if true false) stmts)`, "stmt as expr")
}

//
// Helpers below
//

func checkParseErr(t *testing.T, input, matchErr string) {

	_, err := parseInput(input)
	if err == nil {
		t.Errorf("expected error with \"%s\", but got nil", matchErr)
		return
	}

	if !strings.Contains(err.Error(), matchErr) {
		t.Errorf("expected error with \"%s\", but got %s", matchErr, err)
	}
}

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

func checkLefty(t *testing.T, input string) {

	r, err := parseInput(input)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	if !r.IsLefty() {
		t.Errorf("expected %s to be Lefty, but it's not", input)
	}
}

func checkRighty(t *testing.T, input string) {

	r, err := parseInput(input)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	if !r.IsRighty() {
		t.Errorf("expected %s to be Righty, but it's not", input)
	}
}
