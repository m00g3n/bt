package bt

import "strings"

type inverterNode[T any] struct {
	name  string
	child Node[T]
}

// NewInverter creates an Inverter decorator.
// Swaps Success and Failure.
func NewInverter[T any](name string, child Node[T]) Node[T] {
	return &inverterNode[T]{name: name, child: child}
}

func (n *inverterNode[T]) Process(bb *Blackboard[T]) (State, error) {
	s, err := n.child.Process(bb)
	if err != nil {
		return s, err
	}
	if s == Success {
		return Failure, nil
	}
	return Success, nil
}

func (n *inverterNode[T]) String() string {
	var buf strings.Builder
	buf.WriteString(n.name + " [Inverter]")
	buf.WriteByte('\n')
	treeLines(n.child.(treeNode), &buf, "", true)
	return buf.String()
}

func (n *inverterNode[T]) treeLabel() string { return n.name + " [Inverter]" }
func (n *inverterNode[T]) treeChildren() []treeNode {
	return []treeNode{n.child.(treeNode)}
}
