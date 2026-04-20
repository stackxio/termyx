package termyx

// collectFocusOrder does a depth-first walk of the tree and appends the IDs
// of all focusable nodes in document order.
func collectFocusOrder(node *Node, out *[]string) {
	if node.Props.Focusable {
		*out = append(*out, node.ID)
	}
	for _, child := range node.Children {
		collectFocusOrder(child, out)
	}
}

// applyFocus marks the node with the given ID as focused (and all others as
// not focused). Called after reconciliation, before rendering.
func applyFocus(node *Node, focusedID string) {
	node.Focused = node.ID == focusedID && focusedID != ""
	for _, child := range node.Children {
		applyFocus(child, focusedID)
	}
}

// findNode returns the node with the given ID, or nil.
func findNode(node *Node, id string) *Node {
	if node.ID == id {
		return node
	}
	for _, child := range node.Children {
		if found := findNode(child, id); found != nil {
			return found
		}
	}
	return nil
}
