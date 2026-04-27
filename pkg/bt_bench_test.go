package bt_test

import (
	"testing"

	bt "github.com/m00g3n/bt/pkg"
)

// sink prevents the compiler from optimising away benchmark calls.
var sink bt.State

// emptyState is used for benchmarks that don't access blackboard data.
type emptyState struct{}

// benchState is used for the mixed-tree benchmark that reads visibility flags.
type benchState struct {
	Visible bool
	InRange bool
}

// counterState is used for the blackboard-access benchmark.
type counterState struct {
	Counter int
}

func makeFlatSequence(n int) bt.Node[emptyState] {
	children := make([]bt.Node[emptyState], n)
	for i := range children {
		children[i] = bt.NewLeaf("n", func(*bt.Blackboard[emptyState]) (bt.State, error) { return bt.Success, nil })
	}
	return bt.NewSequence("seq", children...)
}

// makeFlatSelector builds a selector where the first n-1 children fail and the last succeeds.
func makeFlatSelector(n int) bt.Node[emptyState] {
	children := make([]bt.Node[emptyState], n)
	for i := 0; i < n-1; i++ {
		children[i] = bt.NewLeaf("f", func(*bt.Blackboard[emptyState]) (bt.State, error) { return bt.Failure, nil })
	}
	children[n-1] = bt.NewLeaf("ok", func(*bt.Blackboard[emptyState]) (bt.State, error) { return bt.Success, nil })
	return bt.NewSelector("sel", children...)
}

func makeMixedTree() bt.Node[benchState] {
	visible := bt.NewLeaf("visible", func(bb *bt.Blackboard[benchState]) (bt.State, error) {
		if bb.Data.Visible {
			return bt.Success, nil
		}
		return bt.Failure, nil
	})
	inRange := bt.NewLeaf("in_range", func(bb *bt.Blackboard[benchState]) (bt.State, error) {
		if bb.Data.InRange {
			return bt.Success, nil
		}
		return bt.Failure, nil
	})
	fire := bt.NewLeaf("fire", func(*bt.Blackboard[benchState]) (bt.State, error) { return bt.Success, nil })
	flee := bt.NewLeaf("flee", func(*bt.Blackboard[benchState]) (bt.State, error) { return bt.Failure, nil })
	precond := bt.NewParallel("preconditions", 2, visible, inRange)
	attack := bt.NewSequence("attack", precond, fire)
	notFlee := bt.NewInverter("not_fleeing", flee)
	return bt.NewSelector("root", attack, notFlee)
}

func BenchmarkLeaf(b *testing.B) {
	bb := &bt.Blackboard[emptyState]{}
	node := bt.NewLeaf("x", func(*bt.Blackboard[emptyState]) (bt.State, error) { return bt.Success, nil })
	b.ResetTimer()
	for b.Loop() {
		sink, _ = node.Process(bb)
	}
}

func BenchmarkSequence5AllSuccess(b *testing.B) {
	bb := &bt.Blackboard[emptyState]{}
	node := makeFlatSequence(5)
	b.ResetTimer()
	for b.Loop() {
		sink, _ = node.Process(bb)
	}
}

func BenchmarkSequence20AllSuccess(b *testing.B) {
	bb := &bt.Blackboard[emptyState]{}
	node := makeFlatSequence(20)
	b.ResetTimer()
	for b.Loop() {
		sink, _ = node.Process(bb)
	}
}

func BenchmarkSelector5LastSuccess(b *testing.B) {
	bb := &bt.Blackboard[emptyState]{}
	node := makeFlatSelector(5)
	b.ResetTimer()
	for b.Loop() {
		sink, _ = node.Process(bb)
	}
}

func BenchmarkSelector20LastSuccess(b *testing.B) {
	bb := &bt.Blackboard[emptyState]{}
	node := makeFlatSelector(20)
	b.ResetTimer()
	for b.Loop() {
		sink, _ = node.Process(bb)
	}
}

func BenchmarkParallel5AllSuccess(b *testing.B) {
	bb := &bt.Blackboard[emptyState]{}
	node := bt.NewParallel("par", 5,
		bt.NewLeaf("a", func(*bt.Blackboard[emptyState]) (bt.State, error) { return bt.Success, nil }),
		bt.NewLeaf("b", func(*bt.Blackboard[emptyState]) (bt.State, error) { return bt.Success, nil }),
		bt.NewLeaf("c", func(*bt.Blackboard[emptyState]) (bt.State, error) { return bt.Success, nil }),
		bt.NewLeaf("d", func(*bt.Blackboard[emptyState]) (bt.State, error) { return bt.Success, nil }),
		bt.NewLeaf("e", func(*bt.Blackboard[emptyState]) (bt.State, error) { return bt.Success, nil }),
	)
	b.ResetTimer()
	for b.Loop() {
		sink, _ = node.Process(bb)
	}
}

func BenchmarkInverter(b *testing.B) {
	bb := &bt.Blackboard[emptyState]{}
	node := bt.NewInverter("inv",
		bt.NewLeaf("x", func(*bt.Blackboard[emptyState]) (bt.State, error) { return bt.Success, nil }),
	)
	b.ResetTimer()
	for b.Loop() {
		sink, _ = node.Process(bb)
	}
}

func BenchmarkMixedTree(b *testing.B) {
	bb := &bt.Blackboard[benchState]{Data: benchState{Visible: true, InRange: true}}
	node := makeMixedTree()
	b.ResetTimer()
	for b.Loop() {
		sink, _ = node.Process(bb)
	}
}

func BenchmarkBlackboardAccess(b *testing.B) {
	bb := &bt.Blackboard[counterState]{}
	node := bt.NewLeaf("rw", func(bb *bt.Blackboard[counterState]) (bt.State, error) {
		bb.Data.Counter++
		return bt.Success, nil
	})
	b.ResetTimer()
	for b.Loop() {
		sink, _ = node.Process(bb)
	}
}
