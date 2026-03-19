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

// Parse parses given inp and returns parsed attributes
func Parse(input io.Reader) (*Attribute, error) {
	p := parser{
		lexer: newLexer(input),
	}
	attr, err := p.Parse2()
	if err != nil {
		return nil, err
	}
	return attr, nil
}

// parser implements parser for attributes
type parser struct {
	// lexer instance
	lexer *lexer

	// peeked obj
	peekObjs []*peekObj
}

// Parse parses given inp and returns parsed attributes
func (p *parser) Parse() ([]*Attribute, error) {
	attrs, err := p.parseAttributes()
	if err != nil {
		return nil, err
	}
	pos, token, value := p.Lex()
	if token != TokenEOF {
		return nil, NewParseError(pos, "unexpected tok %v: %v", token.String(), value)
	}

	return attrs.Object.Attributes, nil
}

// Peek peeks
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

//func (p *parser) Skip() error {
//	_, tok, value := p.Lex()
//	if tok == TokenError {
//		return NewParseError(nil, "unexpected tok %v: %v", tok.String(), value)
//	}
//	if tok != TokenEOF {
//		return io.EOF
//	}
//	return nil
//}

//func (p *parser) PeekEquals(tokens ...Token) bool {
//	_, tok, _ := p.Peek()
//	for _, t := range tokens {
//		if t == tok {
//			return true
//		}
//	}
//	return false
//}

// Lex returns next tok from lexer
func (p *parser) Lex() (span *SourceSpan, token Token, value string) {
	if lpo := len(p.peekObjs); lpo > 0 {
		span, token, value = p.peekObjs[lpo-1].span, p.peekObjs[lpo-1].tok, p.peekObjs[lpo-1].value
		p.peekObjs = p.peekObjs[:lpo-1]
		return
	}
	return p.lexer.Lex()
}

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

func (p *parser) currentPos() int {
	if len(p.peekObjs) > 0 {
		return p.peekObjs[len(p.peekObjs)-1].span.Position
	}
	return p.lexer.pos
}

// Unlex puts tok back to lexer
func (p *parser) Unlex(span *SourceSpan, token Token, value string) {
	p.peekObjs = append(p.peekObjs, &peekObj{span, token, value})
}

// parseArray parses array of values
func (p *parser) parseArray() (result []*Attribute, _ error) {
	result = make([]*Attribute, 0)

outer:
	for {
		span, token, value := p.Lex()

		switch token {
		case TokenString, TokenIdent, TokenNumber, TokenOpenBracket, TokenOpenSquareBracket:
			switch token {
			case TokenString, TokenIdent, TokenNumber:
				p.Unlex(span, token, value)
				val, err := p.parseValue()
				if err != nil {
					return result, nil
				}
				result = append(result, &Attribute{
					Span:  span,
					Value: &val,
				})
			case TokenOpenBracket:
				attributes, err := p.parseAttributes()
				if err != nil {
					return result, err
				}
				// now check closing brace
				_, token, _ := p.Lex()
				if token != TokenCloseBracket {
					return result, NewParseError(span, "unexpected tok %v: %v", token.String(), value)
				}
				result = append(result, &Attribute{
					Span:   span.withLengthFromPosition(p.currentPos()),
					Object: attributes.Object,
				})
			case TokenOpenSquareBracket:
				array, err := p.parseArray()
				if err != nil {
					return result, err
				}
				// now check closing brace
				_, token, _ := p.Lex()
				if token != TokenCloseSquareBracket {
					return result, NewParseError(span, "unexpected tok %v: %v", token.String(), value)
				}
				result = append(result, &Attribute{
					Span:  span.withLengthFromPosition(p.lexer.pos),
					Array: newAttributes(span.withLengthFromPosition(p.lexer.pos), array...),
				})
			}
			// now check for comma
			span, token, value = p.Lex()
			if token == TokenComma {
				continue outer
			}
			p.Unlex(span, token, value)
			return result, nil
		default:
			p.Unlex(span, token, value)
			return result, nil
		}
	}
}

// parseAttributes parses attributes separated by commas
func (p *parser) parseAttributes() (result *Attribute, _ error) {
	result.Object = newAttributes(newSourceSpan(p.currentPos()))
	var canComma bool
outer:
	for {
		item, err := p.lexV2Match(TokenIdent, TokenCloseSquareBracket, TokenComma)
		if err != nil {
			return nil, fmt.Errorf("not attributes: %s", err)
		}
		if item.Token == TokenComma {
			if canComma {
				canComma = false
				continue outer
			}
		} else {
			canComma = true
		}

		switch item.Token {
		case TokenComma:
			if canComma {
				canComma = false
				break outer
			}
			return nil, fmt.Errorf("unexpected comma at position %d", item.Span.Position)
		case TokenCloseSquareBracket:
			break outer
		case TokenIdent:
			_ = item.Rollback()

			attr, errParseAttribute := p.parseAttribute()
			if errParseAttribute != nil {
				return nil, errParseAttribute
			}

			result.Object.Push(attr)
		}
	}
	return result, nil
}

// parseIdent parses single attribute which can be in form
//   - "ident=val" - key val
//   - "ident[...]" - array
//   - "ident(...)" - object
//   - "ident" - enable feature, equals to "ident=true"
func (p *parser) parseAttribute() (result *Attribute, _ error) {
	span, token, value := p.Lex()
	result.Span = span

	switch token {
	case TokenIdent:
		result.Name = value

		// now find out what type of attribute it is
		sp, token, _ := p.Lex()
		switch token {
		case TokenEqual:
			value, err := p.parseValue()
			if err != nil {
				return result, err
			}
			result.Value = &value
		case TokenOpenBracket: // attributes (object)
			attributes, err := p.parseAttributes()
			if err != nil {
				return result, err
			}
			result.Object = attributes.Object
			result.Object.Span = sp

			// now check closing brace
			_, token, _ := p.Lex()
			if token != TokenCloseBracket {
				return result, NewParseError(span, "unexpected tok %v: %v", token.String(), value)
			}
			result.Object.Span = result.Object.Span.withLengthFromPosition(p.currentPos())

		case TokenOpenSquareBracket: // array parsing
			span = newSourceSpan(p.lexer.pos - 1)
			array, err := p.parseArray()
			if err != nil {
				return result, err
			}
			result.Array = newAttributes(span.withLengthFromPosition(p.lexer.pos), array...)

			// now check closing brace
			_, token, _ := p.Lex()
			if token != TokenCloseSquareBracket {
				return result, NewParseError(span, "unexpected tok %v: %v", token.String(), value)
			}
		default:
			// anything else means that we have flag enabled
			trueValue := "true"
			result.Value = &Value{Span: span, Boolean: &trueValue}
			p.Unlex(span, token, value)
		}
	default:
		return result, NewParseError(span, "unexpected tok %v: %v", token.String(), value)
	}
	return result, nil
}

// parseValue parses single val such as string, number or boolean
func (p *parser) parseValue() (result Value, _ error) {
	pos, token, value := p.Lex()
	result.Span = pos
	switch token {
	case TokenString:
		result.String = &value
		return result, nil
	case TokenIdent: // string or boolean
		if value == "true" || value == "false" {
			result.Boolean = &value
			return result, nil
		}
		result.String = &value
		return result, nil
	case TokenNumber:
		result.Number = &value
		return result, nil
	}
	return result, NewParseError(pos, "unexpected tok %v: %v", token.String(), value)
}

// peekObj is used when Peek is called to return "back"
type peekObj struct {
	span  *SourceSpan
	tok   Token
	value string
}
