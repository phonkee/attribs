# attribs

**attribs** is a Go library that parses human-readable attribute strings into typed Go structs using reflection. It was built for struct tag parsing — instead of hand-rolling per-tag string parsers, you declare a struct once and let **attribs** do the rest.

```
name='user_id', required, span(start=0, end=255), tags['id', 'primary']
```

becomes

```go
Tag{Name: "user_id", Required: true, Span: Span{Start: 0, End: 255}, Tags: []string{"id", "primary"}}
```

The lexer and parser are handwritten (no code generator), so there is no dependency overhead.  
Because `Definition[T]` is built once and reused, the reflection cost is paid only at startup — not on every parse.

---

## Installation

```bash
go get github.com/phonkee/attribs
```

Requires **Go 1.21+** (uses generics).

---

## Quick start

```go
package main

import (
    "fmt"
    "github.com/phonkee/attribs"
)

type FieldTag struct {
    Name        string `attr:"name=name"`
    Description string `attr:"name=description"`
    Required    bool   `attr:"name=required"`
    MaxLength   int    `attr:"name=max_length"`
}

// Build the definition once — safe to call from init().
var fieldDef = attribs.Must(attribs.New(FieldTag{}))

func main() {
    tag, err := fieldDef.Parse("name=email, description='User email address', required, max_length=254", false)
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", tag)
    // {Name:email Description:User email address Required:true MaxLength:254}
}
```

---

## API

### `New[T]` — build a definition

```go
func New[T any](what T) (Definition[T], error)
```

Inspects the struct type and builds a reusable `Definition[T]`.  
Accepts both value and pointer (`New(MyStruct{})` or `New(&MyStruct{})`).  
Returns `ErrNotStruct` if `T` is not a struct.

### `Must` — panic-on-error helper

```go
func Must[T any](d Definition[T], err error) Definition[T]
```

Wraps `New` for use in package-level `var` declarations or `init()` where panicking on a bad definition is acceptable.

```go
var def = attribs.Must(attribs.New(MyTag{}))
```

### `Definition[T].Parse` — parse an attribute string

```go
func (d Definition[T]) Parse(input string, ignoreUnknown bool) (T, error)
```

Parses `input` and returns a populated `T`.

- `ignoreUnknown = false` — returns an error on any attribute name not declared in the struct.
- `ignoreUnknown = true` — silently skips unknown attributes; useful when your tag format carries extra fields consumed by other systems.

---

## Struct field tags

Control how each field is mapped with the `attr` struct tag:

| Option | Type | Description |
|---|---|---|
| `name=<ident>` | string | **Required.** The attribute name as it appears in the input string. |
| `required=true` | bool | Marks the field as required (stored on the definition for your own validation). |
| `disabled=true` | bool | Excludes the field from parsing entirely. |
| `pos=<n>` | int | Marks the field as a positional argument at index `n` (0-based). |

```go
type Example struct {
    Name     string `attr:"name=name"`
    ReadOnly bool   `attr:"name=readonly"`
    Internal string `attr:"name=internal,disabled=true"` // never populated
    First    string `attr:"name=first,pos=0"`             // receives positional arg 0
}
```

> Identifier rules: names must match `^_*[a-zA-Z][a-zA-Z0-9_]*$` — letters, digits, underscores; must contain at least one letter.

---

## Grammar

```
input       = attribute ("," attribute)*

attribute   = ident "=" value           -- key=value pair
            | ident "(" attributes ")"  -- nested object
            | ident "[" items "]"       -- array
            | ident                     -- bare boolean flag  (equivalent to ident=true)
            | string                    -- positional string literal
            | number                    -- positional number
            | "(" attributes ")"        -- positional object
            | "[" items "]"             -- positional array

value       = string | number | ident | "true" | "false"

string      = '"' chars '"' | "'" chars "'"
number      = ["-"] digits ["." digits]
ident       = letter (letter | digit | "_")*
```

Single-quoted strings support `\'` as an escape for a literal apostrophe.  
Double-quoted strings support `\n` for a newline.  
Whitespace between tokens is ignored.

---

## Supported types

