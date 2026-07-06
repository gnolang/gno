# ADR: Persist function-local declared types referenced by saved values

## Context

Package-level declared types are written to the type store (`/t/<TypeID>`) at
addpkg (`saveNewPackageValuesAndTypes`). Function-local declared types
(`type S ...` inside a function body) are created at runtime and were never
`SetType`'d. Any persisted `TypedValue` whose `.T` is such a type serializes
as `RefType{"pkg[loc].Name"}` — a dangling pointer into the type store. The
same process never notices (the live type sits in `cacheTypes`); after a node
restart, loading the object hits `fillTypesOfValue` → `GetType` → miss →
`panic("unexpected type with id ...")`. The state is permanently unreadable.

Escape routes that reproduce on master: an interface-typed package var
(`X = S{...}`) and a closure capture (heap-item slot typed `S`). A third
route — interface-bound method values (`G = i.Get`) — does *not* reproduce on
master because the eager bind resolves promotion at bind time and persists
only the embedded package-level receiver; it starts carrying the local type
once #5737's call-time dispatch lands (which persists the interface-boxed
operand). That makes this PR a prerequisite for #5737, but the fix is a
standalone correction of live corruption.

## Decision

1. **`localTypeSaver` walk in `saveObject`** (realm.go): before `SetObject`,
   walk the object's typed slots and `SetType` every reachable `DeclaredType`
   with a non-zero `ParentLoc`, recursing through `Base` (visited-guarded for
   recursive types). `saveObject` is the single choke point all persistence
   routes go through (tx finalize and addpkg both reach it via
   `FinalizeRealmTransaction`), so the walk covers every route. Dedup is
   centralized in `SetType`'s per-tx `cacheTypes` early-return; the walk stays
   stateless apart from cycle protection. Per-slot cost is a few pointer hops:
   package-level `DeclaredType` prunes immediately without descending `Base`,
   and `SetObject` already performs two full structural walks
   (`copyValueWithRefs`, amino marshal), so this shallower third walk is noise.
2. **`copyTypeWithRefs` preserves `ParentLoc`**: `ParentLoc` is part of the
   TypeID for local types (`pkg[loc].Name`); dropping it in the persist copy
   (as before) would store the type record under a different ID than the one
   values reference.
3. **`debugAssert` invariant in `SetObject`** (store.go): walk the
   persist-copy; a bracketed `RefType` not in `cacheTypes` panics at save
   time, so a future missed route fails loudly inside the (buffered,
   rolled-back) transaction instead of committing unreadable state. The type
   walker panics on unknown type kinds for the same reason; known-but-not-
   currently-persistable kinds (`tupleType`) are walked, structurally-empty
   kinds (`blockType` etc.) are pruned.

## Alternatives considered

- **`SetType` at declaration time (`OpTypeDecl`)**: persists types whose
  values never escape — store bloat and gas for dead types; also detaches the
  write from the save path that defines what actually needs resolving.
- **`containsLocal` bit propagated at type construction** (compiler-style
  `HasTParam` pattern): O(1) save-time check, but every Type construction
  site (preprocess + runtime) must participate and the bit must survive amino
  round trips. The saver's query is cold (once per dirty object per tx,
  dwarfed by amino encode), so the walk wins on blast radius.
- **Resolve lazily on reload**: impossible — after restart nothing can
  reconstruct the type; the source of truth is gone.

## Consequences

- Saves reaching a local type now write `/t/` entries and pay their encode
  gas; type-record bytes gain `ParentLoc`. Deterministic, but a state/gas
  change — coordinate like other consensus-affecting fixes.
- `BoundMethodValue.Receiver` in the saver and the lt1 txtar case are
  currently vacuous (eager bind flattens the receiver); they become live with
  #5737.
- Tests: `restart_local_type*.txtar` are the true reproducers (fail on
  master); `zrealm_localtype0/1.gno` filetests pass on master and act as
  save-side guards via `-tags debugAssert` (`make test.debugAssert`, not yet
  in CI) plus a golden pinning the on-the-wire bracketed `RefType`.
