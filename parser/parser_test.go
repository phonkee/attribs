package parser

import (
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Run("test parse simple attributes", func(t *testing.T) {
		data := []struct {
			input  string
			expect []Attribute
		}{
			//{"", []Attribute{}},
			//{"hello=1", []Attribute{
			//	att("hello", newSourceSpan(0, 5), val(newSourceSpan(6, 1), ptr("1"), nil, nil), nil, nil),
			//}},
			//{"hello='world'", []Attribute{
			//	att("hello", newSourceSpan(0, 5), val(newSourceSpan(6, 7), nil, ptr("world"), nil), nil, nil),
			//}},
			//{"hello='hello world', readonly", []Attribute{
			//	att("hello", newSourceSpan(0, 5), val(newSourceSpan(6, 13), nil, ptr("hello world"), nil), nil, nil),
			//	att("readonly", newSourceSpan(21, 8), val(newSourceSpan(21, 8), nil, nil, ptr("true")), nil, nil),
			//}},
			//{"hello='world', welcome = 'home'", []Attribute{
			//	att("hello", newSourceSpan(0, 5), val(newSourceSpan(6, 7), nil, ptr("world"), nil), nil, nil),
			//	att("welcome", newSourceSpan(15, 7), val(newSourceSpan(25, 6), nil, ptr("home"), nil), nil, nil),
			//}},
			//{"hello='world', welcome = 'home', number = 1.22, bool = true", []Attribute{
			//	att("hello", newSourceSpan(0, 5), val(newSourceSpan(6, 7), nil, ptr("world"), nil), nil, nil),
			//	att("welcome", newSourceSpan(15, 7), val(newSourceSpan(25, 6), nil, ptr("home"), nil), nil, nil),
			//	att("number", newSourceSpan(33, 6), val(newSourceSpan(42, 4), ptr("1.22"), nil, nil), nil, nil),
			//	att("bool", newSourceSpan(48, 4), val(newSourceSpan(55, 4), nil, nil, ptr("true")), nil, nil),
			//}},
			//{"hello='world', welcome = 'home', number = -1.22, bool = true", []Attribute{
			//	att("hello", newSourceSpan(0, 5), val(newSourceSpan(6, 7), nil, ptr("world"), nil), nil, nil),
			//	att("welcome", newSourceSpan(15, 7), val(newSourceSpan(25, 6), nil, ptr("home"), nil), nil, nil),
			//	att("number", newSourceSpan(33, 6), val(newSourceSpan(42, 5), ptr("-1.22"), nil, nil), nil, nil),
			//	att("bool", newSourceSpan(49, 4), val(newSourceSpan(56, 4), nil, nil, ptr("true")), nil, nil),
			//}},
			//{"hello='world', welcome = 'home', number = -.22, bool = true", []Attribute{
			//	att("hello", newSourceSpan(0, 5), val(newSourceSpan(6, 7), nil, ptr("world"), nil), nil, nil),
			//	att("welcome", newSourceSpan(15, 7), val(newSourceSpan(25, 6), nil, ptr("home"), nil), nil, nil),
			//	att("number", newSourceSpan(33, 6), val(newSourceSpan(42, 4), ptr("-0.22"), nil, nil), nil, nil),
			//	att("bool", newSourceSpan(48, 4), val(newSourceSpan(55, 4), nil, nil, ptr("true")), nil, nil),
			//}},
			//{"hello='world', welcome = 'home', number = -1, bool = true", []Attribute{
			//	att("hello", newSourceSpan(0, 5), val(newSourceSpan(6, 7), nil, ptr("world"), nil), nil, nil),
			//	att("welcome", newSourceSpan(15, 7), val(newSourceSpan(25, 6), nil, ptr("home"), nil), nil, nil),
			//	att("number", newSourceSpan(33, 6), val(newSourceSpan(42, 2), ptr("-1"), nil, nil), nil, nil),
			//	att("bool", newSourceSpan(46, 4), val(newSourceSpan(53, 4), nil, nil, ptr("true")), nil, nil),
			//}},
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
			//{"", []Attribute{}, ""},
			//{"hello(inner=1, other=2)", []Attribute{
			//	att("hello", newSourceSpan(0, 5), nil, []Attribute{
			//		att("inner", newSourceSpan(6, 5), val(newSourceSpan(12, 1), ptr("1"), nil, nil), nil, nil),
			//		att("other", newSourceSpan(15, 5), val(newSourceSpan(21, 1), ptr("2"), nil, nil), nil, nil),
			//	}, nil),
			//}, ""},
			//{"hello(inner=1, other=2, span(start=1, end=3))", []Attribute{
			//	att("hello", newSourceSpan(0, 5), nil, []Attribute{
			//		att("inner", newSourceSpan(6, 5), val(newSourceSpan(12, 1), ptr("1"), nil, nil), nil, nil),
			//		att("other", newSourceSpan(15, 5), val(newSourceSpan(21, 1), ptr("2"), nil, nil), nil, nil),
			//		att("span", newSourceSpan(24, 4), nil, []Attribute{
			//			att("start", newSourceSpan(29, 5), val(newSourceSpan(35, 1), ptr("1"), nil, nil), nil, nil),
			//			att("end", newSourceSpan(38, 3), val(newSourceSpan(42, 1), ptr("3"), nil, nil), nil, nil),
			//		}, nil),
			//	}, nil),
			//}, ""},
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
			expect      *Attribute
			expectError string
		}{
			//{"hello[1, true, 'yes', (one=1, two=2)]", att(
			//	"",
			//	newSourceSpan(0, 5),
			//	nil,
			//	newAttributes(
			//		newSourceSpan(0, 5),
			//		att("hello", newSourceSpan(0, 5), nil, nil, newAttributes(
			//			newSourceSpan(5, 32),
			//			att("", newSourceSpan(6, 1), val(newSourceSpan(6, 1), ptr("1"), nil, nil), nil, nil),
			//			att("", newSourceSpan(9, 4), val(newSourceSpan(9, 4), nil, nil, ptr("true")), nil, nil),
			//			att("", newSourceSpan(15, 5), val(newSourceSpan(15, 5), nil, ptr("yes"), nil), nil, nil),
			//			att("", newSourceSpan(22, 14), nil, newAttributes(
			//				newSourceSpan(22, 1),
			//				att("one", newSourceSpan(23, 3), val(newSourceSpan(27, 1), ptr("1"), nil, nil), nil, nil),
			//				att("two", newSourceSpan(30, 3), val(newSourceSpan(34, 1), ptr("2"), nil, nil), nil, nil),
			//			), nil),
			//		)),
			//	), nil,
			//),
			//	""},
			//{"hello[[1, 2]]", att(
			//	"",
			//	newSourceSpan(5, 8),
			//	nil,
			//	newAttributes(
			//		newSourceSpan(5, 8),
			//		att("hello", newSourceSpan(0, 5), nil, nil, newAttributes(
			//			newSourceSpan(5, 8),
			//			att("", newSourceSpan(6, 6), nil, nil, newAttributes(
			//				newSourceSpan(6, 6),
			//				att("", newSourceSpan(7, 1), val(newSourceSpan(7, 1), ptr("1"), nil, nil), nil, nil),
			//				att("", newSourceSpan(10, 1), val(newSourceSpan(10, 1), ptr("2"), nil, nil), nil, nil),
			//			)),
			//		)),
			//	),
			//	nil,
			//), ""},
			{"hello[42, 999, 'world']", att(
				"",
				newSourceSpan(5, 8),
				nil,
				newAttributes(
					newSourceSpan(5, 8),
					att("hello", newSourceSpan(0, 5), nil, nil, newAttributes(
						newSourceSpan(5, 8),
						att("", newSourceSpan(6, 6), nil, nil, newAttributes(
							newSourceSpan(6, 6),
							att("", newSourceSpan(7, 1), val(newSourceSpan(7, 1), ptr("1"), nil, nil), nil, nil),
							att("", newSourceSpan(10, 1), val(newSourceSpan(10, 1), ptr("2"), nil, nil), nil, nil),
							att("", newSourceSpan(10, 1), val(newSourceSpan(10, 1), ptr("world"), nil, nil), nil, nil),
						)),
					)),
				),
				nil,
			), ""},
		}

		for _, item := range data {
			result, err := Parse(strings.NewReader(item.input))
			spew.Dump(result)
			assert.NoError(t, err)
			//assert.Equal(t, item.expect, result)
		}
	})
	t.Run("test objects", func(t *testing.T) {
		data := []struct {
			input       string
			expect      *Attribute
			expectError string
		}{
			{"hello(one=1, two=2)", &Attribute{
				Object: newAttributes(
					newSourceSpan(5, 14),
					att("hello", newSourceSpan(0, 5), nil, newAttributes(
						newSourceSpan(5, 14),
						att("one", newSourceSpan(6, 3), val(newSourceSpan(10, 1), ptr("1"), nil, nil), nil, nil),
						att("two", newSourceSpan(13, 3), val(newSourceSpan(17, 1), ptr("2"), nil, nil), nil, nil),
					), nil),
				),
			}, ""},
		}

		for _, item := range data {
			result, err := Parse(strings.NewReader(item.input))
			assert.NoError(t, err)
			spew.Dump(result)
			//assert.Equal(t, item.expect, result)
		}
	})
}

func ptr[T any](t T) *T {
	return &t
}

func val(span *SourceSpan, number *string, str *string, bool *string) *Value {
	return &Value{
		Span:    span,
		Number:  number,
		String:  str,
		Boolean: bool,
	}
}

func att(name string, span *SourceSpan, value *Value, attribs *Attributes, array *Attributes) *Attribute {
	return &Attribute{
		Name:   name,
		Span:   span,
		Value:  value,
		Object: attribs,
		Array:  array,
	}
}
