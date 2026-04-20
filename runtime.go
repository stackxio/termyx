package termyx

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// App defines the Termyx application.
type App struct {
	// Root is called each frame to produce the current UI tree.
	Root func() *Node

	// OnKey handles keyboard input. Return true to exit the runtime.
	// Keyboard events are first dispatched to the focused node's OnKey handler;
	// unhandled events (where the node's OnKey returns false) fall through here.
	OnKey func(KeyEvent) bool

	// OnMouse handles mouse events at the application level.
	// Return true to exit the runtime (rare; usually return false).
	// Mouse events are first dispatched to the deepest node under the cursor
	// that has an OnMouse handler; unhandled events bubble here.
	OnMouse func(MouseEvent) bool

	// EnableMouse activates terminal mouse reporting (SGR protocol).
	// When true, click, scroll, and motion events are delivered via OnMouse
	// and per-node Clickable handlers. Defaults to false.
	EnableMouse bool

	// Update is an optional channel that signals a re-render.
	// Useful for triggering redraws from background goroutines.
	Update <-chan struct{}

	// InitialFocus is the key of the node that receives focus on startup.
	// Leave empty to start with no focused node.
	InitialFocus string
}

// Run starts the Termyx event loop, blocking until the app exits.
// It puts the terminal in raw mode, handles SIGWINCH, and routes
// keyboard events through the focus system and App.OnKey.
func Run(app *App) error {
	stdinFD := int(os.Stdin.Fd())

	state, err := term.MakeRaw(stdinFD)
	if err != nil {
		return fmt.Errorf("termyx: raw mode: %w", err)
	}
	defer func() {
		term.Restore(stdinFD, state)
		fmt.Print("\x1b[?25h\x1b[0m") // restore cursor + reset attributes
	}()

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width, height = 80, 24
	}

	painter := NewPainter(os.Stdout)

	if app.EnableMouse {
		fmt.Fprint(os.Stdout, mouseTrackingOn())
		defer fmt.Fprint(os.Stdout, mouseTrackingOff())
	}

	sigwinch := make(chan os.Signal, 1)
	signal.Notify(sigwinch, syscall.SIGWINCH)
	defer signal.Stop(sigwinch)

	type inputMsg struct {
		key   *KeyEvent
		mouse *MouseEvent
	}
	inputc := make(chan inputMsg, 8)
	go func() {
		for {
			ev, err := ReadInput()
			if err != nil {
				return
			}
			inputc <- inputMsg{key: ev.Key, mouse: ev.Mouse}
		}
	}()

	var (
		prevTree   *Node
		focusedID  string
		focusOrder []string
	)

	if app.InitialFocus != "" {
		focusedID = app.InitialFocus
	}

	cycleFocus := func(forward bool) {
		if len(focusOrder) == 0 {
			return
		}
		idx := -1
		for i, id := range focusOrder {
			if id == focusedID {
				idx = i
				break
			}
		}
		if forward {
			idx = (idx + 1) % len(focusOrder)
		} else {
			idx = (idx - 1 + len(focusOrder)) % len(focusOrder)
		}
		focusedID = focusOrder[idx]
	}

	frame := func() {
		next := app.Root()
		next = Reconcile(prevTree, next)

		// Collect focusable nodes and apply focus state before rendering.
		focusOrder = focusOrder[:0]
		collectFocusOrder(next, &focusOrder)

		// Auto-focus first node if InitialFocus was set as a key
		if focusedID == "" && len(focusOrder) > 0 && app.InitialFocus == "first" {
			focusedID = focusOrder[0]
		}

		applyFocus(next, focusedID)

		ComputeLayout(next, 0, 0, width, height)
		buf := NewBuffer(width, height)
		Render(next, buf)
		painter.Paint(buf)
		prevTree = next
	}

	frame()

	update := app.Update
	if update == nil {
		update = make(chan struct{}) // never fires
	}

	for {
		select {
		case <-sigwinch:
			width, height, _ = term.GetSize(int(os.Stdout.Fd()))
			painter.Invalidate()
			frame()

		case msg := <-inputc:
			if msg.mouse != nil {
				ev := *msg.mouse
				// Dispatch to deepest node with an OnMouse handler.
				handled := false
				if prevTree != nil {
					if hit := hitTest(prevTree, ev.X, ev.Y); hit != nil && hit.Props.OnMouse != nil {
						hit.Props.OnMouse(ev)
						handled = true
					}
				}
				if !handled && app.OnMouse != nil && app.OnMouse(ev) {
					return nil
				}
				frame()
				continue
			}

			ev := *msg.key
			// Tab / Shift+Tab cycle focus.
			if ev.Key == KeyTab {
				cycleFocus(true)
				frame()
				continue
			}
			if ev.Key == KeyShiftTab {
				cycleFocus(false)
				frame()
				continue
			}

			// Route to the focused node's OnKey handler first.
			routed := false
			if focusedID != "" && prevTree != nil {
				if node := findNode(prevTree, focusedID); node != nil && node.Props.OnKey != nil {
					node.Props.OnKey(ev)
					routed = true
				}
			}

			// Fall through to app-level handler for unrouted or unfocused events.
			if !routed && app.OnKey != nil && app.OnKey(ev) {
				return nil
			}
			if routed && app.OnKey != nil {
				// Also give app a chance to handle quit (ctrl+c) even when focused.
				if app.OnKey(ev) {
					return nil
				}
			}

			frame()

		case <-update:
			frame()
		}
	}
}
