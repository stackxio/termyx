package termyx

import (
	"strings"
	"unicode/utf8"
)

// Truncate shortens s to at most width runes. If truncated, the last rune
// is replaced with "…" (ellipsis). Returns s unchanged if it fits.
func Truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	return string(runes[:width-1]) + "…"
}

// TruncateHard shortens s to exactly width runes with no ellipsis.
func TruncateHard(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	return string(runes[:width])
}

// PadRight pads s with spaces on the right to exactly width runes.
// If s is longer than width, it is truncated (hard, no ellipsis).
func PadRight(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		if len(runes) > width {
			return string(runes[:width])
		}
		return s
	}
	return s + strings.Repeat(" ", width-len(runes))
}

// PadLeft pads s with spaces on the left to exactly width runes.
func PadLeft(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		if len(runes) > width {
			return string(runes[:width])
		}
		return s
	}
	return strings.Repeat(" ", width-len(runes)) + s
}

// Center centers s within width columns, padding with spaces on both sides.
func Center(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		return string(runes[:width])
	}
	pad := width - len(runes)
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// WordWrap wraps text to at most width runes per line, breaking at word
// boundaries where possible. Existing newlines are preserved.
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
	if utf8.RuneCountInString(line) <= width {
		return []string{line}
	}
	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	current := ""
	for _, word := range words {
		wordRunes := []rune(word)
		// Word itself too long — hard break it.
		for len(wordRunes) > width {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
			lines = append(lines, string(wordRunes[:width]))
			wordRunes = wordRunes[width:]
		}
		w := string(wordRunes)
		if current == "" {
			current = w
		} else if utf8.RuneCountInString(current)+1+utf8.RuneCountInString(w) <= width {
			current += " " + w
		} else {
			lines = append(lines, current)
			current = w
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// RuneWidth returns the display width of s in terminal columns.
// For now this is equivalent to rune count; wide-char support can be added later.
func RuneWidth(s string) int {
	return utf8.RuneCountInString(s)
}
