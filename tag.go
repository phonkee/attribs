package attribs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/phonkee/attribs/parser"
)

type attrAttribs struct {
	Alias    string
	Name     string
	Disabled bool
}

// parseAttribsTag parses attribs tag
func parseAttribsTag(tag string) (result attrAttribs, _ error) {
	parsed, err := parser.Parse(strings.NewReader(tag))
	if err != nil {
		return result, err
	}

	// TODO: add validation
	for _, attr := range parsed {
		switch attr.Name {
		case "name":
			if attr.Value == nil || attr.Value.String == nil {
				return result, fmt.Errorf("invalid name: %w", ErrInvalidTag)
			}
			result.Alias = *attr.Value.String
		case "disabled":
			if attr.Value == nil || attr.Value.String == nil {
				b, err := strconv.ParseBool(*attr.Value.String)
				if err != nil {
					return result, fmt.Errorf("%w: invalid value for disabled: %v", ErrInvalidTag, *attr.Value.String)
				}
				result.Disabled = b
			} else {
				return result, fmt.Errorf("%w: invalid value for disabled: %v", ErrInvalidTag, *attr.Value.String)
			}
			result.Disabled = true
		default:
			return result, fmt.Errorf("%w: %v", ErrInvalidTag, attr.Name)
		}
	}

	return result, nil
}
