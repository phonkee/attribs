package attribs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAttribsTag(t *testing.T) {
	newAttrAttribs := func(name string, required bool) attrAttribs {
		return attrAttribs{
			Name:     name,
			Alias:    name,
			Required: required,
		}
	}

	t.Run("test valid", func(t *testing.T) {
		for _, ti := range []struct {
			name   string
			tag    string
			expect attrAttribs
		}{
			{name: "only name", tag: "name=hello", expect: newAttrAttribs("hello", false)},
			{name: "with required", tag: "name=hello, required=false", expect: newAttrAttribs("hello", false)},
			{name: "with required", tag: "name=hello, required=true", expect: newAttrAttribs("hello", true)},
			{name: "with required no value", tag: "name=hello, required", expect: newAttrAttribs("hello", true)},
		} {
			t.Run(ti.name, func(t *testing.T) {
				p, err := parseAttribsTag(ti.tag, true)
				assert.NoError(t, err)
				assert.Equal(t, ti.expect, p)
			})
		}
	})
	t.Run("test invalid", func(t *testing.T) {
		for _, ti := range []struct {
			name          string
			tag           string
			errorContains string
		}{
			{name: "missing name", tag: "", errorContains: "attribute name is required"},
			{name: "invalid name", tag: "name=\"hello-world\"", errorContains: "invalid attribute name"},
			{name: "invalid name", tag: "name=\"_\"", errorContains: "invalid attribute name"},
			{name: "invalid required", tag: "name=hello, required=what", errorContains: "invalid tag: required not boolean"},
		} {
			t.Run(ti.name, func(t *testing.T) {
				_, err := parseAttribsTag(ti.tag, true)
				assert.Error(t, err)
				assert.ErrorContains(t, err, ti.errorContains)
			})
		}
	})

}
