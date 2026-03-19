package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

func newLexer(reader io.Reader) *lexer {
	all, err := io.ReadAll(reader) // read all to get correct EOF position
	if err != nil {
		panic(err)
	}
	stringContent := string(all)
	return &lexer{
		content: stringContent,
		reader:  bufio.NewReader(strings.NewReader(stringContent)),
	}
}

type lexer struct {
	pos     int
	content string
	reader  *bufio.Reader
}

// Snapshot returns snapshot which can be used to "rollback to"
func (l *lexer) Snapshot() *Snapshot {
	return &Snapshot{
		pos: l.pos,
	}
}

// Rollback to given snapshot
func (l *lexer) Rollback(to *Snapshot) error {
	if to.pos > l.pos {
		return io.EOF
	}
	if to.pos == l.pos {
		return nil
	}
	l.reader = bufio.NewReader(strings.NewReader(l.content))
	disc, err := l.reader.Discard(to.pos)
	if err != nil {
		return err
	}
	if disc != to.pos {
		return fmt.Errorf("expected to discard: %d but %d", to.pos, disc)
	}
	l.pos = to.pos

	return nil
}

func (l *lexer) Lex() (*SourceSpan, Token, string) {
	for {
		r, err := l.read()
		pos := l.pos - 1
		span := newSourceSpan(pos)
		if err != nil {
			if err == io.EOF {
				return span.withLengthFromPosition(l.pos), TokenEOF, ""
			}
			// this should not happen
			return span.withLengthFromPosition(l.pos), TokenError, err.Error()
		}

		switch r {
		case '(':
			return span.withLength(1), TokenOpenBracket, "("
		case ')':
			return span.withLength(1), TokenCloseBracket, ")"
		case '[':
			return span.withLength(1), TokenOpenSquareBracket, "["
		case ']':
			return span.withLength(1), TokenCloseSquareBracket, "]"
		case '=':
			return span.withLength(1), TokenEqual, "="
		case ',':
			return span.withLength(1), TokenComma, ","
		case '"':
			str := ""
		outer2:
			for {
				r, err := l.read()
				if err != nil {
					if err == io.EOF {
						return newSourceSpan(l.pos - 1), TokenError, ""
					}
					// this should not happen
					return span.withLengthFromPosition(l.pos), TokenError, err.Error()
				}

				switch {
				case r == '\\':
					pr, errp := l.read()
					if errp != nil {
						if errp == io.EOF {
							return span.withLengthFromPosition(l.pos), TokenError, str
						}
					}
					switch pr {
					case 'n':
						str += string("\n")
						continue outer2
					default:
						l.unread()
					}
				}

				if r == '"' {
					return span.withLengthFromPosition(l.pos), TokenString, str
				}

				str += string(r)
			}

		case '\'':
			str := ""
		outer:
			for {
				r, err := l.read()
				if err != nil {
					if err == io.EOF {
						return newSourceSpan(l.pos - 1), TokenError, ""
					}
					// this should not happen
					return newSourceSpan(l.pos - 1), TokenError, err.Error()
				}

				switch {
				case r == '\\':
					pr, errp := l.read()
					if errp != nil {
						if errp == io.EOF {
							return newSourceSpan(l.pos - 1), TokenError, str
						}
					}
					switch pr {
					case '\'':
						str += string(pr)
						continue outer
					case '\\':
						// TODO: iterate over this over and over until we find a quote
						str += string(r)
					default:
						l.unread()
					}
				case r == '\'':
					return span.withLengthFromPosition(l.pos), TokenString, str
				}

				if r == '\'' {
					return span.withLengthFromPosition(l.pos), TokenString, str
				}

				str += string(r)
			}
		default:
			if unicode.IsSpace(r) {
				continue // nothing to do here, just move on
			}
			if unicode.IsDigit(r) || r == '.' || r == '-' {
				str := string(r)
				if r == '-' {
					r, err := l.read()
					if err != nil {
						if err == io.EOF {
							return newSourceSpan(l.pos - 1), TokenError, "found minus sign at EOF"
						}
						return newSourceSpan(l.pos - 1), TokenError, err.Error()
					}
					l.unread()
					if !(unicode.IsDigit(r) || r == '.') {
						return span.withLengthFromPosition(l.pos - 1), TokenString, ""
					}

					if r == '.' {
						str += "0"
					}
				}

				span = span.withLengthFromPosition(l.pos)
				var foundDot = r == '.'

				for {
					r, err = l.read()
					if err != nil {
						if err == io.EOF {
							break
						}
						return newSourceSpan(l.pos - 1), TokenError, err.Error()
					}

					if unicode.IsNumber(r) {
						str += string(r)
					} else if r == '.' {
						if foundDot {
							return newSourceSpan(l.pos - 1), TokenError, "found multiple dots in number"
						}
						foundDot = true
						str += string(r)
					} else {
						l.unread()
						break
					}
					span = span.withLengthFromPosition(l.pos)
				}
				return span, TokenNumber, str
			}
			if unicode.IsLetter(r) {
				str := string(r)

				for {
					r, err := l.read()
					if err != nil {
						if err == io.EOF {
							break
						}
						return newSourceSpan(l.pos), TokenError, err.Error()
					}

					if unicode.IsNumber(r) || unicode.IsLetter(r) || r == '_' {
						str += string(r)
						span = span.withLengthFromPosition(l.pos)
					} else {
						l.unread()
						break
					}
				}
				return span, TokenIdent, str
			}
		}
	}
}

func (l *lexer) peek() (rune, error) {
	r, err := l.read()
	l.unread()
	return r, err
}

func (l *lexer) read() (rune, error) {
	r, _, err := l.reader.ReadRune()
	l.pos++
	return r, err
}

func (l *lexer) unread() {
	_ = l.reader.UnreadRune()
	l.pos--
}

// Snapshot taken in time
type Snapshot struct {
	pos int
}

func (s *Snapshot) Rollback(p *parser) error {
	return p.lexer.Rollback(s)
}
