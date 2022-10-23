package parser

// Attribute representation
type Attribute struct {
	// name of attribute
	Name string

	// pos in original string
	Position int

	// This is poor man's union in go (not complaining, just saying)
	Value      *Value
	Attributes []Attribute
	Array      []Attribute
}

// HasValue returns whether any val was set to attribute
// if used via parser, one of values is always set
func (p *Attribute) HasValue() bool {
	return p.Value != nil || p.Attributes != nil || p.Array != nil
}
