package attribs

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"reflect"
	"strconv"

	"github.com/phonkee/attribs/parser"
)

// Attribute defines interface for attribute
type Attribute interface {
	// Init initializes Attribute with given type
	Init(value reflect.Type, cache Cache) error

	// SetValue sets value for given attribute
	SetValue(reflect.Value, *parser.Attribute) error

	// Item returns item of attribute
	Item() (*parser.Attribute, bool)
}

// baseAttribute defines base attribute Properties
type baseAttribute struct {
	Name     string
	Alias    string
	Disabled bool
	Nullable bool
}

// Init is base for all embeddees
func (b baseAttribute) Init(value reflect.Type, cache Cache) error { return nil }

type structAttribute struct {
	baseAttribute
	fields map[string]attrInfo
}

type attrInfo struct {
	attr Attribute
	base baseAttribute
}

// Init initializes Attribute with given value and tag
func (s *structAttribute) Init(typ reflect.Type, cache Cache) error {
	// init fields
	if s.fields == nil {
		s.fields = map[string]attrInfo{}
	}

	// iterate over all fields
	for i := 0; i < typ.NumField(); i++ {
		ft := typ.Field(i)
		ftType := ft.Type
		if ft.PkgPath != "" {
			continue
		}
		// support for embedded structs
		if ft.Anonymous {
			if err := s.Init(ftType, cache); err != nil {
				return err
			}
			continue
		}

		// parse attribs tag first
		pa, err := parseAttribsTag(ft.Tag.Get(TagName))
		if err != nil {
			return err
		}
		base := baseAttribute{
			Name:     ft.Name,
			Alias:    pa.Alias,
			Disabled: pa.Disabled,
		}
		if base.Alias == "" {
			base.Alias = base.Name
		}
		if ftType.Kind() == reflect.Ptr {
			ftType = ftType.Elem()
			base.Nullable = true
		}

		p, err := define(ftType, base, cache)
		if err != nil {
			return err
		}
		s.fields[base.Alias] = attrInfo{
			attr: p,
			base: base,
		}
	}

	return nil
}

// SetValue sets value for given attribute
func (s *structAttribute) SetValue(result reflect.Value, attr *parser.Attribute) error {
	// check if we have proper value (attributes - struct)
	if attr.Attributes == nil {
		return parser.NewParseError(attr.Position, "invalid value for %s", attr.Name)
	}
	if !result.CanSet() {
		return fmt.Errorf("cannot set value for struct %s", attr.Name)
	}

	if result.Type().Kind() == reflect.Ptr {
		result.Set(reflect.New(result.Type().Elem()))
	}

	for _, att := range attr.Attributes {
		sa, ok := s.fields[att.Name]
		if !ok {
			return parser.NewParseError(attr.Position, "unknown attribute %s", att.Name)
		}

		var field reflect.Value
		if result.Kind() == reflect.Ptr {
			field = result.Elem().FieldByName(sa.base.Name)
		} else {
			field = result.FieldByName(sa.base.Name)
		}
		if err := sa.attr.SetValue(field, &att); err != nil {
			return err
		}
	}

	return nil
}

// Item is no op
func (*structAttribute) Item() (*parser.Attribute, bool) {
	return nil, false
}

// intAttribute handles numbers
type intAttribute struct {
	baseAttribute
	width    int
	unsigned bool
}

// Item is no op
func (*intAttribute) Item() (*parser.Attribute, bool) {
	return nil, false
}

func (i *intAttribute) SetValue(result reflect.Value, attr *parser.Attribute) error {
	if attr.Value == nil || attr.Value.Number == nil {
		return parser.NewParseError(attr.Position, "invalid value for %s", attr.Name)
	}
	if !result.CanSet() {
		return parser.NewParseError(attr.Position, "cannot set value for %s", attr.Name)
	}

	if i.unsigned {
		if value, err := strconv.ParseUint(*attr.Value.Number, 10, i.width); err != nil {
			return err
		} else {
			if result.Kind() == reflect.Ptr {
				result.Set(reflect.New(result.Type().Elem()))
				result.Elem().SetUint(value)
			} else {
				result.SetUint(value)
			}
		}
	} else {
		if value, err := strconv.ParseInt(*attr.Value.Number, 10, i.width); err != nil {
			return err
		} else {
			if result.Kind() == reflect.Ptr {
				result.Set(reflect.New(result.Type().Elem()))
				result.Elem().SetInt(value)
			} else {
				result.SetInt(value)
			}
		}
	}

	return nil
}

