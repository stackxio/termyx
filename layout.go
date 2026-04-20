package termyx

// ComputeLayout runs a recursive layout pass, assigning X, Y, Width, Height
// to every node in the tree given the available space.
func ComputeLayout(node *Node, x, y, width, height int) {
	node.Layout = LayoutResult{X: x, Y: y, Width: width, Height: height}

	if len(node.Children) == 0 {
		return
	}

	switch node.Props.Direction {
	case DirectionRow:
		layoutChildren(node.Children, x, y, width, height, true)
	case DirectionColumn:
		layoutChildren(node.Children, x, y, width, height, false)
	}
}

// layoutChildren distributes space among children along the main axis.
// horizontal=true means a Row (distribute width), false means a Column (distribute height).
func layoutChildren(children []*Node, x, y, width, height int, horizontal bool) {
	fixedTotal, totalGrow := measureChildren(children, horizontal)

	available := mainAxis(width, height, horizontal) - fixedTotal
	if available < 0 {
		available = 0
	}

	cursor := ternary(horizontal, x, y)

	for i, child := range children {
		main := childMain(child, available, totalGrow, horizontal, i == len(children)-1, cursor, x, y, width, height)
		cross := crossAxis(child, width, height, horizontal)

		if horizontal {
			ComputeLayout(child, cursor, y, main, cross)
		} else {
			ComputeLayout(child, x, cursor, cross, main)
		}

		cursor += main
	}
}

func mainAxis(w, h int, horizontal bool) int {
	if horizontal {
		return w
	}
	return h
}

func crossAxis(child *Node, width, height int, horizontal bool) int {
	if horizontal {
		if child.Props.Height > 0 {
			return child.Props.Height
		}
		return height
	}
	if child.Props.Width > 0 {
		return child.Props.Width
	}
	return width
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
