package termyx

import (
	"fmt"
	"strings"
	"time"
)

// ── List ─────────────────────────────────────────────────────────────────────

// ListItem is a single row in a List.
type ListItem struct {
	Label string
	Style Style // optional per-row style override (zero = use ListStyle.Normal)
}

// ListStyle holds the visual styles for a List component.
type ListStyle struct {
	Normal   Style
	Selected Style
	Header   Style
}

// ListModel holds the state for a List component.
// The application owns this struct and mutates it in response to key events.
type ListModel struct {
	Items    []ListItem
	Selected int   // index of the highlighted row
	Scroll   *ScrollState
	Style    ListStyle
	Header   string // optional header line (empty = no header)
}

// SelectNext moves the selection down one row and adjusts scroll if needed.
func (m *ListModel) SelectNext() {
	if m.Selected < len(m.Items)-1 {
		m.Selected++
	}
}

// SelectPrev moves the selection up one row.
func (m *ListModel) SelectPrev() {
	if m.Selected > 0 {
		m.Selected--
	}
}

// List renders a scrollable, selectable list of items.
// Pass a *ScrollState for scroll support; nil disables scrolling.
func List(model *ListModel) *Node {
	return Custom(func(buf *Buffer, l LayoutResult) {
		c := buf.Region(l)
		y := 0

		if model.Header != "" {
			c.WriteText(0, y, PadRight(model.Header, l.Width), model.Style.Header)
			y++
		}

		for i, item := range model.Items {
			if y >= l.Height {
				break
			}
			style := item.Style
			if (style == Style{}) {
				style = model.Style.Normal
			}
			if i == model.Selected {
				style = model.Style.Selected
			}
			c.WriteText(0, y, PadRight(item.Label, l.Width), style)
			y++
		}
	})
}

// ── Table ─────────────────────────────────────────────────────────────────────

// TableColumn defines a column in a Table.
type TableColumn struct {
	Title string
	Width int // fixed character width; 0 = distribute remaining space equally
}

// TableRow is one data row in a Table.
type TableRow struct {
	Cells []string
	Style Style // optional per-row style override
}

// TableStyle holds visual styles for a Table.
type TableStyle struct {
	Header   Style
	Normal   Style
	Selected Style
	Border   Style
}

// TableModel holds state for a Table component.
type TableModel struct {
	Columns  []TableColumn
	Rows     []TableRow
	Selected int
	Scroll   *ScrollState
	Style    TableStyle
	Border   bool // draw column separator lines
}

// SelectNext moves selection down.
func (m *TableModel) SelectNext() {
	if m.Selected < len(m.Rows)-1 {
		m.Selected++
	}
}

// SelectPrev moves selection up.
func (m *TableModel) SelectPrev() {
	if m.Selected > 0 {
		m.Selected--
	}
}

// Table renders a scrollable table with headers and selectable rows.
func Table(model *TableModel) *Node {
	return Custom(func(buf *Buffer, l LayoutResult) {
		c := buf.Region(l)
		colWidths := resolveColumnWidths(model.Columns, l.Width, model.Border)

		// Header row.
		x := 0
		for ci, col := range model.Columns {
			w := colWidths[ci]
			c.WriteText(x, 0, PadRight(col.Title, w), model.Style.Header)
			if model.Border && ci < len(model.Columns)-1 {
				c.Set(x+w, 0, '│', model.Style.Border)
			}
			x += w
			if model.Border && ci < len(model.Columns)-1 {
				x++
			}
		}

		// Data rows.
		for ri, row := range model.Rows {
			y := ri + 1
			if y >= l.Height {
				break
			}
			style := row.Style
			if (style == Style{}) {
				style = model.Style.Normal
			}
			if ri == model.Selected {
				style = model.Style.Selected
			}
			x = 0
			for ci := range model.Columns {
				w := colWidths[ci]
				cell := ""
				if ci < len(row.Cells) {
					cell = row.Cells[ci]
				}
				c.WriteText(x, y, PadRight(cell, w), style)
				if model.Border && ci < len(model.Columns)-1 {
					c.Set(x+w, y, '│', model.Style.Border)
				}
				x += w
				if model.Border && ci < len(model.Columns)-1 {
					x++
				}
			}
		}
	})
}

