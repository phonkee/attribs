package parser

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── helpers ────────────────────────────────────────────────────────────────

func ptr[T any](t T) *T { return &t }

// mustParse calls Parse and requires no error.
func mustParse(t *testing.T, input string) *Attribute {
	t.Helper()
	result, err := Parse(strings.NewReader(input))
	require.NoError(t, err, "Parse(%q)", input)
	require.NotNil(t, result)
	return result
}

// topAttrs returns the top-level attribute slice from a Parse() result.
func topAttrs(a *Attribute) []*Attribute {
	if a == nil || a.Object == nil {
		return nil
	}
	return a.Object.Attributes
}

// ─── TestParse ───────────────────────────────────────────────────────────────

func TestParse(t *testing.T) {
	t.Run("empty_input", func(t *testing.T) {
		got := mustParse(t, "")
		assert.Empty(t, topAttrs(got))
		assert.NotNil(t, got.Span)
	})

	t.Run("key_equals_integer", func(t *testing.T) {
		a := topAttrs(mustParse(t, "id=42"))
		require.Len(t, a, 1)
		assert.Equal(t, "id", a[0].Name)
		require.NotNil(t, a[0].Value)
		assert.Equal(t, ptr("42"), a[0].Value.Number)
		assert.Nil(t, a[0].Value.String)
		assert.Nil(t, a[0].Value.Boolean)
	})

	t.Run("key_equals_negative_integer", func(t *testing.T) {
		a := topAttrs(mustParse(t, "n=-7"))
		require.Len(t, a, 1)
		assert.Equal(t, ptr("-7"), a[0].Value.Number)
	})

	t.Run("key_equals_float", func(t *testing.T) {
		a := topAttrs(mustParse(t, "f=3.14"))
		require.Len(t, a, 1)
		assert.Equal(t, ptr("3.14"), a[0].Value.Number)
	})

	t.Run("key_equals_negative_float", func(t *testing.T) {
		a := topAttrs(mustParse(t, "f=-1.5"))
		require.Len(t, a, 1)
		assert.Equal(t, ptr("-1.5"), a[0].Value.Number)
	})

	t.Run("key_equals_single_quoted_string", func(t *testing.T) {
		a := topAttrs(mustParse(t, "s='hello world'"))
		require.Len(t, a, 1)
		require.NotNil(t, a[0].Value)
		assert.Equal(t, ptr("hello world"), a[0].Value.String)
		assert.Nil(t, a[0].Value.Boolean)
		assert.Nil(t, a[0].Value.Number)
	})

	t.Run("key_equals_double_quoted_string", func(t *testing.T) {
		a := topAttrs(mustParse(t, `s="hello world"`))
		require.Len(t, a, 1)
		assert.Equal(t, ptr("hello world"), a[0].Value.String)
	})

	t.Run("key_equals_unquoted_ident", func(t *testing.T) {
		a := topAttrs(mustParse(t, "s=world"))
		require.Len(t, a, 1)
		require.NotNil(t, a[0].Value)
		assert.Equal(t, ptr("world"), a[0].Value.String)
		assert.Nil(t, a[0].Value.Boolean)
	})

	t.Run("key_equals_true", func(t *testing.T) {
		a := topAttrs(mustParse(t, "ok=true"))
		require.Len(t, a, 1)
		require.NotNil(t, a[0].Value)
		assert.Equal(t, ptr("true"), a[0].Value.Boolean)
		assert.Equal(t, ptr("true"), a[0].Value.String)
		assert.Nil(t, a[0].Value.Number)
	})

	t.Run("key_equals_false", func(t *testing.T) {
		a := topAttrs(mustParse(t, "ok=false"))
		require.Len(t, a, 1)
		assert.Equal(t, ptr("false"), a[0].Value.Boolean)
		assert.Equal(t, ptr("false"), a[0].Value.String)
	})

	t.Run("bare_boolean_flag", func(t *testing.T) {
		a := topAttrs(mustParse(t, "disabled"))
		require.Len(t, a, 1)
		assert.Equal(t, "disabled", a[0].Name)
		require.NotNil(t, a[0].Value)
		// bare flag sets both Boolean and String to "true"
		assert.Equal(t, ptr("true"), a[0].Value.Boolean)
		assert.Equal(t, ptr("true"), a[0].Value.String)
	})

	t.Run("multiple_key_value_pairs", func(t *testing.T) {
		a := topAttrs(mustParse(t, "a=1, b='hello', c=true"))
		require.Len(t, a, 3)
		assert.Equal(t, "a", a[0].Name)
		assert.Equal(t, ptr("1"), a[0].Value.Number)
		assert.Equal(t, "b", a[1].Name)
		assert.Equal(t, ptr("hello"), a[1].Value.String)
		assert.Equal(t, "c", a[2].Name)
		assert.Equal(t, ptr("true"), a[2].Value.Boolean)
	})

	t.Run("multiple_bare_flags", func(t *testing.T) {
		a := topAttrs(mustParse(t, "x, y, z"))
		require.Len(t, a, 3)
		for i, name := range []string{"x", "y", "z"} {
			assert.Equal(t, name, a[i].Name)
			assert.Equal(t, ptr("true"), a[i].Value.Boolean)
		}
	})

	t.Run("mixed_kv_and_flags", func(t *testing.T) {
		a := topAttrs(mustParse(t, "id=42, disabled, name='foo'"))
		require.Len(t, a, 3)
		assert.Equal(t, ptr("42"), a[0].Value.Number)
		assert.Equal(t, ptr("true"), a[1].Value.Boolean)
		assert.Equal(t, ptr("foo"), a[2].Value.String)
	})

	t.Run("nested_object", func(t *testing.T) {
		a := topAttrs(mustParse(t, "span(start=1, end=2)"))
		require.Len(t, a, 1)
		assert.Equal(t, "span", a[0].Name)
		require.NotNil(t, a[0].Object)
		inner := a[0].Object.Attributes
		require.Len(t, inner, 2)
		assert.Equal(t, "start", inner[0].Name)
		assert.Equal(t, ptr("1"), inner[0].Value.Number)
		assert.Equal(t, "end", inner[1].Name)
		assert.Equal(t, ptr("2"), inner[1].Value.Number)
	})

	t.Run("empty_nested_object", func(t *testing.T) {
		a := topAttrs(mustParse(t, "a()"))
		require.Len(t, a, 1)
		assert.Equal(t, "a", a[0].Name)
		require.NotNil(t, a[0].Object)
		assert.Empty(t, a[0].Object.Attributes)
	})

	t.Run("deeply_nested_object", func(t *testing.T) {
		a := topAttrs(mustParse(t, "a(b(c=99))"))
		require.Len(t, a, 1)
		b := a[0].Object.Attributes
		require.Len(t, b, 1)
		assert.Equal(t, "b", b[0].Name)
		c := b[0].Object.Attributes
		require.Len(t, c, 1)
		assert.Equal(t, "c", c[0].Name)
		assert.Equal(t, ptr("99"), c[0].Value.Number)
	})

	t.Run("object_with_boolean_flag_inside", func(t *testing.T) {
		a := topAttrs(mustParse(t, "opts(verbose, count=3)"))
		inner := a[0].Object.Attributes
		require.Len(t, inner, 2)
		assert.Equal(t, ptr("true"), inner[0].Value.Boolean)
		assert.Equal(t, ptr("3"), inner[1].Value.Number)
	})

	t.Run("simple_array", func(t *testing.T) {
		a := topAttrs(mustParse(t, "nums[1, 2, 3]"))
		require.Len(t, a, 1)
		require.NotNil(t, a[0].Array)
		items := a[0].Array.Attributes
		require.Len(t, items, 3)
		assert.Equal(t, ptr("1"), items[0].Value.Number)
		assert.Equal(t, ptr("2"), items[1].Value.Number)
		assert.Equal(t, ptr("3"), items[2].Value.Number)
	})

	t.Run("empty_array", func(t *testing.T) {
		a := topAttrs(mustParse(t, "tags[]"))
		require.Len(t, a, 1)
		require.NotNil(t, a[0].Array)
		assert.Empty(t, a[0].Array.Attributes)
	})

	t.Run("array_of_strings", func(t *testing.T) {
		a := topAttrs(mustParse(t, "tags['go', 'test']"))
		items := a[0].Array.Attributes
		require.Len(t, items, 2)
		assert.Equal(t, ptr("go"), items[0].Value.String)
		assert.Equal(t, ptr("test"), items[1].Value.String)
	})

	t.Run("array_of_mixed_types", func(t *testing.T) {
		a := topAttrs(mustParse(t, "x[1, 'hello', true]"))
		items := a[0].Array.Attributes
		require.Len(t, items, 3)
		assert.NotNil(t, items[0].Value.Number)
		assert.NotNil(t, items[1].Value.String)
		assert.NotNil(t, items[2].Value.Boolean)
	})

	t.Run("array_of_objects", func(t *testing.T) {
		a := topAttrs(mustParse(t, "pts[(x=1, y=2), (x=3, y=4)]"))
		items := a[0].Array.Attributes
		require.Len(t, items, 2)
		require.NotNil(t, items[0].Object)
		assert.Len(t, items[0].Object.Attributes, 2)
		assert.Equal(t, ptr("1"), items[0].Object.Attributes[0].Value.Number)
		assert.Equal(t, ptr("3"), items[1].Object.Attributes[0].Value.Number)
	})

	t.Run("nested_array", func(t *testing.T) {
		a := topAttrs(mustParse(t, "m[[1, 2], [3, 4]]"))
		items := a[0].Array.Attributes
		require.Len(t, items, 2)
		require.NotNil(t, items[0].Array)
		inner0 := items[0].Array.Attributes
		inner1 := items[1].Array.Attributes
		require.Len(t, inner0, 2)
		assert.Equal(t, ptr("1"), inner0[0].Value.Number)
		assert.Equal(t, ptr("2"), inner0[1].Value.Number)
		require.Len(t, inner1, 2)
		assert.Equal(t, ptr("3"), inner1[0].Value.Number)
		assert.Equal(t, ptr("4"), inner1[1].Value.Number)
	})

	t.Run("positional_number", func(t *testing.T) {
		a := topAttrs(mustParse(t, "42"))
		require.Len(t, a, 1)
		assert.Equal(t, "", a[0].Name)
		assert.Equal(t, ptr("42"), a[0].Value.Number)
	})

	t.Run("positional_string", func(t *testing.T) {
		a := topAttrs(mustParse(t, "'hello'"))
		require.Len(t, a, 1)
		assert.Equal(t, "", a[0].Name)
		assert.Equal(t, ptr("hello"), a[0].Value.String)
	})

	t.Run("positional_object", func(t *testing.T) {
		a := topAttrs(mustParse(t, "(a=1, b=2)"))
		require.Len(t, a, 1)
		assert.Equal(t, "", a[0].Name)
		require.NotNil(t, a[0].Object)
		assert.Len(t, a[0].Object.Attributes, 2)
	})

	t.Run("positional_array", func(t *testing.T) {
		a := topAttrs(mustParse(t, "[1, 2, 3]"))
		require.Len(t, a, 1)
		assert.Equal(t, "", a[0].Name)
		require.NotNil(t, a[0].Array)
		assert.Len(t, a[0].Array.Attributes, 3)
	})

	t.Run("bare_ident_is_boolean_flag_not_positional", func(t *testing.T) {
		// An ident at the start is consumed as key=true (boolean flag), not positional.
		a := topAttrs(mustParse(t, "hello"))
		require.Len(t, a, 1)
		assert.Equal(t, "hello", a[0].Name)
		assert.Equal(t, ptr("true"), a[0].Value.Boolean)
	})

	t.Run("positional_then_named", func(t *testing.T) {
		a := topAttrs(mustParse(t, "42, name=foo"))
		require.Len(t, a, 2)
		assert.Equal(t, "", a[0].Name)
		assert.Equal(t, ptr("42"), a[0].Value.Number)
		assert.Equal(t, "name", a[1].Name)
		assert.Equal(t, ptr("foo"), a[1].Value.String)
	})

	t.Run("whitespace_around_tokens", func(t *testing.T) {
		a := topAttrs(mustParse(t, "  a  =  1  ,  b  =  2  "))
		require.Len(t, a, 2)
		assert.Equal(t, "a", a[0].Name)
		assert.Equal(t, ptr("1"), a[0].Value.Number)
		assert.Equal(t, "b", a[1].Name)
	})

	t.Run("top_level_span_starts_at_zero", func(t *testing.T) {
		got := mustParse(t, "a=1")
		require.NotNil(t, got.Span)
		assert.Equal(t, 0, got.Span.Position)
	})

	t.Run("error_value_after_equals", func(t *testing.T) {
		_, err := Parse(strings.NewReader("a=,"))
		assert.Error(t, err)
	})

	t.Run("error_unclosed_array", func(t *testing.T) {
		_, err := Parse(strings.NewReader("a[1, 2"))
		assert.Error(t, err)
	})

	t.Run("error_double_comma", func(t *testing.T) {
		_, err := Parse(strings.NewReader("a=1,, b=2"))
		assert.Error(t, err)
	})

	t.Run("error_comma_only", func(t *testing.T) {
		_, err := Parse(strings.NewReader(","))
		assert.Error(t, err)
	})

	t.Run("error_unclosed_nested_array", func(t *testing.T) {
		_, err := Parse(strings.NewReader("a[[1, 2]"))
		assert.Error(t, err)
	})

	// ── additional edge cases ──────────────────────────────────────────────────

	t.Run("error_unclosed_object", func(t *testing.T) {
		_, err := Parse(strings.NewReader("a(b=1"))
		assert.Error(t, err)
	})

	t.Run("error_stray_close_bracket", func(t *testing.T) {
		_, err := Parse(strings.NewReader("a=1)"))
		assert.Error(t, err)
	})

	t.Run("error_stray_close_square_bracket", func(t *testing.T) {
		_, err := Parse(strings.NewReader("a=1]"))
		assert.Error(t, err)
	})

	t.Run("error_trailing_comma_top_level", func(t *testing.T) {
		_, err := Parse(strings.NewReader("a=1,"))
		assert.Error(t, err)
	})

	t.Run("error_equals_without_key", func(t *testing.T) {
		_, err := Parse(strings.NewReader("=1"))
		assert.Error(t, err)
	})

	t.Run("double_quoted_string_with_newline_escape", func(t *testing.T) {
		a := topAttrs(mustParse(t, `s="line1\nline2"`))
		require.Len(t, a, 1)
		require.NotNil(t, a[0].Value)
		assert.Equal(t, ptr("line1\nline2"), a[0].Value.String)
	})

	t.Run("single_quoted_string_with_escaped_quote", func(t *testing.T) {
		a := topAttrs(mustParse(t, `s='it\'s fine'`))
		require.Len(t, a, 1)
		require.NotNil(t, a[0].Value)
		assert.Equal(t, ptr("it's fine"), a[0].Value.String)
	})

	t.Run("key_with_underscore", func(t *testing.T) {
		a := topAttrs(mustParse(t, "my_key=1"))
		require.Len(t, a, 1)
		assert.Equal(t, "my_key", a[0].Name)
		assert.Equal(t, ptr("1"), a[0].Value.Number)
	})

	t.Run("key_with_leading_underscores", func(t *testing.T) {
		a := topAttrs(mustParse(t, "__private=true"))
		require.Len(t, a, 1)
		assert.Equal(t, "__private", a[0].Name)
	})

	t.Run("key_alphanumeric", func(t *testing.T) {
		a := topAttrs(mustParse(t, "field123=42"))
		require.Len(t, a, 1)
		assert.Equal(t, "field123", a[0].Name)
	})

	t.Run("float_with_leading_dot", func(t *testing.T) {
		a := topAttrs(mustParse(t, "f=.5"))
		require.Len(t, a, 1)
		assert.Equal(t, ptr(".5"), a[0].Value.Number)
	})

	t.Run("large_integer", func(t *testing.T) {
		a := topAttrs(mustParse(t, "n=1000000"))
		require.Len(t, a, 1)
		assert.Equal(t, ptr("1000000"), a[0].Value.Number)
	})

	t.Run("zero", func(t *testing.T) {
		a := topAttrs(mustParse(t, "n=0"))
		require.Len(t, a, 1)
		assert.Equal(t, ptr("0"), a[0].Value.Number)
	})

	t.Run("negative_float", func(t *testing.T) {
		a := topAttrs(mustParse(t, "v=-3.14"))
		require.Len(t, a, 1)
		assert.Equal(t, ptr("-3.14"), a[0].Value.Number)
	})

	t.Run("array_single_item", func(t *testing.T) {
		a := topAttrs(mustParse(t, "x[42]"))
		require.Len(t, a, 1)
		require.NotNil(t, a[0].Array)
		assert.Len(t, a[0].Array.Attributes, 1)
		assert.Equal(t, ptr("42"), a[0].Array.Attributes[0].Value.Number)
	})

	t.Run("array_boolean_items", func(t *testing.T) {
		a := topAttrs(mustParse(t, "flags[true, false, true]"))
		items := a[0].Array.Attributes
		require.Len(t, items, 3)
		assert.Equal(t, ptr("true"), items[0].Value.Boolean)
		assert.Equal(t, ptr("false"), items[1].Value.Boolean)
		assert.Equal(t, ptr("true"), items[2].Value.Boolean)
	})

	t.Run("object_with_multiple_flags", func(t *testing.T) {
		a := topAttrs(mustParse(t, "opts(a, b, c)"))
		require.Len(t, a, 1)
		inner := a[0].Object.Attributes
		require.Len(t, inner, 3)
		for i, name := range []string{"a", "b", "c"} {
			assert.Equal(t, name, inner[i].Name)
			assert.Equal(t, ptr("true"), inner[i].Value.Boolean)
		}
	})

	t.Run("deeply_nested_array_in_object", func(t *testing.T) {
		a := topAttrs(mustParse(t, "cfg(ids[1, 2, 3], name='x')"))
		require.Len(t, a, 1)
		inner := a[0].Object.Attributes
		require.Len(t, inner, 2)
		assert.Equal(t, "ids", inner[0].Name)
		require.NotNil(t, inner[0].Array)
		assert.Len(t, inner[0].Array.Attributes, 3)
		assert.Equal(t, "name", inner[1].Name)
		assert.Equal(t, ptr("x"), inner[1].Value.String)
	})

	t.Run("array_of_arrays_of_objects", func(t *testing.T) {
		a := topAttrs(mustParse(t, "m[[(x=1)], [(x=2)]]"))
		require.Len(t, a, 1)
		outer := a[0].Array.Attributes
		require.Len(t, outer, 2)
		inner0 := outer[0].Array.Attributes
		require.Len(t, inner0, 1)
		require.NotNil(t, inner0[0].Object)
		assert.Equal(t, ptr("1"), inner0[0].Object.Attributes[0].Value.Number)
	})

	t.Run("whitespace_inside_array", func(t *testing.T) {
		a := topAttrs(mustParse(t, "x[  1  ,  2  ,  3  ]"))
		require.NotNil(t, a[0].Array)
		assert.Len(t, a[0].Array.Attributes, 3)
	})

	t.Run("whitespace_inside_object", func(t *testing.T) {
		a := topAttrs(mustParse(t, "o(  a = 1  ,  b = 2  )"))
		require.NotNil(t, a[0].Object)
		assert.Len(t, a[0].Object.Attributes, 2)
	})

	t.Run("string_with_spaces_double_quoted", func(t *testing.T) {
		a := topAttrs(mustParse(t, `desc="hello world foo"`))
		require.Len(t, a, 1)
		assert.Equal(t, ptr("hello world foo"), a[0].Value.String)
	})

	t.Run("positional_string_then_number", func(t *testing.T) {
		a := topAttrs(mustParse(t, "'hello', 42"))
		require.Len(t, a, 2)
		assert.Equal(t, ptr("hello"), a[0].Value.String)
		assert.Equal(t, ptr("42"), a[1].Value.Number)
	})

	t.Run("many_nested_objects", func(t *testing.T) {
		a := topAttrs(mustParse(t, "a(b(c(d(e=1))))"))
		require.Len(t, a, 1)
		assert.Equal(t, "a", a[0].Name)
		b := a[0].Object.Attributes
		require.Len(t, b, 1)
		assert.Equal(t, "b", b[0].Name)
		c := b[0].Object.Attributes
		require.Len(t, c, 1)
		assert.Equal(t, "c", c[0].Name)
		d := c[0].Object.Attributes
		require.Len(t, d, 1)
		assert.Equal(t, "d", d[0].Name)
		e := d[0].Object.Attributes
		require.Len(t, e, 1)
		assert.Equal(t, "e", e[0].Name)
		assert.Equal(t, ptr("1"), e[0].Value.Number)
	})

	t.Run("empty_string_double_quoted", func(t *testing.T) {
		a := topAttrs(mustParse(t, `s=""`))
		require.Len(t, a, 1)
		assert.Equal(t, ptr(""), a[0].Value.String)
	})

	t.Run("empty_string_single_quoted", func(t *testing.T) {
		a := topAttrs(mustParse(t, "s=''"))
		require.Len(t, a, 1)
		assert.Equal(t, ptr(""), a[0].Value.String)
	})
}

