package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/phonkee/attribs"
)

type AttribOne struct {
	DefaultFirst int  `attr:"name=default"`
	Readonly     bool `attr:"name=readonly"`
}

type AttribTwo struct {
	Hello int  `attr:"name=hello"`
	World bool `attr:"name=world"`
}

type Tags struct {
	AttribOne *AttribOne `attr:"tag=my_tag_one"`
	AttribTwo *AttribTwo `attr:"tag=my_tag_two"`
}

type Something struct {
	First bool `my_tag_two:"hello=420"`
	Other bool `my_tag_one:"default=1" my_tag_two:"hello=42"`
}

func main() {
	m := attribs.Must(attribs.NewMulti(Tags{}))
	//result := attribs.Must(m.ParseStructTag(`my_tag_one:"default=1" my_tag_two:"hello=42"`))
	result := attribs.Must(m.ParseStruct(Something{}))

	spew.Dump(result)
}
