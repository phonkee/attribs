package attribs

// return pointer to given value
func ptr[T any](val T) *T {
	return &val
}
