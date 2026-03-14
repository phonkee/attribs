package main

import "github.com/phonkee/attribs"

type Attribs struct {
	Name        string `attr:"name=name"`
	Description string `attr:"name=description"`
}

var (
	def = attribs.Must(attribs.New(Attribs{}))
)

type Some struct {
	//Field  string `extag:"name=field, description='yeah this works'"`
	Field2 string `extag:"name=field2, description=\"no this doesn't\""`
}

func main() {
	attribs.Debug[Attribs, Some]("extag", Some{})

}
