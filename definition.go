package attribs

import (
	"github.com/phonkee/attribs/parser"
	"reflect"
	"strings"
)

const (
	TagName = "attr"
)

// Must !
func Must[T any](t Definition[T], err error) Definition[T] {
	if err != nil {
		panic(err)
	}
	return t
}

// New analyzes given struct and returns definition. definition can then parse tags and returns values
// If something fails, this function panics
func New[T any](what T) (result Definition[T], _ error) {
	// now we go over all fields and check which are used
	typ := reflect.TypeOf(what)
	isPtr := typ.Kind() == reflect.Ptr

	// support for pointers
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// check if given type is struct
	if typ.Kind() != reflect.Struct {
		return result, ErrNotStruct
	}

	//attr, err := inspect(*new(T), map[reflect.Type]*attr{})
	attr, err := inspect(reflect.Indirect(reflect.New(typ)).Interface(), map[reflect.Type]*attr{})

	if err != nil {
		return result, err
	}

	return Definition[T]{
		attr:  attr,
		isPtr: isPtr,
	}, nil
}

// Definition defies definition of struct
type Definition[T any] struct {
	// structure attribute instance
	attr  *attr
	isPtr bool
}

// Parse parses string with attributes into given type
func (d Definition[T]) Parse(input string) (T, error) {
	typ := reflect.TypeOf(*new(T))
	result := reflect.New(typ)

	// parse input to attribute tree
	attrs, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		return result.Interface().(T), err
	}
	result = result.Elem()

	// create new value from given parsed attributes
	err = d.attr.Set(result, &parser.Attribute{Attributes: attrs})
	if err != nil {
		return result.Interface().(T), err
	}

	return result.Interface().(T), nil
}
