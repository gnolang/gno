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
corruption persist.) `doOpAssign` hit the same class of bug for tuple assignment
and was fixed in PR #5765 by summing `numStackValuesForPointer` over the LHS; the
range handlers, unlike `doOpAssign`, have a frame of their own to anchor to, so
they are fixed differently — see Decision.

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

Address `X` by its **absolute** index in `m.Values` rather than by an offset from
the top of the stack.

The `RangeStmt` setup calls `PushFrameBasic` and only *then* evaluates `X`, so
the value-stack length the frame recorded is exactly `X`'s slot. Everything above
that slot belongs to the loop — the ASSIGN-form LHS pointer operands, and mid-body
whatever the body is evaluating — and none of it can be mistaken for `X`.
`PeekFrameAndContinueRange` already relies on this identity when `continue`
restores the stack to `fr.NumValues+1` ("back to `X` only"):

```go
func (m *Machine) rangeFrame() *Frame {
	fr := m.LastFrame()
	if debugAssert {
		if _, ok := fr.Source.(*RangeStmt); !ok {
			panic(fmt.Sprintf(
				"expected the last frame to be the range's own frame, got %T",
				fr.Source,
			))
		}
	}
	return fr
}
```

Apply it at both consumers, in all three range handlers:

- **Subject read**: `xv := &m.Values[fr.NumValues]`.
- **`bs.NumValues` capture** (the `-2` init): `bs.NumValues = fr.NumValues + 1`,
  recording the X-only length so the `goto` restore agrees with the `continue`
  restore.

This is a no-op for every previously-passing shape — DEFINE form, name-only or
blank ASSIGN, and `for range x` push nothing above `X`, so `fr.NumValues` is the
top of the stack and the read is identical to the historical `PeekValue(1)`.

The point of taking the frame's index rather than computing the operand count is
that the read stops depending on the LHS shape at all: there is no per-LHS
arithmetic to keep in sync with `PushForPointer`/`PopAsPointer`, and no phase
gate to get wrong. The bug class becomes unrepresentable instead of merely
computed correctly.

## Alternatives considered

- **Fix only the three subject reads.** Rejected: leaves the `goto` restore
  mismatch, turning a loud pre-existing panic into silent persisted corruption
  for `goto`-out-of-body cases (measured — see Verification).
- **Special-case range frames in the `GOTO` handler** (restore to
  `fr.NumValues+1` like `continue`). Equivalent in effect, but pushes range-only
  knowledge into the generic branch handler; correcting `bs.NumValues` at the
  capture keeps the invariant local to the range handlers and leaves the generic
  `GOTO`/for-loop path untouched (`OpBody`/`OpForLoop` capture nothing above X).
