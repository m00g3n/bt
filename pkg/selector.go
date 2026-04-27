package bt

import "strings"

type selectorNode[T any] struct {
	name     string
	children []Node[T]
}

// NewSelector creates a Selector node (OR logic).
// Ticks children left-to-right; returns on first non-Failure result.
// Returns Failure only when all children fail.
func NewSelector[T any](name string, children ...Node[T]) Node[T] {
	return &selectorNode[T]{name: name, children: children}
}

func (n *selectorNode[T]) Process(bb *Blackboard[T]) (State, error) {
	for _, child := range n.children {
		s, err := child.Process(bb)
		if err != nil {
			return s, err
		}
		if s != Failure {
			return s, nil
		}
	}
	return Failure, nil
}

func (n *selectorNode[T]) String() string {
	var buf strings.Builder
	buf.WriteString(n.name + " [Selector]")
	ch := n.treeChildren()
	for i, child := range ch {
		buf.WriteByte('\n')
		treeLines(child, &buf, "", i == len(ch)-1)
	}
	return buf.String()
}

func (n *selectorNode[T]) treeLabel() string { return n.name + " [Selector]" }
func (n *selectorNode[T]) treeChildren() []treeNode {
	out := make([]treeNode, len(n.children))
	for i, c := range n.children {
		out[i] = c.(treeNode)
	}
	return out
}
