package compile

import (
	"strconv"

	"github.com/golang/go/src/go/types"
	"github.com/pdk/gosh/token"
	"github.com/pdk/gosh/u"
)

// Values returns a slice of Value.
func Values(v ...Value) []Value {
	return append([]Value{}, v...)
}

// Evaluator is a function that can be evaluated. It may return some Values
// and/or an error.
type Evaluator func(*Variables) ([]Value, error)

// evaluatorProducer produces Evaluators for Nodes.
type evaluatorProducer func(*Node) (Evaluator, error)

var nodeEvaluator [token.TransformResultsEnd]evaluatorProducer

func init() {
	nodeEvaluator = [token.TransformResultsEnd]evaluatorProducer{
		token.IDENT:      VariableLookup,
		token.ASSIGN:     AssignValues,
		token.INT:        IntegerLiteral,
		token.FLOAT:      FloatLiteral,
		token.STRING:     StringLiteral,
		token.PLUS:       AdditionOperator,
		token.MINUS:      SubtractionOperator,
		token.MODULO:     ModuloOperation,
		token.MULT:       MultiplicationOperator,
		token.DIV:        DivisionOperator,
		token.COMMA:      MultiValueOperator,
		token.LPAREN:     MultiValueOperator,
		token.TRUE:       TrueLiteral,
		token.FALSE:      FalseLiteral,
		token.NIL:        NilLiteral,
		token.FUNC:       FuncDefinition,
		token.FUNCAPPLY:  FuncApplication,
		token.STMTS:      StatementsEvaluator,
		token.EXTERN:     Noop,
		token.EQUAL:      EqualOperator,
		token.NOT_EQUAL:  NotEqualOperator,
		token.LESS:       LessThanOperator,
		token.LESS_EQUAL: LessThanEqualOperator,
		token.GRTR:       GreaterThanOperator,
		token.GRTR_EQUAL: GreaterThanEqualOperator,
		token.NOT:        NotOperator,
		token.LOG_AND:    LogicalAndOperator,
		token.LOG_OR:     LogicialOrOperator,
	}
}

// Evaluator converts a parsed & analyzed node into an evaluator.
func (n *Node) Evaluator() (Evaluator, error) {

	producer := nodeEvaluator[n.Token()]

	if producer == nil {
		return nil, n.lexeme.Error("unknown operator %s", n.Literal())
	}

	return producer(n)
}

// FuncApplication applies a function to arguments.
func FuncApplication(n *Node) (Evaluator, error) {

	funcResolver, err := LeftEval(n)
	if err != nil {
		return nil, err
	}

	var paramEvals []Evaluator

	for _, child := range n.children[1:] {
		eval, err := child.Evaluator()
		if err != nil {
			return nil, err
		}

		paramEvals = append(paramEvals, eval)
	}

	e := func(vars *Variables) ([]Value, error) {

		fr, err := funcResolver(vars)
		if err != nil {
			return Values(), n.Error("unable to resolve function: %s", err)
		}

		if len(fr) != 1 || !fr[0].isFunction {
			return Values(), n.Error("cannot apply a non-function")
		}

		f := fr[0].function

		var values []Value
		for _, eachEval := range paramEvals {
			val, err := eachEval(vars)
			if err != nil {
				return Values(), err
			}
			values = append(values, val...)
		}

		if len(f.parameters) != len(values) {
			return Values(), n.Error("number of arguments does not match number of parameters")
		}

		scope := NewScope(vars)

		for n, v := range f.captured.values {
			scope.SetRef(n, v)
		}

		for _, l := range f.locals {
			scope.Set(l, NilValue())
		}

		for i, p := range f.parameters {
			scope.Set(p, values[i])
		}

		return f.body(scope)
	}

	return e, nil
}

// FuncDefinition returns a function.
func FuncDefinition(n *Node) (Evaluator, error) {

	e := func(vars *Variables) ([]Value, error) {

		bodyEval, err := n.analysis.body.Evaluator()
		if err != nil {
			return nil, err
		}

		f := Function{
			parameters: n.analysis.parameters,
			channels:   n.analysis.channels,
			locals:     u.KeysOf(n.analysis.locals),
			body:       bodyEval,
			captured:   NewScope(vars),
		}

		for _, v := range n.analysis.FreeVariables() {
			ref, err := vars.Reference(v)
			if err != nil {
				return Values(), n.Error("unable to capture free variable %s: %s", v, err)
			}
			f.captured.SetRef(v, ref)
		}

		return Values(FunctionValue(f)), nil
	}

	return e, nil
}