// ─── TestMustParse ────────────────────────────────────────────────────────────

func TestMustParse(t *testing.T) {
	t.Run("valid_input_does_not_panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			result := MustParse(strings.NewReader("a=1"))
			assert.NotNil(t, result)
		})
	})

	t.Run("invalid_input_panics", func(t *testing.T) {
		assert.Panics(t, func() {
			MustParse(strings.NewReader("a=,"))
		})
	})
}

// ─── TestValue ────────────────────────────────────────────────────────────────

func TestValue(t *testing.T) {
	t.Run("IsZero", func(t *testing.T) {
		assert.True(t, (&Value{}).IsZero())
		assert.False(t, (&Value{String: ptr("x")}).IsZero())
		assert.False(t, (&Value{Number: ptr("1")}).IsZero())
		assert.False(t, (&Value{Boolean: ptr("true")}).IsZero())
	})

	t.Run("ClearAll", func(t *testing.T) {
		v := &Value{String: ptr("x"), Number: ptr("1"), Boolean: ptr("true")}
		v.ClearAll()
		assert.True(t, v.IsZero())
	})

	t.Run("FromRaw_string", func(t *testing.T) {
		v := &Value{}
		require.NoError(t, v.FromRaw(ValueTypeString, "hello"))
		assert.Equal(t, ptr("hello"), v.String)
		assert.Nil(t, v.Number)
		assert.Nil(t, v.Boolean)
	})

	t.Run("FromRaw_number", func(t *testing.T) {
		v := &Value{}
		require.NoError(t, v.FromRaw(ValueTypeNumber, "42"))
		assert.Equal(t, ptr("42"), v.Number)
		assert.Nil(t, v.String)
	})

	t.Run("FromRaw_boolean_valid", func(t *testing.T) {
		v := &Value{}
		require.NoError(t, v.FromRaw(ValueTypeBoolean, "true"))
		assert.Equal(t, ptr("true"), v.Boolean)
	})

	t.Run("FromRaw_boolean_invalid", func(t *testing.T) {
		v := &Value{}
		err := v.FromRaw(ValueTypeBoolean, "notabool")
		assert.Error(t, err)
		assert.Nil(t, v.Boolean)
	})

	t.Run("FromRaw_combined_mask", func(t *testing.T) {
		v := &Value{}
		require.NoError(t, v.FromRaw(ValueTypeString|ValueTypeBoolean, "true"))
		assert.Equal(t, ptr("true"), v.String)
		assert.Equal(t, ptr("true"), v.Boolean)
	})

	t.Run("AsBool_from_boolean_field", func(t *testing.T) {
		b, err := (&Value{Boolean: ptr("true")}).AsBool()
		require.NoError(t, err)
		assert.True(t, b)

		b, err = (&Value{Boolean: ptr("false")}).AsBool()
		require.NoError(t, err)
		assert.False(t, b)
	})

	t.Run("AsBool_zero_value_returns_true", func(t *testing.T) {
		b, err := (&Value{}).AsBool()
		require.NoError(t, err)
		assert.True(t, b)
	})

	t.Run("AsBool_non_boolean_string_errors", func(t *testing.T) {
		_, err := (&Value{String: ptr("hello")}).AsBool()
		assert.Error(t, err)
	})

	t.Run("AsInt_valid", func(t *testing.T) {
		n, err := (&Value{Number: ptr("123")}).AsInt()
		require.NoError(t, err)
		assert.Equal(t, 123, n)
	})

	t.Run("AsInt_negative", func(t *testing.T) {
		n, err := (&Value{Number: ptr("-5")}).AsInt()
		require.NoError(t, err)
		assert.Equal(t, -5, n)
	})

	t.Run("AsInt_no_number_errors", func(t *testing.T) {
		_, err := (&Value{String: ptr("hello")}).AsInt()
		assert.Error(t, err)

		_, err = (&Value{}).AsInt()
		assert.Error(t, err)
	})

	t.Run("AsInt_float_errors", func(t *testing.T) {
		_, err := (&Value{Number: ptr("3.14")}).AsInt()
		assert.Error(t, err)
	})

	t.Run("AsFloat_valid", func(t *testing.T) {
		f, err := (&Value{Number: ptr("3.14")}).AsFloat()
		require.NoError(t, err)
		assert.InDelta(t, 3.14, f, 0.001)
	})

	t.Run("AsFloat_integer_as_float", func(t *testing.T) {
		f, err := (&Value{Number: ptr("42")}).AsFloat()
		require.NoError(t, err)
		assert.InDelta(t, 42.0, f, 0.001)
	})

	t.Run("AsFloat_no_number_errors", func(t *testing.T) {
		_, err := (&Value{}).AsFloat()
		assert.Error(t, err)
	})

	t.Run("AsString_valid", func(t *testing.T) {
		s, err := (&Value{String: ptr("hello")}).AsString()
		require.NoError(t, err)
		assert.Equal(t, "hello", s)
	})

	t.Run("AsString_no_string_errors", func(t *testing.T) {
		_, err := (&Value{Number: ptr("1")}).AsString()
		assert.Error(t, err)

		_, err = (&Value{}).AsString()
		assert.Error(t, err)
	})

	t.Run("AsTrimmedString_trims_spaces", func(t *testing.T) {
		s, err := (&Value{String: ptr("  hello world  ")}).AsTrimmedString()
		require.NoError(t, err)
		assert.Equal(t, "hello world", s)
	})

	t.Run("AsTrimmedString_no_string_errors", func(t *testing.T) {
		_, err := (&Value{}).AsTrimmedString()
		assert.Error(t, err)
	})

	t.Run("IsBool_IsInt_IsFloat_IsString", func(t *testing.T) {
		vBool := &Value{Boolean: ptr("true")}
		assert.True(t, vBool.IsBool())
		assert.False(t, vBool.IsInt())
		assert.False(t, vBool.IsFloat())
		assert.False(t, vBool.IsString())

		vNum := &Value{Number: ptr("42")}
		assert.True(t, vNum.IsInt())
		assert.True(t, vNum.IsFloat())
		assert.False(t, vNum.IsBool())
		assert.False(t, vNum.IsString())

		vStr := &Value{String: ptr("x")}
		assert.True(t, vStr.IsString())
		assert.False(t, vStr.IsBool())

		// IsZero → AsBool returns true → IsBool is true
		vZero := &Value{}
		assert.True(t, vZero.IsBool())
	})

	t.Run("BuildValue_string", func(t *testing.T) {
		rv, err := (&Value{String: ptr("world")}).BuildValue()
		require.NoError(t, err)
		assert.Equal(t, "world", rv.Interface().(string))
	})

	t.Run("BuildValue_integer", func(t *testing.T) {
		rv, err := (&Value{Number: ptr("7")}).BuildValue()
		require.NoError(t, err)
		assert.Equal(t, int(7), rv.Interface().(int))
	})

	t.Run("BuildValue_float", func(t *testing.T) {
		rv, err := (&Value{Number: ptr("2.5")}).BuildValue()
		require.NoError(t, err)
		assert.InDelta(t, 2.5, rv.Interface().(float64), 0.001)
	})

	t.Run("BuildValue_boolean", func(t *testing.T) {
		rv, err := (&Value{Boolean: ptr("true")}).BuildValue()
		require.NoError(t, err)
		assert.True(t, rv.Interface().(bool))

		rv, err = (&Value{Boolean: ptr("false")}).BuildValue()
		require.NoError(t, err)
		assert.False(t, rv.Interface().(bool))
	})

	t.Run("BuildValue_no_value_errors", func(t *testing.T) {
		_, err := (&Value{}).BuildValue()
		assert.Error(t, err)
	})

	t.Run("BuildValue_invalid_boolean_errors", func(t *testing.T) {
		_, err := (&Value{Boolean: ptr("notbool")}).BuildValue()
		assert.Error(t, err)
	})
}

