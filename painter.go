package termyx

import (
	"fmt"
	"io"
	"strings"

	"github.com/mattn/go-runewidth"
)

// Painter writes ANSI escape sequences to a terminal, diffing against the
// previous frame to minimize bytes written.
type Painter struct {
	out  io.Writer
	prev *Buffer
}

// NewPainter creates a Painter writing to out.
func NewPainter(out io.Writer) *Painter {
	return &Painter{out: out}
}

// Invalidate forces a full repaint on the next Paint call.
func (p *Painter) Invalidate() {
	p.prev = nil
}

// Paint renders curr to the terminal. A full paint is done when there is no
// previous frame or the terminal size changed; otherwise only changed cells
// are written.
func (p *Painter) Paint(curr *Buffer) {
	var sb strings.Builder

	if p.prev == nil || p.prev.Width != curr.Width || p.prev.Height != curr.Height {
		fullPaint(&sb, curr)
	} else {
		diffPaint(&sb, p.prev, curr)
	}

	if sb.Len() > 0 {
		fmt.Fprint(p.out, sb.String())
	}

	p.prev = curr
}

func fullPaint(sb *strings.Builder, buf *Buffer) {
	sb.WriteString("\x1b[?25l") // hide cursor
	sb.WriteString("\x1b[2J")   // clear screen

	var last Style
	styled := false

	for y := 0; y < buf.Height; y++ {
		moveTo(sb, y+1, 1)
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y][x]
			if cell.Wide {
				// Right half of a wide char — terminal already advanced past it.
				continue
			}
			if !styled || cell.Style != last {
				applyStyle(sb, cell.Style)
				last = cell.Style
				styled = true
			}
			if cell.Rune == 0 {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(cell.Rune)
			}
		}
	}

	sb.WriteString("\x1b[0m")
	moveTo(sb, buf.Height, 1)
}

func diffPaint(sb *strings.Builder, prev, curr *Buffer) {
	var last Style
	styled := false
	lastX, lastY := -2, -2
	lastW := 1 // display width of the last written rune

	for y := 0; y < curr.Height; y++ {
		for x := 0; x < curr.Width; x++ {
			cur := curr.Cells[y][x]
			prv := prev.Cells[y][x]

			if cur.Wide {
				// Right half of a wide char. Skip — but if it changed (e.g.,
				// was a regular char before), the terminal position is already
				// correct because we wrote the wide rune at x-1.
				continue
			}

			if cur == prv {
				continue
			}

			// Check if cursor is already in position.
			expectedX := lastX + lastW
			if !(y == lastY && x == expectedX) {
				moveTo(sb, y+1, x+1)
			}

			if !styled || cur.Style != last {
				applyStyle(sb, cur.Style)
				last = cur.Style
				styled = true
			}

			r := cur.Rune
			if r == 0 {
				r = ' '
			}
			sb.WriteRune(r)
			lastX, lastY = x, y
			lastW = runewidth.RuneWidth(r)
			if lastW < 1 {
				lastW = 1
			}
		}
	}

	if sb.Len() > 0 {
		sb.WriteString("\x1b[0m")
	}
}

func moveTo(sb *strings.Builder, row, col int) {
	fmt.Fprintf(sb, "\x1b[%d;%dH", row, col)
}

func applyStyle(sb *strings.Builder, s Style) {
	sb.WriteString("\x1b[0m")
	if s.Bold {
		sb.WriteString("\x1b[1m")
	}
	if s.Italic {
		sb.WriteString("\x1b[3m")
	}
	if s.Underline {
		sb.WriteString("\x1b[4m")
	}
	if s.Reverse {
		sb.WriteString("\x1b[7m")
	}
	if s.Strikethrough {
		sb.WriteString("\x1b[9m")
	}
	if s.FG.Set {
		fmt.Fprintf(sb, "\x1b[38;2;%d;%d;%dm", s.FG.R, s.FG.G, s.FG.B)
	}
	if s.BG.Set {
		fmt.Fprintf(sb, "\x1b[48;2;%d;%d;%dm", s.BG.R, s.BG.G, s.BG.B)
	}
}
