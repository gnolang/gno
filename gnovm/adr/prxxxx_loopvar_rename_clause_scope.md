# PRxxxx: Loop-var rename pass resolves clause expressions in the wrong scope

## Context

A three-clause `for` whose init declares a name that shadows an outer
variable, and reads that outer variable on the RHS, failed to preprocess:

```go
func main() {
	x := 7
	for x, i := x, 0; i < 1; i++ {
		x = 0
		_ = x
	}
	f := func() { println(x) }
	f()
}
```

```
main/loopvar_shadow_init_capture_1.gno:10:14-15: name x.loopvar not declared
```

`go run` prints `7`. The position is the RHS `x` of the init. Neither the
closure nor the two-name init is needed: `for x := x + 1; x < 10; x++ {}`
fails the same way. Present on upstream `master` (checked at `640d2b784`
and `b7845fb2d`); unrelated to PR #6196's `cur` write rules.

While fixing it, three mirror-image cases in the same pass turned up, all
also failing on `master` and all passing under `go run`:

```go
for x := 1; x < 3; x++ {
	for x := range make([]int, x) {}     // name x not declared
}
for x := interface{}(1); x != nil; x = nil {
	switch x := x.(type) {}              // name x not declared
}
for k, v := 0, 0; k < 1; k++ {
	for k, v = range []int{7, 8} {}      // name k not declared
}
```

### Root cause

`initStaticBlocks1` (`gnovm/pkg/gnolang/preprocess.go`) gives each
for-init declared name a `.loopvar` suffix so the interpreter can allocate a
per-iteration copy (go1.22 semantics). It is a single `TranscribeB` walk
over a stack of frames (`map[Name]bool`); a `NameExpr` is renamed iff its
name is in the top frame and its position is not a declaration site.

Each statement mutated its frame at its `TRANS_BLOCK`, i.e. *before* any of
its children were walked. But every one of these statements has a clause
that Go resolves before the statement's own declared names exist:

| stmt | declared names | clause resolved before they exist |
|---|---|---|
| `ForStmt` (DEFINE init) | init LHS, **added** to the frame | init RHS |
| `RangeStmt` (DEFINE) | key/value, **deleted** from the frame | `X` |
| type `SwitchStmt` | `VarName`, **deleted** from the frame | `X` |

So `x, i := x, 0` became `x.loopvar, i.loopvar := x.loopvar, 0` (the RHS
now names the variable this statement is defining), and in the range and
type-switch cases the `X` reference to the enclosing loop's `x.loopvar` was
left as a bare `x`, which no longer exists in that scope.

The fourth case is a declaration-site check that ignored the operator: a
range key/value was treated as a declaration even for `for k = range`,
where it is an assignment target and must be renamed like any other
reference (the assign-LHS check already tested `Op == DEFINE`).

The for-init half is a regression from the O(N) rewrite in PR #5642. The
earlier per-loopvar walk explicitly skipped the loop's own init
(`if last == bn && ftype == TRANS_FOR_INIT { return n, TRANS_SKIP }`), and
that exclusion was not carried over.

## Decision

