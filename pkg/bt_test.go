package bt_test

import (
	"strings"
	"testing"

	bt "github.com/m00g3n/bt/pkg"
)

// noState is used for tests that don't read or write blackboard data.
type noState struct{}

// stateLeaf returns a leaf node that always returns state s.
func stateLeaf(s bt.State) bt.Node[noState] {
	return bt.NewLeaf("", func(*bt.Blackboard[noState]) (bt.State, error) { return s, nil })
}

// countingLeaf returns a leaf node that increments *n on each tick and returns s.
func countingLeaf(n *int, s bt.State) bt.Node[noState] {
	return bt.NewLeaf("", func(*bt.Blackboard[noState]) (bt.State, error) {
		*n++
		return s, nil
	})
}

// ─────────────────────────────────────────
// BT.State
// ─────────────────────────────────────────

func TestStateValues(t *testing.T) {
	if bt.Success == bt.Failure {
		t.Error("Success and Failure must be distinct")
	}

	if bt.Success != 0 {
		t.Errorf("Success must be 0, got %d", bt.Success)
	}
	if bt.Failure != 1 {
		t.Errorf("Failure must be 1, got %d", bt.Failure)
	}
}

// ─────────────────────────────────────────
// Leaf
// ─────────────────────────────────────────

func TestLeaf(t *testing.T) {
	bb := &bt.Blackboard[noState]{}

	for _, tc := range []struct {
		name string
		s    bt.State
	}{
		{"success", bt.Success},
		{"failure", bt.Failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			node := bt.NewLeaf(tc.name, func(*bt.Blackboard[noState]) (bt.State, error) { return tc.s, nil })
			got, err := node.Process(bb)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.s {
				t.Errorf("got %v, want %v", got, tc.s)
			}
		})
	}

	t.Run("nil_fn_returns_failure", func(t *testing.T) {
		node := bt.NewLeaf[noState]("nil", nil)
		got, err := node.Process(bb)
		if got != bt.Failure {
			t.Errorf("nil fn: got %v, want Failure", got)
		}
		if err == nil {
			t.Error("nil fn: want non-nil error")
		}
	})
}

// ─────────────────────────────────────────
// Sequence
// ─────────────────────────────────────────

