package parser

// Token that is emitted by lexer
type Token int

const (
	TokenEOF Token = iota
	TokenIdent
	TokenString
	TokenNumber
	TokenComma
	TokenEqual
	TokenOpenBracket
	TokenCloseBracket
	TokenOpenSquareBracket
	TokenCloseSquareBracket
	TokenError
)

func (t Token) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenIdent:
		return "IDENT"
	case TokenString:
		return "STRING"
	case TokenNumber:
		return "NUMBER"
	case TokenComma:
		return "COMMA"
	case TokenEqual:
		return "EQUAL"
	case TokenOpenBracket:
		return "OPEN_BRACKET"
	case TokenCloseBracket:
		return "CLOSE_BRACKET"
	case TokenOpenSquareBracket:
		return "OPEN_SQUARE_BRACKET"
	case TokenCloseSquareBracket:
		return "CLOSE_SQUARE_BRACKET"
	case TokenError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}
