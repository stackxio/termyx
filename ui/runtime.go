package ui

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
	OnKey func(KeyEvent) bool

	// Update is an optional channel that signals a re-render.
	// Useful for triggering redraws from background goroutines.
	Update <-chan struct{}
}

// Run starts the Termyx event loop, blocking until the app exits.
// It puts the terminal in raw mode, handles SIGWINCH, and routes
// keyboard events through App.OnKey.
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

	sigwinch := make(chan os.Signal, 1)
	signal.Notify(sigwinch, syscall.SIGWINCH)
	defer signal.Stop(sigwinch)

	keyc := make(chan KeyEvent, 8)
	go func() {
		for {
			ev, err := ReadKey()
			if err != nil {
				return
			}
			keyc <- ev
		}
	}()

	var prevTree *Node

	frame := func() {
		next := app.Root()
		next = Reconcile(prevTree, next)
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

		case ev := <-keyc:
			if app.OnKey != nil && app.OnKey(ev) {
				return nil
			}
			frame()

		case <-update:
			frame()
		}
	}
}
