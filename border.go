package termyx

// Border wraps child in a box-drawing border with an optional title.
// The child is laid out inside the 1-cell inset and rendered by the tree traversal.
// The border style controls the color of the box lines and title.
func Border(title string, style Style, child *Node) *Node {
	n := &Node{
		ID:       nextID(),
		Type:     CustomNode,
		Props:    Props{FlexGrow: child.Props.FlexGrow},
		Children: []*Node{child},
	}
	n.Render = func(buf *Buffer, l LayoutResult) {
		drawBox(buf, l, title, style)

		inner := LayoutResult{
			X:      l.X + 1,
			Y:      l.Y + 1,
			Width:  l.Width - 2,
			Height: l.Height - 2,
		}
		if inner.Width < 0 {
			inner.Width = 0
		}
		if inner.Height < 0 {
			inner.Height = 0
		}
		ComputeLayout(child, inner.X, inner.Y, inner.Width, inner.Height)
	}
	return n
}

func drawBox(buf *Buffer, l LayoutResult, title string, style Style) {
	x, y, w, h := l.X, l.Y, l.Width, l.Height
	if w < 2 || h < 2 {
		return
	}

	buf.Set(x, y, '┌', style)
	buf.Set(x+w-1, y, '┐', style)
	buf.Set(x, y+h-1, '└', style)
	buf.Set(x+w-1, y+h-1, '┘', style)

	for i := 1; i < w-1; i++ {
		buf.Set(x+i, y, '─', style)
		buf.Set(x+i, y+h-1, '─', style)
	}

	for i := 1; i < h-1; i++ {
		buf.Set(x, y+i, '│', style)
		buf.Set(x+w-1, y+i, '│', style)
	}

	if title != "" && w > 4 {
		max := w - 4
		runes := []rune(title)
		if len(runes) > max {
			runes = runes[:max]
		}
		buf.Set(x+1, y, ' ', style)
		buf.WriteText(x+2, y, string(runes), style)
		pos := x + 2 + len(runes)
		if pos < x+w-1 {
			buf.Set(pos, y, ' ', style)
		}
	}
}
