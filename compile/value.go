package compile

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/go/src/go/types"
)

// Value is a value that is the result of an evaluation.
type Value struct {
	isNil        bool
	isBasicKind  bool
	basicKind    types.BasicKind
	basicValue   interface{}
	isFunction   bool
	function     Function
	isControl    bool
	controlType  ControlType
	returnValues []Value
}

// ControlType indicates kinds of flow-control Values.
type ControlType byte

// Values for ControlType
const (
	ControlNone ControlType = iota
	ControlReturn
	ControlBreak
	ControlContinue
)

// Function is a evaluatable thing.
type Function struct {
	parameters []string
	channels   []string
	locals     []string
	captured   *Variables
	body       Evaluator
}

// BreakValue constructs a ControlBreak Value.
func BreakValue() []Value {
	return Values(Value{
		isControl:   true,
		controlType: ControlBreak,
	})
}

// ContinueValue constructs a ControlContinue Value.
func ContinueValue() []Value {
	return Values(Value{
		isControl:   true,
		controlType: ControlContinue,
	})
}

// ReturnValue constructs a ControlReturn Value.
func ReturnValue(vals []Value) []Value {
	return Values(Value{
		isControl:    true,
		controlType:  ControlReturn,
		returnValues: vals,
	})
}

// WrappedValues returns the enclosed []Value of the ControlReturn
func WrappedValues(vals []Value) []Value {
	return vals[0].returnValues
}

// IsControlValue returns true if the first Value in the slice is a control return.
func IsControlValue(vals []Value) bool {
	return len(vals) > 0 && vals[0].isControl
}

// IsBreakValue returns true if the first value in the slice is a break value.
func IsBreakValue(vals []Value) bool {
	return IsControlValue(vals) && vals[0].controlType == ControlBreak
}

// IsContinueValue returns true if the first value in the slice is a continue value.
func IsContinueValue(vals []Value) bool {
	return IsControlValue(vals) && vals[0].controlType == ControlContinue
}

// IsReturnValue checks if the first Value in a slice is a ControlReturn.
func IsReturnValue(vals []Value) bool {
	return IsControlValue(vals) && vals[0].controlType == ControlReturn
}

// FunctionValue wraps a Function in a Value.
func FunctionValue(f Function) Value {
	return Value{
		isFunction: true,
		function:   f,
	}
}

// Set overwrites the value of a value from another value.
func (v *Value) Set(from Value) {
	v.isNil = from.isNil
	v.isBasicKind = from.isBasicKind
	v.isFunction = from.isFunction
	v.basicKind = from.basicKind
	v.basicValue = from.basicValue
	v.function = from.function
}

func (v Value) String() string {

	if v.isNil {
		return "nil"
	}

	if v.isFunction {
		return fmt.Sprintf("func(%s)[%s]{...}",
			strings.Join(v.function.parameters, ", "),
			strings.Join(v.function.channels, ", "))
	}

	if !v.isBasicKind {
		return "*unprintable*"
	}

	switch v.basicKind {
	case types.Bool:
		if v.basicValue.(bool) {
			return "true"
		}
		return "false"
	case types.Int64:
		return strconv.FormatInt(v.basicValue.(int64), 10)
	case types.Float64:
		return strconv.FormatFloat(v.basicValue.(float64), 'f', -1, 64)
	default:
		return fmt.Sprintf("%s", v.basicValue)
	}
}

// NilValue returns a nil Value.
func NilValue() Value {
	return Value{
		isNil:       true,
		isBasicKind: true,
		basicKind:   types.UntypedNil,
		basicValue:  nil,
	}
}

// TrueValue returns a true Value.
func TrueValue() Value {

	return Value{
		isBasicKind: true,
		basicKind:   types.Bool,
		basicValue:  true,
	}
}

// IsTruthy returns true if the value is boolean and true, or if it is non-nil.
func (v Value) IsTruthy() bool {

	if v.isNil {
		return false
	}

	if v.isBasicKind && v.basicKind == types.Bool {
		return v.basicValue.(bool)
	}

	return true
}

// FalseValue returns a false Value.
func FalseValue() Value {

	return Value{
		isBasicKind: true,
		basicKind:   types.Bool,
		basicValue:  false,
	}
}

// EqualValues returns true/false if the two values are equal. If they are of
// different types, return an error.
func EqualValues(left, right Value) (bool, error) {

	if left.basicKind == types.Bool && right.basicKind == types.Bool {
		return left.basicValue.(bool) == right.basicValue.(bool), nil
	}

	_, err := checkCompareKinds(left, right)
	if err != nil {
		return false, err
	}

	return left.basicValue == right.basicValue, nil
}

// NotEqualValues returns true/false if the two values are not equal. If they are of
// different types, return an error.
func NotEqualValues(left, right Value) (bool, error) {

	if left.basicKind == types.Bool && right.basicKind == types.Bool {
		return left.basicValue.(bool) != right.basicValue.(bool), nil
	}

	_, err := checkCompareKinds(left, right)
	if err != nil {
		return false, err
	}

	return left.basicValue != right.basicValue, nil
}

// LessThanEqualValue returns true/false and/or an error.
func LessThanEqualValue(left, right Value) (bool, error) {

	t, err := checkCompareKinds(left, right)
	if err != nil {
		return false, err
	}

	l, r := left.basicValue, right.basicValue

	switch t {
	case types.Int64:
		return l.(int64) <= r.(int64), nil
	case types.Float64:
		return l.(float64) <= r.(float64), nil
	case types.String:
		return l.(string) <= r.(string), nil
	}

	panic("invalid value comparison")
}

// GreaterThanEqualValue returns true/false and/or an error.
func GreaterThanEqualValue(left, right Value) (bool, error) {

	t, err := checkCompareKinds(left, right)
	if err != nil {
		return false, err
	}

	l, r := left.basicValue, right.basicValue

	switch t {
	case types.Int64:
		return l.(int64) >= r.(int64), nil
	case types.Float64:
		return l.(float64) >= r.(float64), nil
	case types.String:
		return l.(string) >= r.(string), nil
	}

	panic("invalid value comparison")
}

// GreaterThanValue returns true/false and/or an error.
func GreaterThanValue(left, right Value) (bool, error) {

	t, err := checkCompareKinds(left, right)
	if err != nil {
		return false, err
	}

	l, r := left.basicValue, right.basicValue

	switch t {
	case types.Int64:
		return l.(int64) > r.(int64), nil
	case types.Float64:
		return l.(float64) > r.(float64), nil
	case types.String:
		return l.(string) > r.(string), nil
	}

	panic("invalid value comparison")
}

// LessThanValue returns true/false and/or an error.
func LessThanValue(left, right Value) (bool, error) {

	t, err := checkCompareKinds(left, right)
	if err != nil {
		return false, err
	}

	l, r := left.basicValue, right.basicValue

	switch t {
	case types.Int64:
		return l.(int64) < r.(int64), nil
	case types.Float64:
		return l.(float64) < r.(float64), nil
	case types.String:
		return l.(string) < r.(string), nil
	}

	panic("invalid value comparison")
}

// checkCompareKinds checks if the two operands are types that can be compared.
// Returns the type if so, and error if no.
func checkCompareKinds(left, right Value) (types.BasicKind, error) {

	if left.basicKind != right.basicKind {
		return types.UntypedNil, fmt.Errorf("cannot compare values of different types")
	}

	t := left.basicKind
	if t != types.Int64 &&
		t != types.Float64 &&
		t != types.String {
		// we only know how to compare those types
		return types.UntypedNil, fmt.Errorf("don't know how to compare those types")
	}

	return t, nil
}
