package attribs

import "errors"

var (
	ErrInvalidTag     = errors.New("invalid tag")
	ErrNotStruct      = errors.New("not a struct")
	ErrDuplicateField = errors.New("duplicate field")
)
