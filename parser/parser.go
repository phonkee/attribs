package parser

import (
	"fmt"
	"io"
)

// MustParse panics on error, usually used in init funcs and/or tests
func MustParse(input io.Reader) *Attribute {
	result, err := Parse(input)
	if err != nil {
		panic(err)
	}
	return result
}

// Parse parses the input and returns the top-level attribute (Object holds all parsed attributes).
func Parse(input io.Reader) (*Attribute, error) {
	p := &parser{lexer: newLexer(input)}
	span := newSourceSpan(0)

	oa, err := p.parseAttributeList()
	if err != nil {
		return nil, err
	}

	_, tok, val := p.Lex()
	if tok != TokenEOF {
		return nil, NewParseError(p.currentSpan(), "unexpected token %s: %q", tok.String(), val)
	}

	return &Attribute{
		Span:   span.withLengthFromPosition(p.currentPos()),
		Object: oa,
	}, nil
}

// parser implements parser for attributes
type parser struct {
	lexer    *lexer
	peekObjs []*peekObj
}

// parseAttributeList parses a comma-separated list of attributes, stopping at EOF or ')'.
func (p *parser) parseAttributeList() (*Attributes, error) {
	span := newSourceSpan(p.currentPos())
	result := newAttributes(span)

	first := true
	for {
		_, tok, _ := p.Peek()
		if tok == TokenEOF || tok == TokenCloseBracket {
			break
		}

		if !first {
			commaSpan := p.currentSpan()
			_, tok, val := p.Lex()
			if tok != TokenComma {
				return nil, NewParseError(commaSpan, "expected ',' but got %s %q", tok.String(), val)
			}
			// double-comma or trailing comma check
			_, nextTok, _ := p.Peek()
			if nextTok == TokenComma {
				return nil, NewParseError(p.currentSpan(), "unexpected double comma")
			}
			if nextTok == TokenEOF || nextTok == TokenCloseBracket {
				return nil, NewParseError(commaSpan, "trailing comma not allowed")
			}
		}
		first = false

		attr, err := p.parseAttribute()
		if err != nil {
			return nil, err
		}
		result.Push(attr)
	}

	result.Span = span.withLengthFromPosition(p.currentPos())
	return result, nil
}

// parseAttribute parses a single attribute which can be:
//   - ident=value   (key=value)
//   - ident(attrs)  (nested object)
//   - ident[items]  (array)
//   - ident         (bare boolean flag, equals ident=true)
//   - string        (positional string value)
//   - number        (positional number value)
//   - (attrs)       (positional object)
//   - [items]       (positional array)
func (p *parser) parseAttribute() (*Attribute, error) {
	span := p.currentSpan()
	_, tok, val := p.Peek()

	switch tok {
	case TokenIdent:
		p.Lex()
		result := &Attribute{Name: val, Span: span}

		_, nextTok, _ := p.Peek()
		switch nextTok {
		case TokenEqual:
			p.Lex() // consume '='
			v, err := p.parseValue()
			if err != nil {
				return nil, err
			}
			result.Value = &v

		case TokenOpenBracket:
			p.Lex() // consume '('
			attrs, err := p.parseAttributeList()
			if err != nil {
				return nil, err
			}
			_, closeTok, _ := p.Lex()
			if closeTok != TokenCloseBracket {
				return nil, NewParseError(span, "expected ')' to close object %q", val)
			}
			result.Object = attrs

		case TokenOpenSquareBracket:
			p.Lex() // consume '['
			arr, err := p.parseArray()
			if err != nil {
				return nil, err
			}
			_, closeTok, _ := p.Lex()
			if closeTok != TokenCloseSquareBracket {
				return nil, NewParseError(span, "expected ']' to close array %q", val)
			}
			result.Array = arr

		default:
			// bare boolean flag: ident with no following =, (, or [
			trueStr := "true"
			result.Value = &Value{Span: span, Boolean: &trueStr, String: &trueStr}
		}
		return result, nil

	case TokenString, TokenNumber:
		v, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		return &Attribute{Span: span.withLengthFromPosition(p.currentPos()), Value: &v}, nil

	case TokenOpenBracket:
		p.Lex() // consume '('
		attrs, err := p.parseAttributeList()
		if err != nil {
			return nil, err
		}
		_, closeTok, _ := p.Lex()
		if closeTok != TokenCloseBracket {
			return nil, NewParseError(span, "expected ')' to close positional object")
		}
		return &Attribute{Span: span.withLengthFromPosition(p.currentPos()), Object: attrs}, nil

	case TokenOpenSquareBracket:
		p.Lex() // consume '['
		arr, err := p.parseArray()
		if err != nil {
			return nil, err
		}
		_, closeTok, _ := p.Lex()
		if closeTok != TokenCloseSquareBracket {
			return nil, NewParseError(span, "expected ']' to close positional array")
		}
		return &Attribute{Span: span.withLengthFromPosition(p.currentPos()), Array: arr}, nil

	default:
		return nil, NewParseError(span, "unexpected token %s: %q", tok.String(), val)
	}
}

