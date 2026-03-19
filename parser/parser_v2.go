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
	if _, _, _, err := p.LexSelected(TokenOpenBracket); err != nil {
		if errors.Is(err, ErrNoMatch) {
			return nil, ErrNotObject
		}
		return nil, err
	}
	result := newAttributes(currentSpan)
	if _, _, _, err := p.LexSelected(TokenCloseBracket); err != nil {
		if !errors.Is(err, ErrNoMatch) {
			return nil, err
		}
	} else {
		return result, nil
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
	var canComma bool

outer:
	for {
		pi, err := p.lexV2MatchMap(ErrNotObjectAttribute, TokenIdent, TokenComma, TokenCloseBracket)
		if err != nil {
			return nil, err
		}
		switch pi.Token {
		case TokenComma:
			if canComma {
				canComma = false
				continue outer
			}
			return nil, fmt.Errorf("unexpected comma at position %d", currentSpan.Position)
		case TokenIdent:
			_ = pi.Rollback()
			akv, err := p.parseV2ObjectAttributeKeyValue()
			if err != nil {
				return result, err
			}
			result.Push(akv)
			canComma = true
		case TokenCloseBracket:
			_ = pi.Rollback()
			break outer
		}
	}

	return result, nil
}

func (p *parser) parseV2ObjectAttributeKeyValue() (*Attribute, error) {
	currentSpan := p.currentSpan()
	pi, err := p.lexV2Match(TokenIdent)
	if err != nil {
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
			return nil, NewParseError(result.Span, "%v: unexpected tok", err)
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
	item, err := p.lexV2MatchMap(ErrNotArray, TokenOpenSquareBracket)
	if err != nil {
		return nil, err
	}
	result := newAttributes(item.Span)

	var (
		arrayItem *Attribute
		canComma  bool
	)

	var initial *ParserItem

	piList, errPiList := p.match(MatchToken(TokenIdent, TokenCloseSquareBracket))
	if errPiList != nil {
		return nil, NewParseError(item.Span, "%v: unexpected what piList", errPiList)
	}
	initial = piList[0]

outer:
	for {

		var current *ParserItem
		if initial != nil {
			current = initial
			initial = nil
		}

		if current == nil {
			p.match(
				MatchToken(TokenCloseSquareBracket),
			)
		}

		//ai, errAi := p.lexV2Match(TokenCloseSquareBracket, TokenComma)
		//if errAi != nil {
		//	if !ErrIsNoMatch(errAi) {
		//		return nil, errAi
		//	}
		//}
		//
		switch current.Token {
		case TokenComma:
			if canComma {
				canComma = false
				continue
			}
			return nil, NewParseError(item.Span, "unexpected comma at position %d", item.Span.Position)
		case TokenCloseSquareBracket:
			break outer
		}

		arrayItem, err = p.parseArrayV2Item()
		if err != nil {
			return nil, err
		}
		result.Push(arrayItem)
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
