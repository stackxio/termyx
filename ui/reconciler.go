package ui

// Reconcile merges next onto prev, carrying stable identity (ID) from matched
// nodes so that future state fields survive across frames.
// Matching uses Key when set, otherwise falls back to position.
func Reconcile(prev, next *Node) *Node {
	if prev == nil || next == nil {
		return next
	}
	if prev.Type != next.Type || prev.Key != next.Key {
		return next
	}

	next.ID = prev.ID

	reconcileChildren(prev.Children, next.Children)
	return next
}

func reconcileChildren(prev, next []*Node) {
	byKey := make(map[string]*Node, len(prev))
	for _, c := range prev {
		if c.Key != "" {
			byKey[c.Key] = c
		}
	}

	for i, n := range next {
		var p *Node
		if n.Key != "" {
			p = byKey[n.Key]
		} else if i < len(prev) {
			p = prev[i]
		}
		if p != nil {
			Reconcile(p, n)
		}
	}
}
