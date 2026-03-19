package parser

func (p *parser) match(matchers ...Matcher) ([]*ParserItem, error) {
	snapshot := p.Snapshot()
	var err error

	result := make([]*ParserItem, 0)
	var pi *ParserItem
outer:
	for _, m := range matchers {
		pi, err = lexParserItem(p)
		if err != nil {
			break outer
		}
		if !m.Match(pi) {
			err = ErrNoMatch
			break outer
		}
		result = append(result, pi)
	}

	if err == nil {
		return result, nil
	}

	// rollback and error
	_ = snapshot.Rollback(p)
	return nil, err
}

type Matcher interface {
	Match(*ParserItem) bool
}

type MatcherFunc func(*ParserItem) bool

func (f MatcherFunc) Match(pi *ParserItem) bool {
	return f(pi)
}

func MatchToken(token ...Token) Matcher {
	return MatcherFunc(func(pi *ParserItem) bool {
		for _, t := range token {
			if t == pi.Token {
				return true
			}
		}
		return false
	})
}

func MatchAll(matchers ...Matcher) Matcher {
	valid := make([]Matcher, 0, len(matchers))
	for _, m := range matchers {
		if m == nil {
			continue
		}
		valid = append(valid, m)
	}
	return MatcherFunc(func(pi *ParserItem) bool {
		for _, m := range valid {
			if !m.Match(pi) {
				return false
			}
		}
		return true
	})
}
