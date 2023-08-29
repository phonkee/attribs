# Attribs

This package provides a simple way to parse attributes from a string.
It's more readable than json, and it was implemented to be used for parsing
struct field tags.
Now you don't need to do special parsing for your awesome shiny struct tag, you can
just define a structure and parse 'em.
This package will take care about rest. It provides simple error reporting, with position where
error occurred.
Parser and lexer is handwritten, so there is no overhead from using a parser generator.
You can find parser [here](parser/).
Altough New function returns generic definition, mapping values to struct uses reflection, so please don't use it in
performance critical code.

# Grammar

Attribs defines following grammar for attributes

* `key='value', key2='value2'` - key value pair
* `(key='value')` - object
* `[]` - array

Warning! Top level object must be an object (or pointer to object).

We will show full example

# Example

Let's omit error handling in examples. First we need to define all attributes in single structure:

```go
package main

import (
	"github.com/phonkee/attribs"
)

type Tag struct {
	DefaultFirst int            `attr:"name=default"`
	Readonly     bool           `attr:"name=readonly"`
	Description  string         `attr:"description"`
	Metadata     map[string]any `attr:"name=metadata"`
	Tags         []string       `attr:"name=tags"`
	Inner        Inner          `attr:"name=inner"`
}

type Inner struct {
	Hello string `attr:"name=hello"`
}

var d attribs.Definition[Tag]

func init() {
	// create definition
	d = Must(New(Tag{}))
}

func main() {
	// now parse example attributes definition
	result, _ := d.Parse("default=42, readonly, description='This is a description', tags['tag1', 'tag2'], inner(hello='world')")

	expected := Tag{
		DefaultFirst: 42,
		Readonly:     true,
		Description:  "This is a description",
		Metadata:     map[string]any{},
		Tags:         []string{"tag1", "tag2"},
		Inner: Inner{
			Hello: "world",
		},
	}
}

```

# Supported types

Currently supported (tested) types are

- int, int8, int16, int32, int64 (and pointers)
- uint, uint8, uint16, uint32, uint64  (and pointers)
- string  (and pointers)
- bool  (and pointers)
- float32, float64  (and pointers)
- structs (embedded) (and pointers)
- array (and pointers)
- map (and pointers to map) with string key to any supported type
- `any` type

# Author

Peter Vrba <phonkee@phonkee.eu>
