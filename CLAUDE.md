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
- `NewSequence(children...)` — AND logic; short-circuits on first `Failure`
- `NewSelector(children...)` — OR logic; short-circuits on first `Success`
- `NewParallel(threshold, children...)` — ticks all children; returns `Success` if ≥ threshold succeed
- `NewInverter(child)` — decorator swapping `Success` ↔ `Failure`

**Error propagation:** errors bubble up but do not short-circuit composite nodes (Parallel captures only the first error).

**Tree visualization:** call `.String()` on any node to get an ASCII tree using box-drawing characters.

