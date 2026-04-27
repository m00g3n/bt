package bt

import "errors"

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
		return Failure, errors.New("bt: leaf node " + n.name + ": fn is nil")
	}
	return n.fn(bb)
}

func (n *leafNode[T]) String() string           { return n.name + " [Node]" }
func (n *leafNode[T]) treeLabel() string        { return n.name + " [Node]" }
func (n *leafNode[T]) treeChildren() []treeNode { return nil }
