package termyx

// cachedRender stores the rendered cells for one node's layout region.
type cachedRender struct {
	memo   any
	layout LayoutResult
	cells  [][]Cell
}

// renderCache stores cached renders keyed by node ID.
// Keys are set by the reconciler and stable across frames for the same logical node.
var renderCache = map[string]*cachedRender{}

// ClearRenderCache evicts all cached renders. Call after a forced full-refresh
// or when dynamically changing app structure in ways the reconciler won't detect.
func ClearRenderCache() {
	renderCache = map[string]*cachedRender{}
}

// cacheGet returns a cached cell snapshot if the node ID, memo value, and
// layout are all identical to the last cached render. Returns nil on miss.
func cacheGet(id string, memo any, layout LayoutResult) [][]Cell {
	entry, ok := renderCache[id]
	if !ok {
		return nil
	}
	if entry.layout != layout {
		return nil
	}
	if !memoEqual(entry.memo, memo) {
		return nil
	}
	return entry.cells
}

// cachePut stores a copy of the cells in the node's layout region.
func cachePut(id string, memo any, layout LayoutResult, buf *Buffer) {
	if layout.Width <= 0 || layout.Height <= 0 {
		return
	}
	cells := make([][]Cell, layout.Height)
	for dy := 0; dy < layout.Height; dy++ {
		row := make([]Cell, layout.Width)
		sy := layout.Y + dy
		if sy >= 0 && sy < buf.Height {
			for dx := 0; dx < layout.Width; dx++ {
				sx := layout.X + dx
				if sx >= 0 && sx < buf.Width {
					row[dx] = buf.Cells[sy][sx]
				}
			}
		}
		cells[dy] = row
	}
	renderCache[id] = &cachedRender{memo: memo, layout: layout, cells: cells}
}

// blitCells copies a cached snapshot back into buf at the node's layout position.
func blitCells(buf *Buffer, layout LayoutResult, cells [][]Cell) {
	for dy, row := range cells {
		sy := layout.Y + dy
		if sy < 0 || sy >= buf.Height {
			continue
		}
		for dx, cell := range row {
			sx := layout.X + dx
			if sx < 0 || sx >= buf.Width {
				continue
			}
			buf.Cells[sy][sx] = cell
		}
	}
}

// memoEqual returns true if a == b using Go's built-in == operator.
// Returns false (never equal) for non-comparable types (maps, slices, funcs)
// so that non-comparable memos always trigger a re-render — safe fallback.
func memoEqual(a, b any) (equal bool) {
	defer func() {
		if r := recover(); r != nil {
			equal = false
		}
	}()
	return a == b
}
