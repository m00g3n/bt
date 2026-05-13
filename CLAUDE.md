# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Test
go test ./pkg
go test -v ./pkg
go test ./pkg -run TestName

# Benchmark
go test -bench . ./pkg
go test -bench BenchmarkName ./pkg
```

## Architecture

This is a specialized Behavior Tree library (`github.com/m00g3n/bt`). All code lives in `pkg/`.

**Core types:**
- `Node[T any]` interface — every node implements `Process(bb *Blackboard[T]) (State, error)`
- `Blackboard[T]` — generic shared-state container passed by pointer to all nodes during execution
- `State` — `Success` (0) or `Failure` (1)

**Node types:**
- `NewLeaf(name, fn)` — terminal node executing a `func(*Blackboard[T]) (State, error)`
- `NewSequence(name, children...)` — AND logic; short-circuits on first `Failure` or error
- `NewSelector(name, children...)` — OR logic; short-circuits on first `Success` or error
- `NewParallel(name, threshold, children...)` — ticks all children (no short-circuit); returns `Success` if ≥ threshold succeed
- `NewInverter(name, child)` — decorator swapping `Success` ↔ `Failure`

**Error propagation:** Sequence and Selector short-circuit on the first error. Parallel always ticks all children and captures only the first error.

**Tree visualization:** call `.String()` on any node to get an ASCII tree using box-drawing characters.

## Plans

Implementation plans live in `docs/plans/`. Use this directory when creating or reading plan files.

