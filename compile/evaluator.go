package compile

import (
	"strconv"

	"github.com/pdk/gosh/token"
	"github.com/pdk/gosh/u"
)

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
		token.IF:         ConditionalOperator,
		token.WHILE:      LoopOperator,
		token.RETURN:     ReturnOperator,
		token.BREAK:      BreakOperator,
		token.CONTINUE:   ContinueOperator,
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

		if len(fr) != 1 {
			return Values(), n.Error("cannot apply multiple values as a function")
		}

		f, ok := fr[0].(Function)
		if !ok {
			return Values(), n.Error("cannot apply a non-function")
		}

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
			scope.Set(l, nil)
		}

		for i, p := range f.parameters {
			scope.Set(p, values[i])
		}

		result, err := f.body(scope)

		if IsControlValue(result) {
			return WrappedValues(result), err
		}

		return result, err
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

		return Values(f), nil
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
		return Values(nil), nil
	}

	return e, nil
}

// TrueLiteral returns true.
func TrueLiteral(n *Node) (Evaluator, error) {

	e := func(vars *Variables) ([]Value, error) {
		return Values(true), nil
	}

	return e, nil
}

// FalseLiteral returns false.
func FalseLiteral(n *Node) (Evaluator, error) {

	e := func(vars *Variables) ([]Value, error) {
		return Values(false), nil
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
		return Values(i), nil
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
		return Values(f), nil
	}

	return e, nil
}

