package compile

import (
	"fmt"
	"log"
	"strings"

	"github.com/pdk/gosh/lexer"
	"github.com/pdk/gosh/parse"
	"github.com/pdk/gosh/token"
	"github.com/pdk/gosh/u"
)

// Node is a node in a "compile" processing tree.
type Node struct {
	lexeme   *lexer.Lexeme
	children []*Node
	arity    parse.Arity
	analysis *Analysis
}

// Analysis returns the analysis of the node.
func (n *Node) Analysis() *Analysis {
	return n.analysis
}

// Analysis contains info about what's been parsed.
type Analysis struct {
	identifiers map[string]bool
	parameters  []string
	channels    []string
	locals      map[string]bool
	externs     map[string]bool
	body        *Node
	parent      *Analysis
}

// NewAnalysis returns a new Analysis.
func NewAnalysis() *Analysis {
	return &Analysis{
		identifiers: make(map[string]bool),
		locals:      make(map[string]bool),
		externs:     make(map[string]bool),
	}
}

// Print prints out the results of an analysis.
func (a *Analysis) Print() {
	fmt.Printf("identifiers: %s\n", strings.Join(names(a.identifiers), ", "))
	fmt.Printf("parameters: %s\n", strings.Join(a.parameters, ", "))
	fmt.Printf("channels: %s\n", strings.Join(a.channels, ", "))
	fmt.Printf("locals: %s\n", strings.Join(names(a.locals), ", "))
	fmt.Printf("free: %s\n", strings.Join(a.FreeVariables(), ", "))
	fmt.Printf("externs: %s\n", strings.Join(names(a.externs), ", "))
	fmt.Printf("unbound: %s\n", strings.Join(a.MissingBinding(), ", "))
}

// FreeVariables returns a slice of identifiers which are not local, and are not
// params/local channels. I.e. the names that must be resolved outside the scope
// of the given scope (function).
func (a *Analysis) FreeVariables() []string {

	var free []string

	free = append(free, names(a.externs)...)

	for id := range a.identifiers {
		if !u.StringIn(id, free) &&
			!u.StringIn(id, a.parameters) &&
			!u.StringIn(id, a.channels) &&
			!a.locals[id] {

			free = append(free, id)
		}
	}

	return free
}

// BoundInAncestor recurses the scope chain looking for a binding.
func (a *Analysis) BoundInAncestor(v string) bool {

	if a == nil {
		return false
	}

	if u.StringIn(v, a.parameters) || u.StringIn(v, a.channels) || a.locals[v] {
		return true
	}

	return a.parent.BoundInAncestor(v)
}

// MissingBinding identifies free variables that are not bound in any ancestor.
func (a *Analysis) MissingBinding() []string {

	free := a.FreeVariables()
	var unbound []string

	for _, v := range free {
		if !a.BoundInAncestor(v) {
			unbound = append(unbound, v)
		}
	}

	return unbound
}

func names(m map[string]bool) []string {
	var x []string

	for n := range m {
		x = append(x, n)
	}

	return x
}

// ConvertParseToCompile converts a parse tree to a compile tree in preparation
// for analysis.
func ConvertParseToCompile(ast *parse.Node) *Node {

	node := &Node{
		lexeme: ast.Lexeme(),
		arity:  ast.Arity(),
	}

	for _, child := range ast.Children() {
		node.children = append(node.children, ConvertParseToCompile(child))
	}

	return node
}

// Error returns an error associated with a particular node.
func (n *Node) Error(mesg string, args ...interface{}) error {
	return n.lexeme.Error(mesg, args...)
}

// Token returns the token of the lexeme of the node.
func (n *Node) Token() token.Token {
	return n.lexeme.Token()
}

// Literal returns the literal value of the lexeme of the node.
func (n *Node) Literal() string {
	return n.lexeme.Literal()
}

// IsToken checks if the token of the lexeme of the node is any of the given tokens.
func (n *Node) IsToken(toks ...token.Token) bool {

	for _, tok := range toks {
		if n.Token() == tok {
			return true
		}
	}

	return false
}

// AssertIsToken returns an error if the token is not one of the specified
// options.
func (n *Node) AssertIsToken(toks ...token.Token) error {
	if !n.IsToken(toks...) {
		return n.Error("expected to find (one of) %s", toks)
	}

	return nil
}

// addStrings puts all the names in a slice into a map.
func addStrings(m map[string]bool, names []string) {

	for _, n := range names {
		m[n] = true
	}
}