// Noop does nothing and returns no values.
func Noop(n *Node) (Evaluator, error) {

	return func(vars *Variables) ([]Value, error) {
		return Values(), nil
	}, nil
}

// NilLiteral returns nil.
func NilLiteral(n *Node) (Evaluator, error) {

	e := func(vars *Variables) ([]Value, error) {
		return []Value{
			Value{
				isNil: true,
			},
		}, nil
	}

	return e, nil
}

// TrueLiteral returns true.
func TrueLiteral(n *Node) (Evaluator, error) {

	e := func(vars *Variables) ([]Value, error) {
		return Values(TrueValue()), nil
	}

	return e, nil
}

// FalseLiteral returns false.
func FalseLiteral(n *Node) (Evaluator, error) {

	e := func(vars *Variables) ([]Value, error) {
		return Values(FalseValue()), nil
	}

	return e, nil
}

// IntegerLiteral returns the value of an integer.
func IntegerLiteral(n *Node) (Evaluator, error) {

	i, err := strconv.ParseInt(n.lexeme.Literal(), 10, 0)

	if err != nil {
		return nil, n.Error("%s", err)
	}

	e := func(vars *Variables) ([]Value, error) {
		return []Value{
			Value{
				isBasicKind: true,
				basicKind:   types.Int64,
				basicValue:  i,
			},
		}, err
	}

	return e, nil
}

// FloatLiteral returns the value of an integer.
func FloatLiteral(n *Node) (Evaluator, error) {

	f, err := strconv.ParseFloat(n.lexeme.Literal(), 64)

	if err != nil {
		return nil, n.Error("%s", err)
	}

	e := func(vars *Variables) ([]Value, error) {
		return []Value{
			Value{
				isBasicKind: true,
				basicKind:   types.Float64,
				basicValue:  f,
			},
		}, err
	}

	return e, nil
}

// StringLiteral returns the value of a string literal
func StringLiteral(n *Node) (Evaluator, error) {

	s := n.Literal()

	e := func(vars *Variables) ([]Value, error) {
		return []Value{
			Value{
				isBasicKind: true,
				basicKind:   types.String,
				basicValue:  s,
			},
		}, nil
	}

	return e, nil
}

// StatementsEvaluator evaluates a series of expressions, returning the value of the last expression.
func StatementsEvaluator(n *Node) (Evaluator, error) {

	var evaluators []Evaluator

	for _, child := range n.children {
		eval, err := child.Evaluator()
		if err != nil {
			return nil, err
		}

		evaluators = append(evaluators, eval)
	}

	e := func(vars *Variables) ([]Value, error) {

		var results []Value

		for _, e := range evaluators {
			r, err := e(vars)
			if err != nil {
				return Values(), err
			}

			results = r
		}

		return results, nil
	}

	return e, nil
}

// MultiValueOperator produces multiple values, one per child.
func MultiValueOperator(n *Node) (Evaluator, error) {

	var evaluators []Evaluator

	for _, child := range n.children {
		eval, err := child.Evaluator()
		if err != nil {
			return nil, err
		}

		evaluators = append(evaluators, eval)
	}

	e := func(vars *Variables) ([]Value, error) {

		var results []Value

		for _, e := range evaluators {
			r, err := e(vars)
			if err != nil {
				return Values(), err
			}

			results = append(results, r...)
		}

		return results, nil
	}

	return e, nil
}

// VariableLookup looks up and returns the value of a variable.
func VariableLookup(n *Node) (Evaluator, error) {

	varName := n.Literal()

	e := func(vars *Variables) ([]Value, error) {

		v, err := vars.Value(varName)

		if err != nil {
			return Values(), err
		}

		return Values(v), nil
	}

	return e, nil
}

