package compile

import (
	"fmt"
	"log"
	"strconv"

	"github.com/golang/go/src/go/types"
	"github.com/pdk/gosh/token"
)

// Evaluator is a function that can be evaluated. It may return some Values
// and/or an error.
type Evaluator func() ([]Value, error)

// evaluatorProducer produces Evaluators for Nodes.
type evaluatorProducer func(*Node) Evaluator

// Value is a value that is the result of an evaluation.
type Value struct {
	isBasicKind bool
	basicKind   types.BasicKind
	basicValue  interface{}
}

func (v Value) String() string {

	if !v.isBasicKind {
		return "*unprintable*"
	}

	switch v.basicKind {
	case types.Int64:
		return strconv.FormatInt(v.basicValue.(int64), 10)
	default:
		return fmt.Sprintf("%s", v.basicValue)
	}
}

var nodeEvaluator [token.TransformResultsEnd]evaluatorProducer

func init() {
	nodeEvaluator = [token.TransformResultsEnd]evaluatorProducer{
		token.INT:    IntegerLiteral,
		token.STRING: StringLiteral,
		token.PLUS:   AdditionOperator,
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
		log.Printf("Error converting int literal %s on line %d col %d: %s",
			n.Literal(), n.lexeme.LineNo(), n.lexeme.CharNo(), err)
	}

	return func() ([]Value, error) {
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

	return func() ([]Value, error) {
		return []Value{
			Value{
				isBasicKind: true,
				basicKind:   types.String,
				basicValue:  s,
			},
		}, nil
	}
}

// AdditionOperator returns the value of an addition operation.
func AdditionOperator(n *Node) Evaluator {

	left := n.children[0].Evaluator()
	right := n.children[1].Evaluator()

	return func() ([]Value, error) {

		r1, err1 := left()
		if err1 != nil {
			return []Value{}, err1
		}

		r2, err2 := right()
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

		return n.UnknownOperation()
	}
}

// UnknownOperation returns an error saying we don't know what to do with this
// node.
func (n *Node) UnknownOperation() ([]Value, error) {
	return []Value{},
		fmt.Errorf("unknown operation %s/%s on line %d col %d",
			n.Token(), n.lexeme.Literal(), n.lexeme.LineNo(), n.lexeme.CharNo())
}
