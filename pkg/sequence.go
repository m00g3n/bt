package bt

import "strings"

type sequenceNode[T any] struct {
	name     string
	children []Node[T]
}

// NewSequence creates a Sequence node (AND logic).
// Ticks children left-to-right; returns on first non-Success result.
// Returns Success only when all children succeed.
func NewSequence[T any](name string, children ...Node[T]) Node[T] {
	return &sequenceNode[T]{name: name, children: children}
}

func (n *sequenceNode[T]) Process(bb *Blackboard[T]) (State, error) {
	for _, child := range n.children {
		s, err := child.Process(bb)
		if err != nil {
			return s, err
		}
		if s != Success {
			return s, nil
		}
	}
	return Success, nil
}

func (n *sequenceNode[T]) String() string {
	var buf strings.Builder
	buf.WriteString(n.name + " [Sequence]")
	ch := n.treeChildren()
	for i, child := range ch {
		buf.WriteByte('\n')
		treeLines(child, &buf, "", i == len(ch)-1)
	}
	return buf.String()
}

func (n *sequenceNode[T]) treeLabel() string { return n.name + " [Sequence]" }
func (n *sequenceNode[T]) treeChildren() []treeNode {
	out := make([]treeNode, len(n.children))
	for i, c := range n.children {
		out[i] = c.(treeNode)
	}
	return out
}
