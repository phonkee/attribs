package main

import "github.com/phonkee/attribs"

type Attribs struct {
	Name        string `attr:"name=name"`
	Description string `attr:"name=description"`
	Other       int    `attr:"name=other"`
}

var (
	def = attribs.Must(attribs.New(Attribs{}))
)

type Some struct {
	Field  string `extag:"name=field, description='yeah this works'"`
	Field2 string `extag:"name=field2, description=\"no this doesn't\", other=1, ignore=true"`
}

func main() {
	attribs.Debug[Attribs, Some]("extag", Some{}, true)

}
