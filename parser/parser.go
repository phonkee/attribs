package parser

import (
	"io"
)

// MustParse panics on error, usually used in init funcs and/or tests
func MustParse(input io.Reader) []Attribute {
	result, err := Parse(input)
	if err != nil {
		panic(err)
	}
	return result
}

// Parse parses given inp and returns parsed attributes
func Parse(input io.Reader) ([]Attribute, error) {
	p := parser{
		lexer: newLexer(input),
	}
	return p.Parse()
}

// parser implements parser for attributes
type parser struct {
	// lexer instance
	lexer *lexer

	// peeked obj
	peekObj *peekObj
}

// Parse parses given inp and returns parsed attributes
func (p *parser) Parse() ([]Attribute, error) {
	attrs, err := p.parseAttributes()
	if err != nil {
		return nil, err
	}
	pos, token, value := p.lex()
	if token != TokenEOF {
		return nil, NewParseError(pos, "unexpected tok %v: %v", token.String(), value)
	}

	return attrs.Attributes, nil
}

// lex returns next tok from lexer
func (p *parser) lex() (pos int, token Token, value string) {
	if p.peekObj != nil {
		pos, token, value = p.peekObj.pos, p.peekObj.tok, p.peekObj.value
		p.peekObj = nil
		return
	}
	return p.lexer.Lex()
}

// unlex puts tok back to lexer
func (p *parser) unlex(pos int, token Token, value string) {
	p.peekObj = &peekObj{pos, token, value}
}

// parseArray parses array of values
func (p *parser) parseArray() (result []Attribute, _ error) {
	result = make([]Attribute, 0)

outer:
	for {
		pos, token, value := p.lex()

		switch token {
		case TokenString, TokenIdent, TokenNumber, TokenOpenBracket, TokenOpenSquareBracket:
			switch token {
			case TokenString, TokenIdent, TokenNumber:
				p.unlex(pos, token, value)
				val, err := p.parseValue()
				if err != nil {
					return result, nil
				}
				result = append(result, Attribute{
					Position: pos,
					Value:    &val,
				})
			case TokenOpenBracket:
				attributes, err := p.parseAttributes()
				if err != nil {
					return result, err
				}
				// now check closing brace
				_, token, _ := p.lex()
				if token != TokenCloseBracket {
					return result, NewParseError(pos, "unexpected tok %v: %v", token.String(), value)
				}
				result = append(result, Attribute{
					Position:   pos,
					Attributes: attributes.Attributes,
				})
			case TokenOpenSquareBracket:
				array, err := p.parseArray()
				if err != nil {
					return result, err
				}
				// now check closing brace
				_, token, _ := p.lex()
				if token != TokenCloseSquareBracket {
					return result, NewParseError(pos, "unexpected tok %v: %v", token.String(), value)
				}
				result = append(result, Attribute{
					Position: pos,
					Array:    array,
				})
			}
			// now check for comma
			pos, token, value = p.lex()
			if token == TokenComma {
				continue outer
			}
			p.unlex(pos, token, value)
			return result, nil
		default:
			p.unlex(pos, token, value)
			return result, nil
		}
	}
}

// parseAttributes parses attributes separated by commas
func (p *parser) parseAttributes() (result Attribute, _ error) {
	result.Attributes = make([]Attribute, 0)

outer:
	for {
		pos, token, value := p.lex()

		switch token {
		case TokenIdent:
			p.unlex(pos, token, value)
			attr, err := p.parseAttribute()
			if err != nil {
				return result, err
			}
			result.Attributes = append(result.Attributes, attr)

			// now check if there is a comma after the attribute, or anything else
			pos, token, value := p.lex()
			switch token {
			case TokenComma:
				continue outer
			default:
				p.unlex(pos, token, value)
				return result, nil
			}
		default:
			p.unlex(pos, token, value)
			return result, nil
		}
	}
}

// parseIdent parses single attribute which can be in form
//   - "ident=val" - key val
//   - "ident[...]" - array
//   - "ident(...)" - object
//   - "ident" - enable feature, equals to "ident=true"
func (p *parser) parseAttribute() (result Attribute, _ error) {
	pos, token, value := p.lex()
	result.Position = pos

	switch token {
	case TokenIdent:
		result.Name = value

		// now find out what type of attribute it is
		_, token, _ := p.lex()
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
			result.Attributes = attributes.Attributes

			// now check closing brace
			_, token, _ := p.lex()
			if token != TokenCloseBracket {
				return result, NewParseError(pos, "unexpected tok %v: %v", token.String(), value)
			}
		case TokenOpenSquareBracket: // array parsing
			array, err := p.parseArray()
			if err != nil {
				return result, err
			}
			result.Array = array

			// now check closing brace
			_, token, _ := p.lex()
			if token != TokenCloseSquareBracket {
				return result, NewParseError(pos, "unexpected tok %v: %v", token.String(), value)
			}
		default:
			// anything else means that we have flag enabled
			trueValue := "true"
			result.Value = &Value{Position: pos, String: &trueValue}
			p.unlex(pos, token, value)
		}
	default:
		return result, NewParseError(pos, "unexpected tok %v: %v", token.String(), value)
	}
	return result, nil
}

// parseValue parses single val such as string, number or boolean
func (p *parser) parseValue() (result Value, _ error) {
	pos, token, value := p.lex()
	result.Position = pos
	switch token {
	case TokenString:
		result.String = &value
		return result, nil
	case TokenIdent: // string or boolean
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
	pos   int
	tok   Token
	value string
}
