package compile

import (
	"fmt"
)

// Variables is a standard value-by-name store.
type Variables struct {
	values map[string]Value
	parent *Variables
}

// GlobalScope returns a new global scope map.
func GlobalScope() *Variables {
	v := Variables{
		values: make(map[string]Value),
	}
	return &v
}

// NewScope returns a new child variable scope.
func NewScope(parent *Variables) *Variables {
	v := Variables{
		values: make(map[string]Value),
		parent: parent,
	}

	return &v
}

// Value returns the value for the given name.
func (v *Variables) Value(name string) (Value, error) {

	if v == nil {
		return Nil(), fmt.Errorf("attempt to access undefined variable %s", name)
	}

	val, ok := v.values[name]
	if ok {
		return val, nil
	}

	return v.parent.Value(name)
}

// Set will set a value in the variable map.
func (v *Variables) Set(name string, val Value) (Value, error) {

	cur, ok := v.values[name]

	if !ok {
		v.values[name] = val
		return val, nil
	}

	if cur.isNil {
		v.values[name] = val
		return val, nil
	}

	if cur.isBasicKind != val.isBasicKind || cur.basicKind != val.basicKind {
		return Nil(), fmt.Errorf("attempt to convert variable %s from type %d to type %d", name, cur.basicKind, val.basicKind)
	}

	v.values[name] = val
	return val, nil
}
