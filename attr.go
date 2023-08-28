package attribs

import (
	"fmt"
	"github.com/phonkee/attribs/parser"
	"reflect"
	"strconv"
)

type attrType string

const (
	attrTypeInvalid attrType = "invalid"
	attrTypeInteger attrType = "integer"
	attrTypeString  attrType = "string"
	attrTypeFloat   attrType = "float"
	attrTypeStruct  attrType = "struct"
	attrTypeArray   attrType = "array"
	attrTypeBoolean attrType = "boolean"
	attrTypeMap     attrType = "map"
	attrTypeAny     attrType = "any" // any type is only supported in map, otherwise is impossible to get this type from inspect (since we pass values)
)

// inspect given value and return attribute
// TODO: cache is not supported yet
func inspect(what any, cache map[reflect.Type]*attr) (*attr, error) {
	if _, ok := what.(reflect.Type); ok {
		panic("passing reflect.Type to inspect is not supported")
	}
	if _, ok := what.(reflect.Value); ok {
		panic("passing reflect.Value to inspect is not supported")
	}

	val := reflect.ValueOf(what)

	// prepare result
	result := &attr{
		Nullable: val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface,
	}

	if val.Kind() == reflect.Ptr && val.IsNil() {
		val.Set(reflect.New(val.Type().Elem()))
	}

	// get element from pointer
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	switch val.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		result.Type = attrTypeInteger
		result.Signed = true
		break
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		result.Type = attrTypeInteger
		break
	case reflect.Float32, reflect.Float64:
		result.Type = attrTypeFloat
		break
	case reflect.Bool:
		result.Type = attrTypeBoolean
		break
	case reflect.String:
		result.Type = attrTypeString
		break
	case reflect.Array, reflect.Slice:
		result.Type = attrTypeArray
		typ := val.Type().Elem()
		var (
			elem *attr
			err  error
		)
		// check for any type
		if typ.Kind() == reflect.Interface {
			elem = anyAttr()
		} else {
			newType := reflect.Indirect(reflect.New(typ))
			if newType.Kind() == reflect.Ptr && newType.IsNil() {
				newType.Set(reflect.New(newType.Type().Elem()))
			}
			elem, err = inspect(newType.Interface(), cache)
			if err != nil {
				return nil, err
			}
		}
		elem.Name = val.Type().Elem().String()
		elem.Parent = result
		result.Elem = elem
	case reflect.Struct:
		result.Type = attrTypeStruct

		// TODO: peek into cache, if enabled with 2 same fields, it will issue
		// TODO: recursive structures still not supported

		// prepare all props
		result.Properties = make(map[string]*attr)

		// iterate over struct fields
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldType := val.Type().Field(i)

			// skip unexported fields
			if !fieldType.IsExported() {
				continue
			}

			var (
				err       error
				fieldAttr *attr
			)

			// prepare new value for field, so we can inspect it
			if field.Kind() == reflect.Interface {
				fieldAttr = anyAttr()
			} else {
				if field.Type().Kind() == reflect.Ptr {
					// in case of pointer we get type what it points to and then reflect.New
					// field attribute returned from inspect
					fieldAttr, err = inspect(reflect.Indirect(reflect.New(field.Type().Elem())).Interface(), cache)
				} else {
					// field attribute returned from inspect
					fieldAttr, err = inspect(reflect.Indirect(reflect.New(field.Type())).Interface(), cache)
				}
			}

			// field attribute returned from inspect
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

			// Support for embedded structs
			if fieldType.Anonymous {
				fieldAttr.Name = fieldType.Type.Name()

				// merge properties from embedded struct to current struct
				// this is a naive way, since we don't store whole tree of embedded structs to set values.
				// this is prone to duplicates
				for name, prop := range fieldAttr.Properties {
					if _, ok := result.Properties[name]; ok {
						return nil, fmt.Errorf("%w: %v", ErrDuplicateField, name)
					}

					// naive way
					result.Properties[name] = prop
				}
			}

			// names and aliases
			fieldAttr.Name = fieldType.Name
			fieldAttr.Alias = pa.Alias
			fieldAttr.Parent = result
			if fieldAttr.Alias == "" {
				fieldAttr.Alias = fieldAttr.Name
			}

			// add field attribute to struct properties
			result.Properties[fieldAttr.Alias] = fieldAttr
		}
	case reflect.Map:
		result.Type = attrTypeMap

		// now check if key is string, because we support only string keys
		if val.Type().Key() != reflect.TypeOf("") {
			return nil, fmt.Errorf("%w: %s", ErrMapKeyNotStr, val.Type().Key().String())
		}

		// inspect value type
		elemType := val.Type().Elem()

		var (
			elemAttr *attr
			err      error
		)

		// shorthand for any type
		// this is not supported to pass to inspect (we know why) so we just return attr with any type (simple, huh?)
		if elemType.Kind() == reflect.Interface {
			elemAttr = anyAttr()
		} else {
			// prepare new value for field, so we can inspect it
			newValue := func() any {
				if elemType.Kind() == reflect.Ptr {
					return reflect.Indirect(reflect.New(elemType.Elem())).Interface()
				} else {
					return reflect.Indirect(reflect.New(elemType)).Interface()
				}
			}()

			// field attribute returned from inspect
			elemAttr, err = inspect(newValue, cache)
			if err != nil {
				return nil, err
			}
		}

		result.Elem = elemAttr
		result.Elem.Parent = result
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedType, val.Type().String())
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

	// integer type
	Signed bool

	// array/slice/map (for map we don't need key since only string is supported)
	Elem *attr

	// struct properties
	Properties map[string]*attr

	// Parent for better debugging
	Parent *attr
}