// parseArray parses the body of an array (after '[' has already been consumed).
// Returns when ']' is peeked (does not consume it).
func (p *parser) parseArray() (*Attributes, error) {
	span := p.currentSpan()
	result := newAttributes(span)

	// empty array
	_, tok, _ := p.Peek()
	if tok == TokenCloseSquareBracket {
		return result, nil
	}

	// first item (no leading comma)
	item, err := p.parseArrayItem()
	if err != nil {
		return nil, err
	}
	result.Push(item)

	// subsequent items, each preceded by a comma
	for {
		_, tok, _ := p.Peek()
		if tok == TokenCloseSquareBracket {
			break
		}
		if tok == TokenEOF {
			return nil, NewParseError(span, "unclosed array, expected ']'")
		}

		_, tok, val := p.Lex()
		if tok != TokenComma {
			return nil, NewParseError(span, "expected ',' in array but got %s %q", tok.String(), val)
		}

		next, err := p.parseArrayItem()
		if err != nil {
			return nil, err
		}
		result.Push(next)
	}

	result.Span = span.withLengthFromPosition(p.currentPos())
	return result, nil
}

// parseArrayItem parses one element inside an array: a scalar value, a nested array, or an object.
func (p *parser) parseArrayItem() (*Attribute, error) {
	span := p.currentSpan()
	_, tok, _ := p.Peek()

	switch tok {
	case TokenString, TokenNumber, TokenIdent:
		v, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		return &Attribute{Span: span.withLengthFromPosition(p.currentPos()), Value: &v}, nil

	case TokenOpenBracket:
		p.Lex() // consume '('
		attrs, err := p.parseAttributeList()
		if err != nil {
			return nil, err
		}
		_, closeTok, _ := p.Lex()
		if closeTok != TokenCloseBracket {
			return nil, NewParseError(span, "expected ')' to close object in array")
		}
		return &Attribute{Span: span.withLengthFromPosition(p.currentPos()), Object: attrs}, nil

	case TokenOpenSquareBracket:
		p.Lex() // consume '['
		arr, err := p.parseArray()
		if err != nil {
			return nil, err
		}
		_, closeTok, _ := p.Lex()
		if closeTok != TokenCloseSquareBracket {
			return nil, NewParseError(span, "expected ']' to close nested array")
		}
		return &Attribute{Span: span.withLengthFromPosition(p.currentPos()), Array: arr}, nil

	default:
		return nil, NewParseError(span, "unexpected token %s in array", tok.String())
	}
}

// parseValue parses a scalar value: string, number, or boolean/ident.
func (p *parser) parseValue() (Value, error) {
	span, tok, val := p.Lex()
	result := Value{Span: span}
	switch tok {
	case TokenString:
		result.String = &val
		return result, nil
	case TokenIdent:
		result.String = &val
		if val == "true" || val == "false" {
			result.Boolean = &val
		}
		return result, nil
	case TokenNumber:
		result.Number = &val
		return result, nil
	default:
		return result, NewParseError(span, "expected value, got %s: %q", tok.String(), val)
	}
}