func resolveColumnWidths(cols []TableColumn, totalWidth int, border bool) []int {
	widths := make([]int, len(cols))
	fixed := 0
	flex := 0
	separators := 0
	if border && len(cols) > 1 {
		separators = len(cols) - 1
	}
	for i, col := range cols {
		if col.Width > 0 {
			widths[i] = col.Width
			fixed += col.Width
		} else {
			flex++
		}
	}
	remaining := totalWidth - fixed - separators
	if remaining < 0 {
		remaining = 0
	}
	if flex > 0 {
		each := remaining / flex
		last := remaining - each*(flex-1)
		fi := 0
		for i, col := range cols {
			if col.Width == 0 {
				fi++
				if fi == flex {
					widths[i] = last
				} else {
					widths[i] = each
				}
			}
		}
	}
	return widths
}

// ── TabBar ────────────────────────────────────────────────────────────────────

// TabStyle holds visual styles for a TabBar.
type TabStyle struct {
	Active   Style
	Inactive Style
	Gap      Style // style of spacing between tabs
}

// TabBar renders a horizontal row of tab labels.
// active is the index of the currently selected tab (0-based).
func TabBar(labels []string, active int, style TabStyle) *Node {
	return Custom(func(buf *Buffer, l LayoutResult) {
		c := buf.Region(l)
		x := 0
		for i, label := range labels {
			tab := fmt.Sprintf(" %s ", label)
			s := style.Inactive
			if i == active {
				s = style.Active
			}
			c.WriteText(x, 0, tab, s)
			x += len([]rune(tab))
			if x >= l.Width {
				break
			}
			// Gap between tabs.
			c.Set(x, 0, ' ', style.Gap)
			x++
		}
		// Fill remainder.
		for ; x < l.Width; x++ {
			c.Set(x, 0, ' ', style.Gap)
		}
	})
}

// ── StatusBar ─────────────────────────────────────────────────────────────────

// StatusBar renders a single-row bar with left-aligned and right-aligned text.
// Typically used with Fixed(node, 0, 1) to pin it to one row.
func StatusBar(left, right string, style Style) *Node {
	return Custom(func(buf *Buffer, l LayoutResult) {
		c := buf.Region(l)
		// Fill background.
		c.Fill(0, 0, l.Width, l.Height, ' ', style)
		// Left text.
		c.WriteText(0, 0, TruncateHard(left, l.Width), style)
		// Right text (right-aligned).
		if right != "" {
			runes := []rune(right)
			if len(runes) < l.Width {
				c.WriteText(l.Width-len(runes), 0, right, style)
			}
		}
	})
}

// ── Spinner ───────────────────────────────────────────────────────────────────

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// SpinnerModel holds the state for a Spinner.
// Call Tick() on every update tick to advance the animation.
type SpinnerModel struct {
	frame  int
	Label  string
	Style  Style
	Active bool // when false, the spinner is hidden
}

// Tick advances the spinner animation by one frame.
func (s *SpinnerModel) Tick() {
	s.frame = (s.frame + 1) % len(spinnerFrames)
}

// StartAutoTick starts a goroutine that ticks the spinner at the given interval
// and sends on notify to trigger a re-render. Stop it by closing stopC.
func (s *SpinnerModel) StartAutoTick(interval time.Duration, notify chan<- struct{}, stopC <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.Tick()
				select {
				case notify <- struct{}{}:
				default:
				}
			case <-stopC:
				return
			}
		}
	}()
}

// Spinner renders an animated braille spinner with an optional label.
func Spinner(model *SpinnerModel) *Node {
	return Custom(func(buf *Buffer, l LayoutResult) {
		if !model.Active {
			return
		}
		c := buf.Region(l)
		frame := spinnerFrames[model.frame%len(spinnerFrames)]
		text := frame
		if model.Label != "" {
			text = frame + " " + model.Label
		}
		c.WriteText(0, 0, TruncateHard(text, l.Width), model.Style)
	})
}

// ── TextInput ─────────────────────────────────────────────────────────────────

// TextInputModel holds the state for a single-line text input field.
// The application owns this and wires it to key events.
type TextInputModel struct {
	Value       string
	Cursor      int // byte position
	Placeholder string
	Style       Style
	FocusStyle  Style // style when the node is focused
	Focused     bool
}

