package attribs

import (
	"fmt"
	"strings"

	"github.com/phonkee/attribs/parser"
)

// parseAttribsTag parses attribs tag
func parseAttribsTag(tag string, skipUnknown bool) (result attrAttribs, _ error) {
	parsed, err := parser.Parse(strings.NewReader(tag))
	if err != nil {
		return result, err
	}

	// TODO: add validation
	for _, attr := range parsed {
		switch attr.Name {
		case "name":
			if result.Name, err = attr.Value.AsTrimmedString(); err != nil {
				return result, fmt.Errorf("invalid name: %w", ErrInvalidTag)
			}
			result.Alias = result.Name
		case "disabled":
			if result.Disabled, err = attr.Value.AsBool(); err != nil {
				return result, fmt.Errorf("%w: disabled not boolean", err)
			}
		case "required":
			if result.Required, err = attr.Value.AsBool(); err != nil {
				return result, fmt.Errorf("%w: required not boolean", ErrInvalidTag)
			}
		default:
			if !skipUnknown {
				return result, fmt.Errorf("%w: %v", ErrInvalidTag, attr.Name)
			}
		}
	}

	return result, nil
}

// attrAttribs holds information about defined attribute
type attrAttribs struct {
	Alias    string
	Name     string
	Disabled bool
	Required bool
}

func (a attrAttribs) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("attribute name is required")
	}
	if err := parser.ValidateIdentifier(a.Name); err != nil {
		return fmt.Errorf("invalid attribute name: %v", a.Name)
	}
	return nil
}

func parseAttribsTagDisabled(tag string) (result bool, _ error) {
	parsed, err := parser.Parse(strings.NewReader(tag))
	if err != nil {
		return result, err
	}

	for _, attr := range parsed {
		switch attr.Name {
		case "disabled":
			if result, err = attr.Value.AsBool(); err != nil {
				return result, fmt.Errorf("%w: disabled not boolean", err)
			}
		default:
			continue
		}
	}

	return result, nil
}
