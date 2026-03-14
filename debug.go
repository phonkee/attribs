package attribs

import (
	"fmt"
	"reflect"
)

// Debug debugs attribs
func Debug[A any, T any](tagName string, instance T, print ...bool) map[string]A {
	var attrInstance A
	result := map[string]A{}

	var shouldPrint bool
	if len(print) > 0 {
		shouldPrint = print[0]
	}

	d, err := New(attrInstance)
	if err != nil {
		panic(err)
	}

	typ := reflect.TypeOf(instance)
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		panic("expected struct")
	}

	if shouldPrint {
		fmt.Printf("struct: \"%T\", tag name: %q\n", instance, tagName)
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag, ok := field.Tag.Lookup(tagName)
		if !ok {
			continue
		}

		parsed, err := d.Parse(tag)
		if err != nil {
			panicf("failed to parse tag %q: %v", tagName, err)
		}

		if shouldPrint {
			fmt.Printf("  field: %q, tag: %q, parsed: %#v\n", field.Name, tag, parsed)
		}

		result[field.Name] = parsed
	}
	return result
}

func panicf(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}
