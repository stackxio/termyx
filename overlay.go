package termyx

// OverlayOpts controls the size and position of a modal overlay.
type OverlayOpts struct {
	// Width and Height of the modal. 0 = 75% of the parent dimension.
	Width  int
	Height int
	// XOffset and YOffset nudge the centered position (can be negative).
	XOffset int
	YOffset int
}

// Overlay renders base normally and, when visible is true, renders modal
// centered on top of it. Because Termyx renders into a flat cell buffer,
// "on top" simply means writing after base — later writes overwrite earlier ones.
//
// The modal is given its own layout pass with the resolved size. Pass
// opts.Width/Height to control size; leave at 0 for 75% of the viewport.
//
// Example (help screen):
//
//	termyx.Overlay(mainContent, helpPanel, showHelp, termyx.OverlayOpts{Width: 60, Height: 20})
func Overlay(base, modal *Node, visible bool, opts OverlayOpts) *Node {
	n := &Node{
		ID:       nextID(),
		Type:     CustomNode,
		Props:    Props{FlexGrow: 1},
		Children: []*Node{base},
	}

	n.Render = func(buf *Buffer, l LayoutResult) {
		// Always render the base layer.
		ComputeLayout(base, l.X, l.Y, l.Width, l.Height)
		Render(base, buf)

		if !visible {
			return
		}

		// Resolve modal dimensions.
		mw := opts.Width
		mh := opts.Height
		if mw <= 0 {
			mw = l.Width * 3 / 4
		}
		if mh <= 0 {
			mh = l.Height * 3 / 4
		}
		if mw > l.Width {
			mw = l.Width
		}
		if mh > l.Height {
			mh = l.Height
		}

		// Center modal with optional nudge.
		mx := l.X + (l.Width-mw)/2 + opts.XOffset
		my := l.Y + (l.Height-mh)/2 + opts.YOffset

		// Clamp to buffer bounds.
		if mx < 0 {
			mx = 0
		}
		if my < 0 {
			my = 0
		}
		if mx+mw > l.X+l.Width {
			mw = l.X + l.Width - mx
		}
		if my+mh > l.Y+l.Height {
			mh = l.Y + l.Height - my
		}

		// Layout and render modal over the base.
		ComputeLayout(modal, mx, my, mw, mh)
		Render(modal, buf)
	}
	return n
}

// OverlayFull renders a full-viewport dimmed backdrop and centers modal on top.
// This is the equivalent of lipgloss.Place — it fills the entire terminal with
// bgStyle, then draws the modal centered within it.
//
//	termyx.OverlayFull(base, helpPanel, showHelp, bgStyle,
//	    termyx.OverlayOpts{Width: 60, Height: 20})
func OverlayFull(base, modal *Node, visible bool, bgStyle Style, opts OverlayOpts) *Node {
	if !visible {
		return base
	}
	backdropped := Backdrop(bgStyle, modal)
	return Overlay(base, backdropped, true, opts)
}

// Backdrop fills a region with a semi-transparent overlay color before the modal.
// Use it as the background of a modal to visually separate it from the base layer.
func Backdrop(style Style, child *Node) *Node {
	n := &Node{
		ID:       nextID(),
		Type:     CustomNode,
		Props:    Props{FlexGrow: 1},
		Children: []*Node{child},
	}
	n.Render = func(buf *Buffer, l LayoutResult) {
		buf.Fill(l.X, l.Y, l.Width, l.Height, ' ', style)
		ComputeLayout(child, l.X, l.Y, l.Width, l.Height)
		Render(child, buf)
	}
	return n
}
