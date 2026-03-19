package attribs

import (
	"fmt"
	"reflect"
)

// Debug debugs attribs
func Debug[A any, T any](tagName string, instance T, ignoreUnknown bool) {
	var attrInstance A

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

	fmt.Printf("struct: \"%T\", tag name: %q\n", instance, tagName)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag, ok := field.Tag.Lookup(tagName)
		if !ok {
			continue
		}

		parsed, err := d.Parse(tag, ignoreUnknown)
		if err != nil {
			panicf("failed to parse tag %q: %v", tagName, err)
		}

		fmt.Printf("  field: %q, tag: %q, parsed: %#v\n", field.Name, tag, parsed)
	}
}

func panicf(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}
