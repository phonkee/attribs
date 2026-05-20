/*
 * MIT License
 * Copyright (c) 2023 Peter Vrba
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in the
 * Software without restriction, including without limitation the rights to use, copy,
 * modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
 * and to permit persons to whom the Software is furnished to do so, subject to the
 * following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 * EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 * MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
 * NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
 * HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
 * DEALINGS IN THE SOFTWARE.
 */

package parser

import (
	"errors"
	"fmt"
)

func (p *parser) Parse2() (*Attribute, error) {
	span := p.currentSpan()
	oa, err := p.parseV2ObjectAttributes()
	if err != nil {
		return nil, err
	}

	return &Attribute{
		Object: oa,
		Span:   span.withLengthFromPosition(p.currentPos()),
	}, nil
}

func (p *parser) currentSpan() *SourceSpan {
	return newSourceSpan(p.currentPos())
}

func (p *parser) parseV2Object() (*Attributes, error) {
	currentSpan := p.currentSpan()

	if _, err := p.lexV2MatchMap(ErrNotObject, TokenOpenBracket); err != nil {
		return nil, err
	}

	result := newAttributes(currentSpan)

	// Empty object: ()
	if _, err := p.lexV2Match(TokenCloseBracket); err == nil {
		return result, nil
	} else if !ErrIsNoMatch(err) {
		return nil, err
	}

	obj, err := p.parseV2ObjectAttributes()
	if err != nil {
		return nil, err
	}

	result.Attributes = obj.Attributes

	return result, nil
}

func (p *parser) parseV2ObjectAttributes() (*Attributes, error) {
	currentSpan := p.currentSpan()
	result := newAttributes(currentSpan)

	var isInitial bool = true

outer:
	for {
		pi, err := p.lexV2()
		if err != nil {
			return nil, err
		}

		// CloseBracket ends a nested object; EOF ends top-level input.
		if pi.Token == TokenCloseBracket || pi.Token == TokenEOF {
			break outer
		}

		if isInitial {
			isInitial = false
			// Put the token back so parseV2ObjectAttributeKeyValue can read it.
			_ = pi.Rollback()
		} else {
			if pi.Token != TokenComma {
				return nil, fmt.Errorf("expected comma but found %s", pi.Token.String())
			}
		}

		akv, err := p.parseV2ObjectAttributeKeyValue()
		if err != nil {
			return result, err
		}
		result.Push(akv)
	}

	return result, nil
}

func (p *parser) parseV2ObjectAttributeKeyValue() (*Attribute, error) {
	currentSpan := p.currentSpan()
	pi, err := p.lexV2Match(TokenIdent)
	if err != nil {
		if ErrIsNoMatch(err) {
			// No identifier — try to parse a positional value.
			return p.parseV2PositionalValue()
		}
		return nil, NewParseError(currentSpan, "%v: unexpected tok", err)
	}

	result := &Attribute{
		Name: pi.Value,
		Span: pi.Span,
	}

	// lex next operator
	pi, err = p.lexV2Match(TokenOpenBracket, TokenOpenSquareBracket, TokenEqual)
	if err != nil {
		if ErrIsNoMatch(err) {
			// Boolean flag: bare identifier with no =, (, or [.
			trueStr := "true"
			result.Value = &Value{Span: result.Span, Boolean: &trueStr, String: &trueStr}
			return result, nil
		}
		return nil, err
	}

	switch pi.Token {
	case TokenOpenBracket:
		_ = pi.Rollback()
		if result.Object, err = p.parseV2Object(); err != nil {
			return nil, err
		}
	case TokenOpenSquareBracket:
		_ = pi.Rollback()
		if result.Array, err = p.parseV2Array(); err != nil {
			return nil, err
		}
	case TokenEqual:
		value, err := p.parseV2SimpleValue()
		if err != nil {
			return nil, err
		}
		result.Value = value
	}

	return result, nil
}

func (p *parser) parseV2SimpleValue() (*Value, error) {
	pi, err := p.lexV2MatchMap(ErrNotValue, TokenString, TokenIdent, TokenNumber)
	if err != nil {
		return nil, err
	}

	result := &Value{
		Span: pi.Span,
	}
	switch pi.Token {
	case TokenString, TokenIdent:
		if err := result.FromRaw(ValueTypeString, pi.Value); err != nil {
			return result, err
		}
		_ = result.FromRaw(ValueTypeBoolean, pi.Value)
	case TokenNumber:
		if err := result.FromRaw(ValueTypeNumber, pi.Value); err != nil {
			return result, err
		}
	}

	return result, nil
}

