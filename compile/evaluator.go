package compile

import (
	"fmt"
	"strconv"

	"github.com/golang/go/src/go/types"
	"github.com/pdk/gosh/token"
	"github.com/pdk/gosh/u"
)

// Evaluator is a function that can be evaluated. It may return some Values
// and/or an error.
type Evaluator func(*Variables) ([]Value, error)

// evaluatorProducer produces Evaluators for Nodes.
type evaluatorProducer func(*Node) Evaluator

var nodeEvaluator [token.TransformResultsEnd]evaluatorProducer

func init() {
	nodeEvaluator = [token.TransformResultsEnd]evaluatorProducer{
		token.IDENT:     VariableLookup,
		token.ASSIGN:    AssignValues,
		token.INT:       IntegerLiteral,
		token.STRING:    StringLiteral,
		token.PLUS:      AdditionOperator,
		token.COMMA:     MultiValueOperator,
		token.LPAREN:    MultiValueOperator,
		token.TRUE:      TrueLiteral,
		token.FALSE:     FalseLiteral,
		token.NIL:       NilLiteral,
		token.FUNC:      FuncDefinition,
		token.FUNCAPPLY: FuncApplication,
		token.STMTS:     StatementsEvaluator,
		token.EXTERN:    Noop,
	}
}

// Evaluator converts a parsed & analyzed node into an evaluator.
func (n *Node) Evaluator() Evaluator {

	producer := nodeEvaluator[n.Token()]

	if producer == nil {
		return n.UnknownOperation
	}

	return producer(n)
}

// FuncApplication applies a function to arguments.
func FuncApplication(n *Node) Evaluator {

	funcResolver := LeftEval(n)
	var paramEvals []Evaluator

	for _, child := range n.children[1:] {
		eval := child.Evaluator()
		// if err != nil {
		// 	return func(vars *Variables) ([]Value, error) {
		// 		return []Value{}, fmt.Errorf("unable to construct evaluator: %s", err)
		// 	}
		// }

		paramEvals = append(paramEvals, eval)
	}

	return func(vars *Variables) ([]Value, error) {

		fr, err := funcResolver(vars)
		if err != nil {
			return []Value{}, fmt.Errorf("unable to resolve function: %s", err)
		}

		if len(fr) != 1 || !fr[0].isFunction {
			return []Value{}, fmt.Errorf("cannot apply a non-function")
		}

		f := fr[0].function

		values := []Value{}
		for _, eachEval := range paramEvals {
			val, err := eachEval(vars)
			if err != nil {
				return []Value{}, err
			}
			values = append(values, val...)
		}

		if len(f.parameters) != len(values) {
			return []Value{}, fmt.Errorf("number of arguments does not match number of parameters")
		}

		scope := NewScope(vars)

		for n, v := range f.captured.values {
			scope.SetRef(n, v)
		}

		for _, l := range f.locals {
			scope.Set(l, Nil())
		}

		for i, p := range f.parameters {
			scope.Set(p, values[i])
		}

		return f.body(scope)
	}
}

// FuncDefinition returns a function.
func FuncDefinition(n *Node) Evaluator {

	return func(vars *Variables) ([]Value, error) {

		f := Function{
			parameters: n.analysis.parameters,
			channels:   n.analysis.channels,
			locals:     u.KeysOf(n.analysis.locals),
			body:       n.analysis.body.Evaluator(),
			captured:   NewScope(vars),
		}

		for _, v := range n.analysis.FreeVariables() {
			ref, err := vars.Reference(v)
			if err != nil {
				return []Value{}, fmt.Errorf("unable to capture free variable %s: %s", v, err)
			}
			f.captured.SetRef(v, ref)
		}

		return []Value{FunctionValue(f)}, nil
	}
}

// Noop does nothing and returns no values.
func Noop(n *Node) Evaluator {
	return func(vars *Variables) ([]Value, error) {
		return []Value{}, nil
	}
}

// NilLiteral returns nil.
func NilLiteral(n *Node) Evaluator {
	return func(vars *Variables) ([]Value, error) {
		return []Value{
			Value{
				isNil: true,
			},
		}, nil
	}
}

// TrueLiteral returns true.
func TrueLiteral(n *Node) Evaluator {

	return func(vars *Variables) ([]Value, error) {
		return []Value{
			Value{
				isBasicKind: true,
				basicKind:   types.Bool,
				basicValue:  true,
			},
		}, nil
	}
}

// FalseLiteral returns false.
func FalseLiteral(n *Node) Evaluator {

	return func(vars *Variables) ([]Value, error) {
		return []Value{
			Value{
				isBasicKind: true,
				basicKind:   types.Bool,
				basicValue:  false,
			},
		}, nil
	}
}

// IntegerLiteral returns the value of an integer.
func IntegerLiteral(n *Node) Evaluator {

	i, err := strconv.ParseInt(n.lexeme.Literal(), 10, 0)

	if err != nil {
		n.lexeme.PrintError("Error converting int literal %s: %s", n.Literal(), err)
	}

	return func(vars *Variables) ([]Value, error) {
		return []Value{
			Value{
				isBasicKind: true,
				basicKind:   types.Int64,
				basicValue:  i,
			},
		}, err
	}
}

