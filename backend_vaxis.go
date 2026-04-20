package termyx

import (
	vx "git.sr.ht/~rockorager/vaxis"
)

// VaxisBackend uses the Vaxis terminal library as the rendering and input layer.
// Vaxis provides better compatibility with modern terminal features: Kitty
// keyboard protocol, Sixel graphics, SGR-Pixels mouse, and more.
//
// Usage:
//
//	termyx.Run(&termyx.App{
//	    Backend: termyx.NewVaxisBackend(),
//	    Root:    func() *termyx.Node { … },
//	    OnKey:   func(ev termyx.KeyEvent) bool { … },
//	})
type VaxisBackend struct {
	vaxis  *vx.Vaxis
	inputC chan InputEvent
	width  int
	height int
}

// NewVaxisBackend creates a VaxisBackend. Call this and assign to App.Backend.
func NewVaxisBackend() *VaxisBackend {
	return &VaxisBackend{
		inputC: make(chan InputEvent, 16),
	}
}

func (b *VaxisBackend) Init() error {
	v, err := vx.New(vx.Options{})
	if err != nil {
		return err
	}
	b.vaxis = v
	b.width, b.height = v.Window().Size()

	// Pump vaxis events into our InputEvent channel.
	go func() {
		for ev := range v.Events() {
			switch e := ev.(type) {
			case vx.Key:
				kev := vaxisKeyToKeyEvent(e)
				b.inputC <- InputEvent{Key: &kev}

			case vx.Mouse:
				mev := vaxisMouseToMouseEvent(e)
				b.inputC <- InputEvent{Mouse: &mev}

			case vx.Resize:
				b.width = e.Cols
				b.height = e.Rows
				b.inputC <- InputEvent{Resize: &ResizeEvent{Width: e.Cols, Height: e.Rows}}

			case vx.QuitEvent:
				return
			}
		}
	}()
	return nil
}

func (b *VaxisBackend) Restore() {
	if b.vaxis != nil {
		b.vaxis.Close()
	}
}

func (b *VaxisBackend) Size() (int, int) {
	return b.width, b.height
}

func (b *VaxisBackend) Paint(buf *Buffer) {
	if b.vaxis == nil {
		return
	}
	win := b.vaxis.Window()
	for y := 0; y < buf.Height && y < b.height; y++ {
		for x := 0; x < buf.Width && x < b.width; x++ {
			cell := buf.Cells[y][x]
			if cell.Wide {
				continue // right half; left-half write already advanced cursor
			}
			r := cell.Rune
			if r == 0 {
				r = ' '
			}
			win.SetCell(x, y, vx.Cell{
				Character: vx.Character{Grapheme: string(r), Width: 1},
				Style:     termyxStyleToVaxis(cell.Style),
			})
		}
	}
	b.vaxis.Render()
}

func (b *VaxisBackend) InputCh() <-chan InputEvent {
	return b.inputC
}

func (b *VaxisBackend) EnableMouse(enable bool) {
	// Vaxis handles mouse mode internally via its Options or capability detection.
	// No-op: vaxis enables mouse automatically when the terminal supports it.
}

// ── Type mapping helpers ──────────────────────────────────────────────────────

func termyxStyleToVaxis(s Style) vx.Style {
	var attr vx.AttributeMask
	if s.Bold {
		attr |= vx.AttrBold
	}
	if s.Italic {
		attr |= vx.AttrItalic
	}
	// Note: vaxis does not have AttrUnderline in AttributeMask;
	// underline is set via Style.UnderlineStyle instead.
	_ = s.Underline
	if s.Reverse {
		attr |= vx.AttrReverse
	}
	if s.Strikethrough {
		attr |= vx.AttrStrikethrough
	}
	vs := vx.Style{Attribute: attr}
	if s.FG.Set {
		vs.Foreground = vx.RGBColor(s.FG.R, s.FG.G, s.FG.B)
	}
	if s.BG.Set {
		vs.Background = vx.RGBColor(s.BG.R, s.BG.G, s.BG.B)
	}
	return vs
}

