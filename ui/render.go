package ui

import "strings"

// Render traverses the node tree and draws each node into buf.
// Custom nodes with a Render func are called first; children are always rendered after.
func Render(node *Node, buf *Buffer) {
	if node.Render != nil {
		node.Render(buf, node.Layout)
	} else {
		renderDefault(node, buf)
	}
	for _, child := range node.Children {
		Render(child, buf)
	}
}

func renderDefault(node *Node, buf *Buffer) {
	switch node.Type {
	case TextNode:
		renderText(node, buf)
	case ContainerNode:
		if node.Props.Style.BG.Set {
			l := node.Layout
			buf.Fill(l.X, l.Y, l.Width, l.Height, ' ', node.Props.Style)
		}
	}
}

func renderText(node *Node, buf *Buffer) {
	l := node.Layout
	if l.Width <= 0 || l.Height <= 0 {
		return
	}

	if node.Props.Style.BG.Set {
		buf.Fill(l.X, l.Y, l.Width, l.Height, ' ', node.Props.Style)
	}

	lines := wrapLines(node.Props.Text, l.Width)
	for i, line := range lines {
		if i >= l.Height {
			break
		}
		buf.WriteText(l.X, l.Y+i, line, node.Props.Style)
	}
}

// wrapLines splits text at newlines and wraps long lines to width.
func wrapLines(text string, width int) []string {
	if width <= 0 {
		return nil
	}
	var result []string
	for _, line := range strings.Split(text, "\n") {
		runes := []rune(line)
		if len(runes) == 0 {
			result = append(result, "")
			continue
		}
		for len(runes) > 0 {
			end := width
			if end > len(runes) {
				end = len(runes)
			}
			result = append(result, string(runes[:end]))
			runes = runes[end:]
		}
	}
	return result
}
