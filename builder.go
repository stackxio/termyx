package termyx

import (
	"fmt"
	"sync/atomic"
)

var nodeSeq uint64

func nextID() string {
	return fmt.Sprintf("n%d", atomic.AddUint64(&nodeSeq, 1))
}

// Row creates a horizontal flex container.
func Row(children ...*Node) *Node {
	return &Node{
		ID:       nextID(),
		Type:     ContainerNode,
		Props:    Props{Direction: DirectionRow, FlexGrow: 1},
		Children: children,
	}
}

// Column creates a vertical flex container.
func Column(children ...*Node) *Node {
	return &Node{
		ID:       nextID(),
		Type:     ContainerNode,
		Props:    Props{Direction: DirectionColumn, FlexGrow: 1},
		Children: children,
	}
}

// Text creates a leaf node that renders a string.
func Text(s string) *Node {
	return &Node{
		ID:    nextID(),
		Type:  TextNode,
		Props: Props{Text: s, FlexGrow: 1},
	}
}

// StyledText creates a text node with explicit styling.
func StyledText(s string, style Style) *Node {
	n := Text(s)
	n.Props.Style = style
	return n
}

// Custom creates a node with a user-supplied render function.
// Children are laid out by the engine and rendered by the tree traversal.
func Custom(render RenderFunc, children ...*Node) *Node {
	return &Node{
		ID:       nextID(),
		Type:     CustomNode,
		Props:    Props{FlexGrow: 1},
		Children: children,
		Render:   render,
	}
}

// Grow overrides the flex grow factor on n and returns n.
func Grow(n *Node, factor float64) *Node {
	n.Props.FlexGrow = factor
	return n
}

// Fixed sets an explicit width and height on n and returns n.
func Fixed(n *Node, width, height int) *Node {
	n.Props.Width = width
	n.Props.Height = height
	return n
}

// WithKey attaches a reconciliation key to n and returns n.
func WithKey(n *Node, key string) *Node {
	n.Key = key
	return n
}

// WithStyle sets the style on n and returns n.
func WithStyle(n *Node, style Style) *Node {
	n.Props.Style = style
	return n
}

// Pad sets uniform padding on all four sides and returns n.
func Pad(n *Node, all int) *Node {
	n.Props.PaddingTop = all
	n.Props.PaddingRight = all
	n.Props.PaddingBottom = all
	n.Props.PaddingLeft = all
	return n
}

// PadXY sets horizontal (x) and vertical (y) padding and returns n.
func PadXY(n *Node, x, y int) *Node {
	n.Props.PaddingTop = y
	n.Props.PaddingBottom = y
	n.Props.PaddingLeft = x
	n.Props.PaddingRight = x
	return n
}

// PadSides sets each padding side individually and returns n.
func PadSides(n *Node, top, right, bottom, left int) *Node {
	n.Props.PaddingTop = top
	n.Props.PaddingRight = right
	n.Props.PaddingBottom = bottom
	n.Props.PaddingLeft = left
	return n
}

// MinSize sets minimum width and height constraints on n and returns n.
func MinSize(n *Node, minW, minH int) *Node {
	n.Props.MinWidth = minW
	n.Props.MinHeight = minH
	return n
}

// WithAlign sets the cross-axis alignment for a container's children and returns n.
func WithAlign(n *Node, a Align) *Node {
	n.Props.AlignItems = a
	return n
}

// FixedWidth sets only the width (height remains flex) and returns n.
func FixedWidth(n *Node, width int) *Node {
	n.Props.Width = width
	return n
}

// FixedHeight sets only the height (width remains flex) and returns n.
func FixedHeight(n *Node, height int) *Node {
	n.Props.Height = height
	return n
}
