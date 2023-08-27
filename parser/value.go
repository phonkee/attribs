package parser

import (
	"reflect"
	"strconv"
)

// Value represents val of attribute
// - Number represents any number int/float
// - String represents: string, bool
type Value struct {
	Position int
	Boolean  *string
	Number   *string
	String   *string
}

func (v Value) BuildValue() (reflect.Value, error) {
	switch {
	case v.Boolean != nil:
		b, err := strconv.ParseBool(*v.Boolean)
		if err != nil {
			return reflect.Value{}, NewParseError(v.Position, "invalid boolean value: %s", *v.Boolean)
		}
		target := reflect.Indirect(reflect.New(reflect.TypeOf(b)))
		if !target.CanSet() {
			// TODO: remove this
			panic("cannot set bool value")

		}
		target.SetBool(b)
		return target, nil
	case v.Number != nil:
		r, err := strconv.ParseInt(*v.Number, 10, 64)
		if err != nil {
			r, err := strconv.ParseFloat(*v.Number, 64)
			if err != nil {
				return reflect.Value{}, NewParseError(v.Position, "invalid number value: %s", *v.Number)
			}
			target := reflect.Indirect(reflect.New(reflect.TypeOf(r)))
			if !target.CanSet() {
				// TODO: remove this
				panic("cannot set float value")
			}
			target.SetFloat(r)
			return target, nil
		}
		typ := reflect.TypeOf(int(0))
		target := reflect.Indirect(reflect.New(typ))
		if !target.CanSet() {
			// TODO: remove this
			panic("cannot set int value")
		}
		target.SetInt(r)
		return target, nil
	case v.String != nil:
		typ := reflect.TypeOf(*v.String)
		target := reflect.Indirect(reflect.New(typ))
		if !target.CanSet() {
			panic("cannot set string value")
		}
		target.SetString(*v.String)
		return target, nil
	default:
		return reflect.Value{}, NewParseError(v.Position, "value has no value")
	}
}