// Insert appends a rune at the cursor position.
func (t *TextInputModel) Insert(r rune) {
	runes := []rune(t.Value)
	runes = append(runes[:t.Cursor], append([]rune{r}, runes[t.Cursor:]...)...)
	t.Value = string(runes)
	t.Cursor++
}

// Backspace deletes the rune before the cursor.
func (t *TextInputModel) Backspace() {
	if t.Cursor == 0 {
		return
	}
	runes := []rune(t.Value)
	runes = append(runes[:t.Cursor-1], runes[t.Cursor:]...)
	t.Value = string(runes)
	t.Cursor--
}

// Delete removes the rune at the cursor.
func (t *TextInputModel) Delete() {
	runes := []rune(t.Value)
	if t.Cursor >= len(runes) {
		return
	}
	t.Value = string(append(runes[:t.Cursor], runes[t.Cursor+1:]...))
}

// Clear resets the input value and cursor.
func (t *TextInputModel) Clear() {
	t.Value = ""
	t.Cursor = 0
}

// HandleKey processes a KeyEvent and returns true if it was consumed.
func (t *TextInputModel) HandleKey(ev KeyEvent) bool {
	switch ev.Key {
	case KeyBackspace:
		t.Backspace()
		return true
	case KeyDelete:
		t.Delete()
		return true
	case KeyLeft:
		if t.Cursor > 0 {
			t.Cursor--
		}
		return true
	case KeyRight:
		if t.Cursor < len([]rune(t.Value)) {
			t.Cursor++
		}
		return true
	case KeyHome:
		t.Cursor = 0
		return true
	case KeyEnd:
		t.Cursor = len([]rune(t.Value))
		return true
	}
	if ev.Rune != 0 {
		t.Insert(ev.Rune)
		return true
	}
	return false
}

// TextInput renders a single-line editable text field.
func TextInput(model *TextInputModel) *Node {
	return Custom(func(buf *Buffer, l LayoutResult) {
		c := buf.Region(l)
		style := model.Style
		if model.Focused {
			style = model.FocusStyle
		}
		c.Fill(0, 0, l.Width, 1, ' ', style)

		display := model.Value
		if display == "" && !model.Focused && model.Placeholder != "" {
			dim := style
			dim.FG = Color{R: 120, G: 120, B: 120, Set: true}
			c.WriteText(0, 0, TruncateHard(model.Placeholder, l.Width), dim)
			return
		}

		runes := []rune(display)
		// Scroll view to keep cursor visible.
		viewStart := 0
		if model.Cursor >= l.Width {
			viewStart = model.Cursor - l.Width + 1
		}
		if viewStart > len(runes) {
			viewStart = len(runes)
		}
		visible := runes[viewStart:]
		c.WriteText(0, 0, TruncateHard(string(visible), l.Width), style)

		// Draw cursor.
		if model.Focused {
			cursorX := model.Cursor - viewStart
			if cursorX >= 0 && cursorX < l.Width {
				cursorRune := ' '
				if cursorX < len(visible) {
					cursorRune = visible[cursorX]
				}
				cursorStyle := style
				cursorStyle.Reverse = true
				c.Set(cursorX, 0, cursorRune, cursorStyle)
			}
		}
	})
}

// ── ProgressBar ───────────────────────────────────────────────────────────────

// ProgressBar renders a horizontal progress bar.
// pct is 0.0–100.0. label is appended after the bar (empty = show percentage).
func ProgressBar(pct float64, label string, filled, empty Style) *Node {
	return Custom(func(buf *Buffer, l LayoutResult) {
		c := buf.Region(l)
		if l.Width <= 0 {
			return
		}
		// Reserve space for label.
		suffix := label
		if suffix == "" {
			suffix = fmt.Sprintf(" %3.0f%%", pct)
		}
		barW := l.Width - len([]rune(suffix))
		if barW < 1 {
			barW = 1
		}
		filledW := int(pct / 100.0 * float64(barW))
		if filledW > barW {
			filledW = barW
		}
		for i := 0; i < barW; i++ {
			if i < filledW {
				c.Set(i, 0, '█', filled)
			} else {
				c.Set(i, 0, '░', empty)
			}
		}
		if barW < l.Width {
			c.WriteText(barW, 0, strings.TrimRight(suffix, " "), filled)
		}
	})
}