// FuncAnalysis gathers info about a func/methods variables/identifiers.
func (n *Node) FuncAnalysis(parent *Analysis) (bool, error) {

	if !n.IsToken(token.FUNC) {
		return false, nil
	}

	if len(n.children) != 3 {
		log.Printf("error in parse of func/method, expected 3 children, got %d", len(n.children))
	}

	// start a new collector, attach to current node.
	// identify parameters, channels
	// recurse collecting names

	collector := NewAnalysis()
	collector.parent = parent
	n.analysis = collector

	var err error
	collector.parameters, err = n.children[0].Idents()
	if err != nil {
		return true, err
	}

	collector.channels, err = n.children[1].Idents()
	if err != nil {
		return true, err
	}

	collector.body = n.children[2]

	collector.body.ScopeAnalysis(collector)

	for e := range collector.externs {
		delete(collector.locals, e)
	}

	return true, nil
}

// MethApplyAnalysis checks if the node is a method-apply, and handles if so.
// Return true if handled, false if not.
func (n *Node) MethApplyAnalysis(collector *Analysis) (bool, error) {

	if !n.IsToken(token.METHAPPLY) {
		return false, nil
	}

	// first child is obj
	ident, err := n.children[0].PrimaryIdent()
	if err != nil {
		return true, err
	}

	collector.identifiers[ident] = true
	// second child is meth name. skip
	// third and subsequent are expressions to eval as params
	for _, each := range n.children[2:] {
		each.ScopeAnalysis(collector)
	}

	return true, nil
}

// AssignAnalysis collects identifiers on the LHS of any assignment.
func (n *Node) AssignAnalysis(collector *Analysis) error {

	if !n.IsToken(token.ASSIGN, token.QASSIGN, token.ACCUM) {
		return nil
	}

	// identify left-most identifiers of nodes on left side as local
	// variables.

	leftHandSide := n.children[0]

	if leftHandSide.IsToken(token.COMMA, token.LPAREN) {
		for _, each := range leftHandSide.children {
			id, err := each.PrimaryIdent()
			if err != nil {
				return err
			}
			collector.locals[id] = true
		}

		return nil
	}

	id, err := leftHandSide.PrimaryIdent()
	if err != nil {
		return err
	}

	collector.locals[id] = true
	return nil
}

// ScopeAnalysis crawls the tree and identifies identifiers to find free
// variables and unknown idents.
func (n *Node) ScopeAnalysis(collector *Analysis) error {

	done, err := n.FuncAnalysis(collector)
	if done || err != nil {
		return err
	}

	done, err = n.MethApplyAnalysis(collector)
	if done || err != nil {
		return err
	}

	n.AssignAnalysis(collector)

	if n.IsToken(token.EXTERN) {
		for _, child := range n.children {
			if child.IsToken(token.IDENT) {
				collector.externs[child.Literal()] = true
			}
		}
	}

	if n.IsToken(token.IDENT) {
		collector.identifiers[n.Literal()] = true
	}

	if n.IsToken(token.PERIOD) {
		// ignore names on the right of the dot.
		n.children[0].ScopeAnalysis(collector)
		return nil
	}

	for _, c := range n.children {
		c.ScopeAnalysis(collector)
	}

	return nil
}

// Idents returns a slice of strings of the idents' literal values.
func (n *Node) Idents() ([]string, error) {

	err := n.AssertIsToken(token.IDENT, token.COMMA, token.LSQR, token.LPAREN)
	if err != nil {
		return []string{}, err
	}

	if n.IsToken(token.IDENT) {
		return []string{n.Literal()}, nil
	}

	var ids []string
	for _, ch := range n.children {
		idents, err := ch.Idents()
		if err != nil {
			return ids, err
		}

		ids = append(ids, idents...)
	}

	return ids, nil
}

// PrimaryIdent returns the primary identifier name of an expression.
// a[b] => a
// a.b.c.x[23] => a
func (n *Node) PrimaryIdent() (string, error) {

	err := n.AssertIsToken(token.IDENT, token.PERIOD, token.LSQR)
	if err != nil {
		return "", nil
	}

	if n.IsToken(token.IDENT) {
		return n.lexeme.Literal(), nil
	}

	return n.children[0].PrimaryIdent()
}

// AllFuncs finds all the func declaration nodes in the parse tree and returns
// them as a slice.
func (n *Node) AllFuncs() []*Node {

	if n.IsToken(token.FUNC) {
		return []*Node{n}
	}

	if len(n.children) == 0 {
		return []*Node{}
	}

	var x []*Node
	for _, c := range n.children {
		x = append(x, c.AllFuncs()...)
	}

	return x
}