func vaxisKeyToKeyEvent(k vx.Key) KeyEvent {
	// Check modifiers for Ctrl combinations.
	ctrl := k.Modifiers&vx.ModCtrl != 0
	shift := k.Modifiers&vx.ModShift != 0

	switch k.Keycode {
	case vx.KeyUp:
		return KeyEvent{Key: KeyUp}
	case vx.KeyDown:
		return KeyEvent{Key: KeyDown}
	case vx.KeyLeft:
		return KeyEvent{Key: KeyLeft}
	case vx.KeyRight:
		return KeyEvent{Key: KeyRight}
	case vx.KeyEnter:
		return KeyEvent{Key: KeyEnter}
	case vx.KeyBackspace:
		return KeyEvent{Key: KeyBackspace}
	case vx.KeyDelete:
		return KeyEvent{Key: KeyDelete}
	case vx.KeyEsc:
		return KeyEvent{Key: KeyEscape}
	case vx.KeyTab:
		if shift {
			return KeyEvent{Key: KeyShiftTab}
		}
		return KeyEvent{Key: KeyTab}
	case vx.KeyHome:
		return KeyEvent{Key: KeyHome}
	case vx.KeyEnd:
		return KeyEvent{Key: KeyEnd}
	case vx.KeyPgUp:
		return KeyEvent{Key: KeyPageUp}
	case vx.KeyPgDown:
		return KeyEvent{Key: KeyPageDown}
	case vx.KeyF01:
		return KeyEvent{Key: KeyF1}
	case vx.KeyF02:
		return KeyEvent{Key: KeyF2}
	case vx.KeyF03:
		return KeyEvent{Key: KeyF3}
	case vx.KeyF04:
		return KeyEvent{Key: KeyF4}
	case vx.KeyF05:
		return KeyEvent{Key: KeyF5}
	case vx.KeyF06:
		return KeyEvent{Key: KeyF6}
	case vx.KeyF07:
		return KeyEvent{Key: KeyF7}
	case vx.KeyF08:
		return KeyEvent{Key: KeyF8}
	case vx.KeyF09:
		return KeyEvent{Key: KeyF9}
	case vx.KeyF10:
		return KeyEvent{Key: KeyF10}
	case vx.KeyF11:
		return KeyEvent{Key: KeyF11}
	case vx.KeyF12:
		return KeyEvent{Key: KeyF12}
	}

	// Ctrl+letter combinations.
	if ctrl && k.Keycode >= 'a' && k.Keycode <= 'z' {
		switch k.Keycode {
		case 'c':
			return KeyEvent{Key: KeyCtrlC}
		case 'd':
			return KeyEvent{Key: KeyCtrlD}
		case 'l':
			return KeyEvent{Key: KeyCtrlL}
		}
	}

	// Regular printable rune.
	if k.Keycode > 0 && k.Keycode < 0xE000 {
		return KeyEvent{Rune: k.Keycode}
	}
	return KeyEvent{}
}

func vaxisMouseToMouseEvent(m vx.Mouse) MouseEvent {
	var btn MouseButton
	var act MouseAction

	switch m.Button {
	case vx.MouseLeftButton:
		btn = MouseLeft
	case vx.MouseMiddleButton:
		btn = MouseMiddle
	case vx.MouseRightButton:
		btn = MouseRight
	case vx.MouseWheelUp:
		btn = MouseScrollUp
	case vx.MouseWheelDown:
		btn = MouseScrollDown
	default:
		btn = MouseButtonNone
	}

	switch m.EventType {
	case vx.EventPress:
		act = MousePress
	case vx.EventRelease:
		act = MouseRelease
	case vx.EventMotion:
		act = MouseMove
	}

	return MouseEvent{X: m.Col, Y: m.Row, Button: btn, Action: act}
}
