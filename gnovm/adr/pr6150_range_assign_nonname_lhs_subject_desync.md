# PR6150: Fix range-subject stack desync for ASSIGN-form range with a non-name LHS

## Context

The ASSIGN form of a `range` statement (`for k, v = range x`, as opposed to the
DEFINE form `:= range`) whose key or value target is a non-`NameExpr` lvalue —
an `IndexExpr` (`m[i]`), `SelectorExpr` (`s.F`), or `StarExpr` (`*p`) — ranged
over the **wrong value** and silently corrupted state. Verified against `go run`
as the oracle (built VM diff):

```go
var i int
mm := map[string]string{"abc": "orig"}
for i, mm["abc"] = range []string{"A", "B", "C", "D", "E"} {
}
println(i, mm["abc"], len(mm["abc"]))
```

| | `i` | `mm["abc"]` | `len(...)` |
|---|---|---|---|
| Go (oracle) | 4 | "E" | 1 |
| GnoVM (before) | 2 | 99 | **host panic** `unexpected type for len(): uint8` |

The VM ranged over the map **key** string `"abc"` (length 3, last index 2) and
wrote byte `'c'`=99 as a `uint8` into a `string`-typed map slot. The type-confused
value **persists** (the slot flips from `StringValue` to `uint8`), and a later
`len(mm["abc"])` throws a raw Go panic — not an `m.Panic`, so **not
Gno-recoverable**: the keeper turns it into a failed tx with gas already burned,
and the key is permanently bricked across transactions. `SelectorExpr` and
`StarExpr` targets produce the analogous uncatchable `unexpected type for len():
main.S` / `*string`.

The bug is deterministic (every node computes the same wrong result), so it is a
correctness/soundness hole and latent persistent-state corruption rather than a
consensus fork. It requires a realm author to write the legal-but-unusual
`for k, m[i] = range x` shape. No filetest exercised the `= range` form with a
non-name LHS.

### Root cause

Two stack sites read the range subject `X` at a fixed offset that is only valid
when both range targets are `NameExpr`.

The `RangeStmt` setup (`op_exec.go`, `case *RangeStmt`) evaluates `X` onto the
value stack and then, in the ASSIGN case, pushes each LHS's pointer operands
*above* `X` via `PushForPointer`. `numStackValuesForPointer` (`machine.go`)
gives the per-LHS count: `NameExpr` 0, `IndexExpr` 2, `SelectorExpr`/`StarExpr`/
`CompositeLitExpr` 1. The range handlers then read the subject as a fixed
`xv := m.PeekValue(1)`:

- `OpRangeIter` / `OpRangeIterArrayPtr`,
- `OpRangeIterString`,
- `OpRangeIterMap`.

