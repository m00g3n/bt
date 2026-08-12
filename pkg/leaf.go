package bt

import "fmt"

type leafNode[T any] struct {
	name string
	fn   func(*Blackboard[T]) (State, error)
}

// NewLeaf creates a leaf node that delegates to fn on each tick.
// A nil fn always returns Failure.
func NewLeaf[T any](name string, fn func(*Blackboard[T]) (State, error)) Node[T] {
	return &leafNode[T]{name: name, fn: fn}
}

func (n *leafNode[T]) Process(bb *Blackboard[T]) (State, error) {
	if n.fn == nil {
		return Failure, fmt.Errorf("bt: leaf node %q: fn is nil", n.name)
	}
	s, err := n.fn(bb)
	if err != nil {
		return s, fmt.Errorf("bt: leaf node %q: %w", n.name, err)
	}
	return s, nil
}

func (n *leafNode[T]) String() string           { return n.name + " [Node]" }
func (n *leafNode[T]) treeLabel() string        { return n.name + " [Node]" }
func (n *leafNode[T]) treeChildren() []treeNode { return nil }
