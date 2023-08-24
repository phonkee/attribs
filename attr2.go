package attribs

import (
	"github.com/phonkee/attribs/parser"
	"reflect"
	"strconv"
)

type attrType int

const (
	attrTypeInvalid attrType = iota
	attrTypeInteger
	attrTypeString
	attrTypeFloat
	attrTypeStruct
	attrTypeArray
	attrTypeBoolean
)

// inspect given value and return attribute
func inspect(what any, cache map[reflect.Type]*attr) (*attr, error) {
	val := reflect.ValueOf(what)

	// prepare result
	result := &attr{
		Nullable: val.Kind() == reflect.Pointer,
	}

	// get element from pointer
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	switch val.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		result.Type = attrTypeInteger
		result.Signed = true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		result.Type = attrTypeInteger
	case reflect.Bool:
		result.Type = attrTypeBoolean
	case reflect.String:
		result.Type = attrTypeString
	case reflect.Array, reflect.Slice:
		result.Type = attrTypeArray
		elem, err := inspect(reflect.Indirect(reflect.New(val.Type().Elem())), cache)
		if err != nil {
			return nil, err
		}
		result.Elem = elem
	case reflect.Struct:
		result.Type = attrTypeStruct

		// prepare all props
		result.Properties = make(map[string]*attr)

		// iterate over struct fields
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldType := val.Type().Field(i)

			// skip unexported fields
			// TODO: add support for embedded structs
			if !fieldType.IsExported() || fieldType.Anonymous {
				continue
			}

			var newValue any

			if field.Type().Kind() == reflect.Ptr {
				newValue = reflect.Indirect(reflect.New(field.Type().Elem())).Interface()
			} else {
				newValue = reflect.Indirect(reflect.New(field.Type())).Interface()
			}

			fieldAttr, err := inspect(newValue, cache)
			if err != nil {
				return nil, err
			}

			// parse attribs tag first
			pa, err := parseAttribsTag(fieldType.Tag.Get(TagName))
			if err != nil {
				return nil, err
			}

			// skip disabled fields
			if pa.Disabled {
				continue
			}

			// names and aliases
			fieldAttr.Name = fieldType.Name
			fieldAttr.Alias = pa.Alias
			if fieldAttr.Alias == "" {
				fieldAttr.Alias = fieldAttr.Name
			}

			result.Properties[fieldAttr.Alias] = fieldAttr
		}

	}

	return result, nil
}

// attr implementation
// it holds any supported attribute
type attr struct {
	Name     string
	Alias    string
	Nullable bool
	Type     attrType

	// integer and float types
	//Width int

	// integer type
	Signed bool

	// array/slice
	Elem *attr

	// struct properties
	Properties map[string]*attr
}

// Set sets value to given target from parser.
// it returns error if value cannot be set or parsed attribute is invalid
func (a *attr) Set(target reflect.Value, parsed *parser.Attribute) error {
	// check if pointer is not nil, we need to provide new value
	if target.Kind() == reflect.Ptr && target.IsNil() {
		target.Set(reflect.New(target.Type().Elem()))
	}

	switch a.Type {
	case attrTypeInteger:
		return a.SetInteger(target, parsed)
	case attrTypeString:
		return a.SetString(target, parsed)
	case attrTypeFloat:
		//
	case attrTypeArray:
		//
	case attrTypeBoolean:
		//
	case attrTypeStruct:
		for _, att := range parsed.Attributes {
			prop, ok := a.Properties[att.Name]
			if !ok {
				return parser.NewParseError(att.Position, "unknown attribute %s", att.Name)
			}
			var field reflect.Value
			if target.Kind() == reflect.Ptr {
				field = target.Elem().FieldByName(prop.Name)
			} else {
				field = target.FieldByName(prop.Name)
			}

			// handle pointers
			if target.Type().Kind() == reflect.Ptr {
				field.Set(reflect.New(target.Type().Elem()))
			}

			if err := prop.Set(field, &att); err != nil {
				return err
			}
		}
	default:
		panic("implement me")
	}

	return nil
}

func (a *attr) SetInteger(target reflect.Value, parsed *parser.Attribute) error {
	if parsed.Value == nil || parsed.Value.Number == nil {
		return parser.NewParseError(parsed.Position, "invalid value for %s", parsed.Name)
	}
	if target.Kind() == reflect.Ptr {
		target = target.Elem()
	}

	if a.Signed {
		val, err := strconv.ParseInt(*parsed.Value.Number, 10, 64)
		if err != nil {
			return parser.NewParseError(parsed.Position, "invalid value for %s", parsed.Name)
		}
		target.SetInt(val)
	} else {
		val, err := strconv.ParseUint(*parsed.Value.Number, 10, 64)
		if err != nil {
			return parser.NewParseError(parsed.Position, "invalid value for %s", parsed.Name)
		}
		target.SetUint(val)
	}
	return nil
}

func (a *attr) SetString(target reflect.Value, parsed *parser.Attribute) error {
	if parsed.Value == nil || parsed.Value.String == nil {
		return parser.NewParseError(parsed.Position, "invalid value for %s", parsed.Name)
	}
	if target.Kind() == reflect.Ptr {
		target = target.Elem()
	}
	target.SetString(*parsed.Value.String)

	return nil
}
