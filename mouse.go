package termyx

import (
	"fmt"
	"strconv"
	"strings"
)

// MouseButton identifies which mouse button was involved in an event.
type MouseButton int

const (
	MouseLeft        MouseButton = iota // primary button
	MouseMiddle                         // middle / scroll-wheel click
	MouseRight                          // secondary button
	MouseScrollUp                       // scroll wheel up
	MouseScrollDown                     // scroll wheel down
	MouseButtonNone                     // motion-only event
)

// MouseAction describes what happened to the button.
type MouseAction int

const (
	MousePress   MouseAction = iota // button pressed
	MouseRelease                    // button released
	MouseMove                       // cursor moved (no button change)
)

// MouseEvent is emitted for mouse clicks, releases, and scroll events.
// X and Y are 0-based terminal column/row coordinates.
type MouseEvent struct {
	X, Y   int
	Button MouseButton
	Action MouseAction
}

// parseSGRMouse parses an SGR extended mouse report of the form
// \x1b[<Cb;Cx;CyM  (press) or \x1b[<Cb;Cx;CyM (release, trailing 'm').
// Returns (event, true) on success.
func parseSGRMouse(b []byte) (MouseEvent, bool) {
	// Expect: ESC [ < digits ; digits ; digits M|m
	if len(b) < 9 {
		return MouseEvent{}, false
	}
	if b[0] != 0x1b || b[1] != '[' || b[2] != '<' {
		return MouseEvent{}, false
	}
	s := string(b[3:])
	action := MousePress
	if s[len(s)-1] == 'm' {
		action = MouseRelease
	} else if s[len(s)-1] != 'M' {
		return MouseEvent{}, false
	}
	s = s[:len(s)-1]
	parts := strings.Split(s, ";")
	if len(parts) != 3 {
		return MouseEvent{}, false
	}
	cb, err1 := strconv.Atoi(parts[0])
	cx, err2 := strconv.Atoi(parts[1])
	cy, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return MouseEvent{}, false
	}

	var btn MouseButton
	switch {
	case cb&64 != 0:
		if cb&1 != 0 {
			btn = MouseScrollDown
		} else {
			btn = MouseScrollUp
		}
		action = MousePress
	case cb&32 != 0:
		btn = MouseButtonNone
		action = MouseMove
	default:
		switch cb & 3 {
		case 0:
			btn = MouseLeft
		case 1:
			btn = MouseMiddle
		case 2:
			btn = MouseRight
		default:
			btn = MouseButtonNone
		}
	}

	return MouseEvent{
		X:      cx - 1, // convert from 1-based
		Y:      cy - 1,
		Button: btn,
		Action: action,
	}, true
}

// parseX10Mouse parses a legacy X10 mouse report: ESC [ M <3 bytes>.
// Returns (event, true) on success.
func parseX10Mouse(b []byte) (MouseEvent, bool) {
	if len(b) < 6 || b[0] != 0x1b || b[1] != '[' || b[2] != 'M' {
		return MouseEvent{}, false
	}
	cb := int(b[3]) - 32
	cx := int(b[4]) - 33 // 1-based → 0-based
	cy := int(b[5]) - 33

	action := MousePress
	if cb&3 == 3 {
		action = MouseRelease
	}

	var btn MouseButton
	switch {
	case cb&64 != 0:
		if cb&1 != 0 {
			btn = MouseScrollDown
		} else {
			btn = MouseScrollUp
		}
		action = MousePress
	case cb&32 != 0:
		btn = MouseButtonNone
		action = MouseMove
	default:
		switch cb & 3 {
		case 0:
			btn = MouseLeft
		case 1:
			btn = MouseMiddle
		case 2:
			btn = MouseRight
		default:
			btn = MouseButtonNone
			action = MouseRelease
		}
	}

	return MouseEvent{X: cx, Y: cy, Button: btn, Action: action}, true
}

// mouseTrackingOn returns the ANSI escape sequences to enable SGR mouse
// reporting (button events + scroll, extended coordinates).
func mouseTrackingOn() string {
	return "\x1b[?1000h" + // button events
		"\x1b[?1002h" + // button-motion events
		"\x1b[?1006h" // SGR extended coords (>223 cols)
}

// mouseTrackingOff disables all mouse reporting modes.
func mouseTrackingOff() string {
	return "\x1b[?1006l\x1b[?1002l\x1b[?1000l"
}

// hitTest returns the deepest rendered node in the tree whose LayoutResult
// contains (x, y). Returns nil if no node covers that point.
func hitTest(node *Node, x, y int) *Node {
	l := node.Layout
	if x < l.X || x >= l.X+l.Width || y < l.Y || y >= l.Y+l.Height {
		return nil
	}
	// Depth-first: last child wins (highest z-order).
	for i := len(node.Children) - 1; i >= 0; i-- {
		if hit := hitTest(node.Children[i], x, y); hit != nil {
			return hit
		}
	}
	return node
}

// Clickable wraps a node so it receives mouse events when clicked.
// The OnMouse handler receives events whose coordinates fall within the node's
// rendered area.
func Clickable(node *Node, onMouse func(MouseEvent)) *Node {
	node.Props.OnMouse = onMouse
	return node
}

// String returns a human-readable description of a MouseEvent (for debugging).
func (ev MouseEvent) String() string {
	btn := []string{"Left", "Middle", "Right", "ScrollUp", "ScrollDown", "None"}[ev.Button]
	act := []string{"Press", "Release", "Move"}[ev.Action]
	return fmt.Sprintf("Mouse{%s %s at (%d,%d)}", btn, act, ev.X, ev.Y)
}
