package attribs

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type multiAttribs struct {
	Tag string `attr:"name=tag"`
	ID  string `attr:"name=id"`
}

var (
	dma = Must(New(multiAttribs{}))
)

// NewMulti creates new Multi definition. It analyzes given struct and returns definition. definition can then parse tags and returns values
func NewMulti[T any](multi T) (Multi[T], error) {
	result := Multi[T]{
		definitions: make(map[string]*multiDef),
	}

	typ := reflect.TypeOf(multi)
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return result, errors.New("multi expects a struct")
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldType := field.Type
		tagValue, found := field.Tag.Lookup(TagName)
		if !found {
			continue
		}
		parsed, err := dma.Parse(tagValue)
		if err != nil {
			return result, err
		}
		tag := strings.TrimSpace(parsed.Tag)
		if tag == "" {
			continue
		}

		isPtr := fieldType.Kind() == reflect.Ptr

		for fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		existingMD, ok := result.definitions[tag]
		if ok {
			if existingMD.erased != nil {
				return result, fmt.Errorf("if duplicate, cannot have non id-ed: %q", tag)
			}
		} else {
			result.definitions[tag] = &multiDef{
				index:    i,
				isPtr:    isPtr,
				children: map[string]*definitionErased{},
			}
		}

		de, err := newErased(fieldType)
		if err != nil {
			return result, err
		}

		if parsed.ID != "" {
			result.definitions[parsed.Tag].children[parsed.ID] = de
		} else {
			result.definitions[parsed.Tag].erased = de
		}
	}

	return result, nil
}

type Multi[T any] struct {
	definitions map[string]*multiDef
}

func (m *Multi[T]) ParseStructTag(structTag reflect.StructTag) (*T, error) {
	instance := new(T)

	val := reflect.ValueOf(instance).Elem()

	for tagName, def := range m.definitions {
		tag, ok := structTag.Lookup(tagName)
		if !ok {
			continue
		}

		parsed, err := def.erased.Parse(tag)
		if err != nil {
			return nil, err
		}

		// pointer
		if def.isPtr {
			target := reflect.New(def.erased.typ)
			target.Elem().Set(parsed)
			val.Field(def.index).Set(target)
		} else {
			val.Field(def.index).Set(parsed)
		}
	}

	return instance, nil
}

// ParseStruct parses currently only one level no embed
func (m *Multi[T]) ParseStruct(s any) (map[string]*T, error) {
	typ := reflect.TypeOf(s)
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, errors.New("multi expects a struct")
	}

	result := make(map[string]*T)

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		res, err := m.ParseStructTag(field.Tag)
		if err != nil {
			return nil, err
		}
		result[field.Name] = res
	}

	return result, nil
}

type multiDef struct {
	erased   *definitionErased
	index    int
	isPtr    bool
	children map[string]*definitionErased
}