// ─── TestValueType ────────────────────────────────────────────────────────────

func TestValueType(t *testing.T) {
	t.Run("FromRaw_string_only", func(t *testing.T) {
		v := &Value{}
		require.NoError(t, ValueTypeString.FromRaw("hello", v))
		assert.Equal(t, ptr("hello"), v.String)
		assert.Nil(t, v.Number)
		assert.Nil(t, v.Boolean)
	})

	t.Run("FromRaw_number_only", func(t *testing.T) {
		v := &Value{}
		require.NoError(t, ValueTypeNumber.FromRaw("99", v))
		assert.Equal(t, ptr("99"), v.Number)
	})

	t.Run("FromRaw_boolean_true", func(t *testing.T) {
		v := &Value{}
		require.NoError(t, ValueTypeBoolean.FromRaw("true", v))
		assert.Equal(t, ptr("true"), v.Boolean)
	})

	t.Run("FromRaw_boolean_false", func(t *testing.T) {
		v := &Value{}
		require.NoError(t, ValueTypeBoolean.FromRaw("false", v))
		assert.Equal(t, ptr("false"), v.Boolean)
	})

	t.Run("FromRaw_boolean_invalid_returns_error", func(t *testing.T) {
		v := &Value{}
		err := ValueTypeBoolean.FromRaw("oops", v)
		assert.Error(t, err)
		assert.Nil(t, v.Boolean)
	})

	t.Run("FromRaw_all_types", func(t *testing.T) {
		v := &Value{}
		require.NoError(t, (ValueTypeString|ValueTypeNumber).FromRaw("42", v))
		assert.Equal(t, ptr("42"), v.String)
		assert.Equal(t, ptr("42"), v.Number)
		assert.Nil(t, v.Boolean)
	})
}

