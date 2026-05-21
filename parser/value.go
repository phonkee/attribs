package parser

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type ValueType int

const (
	ValueTypeString ValueType = 1 << iota
	ValueTypeNumber
	ValueTypeBoolean
)

var (
	ValueTypeAll = []ValueType{ValueTypeString, ValueTypeNumber, ValueTypeBoolean}
)

func (v ValueType) FromRaw(raw string, val *Value) error {
	if v&ValueTypeString != 0 {
		val.String = &raw
	}
	if v&ValueTypeNumber != 0 {
		val.Number = &raw
	}
	if v&ValueTypeBoolean != 0 {
		if _, err := strconv.ParseBool(raw); err != nil {
			return fmt.Errorf("cannot parse boolean value: %s", raw)
		}
		val.Boolean = &raw
	}
	return nil
}

// Value represents val of attribute
// - Number represents any number int/float
// - String represents: string, bool
type Value struct {
	Span    *SourceSpan
	Boolean *string
	Number  *string
	String  *string
}

func (v *Value) ClearAll() {
	v.Boolean = nil
	v.Number = nil
	v.String = nil
}

func (v *Value) FromRaw(valueType ValueType, raw string) error {
	return valueType.FromRaw(raw, v)
}

func (v *Value) BuildValue() (reflect.Value, error) {
	switch {
	case v.Boolean != nil:
		b, err := strconv.ParseBool(*v.Boolean)
		if err != nil {
			return reflect.Value{}, NewParseError(v.Span, "invalid boolean value: %s", *v.Boolean)
		}
		target := reflect.Indirect(reflect.New(reflect.TypeOf(b)))
		if !target.CanSet() {
			return reflect.Value{}, NewParseError(v.Span, "cannot set bool value: %s", *v.Boolean)
		}
		target.SetBool(b)
		return target, nil
	case v.Number != nil:
		r, err := strconv.ParseInt(*v.Number, 10, 64)
		if err != nil {
			r, err := strconv.ParseFloat(*v.Number, 64)
			if err != nil {
				return reflect.Value{}, NewParseError(v.Span, "invalid number value: %s", *v.Number)
			}
			target := reflect.Indirect(reflect.New(reflect.TypeOf(r)))
			if !target.CanSet() {
				return reflect.Value{}, NewParseError(v.Span, "cannot set float value: %s", *v.Number)
			}
			target.SetFloat(r)
			return target, nil
		}
		typ := reflect.TypeOf(int(0))
		target := reflect.Indirect(reflect.New(typ))
		if !target.CanSet() {
			return reflect.Value{}, NewParseError(v.Span, "cannot set int value: %s", *v.Number)
		}
		target.SetInt(r)
		return target, nil
	case v.String != nil:
		typ := reflect.TypeOf(*v.String)
		target := reflect.Indirect(reflect.New(typ))
		if !target.CanSet() {
			return reflect.Value{}, NewParseError(v.Span, "cannot set string value: %s", *v.String)
		}
		target.SetString(*v.String)
		return target, nil
	default:
		return reflect.Value{}, NewParseError(v.Span, "value has no value")
	}
}

func (v *Value) IsZero() bool {
	return v.Boolean == nil && v.Number == nil && v.String == nil
}

func (v *Value) AsBool() (bool, error) {
	if v.Boolean != nil {
		return strconv.ParseBool(*v.Boolean)
	}
	if v.IsZero() {
		return true, nil
	}
	return false, NewParseError(v.Span, "value has not boolean value")
}

func (v *Value) AsInt() (int, error) {
	if v.Number != nil {
		parsed, err := strconv.ParseInt(*v.Number, 10, 64)
		if err != nil {
			return 0, NewParseError(v.Span, "invalid number value: %s", *v.Number)
		}
		return int(parsed), nil
	}
	return 0, NewParseError(v.Span, "value has not int value")
}

func (v *Value) AsFloat() (float64, error) {
	if v.Number != nil {
		parsed, err := strconv.ParseFloat(*v.Number, 64)
		if err != nil {
			return 0, NewParseError(v.Span, "invalid number value: %s", *v.Number)
		}
		return parsed, nil
	}
	return 0, NewParseError(v.Span, "value has not float value")
}

func (v *Value) AsString() (string, error) {
	if v.String != nil {
		return *v.String, nil
	}
	return "", NewParseError(v.Span, "value has not string value")
}

func (v *Value) AsTrimmedString() (string, error) {
	result, err := v.AsString()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result), nil
}

func (v *Value) IsBool() bool {
	_, err := v.AsBool()
	return err == nil
}

func (v *Value) IsInt() bool {
	_, err := v.AsInt()
	return err == nil
}

func (v *Value) IsFloat() bool {
	_, err := v.AsFloat()
	return err == nil
}

func (v *Value) IsString() bool {
	_, err := v.AsString()
	return err == nil
}