type floatAttribute struct {
	baseAttribute
	width int
}

func (f *floatAttribute) SetValue(result reflect.Value, attr *parser.Attribute) error {
	if attr.Value == nil || attr.Value.Number == nil {
		return parser.NewParseError(attr.Position, "invalid value for %s", attr.Name)
	}
	if !result.CanSet() {
		return parser.NewParseError(attr.Position, "cannot set value for %s", attr.Name)
	}
	if value, err := strconv.ParseFloat(*attr.Value.Number, f.width); err != nil {
		return parser.NewParseError(attr.Position, "invalid value for %s", attr.Name)
	} else {
		if result.Kind() == reflect.Ptr {
			result.Set(reflect.New(result.Type().Elem()))
			result.Elem().SetFloat(value)
		} else {
			result.SetFloat(value)
		}
	}
	return nil
}

// Item is no op
func (*floatAttribute) Item() (*parser.Attribute, bool) {
	return nil, false
}

type boolAttribute struct {
	baseAttribute
}

func (b *boolAttribute) SetValue(result reflect.Value, attr *parser.Attribute) error {
	if attr.Value == nil || attr.Value.Boolean == nil {
		return parser.NewParseError(attr.Position, "invalid value for %s", attr.Name)
	}
	if !result.CanSet() {
		return parser.NewParseError(attr.Position, "cannot set value for %s", attr.Name)
	}

	val := *attr.Value.Boolean

	if val != "true" && val != "false" {
		return parser.NewParseError(attr.Position, "invalid value for %s", attr.Name)
	}

	// we know it's correct
	value, _ := strconv.ParseBool(val)

	if result.Kind() == reflect.Ptr {
		result.Set(reflect.New(result.Type().Elem()))
		result.Elem().SetBool(value)
	} else {
		result.SetBool(value)
	}

	return nil
}

// Item is no op
func (*boolAttribute) Item() (*parser.Attribute, bool) {
	return nil, false
}

type stringAttribute struct {
	baseAttribute
}

func (s *stringAttribute) SetValue(result reflect.Value, attr *parser.Attribute) error {
	if attr.Value == nil || attr.Value.String == nil {
		return parser.NewParseError(attr.Position, "invalid value for %s", attr.Name)
	}
	if !result.CanSet() {
		return parser.NewParseError(attr.Position, "cannot set value for %s", attr.Name)
	}

	if result.Kind() == reflect.Ptr {
		result.Set(reflect.New(result.Type().Elem()))
		result.Elem().SetString(*attr.Value.String)
	} else {
		result.SetString(*attr.Value.String)
	}

	return nil
}

// Item is no op
func (*stringAttribute) Item() (*parser.Attribute, bool) {
	return nil, false
}

type arrayAttribute struct {
	baseAttribute
	attr    Attribute
	itemTyp reflect.Type
}

// Init initializes arrayAttribute with proper Attribute
func (a *arrayAttribute) Init(typ reflect.Type, cache Cache) (err error) {
	a.attr, err = define(typ.Elem(), baseAttribute{}, cache)
	a.itemTyp = typ.Elem()
	return
}

// SetValue iterates over array values and sets them
func (a *arrayAttribute) SetValue(result reflect.Value, attr *parser.Attribute) error {
	if attr.Array == nil {
		return parser.NewParseError(attr.Position, "invalid value for %s", attr.Name)
	}
	if !result.CanSet() {
		return parser.NewParseError(attr.Position, "cannot set value for %s", attr.Name)
	}

	var (
		typ   = result.Type()
		isPtr = typ.Kind() == reflect.Ptr
	)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	nu := reflect.Indirect(reflect.New(typ))

	for _, item := range attr.Array {
		val := reflect.Indirect(reflect.New(a.itemTyp))
		if err := a.attr.SetValue(val, &item); err != nil {
			return fmt.Errorf("cannot set array value for %s: %s", attr.Name, err)
		}
		nu = reflect.Append(nu, val)
	}

	if isPtr {
		spew.Dump(result)
		result.Set(nu)
	} else {
		result.Set(nu)
	}

	return nil
}

func (a *arrayAttribute) Item() (*parser.Attribute, bool) {
	panic("here")
	return nil, false
}
