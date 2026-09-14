# ADR: Derive `IsAssignableNameAt` from `NameSources`, drop `StaticBlock.UnassignableNames`

## Context

`StaticBlock.UnassignableNames` was introduced by #3198 to let the
preprocessor reject assignments to package-level function names
(`func f(){}; f = nil`). Despite the general-sounding name, the slice
only ever had one writer: the non-method `*FuncDecl` case in
`initStaticBlocks`, which appended the func decl's name right after
calling `Reserve(false, nx, n, NSFuncDecl, -1)` on the same package
block. Every other kind of unassignable name is handled elsewhere —
constants are folded to `ConstExpr` (and tracked in `Consts`), type
names are folded to `constTypeExpr`, and uverse names are refused by an
explicit branch inside `IsAssignableNameAt` (formerly `IsAssignable`;
renamed to avoid confusion with type assignability à la
`checkAssignableTo`).

The name-keyed walk had a second defect, exposed by #6060: it stops at
the first block holding a slot for the name, which for an assignment
placed *before* a shadowing declaration in the same block is the shadow's
reserved slot, not the outer binding the assignment actually targets.
`f = func(){}; f := 1` inside a block therefore overwrote the package-level
`f`. The `const f`/`type f` spellings were rejected on master only by the
name-keyed const check, which #6060 makes path-aware for the same reason.

That `Reserve` call already records the same fact in
`StaticBlock.NameSources`: the entry at the name's local index carries
`Type == NSFuncDecl`. So `UnassignableNames` duplicated, in a second
serialized field, information the block already persists, and the
duplicate invited a wrong reading — "every unassignable name is in this
list" — which is false.

## Decision

- Delete the `UnassignableNames` field. `IsAssignableNameAt(store, path)`
  answers from the block the NameExpr's already-resolved path names
  (`GetBlockNodeForPath`, as `GetIsConstAt` does):
  `NameSources[path.Index].Type != NSFuncDecl`. Indexing by the path
  index is safe: `Define2` panics unless `NumNames == len(NameSources)`.
- Move the check from an `AssignStmt`-only loop in `preprocess.go` into
  `assertValidAssignLhs` (`type_check.go`), after its blank/uverse/const
  branches. That gate is shared by assignments, inc/dec and range
  clauses, so `for _, f = range ...` into a package func is now rejected
  too (`tests/files/assign43.gno`); on master it ran. The error reads
  `cannot assign to func f`, alongside the existing `cannot assign to
  const`. Position-correct: an assignment before a shadowing declaration
  checks the outer binding (`tests/files/assign42.gno`).
- Retire amino field 8 with a blank `_ struct{} `amino:"reserved"``
  field — the mechanism introduced for `Externs` (field 10) in #5301.
  Field numbers are unchanged; decoders skip field 8 if present in old
  encoded data. `gnolang.proto` and `pb3_gen.go` regenerated with
  `misc/genproto2`.

## Alternatives considered

- **Rename to `FuncDeclNames`**: fixes the misleading name but keeps
  duplicate serialized state, and a per-block field that is only ever
  non-empty on `PackageNode` stays awkward under any name.
- **Make the name true** (fold consts/types/uverse into the list): adds
  state for facts that already have cheaper representations.

## Consequences

- One less field to keep in sync with `NameSources`, and a path-keyed
  check in the shared LHS gate instead of a name scan on one statement
  kind. The shadow-before-assignment hole is closed for all three
  spellings once #6060 lands (verified on a merge of both branches).
- #6060 edits the adjacent const branch of `assertValidAssignLhs`
  (`GetIsConst` to `GetIsConstAt`); the two merge cleanly in either
  order, verified with #6060's shadow tests and the probes above.
- One less field in the amino/proto schema for `StaticBlock`. Block
  nodes are not persisted to the store backend today
  (`SetBlockNode`'s backend write is a TODO), so this is schema
  hygiene rather than a live migration; the wire suites (amino,
  `-run Gas`, `TestTestdata`) all pass unchanged.
- Like the `Externs` removal, the reserved slot must stay in place;
  amino field removal remains order-brittle.
- Possible follow-up, out of scope here: merging `Consts` into
  `NameSources` the same way, retiring the last parallel name-list.