// ─── TestAttribute ────────────────────────────────────────────────────────────

func TestAttribute(t *testing.T) {
	span := newSourceSpan(0)

	t.Run("HasValue_with_value", func(t *testing.T) {
		a := &Attribute{Span: span, Value: &Value{String: ptr("x")}}
		assert.True(t, a.HasValue())
	})

	t.Run("HasValue_with_object", func(t *testing.T) {
		a := &Attribute{Span: span, Object: newAttributes(span)}
		assert.True(t, a.HasValue())
	})

	t.Run("HasValue_with_array", func(t *testing.T) {
		a := &Attribute{Span: span, Array: newAttributes(span)}
		assert.True(t, a.HasValue())
	})

	t.Run("HasValue_empty", func(t *testing.T) {
		a := &Attribute{Span: span}
		assert.False(t, a.HasValue())
	})

	t.Run("Build_string", func(t *testing.T) {
		a := &Attribute{Span: span, Value: &Value{String: ptr("hi")}}
		rv, err := a.Build()
		require.NoError(t, err)
		assert.Equal(t, "hi", rv.Interface().(string))
	})

	t.Run("Build_object_to_map", func(t *testing.T) {
		a := &Attribute{
			Span: span,
			Object: newAttributes(span,
				&Attribute{Name: "key", Span: span, Value: &Value{String: ptr("val")}},
				&Attribute{Name: "n", Span: span, Value: &Value{Number: ptr("1")}},
			),
		}
		rv, err := a.Build()
		require.NoError(t, err)
		m, ok := rv.Interface().(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "val", m["key"])
		assert.Equal(t, int(1), m["n"])
	})

	t.Run("Build_array_to_slice", func(t *testing.T) {
		a := &Attribute{
			Span: span,
			Array: newAttributes(span,
				&Attribute{Span: span, Value: &Value{Number: ptr("10")}},
				&Attribute{Span: span, Value: &Value{String: ptr("str")}},
			),
		}
		rv, err := a.Build()
		require.NoError(t, err)
		sl, ok := rv.Interface().([]any)
		require.True(t, ok)
		require.Len(t, sl, 2)
		assert.Equal(t, int(10), sl[0])
		assert.Equal(t, "str", sl[1])
	})

	t.Run("Build_no_value_errors", func(t *testing.T) {
		a := &Attribute{Span: span}
		_, err := a.Build()
		assert.Error(t, err)
	})

	t.Run("Build_nested_object", func(t *testing.T) {
		inner := &Attribute{Name: "x", Span: span, Value: &Value{Boolean: ptr("true")}}
		outer := &Attribute{
			Span:   span,
			Object: newAttributes(span, inner),
		}
		rv, err := outer.Build()
		require.NoError(t, err)
		m := rv.Interface().(map[string]any)
		assert.True(t, m["x"].(bool))
	})

	t.Run("PrepareAny_string", func(t *testing.T) {
		a := &Attribute{Span: span, Value: &Value{String: ptr("x")}}
		rv, err := a.PrepareAny()
		require.NoError(t, err)
		assert.Equal(t, reflect.TypeOf(""), rv.Type())
	})

	t.Run("PrepareAny_integer", func(t *testing.T) {
		a := &Attribute{Span: span, Value: &Value{Number: ptr("5")}}
		rv, err := a.PrepareAny()
		require.NoError(t, err)
		assert.Equal(t, reflect.TypeOf(int64(0)), rv.Type())
	})

	t.Run("PrepareAny_float", func(t *testing.T) {
		a := &Attribute{Span: span, Value: &Value{Number: ptr("3.14")}}
		rv, err := a.PrepareAny()
		require.NoError(t, err)
		assert.Equal(t, reflect.TypeOf(float64(0)), rv.Type())
	})

	t.Run("PrepareAny_boolean", func(t *testing.T) {
		a := &Attribute{Span: span, Value: &Value{Boolean: ptr("true")}}
		rv, err := a.PrepareAny()
		require.NoError(t, err)
		assert.Equal(t, reflect.TypeOf(false), rv.Type())
	})

	t.Run("PrepareAny_object_returns_zero_reflect_value", func(t *testing.T) {
		a := &Attribute{Span: span, Object: newAttributes(span)}
		rv, err := a.PrepareAny()
		require.NoError(t, err)
		assert.False(t, rv.IsValid())
	})

	t.Run("PrepareAny_array_returns_zero_reflect_value", func(t *testing.T) {
		a := &Attribute{Span: span, Array: newAttributes(span)}
		rv, err := a.PrepareAny()
		require.NoError(t, err)
		assert.False(t, rv.IsValid())
	})

	t.Run("PrepareAny_no_value_errors", func(t *testing.T) {
		a := &Attribute{Span: span}
		_, err := a.PrepareAny()
		assert.Error(t, err)
	})
}

