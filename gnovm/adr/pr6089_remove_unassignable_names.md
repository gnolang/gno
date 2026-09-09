# ADR: Derive `IsAssignableName` from `NameSources`, drop `StaticBlock.UnassignableNames`

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
explicit branch inside `IsAssignableName` (formerly `IsAssignable`;
renamed to avoid confusion with type assignability à la
`checkAssignableTo`).

That `Reserve` call already records the same fact in
`StaticBlock.NameSources`: the entry at the name's local index carries
`Type == NSFuncDecl`. So `UnassignableNames` duplicated, in a second
serialized field, information the block already persists, and the
duplicate invited a wrong reading — "every unassignable name is in this
list" — which is false.

## Decision

- Delete the `UnassignableNames` field. `IsAssignableName` now answers from
  the block that declares the name: `NameSources[idx].Type != NSFuncDecl`,
  an O(1) lookup instead of an O(n) scan. Indexing `NameSources` by a
  `GetLocalIndex` result is safe: `Define2` panics unless
  `NumNames == len(NameSources)`, and the same unguarded idiom is
  already used for `NSTypeDecl` lookups in `preprocess.go`.
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

- One less field to keep in sync with `NameSources`, and an O(1) check
  instead of an O(n) scan in the preprocessor's assign check.
- One less field in the amino/proto schema for `StaticBlock`. Block
  nodes are not persisted to the store backend today
  (`SetBlockNode`'s backend write is a TODO), so this is schema
  hygiene rather than a live migration; the wire suites (amino,
  `-run Gas`, `TestTestdata`) all pass unchanged.
- Like the `Externs` removal, the reserved slot must stay in place;
  amino field removal remains order-brittle.
- Possible follow-ups, out of scope here: an `IsAssignableNameAt(store,
  path)` fast path (the `AssignStmt` call site already has a resolved
  `ValuePath`, mirroring `GetIsConstAt`), and merging `Consts` into
  `NameSources` the same way, retiring the last parallel name-list.
