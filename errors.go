package attribs

import "errors"

var (
	ErrInvalidTag      = errors.New("invalid tag")
	ErrNotStruct       = errors.New("not a struct")
	ErrDuplicateField  = errors.New("duplicate field")
	ErrMapKeyNotStr    = errors.New("map key is not a string")
	ErrUnsupportedType = errors.New("unsupported type")
)
