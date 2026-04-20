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
