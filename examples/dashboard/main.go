package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/voxire/termyx/ui"
)

// ── styles ───────────────────────────────────────────────────────────────────

var (
	styleHeader  = ui.Style{FG: ui.RGB(30, 215, 96), Bold: true}
	styleValue   = ui.Style{FG: ui.RGB(255, 200, 50)}
	styleDim     = ui.Style{FG: ui.RGB(100, 100, 110)}
	styleBorder  = ui.Style{FG: ui.RGB(60, 80, 120)}
	styleLog     = ui.Style{FG: ui.RGB(180, 180, 190)}
	styleLogTime = ui.Style{FG: ui.RGB(80, 100, 140)}
)

// ── state ────────────────────────────────────────────────────────────────────

type state struct {
	mu      sync.Mutex
	logs    []string
	cpu     float64
	mem     float64
	uptime  int
	lastKey string
}

func newState() *state {
	return &state{
		logs: []string{"[termyx] dashboard started"},
		cpu:  rand.Float64() * 60,
		mem:  rand.Float64() * 50,
	}
}

func (s *state) tick() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.uptime++
	s.cpu = clamp(s.cpu+rand.Float64()*12-6, 1, 99)
	s.mem = clamp(s.mem+rand.Float64()*6-3, 10, 95)
	s.logs = append(s.logs, fmt.Sprintf(
		"[%04d] cpu=%-5.1f%% mem=%-5.1f%%  goroutines=%d",
		s.uptime, s.cpu, s.mem, rand.Intn(200)+10,
	))
	if len(s.logs) > 2000 {
		s.logs = s.logs[1000:]
	}
}

// ── UI tree ──────────────────────────────────────────────────────────────────

func buildUI(s *state) *ui.Node {
	s.mu.Lock()
	defer s.mu.Unlock()

	return ui.Row(
		// Left 60%: log stream
		ui.Grow(
			ui.Border(" LOGS ", styleBorder, logPane(s.logs)),
			1.5,
		),
		// Right 40%: metrics + input
		ui.Column(
			ui.Border(" METRICS ", styleBorder, metricsPane(s.cpu, s.mem, s.uptime)),
			ui.Border(" INPUT ", styleBorder, inputPane(s.lastKey)),
		),
	)
}

func logPane(logs []string) *ui.Node {
	return ui.Custom(func(buf *ui.Buffer, l ui.LayoutResult) {
		if l.Width <= 0 || l.Height <= 0 {
			return
		}
		// Show only the last l.Height lines
		start := len(logs) - l.Height
		if start < 0 {
			start = 0
		}
		visible := logs[start:]

		for i, line := range visible {
			if i >= l.Height {
				break
			}
			// Dim the timestamp prefix ([0042]) and normal style for the rest
			if len(line) > 6 && line[0] == '[' {
				buf.WriteText(l.X, l.Y+i, line[:6], styleLogTime)
				rest := line[6:]
				if len(rest) > l.Width-6 {
					rest = rest[:l.Width-6]
				}
				buf.WriteText(l.X+6, l.Y+i, rest, styleLog)
			} else {
				if len(line) > l.Width {
					line = line[:l.Width]
				}
				buf.WriteText(l.X, l.Y+i, line, styleLog)
			}
		}
	})
}

func metricsPane(cpu, mem float64, uptime int) *ui.Node {
	cpuBar := progressBar(cpu, 22)
	memBar := progressBar(mem, 22)

	lines := []struct {
		label string
		value string
	}{
		{"CPU", fmt.Sprintf("%s  %.1f%%", cpuBar, cpu)},
		{"MEM", fmt.Sprintf("%s  %.1f%%", memBar, mem)},
		{"UP ", fmt.Sprintf("%ds  (%s)", uptime, formatDuration(uptime))},
	}

	return ui.Custom(func(buf *ui.Buffer, l ui.LayoutResult) {
		y := l.Y + 1
		for _, ln := range lines {
			if y >= l.Y+l.Height {
				break
			}
			buf.WriteText(l.X+2, y, ln.label, styleHeader)
			buf.WriteText(l.X+6, y, ln.value, styleValue)
			y += 2
		}
	})
}

func inputPane(lastKey string) *ui.Node {
	return ui.Custom(func(buf *ui.Buffer, l ui.LayoutResult) {
		if l.Height < 1 {
			return
		}
		label := "  last key: "
		buf.WriteText(l.X, l.Y+1, label, styleDim)
		buf.WriteText(l.X+len(label), l.Y+1, lastKey, styleValue)

		hint := "  q / ctrl+c to quit"
		if l.Height > 3 {
			buf.WriteText(l.X, l.Y+3, hint, styleDim)
		}
	})
}

// ── helpers ──────────────────────────────────────────────────────────────────

func progressBar(pct float64, width int) string {
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}

func formatDuration(secs int) string {
	h := secs / 3600
	m := (secs % 3600) / 60
	s := secs % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func notify(ch chan<- struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}

func keyName(ev ui.KeyEvent) string {
	if ev.Rune != 0 {
		return string(ev.Rune)
	}
	switch ev.Key {
	case ui.KeyUp:
		return "↑"
	case ui.KeyDown:
		return "↓"
	case ui.KeyLeft:
		return "←"
	case ui.KeyRight:
		return "→"
	case ui.KeyEnter:
		return "Enter"
	case ui.KeyBackspace:
		return "Backspace"
	case ui.KeyDelete:
		return "Delete"
	case ui.KeyEscape:
		return "Esc"
	case ui.KeyTab:
		return "Tab"
	case ui.KeyCtrlL:
		return "Ctrl+L"
	case ui.KeyHome:
		return "Home"
	case ui.KeyEnd:
		return "End"
	case ui.KeyPageUp:
		return "PgUp"
	case ui.KeyPageDown:
		return "PgDn"
	default:
		return ""
	}
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	s := newState()
	update := make(chan struct{}, 1)

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			s.tick()
			notify(update)
		}
	}()

	app := &ui.App{
		Root: func() *ui.Node {
			return buildUI(s)
		},
		OnKey: func(ev ui.KeyEvent) bool {
			if ev.Key == ui.KeyCtrlC || ev.Key == ui.KeyCtrlD || ev.Rune == 'q' {
				return true
			}
			s.mu.Lock()
			s.lastKey = keyName(ev)
			s.mu.Unlock()
			return false
		},
		Update: update,
	}

	if err := ui.Run(app); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
