# PRXXXX: Fix `GotoJump` stmt-stack underflow on backward goto out of nested loops

## Context

A backward `goto` that jumps out of two or more nested `for`/`range` loops
crashed the interpreter with:

```
runtime error: slice bounds out of range [:-1]
```

at `Machine.GotoJump` (`gnovm/pkg/gnolang/machine.go`). Minimal reproduction —
three nested loops jumping back to a label in the enclosing function body:

```go
func main() {
	i := 0
top:
	println("i =", i)
	for a := 0; a < 2; a++ {
		for b := 0; b < 2; b++ {
			for c := 0; c < 2; c++ {
				i++
				if i < 3 {
					goto top
				}
			}
		}
	}
	println("done", i)
}
```

The crash is independent of the gas profiler (reproduces with it disabled); it
is a pre-existing accounting bug in `GotoJump`.

### Root cause

At preprocess time `findGotoLabel` computes, for a `GOTO` `BranchStmt`:

- `FrameDepth` — the number of loop frames (`ForStmt`/`RangeStmt`/
  `SwitchClauseStmt`) crossed between the goto and the label, and
- `BlockDepth` — the number of non-frame block scopes crossed since the last
  frame crossing.

Each loop pushes exactly **one** frame (`PushFrameBasic`) and **one** sticky
`bodyStmt` onto `m.Stmts`. A frame records `NumStmts = len(m.Stmts)` *before*
its own `bodyStmt` is pushed.

`GotoJump(depthFrames, depthBlocks)` reset the machine to the outermost popped
frame `fr = m.Frames[len-depthFrames]`, then truncated the stmt stack twice:

```go
m.Stmts = m.Stmts[:fr.NumStmts]           // (1) baseline reset
m.Stmts = m.Stmts[:len(m.Stmts)-depthFrames] // (2) extra pop — BUG
```

Line (1) already drops the `bodyStmt`s of **every** popped loop frame at once,
because `fr.NumStmts` predates the outermost loop's `bodyStmt` and all inner
loops pushed theirs afterward. Line (2) then subtracts `depthFrames` a second
time. Two consequences:

- When `fr.NumStmts >= depthFrames`, it silently over-truncates, but the `GOTO`
  handler in `op_exec.go` immediately re-truncates `m.Stmts` to the target
  block's `bodyStmt.NumStmts` and re-extends the slice via its retained
  capacity, masking the bug.
- When `fr.NumStmts < depthFrames` — which happens once the number of crossed
  loop frames exceeds the number of enclosing sticky `bodyStmt`s (e.g. three
  nested loops directly under the function body, where `fr.NumStmts == 2` but
  `depthFrames == 3`) — the index goes negative and the slice expression
  panics.

The `m.Blocks` handling is *not* symmetric: after resetting to `fr.NumBlocks`
(which likewise excludes all popped loop blocks) it pops an additional
`depthBlocks`, which is correct because `BlockDepth` counts scopes *within* the
target frame between `fr` and the label. There is no equivalent second pop
needed for stmts.

## Decision

Remove the erroneous second truncation (line 2). `m.Stmts[:fr.NumStmts]` is the
correct baseline reset, consistent with the `Ops`/`Values`/`Exprs`/`Blocks`
resets on the surrounding lines, and the `op_exec.go` `GOTO` handler
authoritatively sets the final `m.Stmts` length to the target block's
`bodyStmt.NumStmts` afterward.

The final `m.Stmts` state is therefore unchanged in the non-crashing cases (the
old code relied on the same `op_exec.go` re-truncation to fix its
over-truncation), so this is a pure crash fix with no behavioral change to
previously-working gotos.

## Alternatives considered

- **Clamp the subtraction at zero** (`max(0, len-depthFrames)`): avoids the
  panic but keeps the redundant, misleading double-pop and its reliance on
  slice-capacity re-extension. Rejected in favor of removing the dead
  truncation outright.
- **Recompute stmt depth in `findGotoLabel`** and pop exactly that many:
  unnecessary — `fr.NumStmts` already encodes the correct baseline, and
  `op_exec.go` owns the final positioning.

## Consequences

- Backward (and forward) gotos out of arbitrarily deep nested loops now execute
  correctly, matching Go semantics (verified against `go run` for 3- and
  4-level nesting, nested `range` loops, and gotos additionally crossing block
  scopes).
- No change to any previously-passing behavior; the whole `-run Files` suite and
  the full `goto*` / `heap_alloc_gotoloop*` / `loopvar_goto*` families pass.
- Regression coverage added: `gnovm/tests/files/goto10.gno`.

## Verification

- `go test ./gnovm/pkg/gnolang/ -run Files -test.short` — 0 failures.
- `go test ./gnovm/pkg/gnolang/ -run 'TestFiles/(goto|heap_alloc_gotoloop|loopvar_goto)' -test.short` — ok.
- New filetest `goto10.gno` panics before the fix, passes after.
