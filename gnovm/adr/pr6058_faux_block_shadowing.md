# Shadowing a switch/if init name from a case body

## Context

`IfStmt` and `SwitchStmt` are **faux blocks** in GnoVM. Their init statement
declares names, but there is no runtime `Block` for the `IfStmt`/`SwitchStmt`
itself: `doOpExec` allocates one `Block` for the whole statement and every
`IfCaseStmt`/`SwitchClauseStmt` executes inside it. So that one block can serve
both scopes with a flat, depth-1 index space, the parent's names are **copied
into each case's static block**, occupying its leading slots in the parent's
order:

- `initStaticBlocks` reserves them (`preprocess.go`, `*IfCaseStmt` and
  `*SwitchClauseStmt` under `TRANS_BLOCK`),
- `preprocess1` fills in their types (`copyFromFauxBlock` plus the copy loops in
  the same cases),
- `Block.ExpandWith` grows the runtime block to the matched case's name count.

`fauxChildBlockNode` and the `path.Depth + i - fauxChild` arithmetic in
`fillNameExprPath` collapse the faux level so both scopes resolve at depth 1.

The copy makes the case's static block claim its parent's names as its own
locals. A declaration in the case body reusing one of them therefore found the
name already reserved and **wrote over the copy** instead of getting a slot of
its own:

```go
switch v := 1; v {
case 1:
    v := "shadow" // panic: StaticBlock.Define2(v) cannot change .T; was int, new string
    println(v)
}
```

When the shadowing declaration happened to have the *same* type there was no
panic — just a silently clobbered outer binding. `case 1: v := 99` left the
switch's own `v` reading `99`, observable from a later clause reached by
`fallthrough`. The same applies to `if v := ...` branches, and to both `:=` and
`var`/`const` declarations.

Go accepts all of these: a case body is a nested scope, and the spec's
"innermost declaration wins" applies. The gno type-checker (`go/types`) already
accepted them; only the gno preprocessor rejected or miscompiled them.

## Decision

Give the shadowing declaration a **second slot for the same name in the same
static block**, resolved last-wins.

1. **`StaticBlock.Reserve` appends** rather than no-oping when the name's only
   existing slot is one of the leading faux-copied ones
   (`idx < numFauxCopiedNames()`). It stays idempotent for the copy loop
   itself — which re-runs whenever an already-initialized block node is
   re-preprocessed, as stdlib nodes are — by comparing the reserving
   declaration against the slot's recorded `NameSource`.

   The comparison is on `(Origin, Type, Index)`, not on the `*NameExpr`.
   Every one of the eleven `Reserve` call sites passes a stable triple, but
   one does *not* pass a stable `*NameExpr`: the type switch variable's site
   (`preprocess.go`, `initStaticBlocks`) builds `&NameExpr{Name: ss.VarName}`
   fresh on each pass. Keying on the pointer would have made that site's
   correctness depend on its slot landing exactly at the boundary index, so
   that the earlier `idx >= numFauxCopiedNames()` return fired first.

2. **`StaticBlock.GetLocalIndex` resolves a name to its last slot.** Duplicates
   are otherwise impossible (`Define2` overwrites in place; only `defineNew`
   appends), so this is a no-op for every block except a faux case block that
   contains a shadow. `buildNameIndex` switches to last-wins to match the
   linear scan.

3. **Uses before the shadowing declaration still resolve to the outer name**,
   with no new machinery. The shadow's slot is reserved with a nil type until
   `preprocess1` reaches its declaration, and `fillNameExprPath` already treats
   "reserved but not yet typed" as *not defined here* and walks up the parent
   chain. This is the same mechanism that makes shadowing work in an ordinary
   nested block (`println(x); x := 2` — see `tests/files/define1.gno`).

4. **The faux copy writes by index** (`defineFauxCopy`), because by name it
   would now find the shadow's slot.

A **type switch's variable is deliberately left unshadowable**. Go declares it
in each clause's own block, so a clause body redeclaring it is an error, not a
shadow (`go/types`: "no new variables on left side of :="). `Reserve` refuses
to append over a slot whose `NameSource.Type` is `NSTypeSwitch`, so gno keeps
rejecting the program exactly as before. That is an explicit test rather than a
consequence of the variable's slot index, which would otherwise be the only
thing stopping it.

## Alternatives considered

**Make the case a real child block at depth 2.** The structurally clean fix:
stop copying, give the case its own runtime `Block` whose parent is the
statement's. It removes the faux-block concept outright. Rejected as far too
invasive for this bug — it touches `fauxChildBlockNode`, every depth
computation in `fillNameExprPath` and `GetBlockNodeForPath`, `ExpandWith`,
`doOpIfCond` / `doOpSwitchClauseCase` / `doOpTypeSwitch`, `fallthrough`,
break/continue frame handling, and block persistence.

