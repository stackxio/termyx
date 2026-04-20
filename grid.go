package termyx

// Grid renders children in a 2D grid layout.
//
// colSizes specifies column widths; 0 means flex (share remaining space equally).
// rowSizes specifies row heights; 0 means flex.
// If rowSizes is nil, rows are computed automatically from the number of children.
//
// Children are placed left-to-right, top-to-bottom. Extra children beyond
// cols*rows are silently ignored.
//
//	termyx.Grid(
//	    []int{30, 0, 0},  // col 0 = 30 wide, cols 1+2 flex
//	    []int{10, 0},     // row 0 = 10 tall,  row 1 flex
//	    cell1, cell2, cell3,
//	    cell4, cell5, cell6,
//	)
func Grid(colSizes []int, rowSizes []int, children ...*Node) *Node {
	if len(colSizes) == 0 {
		colSizes = []int{0}
	}
	cols := len(colSizes)

	n := &Node{Type: CustomNode}
	n.Children = children
	n.Render = func(buf *Buffer, l LayoutResult) {
		if len(children) == 0 {
			return
		}

		rs := rowSizes
		rows := len(rs)
		if rows == 0 {
			rows = (len(children) + cols - 1) / cols
			rs = make([]int, rows)
		}

		cw := resolveGridSizes(colSizes, l.Width)
		rh := resolveGridSizes(rs, l.Height)

		// Precompute cumulative x offsets.
		xOff := make([]int, cols+1)
		for c := 0; c < cols; c++ {
			xOff[c+1] = xOff[c] + cw[c]
		}
		// Precompute cumulative y offsets.
		yOff := make([]int, rows+1)
		for r := 0; r < rows; r++ {
			yOff[r+1] = yOff[r] + rh[r]
		}

		for i, child := range children {
			col := i % cols
			row := i / cols
			if row >= rows {
				break
			}
			if cw[col] <= 0 || rh[row] <= 0 {
				continue
			}
			ComputeLayout(child, l.X+xOff[col], l.Y+yOff[row], cw[col], rh[row])
			Render(child, buf)
		}
	}
	return n
}

// GridCell wraps a child node and places it at a specific grid position.
// Use this when you need explicit positioning rather than sequential fill order.
// col and row are 0-based. colSpan/rowSpan default to 1 if ≤ 0.
//
// GridCell is used together with GridExplicit, which takes a slice of GridCells
// and places them according to their coordinates.
type GridCell struct {
	Col, Row         int
	ColSpan, RowSpan int // defaults to 1
	Child            *Node
}

// GridExplicit renders children at explicit grid coordinates with optional spanning.
//
//	termyx.GridExplicit(
//	    []int{0, 0, 0},   // 3 flex columns
//	    []int{0, 0},      // 2 flex rows
//	    termyx.GridCell{Col: 0, Row: 0, ColSpan: 2, Child: header},
//	    termyx.GridCell{Col: 2, Row: 0, Child: sidebar},
//	    termyx.GridCell{Col: 0, Row: 1, ColSpan: 3, Child: footer},
//	)
func GridExplicit(colSizes []int, rowSizes []int, cells ...GridCell) *Node {
	if len(colSizes) == 0 {
		colSizes = []int{0}
	}
	cols := len(colSizes)
	rows := len(rowSizes)
	if rows == 0 {
		rows = 1
		rowSizes = []int{0}
	}

	children := make([]*Node, len(cells))
	for i, cell := range cells {
		children[i] = cell.Child
	}

	n := &Node{Type: CustomNode}
	n.Children = children
	n.Render = func(buf *Buffer, l LayoutResult) {
		cw := resolveGridSizes(colSizes, l.Width)
		rh := resolveGridSizes(rowSizes, l.Height)

		xOff := make([]int, cols+1)
		for c := 0; c < cols; c++ {
			xOff[c+1] = xOff[c] + cw[c]
		}
		yOff := make([]int, rows+1)
		for r := 0; r < rows; r++ {
			yOff[r+1] = yOff[r] + rh[r]
		}

		for _, cell := range cells {
			if cell.Child == nil {
				continue
			}
			c := cell.Col
			r := cell.Row
			if c < 0 || c >= cols || r < 0 || r >= rows {
				continue
			}
			cs := cell.ColSpan
			if cs <= 0 {
				cs = 1
			}
			rs := cell.RowSpan
			if rs <= 0 {
				rs = 1
			}
			if c+cs > cols {
				cs = cols - c
			}
			if r+rs > rows {
				rs = rows - r
			}

			w := xOff[c+cs] - xOff[c]
			h := yOff[r+rs] - yOff[r]
			if w <= 0 || h <= 0 {
				continue
			}
			ComputeLayout(cell.Child, l.X+xOff[c], l.Y+yOff[r], w, h)
			Render(cell.Child, buf)
		}
	}
	return n
}

func resolveGridSizes(sizes []int, total int) []int {
	result := make([]int, len(sizes))
	fixed := 0
	flex := 0
	for _, s := range sizes {
		if s > 0 {
			fixed += s
		} else {
			flex++
		}
	}
	remaining := total - fixed
	if remaining < 0 {
		remaining = 0
	}
	each := 0
	last := 0
	if flex > 0 {
		each = remaining / flex
		last = remaining - each*(flex-1)
	}
	fi := 0
	for i, s := range sizes {
		if s > 0 {
			result[i] = s
		} else {
			fi++
			if fi == flex {
				result[i] = last
			} else {
				result[i] = each
			}
		}
	}
	return result
}
