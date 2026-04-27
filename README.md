# bt

A lightweight, generic Behavior Tree library for Go.

## Install

```bash
go get github.com/m00g3n/bt
```

Requires Go 1.21+ (uses generics).

## Concepts

A behavior tree is composed of **nodes** that each return `Success` or `Failure` when ticked. Shared state is passed through a typed **Blackboard**.

### Node types

| Constructor | Type | Logic |
|---|---|---|
| `NewLeaf(name, fn)` | Leaf | Executes a function; terminal node |
| `NewSequence(name, children...)` | Composite | AND — short-circuits on first `Failure` |
| `NewSelector(name, children...)` | Composite | OR — short-circuits on first `Success` |
| `NewParallel(name, threshold, children...)` | Composite | Ticks all children; `Success` if ≥ threshold succeed |
| `NewInverter(name, child)` | Decorator | Swaps `Success` ↔ `Failure` |

## Usage

```go
package main

import (
    "fmt"
    bt "github.com/m00g3n/bt/pkg"
)

type K8sState struct {
    ReplicasReady    bool
    NotDegraded      bool
    RolloutTriggered bool
}

func main() {
    replicasReady := bt.NewLeaf("replicas_ready", func(bb *bt.Blackboard[K8sState]) (bt.State, error) {
        if bb.Data.ReplicasReady {
            return bt.Success, nil
        }
        return bt.Failure, nil
    })

    notDegraded := bt.NewLeaf("not_degraded", func(bb *bt.Blackboard[K8sState]) (bt.State, error) {
        if bb.Data.NotDegraded {
            return bt.Success, nil
        }
        return bt.Failure, nil
    })

    rollout := bt.NewLeaf("rollout", func(bb *bt.Blackboard[K8sState]) (bt.State, error) {
        bb.Data.RolloutTriggered = true
        return bt.Success, nil
    })

    rollback := bt.NewLeaf("rollback", func(bb *bt.Blackboard[K8sState]) (bt.State, error) {
        return bt.Failure, nil
    })

    precond    := bt.NewParallel("preconditions", 2, replicasReady, notDegraded)
    deploy     := bt.NewSequence("deploy", precond, rollout)
    noRollback := bt.NewInverter("no_rollback", rollback)
    root       := bt.NewSelector("root", deploy, noRollback)

    bb := &bt.Blackboard[K8sState]{Data: K8sState{ReplicasReady: true, NotDegraded: true}}
    state, err := root.Process(bb)
    fmt.Println(state, err, bb.Data.RolloutTriggered) // Success <nil> true

    // Print the tree structure
    fmt.Println(root)
}
```

### Tree visualization

Every node implements `String()`, producing an ASCII tree:

```
root [Selector]
├── deploy [Sequence]
│   ├── preconditions [Parallel]
│   │   ├── replicas_ready [Node]
│   │   └── not_degraded [Node]
│   └── rollout [Node]
└── no_rollback [Inverter]
    └── rollback [Node]
```

### Blackboard

`Blackboard[T]` is a generic container for your state type. Pass it by pointer so leaf functions can mutate `Data` in place:

```go
type K8sState struct {
    ReplicasReady bool
    Namespace     string
}

bb := &bt.Blackboard[K8sState]{Data: K8sState{ReplicasReady: true}}
root.Process(bb)
```

### Error handling

Errors returned by leaf functions bubble up through the tree. Composite nodes (`Sequence`, `Selector`) stop and propagate the error immediately. `Parallel` ticks all children regardless and returns the first error encountered.

## Testing

```bash
go test ./pkg
go test -v ./pkg
go test -bench . ./pkg
```

## License

MIT
