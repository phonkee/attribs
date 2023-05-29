package attribs

import "errors"

var (
	ErrInvalidTag      = errors.New("invalid tag")
	ErrNotStruct       = errors.New("not a struct")
	ErrUnsupportedType = errors.New("unsupported type")
)
