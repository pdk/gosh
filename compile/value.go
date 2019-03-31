package compile

import (
	"fmt"
	"strconv"
	"strings"

	"reflect"
)

// Value is a value that is the result of an evaluation.
type Value struct {
	value interface{}
}

// Values wraps particular values in a []Value.
func Values(vals ...interface{}) []Value {

	var r []Value

	for _, v := range vals {
		x, ok := v.(Value)
		if ok {
			r = append(r, x)
			continue
		}
		r = append(r, Value{value: v})
	}

	return r
}

// TypeOf returns the reflect.Type of the value.
func (v Value) TypeOf() reflect.Type {
	return reflect.TypeOf(v.value)
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

// IsNil returns true if the value is nil.
func (v Value) IsNil() bool {
	return v.value == nil
}

// IsBool returns true if the value is a bool.
func (v Value) IsBool() bool {
	return v.TypeOf().Kind() == reflect.Bool
}

// Bool returns the value as a bool. Will panic if value is not a bool.
func (v Value) Bool() bool {
	return v.value.(bool)
}

// IsString returns true if the value is a string.
func (v Value) IsString() bool {
	return v.TypeOf().Kind() == reflect.String
}

// AsString returns the value as a string. Will panic if underlying value is not
// a string.
func (v Value) AsString() string {
	return v.value.(string)
}

// IsInt64 returns true if the value is an int64.
func (v Value) IsInt64() bool {
	return v.TypeOf().Kind() == reflect.Int64
}

// Int64 returns the value as an int64. Will panic if value is not int64.
func (v Value) Int64() int64 {
	return v.value.(int64)
}

// IsFloat64 returns true if the value is an float64.
func (v Value) IsFloat64() bool {
	return v.TypeOf().Kind() == reflect.Float64
}

// Float64 returns the value as an float64. Will panic if value is not float64.
func (v Value) Float64() float64 {
	return v.value.(float64)
}

// IsFunction returns true if the value is a Function.
func (v Value) IsFunction() bool {
	return v.TypeOf() == reflect.TypeOf(Function{})
}

// Function returns the value as a Function. Program will panic if called on non-Function.
func (v Value) Function() Function {
	return v.value.(Function)
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

// ControlValue returns the value as a ControlValue. Will panic if not actually a ControlValue.
func (v Value) ControlValue() ControlValue {
	return v.value.(ControlValue)
}

// WrappedValues returns the enclosed []Value of the ControlReturn
func WrappedValues(vals []Value) []Value {
	return vals[0].ControlValue().returnValues
}

// IsControlValue returns true if values has 1 value, and it is a ControlValue.
func IsControlValue(vals []Value) bool {

	if len(vals) != 1 {
		return false
	}

	_, ok := vals[0].value.(ControlValue)

	return ok
}

// IsBreakValue returns true if the first value in the slice is a break value.
func IsBreakValue(vals []Value) bool {
	return IsControlValue(vals) && vals[0].ControlValue().which == ControlBreak
}

// IsContinueValue returns true if the first value in the slice is a continue value.
func IsContinueValue(vals []Value) bool {
	return IsControlValue(vals) && vals[0].ControlValue().which == ControlContinue
}

// IsReturnValue checks if the first Value in a slice is a ControlReturn.
func IsReturnValue(vals []Value) bool {
	return IsControlValue(vals) && vals[0].ControlValue().which == ControlReturn
}

// Set overwrites the value of a value from another value.
func (v *Value) Set(from Value) {
	v.value = from.value
}

func (v Value) String() string {

	if v.IsNil() {
		return "nil"
	}

	if v.IsFunction() {
		return fmt.Sprintf("func(%s)[%s]{...}",
			strings.Join(v.Function().parameters, ", "),
			strings.Join(v.Function().channels, ", "))
	}

	switch v.TypeOf().Kind() {
	case reflect.Bool:
		if v.value.(bool) {
			return "true"
		}
		return "false"
	case reflect.Int64:
		return strconv.FormatInt(v.Int64(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(v.Float64(), 'f', -1, 64)
	default:
		return fmt.Sprintf("%s", v.value)
	}
}

// NilValue returns a nil Value.
func NilValue() Value {
	return Value{value: nil}
}

// TrueValue returns a true Value.
func TrueValue() Value {
	return Value{value: true}
}

// IsTruthy returns true if the value is boolean and true, or if it is non-nil.
func (v Value) IsTruthy() bool {

	if v.IsNil() {
		return false
	}

	if v.IsBool() {
		return v.Bool()
	}

	return true
}

// FalseValue returns a false Value.
func FalseValue() Value {
	return Value{value: false}
}

// EqualValues returns true/false if the two values are equal. If they are of
// different types, return an error.
func EqualValues(left, right Value) (bool, error) {

	if left.TypeOf().Kind() != right.TypeOf().Kind() {
		return false, fmt.Errorf("cannot compare values of different types")
	}

	return left.value == right.value, nil
}

// NotEqualValues returns true/false if the two values are not equal. If they are of
// different types, return an error.
func NotEqualValues(left, right Value) (bool, error) {

	r, err := EqualValues(left, right)

	return !r, err
}

// LessThanEqualValue returns true/false and/or an error.
func LessThanEqualValue(left, right Value) (bool, error) {

	t, err := checkCompareKinds(left, right)
	if err != nil {
		return false, err
	}

	l, r := left.value, right.value

	switch t {
	case reflect.Int64:
		return l.(int64) <= r.(int64), nil
	case reflect.Float64:
		return l.(float64) <= r.(float64), nil
	case reflect.String:
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

	l, r := left.value, right.value

	switch t {
	case reflect.Int64:
		return l.(int64) >= r.(int64), nil
	case reflect.Float64:
		return l.(float64) >= r.(float64), nil
	case reflect.String:
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

	l, r := left.value, right.value

	switch t {
	case reflect.Int64:
		return l.(int64) > r.(int64), nil
	case reflect.Float64:
		return l.(float64) > r.(float64), nil
	case reflect.String:
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

	l, r := left.value, right.value

	switch t {
	case reflect.Int64:
		return l.(int64) < r.(int64), nil
	case reflect.Float64:
		return l.(float64) < r.(float64), nil
	case reflect.String:
		return l.(string) < r.(string), nil
	}

	panic("invalid value comparison")
}

// checkCompareKinds checks if the two operands are types that can be compared.
// Returns the type if so, and error if no.
func checkCompareKinds(left, right Value) (reflect.Kind, error) {

	if left.TypeOf().Kind() != right.TypeOf().Kind() {
		return reflect.Invalid, fmt.Errorf("cannot compare values of different types")
	}

	t := left.TypeOf().Kind()
	if t != reflect.Int64 &&
		t != reflect.Float64 &&
		t != reflect.String {
		// we only know how to compare those types
		return reflect.Invalid, fmt.Errorf("don't know how to compare those types")
	}

	return t, nil
}
