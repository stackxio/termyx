package termyx

import (
	"fmt"
	"strings"
	"time"
)

// ── Tree ─────────────────────────────────────────────────────────────────────

// TreeNode is one item in a Tree. Children are shown when Expanded is true.
type TreeNode struct {
	Label    string
	Style    Style
	Children []*TreeNode
	Expanded bool
	Data     any // application-defined payload
}

// TreeStyle holds visual styles for the Tree component.
type TreeStyle struct {
	Normal   Style
	Selected Style
	Expanded Style  // style for expanded parent nodes
	Indent   int    // spaces per depth level (default 2)
	Icon     string // prefix icon for leaf nodes (default "  ")
	OpenIcon string // icon for expanded nodes (default "▾ ")
	ClosedIcon string // icon for collapsed nodes (default "▸ ")
}

// TreeModel holds the state for a Tree component.
type TreeModel struct {
	Roots    []*TreeNode
	Selected *TreeNode
	Style    TreeStyle
}

// treeItem is a flattened row used during rendering.
type treeItem struct {
	node  *TreeNode
	depth int
}

func flattenTree(nodes []*TreeNode, depth int, out *[]treeItem) {
	for _, n := range nodes {
		*out = append(*out, treeItem{node: n, depth: depth})
		if n.Expanded && len(n.Children) > 0 {
			flattenTree(n.Children, depth+1, out)
		}
	}
}

// SelectNext moves the selection to the next visible row.
func (m *TreeModel) SelectNext() {
	rows := m.rows()
	for i, r := range rows {
		if r.node == m.Selected && i+1 < len(rows) {
			m.Selected = rows[i+1].node
			return
		}
	}
}

// SelectPrev moves the selection to the previous visible row.
func (m *TreeModel) SelectPrev() {
	rows := m.rows()
	for i, r := range rows {
		if r.node == m.Selected && i > 0 {
			m.Selected = rows[i-1].node
			return
		}
	}
}

// Toggle expands or collapses the selected node.
func (m *TreeModel) Toggle() {
	if m.Selected != nil {
		m.Selected.Expanded = !m.Selected.Expanded
	}
}

func (m *TreeModel) rows() []treeItem {
	var rows []treeItem
	flattenTree(m.Roots, 0, &rows)
	return rows
}

// Tree renders an expandable / collapsible tree view.
func Tree(model *TreeModel) *Node {
	indent := model.Style.Indent
	if indent <= 0 {
		indent = 2
	}
	icon := model.Style.Icon
	if icon == "" {
		icon = "  "
	}
	openIcon := model.Style.OpenIcon
	if openIcon == "" {
		openIcon = "▾ "
	}
	closedIcon := model.Style.ClosedIcon
	if closedIcon == "" {
		closedIcon = "▸ "
	}

	return Custom(func(buf *Buffer, l LayoutResult) {
		c := buf.Region(l)
		rows := model.rows()
		for y, r := range rows {
			if y >= l.Height {
				break
			}
			n := r.node
			prefix := strings.Repeat(" ", r.depth*indent)
			var bullet string
			if len(n.Children) > 0 {
				if n.Expanded {
					bullet = openIcon
				} else {
					bullet = closedIcon
				}
			} else {
				bullet = icon
			}
			text := prefix + bullet + n.Label
			style := n.Style
			if (style == Style{}) {
				style = model.Style.Normal
			}
			if n.Expanded && len(n.Children) > 0 {
				style = model.Style.Expanded
			}
			if n == model.Selected {
				style = model.Style.Selected
			}
			c.WriteText(0, y, PadRight(text, l.Width), style)
		}
	})
}

// ── Select (dropdown) ─────────────────────────────────────────────────────────

// SelectStyle holds visual styles for a Select dropdown.
type SelectStyle struct {
	Normal   Style
	Selected Style
	Open     Style // style of the trigger line when open
	Dropdown Style // style of dropdown items
}

// SelectModel holds the state for a Select (dropdown) component.
type SelectModel struct {
	Options  []string
	Value    int // index of the currently selected option
	Open     bool
	Style    SelectStyle
	Placeholder string // shown when Options is empty
}

// SelectNext moves the cursor to the next option (when open).
func (m *SelectModel) SelectNext() {
	if m.Value < len(m.Options)-1 {
		m.Value++
	}
}

// SelectPrev moves the cursor to the previous option (when open).
func (m *SelectModel) SelectPrev() {
	if m.Value > 0 {
		m.Value--
	}
}

// Toggle opens or closes the dropdown.
func (m *SelectModel) Toggle() { m.Open = !m.Open }

// Confirm closes the dropdown, keeping the current Value.
func (m *SelectModel) Confirm() { m.Open = false }

// Selected returns the currently selected option string.
func (m *SelectModel) Selected() string {
	if len(m.Options) == 0 {
		return m.Placeholder
	}
	return m.Options[m.Value]
}

