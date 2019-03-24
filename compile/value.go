package compile

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/go/src/go/types"
)

// Value is a value that is the result of an evaluation.
type Value struct {
	isNil       bool
	isBasicKind bool
	isFunction  bool
	basicKind   types.BasicKind
	basicValue  interface{}
	function    Function
}

// Function is a evaluatable thing.
type Function struct {
	parameters []string
	channels   []string
	locals     []string
	captured   *Variables
	body       Evaluator
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
	default:
		return fmt.Sprintf("%s", v.basicValue)
	}
}

// Nil returns a nil Value.
func Nil() Value {
	return Value{
		isNil:       true,
		isBasicKind: true,
		basicKind:   types.UntypedNil,
		basicValue:  nil,
	}
}