// ─── TestAttributes ───────────────────────────────────────────────────────────

func TestAttributes(t *testing.T) {
	span := newSourceSpan(0)

	t.Run("newAttributes_empty", func(t *testing.T) {
		a := newAttributes(span)
		assert.NotNil(t, a)
		assert.Empty(t, a.Attributes)
		assert.Equal(t, span, a.Span)
	})

	t.Run("newAttributes_with_items", func(t *testing.T) {
		child := &Attribute{Name: "x", Span: span}
		a := newAttributes(span, child)
		require.Len(t, a.Attributes, 1)
		assert.Equal(t, child, a.Attributes[0])
	})

	t.Run("Push_single", func(t *testing.T) {
		a := newAttributes(span)
		child := &Attribute{Name: "a", Span: span}
		a.Push(child)
		require.Len(t, a.Attributes, 1)
		assert.Equal(t, child, a.Attributes[0])
	})

	t.Run("Push_multiple", func(t *testing.T) {
		a := newAttributes(span)
		c1 := &Attribute{Name: "a", Span: span}
		c2 := &Attribute{Name: "b", Span: span}
		a.Push(c1, c2)
		assert.Len(t, a.Attributes, 2)
	})

	t.Run("Push_returns_receiver_for_chaining", func(t *testing.T) {
		a := newAttributes(span)
		ret := a.Push(&Attribute{Name: "x", Span: span})
		assert.Same(t, a, ret)
	})
}

