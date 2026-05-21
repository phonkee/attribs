package parser

import "fmt"

func newParserItem(span *SourceSpan, tok Token, value string) (*ParserItem, error) {
	if tok.IsError() {
		return nil, tok.AsError(span, value)
	}

	return &ParserItem{Span: span, Token: tok, Value: value}, nil
}

func lexParserItem(p *parser) (*ParserItem, error) {
	snapshot := p.Snapshot()

	span, tok, value := p.lexer.Lex()
	pi, err := newParserItem(span, tok, value)
	if err != nil {
		return nil, err
	}
	pi.snapshot = snapshot
	pi.lexer = p.lexer
	return pi, nil
}

type ParserItem struct {
	Span     *SourceSpan
	Token    Token
	Value    string
	snapshot *Snapshot
	lexer    *lexer
}

func (p *ParserItem) Rollback() error {
	if p.lexer == nil || p.snapshot == nil {
		return fmt.Errorf("cannot rollback: lexer or snapshot is nil")
	}
	return p.lexer.Rollback(p.snapshot)
}
