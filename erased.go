package attribs

import (
	"reflect"
	"strings"

	"github.com/phonkee/attribs/parser"
)

func newErased(typ reflect.Type) (*definitionErased, error) {
	// now we go over all fields and check which are used
	isPtr := typ.Kind() == reflect.Ptr

	// support for pointers
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// check if given type is struct
	if typ.Kind() != reflect.Struct {
		return nil, ErrNotStruct
	}

	//attr, err := inspect(*new(T), map[reflect.Type]*attr{})
	attr, err := inspect(reflect.Indirect(reflect.New(typ)).Interface(), map[reflect.Type]*attr{})

	if err != nil {
		return nil, err
	}

	return &definitionErased{
		attr:  attr,
		isPtr: isPtr,
		typ:   typ,
	}, nil
}

type definitionErased struct {
	attr  *attr
	isPtr bool
	typ   reflect.Type
}

func (d definitionErased) Parse(input string) (reflect.Value, error) {
	result := reflect.New(d.typ)

	// parse input to attribute tree
	attrs, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		return reflect.Value{}, err
	}
	result = result.Elem()

	// create new value from given parsed attributes
	err = d.attr.Set(result, &parser.Attribute{Attributes: attrs})
	if err != nil {
		return reflect.Value{}, err
	}

	return result, nil
}
