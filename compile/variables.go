package compile

import (
	"fmt"
	"reflect"
)

// Variables is a standard value-by-name store.
type Variables struct {
	values map[string]*Value
	parent *Variables
}

// GlobalScope returns a new global scope map.
func GlobalScope() *Variables {
	v := Variables{
		values: make(map[string]*Value),
	}
	return &v
}

// NewScope returns a new child variable scope.
func NewScope(parent *Variables) *Variables {
	v := Variables{
		values: make(map[string]*Value),
		parent: parent,
	}

	return &v
}

// Reference returns a reference (pointer) to a value.
func (v *Variables) Reference(name string) (*Value, error) {

	if v == nil {
		return nil, fmt.Errorf("attempt to access undefined variable %s", name)
	}

	val, ok := v.values[name]
	if ok {
		return val, nil
	}

	return v.parent.Reference(name)
}

// Value returns the value for the given name.
func (v *Variables) Value(name string) (Value, error) {

	if v == nil {
		return nil, fmt.Errorf("attempt to access undefined variable %s", name)
	}

	val, ok := v.values[name]
	if ok {
		return *val, nil
	}

	return v.parent.Value(name)
}

// SetRef sets a variable reference (pointer).
func (v *Variables) SetRef(name string, val *Value) {
	v.values[name] = val
}

// Set will set a value in the variable map.
func (v *Variables) Set(name string, val Value) (Value, error) {

	cur, ok := v.values[name]

	if !ok || *cur == nil {
		v.values[name] = &val
		return val, nil
	}

	if reflect.TypeOf(*cur) != reflect.TypeOf(val) {
		return val, fmt.Errorf("attempt to convert variable %s from type %T to type %T",
			name, *cur, val)
	}

	*cur = val

	return val, nil
}
