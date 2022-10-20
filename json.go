package attribs

import (
	"reflect"
	"strings"
)

func GetJsonTagName(field reflect.StructField) string {
	if tag := strings.TrimSpace(field.Tag.Get("json")); tag != "" {
		return strings.Split(tag, ",")[0]
	}
	return ""
}