// Select renders a single-line selector that expands into a dropdown list.
// The node's height should accommodate the trigger row plus len(Options) rows
// when open. Use Fixed or an outer container to reserve the space.
func Select(model *SelectModel) *Node {
	return Custom(func(buf *Buffer, l LayoutResult) {
		c := buf.Region(l)

		label := model.Selected()
		triggerStyle := model.Style.Normal
		if model.Open {
			triggerStyle = model.Style.Open
		}
		// Trigger row.
		arrow := " ▾"
		if model.Open {
			arrow = " ▴"
		}
		line := PadRight(label+arrow, l.Width)
		c.WriteText(0, 0, line, triggerStyle)

		if !model.Open || l.Height <= 1 {
			return
		}
		// Dropdown items.
		for i, opt := range model.Options {
			y := i + 1
			if y >= l.Height {
				break
			}
			style := model.Style.Dropdown
			if i == model.Value {
				style = model.Style.Selected
			}
			c.WriteText(0, y, PadRight(opt, l.Width), style)
		}
	})
}

// ── Toast / Notification ───────────────────────────────────────────────────────

// ToastLevel controls the visual severity of a Toast notification.
type ToastLevel int

const (
	ToastInfo    ToastLevel = iota
	ToastSuccess            // green
	ToastWarning            // yellow
	ToastError              // red
)

// ToastStyle holds visual styles for each Toast severity level.
type ToastStyle struct {
	Info    Style
	Success Style
	Warning Style
	Error   Style
	// Padding around the message (default 1).
	PadX int
}

// ToastModel holds the state for a Toast notification.
type ToastModel struct {
	Message string
	Level   ToastLevel
	Visible bool
	Style   ToastStyle
}

// Show makes the toast visible with the given message and level.
func (t *ToastModel) Show(msg string, level ToastLevel) {
	t.Message = msg
	t.Level = level
	t.Visible = true
}

// Hide hides the toast.
func (t *ToastModel) Hide() { t.Visible = false }

// ShowFor shows the toast and auto-hides it after d. notify triggers a
// re-render after each state change; close stop to cancel.
func (t *ToastModel) ShowFor(msg string, level ToastLevel, d time.Duration, notify chan<- struct{}, stop <-chan struct{}) {
	t.Show(msg, level)
	select {
	case notify <- struct{}{}:
	default:
	}
	go func() {
		select {
		case <-time.After(d):
		case <-stop:
			return
		}
		t.Hide()
		select {
		case notify <- struct{}{}:
		default:
		}
	}()
}

// Toast renders the notification as a single styled line. Typically used
// inside an Overlay to float it over the rest of the UI.
func Toast(model *ToastModel) *Node {
	return Custom(func(buf *Buffer, l LayoutResult) {
		if !model.Visible {
			return
		}
		c := buf.Region(l)
		style := model.Style.Info
		switch model.Level {
		case ToastSuccess:
			style = model.Style.Success
		case ToastWarning:
			style = model.Style.Warning
		case ToastError:
			style = model.Style.Error
		}
		padX := model.Style.PadX
		if padX <= 0 {
			padX = 1
		}
		prefix := strings.Repeat(" ", padX)
		msg := prefix + model.Message + prefix
		c.Fill(0, 0, l.Width, l.Height, ' ', style)
		c.WriteText(0, 0, TruncateHard(msg, l.Width), style)
	})
}

// ── Checkbox ──────────────────────────────────────────────────────────────────

// CheckboxStyle holds visual styles for a Checkbox.
type CheckboxStyle struct {
	Normal   Style
	Focused  Style
	Checked  string // rune string for checked state (default "■")
	Unchecked string // rune string for unchecked state (default "□")
}

// CheckboxModel holds the state for a Checkbox toggle.
type CheckboxModel struct {
	Checked bool
	Label   string
	Style   CheckboxStyle
	Focused bool
}

// Toggle flips the checked state.
func (m *CheckboxModel) Toggle() { m.Checked = !m.Checked }

// HandleKey processes a KeyEvent; Space/Enter toggle the checkbox.
// Returns true if the event was consumed.
func (m *CheckboxModel) HandleKey(ev KeyEvent) bool {
	if ev.Key == KeyEnter || ev.Rune == ' ' {
		m.Toggle()
		return true
	}
	return false
}

// Checkbox renders a single checkbox with an optional label.
func Checkbox(model *CheckboxModel) *Node {
	return Custom(func(buf *Buffer, l LayoutResult) {
		c := buf.Region(l)
		checked := model.Style.Checked
		if checked == "" {
			checked = "■"
		}
		unchecked := model.Style.Unchecked
		if unchecked == "" {
			unchecked = "□"
		}
		box := unchecked
		if model.Checked {
			box = checked
		}
		style := model.Style.Normal
		if model.Focused {
			style = model.Style.Focused
		}
		text := fmt.Sprintf("%s %s", box, model.Label)
		c.WriteText(0, 0, TruncateHard(text, l.Width), style)
	})
}
