# Bound Go2Gno copies before conversion

## Status

Implemented. Supersedes the inherited-const node budget of
`pr6504_go2gno_const_inheritance_bound.md`.

## Context

`Go2Gno` copies source subtrees in two places:

- `toFields` converts `a, b T` into one `FieldTypeExpr` per name, and gives each
  its own converted copy of `T`.
- `toDecls` gives every bare spec in a const block (`c1` in
  `const ( c0 T = x; c1 )`) its own copy of the last type and values.

Copies nest and multiply. With `k` names per level over `d` levels, a ~450
byte package converts to 10^8 fields. A 16.9KB const block whose first value
sums 2000 ones converts to 8M nodes (795MB). Parse gas is charged per source
token, and `Go2Gno` does not use the VM allocator, so none of that work is
metered. `MsgAddPackage`, `MsgRun` and the unauthenticated `vm/qeval` query all
reach it.

The first attempt shared the converted `T` across names and memoized shared
nodes in `Transcribe`. Review showed that this only moves the blowup. Passes
outside `Transcribe` still walk each name separately. `findUndefinedAny` in
predefine takes 31.7s at depth 30 with 2 names. `StructType.TypeID` during
`SaveBlockNodes` builds a 10^8 field string and runs out of memory at depth 8,
which is within `MaxTypeDepth` because that limit bounds depth, not width.
`toDecls` already shared the last const type the same way, so one `Transcribe`
walk visited it once per spec.

## Decision

Keep every copy, and reject the source before conversion if the copies would
add more nodes than it has bytes. `toDecls` copies the values into every bare
spec, since preprocess rewrites them per spec. The last const type stays shared
across bare specs, as it already was: `constCopies` counts it once per bare
spec, so a walk over the shared node is bounded by the same count.

`copiedNodes` counts the added nodes on the `go/ast` in one walk. A field with
`k` names is one `Field` with its names, type and tag, and becomes `k`
`FieldTypeExpr`s, each with a name, a converted copy of the type and the tag.
An unnamed field also gains an empty name. `constCopies` follows `toDecls`: a
bare spec adds the converted size of the last type and values, and a spec whose
name count differs from the last value count drops the values for good. Copies
of copies are counted at their converted size, so nesting multiplies in the
count as it does in the conversion. If the count exceeds `maxCopiesPerByte` (1)
nodes per source byte, the source is rejected.

The count only grows, so nothing else in the source can offset it. Comments are
not converted and are not counted. The `go/ast` holds a single copy of each
subtree, so the walk is linear. The added count is clamped just above the limit
at every level. With a 64-bit `int` the products before the clamp stay far from
overflow. Only 64-bit builds are supported.

The check runs where untrusted source enters:

- `TypeCheckMemPackage`, on each file of the checked package after Go parsing
  and before `go/types`. `MsgAddPackage`, `MsgRun` and `MsgEnablePackage` all
  type check before they call `ParseFile`. Imported packages are not checked.
- `ParseExpr` and `ParseStmts`, which `vm/qeval` and the REPL use directly.

`ParseFile` does not check. At boot, `VMKeeper.Initialize` re-parses every
stored package through it, and a package stored before this bound may exceed
it. Genesis replay sets `TypeCheckOptions.LegacyCopies` for the same reason: a
fork replays its source chain's history rather than admitting new code.

This replaces the inherited-const budget that
`pr6504_go2gno_const_inheritance_bound.md` added to `toDecls` (twice the
declaration's bytes, counted in Gno nodes). That budget ran inside `ParseFile`,
so it rejected stored packages at boot: a 616 byte
`const ( c0 = 1+...+1 (50 terms); c1 ... c99 )` that master accepts made
`PreprocessAllFilesAndSaveBlockNodes` panic. It also ran after `go/types`, which
checks an inherited value once per spec and took 21.7s on a 71KB block.
`checkCopies` covers inherited values before `go/types` runs, so the budget is
removed. That change's refusal to inherit a const value containing a function literal
stays: `checkCopies` counts nodes, not bytes, and a function literal is the only
way to put block nodes, several times larger than any expression node, into a
const. `go/types` already rejects one in a const, so the refusal only fires on
the `ParseExpr` and `ParseStmts` paths.

The bound scales with the source. A fixed bound per file was the first version
of this check, and a package has no file count limit: a 241 byte file could add
57272 nodes (about 4.9MB retained), and a 1MB transaction could carry about
3700 such files for about 18GB, at about 1.25B gas under the 3B block limit.
With the bound per byte, the nodes added across all files of a transaction are
at most its size in bytes.

## Alternatives considered

- Share the converted type across names (the first attempt). The gno AST becomes a DAG, and every pass that walks it per
  field (predefine, `TypeID`, the sealer) stays exponential unless each one is
  memoized. Sharing const values is also unsafe: preprocess reads `ATTR_IOTA`
  per declaration and rewrites the values in place.
- Check in `ParseFile`. It rejects stored packages at boot, where
  `Initialize` has no recover, so a node could not start on its own state.
  The inherited-const budget did this; see the Decision.
- Check in `ParseFile` with a legacy option for boot and replay. It keeps
  future callers covered by default, but threads an option through every
  machine. Every path that admits code already type checks first.
- A fixed bound per file. Splitting the source across files multiplies it, as
  above.
- Bound depth only. `MaxTypeDepth` already does this, and width alone
  reaches 10^8 within it.
- Charge gas per converted node. The work happens before the gas meter covers
  later passes, and the size is known up front anyway.

## Consequences

- Source whose copies add more nodes than it has bytes no longer type checks.
  Across the 4578 `.gno` files in `examples/`, `gnovm/stdlibs`, `gnovm/tests`,
  `contribs` and the integration testdata, the highest ratio is 0.1515
  (`examples/quarantined/gno.land/p/moul/printfdebugging/color.gno`, an iota
  const block adding 158 nodes for 1043 bytes), so the bound leaves about 6x
  headroom. Short sources
  are the closest to it, because the bound is per source: the 53 byte
  expression `struct{ a, b, c, d, e, f, g, h struct{ x, y int } }{}` passed
  alone to `vm/qeval` is rejected, while `struct{ a, b struct{ x, y int } }{}`
  is not.
- A transaction's added nodes are bounded by its size. At the measured ~86
  bytes per added node, a 1MB transaction adds at most about 86MB.
- `go/types` also checks an inherited const value once per spec, which took
  2.27s for a 35KB block. The check runs first, and rejects it in 0.6ms.
- Type parameter lists are counted although `Go2Gno` does not convert them,
  which can only reject more, never less.
- A new path that calls `ParseFile` on untrusted source must type check it
  first, or call `checkCopies` itself.
- `copiedNodes` restates the copy rules of `toFields` and `toDecls`. A new copy
  in `Go2Gno` must be added to it. `TestCopiedNodesTracksGo2Gno` compares the
  count with the converted AST over synthetic sources and the `.gno` corpus,
  so it catches a missed copy that one of them repeats.
- Packages stored before the bound keep loading at boot and in genesis replay
  unchecked. Each node already converted them once when they were added.
- `TypeID` growth is bounded by `pr6504_type_id_bound.md`.
- The error is a tm2 `errors.New` with no args. The keeper's bounded recover
  path, used by `vm/qeval`, prints only whitelisted error types, and renders a
  plain `fmt.Errorf` as `<error: *errors.errorString>`.