`transcribe` already fires a `TRANS_BLOCK2` stage on `SwitchStmt` between
`.Init`/`.X` and the clauses ("special block case for after .Init and
.X"). `ForStmt` and `RangeStmt` have the same shape, so the walker now
fires `TRANS_BLOCK2` for them too, after `.Init` and after `.X`
respectively (`gnovm/pkg/gnolang/transcribe.go`). `initStaticBlocks1` then
has one rule instead of three hooks:

- `TRANS_BLOCK`: every block-introducing node does a plain `pushClone(nil)`.
- `TRANS_BLOCK2`: the statement applies its own declared names to that
  frame — `ForStmt` renames the init LHS and adds the names, `RangeStmt`
  (`DEFINE`) deletes key/value, a type `SwitchStmt` deletes `VarName`.
- `TRANS_LEAVE`: unchanged — an inner `DEFINE`/`var`/`type` deletes, a
  `NameExpr` is renamed iff in the top frame and not a declaration site.

`isLoopvarDeclSite` is extracted from the `NameExpr` case; range key/value
count as a declaration site only when the `RangeStmt` is a `DEFINE`, the
same test the assign-LHS case already made.

Because the for-init add now sits next to the rename, the existing
`HasSuffix(".loopvar")` skip covers both. `initStaticBlocks` runs twice on
file nodes (`PredefineFileSet`, then `Preprocess`) and must stay idempotent:
on the second run every LHS is already suffixed, nothing is renamed and
nothing is added, so every frame stays empty and the whole body takes the
existing empty-frame fast path — exactly the pre-#5642 and pre-fix
behavior. A small `ownFrame()` helper pushes the loop's frame lazily on the
first add when the fast path had skipped it (the parent was empty, so a
fresh map is the clone); deletes never need it, since a delete on a shared
empty frame is a no-op.

References in a clause to an *enclosing* loop's variable are still renamed,
because those names are in the cloned parent frame throughout:

```go
for i := 0; i < 2; i++ {
	for i := i * 10; i%10 < 2; i++ { ... } // RHS i -> outer i.loopvar
}
```

This resolved correctly before (inner block defines `i.loopvar`, RHS finds
the outer block's `i.loopvar`); the fix leaves that AST unchanged.

Emitting `TRANS_BLOCK2` for two more node types touches every `Transcribe`
/`TranscribeB` callback: all 14 live in `preprocess.go`, none switch on
stage with a panicking default, `preprocess1`'s `TRANS_BLOCK2` case matches
only `*SwitchStmt` and returns `TRANS_CONTINUE` otherwise, and `TranscribeB`
already handles `TRANS_BLOCK2` generically (including the pop on
`TRANS_SKIP`). No callback outside the package exists.

## Alternatives considered

- **Hook each statement where the pass already gets control**, without
  touching the walker: add the for-init names at `TRANS_LEAVE` of the init
  `AssignStmt` (`ftype == TRANS_FOR_INIT`), delete range key/value at
  `TRANS_LEAVE` of `X` (`ftype == TRANS_RANGE_X`, in a trailing check after
  the node-type switch), and use the existing `TRANS_BLOCK2` for the type
  switch. This was the first version. Rejected: three hook shapes for one
  rule; the range hook only fired for a bare-name `X` because the `NameExpr`
  case had been restructured to fall through instead of `return`ing (a
  future early return would silently skip the delete and rename the body's
  `xs` in `for _, xs := range xs` to the enclosing loop's slice); and
  recovering the for-init names at the leave by stripping the suffix
  re-added them on the second `initStaticBlocks` run, which repopulated
  every loop frame and defeated the empty-frame fast path for the whole
  body (measured: +97 mallocs on `stdlibs/strings/strings.gno`, +1199 on a
  loop with 200 `if`s, all on the second pass).
- **Detect "inside the loop's own init" at the `NameExpr`** by walking `ns`
  up to the `AssignStmt` and checking it is `ForStmt.Init`. Rejected: the RHS
  can nest arbitrarily (`x + 1`, a call, a func literal whose body reads
  `x`), and a func literal pushes its own frame cloned from the ForStmt's, so
  every nested frame would need the same exception.
- **Restore a `TRANS_SKIP` on `TRANS_FOR_INIT`** as the pre-#5642 walk did.
  Rejected: in a single walk it would also skip renaming enclosing-loop
  references in the init RHS (the nested case above), which the old walk
  only got right because it ran once per loopvar with `last == bn` scoping
  the skip to that loop.
- **Delete the type-switch var per clause** (at each `SwitchClauseStmt`
  push), matching Go's per-clause scoping exactly. Equivalent, since the
  clauses are the only children walked after `TRANS_BLOCK2`; the single
  delete keeps the three cases parallel.
- **Fix only the for-init case** as reported. Rejected: the other three are
  the same confusion in the same function, each a few lines, and a reviewer
  reading the new frame discipline would otherwise have to re-derive why
  the range and type-switch pushes are still allowed to run early.

## Consequences

- `for x := <expr reading outer x>; ...` and multi-name variants
  (`for i, j := 0, i`) preprocess and run with Go semantics; a func literal in
  the init RHS captures the outer variable, not the loop's.
- `for x := range <expr reading loop x>`, `switch x := x.(type)` and
  `for k, v = range ...` inside a loop declaring `x`/`k`/`v` now work.
- Per-iteration copies of a loop's own variable are unchanged (closures in
  the body still see one value per iteration).
- No gas change. `initStaticBlocks1` stays O(N) and allocates fewer
  frames: `ForStmt`, `RangeStmt` and `SwitchStmt` now take the empty-parent
  fast path (one map per statement before), and the second
  `initStaticBlocks` pass over a file allocates none at all.
- `transcribe` fires one extra callback per `ForStmt`/`RangeStmt`
  (`TRANS_BLOCK2`), a stage-switch miss in every callback but this one.
- Regression coverage, each failing before the fix with the quoted error:
  `gnovm/tests/files/loopvar_shadow_init_capture_{1,2,3}.gno`,
  `loopvar_range_x_shadow_1.gno` (including a bare-name `X`),
  `loopvar_typeswitch_x_shadow_1.gno`, `loopvar_range_assign_1.gno`. The
  `Preprocessed:` directives pin the resolution: the clause reads the
  enclosing `x<~VPBlock(2,0)>` / `x.loopvar<VPBlock(2,0)>` while the body
  sees the statement's own name.

## Verification

- Expected outputs of every new filetest taken from `go run` on the same
  source, not derived by hand.
- `go test ./gnovm/pkg/gnolang/ -short` (whole package, includes
  `TestFiles`, since the walker changed)
- `cd examples && go run ../gnovm/cmd/gno test ./...`
