package bt

import "strings"

// Node is the core interface every behavior tree node implements.
type Node[T any] interface {
	Process(bb *Blackboard[T]) (State, error)
	String() string
}

// treeNode is an unexported interface used only for tree visualization.
// It avoids generic complexity in the rendering helpers.
type treeNode interface {
	treeLabel() string
	treeChildren() []treeNode
}

// treeLines writes the subtree rooted at n into buf with box-drawing connectors.
func treeLines(n treeNode, buf *strings.Builder, prefix string, isLast bool) {
	connector := "├── "
	childPrefix := prefix + "│   "
	if isLast {
		connector = "└── "
		childPrefix = prefix + "    "
	}
	buf.WriteString(prefix + connector + n.treeLabel())
	ch := n.treeChildren()
	for i, child := range ch {
		buf.WriteByte('\n')
		treeLines(child, buf, childPrefix, i == len(ch)-1)
	}
}
