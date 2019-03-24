package compile

import (
	"fmt"
	"strconv"

	"github.com/golang/go/src/go/types"
)

// Value is a value that is the result of an evaluation.
type Value struct {
	isNil       bool
	isBasicKind bool
	basicKind   types.BasicKind
	basicValue  interface{}
}

func (v Value) String() string {

	if !v.isBasicKind {
		return "*unprintable*"
	}

	if v.isNil {
		return ""
	}

	switch v.basicKind {
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
