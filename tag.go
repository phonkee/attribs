package attribs

import (
	"fmt"
	"strings"

	"github.com/phonkee/attribs/parser"
)

// parseAttribsTag parses attribs tag
func parseAttribsTag(tag string, skipUnknown bool) (result attrAttribs, _ error) {
	result.Position = -1

	parsed, err := parser.Parse(strings.NewReader(tag))
	if err != nil {
		return result, err
	}

	if parsed.Object == nil {
		return result, nil
	}

	for _, attr := range parsed.Object.Attributes {
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
		case "pos":
			pos, err := attr.Value.AsInt()
			if err != nil {
				return result, fmt.Errorf("%w: pos must be an integer", ErrInvalidTag)
			}
			result.Position = pos
			result.IsPositional = true
		default:
			if !skipUnknown {
				return result, fmt.Errorf("%w: %v", ErrInvalidTag, attr.Name)
			}
		}
	}

	if err = result.Validate(); err != nil {
		return result, err
	}

	return result, nil
}

// attrAttribs holds information about defined attribute
type attrAttribs struct {
	Alias        string
	Name         string
	Disabled     bool
	Required     bool
	Position     int // -1 = not positional
	IsPositional bool
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

	if parsed.Object == nil {
		return result, nil
	}

	for _, attr := range parsed.Object.Attributes {
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
