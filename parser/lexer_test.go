package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexer(t *testing.T) {
	t.Run("test basic values", func(t *testing.T) {
		data := []struct {
			inp string
			tok Token
			val string
			pos int
		}{
			{inp: "=", tok: TokenEqual},
			{inp: "", tok: TokenEOF},
			{inp: "(", tok: TokenOpenBracket},
			{inp: ")", tok: TokenCloseBracket},
			{inp: "[", tok: TokenOpenSquareBracket},
			{inp: "]", tok: TokenCloseSquareBracket},
			{inp: "'hello world'", tok: TokenString, val: "hello world"},
			{inp: "'hello world", tok: TokenError, pos: 12},
			{inp: "'hello \\' world'", tok: TokenString, val: "hello \\' world"},
			{inp: "1234", tok: TokenNumber, val: "1234"},
			{inp: "1234xxx", tok: TokenNumber, val: "1234"},
			{inp: "1234.566", tok: TokenNumber, val: "1234.566"},
			{inp: "ident", tok: TokenIdent, val: "ident"},
			{inp: "ident_1", tok: TokenIdent, val: "ident_1"},
			{inp: "ident_12a", tok: TokenIdent, val: "ident_12a"},
		}

		for _, item := range data {
			pos, token, _ := newLexer(strings.NewReader(item.inp)).Lex()
			assert.Equal(t, item.pos, pos, "inp: %v", item.inp)
			assert.Equal(t, item.tok, token)
		}
	})

	t.Run("test string values", func(t *testing.T) {
		data := []struct {
			inp string
			tok Token
			val string
			pos int
		}{
			{inp: "'hello world'", tok: TokenString, val: "hello world"},
			{inp: "'hello world", tok: TokenError, pos: 12},
			{inp: "'hello \\' world'", tok: TokenString, val: "hello \\' world"},
		}

		for _, item := range data {
			pos, token, _ := newLexer(strings.NewReader(item.inp)).Lex()
			assert.Equal(t, item.pos, pos, "inp: %v", item.inp)
			assert.Equal(t, item.tok, token)
		}
	})

	t.Run("test identifier", func(t *testing.T) {
		data := []struct {
			inp string
			tok Token
			val string
			pos int
		}{
			{inp: "ident", tok: TokenIdent, val: "ident"},
			{inp: "ident_1", tok: TokenIdent, val: "ident_1"},
		}

		for _, item := range data {
			pos, token, _ := newLexer(strings.NewReader(item.inp)).Lex()
			assert.Equal(t, item.pos, pos, "inp: %v", item.inp)
			assert.Equal(t, item.tok, token)
		}
	})

	t.Run("test numbers", func(t *testing.T) {
		data := []struct {
			input   string
			tok     Token
			val     string
			pos     int
			errCont string
		}{
			{input: "1234", tok: TokenNumber, val: "1234", pos: 0},
			{input: "1234xxx", tok: TokenNumber, val: "1234", pos: 0},
			{input: "1234.566", tok: TokenNumber, val: "1234.566", pos: 0},
			{input: "1234.566.888", tok: TokenNumber, pos: 0, errCont: "found multiple dots in number"},
			{input: ".888", tok: TokenNumber, val: ".888", pos: 0},
			{input: ".888.123", tok: TokenNumber, pos: 0, errCont: "found multiple dots in number"},
		}

		for _, item := range data {
			pos, token, value := newLexer(strings.NewReader(item.input)).Lex()
			if item.errCont != "" {
				assert.True(t, token == TokenError, "inp: %v", item.input)
				assert.Contains(t, value, item.errCont)
			} else {
				assert.Equal(t, item.pos, pos, "inp: %v", item.input)
				assert.Equal(t, item.tok, token)
				assert.Equal(t, item.val, value)
			}
		}
	})

	t.Run("test negative numbers", func(t *testing.T) {
		data := []struct {
			input   string
			tok     Token
			val     string
			pos     int
			errCont string
		}{
			{input: "-1234", tok: TokenNumber, val: "-1234", pos: 0},
			{input: "-.1234", tok: TokenNumber, val: "-0.1234", pos: 0},
			{input: "-1234xxx", tok: TokenNumber, val: "-1234", pos: 0},
			{input: "-1234.566", tok: TokenNumber, val: "-1234.566", pos: 0},
			{input: "-1234.566.888", tok: TokenNumber, pos: 0, errCont: "found multiple dots in number"},
			{input: "-.888", tok: TokenNumber, val: "-0.888", pos: 0},
			{input: "-.888.123", tok: TokenNumber, pos: 0, errCont: "found multiple dots in number"},
		}

		for _, item := range data {
			pos, token, value := newLexer(strings.NewReader(item.input)).Lex()
			if item.errCont != "" {
				assert.True(t, token == TokenError, "inp: %v", item.input)
				assert.Contains(t, value, item.errCont)
			} else {
				assert.Equal(t, item.pos, pos, "inp: %v", item.input)
				assert.Equal(t, item.tok, token)
				assert.Equal(t, item.val, value)
			}
		}

	})

}
