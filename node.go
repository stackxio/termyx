package termyx

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

// Color holds an RGB terminal color.
type Color struct {
	R, G, B uint8
	Set      bool
}

// RGB returns a set Color with the given values.
func RGB(r, g, b uint8) Color {
	return Color{R: r, G: g, B: b, Set: true}
}

// Style defines visual appearance of a cell.
type Style struct {
	FG   Color
	BG   Color
	Bold bool
}

// Props holds all configuration for a Node.
type Props struct {
	Direction Direction
	FlexGrow  float64
	Width     int // 0 = flex
	Height    int // 0 = flex
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
type RenderFunc func(buf *Buffer, layout LayoutResult)

// Node is the fundamental building block of a Termyx UI tree.
type Node struct {
	ID       string
	Key      string
	Type     NodeType
	Props    Props
	Children []*Node
	Layout   LayoutResult
	Render   RenderFunc
}
