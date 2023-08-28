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
Altough New function returns generic definition, mapping values to struct uses reflection, so please don't use it in performance critical code.

# Example

Let's omit error handling in examples. First we need to define all attributes in single structure:
```go
type Tag struct {
    DefaultFirst int `attr:"name=default"`
    Readonly bool `attr:"name=readonly"`
    Description string `attr:"description"`
}
```

Then we create a new instance of attribs. We use generics so parser will return everytime
correct type.

```go
parser, _ := New(Tag{})
```

And then we can parse fields of struct

```go
result, _ := parser.Parse("default=42, readonly, description = 'This is a description'")
```

Now result is:

```go
Tag{
    DefaultFirst: 42,
    Readonly: true,
    Description: "This is a description",
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
- array (and pointers) - (arrays of arrays yet not supported)
- map (and pointers to map) with string key to any supported type
- any type in map/slice/struct

# Author

Peter Vrba <phonkee@phonkee.eu>
