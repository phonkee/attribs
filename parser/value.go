package parser

// Value represents val of attribute
// - Number represents any number int/float
// - String represents: string, bool
type Value struct {
	Position int
	Boolean  *string
	Number   *string
	String   *string
}
