package ui

// Cell represents a single character position in the terminal.
type Cell struct {
	Rune  rune
	Style Style
}

var blankCell = Cell{Rune: ' '}

// Buffer is an in-memory grid of Cells representing one terminal frame.
type Buffer struct {
	Width  int
	Height int
	Cells  [][]Cell
}

// NewBuffer allocates a Buffer of the given dimensions filled with spaces.
func NewBuffer(width, height int) *Buffer {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	cells := make([][]Cell, height)
	for i := range cells {
		row := make([]Cell, width)
		for j := range row {
			row[j] = blankCell
		}
		cells[i] = row
	}
	return &Buffer{Width: width, Height: height, Cells: cells}
}

// Set writes a cell at (x, y), silently ignoring out-of-bounds writes.
func (b *Buffer) Set(x, y int, r rune, style Style) {
	if x < 0 || x >= b.Width || y < 0 || y >= b.Height {
		return
	}
	b.Cells[y][x] = Cell{Rune: r, Style: style}
}

// WriteText writes a string starting at (x, y), clipping at buffer edges.
func (b *Buffer) WriteText(x, y int, text string, style Style) {
	for i, r := range text {
		b.Set(x+i, y, r, style)
	}
}

// Fill fills a rectangle with a rune and style.
func (b *Buffer) Fill(x, y, w, h int, r rune, style Style) {
	for row := y; row < y+h; row++ {
		for col := x; col < x+w; col++ {
			b.Set(col, row, r, style)
		}
	}
}
