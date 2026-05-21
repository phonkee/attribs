package parser

import (
	"errors"
	"fmt"
)

var (
	ErrNotObjectAttribute = errors.New("not object attribute")
	ErrNot                = errors.New("not")
	ErrNotArray           = fmt.Errorf("%w array", ErrNot)
	ErrNotArrayItem       = fmt.Errorf("%w array item", ErrNot)
	ErrNotObject          = fmt.Errorf("%w object", ErrNot)
	ErrNotValue           = fmt.Errorf("%w value", ErrNot)
	ErrNoMatch            = errors.New("no match")
)

func ErrIsNoMatch(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrNoMatch)
}

// ParseError defines error with additional information
// Currently only position is supported, but in future it can be extended
type ParseError interface {
	error

	// Position in parsed text
	Position() int
}

// NewParseError instantiates new parse error
func NewParseError(span *SourceSpan, message string, args ...interface{}) ParseError {
	return parseError{
		span:    span,
		message: fmt.Sprintf(message, args...),
	}
}

type parseError struct {
	message string
	span    *SourceSpan
}

func (p parseError) Error() string {
	return fmt.Sprintf("[span: %v] %v", p.span, p.message)
}

func (p parseError) Position() int {
	return p.span.Position
}