// AssignValues evaluates the right-hand side and sets variables on the left-hand side.
func AssignValues(n *Node) (Evaluator, error) {

	var varNames []string

	lhs := n.children[0]
	switch lhs.Token() {
	case token.IDENT:
		varNames = append(varNames, lhs.Literal())
	case token.COMMA:
		for _, v := range lhs.children {
			if !v.IsToken(token.IDENT) {
				return nil, n.Error("left-hand side of assignment must be one or more identifiers")
			}
			varNames = append(varNames, v.Literal())
		}
	default:
		return nil, n.Error("left-hand side of assignment must be one or more identifiers")
	}

	right, err := RightEval(n)
	if err != nil {
		return nil, err
	}

	e := func(vars *Variables) ([]Value, error) {

		r, err := right(vars)
		if err != nil {
			return Values(), err
		}

		if len(varNames) != len(r) {
			return Values(), n.Error("count of variables on left does not match number of results on right side")
		}

		for i, n := range varNames {
			_, err := vars.Set(n, r[i])
			if err != nil {
				return Values(), err
			}
		}

		return r, nil
	}

	return e, nil
}

// SingleValue checks that we're dealing with a single value.
func SingleValue(n *Node, vals []Value) (Value, error) {

	if len(vals) == 1 {
		return vals[0], nil
	}

	return Value{}, n.Error("expected single value but got %d", len(vals))
}

// StandardSingleEval evaluates and Evaluator and expects a single result,
// otherwise returns an error.
func StandardSingleEval(n *Node, eval Evaluator, vars *Variables) (Value, error) {

	r, err := eval(vars)
	if err != nil {
		return Value{}, err
	}

	return SingleValue(n, r)
}

// StandardBinaryEval evaluates a left side, then a right side. If either
// produces an error, or if either does not produce a single value result, then
// an error is returned.
func StandardBinaryEval(n *Node, left, right Evaluator, vars *Variables) (Value, Value, error) {

	leftVal, err := StandardSingleEval(n, left, vars)
	if err != nil {
		return leftVal, Value{}, err
	}

	rightVal, err := StandardSingleEval(n, right, vars)

	return leftVal, rightVal, err
}

// LeftEval returns the evaluator of the first child.
func LeftEval(n *Node) (Evaluator, error) {
	return n.children[0].Evaluator()
}

// RightEval returns the evaluator of the second child.
func RightEval(n *Node) (Evaluator, error) {
	return n.children[1].Evaluator()
}

// LeftRightEvaluators produces evaluators for a standard binary operation (two child nodes).
func LeftRightEvaluators(n *Node) (Evaluator, Evaluator, error) {

	left, err := n.children[0].Evaluator()
	if err != nil {
		return nil, nil, err
	}

	right, err := n.children[1].Evaluator()

	return left, right, err
}

// valueEvaluator is to construct a minimal Evaluator when we've already
// evaluated the result. Don't double-eval an evaluator!
func valueEvaluator(v Value) Evaluator {
	return func(vars *Variables) ([]Value, error) {
		return []Value{v}, nil
	}
}

// AdditionOperator returns the value of an addition operation.
func AdditionOperator(n *Node) (Evaluator, error) {

	left, right, err := LeftRightEvaluators(n)
	if err != nil {
		return nil, err
	}

	e := func(vars *Variables) ([]Value, error) {

		leftVal, rightVal, err := StandardBinaryEval(n, left, right, vars)
		if err != nil {
			return Values(), err
		}

		if leftVal.isBasicKind && leftVal.basicKind == types.String &&
			rightVal.isBasicKind && rightVal.basicKind == types.String {

			s1 := leftVal.basicValue.(string)
			s2 := rightVal.basicValue.(string)

			return []Value{
				Value{
					isBasicKind: true,
					basicKind:   types.String,
					basicValue:  s1 + s2,
				},
			}, nil
		}

		return BinaryNumericOperation(n, valueEvaluator(leftVal), valueEvaluator(rightVal), vars,
			func(a, b int64) int64 {
				return a + b
			}, func(a, b float64) float64 {
				return a + b
			})
	}

	return e, nil
}

