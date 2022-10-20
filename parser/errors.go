package parser

import (
	"fmt"
)

// ParseError defines error with additional information
// Currently only position is supported, but in future it can be extended
type ParseError interface {
	error

	// Position in parsed text
	Position() int
}

// NewParseError instantiates new parse error
func NewParseError(position int, message string, args ...interface{}) ParseError {
	return parseError{
		position: position,
		message:  fmt.Sprintf(message, args...),
	}
}

type parseError struct {
	message  string
	position int
}

func (p parseError) Error() string {
	return fmt.Sprintf("[pos: %v] %v", p.position, p.message)
}

func (p parseError) Position() int {
	return p.position
}
