package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLexer(t *testing.T) {
	t.Run("test basic values", func(t *testing.T) {
		data := []struct {
			inp    string
			tok    Token
			val    string
			pos    int
			length int
		}{
			{inp: "=", tok: TokenEqual, val: "=", length: 1},
			{inp: "", tok: TokenEOF, val: "", length: 1},
			{inp: "(", tok: TokenOpenBracket, val: "(", length: 1},
			{inp: ")", tok: TokenCloseBracket, val: ")", length: 1},
			{inp: "[", tok: TokenOpenSquareBracket, val: "[", length: 1},
			{inp: "]", tok: TokenCloseSquareBracket, val: "]", length: 1},
			{inp: "'hello world'", tok: TokenString, val: "hello world"},
			{inp: "'hello world", tok: TokenError, pos: 12},
			{inp: `'hello \' world'`, tok: TokenString, val: "hello ' world"},
			{inp: `'hello \\ world'`, tok: TokenString, val: "hello \\\\ world"},
			{inp: "1234", tok: TokenNumber, val: "1234"},
			{inp: "1234xxx", tok: TokenNumber, val: "1234"},
			{inp: "1234.566", tok: TokenNumber, val: "1234.566"},
			{inp: "ident", tok: TokenIdent, val: "ident"},
			{inp: "ident_1", tok: TokenIdent, val: "ident_1"},
			{inp: "ident_12a", tok: TokenIdent, val: "ident_12a"},
		}

		for _, item := range data {
			span, token, val := newLexer(strings.NewReader(item.inp)).Lex()
			assert.Equal(t, item.pos, span.Position, "inp: %v", item.inp)
			assert.Equal(t, item.tok, token)
			assert.Equal(t, item.val, val)
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
			{inp: `'hello \' world'`, tok: TokenString, val: `hello ' world`},
			{inp: `"hello world"`, tok: TokenString, val: `hello world`},
		}

		for _, item := range data {
			span, token, val := newLexer(strings.NewReader(item.inp)).Lex()
			assert.Equal(t, item.pos, span.Position, "inp: %v", item.inp)
			assert.Equal(t, item.tok, token)
			assert.Equal(t, item.val, val)
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
			span, token, _ := newLexer(strings.NewReader(item.inp)).Lex()
			assert.Equal(t, item.pos, span.Position, "inp: %v", item.inp)
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
			span, token, value := newLexer(strings.NewReader(item.input)).Lex()
			if item.errCont != "" {
				assert.True(t, token == TokenError, "inp: %v", item.input)
				assert.Contains(t, value, item.errCont)
			} else {
				assert.Equal(t, item.pos, span.Position, "inp: %v", item.input)
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
			span, token, value := newLexer(strings.NewReader(item.input)).Lex()
			if item.errCont != "" {
				assert.True(t, token == TokenError, "inp: %v", item.input)
				assert.Contains(t, value, item.errCont)
			} else {
				assert.Equal(t, item.pos, span.Position, "inp: %v", item.input)
				assert.Equal(t, item.tok, token)
				assert.Equal(t, item.val, value)
			}
		}

	})

}

// lexToken holds the token type and value from a single Lex call, used by lexAll.
type lexToken struct {
	tok Token
	val string
}

// lexAll lexes the entire input and returns every token up to and including EOF or Error.
func lexAll(input string) []lexToken {
	l := newLexer(strings.NewReader(input))
	var result []lexToken
	for {
		_, tok, val := l.Lex()
		result = append(result, lexToken{tok, val})
		if tok == TokenEOF || tok == TokenError {
			break
		}
	}
	return result
}

func TestTokenString(t *testing.T) {
	cases := []struct {
		tok      Token
		expected string
	}{
		{TokenEOF, "EOF"},
		{TokenIdent, "IDENT"},
		{TokenString, "STRING"},
		{TokenNumber, "NUMBER"},
		{TokenComma, "COMMA"},
		{TokenEqual, "EQUAL"},
		{TokenOpenBracket, "OPEN_BRACKET"},
		{TokenCloseBracket, "CLOSE_BRACKET"},
		{TokenOpenSquareBracket, "OPEN_SQUARE_BRACKET"},
		{TokenCloseSquareBracket, "CLOSE_SQUARE_BRACKET"},
		{TokenError, "ERROR"},
		{Token(999), "UNKNOWN"},
	}
	for _, c := range cases {
		assert.Equal(t, c.expected, c.tok.String(), "Token(%d).String()", int(c.tok))
	}
}

func TestLexerExtended(t *testing.T) {
	t.Run("structural tokens", func(t *testing.T) {
		cases := []struct {
			inp    string
			tok    Token
			val    string
			pos    int
			length int
		}{
			{inp: ",", tok: TokenComma, val: ",", pos: 0, length: 1},
			{inp: "=", tok: TokenEqual, val: "=", pos: 0, length: 1},
			{inp: "(", tok: TokenOpenBracket, val: "(", pos: 0, length: 1},
			{inp: ")", tok: TokenCloseBracket, val: ")", pos: 0, length: 1},
			{inp: "[", tok: TokenOpenSquareBracket, val: "[", pos: 0, length: 1},
			{inp: "]", tok: TokenCloseSquareBracket, val: "]", pos: 0, length: 1},
		}
		for _, c := range cases {
			span, tok, val := newLexer(strings.NewReader(c.inp)).Lex()
			assert.Equal(t, c.tok, tok, "inp: %q", c.inp)
			assert.Equal(t, c.val, val, "inp: %q", c.inp)
			assert.Equal(t, c.pos, span.Position, "inp: %q", c.inp)
			assert.Equal(t, c.length, span.Length, "inp: %q", c.inp)
		}
	})

	t.Run("whitespace skipping", func(t *testing.T) {
		cases := []struct {
			inp string
			pos int
		}{
			{inp: "  =", pos: 2},
			{inp: "\t=", pos: 1},
			{inp: "\n=", pos: 1},
			{inp: "  \t\n  =", pos: 6},
		}
		for _, c := range cases {
			span, tok, _ := newLexer(strings.NewReader(c.inp)).Lex()
			assert.Equal(t, TokenEqual, tok, "inp: %q", c.inp)
			assert.Equal(t, c.pos, span.Position, "inp: %q", c.inp)
		}
	})

	t.Run("single-quoted strings", func(t *testing.T) {
		cases := []struct {
			inp     string
			tok     Token
			val     string
			isError bool
		}{
			{inp: `''`, tok: TokenString, val: ""},
			{inp: `'hello'`, tok: TokenString, val: "hello"},
			{inp: `'  spaces  '`, tok: TokenString, val: "  spaces  "},
			{inp: `'hello \' world'`, tok: TokenString, val: "hello ' world"},
			// documented behavior: \\ in single-quoted strings produces two backslashes
			{inp: `'hello \\ world'`, tok: TokenString, val: "hello \\\\ world"},
			// documented behavior: unknown escape sequence \n passes backslash through literally
			{inp: `'unknown \n escape'`, tok: TokenString, val: "unknown \\n escape"},
			{inp: `'unterminated`, isError: true},
		}
		for _, c := range cases {
			_, tok, val := newLexer(strings.NewReader(c.inp)).Lex()
			if c.isError {
				assert.Equal(t, TokenError, tok, "inp: %q", c.inp)
			} else {
				assert.Equal(t, c.tok, tok, "inp: %q", c.inp)
				assert.Equal(t, c.val, val, "inp: %q", c.inp)
			}
		}
	})

	t.Run("double-quoted strings", func(t *testing.T) {
		cases := []struct {
			inp     string
			tok     Token
			val     string
			isError bool
		}{
			{inp: `""`, tok: TokenString, val: ""},
			{inp: `"hello"`, tok: TokenString, val: "hello"},
			{inp: `"  spaces  "`, tok: TokenString, val: "  spaces  "},
			// \n is the only escape sequence recognised in double-quoted strings
			{inp: `"hello\nworld"`, tok: TokenString, val: "hello\nworld"},
			// documented behavior: unknown escape \t passes backslash through literally
			{inp: `"unknown \t escape"`, tok: TokenString, val: "unknown \\t escape"},
			{inp: `"unterminated`, isError: true},
		}
		for _, c := range cases {
			_, tok, val := newLexer(strings.NewReader(c.inp)).Lex()
			if c.isError {
				assert.Equal(t, TokenError, tok, "inp: %q", c.inp)
			} else {
				assert.Equal(t, c.tok, tok, "inp: %q", c.inp)
				assert.Equal(t, c.val, val, "inp: %q", c.inp)
			}
		}
	})

	t.Run("numbers edge cases", func(t *testing.T) {
		cases := []struct {
			inp         string
			tok         Token
			val         string
			errContains string
		}{
			{inp: "0", tok: TokenNumber, val: "0"},
			{inp: "3.14", tok: TokenNumber, val: "3.14"},
			{inp: "-3.14", tok: TokenNumber, val: "-3.14"},
			{inp: "-0", tok: TokenNumber, val: "-0"},
			// documented behavior: a lone dot is returned as a valid number token
			{inp: ".", tok: TokenNumber, val: "."},
			// documented behavior: bare minus at EOF produces a TokenError
			{inp: "-", errContains: "found minus sign at EOF"},
			// documented behavior: minus before a non-digit produces an empty TokenString (minus vanishes)
			{inp: "-x", tok: TokenString, val: ""},
		}
		for _, c := range cases {
			_, tok, val := newLexer(strings.NewReader(c.inp)).Lex()
			if c.errContains != "" {
				assert.Equal(t, TokenError, tok, "inp: %q", c.inp)
				assert.Contains(t, val, c.errContains, "inp: %q", c.inp)
			} else {
				assert.Equal(t, c.tok, tok, "inp: %q", c.inp)
				assert.Equal(t, c.val, val, "inp: %q", c.inp)
			}
		}
	})

	t.Run("identifiers edge cases", func(t *testing.T) {
		cases := []struct {
			inp string
			tok Token
			val string
		}{
			{inp: "hello", tok: TokenIdent, val: "hello"},
			{inp: "Hello123", tok: TokenIdent, val: "Hello123"},
			{inp: "hello_world", tok: TokenIdent, val: "hello_world"},
			{inp: "_hello", tok: TokenIdent, val: "_hello"},
			{inp: "__private", tok: TokenIdent, val: "__private"},
			{inp: "true", tok: TokenIdent, val: "true"}, // lexer emits Ident; the parser handles boolean semantics
			{inp: "false", tok: TokenIdent, val: "false"},
			{inp: "a", tok: TokenIdent, val: "a"},
		}
		for _, c := range cases {
			_, tok, val := newLexer(strings.NewReader(c.inp)).Lex()
			assert.Equal(t, c.tok, tok, "inp: %q", c.inp)
			assert.Equal(t, c.val, val, "inp: %q", c.inp)
		}
	})

	t.Run("token sequences", func(t *testing.T) {
		cases := []struct {
			inp      string
			expected []lexToken
		}{
			{
				inp: "key=value",
				expected: []lexToken{
					{TokenIdent, "key"}, {TokenEqual, "="}, {TokenIdent, "value"}, {TokenEOF, ""},
				},
			},
			{
				inp: "key='value'",
				expected: []lexToken{
					{TokenIdent, "key"}, {TokenEqual, "="}, {TokenString, "value"}, {TokenEOF, ""},
				},
			},
			{
				inp: "key, key2",
				expected: []lexToken{
					{TokenIdent, "key"}, {TokenComma, ","}, {TokenIdent, "key2"}, {TokenEOF, ""},
				},
			},
			{
				inp: "[1, 2, 3]",
				expected: []lexToken{
					{TokenOpenSquareBracket, "["},
					{TokenNumber, "1"}, {TokenComma, ","}, {TokenNumber, "2"}, {TokenComma, ","}, {TokenNumber, "3"},
					{TokenCloseSquareBracket, "]"},
					{TokenEOF, ""},
				},
			},
			{
				inp: "()",
				expected: []lexToken{
					{TokenOpenBracket, "("}, {TokenCloseBracket, ")"}, {TokenEOF, ""},
				},
			},
			{
				inp: "  key  =  value  ",
				expected: []lexToken{
					{TokenIdent, "key"}, {TokenEqual, "="}, {TokenIdent, "value"}, {TokenEOF, ""},
				},
			},
			{
				inp: "key=true",
				expected: []lexToken{
					{TokenIdent, "key"}, {TokenEqual, "="}, {TokenIdent, "true"}, {TokenEOF, ""},
				},
			},
			{
				inp: "key=-1",
				expected: []lexToken{
					{TokenIdent, "key"}, {TokenEqual, "="}, {TokenNumber, "-1"}, {TokenEOF, ""},
				},
			},
		}
		for _, c := range cases {
			got := lexAll(c.inp)
			assert.Equal(t, c.expected, got, "inp: %q", c.inp)
		}
	})
}

func TestLexerSnapshotRollback(t *testing.T) {
	t.Run("rollback to start re-lexes same first token", func(t *testing.T) {
		l := newLexer(strings.NewReader("key=value"))
		snap := l.Snapshot()

		_, tok1, val1 := l.Lex()
		require.Equal(t, TokenIdent, tok1)
		require.Equal(t, "key", val1)

		require.NoError(t, l.Rollback(snap))

		_, tok2, val2 := l.Lex()
		assert.Equal(t, tok1, tok2)
		assert.Equal(t, val1, val2)
	})

	t.Run("mid-stream rollback replays from snapshot position", func(t *testing.T) {
		l := newLexer(strings.NewReader("key=value"))
		l.Lex() // consume "key"

		snap := l.Snapshot()
		_, tok1, val1 := l.Lex() // consume "="
		require.Equal(t, TokenEqual, tok1)
		require.Equal(t, "=", val1)

		l.Lex() // consume "value"

		require.NoError(t, l.Rollback(snap))

		_, tok2, val2 := l.Lex()
		assert.Equal(t, TokenEqual, tok2)
		assert.Equal(t, "=", val2)
	})

	t.Run("rollback after EOF re-lexes from start", func(t *testing.T) {
		l := newLexer(strings.NewReader("a"))
		snap := l.Snapshot()

		_, tok1, val1 := l.Lex()
		require.Equal(t, TokenIdent, tok1)
		require.Equal(t, "a", val1)

		_, eof, _ := l.Lex()
		require.Equal(t, TokenEOF, eof)

		require.NoError(t, l.Rollback(snap))

		_, tok2, val2 := l.Lex()
		assert.Equal(t, TokenIdent, tok2)
		assert.Equal(t, "a", val2)
	})

	t.Run("no-op rollback to current position yields same next token", func(t *testing.T) {
		l := newLexer(strings.NewReader("hello"))
		snap := l.Snapshot()

		require.NoError(t, l.Rollback(snap))

		_, tok, val := l.Lex()
		assert.Equal(t, TokenIdent, tok)
		assert.Equal(t, "hello", val)
	})

	t.Run("full sequence survives multiple rollbacks", func(t *testing.T) {
		l := newLexer(strings.NewReader("a=b"))
		snap := l.Snapshot()

		for range 3 {
			require.NoError(t, l.Rollback(snap))
			got := make([]lexToken, 0, 4)
			for {
				_, tok, val := l.Lex()
				got = append(got, lexToken{tok, val})
				if tok == TokenEOF || tok == TokenError {
					break
				}
			}
			assert.Equal(t, []lexToken{
				{TokenIdent, "a"}, {TokenEqual, "="}, {TokenIdent, "b"}, {TokenEOF, ""},
			}, got)
		}
	})
}