// BinaryIntOperation evaluates left and right sides, confirms they are int64
// results, and produces a new result by applying the given function.
func BinaryIntOperation(n *Node, left, right Evaluator, vars *Variables, op func(int64, int64) int64) ([]Value, error) {

	leftVal, rightVal, err := StandardBinaryEval(n, left, right, vars)
	if err != nil {
		return Values(), err
	}

	if leftVal.isBasicKind && leftVal.basicKind == types.Int64 &&
		rightVal.isBasicKind && rightVal.basicKind == types.Int64 {

		i1 := leftVal.basicValue.(int64)
		i2 := rightVal.basicValue.(int64)

		return []Value{
			Value{
				isBasicKind: true,
				basicKind:   types.Int64,
				basicValue:  op(i1, i2),
			},
		}, nil
	}

	return Values(), n.Error("cannot apply %s to these types", n.Literal())
}

// BinaryNumericOperation evaluates left and right sides, confirms they are int64
// results, and produces a new result by applying the given function.
func BinaryNumericOperation(n *Node, left, right Evaluator, vars *Variables,
	intOp func(int64, int64) int64,
	floatOp func(float64, float64) float64) ([]Value, error) {

	leftVal, rightVal, err := StandardBinaryEval(n, left, right, vars)
	if err != nil {
		return Values(), err
	}

	if leftVal.isBasicKind && leftVal.basicKind == types.Int64 &&
		rightVal.isBasicKind && rightVal.basicKind == types.Int64 {

		i1 := leftVal.basicValue.(int64)
		i2 := rightVal.basicValue.(int64)

		return []Value{
			Value{
				isBasicKind: true,
				basicKind:   types.Int64,
				basicValue:  intOp(i1, i2),
			},
		}, nil
	}

	if leftVal.isBasicKind && leftVal.basicKind == types.Float64 &&
		rightVal.isBasicKind && rightVal.basicKind == types.Float64 {

		f1 := leftVal.basicValue.(float64)
		f2 := rightVal.basicValue.(float64)

		return []Value{
			Value{
				isBasicKind: true,
				basicKind:   types.Float64,
				basicValue:  floatOp(f1, f2),
			},
		}, nil
	}

	return Values(), n.Error("cannot apply %s to these types", n.Literal())
}

// SubtractionOperator returns the value of an addition operation.
func SubtractionOperator(n *Node) (Evaluator, error) {

	if len(n.children) == 1 {
		return NegativeOperation(n)
	}

	left, right, err := LeftRightEvaluators(n)
	if err != nil {
		return nil, err
	}

	e := func(vars *Variables) ([]Value, error) {

		return BinaryNumericOperation(n, left, right, vars, func(a, b int64) int64 {
			return a - b
		}, func(a, b float64) float64 {
			return a - b
		})
	}

	return e, nil
}

// MultiplicationOperator returns the value of an addition operation.
func MultiplicationOperator(n *Node) (Evaluator, error) {

	left, right, err := LeftRightEvaluators(n)
	if err != nil {
		return nil, err
	}

	e := func(vars *Variables) ([]Value, error) {

		return BinaryNumericOperation(n, left, right, vars, func(a, b int64) int64 {
			return a * b
		}, func(a, b float64) float64 {
			return a * b
		})
	}

	return e, nil
}

// DivisionOperator returns the value of an addition operation.
func DivisionOperator(n *Node) (Evaluator, error) {

	left, right, err := LeftRightEvaluators(n)
	if err != nil {
		return nil, err
	}

	e := func(vars *Variables) ([]Value, error) {

		return BinaryNumericOperation(n, left, right, vars, func(a, b int64) int64 {
			return a / b
		}, func(a, b float64) float64 {
			return a / b
		})

	}

	return e, nil
}

// ModuloOperation returns the value of an addition operation.
func ModuloOperation(n *Node) (Evaluator, error) {

	left, right, err := LeftRightEvaluators(n)
	if err != nil {
		return nil, err
	}

	e := func(vars *Variables) ([]Value, error) {

		return BinaryIntOperation(n, left, right, vars, func(a, b int64) int64 {
			return a % b
		})
	}

	return e, nil
}