func (p *parser) parseV2Array() (*Attributes, error) {
	open, err := p.lexV2MatchMap(ErrNotArray, TokenOpenSquareBracket)
	if err != nil {
		return nil, err
	}
	result := newAttributes(open.Span)

	// Empty array.
	if _, err = p.lexV2Match(TokenCloseSquareBracket); err == nil {
		return result, nil
	} else if !ErrIsNoMatch(err) {
		return nil, NewParseError(open.Span, "unexpected token in array: %v", err)
	}

	// First item (no leading comma).
	first, err := p.parseArrayV2Item()
	if err != nil {
		return nil, NewParseError(open.Span, "expected array item: %v", err)
	}
	result.Push(first)

	// Subsequent items, each preceded by a comma.
	for {
		pi, err := p.lexV2Match(TokenCloseSquareBracket, TokenComma)
		if err != nil {
			return nil, NewParseError(open.Span, "expected ',' or ']': %v", err)
		}
		if pi.Token == TokenCloseSquareBracket {
			break
		}
		next, err := p.parseArrayV2Item()
		if err != nil {
			return nil, NewParseError(open.Span, "expected array item after ',': %v", err)
		}
		result.Push(next)
	}

	return result, nil
}

func (p *parser) parseArrayV2Item() (*Attribute, error) {
	span := p.currentSpan()
	if simpleValue, errSimpleValue := p.parseV2SimpleValue(); errSimpleValue != nil {
		if !errors.Is(errSimpleValue, ErrNotValue) {
			return nil, errSimpleValue
		}
	} else {
		return &Attribute{
			Span:  span.withLengthFromPosition(p.currentPos()),
			Value: simpleValue,
		}, nil
	}

	if arr, errArray := p.parseV2Array(); errArray != nil {
		if !errors.Is(errArray, ErrNotArray) {
			return nil, errArray
		}
	} else {
		return &Attribute{
			Span:  span.withLengthFromPosition(p.currentPos()),
			Array: arr,
		}, nil
	}

	if object, errObject := p.parseV2Object(); errObject != nil {
		if !errors.Is(errObject, ErrNotObject) {
			return nil, errObject
		}
	} else {
		return &Attribute{
			Span:   span.withLengthFromPosition(p.currentPos()),
			Object: object,
		}, nil
	}

	return nil, fmt.Errorf("%w: yep", ErrNotArrayItem)
}

// parseV2PositionalValue parses a value with no key name (positional argument).
// Accepted forms: string literal, number, [array], (object).
func (p *parser) parseV2PositionalValue() (*Attribute, error) {
	span := p.currentSpan()

	if sv, err := p.parseV2SimpleValue(); err == nil {
		return &Attribute{Span: span.withLengthFromPosition(p.currentPos()), Value: sv}, nil
	} else if !errors.Is(err, ErrNotValue) {
		return nil, err
	}

	if arr, err := p.parseV2Array(); err == nil {
		return &Attribute{Span: span.withLengthFromPosition(p.currentPos()), Array: arr}, nil
	} else if !errors.Is(err, ErrNotArray) {
		return nil, err
	}

	if obj, err := p.parseV2Object(); err == nil {
		return &Attribute{Span: span.withLengthFromPosition(p.currentPos()), Object: obj}, nil
	} else if !errors.Is(err, ErrNotObject) {
		return nil, err
	}

	return nil, NewParseError(span, "expected positional value (string, number, array, or object)")
}

// Lex returns next tok from lexer
func (p *parser) lexV2() (*ParserItem, error) {
	return lexParserItem(p)
}

// lexV2Match
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

func (p *parser) lexV2MatchMap(origin error, selected ...Token) (*ParserItem, error) {
	pi, err := p.lexV2Match(selected...)
	if err != nil && ErrIsNoMatch(err) {
		return nil, fmt.Errorf("%w: %v", origin, err)
	}
	return pi, err
}

func (p *parser) peekV2Match(selected ...Token) bool {
	result, err := p.lexV2Match(selected...)
	if err != nil {
		return false
	}
	return result != nil
}

func (p *parser) dbgPrint() {
	span := p.currentSpan()

	println("first", p.lexer.content[:span.Position])
	println("remaining", p.lexer.content[span.Position:])

}