**Mangle the shadow's name** (`v` → `v~1`, as `addHeapCapture` does with
`~name`). Keeps `Names` unique, but every by-name lookup reachable from a case
body (`GetStaticTypeOf`, `GetSlot`, `GetPathForName`, `GetIsConst`,
`defineOrDecl`) would need to apply the mapping, and each would need to know
whether the shadow is in scope yet. Strictly more surface than last-wins, for
the same result.

**Leave the single slot and only relax the type check.** Would silence the
panic but not the clobbering, and could not represent both bindings being live
at different points in the same case.

## Consequences

- Case blocks now differ in name count more often, so this **depends on the
  `fallthrough` block-shrinkage fix** (`b.Values` truncation in `doOpExec`),
  which is included in the same PR. Without it, `case 1: v := 99; fallthrough`
  panics with `unexpected block size shrinkage`. `tests/files/switch53.gno`
  covers the combination.

  That truncation composes with the block pool added in #5813: `ExpandWith`
  now grows through `growBlockValues`, which reslices within capacity, and
  then writes every index in `[oldNames, numNames)`. Since the truncation only
  lowers `oldNames`, every slot it exposed is overwritten before use, and no
  recycled value survives into the next clause.

  `ExpandWith` also reassigns `b.Source` *before* its equal-size early return:
  falling through into a clause that declares no names of its own would
  otherwise leave `b.Source` on the fallen-from clause, making the debugger's
  `print` resolve that clause's (dropped) names against the truncated block —
  an index-out-of-range error instead of "could not find symbol".

- `GetLocalIndex` is a hot path and changed iteration direction. Duplicates
  remain impossible outside faux case blocks, so results are unchanged
  everywhere else. Cost was measured rather than assumed: `nameIndexThreshold`
  (32) sends every wide block down the map branch, so the linear scan runs on
  blocks averaging 3.06 names. Instrumenting a full `TestFiles` run counted
  18.4M name comparisons forward vs 19.4M backward (+5.3%; misses, which are
  direction-neutral, are 2.4× hits). Replaying the measured hit-distance
  histogram in a microbenchmark put the difference at 1.16 ns per lookup, or
  about 2.9 ms across a 48 s suite. Removing the `isLocallyReserved` pre-check,
  which duplicated a `GetLocalIndex` that `Reserve` performed anyway, saves a
  full lookup per `:=` LHS name on the same path and more than covers it.

- The invariant that only faux case blocks hold duplicates is enforced, not
  just documented: `defineNew` is the sole append path, and a `debugAssert`
  check there panics on any duplicate at or past the boundary. A full
  `-tags debugAssert` filetest run does not trip it, and its failure set is
  identical to the base commit's apart from the three new tests, which the base
  fails for lack of the fix.

- Heap-item marking and closure capture follow the same slots: `SetIsHeapItem`
  and `addHeapCapture` reach the block via `GetBlockNodeForPath`, which is
  path-based, so a capture of the outer name lands on the parent's slot and a
  capture of the shadow lands on the case's. `switch53.gno` asserts a closure
  made before the shadowing declaration still reads the outer value.

- `getLocalIsConst` remains name-based (`slices.Contains(sb.Consts, n)`), which
  is not index-parallel, so a `const v` in a case marks the name const for both
  slots. For a use of the outer `v` textually *before* the `const v`, the
  name-keyed `GetIsConst` therefore reports true and the use is const-folded
  against the copy's static (value-less) slot, printing its zero value instead
  of the outer value; an assignment there is rejected with "cannot assign to
  const". This divergence is **not introduced here**: an ordinary nested block
  (`v := 1; { println(v); const v = "c" }`) misbehaves identically on master,
  because the shadow's `Consts` entry exists from `initStaticBlocks` on while
  `GetIsConst` has no notion of statement position. The faux-block fix brings
  case bodies to parity with that pre-existing behavior; fixing it for both
  block kinds is left to a follow-up.

- One behavior change outside the compiler: the debugger's `print v`, stopped
  in a case body *before* the shadowing declaration, now reports the shadow's
  slot (zero value) rather than the outer name. `tryGetPathForName` resolves by
  name with no notion of statement position, so it inherits last-wins
  unconditionally. This is the one place where "last slot" and "what is in
  scope here" differ.

## Tests

- `tests/files/switch52.gno` — `:=`, `var` and `const` shadows, use before the
  shadow, and a nested switch shadowing again.
- `tests/files/switch53.gno` — closure capturing the outer name before the
  shadow, closure capturing the shadow, and shadow + `fallthrough`.
- `tests/files/if9.gno` — if-init shadowing in both branches.

Each was run under real `go run` first; the `// Output:` blocks are Go's actual
output.
