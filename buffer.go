package termyx

import "github.com/mattn/go-runewidth"

// Cell represents a single character position in the terminal.
// Wide is true for the right half of a 2-column (CJK/emoji) character;
// those cells are skipped by the painter — the terminal advances past them
// automatically when the left-half wide rune is written.
type Cell struct {
	Rune  rune
	Style Style
	Wide  bool
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
// For wide runes use SetWide so the continuation cell is marked correctly.
func (b *Buffer) Set(x, y int, r rune, style Style) {
	if x < 0 || x >= b.Width || y < 0 || y >= b.Height {
		return
	}
	b.Cells[y][x] = Cell{Rune: r, Style: style}
}

// SetWide writes a 2-column rune at (x, y) and marks (x+1, y) as its
// right-half continuation so the painter skips it.
func (b *Buffer) SetWide(x, y int, r rune, style Style) {
	b.Set(x, y, r, style)
	if x+1 < b.Width && y >= 0 && y < b.Height {
		b.Cells[y][x+1] = Cell{Rune: 0, Style: style, Wide: true}
	}
}

// WriteText writes a UTF-8 string starting at (x, y), advancing by the
// display width of each rune (1 for ASCII/narrow, 2 for wide/CJK).
// Clips at buffer edges; a wide rune that would overflow is skipped.
func (b *Buffer) WriteText(x, y int, text string, style Style) {
	col := x
	for _, r := range text {
		if col >= b.Width {
			break
		}
		w := runewidth.RuneWidth(r)
		if w == 2 {
			if col+1 >= b.Width {
				break // no room for a wide rune
			}
			b.SetWide(col, y, r, style)
			col += 2
		} else {
			b.Set(col, y, r, style)
			col++
		}
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