// ─── TestValidateIdentifier ───────────────────────────────────────────────────

func TestValidateIdentifier(t *testing.T) {
	valid := []string{
		"hello",
		"Hello",
		"Hello123",
		"hello_world",
		"_hello",
		"__hello",
		"_h1",
		"a",
		"z9",
	}
	for _, id := range valid {
		t.Run("valid_"+id, func(t *testing.T) {
			assert.NoError(t, ValidateIdentifier(id), "expected %q to be valid", id)
		})
	}

	invalid := []string{
		"",            // empty
		"_",           // underscore only, no letter
		"__",          // underscores only
		"_123",        // underscore then digits, no letter
		"hello-world", // hyphen not allowed
		"123abc",      // starts with digit
		"hello world", // space not allowed
		"-name",       // leading hyphen
		"a b",         // space inside
	}
	for _, id := range invalid {
		t.Run("invalid_"+id, func(t *testing.T) {
			assert.Error(t, ValidateIdentifier(id), "expected %q to be invalid", id)
		})
	}
}

// ─── TestMatchToken ───────────────────────────────────────────────────────────

func TestMatchToken(t *testing.T) {
	t.Run("matches_single_token", func(t *testing.T) {
		m := MatchToken(TokenIdent)
		assert.True(t, m.Match(&ParserItem{Token: TokenIdent}))
		assert.False(t, m.Match(&ParserItem{Token: TokenEqual}))
	})

	t.Run("matches_any_of_multiple_tokens", func(t *testing.T) {
		m := MatchToken(TokenIdent, TokenEqual, TokenNumber)
		assert.True(t, m.Match(&ParserItem{Token: TokenIdent}))
		assert.True(t, m.Match(&ParserItem{Token: TokenEqual}))
		assert.True(t, m.Match(&ParserItem{Token: TokenNumber}))
		assert.False(t, m.Match(&ParserItem{Token: TokenComma}))
		assert.False(t, m.Match(&ParserItem{Token: TokenEOF}))
	})

	t.Run("empty_token_list_never_matches", func(t *testing.T) {
		m := MatchToken()
		assert.False(t, m.Match(&ParserItem{Token: TokenIdent}))
		assert.False(t, m.Match(&ParserItem{Token: TokenEOF}))
	})
}

// ─── TestMatchAll ─────────────────────────────────────────────────────────────

func TestMatchAll(t *testing.T) {
	t.Run("all_matchers_pass", func(t *testing.T) {
		identMatcher := MatchToken(TokenIdent)
		valueMatcher := MatcherFunc(func(pi *ParserItem) bool { return pi.Value == "hello" })
		m := MatchAll(identMatcher, valueMatcher)
		assert.True(t, m.Match(&ParserItem{Token: TokenIdent, Value: "hello"}))
		assert.False(t, m.Match(&ParserItem{Token: TokenIdent, Value: "world"}))
		assert.False(t, m.Match(&ParserItem{Token: TokenEqual, Value: "hello"}))
	})

	t.Run("nil_matchers_are_skipped", func(t *testing.T) {
		m := MatchAll(nil, MatchToken(TokenIdent), nil)
		assert.True(t, m.Match(&ParserItem{Token: TokenIdent}))
		assert.False(t, m.Match(&ParserItem{Token: TokenComma}))
	})

	t.Run("empty_matcher_list_always_passes", func(t *testing.T) {
		m := MatchAll()
		assert.True(t, m.Match(&ParserItem{Token: TokenIdent}))
		assert.True(t, m.Match(&ParserItem{Token: TokenEOF}))
	})

	t.Run("all_nil_always_passes", func(t *testing.T) {
		m := MatchAll(nil, nil)
		assert.True(t, m.Match(&ParserItem{Token: TokenComma}))
	})
}

// ─── TestMatcherFunc ─────────────────────────────────────────────────────────

func TestMatcherFunc(t *testing.T) {
	t.Run("calls_underlying_function", func(t *testing.T) {
		called := false
		f := MatcherFunc(func(pi *ParserItem) bool {
			called = true
			return pi.Value == "expected"
		})
		assert.True(t, f.Match(&ParserItem{Value: "expected"}))
		assert.True(t, called)
		assert.False(t, f.Match(&ParserItem{Value: "other"}))
	})
}

// ─── TestParserMatch (internal p.match) ──────────────────────────────────────

