package termyx

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// ResizeEvent is included in InputEvent when the terminal is resized.
type ResizeEvent struct {
	Width, Height int
}

// Backend abstracts the terminal I/O layer used by the runtime.
// The zero value (nil in App.Backend) uses the built-in ANSI backend.
type Backend interface {
	// Init sets up the terminal. Called once at startup.
	Init() error
	// Restore restores the terminal to its previous state. Called on exit.
	Restore()
	// Size returns the current terminal dimensions.
	Size() (width, height int)
	// Paint renders a completed Buffer frame to the terminal.
	Paint(*Buffer)
	// InputCh returns the channel on which all input events are delivered:
	// keyboard events, mouse events, and resize events.
	InputCh() <-chan InputEvent
	// EnableMouse activates or deactivates mouse event reporting.
	EnableMouse(enable bool)
}

// ── ANSI backend (default) ────────────────────────────────────────────────────

// ANSIBackend is the default terminal backend: raw ANSI escape sequences,
// SIGWINCH resize detection, and the built-in ANSI painter.
type ANSIBackend struct {
	stdinFD int
	state   *term.State
	painter *Painter
	inputC  chan InputEvent
	sigwinch chan os.Signal
	mouse   bool
}

// NewANSIBackend creates the default backend. Most applications do not need to
// call this directly — leaving App.Backend nil selects it automatically.
func NewANSIBackend() *ANSIBackend {
	return &ANSIBackend{
		stdinFD: int(os.Stdin.Fd()),
		inputC:  make(chan InputEvent, 16),
	}
}

func (b *ANSIBackend) Init() error {
	state, err := term.MakeRaw(b.stdinFD)
	if err != nil {
		return fmt.Errorf("termyx: raw mode: %w", err)
	}
	b.state = state
	// Enter alternate screen so terminal contents are preserved on exit.
	fmt.Fprint(os.Stdout, "\x1b[?1049h\x1b[2J\x1b[H")
	b.painter = NewPainter(os.Stdout)

	// SIGWINCH → resize events.
	b.sigwinch = make(chan os.Signal, 1)
	signal.Notify(b.sigwinch, syscall.SIGWINCH)

	// Keyboard / mouse reader goroutine.
	go func() {
		for {
			ev, err := ReadInput()
			if err != nil {
				return
			}
			b.inputC <- ev
		}
	}()

	// SIGWINCH → ResizeEvent goroutine.
	go func() {
		for range b.sigwinch {
			w, h, _ := term.GetSize(int(os.Stdout.Fd()))
			b.painter.Invalidate()
			b.inputC <- InputEvent{Resize: &ResizeEvent{Width: w, Height: h}}
		}
	}()

	return nil
}

func (b *ANSIBackend) Restore() {
	if b.mouse {
		fmt.Fprint(os.Stdout, mouseTrackingOff())
	}
	signal.Stop(b.sigwinch)
	close(b.sigwinch)
	if b.state != nil {
		term.Restore(b.stdinFD, b.state)
	}
	// Exit alternate screen, restore cursor and attributes.
	fmt.Print("\x1b[?25h\x1b[0m\x1b[?1049l")
}

// Suspend temporarily restores the terminal (for running external commands).
// Call Resume when done.
func (b *ANSIBackend) Suspend() {
	if b.mouse {
		fmt.Fprint(os.Stdout, mouseTrackingOff())
	}
	fmt.Print("\x1b[?25h\x1b[0m\x1b[?1049l")
	if b.state != nil {
		term.Restore(b.stdinFD, b.state)
	}
}

// Resume re-enters raw mode and alt screen after a Suspend.
func (b *ANSIBackend) Resume() error {
	state, err := term.MakeRaw(b.stdinFD)
	if err != nil {
		return err
	}
	b.state = state
	fmt.Fprint(os.Stdout, "\x1b[?1049h\x1b[2J\x1b[H")
	if b.mouse {
		fmt.Fprint(os.Stdout, mouseTrackingOn())
	}
	b.painter.Invalidate()
	return nil
}

func (b *ANSIBackend) Size() (int, int) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80, 24
	}
	return w, h
}

func (b *ANSIBackend) Paint(buf *Buffer) {
	b.painter.Paint(buf)
}

func (b *ANSIBackend) InputCh() <-chan InputEvent {
	return b.inputC
}

func (b *ANSIBackend) EnableMouse(enable bool) {
	if enable == b.mouse {
		return
	}
	b.mouse = enable
	if enable {
		fmt.Fprint(os.Stdout, mouseTrackingOn())
	} else {
		fmt.Fprint(os.Stdout, mouseTrackingOff())
	}
}
