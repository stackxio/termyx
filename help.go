package termyx

import "strings"

// KeyBinding associates a human-readable key label with a description.
//
//	termyx.KeyBinding{Key: "↑/↓", Desc: "navigate"}
//	termyx.KeyBinding{Key: "Enter", Desc: "select"}
//	termyx.KeyBinding{Key: "q", Desc: "quit"}
type KeyBinding struct {
	Key  string // display label, e.g. "Ctrl+C", "↑/↓", "Tab"
	Desc string // short description, e.g. "quit", "move up/down"
}

// KeyMap is an ordered slice of KeyBindings for one context.
type KeyMap []KeyBinding

// HelpStyle holds the visual styles for a Help panel.
type HelpStyle struct {
	Key   Style // style applied to the key label column
	Desc  Style // style applied to the description column
	Title Style // style for the optional title row
	BG    Style // background fill style
}

// HelpModel holds the state for a Help overlay.
type HelpModel struct {
	Title    string  // optional panel title (empty = no title row)
	Bindings KeyMap  // the bindings to display
	Visible  bool    // whether the overlay is shown
	Style    HelpStyle
}

// Toggle shows or hides the help overlay.
func (h *HelpModel) Toggle() { h.Visible = !h.Visible }

// Show makes the help overlay visible.
func (h *HelpModel) Show() { h.Visible = true }

// Hide hides the help overlay.
func (h *HelpModel) Hide() { h.Visible = false }

// Help renders a formatted two-column binding table (key | description).
// Typically used inside an Overlay node so it floats over the rest of the UI.
//
//	termyx.Overlay(base, termyx.Help(helpModel), helpModel.Visible,
//	    termyx.OverlayOpts{Width: 40, Height: 20})
func Help(model *HelpModel) *Node {
	return Custom(func(buf *Buffer, l LayoutResult) {
		if !model.Visible || l.Width <= 0 || l.Height <= 0 {
			return
		}
		c := buf.Region(l)

		// Fill background.
		c.Fill(0, 0, l.Width, l.Height, ' ', model.Style.BG)

		y := 0
		// Title row.
		if model.Title != "" {
			title := Center(model.Title, l.Width)
			c.WriteText(0, y, title, model.Style.Title)
			y++
			// Divider.
			if y < l.Height {
				c.HLine(y, '─', model.Style.Title)
				y++
			}
		}

		// Compute key column width: longest key label + 2 padding.
		keyW := 0
		for _, b := range model.Bindings {
			if w := RuneWidth(b.Key); w > keyW {
				keyW = w
			}
		}
		keyW += 2 // padding
		descW := l.Width - keyW - 3 // " │ " separator

		for _, b := range model.Bindings {
			if y >= l.Height {
				break
			}
			// Key column.
			keyLabel := PadRight(b.Key, keyW)
			c.WriteText(0, y, keyLabel, model.Style.Key)
			// Separator.
			if keyW+1 < l.Width {
				c.Set(keyW, y, '│', model.Style.BG)
				c.Set(keyW+1, y, ' ', model.Style.BG)
			}
			// Description column (may wrap).
			if descW > 0 {
				descLines := WordWrap(b.Desc, descW)
				for di, dl := range descLines {
					if y+di >= l.Height {
						break
					}
					if di > 0 {
						// continuation line: indent key column with spaces
						c.WriteText(0, y+di, strings.Repeat(" ", keyW), model.Style.Key)
						if keyW+1 < l.Width {
							c.Set(keyW, y+di, '│', model.Style.BG)
							c.Set(keyW+1, y+di, ' ', model.Style.BG)
						}
					}
					c.WriteText(keyW+2, y+di, dl, model.Style.Desc)
				}
				if len(descLines) > 1 {
					y += len(descLines) - 1
				}
			}
			y++
		}
	})
}

// HelpBar renders a compact single-line summary of bindings for a status bar.
// Bindings are formatted as "key desc  key desc  …" and truncated to fit.
//
//	termyx.HelpBar(bindings, dimStyle)
func HelpBar(bindings KeyMap, keyStyle, descStyle Style) *Node {
	return Custom(func(buf *Buffer, l LayoutResult) {
		c := buf.Region(l)
		x := 0
		for i, b := range bindings {
			if x >= l.Width {
				break
			}
			// Key label.
			key := b.Key
			for _, r := range []rune(key) {
				if x >= l.Width {
					break
				}
				c.Set(x, 0, r, keyStyle)
				x++
			}
			// Space between key and desc.
			if x < l.Width {
				c.Set(x, 0, ' ', descStyle)
				x++
			}
			// Description.
			for _, r := range []rune(b.Desc) {
				if x >= l.Width {
					break
				}
				c.Set(x, 0, r, descStyle)
				x++
			}
			// Gap between bindings (except last).
			if i < len(bindings)-1 && x+2 < l.Width {
				c.Set(x, 0, ' ', descStyle)
				c.Set(x+1, 0, ' ', descStyle)
				x += 2
			}
		}
	})
}

// DefaultHelpStyle derives a HelpStyle from a Theme.
func (t Theme) StyleHelp() HelpStyle {
	return HelpStyle{
		Key:   Style{FG: t.Accent, BG: t.BgAlt, Bold: true},
		Desc:  Style{FG: t.Fg, BG: t.BgAlt},
		Title: Style{FG: t.FgBright, BG: t.BgAlt, Bold: true},
		BG:    Style{FG: t.Fg, BG: t.BgAlt},
	}
}
