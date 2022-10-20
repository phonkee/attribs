package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

//func TestPlayground(t *testing.T) {
//	// Test playground
//	attrs := "hello(inner=1, other=2, span(start=1, end=3, inner(required,disabled,outer(required,disabled)))), blank()"
//	x, err := Parse(strings.NewReader(attrs))
//	assert.NoError(t, err)
//	spew.Dump(x)
//}

func TestParse(t *testing.T) {
	t.Run("test parse simple attributes", func(t *testing.T) {
		data := []struct {
			input  string
			expect []Attribute
		}{
			{"", []Attribute{}},
			{"hello=1", []Attribute{
				att("hello", 0, val(6, ptr("1"), nil), nil, nil),
			}},
			{"hello='world'", []Attribute{
				att("hello", 0, val(6, nil, ptr("world")), nil, nil),
			}},
			{"hello='hello world', readonly", []Attribute{
				att("hello", 0, val(6, nil, ptr("hello world")), nil, nil),
				att("readonly", 21, val(21, nil, ptr("true")), nil, nil),
			}},
			{"hello='world', welcome = 'home'", []Attribute{
				att("hello", 0, val(6, nil, ptr("world")), nil, nil),
				att("welcome", 15, val(25, nil, ptr("home")), nil, nil),
			}},
			{"hello='world', welcome = 'home', number = 1.22, bool = true", []Attribute{
				att("hello", 0, val(6, nil, ptr("world")), nil, nil),
				att("welcome", 15, val(25, nil, ptr("home")), nil, nil),
				att("number", 33, val(42, ptr("1.22"), nil), nil, nil),
				att("bool", 48, val(55, nil, ptr("true")), nil, nil),
			}},
		}

		for _, item := range data {
			result, err := Parse(strings.NewReader(item.input))
			assert.NoError(t, err)
			assert.Equal(t, item.expect, result)
		}
	})

	t.Run("test recursive attributes", func(t *testing.T) {
		data := []struct {
			input       string
			expect      []Attribute
			expectError string
		}{
			{"", []Attribute{}, ""},
			{"hello(inner=1, other=2)", []Attribute{
				att("hello", 0, nil, []Attribute{
					att("inner", 6, val(12, ptr("1"), nil), nil, nil),
					att("other", 15, val(21, ptr("2"), nil), nil, nil),
				}, nil),
			}, ""},
			{"hello(inner=1, other=2, span(start=1, end=3))", []Attribute{
				att("hello", 0, nil, []Attribute{
					att("inner", 6, val(12, ptr("1"), nil), nil, nil),
					att("other", 15, val(21, ptr("2"), nil), nil, nil),
					att("span", 24, nil, []Attribute{
						att("start", 29, val(35, ptr("1"), nil), nil, nil),
						att("end", 38, val(42, ptr("3"), nil), nil, nil),
					}, nil),
				}, nil),
			}, ""},
		}

		for _, item := range data {
			result, err := Parse(strings.NewReader(item.input))
			assert.NoError(t, err)
			assert.Equal(t, item.expect, result)
		}
	})

	t.Run("test arrays", func(t *testing.T) {
		data := []struct {
			input       string
			expect      []Attribute
			expectError string
		}{
			{"hello[1, true, 'yes', (one=1, two=2)]", []Attribute{
				att("hello", 0, nil, nil, []Attribute{
					att("", 6, val(6, ptr("1"), nil), nil, nil),
					att("", 9, val(9, nil, ptr("true")), nil, nil),
					att("", 15, val(15, nil, ptr("yes")), nil, nil),
					att("", 22, nil, []Attribute{
						att("one", 23, val(27, ptr("1"), nil), nil, nil),
						att("two", 30, val(34, ptr("2"), nil), nil, nil),
					}, nil),
				}),
			}, ""},
			{"hello[[1, 2]]", []Attribute{
				att("hello", 0, nil, nil, []Attribute{
					att("", 6, nil, nil, []Attribute{
						att("", 7, val(7, ptr("1"), nil), nil, nil),
						att("", 10, val(10, ptr("2"), nil), nil, nil),
					}),
				}),
			}, ""},
		}

		for _, item := range data {
			result, err := Parse(strings.NewReader(item.input))
			assert.NoError(t, err)
			assert.Equal(t, item.expect, result)
		}
	})

}

func ptr[T any](t T) *T {
	return &t
}

func val(pos int, number *string, str *string) *Value {
	return &Value{
		Position: pos,
		Number:   number,
		String:   str,
	}
}

func att(name string, pos int, value *Value, attribs []Attribute, array []Attribute) Attribute {
	return Attribute{
		Name:       name,
		Position:   pos,
		Value:      value,
		Attributes: attribs,
		Array:      array,
	}
}
