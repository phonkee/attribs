package parser

import (
	"fmt"
	"regexp"
)

var (
	regexIdentifier *regexp.Regexp
)

func init() {
	regexIdentifier = regexp.MustCompile("^_*[a-zA-Z][a-zA-Z0-9_]*$")
}

func ValidateIdentifier(input string) error {
	if !regexIdentifier.MatchString(input) {
		return fmt.Errorf("invalid identifier: %s", input)
	}
	return nil
}