| Go type | Example attribute |
|---|---|
| `int`, `int8`…`int64` | `count=42` |
| `uint`, `uint8`…`uint64` | `size=1024` |
| `float32`, `float64` | `ratio=0.75` |
| `string` | `label='hello'` or `label=hello` |
| `bool` | `enabled=true` or bare `enabled` |
| `struct` | `span(start=1, end=10)` |
| `[]T` / `[N]T` | `ids[1, 2, 3]` |
| `map[string]T` | `meta(key='val', n=42)` |
| `any` | accepts any of the above |
| Pointer to any of the above | omitting the field leaves it `nil` |

---

## Examples

### Scalars and bare flags

```go
type Options struct {
    Name    string  `attr:"name=name"`
    Count   int     `attr:"name=count"`
    Ratio   float64 `attr:"name=ratio"`
    Verbose bool    `attr:"name=verbose"`
}

def := attribs.Must(attribs.New(Options{}))

opts, _ := def.Parse("name='worker', count=8, ratio=0.5, verbose", false)
// Options{Name:"worker", Count:8, Ratio:0.5, Verbose:true}
```

### Pointer fields — optional values

Pointer fields are left `nil` when the attribute is absent, letting you distinguish "not provided" from a zero value.

```go
type Filter struct {
    MinAge *int    `attr:"name=min_age"`
    Label  *string `attr:"name=label"`
}

def := attribs.Must(attribs.New(Filter{}))

f, _ := def.Parse("min_age=18", false)
// Filter{MinAge: ptr(18), Label: nil}
```

### Nested structs (objects)

```go
type Span struct {
    Start int `attr:"name=start"`
    End   int `attr:"name=end"`
}
type Field struct {
    Name string `attr:"name=name"`
    Span *Span  `attr:"name=span"`
}

def := attribs.Must(attribs.New(Field{}))

f, _ := def.Parse("name=id, span(start=0, end=255)", false)
// Field{Name:"id", Span:&Span{Start:0, End:255}}
```

### Arrays and slices

```go
type Config struct {
    Tags    []string `attr:"name=tags"`
    Weights []int    `attr:"name=weights"`
}

def := attribs.Must(attribs.New(Config{}))

c, _ := def.Parse("tags['alpha', 'beta', 'gamma'], weights[10, 20, 30]", false)
// Config{Tags:["alpha","beta","gamma"], Weights:[10,20,30]}
```

### Arrays of structs

```go
type User struct {
    Username string `attr:"name=username"`
    Admin    bool   `attr:"name=admin"`
}
type Access struct {
    Users []*User `attr:"name=users"`
}

def := attribs.Must(attribs.New(Access{}))

a, _ := def.Parse("users[(username='alice', admin), (username='bob')]", false)
// Access{Users:[{alice true} {bob false}]}
```

### Nested arrays (arrays of arrays)

```go
type Matrix struct {
    Rows [][]int `attr:"name=rows"`
}

def := attribs.Must(attribs.New(Matrix{}))

m, _ := def.Parse("rows[[1, 2, 3], [4, 5, 6]]", false)
// Matrix{Rows:[[1 2 3] [4 5 6]]}
```

### Maps

`map[string]any` fields accept any object-shaped attribute and build a `map[string]any` automatically.

```go
type Doc struct {
    Metadata map[string]any `attr:"name=metadata"`
}

def := attribs.Must(attribs.New(Doc{}))

d, _ := def.Parse("metadata(author='Alice', version=2, public=true)", false)
// Doc{Metadata:map[author:Alice public:true version:2]}
```

### `any` fields

A field typed `any` accepts any scalar, object, or array value and stores the most specific Go type.

```go
type Dynamic struct {
    Value any `attr:"name=value"`
}

def := attribs.Must(attribs.New(Dynamic{}))

d1, _ := def.Parse("value=42",          false) // int
d2, _ := def.Parse("value='hello'",     false) // string
d3, _ := def.Parse("value=3.14",        false) // float64
d4, _ := def.Parse("value",             false) // bool true  (bare flag)
```

### Embedded structs

Embedded struct fields are flattened — their attributes appear at the same level as the parent struct.

```go
type Base struct {
    Label string `attr:"name=label"`
}
type Mixin struct {
    Description string `attr:"name=description"`
}
type Tag struct {
    Base
    Mixin
    Required bool `attr:"name=required"`
}

def := attribs.Must(attribs.New(Tag{}))

t, _ := def.Parse("label=email, description='Email address', required", false)
// Tag{Base:{label:email}, Mixin:{Description:Email address}, Required:true}
```

