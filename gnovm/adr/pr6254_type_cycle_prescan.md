# Validate type-declaration cycles before predefinition

## Context

Issue #6036: gno rejected legal cyclic type declarations that pass through
aliases, and its verdict on some others depended on declaration order.
The Go rule (golang/go#25838): a cycle among type declarations is invalid
when every edge on it is a direct containment, which gives the type
infinite size, or when every declaration on it is an alias, which leaves
nothing to resolve to. Every other cycle is legal.

Cycle detection lived inside the dependency walk (`findUndefinedAny`).
That walk resolves one declaration at a time, stops at the first undefined
name, and forgets a name once it has a slot, so it only ever sees one path
through the graph. PR #6048 kept the walk's verdict and added a second
walk over the constructed types at `*TypeDecl` LEAVE for value cycles,
mirroring go/types (`cycleError` plus `validType`). That split exists in
go/types because it resolves declarations lazily; gno holds the whole
declaration set in `PredefineFileSet` before anything is built.

## Decision

Validate the declaration graph first, then predefine.

1. **`typedecl_group.go`** collects, for the type declarations of a group, the
   names each one refers to with a direct flag (struct field, array
   element, embedded interface and bare name are direct; pointer, slice,
   map, chan, func and method signature are not). Names the group does not
   declare are leaves.
2. **Two plain cycle searches** implement the rule: one over the direct
   edges (infinite size), one over the edges between aliases (alias only).
   A single DFS judging the cycle it happens to close is wrong when a legal
   and an invalid cycle share nodes; see `TestTypeDeclInvalidCycle`.
3. **Call sites**: `PredefineFileSet` for the package, the single-file
   fallback in `Preprocess` (a no-op once the file set is predefined), and
   each type declaration inside a function body. A cycle is reported with
   its earliest-declared member first and located at that declaration, so
   the message is a property of the cycle, not of the search order.
4. **Every declaration in a group gets its slot before any is predefined**
   (`reserveTypeDecls`): a shell of its kind, or for an alias of a group
   name, the slot its chain ends at. A legal cycle can then close on any
   member whichever declaration is predefined first. The one ordering rule
   left in `tryPredefine` is that a directly contained type (base, array
   element, struct field, embedded interface) is built before its
   container, since sealing looks through it; it reuses `collectTypeDeps`,
   so direct and indirect are classified in one place.
5. **The walk carries no verdict.** Its `direct` flag, the LEAVE check and
   the alias-chain resolver are gone.
6. **Map-key comparability and embed depth** are checked once per group
   after its types are built (`checkBuiltTypeDecl`), since a member
   reached through a pointer may still have a nil base when its container
   is sealed. `Seal` no longer checks embed depth; every key and base is
   settled by then, so the map check cannot poison the `comparable` memo.

7. **Gas.** The walks run outside the op loop, where only the flat
   `PreprocessGasPerByte` applies, so they bill the tx's preprocess meter
   per unit of work through `chargeCPUGas`, as `embedWalk` does: per
   declaration, per type-expression node (dependency scan, map-key walk)
   and per graph step (cycle-search edge, `embedDepth` visit). Slopes are
   dev-box fits from `BenchmarkTypeDeclGroup` converted with a machine
   factor measured the way #6164 did it, gas table divided by measured
   pure ns over seven flat ops run alongside (1.9 to 2.2 on the dev box;
   2.1 used), rounded up so every grid shape is a floor. The run is
   recorded in `cmd/calibrate/typedeclgroup_bench_m5_arm64.txt`; a re-fit
   on the reference hardware is a follow-up, as for #6164. Only
   `embedDepth` can exceed linear in the source: a subgraph shared by many
   embedders is re-walked once per declaration, which `embed_gas` pins.

## Alternatives considered

- **Keep the two-verdict design of #6048.** Works and matches go/types, but
  the verdict is split across two places, the walk needs the `direct`
  plumbing only for it, and an invalid alias cycle must be caught before
  construction or `TypeID` overflows the Go stack.
- **Extend the walk to see the whole graph.** It would have to record all
  edges of all declarations, which is this pre-scan under another name.

## Consequences

- Invalid cycles are reported before any type is built, always with the
  full path (`A -> B -> A`). Goldens for `recursive9i` and `recursive9j`
  changed from `refers to itself` to the path form.
- `findUndefinedV/T/Any` and `tryPredefine` lose the `direct`, `stack` and
  `defining` parameters and the `directR` result: the walk only returns the
  first undefined name, which can now only be a value. `tryPredefine`
  builds nothing: it only resolves an alias of a name outside the group,
  which reservation cannot bind before imports are predefined. `predefineRecursively2` keeps `stack` and
  `defining`: besides value cycles they catch the one type cycle the
  pre-scan cannot see, one that passes through an array-length constant
  (`recursive16`, `recursive17`). Outside a type declaration's own
  expression a reserved type counts as undefined until built, so a value
  or variable type that names it (`const N = T(3)` reached from an array
  length, `len(T{})`, `var x [3]T` behind `len(x)`) builds it first
  (`decltype_constlen*`).
- `embed_depth1` pins the post-group embed-depth check with a chain
  declared top-down, which `Seal` alone accepted. `typedecl_group_gas_test.go`
  pins each charge with a budget that only the metered walk exceeds.
- Runtime construction of an inline struct or interface type
  (`doOpStructType`, `doOpInterfaceType`) now bills its embed-depth walk
  too, `OpCPUSlopeTypeDeclStep` per type visited or field scanned; eleven `gas/` goldens
  moved by that amount.
- Forward references among function-local type declarations are invalid
  Go and still fail with `not defined in fileset`, as before.
- Bare-name references into a cycle (`type A B` with B pointing back
  through an indirection, aliases of a member referenced from the member)
  no longer die with `should not happen` or `name not defined`; fixtures
  `recursive14`, `recursive15`. A differential harness of 373 generated
  declaration groups agrees with go/types on every case except channel
  types, which gno rejects outright.
