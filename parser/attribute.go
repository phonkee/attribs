package parser

import (
	"reflect"
	"strconv"
)

// Attribute representation
type Attribute struct {
	// name of attribute
	Name string

	// pos in original string
	Position int

	// This is poor man's union in go (not complaining, just saying)
	Value      *Value
	Attributes []Attribute
	Array      []Attribute
}

// HasValue returns whether any val was set to attribute
// if used via parser, one of values is always set
func (a *Attribute) HasValue() bool {
	return a.Value != nil || a.Attributes != nil || a.Array != nil
}

// Build builds any value from attribute
func (a *Attribute) Build() (reflect.Value, error) {
	switch {
	case a.Value != nil:
		return a.Value.BuildValue()
	case a.Attributes != nil:
		mt := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf((*any)(nil)).Elem())
		newValue := reflect.MakeMapWithSize(mt, 0)

		for _, attr := range a.Attributes {
			value, err := attr.Build()
			if err != nil {
				return reflect.Value{}, err
			}
			newValue.SetMapIndex(reflect.ValueOf(attr.Name), value)
		}

		return newValue, nil
	case a.Array != nil:
		newValue := reflect.MakeSlice(reflect.TypeOf([]any{}), 0, len(a.Array))
		for _, attr := range a.Array {
			value, err := attr.Build()
			if err != nil {
				return reflect.Value{}, err
			}
			newValue = reflect.Append(newValue, value)
		}
		return newValue, nil
	}

	return reflect.Value{}, NewParseError(a.Position, "Attribute has no value")
}

func (a *Attribute) PrepareAny() (reflect.Value, error) {
	switch {
	case a.Value != nil:
		switch {
		case a.Value.String != nil:
			return reflect.ValueOf(""), nil
		case a.Value.Number != nil:
			_, err := strconv.ParseInt(*a.Value.Number, 10, 64)
			if err != nil {
				_, err := strconv.ParseFloat(*a.Value.Number, 64)
				if err != nil {
					return reflect.Value{}, NewParseError(a.Position, "Invalid number: %s", *a.Value.Number)
				}
				return reflect.ValueOf(float64(0)), nil
			}
			return reflect.ValueOf(int64(0)), nil
		case a.Value.Boolean != nil:
			return reflect.ValueOf(false), nil
		}
	case a.Attributes != nil:
		// map
		return reflect.Value{}, nil
	case a.Array != nil:
		return reflect.Value{}, nil
	}
	return reflect.Value{}, NewParseError(a.Position, "Attribute has no value")
}
