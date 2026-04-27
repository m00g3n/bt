package bt

import "strings"

type parallelNode[T any] struct {
	name      string
	threshold int
	children  []Node[T]
}

// NewParallel creates a Parallel node.
// Always ticks ALL children (no short-circuit).
// Returns Success if the number of successful children >= threshold, Failure otherwise.
func NewParallel[T any](name string, threshold int, children ...Node[T]) Node[T] {
	return &parallelNode[T]{name: name, threshold: threshold, children: children}
}

func (n *parallelNode[T]) Process(bb *Blackboard[T]) (State, error) {
	successes := 0
	var firstErr error
	for _, child := range n.children {
		s, err := child.Process(bb)
		if err != nil && firstErr == nil {
			firstErr = err
		}
		if s == Success {
			successes++
		}
	}
	if successes >= n.threshold {
		return Success, firstErr
	}
	return Failure, firstErr
}

func (n *parallelNode[T]) String() string {
	var buf strings.Builder
	buf.WriteString(n.name + " [Parallel]")
	ch := n.treeChildren()
	for i, child := range ch {
		buf.WriteByte('\n')
		treeLines(child, &buf, "", i == len(ch)-1)
	}
	return buf.String()
}

func (n *parallelNode[T]) treeLabel() string { return n.name + " [Parallel]" }
func (n *parallelNode[T]) treeChildren() []treeNode {
	out := make([]treeNode, len(n.children))
	for i, c := range n.children {
		out[i] = c.(treeNode)
	}
	return out
}