func TestParserMatch(t *testing.T) {
	t.Run("matches_sequence", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("a=1"))}
		items, err := p.match(MatchToken(TokenIdent), MatchToken(TokenEqual))
		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.Equal(t, TokenIdent, items[0].Token)
		assert.Equal(t, "a", items[0].Value)
		assert.Equal(t, TokenEqual, items[1].Token)
	})

	t.Run("rolls_back_on_mismatch", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("a=1"))}
		// Attempt ident+ident — will fail on "=" (not ident).
		_, err := p.match(MatchToken(TokenIdent), MatchToken(TokenIdent))
		assert.Error(t, err)
		// After failed match, lexer is rolled back — try again successfully.
		items, err := p.match(MatchToken(TokenIdent), MatchToken(TokenEqual))
		require.NoError(t, err)
		require.Len(t, items, 2)
	})

	t.Run("empty_matchers_succeeds_with_no_items", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("a=1"))}
		items, err := p.match()
		require.NoError(t, err)
		assert.Empty(t, items)
	})
}

// ─── TestParseError ───────────────────────────────────────────────────────────

func TestParseError(t *testing.T) {
	span := newSourceSpan(5)

	t.Run("Error_includes_position_and_message", func(t *testing.T) {
		err := NewParseError(span, "something went wrong")
		assert.Contains(t, err.Error(), "something went wrong")
		assert.Contains(t, err.Error(), "5") // position
	})

	t.Run("Error_formats_args", func(t *testing.T) {
		err := NewParseError(span, "bad token %q at %d", "=", 5)
		assert.Contains(t, err.Error(), `"="`)
		assert.Contains(t, err.Error(), "5")
	})

	t.Run("Position_returns_span_position", func(t *testing.T) {
		err := NewParseError(span, "oops")
		pe, ok := err.(ParseError)
		require.True(t, ok)
		assert.Equal(t, 5, pe.Position())
	})

	t.Run("nil_span_does_not_panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			err := NewParseError(nil, "message")
			_ = err.Error()
		})
	})
}

// ─── TestErrIsNoMatch ─────────────────────────────────────────────────────────

func TestErrIsNoMatch(t *testing.T) {
	assert.False(t, ErrIsNoMatch(nil))
	assert.False(t, ErrIsNoMatch(io.EOF))
	assert.False(t, ErrIsNoMatch(errors.New("some error")))
	assert.False(t, ErrIsNoMatch(ErrNotValue))

	assert.True(t, ErrIsNoMatch(ErrNoMatch))
	assert.True(t, ErrIsNoMatch(fmt.Errorf("wrapped: %w", ErrNoMatch)))
}

// ─── TestToken ────────────────────────────────────────────────────────────────

func TestToken(t *testing.T) {
	t.Run("IsError", func(t *testing.T) {
		assert.True(t, TokenError.IsError())
		assert.False(t, TokenIdent.IsError())
		assert.False(t, TokenEOF.IsError())
	})

	t.Run("AsError_on_error_token", func(t *testing.T) {
		span := newSourceSpan(0)
		err := TokenError.AsError(span, "bad input")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bad input")
	})

	t.Run("AsError_on_non_error_token_returns_nil", func(t *testing.T) {
		assert.Nil(t, TokenIdent.AsError(newSourceSpan(0), ""))
	})

	t.Run("OneOf_single", func(t *testing.T) {
		assert.True(t, TokenIdent.OneOf(TokenIdent))
		assert.False(t, TokenIdent.OneOf(TokenEqual))
	})

	t.Run("OneOf_multiple", func(t *testing.T) {
		assert.True(t, TokenComma.OneOf(TokenIdent, TokenComma, TokenEOF))
		assert.False(t, TokenComma.OneOf(TokenIdent, TokenEqual))
	})

	t.Run("OneOf_empty_list", func(t *testing.T) {
		assert.False(t, TokenIdent.OneOf())
	})
}

// ─── TestSourceSpan ───────────────────────────────────────────────────────────

func TestSourceSpan(t *testing.T) {
	t.Run("newSourceSpan_sets_position", func(t *testing.T) {
		s := newSourceSpan(7)
		assert.Equal(t, 7, s.Position)
		assert.Equal(t, 0, s.Length)
	})

	t.Run("newSourceSpan_with_length", func(t *testing.T) {
		s := newSourceSpan(3, 5)
		assert.Equal(t, 3, s.Position)
		assert.Equal(t, 5, s.Length)
	})

	t.Run("withLength", func(t *testing.T) {
		s := newSourceSpan(2).withLength(10)
		assert.Equal(t, 2, s.Position)
		assert.Equal(t, 10, s.Length)
	})

	t.Run("withPosition", func(t *testing.T) {
		s := newSourceSpan(0, 5).withPosition(9)
		assert.Equal(t, 9, s.Position)
		assert.Equal(t, 5, s.Length)
	})

	t.Run("withLengthFromPosition_computes_length", func(t *testing.T) {
		s := newSourceSpan(3).withLengthFromPosition(8)
		assert.Equal(t, 3, s.Position)
		assert.Equal(t, 5, s.Length) // 8-3
	})

	t.Run("withLengthFromPosition_same_pos_gives_zero_length", func(t *testing.T) {
		s := newSourceSpan(4).withLengthFromPosition(4)
		assert.Equal(t, 0, s.Length)
	})

	t.Run("incrLength_adds_one", func(t *testing.T) {
		s := newSourceSpan(0, 3).incrLength()
		assert.Equal(t, 4, s.Length)
	})

	t.Run("incrLengthBy_adds_n", func(t *testing.T) {
		s := newSourceSpan(0, 3).incrLengthBy(5)
		assert.Equal(t, 8, s.Length)
	})

	t.Run("String_format", func(t *testing.T) {
		s := newSourceSpan(4, 6)
		str := s.String()
		assert.Contains(t, str, "4")
		assert.Contains(t, str, "6")
	})

	t.Run("immutability_original_unchanged", func(t *testing.T) {
		original := newSourceSpan(1, 2)
		_ = original.withLength(99)
		_ = original.withPosition(99)
		_ = original.withLengthFromPosition(99)
		assert.Equal(t, 1, original.Position)
		assert.Equal(t, 2, original.Length)
	})
}

// ─── TestParserItem ───────────────────────────────────────────────────────────

func TestParserItem(t *testing.T) {
	t.Run("Rollback_restores_lexer_position", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("ab"))}
		pi, err := lexParserItem(p)
		require.NoError(t, err)
		assert.Equal(t, TokenIdent, pi.Token)
		// Rollback then re-lex should give the same token.
		require.NoError(t, pi.Rollback())
		pi2, err := lexParserItem(p)
		require.NoError(t, err)
		assert.Equal(t, TokenIdent, pi2.Token)
		assert.Equal(t, "ab", pi2.Value)
	})

	t.Run("newParserItem_error_token_returns_error", func(t *testing.T) {
		_, err := newParserItem(newSourceSpan(0), TokenError, "something bad")
		assert.Error(t, err)
	})

	t.Run("newParserItem_valid_token_succeeds", func(t *testing.T) {
		pi, err := newParserItem(newSourceSpan(0), TokenIdent, "hello")
		require.NoError(t, err)
		assert.Equal(t, TokenIdent, pi.Token)
		assert.Equal(t, "hello", pi.Value)
	})

	t.Run("Rollback_nil_snapshot_returns_error", func(t *testing.T) {
		pi := &ParserItem{Token: TokenIdent}
		assert.Error(t, pi.Rollback())
	})
}

