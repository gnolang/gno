# Order-independent type-cycle checks and deferred map-key validation

## Context

PR #6048 / issue #6036: Gno rejected some legal cyclic type declarations
(especially those involving aliases) and accepted some illegal ones depending
on declaration / visit order. Cycle detection lived inside the dependency walk
(`findUndefinedAny`), which only reports a cycle for a name it is currently
defining, so a direct containment edge could be missed once an earlier
indirect edge had already caused that name to be defined.

A follow-on map-key check (`assertValidMapKeys`) was added at
`*TypeDecl:LEAVE` so uncomparable keys are rejected once the declared type is
filled in. The first version of that check treated a key as "settled" when a
top-level `*DeclaredType` had a non-nil `Base`. That is not enough: a key
struct can already be filled while a nested field's `*DeclaredType` still has
a nil `Base`. `isComparable` treats a nil base as uncomparable and memoizes
the result on `*StructType`, so valid Go such as

```go
type D X
type X [2]*Z
type Z map[S]int
type S struct{ d D }
```

was rejected with `invalid map key type main.S`, and `a == b` on `S` could
see a poisoned memo.

## Decision

1. **Containment cycles** stay checked in `assertNoDirectTypeCycle` at
   `*TypeDecl:LEAVE` on the finished type: walk defined-type bases, struct
   fields and array elements; pointers / slices / maps / funcs / interfaces
   end the walk. Verdict no longer depends on resolution order.

2. **Alias / indirection cycles** in `findUndefinedAny` match go/types: a
   cycle among type decls is legal only when it passes through both an
   indirection and a defined type; alias-only cycles stay invalid. The check
   applies only to type declarations so value names that merely shadow a
   type are unaffected.

3. **Map-key settledness is deep.** `typeIsSettled` walks the same graph
   `isComparable` will walk and returns false if any nested `*DeclaredType`
   still has a nil `Base`. Unsettled keys are skipped; a later
   `*TypeDecl:LEAVE` that walks back into the map re-checks once the graph
   is complete. This mirrors go/types deferring map-key comparability to a
   later pass, without collecting a separate pending list.

## Alternatives considered

- **Collect pending map types in `PredefineFileSet` and check after the
  type-decl loop.** Equivalent end state; deferred until a later LEAVE walk
  reaches the map is enough with the deep settled check and keeps the check
  local to `assertValidMapKeys`.
- **Split map-key validation into its own PR.** Rejected for this follow-up:
  the false reject was introduced with the cycle work and blocks merge.
- **Wrap type aliases in `*DeclaredType` so errors name `K` instead of the
  underlying `struct{...}`.** Pre-existing alias representation (`Define2`
  stores the RHS type directly for `type K = T`). Out of scope here; note as
  a nit / follow-up.

## Consequences

- Legal pointer-broken cycles used as map keys (and `==` on those structs)
  match Go again; regression filetests `maptype2*.gno`.
- Some legal alias cycles still fail with internal panics
  (`name not defined` / `should not happen`) because of an early return in
  the `*NameExpr` branch of `findUndefinedAny` when a name is in `defining`
  but has no slot yet. Tracked as a follow-up issue; not a merge blocker for
  the false map-key reject.
- Map-key errors for `type K = struct{...}` may still spell the underlying
  struct rather than `K` (pre-existing).
