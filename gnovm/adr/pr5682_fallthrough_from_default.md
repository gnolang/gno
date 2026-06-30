# PR5682: Allow fallthrough from `default:` when not textually last

## Context

In Go, `fallthrough` is permitted as the last statement of any clause that is
not the textually-last clause of an expression switch — including a `default:`
clause that appears earlier in the source. Gno rejected this with
`cannot fallthrough final case in switch`, even when `default:` was textually
first:

```go
switch x {
default:
    println("d"); fallthrough // gno rejected this
case 1:
    println("1")
}
```

The root cause was in `gnovm/pkg/gnolang/go2gno.go:toClauses`, which
unconditionally moved the `default` clause to the end of `SwitchStmt.Clauses`.
This made the runtime simpler (default ran when iteration reached the last
index) but corrupted textual ordering, which `fallthrough` requires:

- The preprocess check `i == len(swch.Clauses)-1` then flagged the reordered
  default as the final clause and rejected the `fallthrough`.
- Even if accepted, the runtime jump `BodyIndex + 1` would have landed past
  the end of the (reordered) slice.

## Decision

### `toClauses` preserves textual order

`SwitchStmt.Clauses` now reflects source order exactly. Duplicate-default
detection is preserved.

### `doOpSwitchClause` defers `default` until cases are exhausted

The matching pass now skips `default` clauses regardless of position. When
all cases have been tried without a match, a single backward scan finds the
`default` (if any) and runs it.

### `doOpTypeSwitch` defers `default` similarly

Type switches iterate clauses inline. The loop now records the default index
without matching it, executes the first matching non-default clause, and
falls back to default only when no case matches.

### `fallthrough` preprocess check is unchanged

With textual ordering preserved, the existing index-based check
(`i == len(swch.Clauses)-1`) now correctly identifies the textually-last
clause and continues to reject `fallthrough` from it — matching Go.

## Key files

| File | Role |
|------|------|
| `gnovm/pkg/gnolang/go2gno.go` | `toClauses`: keep textual order |
| `gnovm/pkg/gnolang/op_exec.go` | `doOpSwitchClause`, `doOpTypeSwitch`: defer default |
| `gnovm/tests/files/switch42.gno` | `default:` first, fallthrough to next case |
| `gnovm/tests/files/switch43.gno` | `case` falling through into mid-block `default` |
| `gnovm/tests/files/switch44.gno` | `default:` textually last with fallthrough still errors |
| `gnovm/tests/files/typeswitch9.gno` | type switch with `default` first, matching case wins |
| `gnovm/tests/files/typeswitch10.gno` | type switch with `default` first, no case matches |

## Alternatives considered

- **Add a `DefaultClauseIndex` field to `SwitchStmt`.** Slightly faster runtime
  (no scan when no case matches), but adds persistent node state for an
  already-cheap O(n) scan over what is typically a handful of clauses. Rejected
  for minimal-change reasons.
- **Track textual positions separately and keep reordering.** Would require
  parallel index maps in both the parser and the fallthrough logic. The
  ordering change is the simpler invariant.

## Consequences

- `fallthrough` from a non-last `default:` now behaves as in Go.
- `default:` selection semantics are unchanged (still "no case matched").
- A small runtime cost in `doOpSwitchClause`: when no case matches, a linear
  scan over clauses locates the default. Switches with very many clauses pay
  O(n) extra work in this path only.