- **Sum `numStackValuesForPointer` over the LHS and read at
  `PeekValue(1+depth)`**, the way `doOpAssign` (PR #5765) does for tuple
  assignment. This was the first form of the fix and it is correct — the two
  formulations were run against each other, asserting equality at all six sites
  over 2029 programs, with no disagreement. It was dropped because it is
  strictly more fragile: the depth has to stay in sync with `PushForPointer`
  and `PopAsPointer` for every pointer-LHS shape, and it needs an
  `Op != ASSIGN || NextBodyIndex >= 0` gate encoding *when* the operands happen
  to be on the stack. Both are invariants about other code, and if either drifts
  the handlers silently go back to reading the wrong slot. `doOpAssign` has no
  such choice — it runs in a single phase with no frame of its own — so the
  duplication is not worth preserving for symmetry.
- **Cache the operand count in a `bodyStmt` field** instead of recomputing it.
  Not viable: `Block` embeds `bodyStmt` and `alloc.go` asserts
  `_allocBlock = 536 == unsafe.Sizeof(Block{})`, so a new field shifts
  allocation gas and forces a `pb3_gen.go` regeneration of a persisted type.

## Consequences

- `for k, m[i]/s.F/*p = range x` over slice, array, array-pointer, string, and
  map subjects now matches Go, including with `break`, `continue`, `goto`, and
  nested loops.
- No change to any previously-passing behavior (nothing sits above `X` in any of
  those shapes, so the frame index and the old `PeekValue(1)` coincide).
- Scope caveat — two **pre-existing** divergences in the `RangeStmt` setup are
  unchanged and intentionally left out of this fix. Both are rooted in the same
  place: the setup evaluates the LHS pointer operands unconditionally, once,
  before the loop starts (`PushForPointer(cs.Key)` then `PushForPointer(cs.Value)`),
  where Go evaluates them as part of each iteration's assignment. Both predate
  the fix — the base binary behaves identically — and both are deterministic (no
  consensus fork), but this fix is what makes the affected programs run far
  enough to observe them.
  - **Order.** When BOTH targets are non-name lvalues AND each carries a side
    effect in its index/base (`for a[f()], b[g()] = range x`), the value
    target's operands evaluate before the key target's, opposite to Go's
    left-to-right order — the ops are pushed Key-then-Value and the op stack is
    LIFO. Go: `f g`; GnoVM: `g f`.
  - **Count, on a zero-iteration range.** With even a SINGLE non-name target,
    a subject that yields no iterations still evaluates that target's operands
    once, where Go evaluates them zero times. `for i, m[f()] = range []string{}`
    calls `f` once (Go: never); `var p *[3]int; for i, (*p)[0] = range []int{}`
    panics with a nil pointer dereference (Go: runs clean).

  Correcting either means reworking the setup push order and moving the pushes
  inside the iteration, together with the `-1`-phase pop order — a separate
  change.
- **This is a state-machine change**, not a no-op. It is not a gas-*constant*
  change — no metering constant is touched — but for the affected shape both the
  result and the gas differ, because `m.incrCPU(OpCPUSlopeRangeIterArray * ll)`
  is fed by the subject whose read this PR corrects, and the trip count changes
  with it. The `range_assign_lhs_persist.txtar` `Corrupt` tx is identical either
  side and burns 1,786,673 gas returning `(2 int)` before, 1,789,138 gas
  returning `(4 int)` after. Any node replaying a block containing such a tx on
  a different binary version computes a different AppHash, so this needs a
  coordinated upgrade like any other VM semantics fix; there is no VM-level
  version gate in the tree to hide it behind.
- Regression coverage added:
  - `gnovm/tests/files/types/assign_range_lhs_index.gno` (IndexExpr LHS over
    slice, array, array-pointer, string and map subjects, plus the key-only
    `for m[k] = range x` form where `bs.Value` is nil),
  - `gnovm/tests/files/types/assign_range_lhs_selector_star.gno`
    (SelectorExpr/StarExpr LHS),
  - `gnovm/tests/files/types/assign_range_lhs_goto.gno` (the `goto` interaction
    against all three handlers — `OpRangeIter`, `OpRangeIterArrayPtr`,
    `OpRangeIterString`, `OpRangeIterMap` — including IndexExpr targets so more
    than one stack entry sits above `X`, plus a name-only control),
  - `gno.land/pkg/integration/testdata/range_assign_lhs_persist.txtar`
    (cross-transaction persistence: pre-fix tx1 returns `(2 int)` and a later
    `len(m[k])` fails; post-fix tx1 returns `(4 int)` and the value reads back).

  The `goto` coverage is deliberately spread across all three handlers: with the
  earlier slice-only test, reverting the `bs.NumValues` fix in the string and map
  handlers left the entire `TestFiles` suite green while a two-line program
  (`for k, m["k"] = range "XYZ" { goto L; L: }`) crashed with
  `slice bounds out of range [1:0]`.

## Verification

- Differential vs `go run` for the five original repros and the reviewer's
  `goto`/`break`/`continue`/nested attack set — all match.
- Full enumeration fuzz — 375 programs (5 subject kinds {slice, array,
  array-pointer, string, map} × 3 key-LHS × 4–5 value-LHS {name, blank, index,
  selector, star} × 5 control-flow shapes {none, forward goto, continue, break,
  nested}): the fixed VM matches `go run` on all 375; the unpatched HEAD binary
  diverges on 275 — exactly the shapes with operands above `X` (the 100 that
  agree are the key∈{name,blank} ∧ value∈{name,blank} cases, where the
  historical fixed offset was already correct).
- Wider re-run on review — 2016 programs (7 subject kinds, adding empty-slice
  and nil-map, × 6 key-LHS × 6 value-LHS {name, blank, slice-index, map-index,
  selector, star} × 8 control-flow shapes, adding backward goto, goto out of a
  nested block, and a body that leaves values on the stack): 0 divergences from
  `go run`. Plus hand-written probes for nested `IndexExpr`/`SelectorExpr`,
  `(*p).A.B`, `**pp`, `mm["a"]["b"]`, parenthesized lvalues, the key-only form,
  panic/recover mid-range, labeled break/continue, goto from an inner range out
  to an outer range's body label, `for i, sl[i] = range src` (the index depends
  on the key being assigned), and an LHS aliasing the subject — all match `go
  run`; the base binary panics on nearly all of them.
- Equivalence of the two candidate formulations: a build asserting
  `len(m.Values)-(1+depth) == m.LastFrame().NumValues` and
  `bs.NumValues == m.LastFrame().NumValues+1` at all six sites ran the 2016 fuzz
  programs plus the probes with zero assertion failures, i.e. the frame index
  and the summed operand depth never disagree.
- `gnovm/tests/files/types/assign_range_lhs_*.gno` panic/misbehave before the
  fix, pass after.
- `range_assign_lhs_persist.txtar` fails before the fix (tx1 `no match for
  (4 int)`), passes after.
- AGENTS.md suite: `go test ./gnovm/pkg/gnolang/ -run Files -test.short`,
  `go test ./gno.land/pkg/sdk/vm/ -run Gas`,
  `go test ./gno.land/pkg/integration/ -run TestTestdata`, and
  `cd examples && go run ../gnovm/cmd/gno test ./...`.
