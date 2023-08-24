package parser

import (
	"bufio"
	"io"
	"unicode"
)

func newLexer(reader io.Reader) *lexer {
	return &lexer{
		reader: bufio.NewReader(reader),
	}
}

type lexer struct {
	pos    int
	reader *bufio.Reader
}

func (l *lexer) Lex() (int, Token, string) {
	for {
		r, err := l.read()
		pos := l.pos - 1
		if err != nil {
			if err == io.EOF {
				return pos, TokenEOF, ""
			}
			// this should not happen
			return pos, TokenError, err.Error()
		}

		switch r {
		case '(':
			return pos, TokenOpenBracket, ""
		case ')':
			return pos, TokenCloseBracket, ""
		case '[':
			return pos, TokenOpenSquareBracket, ""
		case ']':
			return pos, TokenCloseSquareBracket, ""
		case '=':
			return pos, TokenEqual, ""
		case ',':
			return pos, TokenComma, ""
		case '\'':
			str := ""
			pos := l.pos - 1
		outer:
			for {
				r, err := l.read()
				if err != nil {
					if err == io.EOF {
						return l.pos - 1, TokenError, ""
					}
					// this should not happen
					return l.pos - 1, TokenError, err.Error()
				}

				switch {
				case r == '\\':
					pr, errp := l.read()
					if errp != nil {
						if errp == io.EOF {
							return l.pos - 1, TokenError, str
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
					return pos, TokenString, str
				}

				if r == '\'' {
					return pos, TokenString, str
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
							return l.pos - 1, TokenError, "found minus sign at EOF"
						}
						return l.pos - 1, TokenError, err.Error()
					}
					l.unread()
					if !(unicode.IsDigit(r) || r == '.') {
						return l.pos - 1, TokenString, ""
					}

					if r == '.' {
						str += "0"
					}
				}

				var foundDot = r == '.'

				for {
					r, err := l.read()
					if err != nil {
						if err == io.EOF {
							break
						}
						return l.pos - 1, TokenError, err.Error()
					}

					if unicode.IsNumber(r) {
						str += string(r)
					} else if r == '.' {
						if foundDot {
							return l.pos - 1, TokenError, "found multiple dots in number"
						}
						foundDot = true
						str += string(r)
					} else {
						l.unread()
						break
					}
				}
				return pos, TokenNumber, str
			}
			if unicode.IsLetter(r) {
				str := string(r)

				for {
					r, err := l.read()
					if err != nil {
						if err == io.EOF {
							break
						}
						return pos, TokenError, err.Error()
					}

					if unicode.IsNumber(r) || unicode.IsLetter(r) || r == '_' {
						str += string(r)
					} else {
						l.unread()
						break
					}
				}
				return pos, TokenIdent, str
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