// StringLiteral returns the value of a string literal
func StringLiteral(n *Node) Evaluator {

	s := n.lexeme.Literal()

	return func(vars *Variables) ([]Value, error) {
		return []Value{
			Value{
				isBasicKind: true,
				basicKind:   types.String,
				basicValue:  s,
			},
		}, nil
	}
}

// StatementsEvaluator evaluates a series of expressions, returning the value of the last expression.
func StatementsEvaluator(n *Node) Evaluator {

	var evaluators []Evaluator

	for _, child := range n.children {
		eval := child.Evaluator()
		// if err != nil {
		// 	return func(vars *Variables) ([]Value, error) {
		// 		return []Value{}, fmt.Errorf("unable to construct evaluator: %s", err)
		// 	}
		// }

		evaluators = append(evaluators, eval)
	}

	return func(vars *Variables) ([]Value, error) {

		var results []Value

		for _, e := range evaluators {
			r, err := e(vars)
			if err != nil {
				return []Value{}, err
			}

			results = r
		}

		return results, nil
	}
}

// MultiValueOperator produces multiple values, one per child.
func MultiValueOperator(n *Node) Evaluator {

	var evaluators []Evaluator

	for _, child := range n.children {
		eval := child.Evaluator()
		// if err != nil {
		// 	return func(vars *Variables) ([]Value, error) {
		// 		return []Value{}, fmt.Errorf("unable to construct evaluator: %s", err)
		// 	}
		// }

		evaluators = append(evaluators, eval)
	}

	return func(vars *Variables) ([]Value, error) {

		var results []Value

		for _, e := range evaluators {
			r, err := e(vars)
			if err != nil {
				return []Value{}, err
			}

			results = append(results, r...)
		}

		return results, nil
	}
}

// VariableLookup looks up and returns the value of a variable.
func VariableLookup(n *Node) Evaluator {

	varName := n.Literal()

	return func(vars *Variables) ([]Value, error) {

		v, err := vars.Value(varName)

		if err != nil {
			return []Value{}, err
		}

		return []Value{v}, nil
	}
}

// AssignValues evaluates the right-hand side and sets variables on the left-hand side.
func AssignValues(n *Node) Evaluator {

	badVars := func(vars *Variables) ([]Value, error) {
		return []Value{}, fmt.Errorf("left-hand side of assignment must be one or more identifiers")
	}

	var varNames []string

	lhs := n.children[0]
	switch lhs.Token() {
	case token.IDENT:
		varNames = append(varNames, lhs.Literal())
	case token.COMMA:
		for _, v := range lhs.children {
			if !v.IsToken(token.IDENT) {
				return badVars
			}
			varNames = append(varNames, v.Literal())
		}
	default:
		return badVars
	}

	right := RightEval(n)

	return func(vars *Variables) ([]Value, error) {

		r, err := right(vars)
		if err != nil {
			return []Value{}, err
		}

		if len(varNames) != len(r) {
			return []Value{}, fmt.Errorf("count of variables on left does not match number of results on right side")
		}

		for i, n := range varNames {
			_, err := vars.Set(n, r[i])
			if err != nil {
				return []Value{}, err
			}
		}

		return r, nil
	}
}

// LeftEval returns the evaluator of the first child.
func LeftEval(n *Node) Evaluator {
	return n.children[0].Evaluator()
}

// RightEval returns the evaluator of the second child.
func RightEval(n *Node) Evaluator {
	return n.children[1].Evaluator()
}

// AdditionOperator returns the value of an addition operation.
func AdditionOperator(n *Node) Evaluator {

	left := LeftEval(n)
	right := RightEval(n)

	return func(vars *Variables) ([]Value, error) {

		r1, err1 := left(vars)
		if err1 != nil {
			return []Value{}, err1
		}

		r2, err2 := right(vars)
		if err2 != nil {
			return []Value{}, err2
		}

		if r1[0].isBasicKind && r1[0].basicKind == types.Int64 &&
			r2[0].isBasicKind && r2[0].basicKind == types.Int64 {

			i1 := r1[0].basicValue.(int64)
			i2 := r2[0].basicValue.(int64)

			return []Value{
				Value{
					isBasicKind: true,
					basicKind:   types.Int64,
					basicValue:  i1 + i2,
				},
			}, nil
		}

		if r1[0].isBasicKind && r1[0].basicKind == types.String &&
			r2[0].isBasicKind && r2[0].basicKind == types.String {

			s1 := r1[0].basicValue.(string)
			s2 := r2[0].basicValue.(string)

			return []Value{
				Value{
					isBasicKind: true,
					basicKind:   types.String,
					basicValue:  s1 + s2,
				},
			}, nil
		}

		return n.UnknownOperation(vars)
	}
}

// UnknownOperation returns an error saying we don't know what to do with this
// node.
func (n *Node) UnknownOperation(vars *Variables) ([]Value, error) {
	return []Value{},
		fmt.Errorf("unknown operation %s/%s on line %d col %d",
			n.Token(), n.lexeme.Literal(), n.lexeme.LineNo(), n.lexeme.CharNo())
}
