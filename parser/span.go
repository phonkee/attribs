package parser

import "fmt"

func newSourceSpan(pos int, length ...int) *SourceSpan {
	result := &SourceSpan{
		Position: pos,
	}
	if len(length) > 0 {
		result.Length = length[0]
	}
	return result
}

type SourceSpan struct {
	Position int
	Length   int
}

func (s *SourceSpan) withLength(length int) *SourceSpan {
	return &SourceSpan{
		Position: s.Position,
		Length:   length,
	}
}

func (s *SourceSpan) withPosition(pos int) *SourceSpan {
	return &SourceSpan{
		Position: pos,
		Length:   s.Length,
	}
}

func (s *SourceSpan) withLengthFromPosition(pos int) *SourceSpan {
	return &SourceSpan{
		Position: s.Position,
		Length:   pos - s.Position,
	}
}

func (s *SourceSpan) incrLength() *SourceSpan {
	return s.incrLengthBy(1)

}
func (s *SourceSpan) incrLengthBy(by int) *SourceSpan {
	return &SourceSpan{
		Length:   s.Length + by,
		Position: s.Position,
	}
}

func (s *SourceSpan) String() string {
	return fmt.Sprintf("Span[Position: %d, Length: %d]", s.Position, s.Length)
}
