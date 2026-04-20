package termyx

// ScrollState tracks scroll position for a Scroll node.
// The application owns this struct; pass a pointer to Scroll() each frame.
// Mutate it in response to key events before or after building the UI tree.
type ScrollState struct {
	Offset     int  // first visible row (0 = top)
	AutoScroll bool // if true, always show the bottom (live logs)

	// ContentHeight is set by the Scroll node after each render.
	// Use it to clamp Offset or display a scrollbar.
	ContentHeight int
}

// ScrollDown moves the viewport down by n rows.
func (s *ScrollState) ScrollDown(n int) {
	s.Offset += n
	s.AutoScroll = false
}

// ScrollUp moves the viewport up by n rows.
func (s *ScrollState) ScrollUp(n int) {
	s.Offset -= n
	if s.Offset < 0 {
		s.Offset = 0
	}
	s.AutoScroll = false
}

// ScrollToBottom enables auto-scroll mode (viewport sticks to the bottom).
func (s *ScrollState) ScrollToBottom() {
	s.AutoScroll = true
}

// ScrollToTop moves the viewport to the top.
func (s *ScrollState) ScrollToTop() {
	s.Offset = 0
	s.AutoScroll = false
}

// Scroll wraps child in a clipping viewport that can scroll vertically.
// state is owned by the caller and must be preserved across frames.
// maxContentH is the maximum height to allocate for child's layout pass;
// pass 0 to use 10× the viewport height.
//
// Example:
//
//	var scroll termyx.ScrollState
//
//	// in OnKey:
//	case termyx.KeyDown: scroll.ScrollDown(1)
//	case termyx.KeyUp:   scroll.ScrollUp(1)
//
//	// in Root:
//	termyx.Scroll(&scroll, logNode, 0)
func Scroll(state *ScrollState, child *Node, maxContentH int) *Node {
	n := &Node{
		ID:       nextID(),
		Type:     CustomNode,
		Props:    Props{FlexGrow: 1},
		Children: []*Node{child},
	}
	n.Render = func(buf *Buffer, l LayoutResult) {
		if l.Width <= 0 || l.Height <= 0 {
			return
		}

		tall := maxContentH
		if tall <= 0 {
			tall = l.Height * 10
		}
		if tall < l.Height {
			tall = l.Height
		}

		// Layout child in a virtual tall space.
		ComputeLayout(child, 0, 0, l.Width, tall)

		// Render child into a temporary buffer.
		tmp := NewBuffer(l.Width, tall)
		Render(child, tmp)

		// Determine actual content height (last non-blank row + 1).
		contentH := tall
		state.ContentHeight = contentH

		// Clamp and apply auto-scroll.
		maxOffset := contentH - l.Height
		if maxOffset < 0 {
			maxOffset = 0
		}
		if state.AutoScroll {
			state.Offset = maxOffset
		}
		if state.Offset > maxOffset {
			state.Offset = maxOffset
		}
		if state.Offset < 0 {
			state.Offset = 0
		}

		// Blit the viewport from tmp into the main buffer.
		for row := 0; row < l.Height; row++ {
			srcRow := state.Offset + row
			if srcRow >= tall {
				break
			}
			for col := 0; col < l.Width; col++ {
				cell := tmp.Cells[srcRow][col]
				buf.Set(l.X+col, l.Y+row, cell.Rune, cell.Style)
			}
		}
	}
	return n
}