// ── lexer helpers ────────────────────────────────────────────────────────────

// currentSpan returns a zero-length span at the current lexer position.
func (p *parser) currentSpan() *SourceSpan {
	return newSourceSpan(p.currentPos())
}

// Peek peeks at the next token without consuming it.
func (p *parser) Peek() (*SourceSpan, Token, string) {
	snapshot := p.lexer.Snapshot()
	span, tok, value := p.Lex()
	if err := p.lexer.Rollback(snapshot); err != nil {
		return nil, TokenError, err.Error()
	}
	return span, tok, value
}

func (p *parser) Snapshot() *Snapshot {
	return p.lexer.Snapshot()
}

// Lex returns the next token from the lexer (or from the unlex buffer).
func (p *parser) Lex() (span *SourceSpan, token Token, value string) {
	if lpo := len(p.peekObjs); lpo > 0 {
		span, token, value = p.peekObjs[lpo-1].span, p.peekObjs[lpo-1].tok, p.peekObjs[lpo-1].value
		p.peekObjs = p.peekObjs[:lpo-1]
		return
	}
	return p.lexer.Lex()
}

// LexSelected lexes the next token and returns it if it matches one of the given tokens.
// On mismatch the token is put back and ErrNoMatch is returned.
func (p *parser) LexSelected(tokens ...Token) (*SourceSpan, Token, string, error) {
	nextSpan, nextToken, nextValue := p.Lex()
	for _, token := range tokens {
		if token == nextToken {
			return nextSpan, nextToken, nextValue, nil
		}
	}
	p.Unlex(nextSpan, nextToken, nextValue)
	return nextSpan, nextToken, nextValue, ErrNoMatch
}

// currentPos returns the current byte position in the input.
func (p *parser) currentPos() int {
	if len(p.peekObjs) > 0 {
		return p.peekObjs[len(p.peekObjs)-1].span.Position
	}
	return p.lexer.pos
}

// Unlex puts a token back into the buffer (LIFO).
func (p *parser) Unlex(span *SourceSpan, token Token, value string) {
	p.peekObjs = append(p.peekObjs, &peekObj{span, token, value})
}

// ── v2-style item helpers (used by matcher/item infrastructure and tests) ───

// lexV2 lexes one token as a ParserItem.
func (p *parser) lexV2() (*ParserItem, error) {
	return lexParserItem(p)
}

// lexV2Match lexes the next token and returns it if it matches one of the given tokens.
// On mismatch the token is put back and ErrNoMatch is returned.
func (p *parser) lexV2Match(selected ...Token) (*ParserItem, error) {
	pi, err := lexParserItem(p)
	if err != nil {
		return nil, err
	}

	for _, tok := range selected {
		if tok == pi.Token {
			return pi, nil
		}
	}

	if err = pi.Rollback(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("%w: %+v", ErrNoMatch, selected)
}

// lexV2MatchMap is like lexV2Match but remaps ErrNoMatch to origin.
func (p *parser) lexV2MatchMap(origin error, selected ...Token) (*ParserItem, error) {
	pi, err := p.lexV2Match(selected...)
	if err != nil && ErrIsNoMatch(err) {
		return nil, fmt.Errorf("%w: %v", origin, err)
	}
	return pi, err
}

// peekV2Match returns true (and consumes the token) when the next token is one of selected.
// Returns false and rolls back when it does not match.
func (p *parser) peekV2Match(selected ...Token) bool {
	result, err := p.lexV2Match(selected...)
	if err != nil {
		return false
	}
	return result != nil
}

// dbgPrint prints the consumed and remaining input at the current position (debugging aid).
func (p *parser) dbgPrint() {
	span := p.currentSpan()
	println("consumed", p.lexer.content[:span.Position])
	println("remaining", p.lexer.content[span.Position:])
}

// peekObj holds a token that has been put back via Unlex.
type peekObj struct {
	span  *SourceSpan
	tok   Token
	value string
}
