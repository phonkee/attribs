package attribs_test

import (
	"github.com/davecgh/go-spew/spew"
	"testing"

	"github.com/phonkee/attribs"
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
	t.Run("test value", func(t *testing.T) {
		for _, item := range []struct {
			input       string
			expected    AttrDef
			errExpected string
		}{
			//{
			//	input:    "id=42, category_id=64, string='hello world'",
			//	expected: AttrDef{ID: 42, CategoryID: ptr(64), String: "hello world"},
			//},
			//{
			//	input:    "id=42, category_id=64, string_opt='hello world', uint=99",
			//	expected: AttrDef{ID: 42, CategoryID: ptr(64), StringOpt: ptr("hello world"), Uint: 99},
			//},
			//{
			//	input:    "id=42, category_id=64, string_opt='hello world', uint_opt=99",
			//	expected: AttrDef{ID: 42, CategoryID: ptr(64), StringOpt: ptr("hello world"), UintOpt: ptr(uint(99))},
			//},
			//{
			//	input:    "id=42 , category_id = 44,   disabled",
			//	expected: AttrDef{ID: 42, CategoryID: ptr(44), Disabled: true},
			//},
			//{
			//	input:    "",
			//	expected: AttrDef{},
			//},
			//{
			//	input: "string = 'hello world'",
			//	expected: AttrDef{
			//		String: "hello world",
			//	},
			//},
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
			//{
			//	input: "id=1, string_opt=foo, uint=42",
			//	expected: AttrDef{
			//		ID:        1,
			//		StringOpt: ptr("foo"),
			//		Uint:      42,
			//	},
			//},
			//{
			//	input:    "id=42, category_id=64, interval(start=1, end=2)",
			//	expected: AttrDef{ID: 42, CategoryID: ptr(64), Interval: &Interval{Start: 1, End: 2}},
			//},
			//{
			//	input:    "id=42 , category_id = 44,   disabled",
			//	expected: AttrDef{ID: 42, CategoryID: ptr(44), Disabled: true},
			//},
			//{
			//	input:    "",
			//	expected: AttrDef{},
			//},
			//{
			//	input: "string = 'hello world'",
			//	expected: AttrDef{
			//		String: "hello world",
			//	},
			//},
			//{
			//	input: "id=1, interval(start=42, end=65535), string_opt=foo",
			//	expected: AttrDef{
			//		ID: 1,
			//		Interval: &Interval{
			//			Start: 42,
			//			End:   65535,
			//		},
			//		StringOpt: ptr("foo"),
			//	},
			//},
			//{
			//	input: "f32=1.2, f64=2.1",
			//	expected: AttrDef{
			//		F32: 1.2,
			//		F64: 2.1,
			//	},
			//},
		} {
			defined, err := attribs.New(AttrDef{})
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
			defined, err := attribs.New(AttrDef{})
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
		defined, err := attribs.New(AttrDef{})
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
			//IDs   []int     `attr:"name=ids"`
			//Users []*User   `attr:"name=users"`
			Enums *[]string `attr:"name=enums"`
		}

		for _, item := range []struct {
			input       string
			expected    Example
			errExpected string
		}{
			//{input: "", expected: Example{}},
			//{input: "ids[1,2,3], users[(username='me')]", expected: Example{IDs: []int{1, 2, 3}, Users: []*User{{Username: "me"}}}},
			{input: "enums['hello', 'world']", expected: Example{Enums: ptr([]string{"hello", "world"})}},
			//{input: "enums['hello', 'world']", expected: Example{IDs: []int{1, 2, 3}, Users: []*User{{Username: "me"}}}},
		} {
			defined, err := attribs.New(Example{})
			assert.NoError(t, err)
			spew.Dump(defined)

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
		}

		defined, err := attribs.New(Field{})
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

		defined, err := attribs.New(Field{})
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
}
