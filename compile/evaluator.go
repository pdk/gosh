package compile

import (
	"fmt"
	"strconv"

	"github.com/golang/go/src/go/types"
	"github.com/pdk/gosh/token"
)

// ValueMap provides access to current variable values.
type ValueMap interface {
	Value(string) (Value, error)
	Set(string, Value) (Value, error)
}

// Evaluator is a function that can be evaluated. It may return some Values
// and/or an error.
type Evaluator func(ValueMap) ([]Value, error)

// evaluatorProducer produces Evaluators for Nodes.
type evaluatorProducer func(*Node) Evaluator

var nodeEvaluator [token.TransformResultsEnd]evaluatorProducer

func init() {
	nodeEvaluator = [token.TransformResultsEnd]evaluatorProducer{
		token.IDENT:  VariableLookup,
		token.ASSIGN: AssignValues,
		token.INT:    IntegerLiteral,
		token.STRING: StringLiteral,
		token.PLUS:   AdditionOperator,
		token.COMMA:  MultiValueOperator,
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

// IntegerLiteral returns the value of an integer.
func IntegerLiteral(n *Node) Evaluator {

	i, err := strconv.ParseInt(n.lexeme.Literal(), 10, 0)

	if err != nil {
		n.lexeme.PrintError("Error converting int literal %s: %s", n.Literal(), err)
	}

	return func(vars ValueMap) ([]Value, error) {
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

	return func(vars ValueMap) ([]Value, error) {
		return []Value{
			Value{
				isBasicKind: true,
				basicKind:   types.String,
				basicValue:  s,
			},
		}, nil
	}
}

// MultiValueOperator produces multiple values, one per child.
func MultiValueOperator(n *Node) Evaluator {

	var evaluators []Evaluator

	for _, child := range n.children {
		eval := child.Evaluator()
		// if err != nil {
		// 	return func(vars ValueMap) ([]Value, error) {
		// 		return []Value{}, fmt.Errorf("unable to construct evaluator: %s", err)
		// 	}
		// }

		evaluators = append(evaluators, eval)
	}

	return func(vars ValueMap) ([]Value, error) {

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

	return func(vars ValueMap) ([]Value, error) {

		v, err := vars.Value(varName)

		if err != nil {
			return []Value{}, err
		}

		return []Value{v}, nil
	}
}

// AssignValues evaluates the right-hand side and sets variables on the left-hand side.
func AssignValues(n *Node) Evaluator {

	badVars := func(vars ValueMap) ([]Value, error) {
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

	return func(vars ValueMap) ([]Value, error) {

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

	return func(vars ValueMap) ([]Value, error) {

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
func (n *Node) UnknownOperation(vars ValueMap) ([]Value, error) {
	return []Value{},
		fmt.Errorf("unknown operation %s/%s on line %d col %d",
			n.Token(), n.lexeme.Literal(), n.lexeme.LineNo(), n.lexeme.CharNo())
}