> Duplicate attribute names across embedded structs return `ErrDuplicateField` at definition time.

### Recursive structs

Mutually recursive types are handled without infinite loops.

```go
type Node struct {
    Value    int   `attr:"name=value"`
    Children *Node `attr:"name=children"`
}

def := attribs.Must(attribs.New(Node{}))

n, _ := def.Parse("value=1, children(value=2, children(value=3))", false)
```

### Positional arguments

Fields tagged with `pos=N` receive the Nth unnamed (positional) argument.

```go
type Vec2 struct {
    X float64 `attr:"name=x,pos=0"`
    Y float64 `attr:"name=y,pos=1"`
}

def := attribs.Must(attribs.New(Vec2{}))

v, _ := def.Parse("1.5, 2.5", false)
// Vec2{X:1.5, Y:2.5}
```

### Ignoring unknown attributes

Pass `ignoreUnknown = true` when your attribute string may contain fields owned by another system.

```go
type MyTag struct {
    Name string `attr:"name=name"`
}

def := attribs.Must(attribs.New(MyTag{}))

// "priority" and "owner" are unknown — silently skipped
t, _ := def.Parse("name=worker, priority=high, owner=ops", true)
// MyTag{Name:"worker"}
```

---

## Error handling

Parse errors implement the `parser.ParseError` interface and include the byte position where the problem occurred.

```go
result, err := def.Parse("name=, broken", false)
if err != nil {
    fmt.Println(err) // [span: Span[Position:5, Length:1]] expected value, got COMMA: ","
    if pe, ok := err.(interface{ Position() int }); ok {
        fmt.Println("error at byte", pe.Position())
    }
}
```

Package-level sentinel errors:

| Error | When |
|---|---|
| `ErrNotStruct` | `New` called with a non-struct type |
| `ErrDuplicateField` | Two embedded fields map to the same attribute name |
| `ErrMapKeyNotStr` | A `map` field has a non-string key type |
| `ErrUnsupportedType` | A field's Go type is not supported |
| `ErrInvalidTag` | The `attr:"…"` struct tag itself is malformed |

---

## Using the parser directly

The `parser` sub-package can be used standalone — it has no reflection dependency and produces a generic AST.

```go
import "github.com/phonkee/attribs/parser"

root, err := parser.Parse(strings.NewReader("id=42, tags['a', 'b'], meta(x=1)"))
if err != nil {
    panic(err)
}

for _, attr := range root.Object.Attributes {
    fmt.Println(attr.Name, attr.Value, attr.Array, attr.Object)
}
```

`parser.MustParse` panics on error — useful in tests and `init()` functions.

---

## Debug utility

`Debug` iterates the fields of any struct, reads a named tag from each, parses it through a given `Definition`, and prints the result. Handy during development.

```go
type Attribs struct {
    Name        string `attr:"name=name"`
    Description string `attr:"name=description"`
}

type MyStruct struct {
    Field1 string `mytag:"name=field1, description='first field'"`
    Field2 int    `mytag:"name=field2, description='second field'"`
}

def := attribs.Must(attribs.New(Attribs{}))
// (Debug uses the generic type parameters to find the definition and the struct)
attribs.Debug[Attribs, MyStruct]("mytag", MyStruct{}, true)
```

---

## Architecture

```
attribs/
├── definition.go   — public generic API: New, Must, Definition[T].Parse
├── attr.go         — reflection tree built by inspect(); Set() dispatchers
├── tag.go          — parses attr:"…" struct field tags
├── errors.go       — package-level sentinel errors
├── debug.go        — Debug[A,T] development helper
└── parser/
    ├── lexer.go    — hand-written rune-level lexer with snapshot/rollback
    ├── parser.go   — recursive-descent parser; produces *Attribute AST
    ├── attribute.go — AST nodes: Attribute, Attributes, Build()
    ├── value.go    — Value type with typed accessors
    ├── span.go     — SourceSpan for byte-position error reporting
    ├── token.go    — Token enum
    ├── errors.go   — ParseError interface and sentinel errors
    ├── item.go     — ParserItem with rollback support
    └── matcher.go  — Matcher interface and helpers
```

---

## License

MIT — see [LICENSE](LICENSE) file.

## Author

Peter Vrba — [phonkee@phonkee.eu](mailto:phonkee@phonkee.eu)