func TestSequence(t *testing.T) {
	bb := &bt.Blackboard[noState]{}
	S, F := bt.Success, bt.Failure

	cases := []struct {
		name     string
		children []bt.State
		want     bt.State
	}{
		{"all_success", []bt.State{S, S, S}, bt.Success},
		{"fail_first", []bt.State{F, S, S}, bt.Failure},
		{"fail_middle", []bt.State{S, F, S}, bt.Failure},
		{"fail_last", []bt.State{S, S, F}, bt.Failure},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nodes := make([]bt.Node[noState], len(tc.children))
			for i, s := range tc.children {
				nodes[i] = stateLeaf(s)
			}
			seq := bt.NewSequence("seq", nodes...)
			got, err := seq.Process(bb)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSequenceShortCircuitOnFailure(t *testing.T) {
	bb := &bt.Blackboard[noState]{}
	firstCalls, lateCalls := 0, 0
	first := countingLeaf(&firstCalls, bt.Failure)
	late := countingLeaf(&lateCalls, bt.Success)
	seq := bt.NewSequence("seq", first, late)
	if _, err := seq.Process(bb); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if firstCalls != 1 {
		t.Errorf("first child: want 1 call, got %d", firstCalls)
	}
	if lateCalls != 0 {
		t.Errorf("later child must not be called after failure, got %d calls", lateCalls)
	}
}

func TestSequenceNesting(t *testing.T) {
	bb := &bt.Blackboard[noState]{}
	inner := bt.NewSequence("inner",
		stateLeaf(bt.Success),
		stateLeaf(bt.Success),
	)
	outer := bt.NewSequence("outer", inner, stateLeaf(bt.Success))
	got, err := outer.Process(bb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != bt.Success {
		t.Errorf("nested sequence: got %v, want Success", got)
	}
}

// ─────────────────────────────────────────
// Selector
// ─────────────────────────────────────────

func TestSelector(t *testing.T) {
	bb := &bt.Blackboard[noState]{}
	S, F := bt.Success, bt.Failure

	cases := []struct {
		name     string
		children []bt.State
		want     bt.State
	}{
		{"all_fail", []bt.State{F, F, F}, bt.Failure},
		{"success_first", []bt.State{S, F, F}, bt.Success},
		{"success_middle", []bt.State{F, S, F}, bt.Success},
		{"success_last", []bt.State{F, F, S}, bt.Success},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nodes := make([]bt.Node[noState], len(tc.children))
			for i, s := range tc.children {
				nodes[i] = stateLeaf(s)
			}
			sel := bt.NewSelector("sel", nodes...)
			got, err := sel.Process(bb)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSelectorShortCircuitOnSuccess(t *testing.T) {
	bb := &bt.Blackboard[noState]{}
	firstCalls, lateCalls := 0, 0
	first := countingLeaf(&firstCalls, bt.Success)
	late := countingLeaf(&lateCalls, bt.Failure)
	sel := bt.NewSelector("sel", first, late)
	if _, err := sel.Process(bb); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if firstCalls != 1 {
		t.Errorf("first child: want 1 call, got %d", firstCalls)
	}
	if lateCalls != 0 {
		t.Errorf("later child must not be called after success, got %d calls", lateCalls)
	}
}

func TestSelectorNesting(t *testing.T) {
	bb := &bt.Blackboard[noState]{}
	inner := bt.NewSelector("inner",
		stateLeaf(bt.Failure),
		stateLeaf(bt.Success),
	)
	outer := bt.NewSelector("outer", stateLeaf(bt.Failure), inner)
	got, err := outer.Process(bb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != bt.Success {
		t.Errorf("nested selector: got %v, want Success", got)
	}
}

// ─────────────────────────────────────────
// Parallel
// ─────────────────────────────────────────

func TestParallel(t *testing.T) {
	bb := &bt.Blackboard[noState]{}
	S, F := bt.Success, bt.Failure

	cases := []struct {
		name      string
		threshold int
		children  []bt.State
		want      bt.State
	}{
		{"all_success_default", 3, []bt.State{S, S, S}, bt.Success},
		{"one_fail_default", 3, []bt.State{S, F, S}, bt.Failure},
		{"threshold_1_any_success", 1, []bt.State{F, S, F}, bt.Success},
		{"threshold_1_all_fail", 1, []bt.State{F, F}, bt.Failure},
		{"custom_threshold_2_of_3", 2, []bt.State{S, S, F}, bt.Success},
		{"custom_threshold_2_fail", 2, []bt.State{S, F, F}, bt.Failure},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nodes := make([]bt.Node[noState], len(tc.children))
			for i, s := range tc.children {
				nodes[i] = stateLeaf(s)
			}
			par := bt.NewParallel("par", tc.threshold, nodes...)
			got, err := par.Process(bb)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParallelTicksAllChildren(t *testing.T) {
	bb := &bt.Blackboard[noState]{}
	counts := [3]int{}
	nodes := [3]bt.Node[noState]{
		countingLeaf(&counts[0], bt.Failure),
		countingLeaf(&counts[1], bt.Failure),
		countingLeaf(&counts[2], bt.Failure),
	}
	bt.NewParallel("par", 3, nodes[0], nodes[1], nodes[2]).Process(bb)
	for i, c := range counts {
		if c != 1 {
			t.Errorf("child %d: want 1 tick, got %d", i, c)
		}
	}
}

// ─────────────────────────────────────────
// Inverter
// ─────────────────────────────────────────

func TestInverter(t *testing.T) {
	bb := &bt.Blackboard[noState]{}

	cases := []struct {
		input, want bt.State
	}{
		{bt.Success, bt.Failure},
		{bt.Failure, bt.Success},
	}

	for _, tc := range cases {
		t.Run(tc.input.String(), func(t *testing.T) {
			inv := bt.NewInverter("inv", stateLeaf(tc.input))
			got, err := inv.Process(bb)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestInverterDoubleInversion(t *testing.T) {
	bb := &bt.Blackboard[noState]{}
	double := bt.NewInverter("outer",
		bt.NewInverter("inner", stateLeaf(bt.Success)),
	)
	got, err := double.Process(bb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != bt.Success {
		t.Errorf("double inversion: got %v, want Success", got)
	}
}

func TestInverterWrapsComposite(t *testing.T) {
	bb := &bt.Blackboard[noState]{}
	seq := bt.NewSequence("seq",
		stateLeaf(bt.Success),
		stateLeaf(bt.Failure),
	)
	inv := bt.NewInverter("inv", seq)
	got, err := inv.Process(bb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != bt.Success {
		t.Errorf("inverter(failing sequence): got %v, want Success", got)
	}
}

// ─────────────────────────────────────────
// Blackboard
// ─────────────────────────────────────────

type flagState struct {
	Flag bool
}

func TestBlackboard(t *testing.T) {
	writer := bt.NewLeaf("write", func(bb *bt.Blackboard[flagState]) (bt.State, error) {
		bb.Data.Flag = true
		return bt.Success, nil
	})
	reader := bt.NewLeaf("read", func(bb *bt.Blackboard[flagState]) (bt.State, error) {
		if bb.Data.Flag {
			return bt.Success, nil
		}
		return bt.Failure, nil
	})
	seq := bt.NewSequence("seq", writer, reader)
	bb := &bt.Blackboard[flagState]{}
	got, err := seq.Process(bb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != bt.Success {
		t.Errorf("blackboard r/w sequence: got %v, want Success", got)
	}
	if !bb.Data.Flag {
		t.Error("bb.Data.Flag should be true after write node")
	}
}

// ─────────────────────────────────────────
// String / tree visualization
// ─────────────────────────────────────────

func TestStringLeaf(t *testing.T) {
	node := bt.NewLeaf("patrol", func(*bt.Blackboard[noState]) (bt.State, error) { return bt.Success, nil })
	if got := node.String(); got != "patrol [Node]" {
		t.Errorf("leaf String(): got %q, want %q", got, "patrol [Node]")
	}
}

func TestStringSingleChild(t *testing.T) {
	child := bt.NewLeaf("child", func(*bt.Blackboard[noState]) (bt.State, error) { return bt.Success, nil })
	root := bt.NewSequence("root", child)
	expected := "root [Sequence]\n└── child [Node]"
	if got := root.String(); got != expected {
		t.Errorf("single-child:\ngot:  %q\nwant: %q", got, expected)
	}
}

func TestStringMultipleChildren(t *testing.T) {
	a := bt.NewLeaf("a", func(*bt.Blackboard[noState]) (bt.State, error) { return bt.Success, nil })
	b := bt.NewLeaf("b", func(*bt.Blackboard[noState]) (bt.State, error) { return bt.Success, nil })
	c := bt.NewLeaf("c", func(*bt.Blackboard[noState]) (bt.State, error) { return bt.Success, nil })
	root := bt.NewSelector("root", a, b, c)
	got := root.String()
	for _, want := range []string{"├── a [Node]", "├── b [Node]", "└── c [Node]"} {
		if !strings.Contains(got, want) {
			t.Errorf("String() missing %q\ngot:\n%s", want, got)
		}
	}
}

func TestStringNestedContinuationPrefix(t *testing.T) {
	leaf := bt.NewLeaf("leaf", func(*bt.Blackboard[noState]) (bt.State, error) { return bt.Success, nil })
	mid := bt.NewSequence("mid", leaf)
	last := bt.NewLeaf("last", func(*bt.Blackboard[noState]) (bt.State, error) { return bt.Success, nil })
	root := bt.NewSelector("root", mid, last)
	if !strings.Contains(root.String(), "│   └── leaf [Node]") {
		t.Errorf("continuation prefix missing\ngot:\n%s", root.String())
	}
}

func TestStringFullTree(t *testing.T) {
	replicasReady := bt.NewLeaf("replicas_ready", func(*bt.Blackboard[noState]) (bt.State, error) { return bt.Success, nil })
	notDegraded := bt.NewLeaf("not_degraded", func(*bt.Blackboard[noState]) (bt.State, error) { return bt.Success, nil })
	rollout := bt.NewLeaf("rollout", func(*bt.Blackboard[noState]) (bt.State, error) { return bt.Success, nil })
	rollback := bt.NewLeaf("rollback", func(*bt.Blackboard[noState]) (bt.State, error) { return bt.Failure, nil })
	precond := bt.NewParallel("preconditions", 2, replicasReady, notDegraded)
	deploy := bt.NewSequence("deploy", precond, rollout)
	noRollback := bt.NewInverter("no_rollback", rollback)
	root := bt.NewSelector("root", deploy, noRollback)

	expected := strings.Join([]string{
		"root [Selector]",
		"├── deploy [Sequence]",
		"│   ├── preconditions [Parallel]",
		"│   │   ├── replicas_ready [Node]",
		"│   │   └── not_degraded [Node]",
		"│   └── rollout [Node]",
		"└── no_rollback [Inverter]",
		"    └── rollback [Node]",
	}, "\n")

	if got := root.String(); got != expected {
		t.Errorf("full tree:\ngot:\n%s\nwant:\n%s", got, expected)
	}
}

func TestStringAllNodeTypesAsRoot(t *testing.T) {
	leaf := bt.NewLeaf("x", func(*bt.Blackboard[noState]) (bt.State, error) { return bt.Success, nil })
	cases := []struct {
		node bt.Node[noState]
		want string
	}{
		{bt.NewSequence("s", leaf), "s [Sequence]"},
		{bt.NewSelector("s", leaf), "s [Selector]"},
		{bt.NewParallel("s", 1, leaf), "s [Parallel]"},
		{bt.NewInverter("s", leaf), "s [Inverter]"},
	}
	for _, tc := range cases {
		if !strings.Contains(tc.node.String(), tc.want) {
			t.Errorf("String() missing %q\ngot:\n%s", tc.want, tc.node.String())
		}
	}
}

// ─────────────────────────────────────────
// Integration
// ─────────────────────────────────────────

func TestIntegrationGuardedPatrol(t *testing.T) {
	bb := &bt.Blackboard[noState]{}
	replicasReady := stateLeaf(bt.Success)
	rollout := stateLeaf(bt.Success)
	guarded := bt.NewSequence("guarded_deploy", replicasReady, rollout)
	got, err := guarded.Process(bb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != bt.Success {
		t.Errorf("guarded deploy (replicas ready): got %v, want Success", got)
	}

	notReady := stateLeaf(bt.Failure)
	skipped := bt.NewSequence("skipped_deploy", notReady, rollout)
	got, err = skipped.Process(bb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != bt.Failure {
		t.Errorf("guarded deploy (replicas not ready): got %v, want Failure", got)
	}
}

func TestIntegrationAttackFleeFallback(t *testing.T) {
	bb := &bt.Blackboard[noState]{}
	deployFails := stateLeaf(bt.Failure)
	rollbackSucceeds := stateLeaf(bt.Success)
	chain := bt.NewSelector("reconcile", deployFails, rollbackSucceeds)
	got, err := chain.Process(bb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != bt.Success {
		t.Errorf("deploy/rollback fallback: got %v, want Success", got)
	}

	// When deploy succeeds, rollback must NOT be ticked.
	deployCalls, rollbackCalls := 0, 0
	deployOK := countingLeaf(&deployCalls, bt.Success)
	rollbackSpy := countingLeaf(&rollbackCalls, bt.Success)
	bt.NewSelector("reconcile2", deployOK, rollbackSpy).Process(bb)
	if deployCalls != 1 {
		t.Errorf("deploy: want 1 call, got %d", deployCalls)
	}
	if rollbackCalls != 0 {
		t.Errorf("rollback must not be called when deploy succeeds, got %d calls", rollbackCalls)
	}
}

func TestIntegrationParallelPreconditions(t *testing.T) {
	bb := &bt.Blackboard[noState]{}
	replicasReady := stateLeaf(bt.Success)
	notDegraded := stateLeaf(bt.Success)
	precond := bt.NewParallel("preconditions", 2, replicasReady, notDegraded)
	got, err := precond.Process(bb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != bt.Success {
		t.Errorf("both preconditions met: got %v, want Success", got)
	}

	degraded := stateLeaf(bt.Failure)
	precondFail := bt.NewParallel("preconditions_fail", 2, replicasReady, degraded)
	got, err = precondFail.Process(bb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != bt.Failure {
		t.Errorf("one precondition failed: got %v, want Failure", got)
	}
}

// k8sState is the typed blackboard data for the complex integration tests.
type k8sState struct {
	ReplicasReady    bool
	NotDegraded      bool
	RolloutTriggered bool
	RollbackBlocked  bool
}

func makeComplexTree() bt.Node[k8sState] {
	replicasReady := bt.NewLeaf("replicas_ready", func(bb *bt.Blackboard[k8sState]) (bt.State, error) {
		if bb.Data.ReplicasReady {
			return bt.Success, nil
		}
		return bt.Failure, nil
	})
	notDegraded := bt.NewLeaf("not_degraded", func(bb *bt.Blackboard[k8sState]) (bt.State, error) {
		if bb.Data.NotDegraded {
			return bt.Success, nil
		}
		return bt.Failure, nil
	})
	rollout := bt.NewLeaf("rollout", func(bb *bt.Blackboard[k8sState]) (bt.State, error) {
		bb.Data.RolloutTriggered = true
		return bt.Success, nil
	})
	rollback := bt.NewLeaf("rollback", func(bb *bt.Blackboard[k8sState]) (bt.State, error) {
		if bb.Data.RollbackBlocked {
			return bt.Success, nil
		}
		return bt.Failure, nil
	})
	precond := bt.NewParallel("preconditions", 2, replicasReady, notDegraded)
	deploy := bt.NewSequence("deploy", precond, rollout)
	noRollback := bt.NewInverter("no_rollback", rollback)
	return bt.NewSelector("root", deploy, noRollback)
}

func TestIntegrationComplexTreeScenario1(t *testing.T) {
	root := makeComplexTree()
	bb := &bt.Blackboard[k8sState]{Data: k8sState{ReplicasReady: true, NotDegraded: true}}
	got, err := root.Process(bb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != bt.Success {
		t.Errorf("scenario1: got %v, want Success", got)
	}
	if !bb.Data.RolloutTriggered {
		t.Error("scenario1: bb.Data.RolloutTriggered should be true")
	}
}

func TestIntegrationComplexTreeScenario2(t *testing.T) {
	root := makeComplexTree()
	bb := &bt.Blackboard[k8sState]{Data: k8sState{ReplicasReady: false, NotDegraded: true}}
	// preconditions fail → rollback fallback (rollback fails → inverter succeeds)
	got, err := root.Process(bb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != bt.Success {
		t.Errorf("scenario2: got %v, want Success", got)
	}
	if bb.Data.RolloutTriggered {
		t.Error("scenario2: rollout should not have been triggered")
	}
}
