package attribs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func ptr[T any](v T) *T {
	return &v
}

// AttrDef defines attribute for all attributes tests
type AttrDef struct {
	ID         int       `attr:"name=id"`
	CategoryID *int      `attr:"name=category_id"`
	Disabled   bool      `attr:"name=disabled"`
	String     string    `attr:"name=string"`
	StringOpt  *string   `attr:"name=string_opt"`
	Interval   *Interval `attr:"name=interval"`
	Interval2  Interval  `attr:"name=interval2"`
	F32        float32   `attr:"name=f32"`
	F64        float64   `attr:"name=f64"`
	Uint       uint      `attr:"name=uint"`
	UintOpt    *uint     `attr:"name=uint_opt"`
}

type Interval struct {
	Start int `attr:"name=start"`
	End   int `attr:"name=end"`
}

func TestAttrs(t *testing.T) {
	t.Run("test pointer to struct", func(t *testing.T) {
		type Struct struct {
			Hello string `attr:"name=hello"`
		}

		d, err := New(&Struct{})
		assert.NoError(t, err)
		assert.NotNil(t, d)

	})

	t.Run("test value", func(t *testing.T) {
		for _, item := range []struct {
			input       string
			expected    AttrDef
			errExpected string
		}{
			{
				input:    "id=42, category_id=64, string='hello world'",
				expected: AttrDef{ID: 42, CategoryID: ptr(64), String: "hello world"},
			},
			{
				input:    "id=42, category_id=64, string_opt='hello world', uint=99",
				expected: AttrDef{ID: 42, CategoryID: ptr(64), StringOpt: ptr("hello world"), Uint: 99},
			},
			{
				input:    "id=42, category_id=64, string_opt='hello world', uint_opt=99",
				expected: AttrDef{ID: 42, CategoryID: ptr(64), StringOpt: ptr("hello world"), UintOpt: ptr(uint(99))},
			},
			{
				input:    "id=42 , category_id = 44,   disabled",
				expected: AttrDef{ID: 42, CategoryID: ptr(44), Disabled: true},
			},
			{
				input:    "",
				expected: AttrDef{},
			},
			{
				input: "string = 'hello world'",
				expected: AttrDef{
					String: "hello world",
				},
			},
			{
				input: "id=1, interval(start=998, end=65535), string_opt=foo, uint=42",
				expected: AttrDef{
					ID: 1,
					Interval: &Interval{
						Start: 998,
						End:   65535,
					},
					StringOpt: ptr("foo"),
					Uint:      42,
				},
			},
			{
				input: "id=1, string_opt=foo, uint=42",
				expected: AttrDef{
					ID:        1,
					StringOpt: ptr("foo"),
					Uint:      42,
				},
			},
			{
				input:    "id=42, category_id=64, interval(start=1, end=2)",
				expected: AttrDef{ID: 42, CategoryID: ptr(64), Interval: &Interval{Start: 1, End: 2}},
			},
			{
				input:    "id=42 , category_id = 44,   disabled",
				expected: AttrDef{ID: 42, CategoryID: ptr(44), Disabled: true},
			},
			{
				input:    "",
				expected: AttrDef{},
			},
			{
				input: "string = 'hello world'",
				expected: AttrDef{
					String: "hello world",
				},
			},
			{
				input: "id=1, interval(start=42, end=65535), string_opt=foo",
				expected: AttrDef{
					ID: 1,
					Interval: &Interval{
						Start: 42,
						End:   65535,
					},
					StringOpt: ptr("foo"),
				},
			},
			{
				input: "f32=1.2, f64=2.1",
				expected: AttrDef{
					F32: 1.2,
					F64: 2.1,
				},
			},
		} {
			defined, err := New(AttrDef{})
			assert.NoError(t, err)

			value, err := defined.Parse(item.input)
			if item.errExpected != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), item.errExpected)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, item.expected, value)
			}
		}
	})

	t.Run("test value", func(t *testing.T) {
		for _, item := range []struct {
			input       string
			expected    AttrDef
			errExpected string
		}{
			{
				input:    "id=42, category_id=64, interval(start=1, end=2)",
				expected: AttrDef{ID: 42, CategoryID: ptr(64), Interval: &Interval{Start: 1, End: 2}},
			},
			{
				input:    "id=42 , category_id = 44,   disabled",
				expected: AttrDef{ID: 42, CategoryID: ptr(44), Disabled: true},
			},
			{
				input:    "",
				expected: AttrDef{},
			},
			{
				input: "string = 'hello world'",
				expected: AttrDef{
					String: "hello world",
				},
			},
			{
				input: "id=1, interval(start=42, end=65535), string_opt=foo, uint=42",
				expected: AttrDef{
					ID: 1,
					Interval: &Interval{
						Start: 42,
						End:   65535,
					},
					StringOpt: ptr("foo"),
					Uint:      42,
				},
			},
			{
				input:    "id=42, category_id=64, interval(start=1, end=2)",
				expected: AttrDef{ID: 42, CategoryID: ptr(64), Interval: &Interval{Start: 1, End: 2}},
			},
			{
				input:    "id=42 , category_id = 44,   disabled",
				expected: AttrDef{ID: 42, CategoryID: ptr(44), Disabled: true},
			},
			{
				input:    "",
				expected: AttrDef{},
			},
			{
				input: "string = 'hello world'",
				expected: AttrDef{
					String: "hello world",
				},
			},
			{
				input: "id=1, interval(start=42, end=65535), string_opt=foo",
				expected: AttrDef{
					ID: 1,
					Interval: &Interval{
						Start: 42,
						End:   65535,
					},
					StringOpt: ptr("foo"),
				},
			},
			{
				input: "f32=1.2, f64=2.1",
				expected: AttrDef{
					F32: 1.2,
					F64: 2.1,
				},
			},
		} {
			defined, err := New(AttrDef{})
			assert.NoError(t, err)

			value, err := defined.Parse(item.input)
			if item.errExpected != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), item.errExpected)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, item.expected, value)
			}
		}
	})
	t.Run("test pointer", func(t *testing.T) {
		defined, err := New(AttrDef{})
		assert.NoError(t, err)

		for _, item := range []struct {
			input       string
			expected    AttrDef
			errExpected string
		}{
			{
				input:    "id=42, category_id=64, interval(start=1, end=2)",
				expected: AttrDef{ID: 42, CategoryID: ptr(64), Interval: &Interval{Start: 1, End: 2}},
			},

			{
				input:    "id=42 , category_id = 44,   disabled",
				expected: AttrDef{ID: 42, CategoryID: ptr(44), Disabled: true},
			},
			{
				input:    "",
				expected: AttrDef{},
			},
			{
				input: "id=1, interval(start=42, end=65535)",
				expected: AttrDef{
					ID: 1,
					Interval: &Interval{
						Start: 42,
						End:   65535,
					},
				},
			},
		} {
			value, err := defined.Parse(item.input)
			if item.errExpected != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), item.errExpected)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, item.expected, value)
			}
		}
	})

	t.Run("test array", func(t *testing.T) {
		type User struct {
			Username string `attr:"name=username"`
		}
		type Example struct {
			IDs   []int     `attr:"name=ids"`
			Users []*User   `attr:"name=users"`
			Enums *[]string `attr:"name=enums"`
		}

		for _, item := range []struct {
			input       string
			expected    Example
			errExpected string
		}{
			{input: "", expected: Example{}},
			{input: "ids[1,2,3], users[(username='me')]", expected: Example{IDs: []int{1, 2, 3}, Users: []*User{{Username: "me"}}}},
			{input: "enums['hello', 'world']", expected: Example{Enums: ptr([]string{"hello", "world"})}},
			{input: "enums['hello', 'world'], ids[1,2,3], users[(username='me')]", expected: Example{
				IDs:   []int{1, 2, 3},
				Users: []*User{{Username: "me"}},
				Enums: ptr([]string{"hello", "world"}),
			},
			},
		} {
			defined, err := New(Example{})
			assert.NoError(t, err)
			value, err := defined.Parse(item.input)
			if item.errExpected != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), item.errExpected)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, item.expected, value)
			}
		}
	})

	t.Run("test all", func(t *testing.T) {
		type Span struct {
			Start int `attr:"name=start"`
			End   int `attr:"name=end"`
		}
		type Field struct {
			Name     string `attr:"name=name"`
			Required bool   `attr:"name=required"`
			Span     *Span  `attr:"name=span"`
			SomeAny  any    `attr:"name=some_any"`
		}

		defined, err := New(Field{})
		assert.NoError(t, err)

		for _, item := range []struct {
			input       string
			expected    Field
			errExpected string
		}{
			{
				input: "name = 'this is the name'",
				expected: Field{
					Name:     "this is the name",
					Required: false,
				},
			},
			{
				input: "name = 'this is the name', span(start=42, end=1024)",
				expected: Field{
					Name:     "this is the name",
					Required: false,
					Span: &Span{
						Start: 42,
						End:   1024,
					},
				},
			},
		} {
			value, err := defined.Parse(item.input)
			if item.errExpected != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), item.errExpected)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, item.expected, value)
			}
		}

	})

	t.Run("test any", func(t *testing.T) {
		type Struct struct {
			Any any `attr:"name=any"`
		}
		defined, err := New(Struct{})
		assert.NoError(t, err)

		for _, item := range []struct {
			input  string
			err    error
			expect Struct
		}{
			{
				input: "any=42",
				expect: Struct{
					Any: 42,
				},
			},
			{
				input: "any='hello'",
				expect: Struct{
					Any: "hello",
				},
			},
			{
				input: "any",
				expect: Struct{
					Any: true,
				},
			},
			{
				input: "any=3.141592653",
				expect: Struct{
					Any: 3.141592653,
				},
			},
		} {
			value, err := defined.Parse(item.input)
			if item.err != nil {
				assert.NotNil(t, err)
				assert.ErrorIs(t, err, item.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, item.expect, value)
			}
		}
	})

	t.Run("test embedded struct", func(t *testing.T) {
		type Inside struct {
			Hello string `attr:"name=hello"`
			World string `attr:"name=world"`
		}
		type Span struct {
			Inside
			Start int `attr:"name=start"`
			End   int `attr:"name=end"`
		}
		type Fourth struct {
			Yes bool `attr:"name=yes"`
		}
		type Other struct {
			Fourth
		}
		type Field struct {
			Name     string `attr:"name=name"`
			Required bool   `attr:"name=required"`
			Other    Other  `attr:"name=other"`
			Span
		}

		defined, err := New(Field{})
		assert.NoError(t, err)

		for _, item := range []struct {
			input       string
			expected    Field
			errExpected string
		}{
			{
				input: "name = 'this is the name'",
				expected: Field{
					Name:     "this is the name",
					Required: false,
				},
			},
			{
				input: "name = 'this is the name', start=42, end=1024, hello='hello', world='world', other(yes=true)",
				expected: Field{
					Name:     "this is the name",
					Required: false,
					Span: Span{
						Inside: Inside{
							Hello: "hello",
							World: "world",
						},
						Start: 42,
						End:   1024,
					},
					Other: Other{
						Fourth: Fourth{
							Yes: true,
						},
					},
				},
			},
		} {
			value, err := defined.Parse(item.input)
			if item.errExpected != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), item.errExpected)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, item.expected, value)
			}
		}
	})

	t.Run("test map support", func(t *testing.T) {

		t.Run("test valid data", func(t *testing.T) {
			type Test struct {
				Metadata    map[string]string  `attr:"name=metadata"`
				MetadataPtr *map[string]string `attr:"name=metadata_ptr"`
			}
			def, err := New(Test{})
			assert.NoError(t, err)
			assert.NotNil(t, def)

			data := []struct {
				input    string
				expected Test
				err      error
			}{
				{input: "", expected: Test{}},
				{input: "metadata()", expected: Test{Metadata: map[string]string{}}},
			}

			for _, item := range data {
				value, err := def.Parse(item.input)
				if item.err != nil {
					assert.ErrorIs(t, err, item.err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, item.expected, value)
				}
			}
		})

	})

	t.Run("test any", func(t *testing.T) {
		type Test struct {
			Metadata  map[string]any `attr:"name=metadata"`
			Metadatas []any          `attr:"name=metadatas"`
			Field     any            `attr:"name=field"`
		}
		def, err := New(Test{})
		assert.NoError(t, err)
		assert.NotNil(t, def)

		data := []struct {
			input    string
			expected Test
			err      error
		}{
			{input: "", expected: Test{}},
			{input: "metadata()", expected: Test{Metadata: map[string]any{}}},
			{input: "metadata(hello='world', priority=42)", expected: Test{Metadata: map[string]any{
				"hello":    "world",
				"priority": 42,
			}}},
			{input: "field=1", expected: Test{Field: 1}},
			{input: "field=true", expected: Test{Field: true}},
			{input: "field", expected: Test{Field: true}},
			{input: "field='world'", expected: Test{Field: "world"}},
		}

		for _, item := range data {
			value, err := def.Parse(item.input)
			if item.err != nil {
				assert.ErrorIs(t, err, item.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, item.expected, value)
			}
		}

	})

	t.Run("test arrays of arrays", func(t *testing.T) {
		type Struct struct {
			ID int `attr:"name=id"`
		}
		type Test struct {
			Metadatass    [][]int     `attr:"name=metadatass"`
			MetadatassAny [][]any     `attr:"name=metadatass_any"`
			Structs       [][]*Struct `attr:"name=structs"`
		}
		def, err := New(Test{})
		assert.NoError(t, err)
		assert.NotNil(t, def)

		data := []struct {
			input    string
			expected Test
			err      error
		}{
			{input: "metadatass[[1,2,3]]", expected: Test{
				Metadatass: [][]int{{1, 2, 3}},
			}},
			{input: "metadatass_any[[1,'hello',4.2]]", expected: Test{
				MetadatassAny: [][]any{{1, "hello", 4.2}},
			}},
			{input: "structs[[(id=42)]]", expected: Test{
				Structs: [][]*Struct{{{
					ID: 42,
				}}},
			}},
			{input: "structs[[()]]", expected: Test{
				Structs: [][]*Struct{{{
					ID: 0,
				}}},
			}},
		}

		for _, item := range data {
			value, err := def.Parse(item.input)
			if item.err != nil {
				assert.ErrorIs(t, err, item.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, item.expected, value)
			}
		}
	})
}

func TestAttrsMust(t *testing.T) {
	assert.NotPanics(t, func() {
		Must(New(AttrDef{}))
	})
	assert.Panics(t, func() {
		Must(New(make(chan int)))
	})

}
