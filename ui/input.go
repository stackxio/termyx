package ui

import "os"

// Key represents a named keyboard key.
type Key int

const (
	KeyNone      Key = iota
	KeyUp            // ↑
	KeyDown          // ↓
	KeyLeft          // ←
	KeyRight         // →
	KeyEnter         // Enter / Return
	KeyBackspace     // Backspace
	KeyDelete        // Delete
	KeyEscape        // Escape
	KeyTab           // Tab
	KeyCtrlC         // Ctrl+C
	KeyCtrlD         // Ctrl+D
	KeyCtrlL         // Ctrl+L
	KeyHome          // Home
	KeyEnd           // End
	KeyPageUp        // Page Up
	KeyPageDown      // Page Down
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
)

// KeyEvent represents a single keyboard input event.
type KeyEvent struct {
	Rune rune
	Key  Key
	Raw  []byte
}

// ReadKey reads one key event from stdin (blocking).
func ReadKey() (KeyEvent, error) {
	buf := make([]byte, 16)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return KeyEvent{}, err
	}
	return parseKey(buf[:n]), nil
}

func parseKey(b []byte) KeyEvent {
	if len(b) == 0 {
		return KeyEvent{Raw: b}
	}

	if b[0] == 0x1b && len(b) >= 3 && b[1] == '[' {
		switch {
		case b[2] == 'A':
			return KeyEvent{Key: KeyUp, Raw: b}
		case b[2] == 'B':
			return KeyEvent{Key: KeyDown, Raw: b}
		case b[2] == 'C':
			return KeyEvent{Key: KeyRight, Raw: b}
		case b[2] == 'D':
			return KeyEvent{Key: KeyLeft, Raw: b}
		case b[2] == 'H':
			return KeyEvent{Key: KeyHome, Raw: b}
		case b[2] == 'F':
			return KeyEvent{Key: KeyEnd, Raw: b}
		case len(b) >= 4 && b[2] == '5' && b[3] == '~':
			return KeyEvent{Key: KeyPageUp, Raw: b}
		case len(b) >= 4 && b[2] == '6' && b[3] == '~':
			return KeyEvent{Key: KeyPageDown, Raw: b}
		case len(b) >= 4 && b[2] == '3' && b[3] == '~':
			return KeyEvent{Key: KeyDelete, Raw: b}
		}
		return KeyEvent{Key: KeyEscape, Raw: b}
	}

	if b[0] == 0x1b && len(b) >= 3 && b[1] == 'O' {
		switch b[2] {
		case 'P':
			return KeyEvent{Key: KeyF1, Raw: b}
		case 'Q':
			return KeyEvent{Key: KeyF2, Raw: b}
		case 'R':
			return KeyEvent{Key: KeyF3, Raw: b}
		case 'S':
			return KeyEvent{Key: KeyF4, Raw: b}
		}
	}

	switch b[0] {
	case 0x1b:
		return KeyEvent{Key: KeyEscape, Raw: b}
	case '\r', '\n':
		return KeyEvent{Key: KeyEnter, Raw: b}
	case 127:
		return KeyEvent{Key: KeyBackspace, Raw: b}
	case '\t':
		return KeyEvent{Key: KeyTab, Raw: b}
	case 3:
		return KeyEvent{Key: KeyCtrlC, Raw: b}
	case 4:
		return KeyEvent{Key: KeyCtrlD, Raw: b}
	case 12:
		return KeyEvent{Key: KeyCtrlL, Raw: b}
	}

	runes := []rune(string(b))
	if len(runes) > 0 {
		return KeyEvent{Rune: runes[0], Raw: b}
	}
	return KeyEvent{Raw: b}
}
