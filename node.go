package termyx

import (
	"strconv"
	"strings"
)

// Direction defines the layout axis for a container node.
type Direction int

const (
	DirectionRow    Direction = iota
	DirectionColumn Direction = iota
)

// NodeType identifies the kind of node.
type NodeType int

const (
	ContainerNode NodeType = iota
	TextNode
	CustomNode
)

// Align controls cross-axis alignment of children inside a container.
type Align int

const (
	AlignStretch Align = iota // children fill the cross axis (default)
	AlignStart                // children align to the start of the cross axis
	AlignCenter               // children center on the cross axis
	AlignEnd                  // children align to the end of the cross axis
)

// Color holds an RGB terminal color.
type Color struct {
	R, G, B uint8
	Set      bool
}

// RGB constructs a Color from r, g, b components (0–255).
func RGB(r, g, b uint8) Color {
	return Color{R: r, G: g, B: b, Set: true}
}

// Hex parses a hex color string ("#RRGGBB" or "RRGGBB").
// Returns a zero Color if the string is invalid.
func Hex(s string) Color {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return Color{}
	}
	r, e1 := strconv.ParseUint(s[0:2], 16, 8)
	g, e2 := strconv.ParseUint(s[2:4], 16, 8)
	b, e3 := strconv.ParseUint(s[4:6], 16, 8)
	if e1 != nil || e2 != nil || e3 != nil {
		return Color{}
	}
	return Color{R: uint8(r), G: uint8(g), B: uint8(b), Set: true}
}

// Style defines the visual appearance of a cell.
type Style struct {
	FG           Color
	BG           Color
	Bold         bool
	Italic       bool
	Underline    bool
	Strikethrough bool
	Reverse      bool // swap FG and BG (cursor/selection highlight)
}

// Props holds all configuration for a Node.
type Props struct {
	Direction Direction
	FlexGrow  float64
	Width     int // 0 = flex
	Height    int // 0 = flex
	MinWidth  int
	MinHeight int

	// Padding shrinks the content area available to children.
	PaddingTop    int
	PaddingRight  int
	PaddingBottom int
	PaddingLeft   int

	// AlignItems controls cross-axis alignment of children.
	AlignItems Align

	Text      string
	Style     Style
	Focusable bool
	OnKey     func(KeyEvent)
}

// LayoutResult holds the computed position and size of a node after layout.
type LayoutResult struct {
	X, Y, Width, Height int
}

// RenderFunc is called during the render pass to draw a node into the buffer.
// Use buf.Region(layout) to get a Canvas with relative coordinates.
type RenderFunc func(buf *Buffer, layout LayoutResult)

// Node is the fundamental building block of a Termyx UI tree.
type Node struct {
	ID       string
	Key      string
	Type     NodeType
	Props    Props
	Children []*Node
	Layout   LayoutResult
	Focused  bool // set by the runtime's focus system
	Render   RenderFunc
}
