package render

import "github.com/grokify/pidl"

// includeBox describes how to draw a scenario include step: a labeled box
// across the participants of the included fragment.
type includeBox struct {
	Label string
	Span  []string // participants in entities order; never empty when p has entities
}

// newIncludeBox resolves an include step against p. When the fragment
// cannot be found (or p is nil), the box spans the first entity so the
// output still renders.
func newIncludeBox(p *pidl.Protocol, f pidl.Flow) includeBox {
	if p == nil {
		return includeBox{Label: f.Include}
	}
	label, span := p.IncludeSpan(f)
	if len(span) == 0 && len(p.Entities) > 0 {
		span = []string{p.Entities[0].ID}
	}
	return includeBox{Label: label, Span: span}
}

// First returns the leftmost participant the box spans, or "".
func (b includeBox) First() string {
	if len(b.Span) == 0 {
		return ""
	}
	return b.Span[0]
}

// Last returns the rightmost participant the box spans, or "".
func (b includeBox) Last() string {
	if len(b.Span) == 0 {
		return ""
	}
	return b.Span[len(b.Span)-1]
}
