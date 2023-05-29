package attribs

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/phonkee/attribs/parser"
)

const (
	TagName = "attr"
)

// Definition defies definition of struct
type Definition[T any] struct {
	// structure attribute instance
	attr Attribute
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
	err = d.attr.SetValue(val, &parser.Attribute{Attributes: attrs})
	if err != nil {
		return *result, err
	}

	return val.Interface().(T), nil
}

// MustNew defines new definition and panics if something fails
func MustNew[T any](what T) Definition[T] {
	result, err := New[T](what)
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

	if typ.Kind() != reflect.Struct && typ.Kind() != reflect.Ptr {
		return result, ErrNotStruct
	}

	// prepare cache for attributes to support recursive structs
	c := newCache()

	// TODO: add support for recursive structs: initialize cache
	attr, err := define(reflect.TypeOf(what), baseAttribute{}, c)
	if err != nil {
		return result, err
	}

	return Definition[T]{
		attr: attr,
	}, nil
}

// define defines given attribute and returns definition
func define(typ reflect.Type, base baseAttribute, cache Cache) (result Attribute, _ error) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		base.Nullable = true
	}

	// check cache first
	if attr, ok := cache.Get(typ); ok {
		return attr, nil
	}

	switch typ.Kind() {
	case reflect.Struct:
		result = &structAttribute{
			baseAttribute: base,
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var width int
		switch typ.Kind() {
		case reflect.Int:
			width = int(unsafe.Sizeof(int(1))) * 8
		case reflect.Int8:
			width = 8
		case reflect.Int16:
			width = 16
		case reflect.Int32:
			width = 32
		case reflect.Int64:
			width = 64
		}
		result = &intAttribute{
			baseAttribute: base,
			width:         width,
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var width int
		switch typ.Kind() {
		case reflect.Uint:
			width = int(unsafe.Sizeof(uint(1))) * 8
		case reflect.Uint8:
			width = 8
		case reflect.Uint16:
			width = 16
		case reflect.Uint32:
			width = 32
		case reflect.Uint64:
			width = 64
		}
		result = &intAttribute{
			baseAttribute: base,
			width:         width,
			unsigned:      true,
		}
	case reflect.Bool:
		result = &boolAttribute{
			baseAttribute: base,
		}
	case reflect.String:
		result = &stringAttribute{
			baseAttribute: base,
		}
	case reflect.Array, reflect.Slice:
		// now parse array (only - item)
		result = &arrayAttribute{
			baseAttribute: base,
		}
	case reflect.Float32, reflect.Float64:
		var width int
		switch typ.Kind() {
		case reflect.Float32:
			width = 32
		case reflect.Float64:
			width = 64
		}
		result = &floatAttribute{
			baseAttribute: base,
			width:         width,
		}
	default:
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedType, typ.String())
	}

	// set to cache before initialization to support recursive structs
	if err := cache.Set(typ, result); err != nil {
		return nil, err
	}

	// initialize given attribute
	if err := result.Init(typ, cache); err != nil {
		return nil, err
	}

	return result, nil
}