With a non-name LHS, `PeekValue(1)` is a pointer operand, not `X`. (The write
*target* is resolved correctly — `PopAsPointer` pops from the top — so the
type-confused value lands in the intended slot, which is what makes the
corruption persist.) `doOpAssign` is the correctly-fixed sibling for tuple
assignment (PR #5765): it sums `numStackValuesForPointer` over the LHS. The fix
here brings the range handlers in line.

A **second, coupled** offset appears on the `goto` path. `bs.NumValues` is
captured at the `-2` init phase *while the LHS operands are still on the stack*,
and the `GOTO` handler (`op_exec.go`) restores `m.Values` to that length while
setting `NextBodyIndex >= 0` (a body position). The `continue` path
(`PeekFrameAndContinueRange`, `machine.go`) instead restores to `fr.NumValues+1`
= **X-only**. For a non-name ASSIGN LHS these two restore points disagree by the
operand count, so after a `goto` out of the body the value stack is left too tall
and the next subject read (which runs with the operands already consumed, offset
0) reads a stale operand slot. Fixing only the three reads would convert the
pre-existing loud `goto` panic into **silent** persisted corruption — strictly
worse — so both sites must be corrected together.

## Decision

Introduce one helper that reports how many value-stack entries sit above `X`
because of ASSIGN-form LHS pointer operands, gated so it is non-zero only while
those operands are actually present (`NextBodyIndex < 0`, the `-2`/`-1` phases;
they are popped in the `-1` phase):

```go
func (bs *bodyStmt) rangeSubjectDepth() int {
	if bs.Op != ASSIGN || bs.NextBodyIndex >= 0 {
		return 0
	}
	n := 0
	if bs.Key != nil {
		n += numStackValuesForPointer(bs.Key)
	}
	if bs.Value != nil {
		n += numStackValuesForPointer(bs.Value)
	}
	return n
}
```

Apply it at both consumers, in all three range handlers:

- **Subject read**: `xv := m.PeekValue(1 + bs.rangeSubjectDepth())`.
- **`bs.NumValues` capture** (the `-2` init, where `NextBodyIndex == -2` so the
  depth is the true operand count): `bs.NumValues = len(m.Values) - bs.rangeSubjectDepth()`,
  recording the X-only length so the `goto` restore agrees with the `continue`
  restore.

This is a no-op for every previously-passing shape — DEFINE form, name-only or
blank ASSIGN, and `for range x` all have depth 0 — and changes only the
non-name-ASSIGN path.

## Alternatives considered

- **Fix only the three subject reads.** Rejected: leaves the `goto` restore
  mismatch, turning a loud pre-existing panic into silent persisted corruption
  for `goto`-out-of-body cases (measured — see Verification).
- **Special-case range frames in the `GOTO` handler** (restore to
  `fr.NumValues+1` like `continue`). Equivalent in effect, but pushes range-only
  knowledge into the generic branch handler; correcting `bs.NumValues` at the
  capture keeps the invariant local to the range handlers and leaves the generic
  `GOTO`/for-loop path untouched (`OpBody`/`OpForLoop` captures have depth 0).
- **Compute the offset unconditionally (no `NextBodyIndex` gate).** Rejected:
  during body execution the operands have already been popped, so a non-zero
  offset would read past the top of the stack (out-of-bounds) — and
  `OpRangeIterString` re-reads the subject in that phase.

## Consequences

- `for k, m[i]/s.F/*p = range x` over slice, array, array-pointer, string, and
  map subjects now matches Go, including with `break`, `continue`, `goto`, and
  nested loops.
- No change to any previously-passing behavior (depth 0 everywhere else).
- Scope caveat — one **pre-existing, unrelated** divergence is unchanged and
  intentionally left out of this fix: when BOTH the key and value targets are
  non-name lvalues AND each carries a side effect in its index/base (e.g.
  `for a[f()], b[g()] = range x`), the VM evaluates the value target's operands
  before the key target's, opposite to Go's left-to-right order. This predates
  the fix (such programs previously panicked before reaching the assignment, so
  the fix only makes them run) and is rooted in the unchanged `RangeStmt` setup
  push order (Key pushed before Value; the LIFO op stack reverses them), not in
  the offset logic changed here. It is deterministic (no consensus fork) and
  affects only that two-non-name-target shape — the single-non-name-target
  examples above match Go exactly. Correcting it means reworking the setup
  push order together with the `-1`-phase pop order, a separate change.
- Not a consensus/gas change: the edit only moves a read offset; no metering path
  depends on it.
- Regression coverage added:
  - `gnovm/tests/files/types/assign_range_lhs_index.gno` (slice/string/map,
    IndexExpr LHS),
  - `gnovm/tests/files/types/assign_range_lhs_selector_star.gno`
    (SelectorExpr/StarExpr LHS),
  - `gnovm/tests/files/types/assign_range_lhs_goto.gno` (the `goto` interaction,
    plus a name-only control),
  - `gno.land/pkg/integration/testdata/range_assign_lhs_persist.txtar`
    (cross-transaction persistence: pre-fix tx1 returns `(2 int)` and a later
    `len(m[k])` fails; post-fix tx1 returns `(4 int)` and the value reads back).

## Verification

- Differential vs `go run` for the five original repros and the reviewer's
  `goto`/`break`/`continue`/nested attack set — all match.
- Full enumeration fuzz — 375 programs (5 subject kinds {slice, array,
  array-pointer, string, map} × 3 key-LHS × 4–5 value-LHS {name, blank, index,
  selector, star} × 5 control-flow shapes {none, forward goto, continue, break,
  nested}): the fixed VM matches `go run` on all 375; the unpatched HEAD binary
  diverges on 275 — exactly the non-depth-0 shapes (the 100 that agree are the
  key∈{name,blank} ∧ value∈{name,blank} cases, where the historical fixed offset
  was already correct).
- `gnovm/tests/files/types/assign_range_lhs_*.gno` panic/misbehave before the
  fix, pass after.
- `range_assign_lhs_persist.txtar` fails before the fix (tx1 `no match for
  (4 int)`), passes after.
- AGENTS.md suite: `go test ./gnovm/pkg/gnolang/ -run Files -test.short`,
  `go test ./gno.land/pkg/sdk/vm/ -run Gas`,
  `go test ./gno.land/pkg/integration/ -run TestTestdata`, and
  `cd examples && go run ../gnovm/cmd/gno test ./...`.
