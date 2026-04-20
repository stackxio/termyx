package termyx

// ComputeLayout runs a recursive layout pass, assigning X, Y, Width, Height
// to every node in the tree given the available space.
func ComputeLayout(node *Node, x, y, width, height int) {
	// Apply MinWidth / MinHeight constraints.
	if node.Props.MinWidth > 0 && width < node.Props.MinWidth {
		width = node.Props.MinWidth
	}
	if node.Props.MinHeight > 0 && height < node.Props.MinHeight {
		height = node.Props.MinHeight
	}

	node.Layout = LayoutResult{X: x, Y: y, Width: width, Height: height}

	if len(node.Children) == 0 {
		return
	}

	// Shrink available space by padding before laying out children.
	p := node.Props
	innerX := x + p.PaddingLeft
	innerY := y + p.PaddingTop
	innerW := width - p.PaddingLeft - p.PaddingRight
	innerH := height - p.PaddingTop - p.PaddingBottom
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}

	switch node.Props.Direction {
	case DirectionRow:
		layoutChildren(node.Children, innerX, innerY, innerW, innerH, true, node.Props.AlignItems)
	case DirectionColumn:
		layoutChildren(node.Children, innerX, innerY, innerW, innerH, false, node.Props.AlignItems)
	}
}

// layoutChildren distributes space among children along the main axis.
// horizontal=true means a Row (distribute width), false means a Column (distribute height).
func layoutChildren(children []*Node, x, y, width, height int, horizontal bool, align Align) {
	fixedTotal, totalGrow := measureChildren(children, horizontal)

	available := mainAxis(width, height, horizontal) - fixedTotal
	if available < 0 {
		available = 0
	}

	cursor := ternary(horizontal, x, y)

	for i, child := range children {
		main := childMain(child, available, totalGrow, horizontal, i == len(children)-1, cursor, x, y, width, height)
		childX, childY, childW, childH := crossLayout(child, x, y, width, height, cursor, main, horizontal, align)
		ComputeLayout(child, childX, childY, childW, childH)
		cursor += main
	}
}

// crossLayout computes the position and size of a child on both axes.
func crossLayout(child *Node, x, y, width, height, cursor, main int, horizontal bool, align Align) (cx, cy, cw, ch int) {
	cross := mainAxis(height, width, horizontal) // cross-axis total

	var childCross int
	switch align {
	case AlignStretch:
		childCross = cross
	default:
		if horizontal && child.Props.Height > 0 {
			childCross = child.Props.Height
		} else if !horizontal && child.Props.Width > 0 {
			childCross = child.Props.Width
		} else {
			childCross = cross
		}
	}

	crossOffset := 0
	switch align {
	case AlignCenter:
		crossOffset = (cross - childCross) / 2
	case AlignEnd:
		crossOffset = cross - childCross
	}
	if crossOffset < 0 {
		crossOffset = 0
	}

	if horizontal {
		return cursor, y + crossOffset, main, childCross
	}
	return x + crossOffset, cursor, childCross, main
}

func mainAxis(w, h int, horizontal bool) int {
	if horizontal {
		return w
	}
	return h
}

func childMain(child *Node, available int, totalGrow float64, horizontal bool, isLast bool, cursor, originX, originY, width, height int) int {
	if horizontal && child.Props.Width > 0 {
		return child.Props.Width
	}
	if !horizontal && child.Props.Height > 0 {
		return child.Props.Height
	}

	grow := child.Props.FlexGrow
	if grow <= 0 {
		grow = 1
	}

	if totalGrow <= 0 {
		return available
	}

	// Give the last flex child any remaining space lost to integer truncation.
	if isLast {
		origin := ternary(horizontal, originX, originY)
		total := ternary(horizontal, width, height)
		return origin + total - cursor
	}

	return int(float64(available) * grow / totalGrow)
}

// measureChildren sums fixed sizes and total flex-grow across children.
func measureChildren(children []*Node, horizontal bool) (fixedTotal int, totalGrow float64) {
	for _, c := range children {
		if horizontal && c.Props.Width > 0 {
			fixedTotal += c.Props.Width
			continue
		}
		if !horizontal && c.Props.Height > 0 {
			fixedTotal += c.Props.Height
			continue
		}
		g := c.Props.FlexGrow
		if g <= 0 {
			g = 1
		}
		totalGrow += g
	}
	return
}

func ternary(cond bool, a, b int) int {
	if cond {
		return a
	}
	return b
}