func anyAttr() *attr {
	return &attr{
		Type:     attrTypeAny,
		Nullable: true,
	}
}

// Set sets value to given target from parser.
// it returns error if value cannot be set or parsed attribute is invalid
func (a *attr) Set(target reflect.Value, parsed *parser.Attribute) error {
	// check if pointer is not nil, we need to provide new value
	if target.Kind() == reflect.Ptr && target.IsNil() {
		target.Set(reflect.New(target.Type().Elem()))
	}

	switch a.Type {
	case attrTypeArray:
		return a.setArray(target, parsed)
	case attrTypeBoolean:
		return a.setBoolean(target, parsed)
	case attrTypeFloat:
		return a.setFloat(target, parsed)
	case attrTypeInteger:
		return a.setInteger(target, parsed)
	case attrTypeString:
		return a.setString(target, parsed)
	case attrTypeStruct:
		return a.setStruct(target, parsed)
	case attrTypeMap:
		return a.setMap(target, parsed)
	default:
		return parser.NewParseError(parsed.Position, "invalid attribute type %v", a.Type)
	}
}

func (a *attr) setArray(target reflect.Value, parsed *parser.Attribute) error {
	if parsed.Array == nil {
		return parser.NewParseError(parsed.Position, "invalid value for %s", parsed.Name)
	}

	if target.Kind() == reflect.Ptr {
		target = target.Elem()
	}

	nu := reflect.Indirect(reflect.New(target.Type()))
	if nu.Kind() == reflect.Ptr && nu.IsNil() {
		nu.Set(reflect.New(nu.Type().Elem()))
	}

	// iterate over all values and set one by one
	for _, item := range parsed.Array {
		var (
			err error
			val reflect.Value
		)
		if a.Elem.Type == attrTypeAny {
			val, err = item.Build()
			if err != nil {
				return err
			}
		} else {
			val = reflect.Indirect(reflect.New(target.Type().Elem()))

			if err := a.Elem.Set(val, &item); err != nil {
				return fmt.Errorf("cannot set array value for %s: %s", parsed.Name, err)
			}
		}
		nu = reflect.Append(nu, val)
	}

	target.Set(nu)

	return nil
}

func (a *attr) setBoolean(target reflect.Value, parsed *parser.Attribute) error {
	if parsed.Value == nil || parsed.Value.Boolean == nil {
		return parser.NewParseError(parsed.Position, "invalid value for %s", parsed.Name)
	}

	val := *parsed.Value.Boolean

	if val != "true" && val != "false" {
		return parser.NewParseError(parsed.Position, "invalid value for %s", parsed.Name)
	}

	// we know it's correct
	value, _ := strconv.ParseBool(val)

	target.SetBool(value)

	return nil
}

func (a *attr) setFloat(target reflect.Value, parsed *parser.Attribute) error {
	if parsed.Value == nil || parsed.Value.Number == nil {
		return parser.NewParseError(parsed.Position, "invalid value for %s", parsed.Name)
	}
	if value, err := strconv.ParseFloat(*parsed.Value.Number, 64); err != nil {
		return parser.NewParseError(parsed.Position, "invalid value for %s", parsed.Name)
	} else {
		target.SetFloat(value)
	}
	return nil

}

func (a *attr) setInteger(target reflect.Value, parsed *parser.Attribute) error {
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

func (a *attr) setMap(target reflect.Value, parsed *parser.Attribute) error {
	// check if we really have object type, otherwise it's invalid
	if parsed.Attributes == nil {
		return parser.NewParseError(parsed.Position, "invalid value for %s", parsed.Name)
	}

	// check if target is nil (map is not initialized)
	if target.Kind() == reflect.Map && target.IsNil() {
		target.Set(reflect.MakeMap(target.Type()))
	}

	// special case for any type
	switch a.Elem.Type {
	case attrTypeAny:
		// special case for any type, we need to build recursively maps and stuff
		v, err := parsed.Build()
		if err != nil {
			return err
		}
		target.Set(v)
	default:
		// TODO: non any elem type

	}

	return nil
}

func (a *attr) setString(target reflect.Value, parsed *parser.Attribute) error {
	if parsed.Value == nil || parsed.Value.String == nil {
		return parser.NewParseError(parsed.Position, "invalid value for %s", parsed.Name)
	}
	if target.Kind() == reflect.Ptr {
		target = target.Elem()
	}

	target.SetString(*parsed.Value.String)

	return nil
}

func (a *attr) setStruct(target reflect.Value, parsed *parser.Attribute) error {
	// TODO: check other than struct types
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
		// if any type, we just build reflect.Value and set it
		if field.Kind() == reflect.Interface {
			v, err := att.Build()
			if err != nil {
				return err
			}
			field.Set(v)
		} else {
			// set property
			if err := prop.Set(field, &att); err != nil {
				return err
			}
		}

	}
	return nil
}