// StringLiteral returns the value of a string literal
func StringLiteral(n *Node) (Evaluator, error) {

	s := n.Literal()

	e := func(vars *Variables) ([]Value, error) {
		return Values(s), nil
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

			if IsControlValue(r) {
				return r, nil
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

	return Values(), n.Error("expected single value but got %d", len(vals))
}

// StandardSingleEval evaluates and Evaluator and expects a single result,
// otherwise returns an error.
func StandardSingleEval(n *Node, eval Evaluator, vars *Variables) (Value, error) {

	r, err := eval(vars)
	if err != nil {
		return nil, err
	}

	return SingleValue(n, r)
}

// StandardBinaryEval evaluates a left side, then a right side. If either
// produces an error, or if either does not produce a single value result, then
// an error is returned.
func StandardBinaryEval(n *Node, left, right Evaluator, vars *Variables) (Value, Value, error) {

	leftVal, err := StandardSingleEval(n, left, vars)
	if err != nil {
		return leftVal, nil, err
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

// LeftRightEvaluators produces evaluators for binary operation (two child nodes).
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

		r, ok := TryBinaryStringOp(leftVal, rightVal, func(s1, s2 string) string {
			return s1 + s2
		})
		if ok {
			return Values(r), nil
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

// TryBinaryStringOp checks if the two values are strings, and then applies the
// op if so. Returns false if either value is not a string.
func TryBinaryStringOp(left, right Value, op func(string, string) string) (string, bool) {

	lStr, ok := left.(string)
	if !ok {
		return "", false
	}

	rStr, ok := right.(string)
	if !ok {
		return "", false
	}

	return op(lStr, rStr), true
}

// TryBinaryInt64Op checks if the two values are int64s, and then applies the
// op if so. Returns false if either value is not an int64.
func TryBinaryInt64Op(left, right Value, op func(int64, int64) int64) (int64, bool) {

	lInt, ok := left.(int64)
	if !ok {
		return 0, false
	}

	rInt, ok := right.(int64)
	if !ok {
		return 0, false
	}

	return op(lInt, rInt), true
}

// TryBinaryFloat64Op checks if the two values are float64s, and then applies the
// op if so. Returns false if either value is not a float64.
func TryBinaryFloat64Op(left, right Value, op func(float64, float64) float64) (float64, bool) {

	lFloat, ok := left.(float64)
	if !ok {
		return 0, false
	}

	rFloat, ok := right.(float64)
	if !ok {
		return 0, false
	}

	return op(lFloat, rFloat), true
}

// BinaryIntOperation evaluates left and right sides, confirms they are int64
// results, and produces a new result by applying the given function.
func BinaryIntOperation(n *Node, left, right Evaluator, vars *Variables, op func(int64, int64) int64) ([]Value, error) {

	leftVal, rightVal, err := StandardBinaryEval(n, left, right, vars)
	if err != nil {
		return Values(), err
	}

	r, ok := TryBinaryInt64Op(leftVal, rightVal, op)
	if ok {
		return Values(r), nil
	}

	return Values(), n.Error("cannot apply %s to %T and %T",
		n.Literal(), leftVal, rightVal)
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

	r, ok := TryBinaryInt64Op(leftVal, rightVal, intOp)
	if ok {
		return Values(r), nil
	}

	r2, ok := TryBinaryFloat64Op(leftVal, rightVal, floatOp)
	if ok {
		return Values(r2), nil
	}

	return Values(), n.Error("cannot apply %s to %T and %T",
		n.Literal(), leftVal, rightVal)
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

		switch v := val.(type) {
		case int64:
			return Values(-v), nil
		case float64:
			return Values(-v), nil
		}

		return Values(), n.Error("cannot apply - (negative) to %T", val)
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

		result, err := comp(leftVal, rightVal)

		return Values(result), n.IfError(err, "%s", err)
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

		return Values(!IsTruthy(val)), nil
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

		if !IsTruthy(leftVal) {
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

		if IsTruthy(leftVal) {
			// short-circuit return on True
			return Values(leftVal), nil
		}

		rightVal, err := StandardSingleEval(n, right, vars)
		return Values(rightVal), err
	}

	return e, nil
}

// LoopOperator handles while ... { ... }
func LoopOperator(n *Node) (Evaluator, error) {

	cond, body, err := LeftRightEvaluators(n)
	if err != nil {
		return nil, err
	}

	e := func(vars *Variables) ([]Value, error) {

		for {

			condVal, err := StandardSingleEval(n, cond, vars)
			if err != nil {
				return Values(condVal), err
			}

			if !IsTruthy(condVal) {
				return Values(condVal), nil
			}

			bodyResults, err := body(vars)
			if err != nil {
				return bodyResults, err
			}

			if IsBreakValue(bodyResults) {
				return WrappedValues(bodyResults), nil
			}

			if IsReturnValue(bodyResults) {
				return bodyResults, nil
			}
		}
	}

	return e, nil
}

// ReturnOperator wraps values in a ReturnValue.
func ReturnOperator(n *Node) (Evaluator, error) {

	eval := func(vars *Variables) ([]Value, error) {
		return Values(), nil
	}
	var err error
	if len(n.children) > 0 {
		eval, err = n.children[0].Evaluator()
	}
	if err != nil {
		return nil, err
	}

	e := func(vars *Variables) ([]Value, error) {
		result, err := eval(vars)
		return ReturnValue(result), err
	}

	return e, nil
}

// BreakOperator returns a BreakValue.
func BreakOperator(n *Node) (Evaluator, error) {

	return func(vars *Variables) ([]Value, error) {
		return BreakValue(), nil
	}, nil
}

// ContinueOperator returns a ContinueValue.
func ContinueOperator(n *Node) (Evaluator, error) {

	return func(vars *Variables) ([]Value, error) {
		return ContinueValue(), nil
	}, nil
}

// ConditionalOperator handle if ... { ... } else ...
func ConditionalOperator(n *Node) (Evaluator, error) {

	// Conditionals looks like
	// if children[0] { children[1] } else if children[2] { children[3] } else { children[4] }
	// so a series of if-thens pairs, followed optionally by a single else-child

	var evals []Evaluator
	for _, child := range n.children {

		one, err := child.Evaluator()
		if err != nil {
			return nil, err
		}

		evals = append(evals, one)
	}

	e := func(vars *Variables) ([]Value, error) {

		var results []Value
		var err error

		for i := 0; i < len(evals); i += 2 {

			results, err = evals[i](vars)
			if i >= len(evals)-1 || err != nil {
				// either err or final else clause
				return results, err
			}

			oneVal, err := SingleValue(n, results)
			if err != nil {
				// conditionals should only have 1 value result
				return results, err
			}

			if IsTruthy(oneVal) {
				// found a truthy conditional, evaluate and return the then-clause
				return evals[i+1](vars)
			}
		}

		return results, nil
	}

	return e, nil
}
