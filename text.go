package termyx

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// RuneWidth returns the display width of s in terminal columns.
// Narrow (ASCII) characters count as 1; wide (CJK, emoji) count as 2.
func RuneWidth(s string) int {
	return runewidth.StringWidth(s)
}

// Truncate shortens s to fit within width display columns. If truncated,
// the last visible column is replaced with "…". Returns s unchanged if it fits.
func Truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	// Build up to width-1 columns, then append "…".
	target := width - 1 // runewidth of "…" is 1
	var b strings.Builder
	used := 0
	for _, r := range s {
		w := runewidth.RuneWidth(r)
		if used+w > target {
			break
		}
		b.WriteRune(r)
		used += w
	}
	b.WriteRune('…')
	return b.String()
}

// TruncateHard shortens s to fit within width display columns with no ellipsis.
// A wide rune that would overflow the boundary is dropped entirely (leaving a
// trailing space is the caller's responsibility if alignment matters).
func TruncateHard(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= width {
		return s
	}
	var b strings.Builder
	used := 0
	for _, r := range s {
		w := runewidth.RuneWidth(r)
		if used+w > width {
			break
		}
		b.WriteRune(r)
		used += w
	}
	return b.String()
}

// PadRight pads s with spaces on the right to exactly width display columns.
// If s is wider than width it is hard-truncated.
func PadRight(s string, width int) string {
	sw := runewidth.StringWidth(s)
	if sw > width {
		return TruncateHard(s, width)
	}
	return s + strings.Repeat(" ", width-sw)
}

// PadLeft pads s with spaces on the left to exactly width display columns.
func PadLeft(s string, width int) string {
	sw := runewidth.StringWidth(s)
	if sw > width {
		return TruncateHard(s, width)
	}
	return strings.Repeat(" ", width-sw) + s
}

// Center centers s within width display columns, padding with spaces on both sides.
func Center(s string, width int) string {
	sw := runewidth.StringWidth(s)
	if sw >= width {
		return TruncateHard(s, width)
	}
	pad := width - sw
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// WordWrap wraps text to at most width display columns per line, breaking at
// word boundaries where possible. Existing newlines are preserved.
func WordWrap(text string, width int) []string {
	if width <= 0 {
		return nil
	}
	var result []string
	for _, line := range strings.Split(text, "\n") {
		result = append(result, wrapLine(line, width)...)
	}
	return result
}

func wrapLine(line string, width int) []string {
	if runewidth.StringWidth(line) <= width {
		return []string{line}
	}
	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	current := ""
	currentW := 0
	for _, word := range words {
		ww := runewidth.StringWidth(word)
		// Word itself too long — hard break it.
		for ww > width {
			remaining := width - currentW
			if currentW > 0 && remaining < ww {
				lines = append(lines, current)
				current = ""
				currentW = 0
				remaining = width
			}
			// Take as many display columns from word as fit.
			chunk, rest := splitDisplayWidth(word, remaining)
			lines = append(lines, chunk)
			word = rest
			ww = runewidth.StringWidth(word)
			current = ""
			currentW = 0
		}
		if ww == 0 {
			continue
		}
		if currentW == 0 {
			current = word
			currentW = ww
		} else if currentW+1+ww <= width {
			current += " " + word
			currentW += 1 + ww
		} else {
			lines = append(lines, current)
			current = word
			currentW = ww
		}
	}
	if currentW > 0 {
		lines = append(lines, current)
	}
	return lines
}

// splitDisplayWidth splits s at exactly cols display columns.
// Returns the prefix that fits and the remainder.
func splitDisplayWidth(s string, cols int) (string, string) {
	used := 0
	for i, r := range s {
		w := runewidth.RuneWidth(r)
		if used+w > cols {
			return s[:i], s[i:]
		}
		used += w
	}
	return s, ""
}
