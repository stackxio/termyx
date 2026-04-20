package termyx

// Span is a styled text segment used in RichText nodes and Canvas.WriteSpans.
//
//	termyx.Span{Text: "ERROR", Style: theme.StyleError()}
type Span struct {
	Text  string
	Style Style
}

// S is a convenience constructor for a Span (saves typing in dense layouts).
//
//	termyx.S("hello", boldStyle)
func S(text string, style Style) Span {
	return Span{Text: text, Style: style}
}

// RichText renders a sequence of Spans on a single line (or wrapping across
// multiple lines when Wrap is true). Each span carries its own style.
//
//	termyx.RichText(
//	    termyx.S("pod-abc  ", theme.StyleNormal()),
//	    termyx.S("Running", theme.StyleSuccess()),
//	    termyx.S("  2d", theme.StyleDim()),
//	)
func RichText(spans ...Span) *Node {
	return Custom(func(buf *Buffer, l LayoutResult) {
		c := buf.Region(l)
		x := 0
		for _, sp := range spans {
			runes := []rune(sp.Text)
			for _, r := range runes {
				if x >= l.Width {
					return
				}
				c.Set(x, 0, r, sp.Style)
				x++
			}
		}
	})
}

// RichTextWrap renders spans wrapping across multiple lines. Long words are
// hard-broken at the column boundary.
func RichTextWrap(spans ...Span) *Node {
	return Custom(func(buf *Buffer, l LayoutResult) {
		if l.Width <= 0 || l.Height <= 0 {
			return
		}
		c := buf.Region(l)
		x, y := 0, 0
		for _, sp := range spans {
			runes := []rune(sp.Text)
			for _, r := range runes {
				if r == '\n' {
					x = 0
					y++
					if y >= l.Height {
						return
					}
					continue
				}
				if x >= l.Width {
					x = 0
					y++
					if y >= l.Height {
						return
					}
				}
				c.Set(x, y, r, sp.Style)
				x++
			}
		}
	})
}

// WriteSpans writes a sequence of Spans at (x, y) in the Canvas, returning
// the x position after the last rune. Stops at the canvas right edge.
func (c *Canvas) WriteSpans(x, y int, spans []Span) int {
	for _, sp := range spans {
		for _, r := range []rune(sp.Text) {
			if x >= c.Width {
				return x
			}
			c.Set(x, y, r, sp.Style)
			x++
		}
	}
	return x
}

// SpansWidth returns the total rune width of all spans combined.
func SpansWidth(spans []Span) int {
	n := 0
	for _, sp := range spans {
		n += len([]rune(sp.Text))
	}
	return n
}

// TruncateSpans shortens a span slice so the total rune width fits within
// maxWidth. If truncation occurs, the last visible rune is replaced by "…".
func TruncateSpans(spans []Span, maxWidth int) []Span {
	if maxWidth <= 0 {
		return nil
	}
	total := 0
	for _, sp := range spans {
		total += len([]rune(sp.Text))
	}
	if total <= maxWidth {
		return spans
	}

	budget := maxWidth - 1 // leave 1 cell for "…"
	var out []Span
	for _, sp := range spans {
		runes := []rune(sp.Text)
		if budget <= 0 {
			break
		}
		if len(runes) <= budget {
			out = append(out, sp)
			budget -= len(runes)
		} else {
			out = append(out, Span{Text: string(runes[:budget]), Style: sp.Style})
			budget = 0
		}
	}
	// Append ellipsis with the style of the last span.
	ellStyle := Style{}
	if len(spans) > 0 {
		ellStyle = spans[len(spans)-1].Style
	}
	out = append(out, Span{Text: "…", Style: ellStyle})
	return out
}
