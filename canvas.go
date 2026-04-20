package termyx

// Canvas is a clipped, offset view into a Buffer that lets custom render
// functions work in local (0,0)-relative coordinates without manually adding
// the node's X/Y offsets. Writes outside the canvas bounds are silently dropped.
//
// Get one from buf.Region(layout) at the start of a RenderFunc:
//
//	n.Render = func(buf *Buffer, l LayoutResult) {
//	    c := buf.Region(l)
//	    c.Fill(0, 0, c.Width, c.Height, ' ', bgStyle)
//	    c.WriteText(2, 1, "hello", titleStyle)
//	}
type Canvas struct {
	buf    *Buffer
	X      int
	Y      int
	Width  int
	Height int
}

// Region returns a Canvas scoped to the given LayoutResult.
// All Canvas methods use coordinates relative to (layout.X, layout.Y).
func (b *Buffer) Region(l LayoutResult) *Canvas {
	return &Canvas{buf: b, X: l.X, Y: l.Y, Width: l.Width, Height: l.Height}
}

// Set writes a single cell at local (x, y). Out-of-bounds writes are ignored.
func (c *Canvas) Set(x, y int, r rune, style Style) {
	if x < 0 || x >= c.Width || y < 0 || y >= c.Height {
		return
	}
	c.buf.Set(c.X+x, c.Y+y, r, style)
}

// WriteText writes a string starting at local (x, y), clipping at canvas edges.
func (c *Canvas) WriteText(x, y int, text string, style Style) {
	for i, r := range text {
		if x+i >= c.Width {
			break
		}
		c.Set(x+i, y, r, style)
	}
}

// WriteTextTrunc writes text truncated to maxW runes starting at local (x, y).
func (c *Canvas) WriteTextTrunc(x, y, maxW int, text string, style Style) {
	runes := []rune(text)
	if len(runes) > maxW {
		runes = runes[:maxW]
	}
	c.WriteText(x, y, string(runes), style)
}

// Fill fills a rectangle at local coordinates with a rune and style.
func (c *Canvas) Fill(x, y, w, h int, r rune, style Style) {
	for row := y; row < y+h; row++ {
		for col := x; col < x+w; col++ {
			c.Set(col, row, r, style)
		}
	}
}

// Sub returns a Canvas scoped to a sub-region at local (x, y) with given size.
// Coordinates within the sub-canvas are relative to its own origin.
func (c *Canvas) Sub(x, y, w, h int) *Canvas {
	// Clamp to parent bounds.
	if x+w > c.Width {
		w = c.Width - x
	}
	if y+h > c.Height {
		h = c.Height - y
	}
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	return &Canvas{buf: c.buf, X: c.X + x, Y: c.Y + y, Width: w, Height: h}
}

// HLine draws a horizontal line of rune r across the full canvas width at row y.
func (c *Canvas) HLine(y int, r rune, style Style) {
	c.Fill(0, y, c.Width, 1, r, style)
}

// VLine draws a vertical line of rune r down the full canvas height at col x.
func (c *Canvas) VLine(x int, r rune, style Style) {
	c.Fill(x, 0, 1, c.Height, r, style)
}
