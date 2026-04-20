package termyx

import (
	"fmt"
	"io"
	"os"
	"os/exec"
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

	// EnableMouse activates terminal mouse reporting.
	// When true, click, scroll, and motion events are delivered via OnMouse
	// and per-node Clickable handlers.
	EnableMouse bool

	// Update is an optional channel that signals a re-render.
	// Useful for triggering redraws from background goroutines.
	Update <-chan struct{}

	// InitialFocus is the key of the node that receives focus on startup.
	// Leave empty to start with no focused node.
	InitialFocus string

	// Backend selects the terminal I/O layer.
	// Nil (default) uses the built-in ANSI painter.
	// Set to NewVaxisBackend() to use the Vaxis library.
	Backend Backend
}

// ExecProcess suspends the Termyx runtime, runs cmd with the terminal
// connected to stdin/stdout/stderr, then restores the TUI. Use this for
// interactive commands like `kubectl exec -it`.
//
// Only works with ANSIBackend (the default). Returns the command's error.
// Typical usage in an OnKey handler:
//
//	termyx.ExecProcess(exec.Command("kubectl", "exec", "-it", pod, "--", "/bin/sh"))
//
// Because Run is blocking, wire this via App.Update: run the command in a
// goroutine, send on updateCh when done, and redraw.
func ExecProcess(be Backend, cmd *exec.Cmd) error {
	ab, ok := be.(*ANSIBackend)
	if !ok {
		return fmt.Errorf("termyx: ExecProcess requires ANSIBackend")
	}
	ab.Suspend()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if resumeErr := ab.Resume(); resumeErr != nil && err == nil {
		err = resumeErr
	}
	return err
}

// ExecProcessPiped runs cmd with piped stdout/stderr for capturing output,
// suspending the TUI first. Useful for `kubectl describe` / `kubectl logs`.
func ExecProcessPiped(be Backend, cmd *exec.Cmd, stdout, stderr io.Writer) error {
	ab, ok := be.(*ANSIBackend)
	if !ok {
		return fmt.Errorf("termyx: ExecProcessPiped requires ANSIBackend")
	}
	ab.Suspend()
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if resumeErr := ab.Resume(); resumeErr != nil && err == nil {
		err = resumeErr
	}
	return err
}

// Run starts the Termyx event loop, blocking until the app exits.
func Run(app *App) error {
	be := app.Backend
	if be == nil {
		be = NewANSIBackend()
	}

	if err := be.Init(); err != nil {
		return fmt.Errorf("termyx: backend init: %w", err)
	}
	defer be.Restore()

	if app.EnableMouse {
		be.EnableMouse(true)
	}

	width, height := be.Size()
	painter := &painterAdapter{be: be}

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

		focusOrder = focusOrder[:0]
		collectFocusOrder(next, &focusOrder)

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

	inputC := be.InputCh()

	for {
		select {
		case ev := <-inputC:
			if ev.Resize != nil {
				width = ev.Resize.Width
				height = ev.Resize.Height
				painter.Invalidate()
				frame()
				continue
			}

			if ev.Mouse != nil {
				mouse := *ev.Mouse
				handled := false
				if prevTree != nil {
					if hit := hitTest(prevTree, mouse.X, mouse.Y); hit != nil && hit.Props.OnMouse != nil {
						hit.Props.OnMouse(mouse)
						handled = true
					}
				}
				if !handled && app.OnMouse != nil && app.OnMouse(mouse) {
					return nil
				}
				frame()
				continue
			}

			if ev.Key != nil {
				key := *ev.Key
				if key.Key == KeyTab {
					cycleFocus(true)
					frame()
					continue
				}
				if key.Key == KeyShiftTab {
					cycleFocus(false)
					frame()
					continue
				}

				routed := false
				if focusedID != "" && prevTree != nil {
					if node := findNode(prevTree, focusedID); node != nil && node.Props.OnKey != nil {
						node.Props.OnKey(key)
						routed = true
					}
				}

				if !routed && app.OnKey != nil && app.OnKey(key) {
					return nil
				}
				if routed && app.OnKey != nil {
					if app.OnKey(key) {
						return nil
					}
				}
				frame()
			}

		case <-update:
			frame()
		}
	}
}

// painterAdapter wraps a Backend to look like the old Painter interface.
type painterAdapter struct {
	be      Backend
	invalid bool
}

func (p *painterAdapter) Paint(buf *Buffer) {
	p.be.Paint(buf)
}

func (p *painterAdapter) Invalidate() {
	// ANSIBackend's Invalidate is implicit via resize.
	if ab, ok := p.be.(*ANSIBackend); ok {
		ab.painter.Invalidate()
	}
}
