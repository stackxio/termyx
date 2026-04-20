package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	termyx "github.com/stackxio/termyx"
)

// ── styles ───────────────────────────────────────────────────────────────────

var (
	styleHeader  = termyx.Style{FG: termyx.RGB(30, 215, 96), Bold: true}
	styleValue   = termyx.Style{FG: termyx.RGB(255, 200, 50)}
	styleDim     = termyx.Style{FG: termyx.RGB(100, 100, 110)}
	styleBorder  = termyx.Style{FG: termyx.RGB(60, 80, 120)}
	styleLog     = termyx.Style{FG: termyx.RGB(180, 180, 190)}
	styleLogTime = termyx.Style{FG: termyx.RGB(80, 100, 140)}
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

func buildUI(s *state) *termyx.Node {
	s.mu.Lock()
	defer s.mu.Unlock()

	return termyx.Row(
		termyx.Grow(
			termyx.Border(" LOGS ", styleBorder, logPane(s.logs)),
			1.5,
		),
		termyx.Column(
			termyx.Border(" METRICS ", styleBorder, metricsPane(s.cpu, s.mem, s.uptime)),
			termyx.Border(" INPUT ", styleBorder, inputPane(s.lastKey)),
		),
	)
}

func logPane(logs []string) *termyx.Node {
	return termyx.Custom(func(buf *termyx.Buffer, l termyx.LayoutResult) {
		if l.Width <= 0 || l.Height <= 0 {
			return
		}
		start := len(logs) - l.Height
		if start < 0 {
			start = 0
		}
		for i, line := range logs[start:] {
			if i >= l.Height {
				break
			}
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

func metricsPane(cpu, mem float64, uptime int) *termyx.Node {
	cpuBar := progressBar(cpu, 22)
	memBar := progressBar(mem, 22)
	rows := []struct{ label, value string }{
		{"CPU", fmt.Sprintf("%s  %.1f%%", cpuBar, cpu)},
		{"MEM", fmt.Sprintf("%s  %.1f%%", memBar, mem)},
		{"UP ", fmt.Sprintf("%ds  (%s)", uptime, formatDuration(uptime))},
	}
	return termyx.Custom(func(buf *termyx.Buffer, l termyx.LayoutResult) {
		y := l.Y + 1
		for _, r := range rows {
			if y >= l.Y+l.Height {
				break
			}
			buf.WriteText(l.X+2, y, r.label, styleHeader)
			buf.WriteText(l.X+6, y, r.value, styleValue)
			y += 2
		}
	})
}

func inputPane(lastKey string) *termyx.Node {
	return termyx.Custom(func(buf *termyx.Buffer, l termyx.LayoutResult) {
		if l.Height < 1 {
			return
		}
		label := "  last key: "
		buf.WriteText(l.X, l.Y+1, label, styleDim)
		buf.WriteText(l.X+len(label), l.Y+1, lastKey, styleValue)
		if l.Height > 3 {
			buf.WriteText(l.X, l.Y+3, "  q / ctrl+c to quit", styleDim)
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
	h, m, s := secs/3600, (secs%3600)/60, secs%60
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

func keyName(ev termyx.KeyEvent) string {
	if ev.Rune != 0 {
		return string(ev.Rune)
	}
	switch ev.Key {
	case termyx.KeyUp:
		return "↑"
	case termyx.KeyDown:
		return "↓"
	case termyx.KeyLeft:
		return "←"
	case termyx.KeyRight:
		return "→"
	case termyx.KeyEnter:
		return "Enter"
	case termyx.KeyBackspace:
		return "Backspace"
	case termyx.KeyDelete:
		return "Delete"
	case termyx.KeyEscape:
		return "Esc"
	case termyx.KeyTab:
		return "Tab"
	case termyx.KeyCtrlL:
		return "Ctrl+L"
	case termyx.KeyHome:
		return "Home"
	case termyx.KeyEnd:
		return "End"
	case termyx.KeyPageUp:
		return "PgUp"
	case termyx.KeyPageDown:
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

	app := &termyx.App{
		Root: func() *termyx.Node {
			return buildUI(s)
		},
		OnKey: func(ev termyx.KeyEvent) bool {
			if ev.Key == termyx.KeyCtrlC || ev.Key == termyx.KeyCtrlD || ev.Rune == 'q' {
				return true
			}
			s.mu.Lock()
			s.lastKey = keyName(ev)
			s.mu.Unlock()
			return false
		},
		Update: update,
	}

	if err := termyx.Run(app); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