// ─── TestParserLex (Lex / Unlex / Peek / LexSelected / currentPos) ──────────

func TestParserLex(t *testing.T) {
	t.Run("Lex_returns_tokens_in_order", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("a=1"))}
		_, tok1, v1 := p.Lex()
		_, tok2, _ := p.Lex()
		_, tok3, v3 := p.Lex()
		assert.Equal(t, TokenIdent, tok1)
		assert.Equal(t, "a", v1)
		assert.Equal(t, TokenEqual, tok2)
		assert.Equal(t, TokenNumber, tok3)
		assert.Equal(t, "1", v3)
	})

	t.Run("Unlex_pushes_token_back", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("a=1"))}
		span, tok, val := p.Lex()
		assert.Equal(t, TokenIdent, tok)
		p.Unlex(span, tok, val)
		// Re-read: should get same token.
		_, tok2, val2 := p.Lex()
		assert.Equal(t, TokenIdent, tok2)
		assert.Equal(t, "a", val2)
	})

	t.Run("Unlex_stacks_multiple_tokens", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("a=1"))}
		s1, t1, v1 := p.Lex()
		s2, t2, v2 := p.Lex()
		p.Unlex(s2, t2, v2)
		p.Unlex(s1, t1, v1)
		// LIFO: last unlex'd is first out.
		_, tok, val := p.Lex()
		assert.Equal(t, t1, tok)
		assert.Equal(t, v1, val)
	})

	t.Run("Peek_does_not_consume_token", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("hello"))}
		_, tok1, v1 := p.Peek()
		_, tok2, v2 := p.Peek()
		assert.Equal(t, tok1, tok2)
		assert.Equal(t, v1, v2)
		// After two peeks, Lex should still return the same token.
		_, tok3, v3 := p.Lex()
		assert.Equal(t, tok1, tok3)
		assert.Equal(t, v1, v3)
	})

	t.Run("LexSelected_returns_matching_token", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("a=1"))}
		_, tok, val, err := p.LexSelected(TokenIdent, TokenEqual)
		require.NoError(t, err)
		assert.Equal(t, TokenIdent, tok)
		assert.Equal(t, "a", val)
	})

	t.Run("LexSelected_unlex_on_mismatch", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("a=1"))}
		_, _, _, err := p.LexSelected(TokenEqual) // "a" is ident, not equal
		assert.ErrorIs(t, err, ErrNoMatch)
		// Token was put back — next Lex should return "a".
		_, tok, val := p.Lex()
		assert.Equal(t, TokenIdent, tok)
		assert.Equal(t, "a", val)
	})

	t.Run("currentPos_with_empty_peekObjs_uses_lexer", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("abc"))}
		assert.Equal(t, 0, p.currentPos())
		p.Lex()
		assert.Greater(t, p.currentPos(), 0)
	})

	t.Run("currentPos_with_peekObjs_uses_topmost_span", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("a=1"))}
		span, tok, val := p.Lex()
		startPos := span.Position
		p.Unlex(span, tok, val)
		assert.Equal(t, startPos, p.currentPos())
	})
}

// ─── TestPeekV2Match ─────────────────────────────────────────────────────────

func TestPeekV2Match(t *testing.T) {
	t.Run("returns_true_when_token_matches", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("hello"))}
		assert.True(t, p.peekV2Match(TokenIdent))
	})

	t.Run("returns_false_when_token_does_not_match", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("hello"))}
		assert.False(t, p.peekV2Match(TokenEqual))
	})

	t.Run("consumes_token_on_match_rolls_back_on_mismatch", func(t *testing.T) {
		// peekV2Match consumes the token when it matches, rolls back when it doesn't.
		p := &parser{lexer: newLexer(strings.NewReader("a=1"))}
		assert.True(t, p.peekV2Match(TokenIdent))  // "a" consumed
		assert.False(t, p.peekV2Match(TokenIdent)) // "=" doesn't match — rolled back
		// Next lex sees "=" (not consumed by failed peek).
		pi, err := p.lexV2()
		require.NoError(t, err)
		assert.Equal(t, TokenEqual, pi.Token)
	})
}

// ─── TestParserSnapshot ───────────────────────────────────────────────────────

func TestParserSnapshot(t *testing.T) {
	t.Run("Snapshot_and_rollback_via_parser", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("a=1"))}
		snap := p.Snapshot()
		_, tok1, v1 := p.Lex()
		assert.Equal(t, TokenIdent, tok1)
		// Rollback via the Snapshot helper.
		require.NoError(t, snap.Rollback(p))
		_, tok2, v2 := p.Lex()
		assert.Equal(t, tok1, tok2)
		assert.Equal(t, v1, v2)
	})
}

// ─── TestLexerV2Match ────────────────────────────────────────────────────────

func TestLexerV2Match(t *testing.T) {
	t.Run("lexV2Match_returns_matching_item", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("hello=42"))}
		pi, err := p.lexV2Match(TokenIdent)
		require.NoError(t, err)
		assert.Equal(t, "hello", pi.Value)
	})

	t.Run("lexV2Match_rolls_back_on_mismatch_and_returns_ErrNoMatch", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("hello=42"))}
		_, err := p.lexV2Match(TokenEqual) // "hello" is not "="
		assert.True(t, ErrIsNoMatch(err))
		// Rolled back — can read "hello" again.
		pi, err := p.lexV2Match(TokenIdent)
		require.NoError(t, err)
		assert.Equal(t, "hello", pi.Value)
	})

	t.Run("lexV2MatchMap_wraps_no_match_with_origin_error", func(t *testing.T) {
		p := &parser{lexer: newLexer(strings.NewReader("hello"))}
		_, err := p.lexV2MatchMap(ErrNotValue, TokenEqual)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotValue))
	})
}

// ─── Round-trip: Parse then Build ────────────────────────────────────────────

func TestParseAndBuild(t *testing.T) {
	t.Run("scalar_round_trip", func(t *testing.T) {
		a := topAttrs(mustParse(t, "n=7"))[0]
		rv, err := a.Value.BuildValue()
		require.NoError(t, err)
		assert.Equal(t, int(7), rv.Interface().(int))
	})

	t.Run("object_round_trip_to_map", func(t *testing.T) {
		a := topAttrs(mustParse(t, "meta(x=1, y=2)"))[0]
		rv, err := a.Build()
		require.NoError(t, err)
		m := rv.Interface().(map[string]any)
		assert.Equal(t, int(1), m["x"])
		assert.Equal(t, int(2), m["y"])
	})

	t.Run("array_round_trip_to_slice", func(t *testing.T) {
		a := topAttrs(mustParse(t, "ids[10, 20, 30]"))[0]
		rv, err := a.Build()
		require.NoError(t, err)
		sl := rv.Interface().([]any)
		require.Len(t, sl, 3)
		assert.Equal(t, int(10), sl[0])
	})
}
