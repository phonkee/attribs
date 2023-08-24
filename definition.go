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
func Must[T any](init func() (Definition[T], error)) Definition[T] {
	result, err := init()
	if err != nil {
		panic(err)
	}
	return result
}

// New analyzes given struct and returns definition. definition can then parse tags and returns values
// If something fails, this function panics
func New[T any](what T) (result Definition[T], _ error) {
	// now we go over all fields and check which are used
	typ := reflect.TypeOf(what)

	if typ.Kind() != reflect.Struct {
		return result, ErrNotStruct
	}

	attr, err := inspect(*new(T), map[reflect.Type]*attr{})

	if err != nil {
		return result, err
	}

	return Definition[T]{
		attr: attr,
	}, nil
}

// Definition defies definition of struct
type Definition[T any] struct {
	// structure attribute instance
	attr *attr
}

// Parse parses string with attributes into given type
func (d Definition[T]) Parse(input string) (T, error) {
	result := new(T)

	// parse input to attribute tree
	attrs, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		return *result, err
	}

	// prepare settable value
	val := reflect.Indirect(reflect.ValueOf(result))

	//// create new value from given parsed attributes
	err = d.attr.Set(val, &parser.Attribute{Attributes: attrs})
	if err != nil {
		return *result, err
	}

	return val.Interface().(T), nil
}