// NegativeOperation handles unary -, return the negative of the value.
func NegativeOperation(n *Node) (Evaluator, error) {

	operand, err := n.children[0].Evaluator()
	if err != nil {
		return nil, err
	}

	e := func(vars *Variables) ([]Value, error) {

		result, err := operand(vars)
		if err != nil {
			return Values(), err
		}

		val, err := SingleValue(n, result)
		if err != nil {
			return Values(), err
		}

		if val.isBasicKind && val.basicKind == types.Int64 {

			i := val.basicValue.(int64)

			return []Value{
				Value{
					isBasicKind: true,
					basicKind:   types.Int64,
					basicValue:  -1 * i,
				},
			}, nil
		}

		if val.isBasicKind && val.basicKind == types.Float64 {

			f := val.basicValue.(float64)

			return []Value{
				Value{
					isBasicKind: true,
					basicKind:   types.Float64,
					basicValue:  -f,
				},
			}, nil
		}

		return Values(), n.Error("cannot apply - (negative) to this type")
	}

	return e, nil
}

// ComparisonOperator applies the given comparison.
func ComparisonOperator(n *Node, comp func(Value, Value) (bool, error)) (Evaluator, error) {

	left, right, err := LeftRightEvaluators(n)
	if err != nil {
		return nil, err
	}

	e := func(vars *Variables) ([]Value, error) {

		leftVal, rightVal, err := StandardBinaryEval(n, left, right, vars)
		if err != nil {
			return Values(), err
		}

		isTrue, err := comp(leftVal, rightVal)

		if err != nil {
			return Values(), n.Error("%s", err)
		}

		if isTrue {
			return Values(TrueValue()), nil
		}

		return Values(FalseValue()), nil
	}

	return e, nil
}

// EqualOperator checks equality.
func EqualOperator(n *Node) (Evaluator, error) {
	return ComparisonOperator(n, EqualValues)
}

// NotEqualOperator checks inequality.
func NotEqualOperator(n *Node) (Evaluator, error) {
	return ComparisonOperator(n, NotEqualValues)
}

// LessThanOperator checks less than.
func LessThanOperator(n *Node) (Evaluator, error) {
	return ComparisonOperator(n, LessThanValue)
}

// LessThanEqualOperator checks greater than.
func LessThanEqualOperator(n *Node) (Evaluator, error) {
	return ComparisonOperator(n, LessThanEqualValue)
}

// GreaterThanOperator checks less than.
func GreaterThanOperator(n *Node) (Evaluator, error) {
	return ComparisonOperator(n, GreaterThanValue)
}

// GreaterThanEqualOperator checks greater than.
func GreaterThanEqualOperator(n *Node) (Evaluator, error) {
	return ComparisonOperator(n, GreaterThanEqualValue)
}

// NotOperator evaluate logical !
func NotOperator(n *Node) (Evaluator, error) {

	operand, err := n.children[0].Evaluator()
	if err != nil {
		return nil, err
	}

	e := func(vars *Variables) ([]Value, error) {

		val, err := StandardSingleEval(n, operand, vars)
		if err != nil {
			return Values(), err
		}

		if val.IsTruthy() {
			return Values(TrueValue()), nil
		}

		return Values(FalseValue()), nil
	}

	return e, nil
}

// LogicalAndOperator handles ... && ...
func LogicalAndOperator(n *Node) (Evaluator, error) {

	left, right, err := LeftRightEvaluators(n)
	if err != nil {
		return nil, err
	}

	e := func(vars *Variables) ([]Value, error) {

		leftVal, err := StandardSingleEval(n, left, vars)
		if err != nil {
			return Values(leftVal), err
		}

		if !leftVal.IsTruthy() {
			// short-circuit return on False
			return Values(leftVal), nil
		}

		rightVal, err := StandardSingleEval(n, right, vars)
		return Values(rightVal), err
	}

	return e, nil
}

// LogicialOrOperator handles ... || ...
func LogicialOrOperator(n *Node) (Evaluator, error) {

	left, right, err := LeftRightEvaluators(n)
	if err != nil {
		return nil, err
	}

	e := func(vars *Variables) ([]Value, error) {

		leftVal, err := StandardSingleEval(n, left, vars)
		if err != nil {
			return Values(leftVal), err
		}

		if leftVal.IsTruthy() {
			// short-circuit return on True
			return Values(leftVal), nil
		}

		rightVal, err := StandardSingleEval(n, right, vars)
		return Values(rightVal), err
	}

	return e, nil
}
