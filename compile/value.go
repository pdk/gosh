package compile

import (
	"fmt"
	"strconv"
	"strings"

	"reflect"
)

// Value is a value that is the result of an evaluation.
type Value = interface{}

// Values wraps particular values in a []Value.
// nasty, sneaky hobbitses
func Values(vals ...interface{}) []Value {
	return vals
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

// ControlValue is a Value which is returned to indicate some flow-of-control
// change.
type ControlValue struct {
	which        ControlType
	returnValues []Value
}

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
	return Values(ControlValue{
		which: ControlBreak,
	})
}

// ContinueValue constructs a ControlContinue Value.
func ContinueValue() []Value {
	return Values(ControlValue{which: ControlContinue})
}

// ReturnValue constructs a ControlReturn Value.
func ReturnValue(vals []Value) []Value {
	return Values(
		ControlValue{
			which:        ControlReturn,
			returnValues: vals,
		},
	)
}

// WrappedValues returns the enclosed []Value of the ControlReturn
func WrappedValues(vals []Value) []Value {
	return vals[0].(ControlValue).returnValues
}

// IsControlValue returns true if values has 1 value, and it is a ControlValue.
func IsControlValue(vals []Value) bool {

	if len(vals) != 1 {
		return false
	}

	_, ok := vals[0].(ControlValue)

	return ok
}

// IsBreakValue returns true if the first value in the slice is a break value.
func IsBreakValue(vals []Value) bool {
	return IsControlValue(vals) && vals[0].(ControlValue).which == ControlBreak
}

// IsContinueValue returns true if the first value in the slice is a continue value.
func IsContinueValue(vals []Value) bool {
	return IsControlValue(vals) && vals[0].(ControlValue).which == ControlContinue
}

// IsReturnValue checks if the first Value in a slice is a ControlReturn.
func IsReturnValue(vals []Value) bool {
	return IsControlValue(vals) && vals[0].(ControlValue).which == ControlReturn
}

// ToString converts a value to a string representation.
func ToString(v Value) string {

	if v == nil {
		return "nil"
	}

	f, ok := v.(Function)
	if ok {
		return fmt.Sprintf("func(%s)[%s]{...}",
			strings.Join(f.parameters, ", "),
			strings.Join(f.channels, ", "))
	}

	switch v2 := v.(type) {
	case bool:
		if v2 {
			return "true"
		}
		return "false"
	case int64:
		return strconv.FormatInt(v2, 10)
	case float64:
		return strconv.FormatFloat(v2, 'f', -1, 64)
	default:
		return fmt.Sprintf("%s", v)
	}
}

// IsTruthy returns true if the value is boolean and true, or if it is non-nil.
func IsTruthy(v interface{}) bool {

	if v == nil {
		return false
	}

	b, ok := v.(bool)
	if ok {
		return b
	}

	// not nil, not a bool, so it's a non-nil value, aka "truthy"
	return true
}

// EqualValues returns true/false if the two values are equal. If they are of
// different types, return an error.
func EqualValues(left, right Value) (bool, error) {

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return false, fmt.Errorf("cannot compare values of different types")
	}

	return left == right, nil
}

// NotEqualValues returns true/false if the two values are not equal. If they are of
// different types, return an error.
func NotEqualValues(left, right Value) (bool, error) {

	r, err := EqualValues(left, right)

	return !r, err
}

// LessThanEqualValue returns true/false and/or an error.
func LessThanEqualValue(left, right Value) (bool, error) {

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return false, fmt.Errorf("cannot compare values of different types, %T and %T", left, right)
	}

	switch lv := left.(type) {
	case int64:
		return lv <= right.(int64), nil
	case float64:
		return lv <= right.(float64), nil
	case string:
		return lv <= right.(string), nil
	}

	return false, fmt.Errorf("don't know how to compare those types: %T, %T", left, right)
}

// GreaterThanEqualValue returns true/false and/or an error.
func GreaterThanEqualValue(left, right Value) (bool, error) {

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return false, fmt.Errorf("cannot compare values of different types, %T and %T", left, right)
	}

	switch lv := left.(type) {
	case int64:
		return lv >= right.(int64), nil
	case float64:
		return lv >= right.(float64), nil
	case string:
		return lv >= right.(string), nil
	}

	return false, fmt.Errorf("don't know how to compare those types: %T, %T", left, right)
}

// GreaterThanValue returns true/false and/or an error.
func GreaterThanValue(left, right Value) (bool, error) {

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return false, fmt.Errorf("cannot compare values of different types, %T and %T", left, right)
	}

	switch lv := left.(type) {
	case int64:
		return lv > right.(int64), nil
	case float64:
		return lv > right.(float64), nil
	case string:
		return lv > right.(string), nil
	}

	return false, fmt.Errorf("don't know how to compare those types: %T, %T", left, right)
}

// LessThanValue returns true/false and/or an error.
func LessThanValue(left, right Value) (bool, error) {

	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return false, fmt.Errorf("cannot compare values of different types, %T and %T", left, right)
	}

	switch lv := left.(type) {
	case int64:
		return lv < right.(int64), nil
	case float64:
		return lv < right.(float64), nil
	case string:
		return lv < right.(string), nil
	}

	return false, fmt.Errorf("don't know how to compare those types: %T, %T", left, right)
}
