# ADR: Persist function-local declared types referenced by saved values

## Context

Package-level declared types are written to the type store (`/t/<TypeID>`) at
addpkg (`saveNewPackageValuesAndTypes`). Function-local declared types
(`type S ...` inside a function body) never were. Any persisted `TypedValue`
whose `.T` is such a type serializes as `RefType{"pkg[loc].Name"}` — a
dangling pointer into the type store. The same process never notices (the
live type sits in `cacheTypes`); after a node restart, loading the object
hits `fillTypesOfValue` → `GetType` → miss →
`panic("unexpected type with id ...")`. The state is permanently unreadable.

Escape routes that reproduce on master: an interface-typed package var
(`X = S{...}`), a closure capture (heap-item slot typed `S`), and — since
#5737's call-time dispatch landed (68111d9e6), which persists the
interface-boxed operand — interface-bound method values (`G = i.Get`).
Before #5737 the third route was inert (eager bind persisted only the
embedded package-level receiver); with it on master, all three corruption
routes are live today.

## Decision

1. **Eager persistence at addpkg (`saveFuncLocalTypes`, machine.go)**:
   `saveNewPackageValuesAndTypes` walks the package's fileset AST for
   `*TypeDecl` nodes (function bodies, closures, nested blocks included) and
   `SetType`s every non-alias `DeclaredType` with `IsFuncLocal()`. Local
   `DeclaredType`s are materialized at preprocess time (`declareWith`), so
   the AST enumerates them completely and no `Base` recursion is needed: any
   local type reachable from another's `Base` is itself a `*TypeDecl`. This
   mirrors how package-level types are persisted — the entire type-storage
   cost lands at addpkg with the deployer, and transaction saves stay free
   of type writes and of any per-save traversal. The walk runs *before*
   `FinalizeRealmTransaction` because file-level var initializers may
   already hold local-typed values at addpkg-save time (pinned by
   `zrealm_localtype3.gno` under debugAssert). Its preconditions — live
   `*Block`, `*PackageNode` source, attached fileset — are invariants,
   not fallbacks: they panic if broken, because a silent skip would
   recreate the dangling-ref corruption.
2. **`copyTypeWithRefs` preserves `ParentLoc`**: `ParentLoc` is part of the
   TypeID for local types (`pkg[loc].Name`); dropping it in the persist copy
   (as before) would store the type record under a different ID than the one
   values reference.
3. **`debugAssert` invariant in `SetObject`** (store.go): walk the
   persist-copy; a bracketed `RefType` that is neither in `cacheTypes` nor
   in the backend type store panics at save time, so a missed declaration
   route fails loudly inside the (buffered, rolled-back) transaction instead
   of committing unreadable state. The backend probe is a raw key check —
   later transactions see addpkg-persisted types in the backend, not in
   their per-tx `cacheTypes` — with a nil `GasContext`, so debugAssert
   builds consume exactly the same gas as release builds. Both the value
   and type walkers panic on unhandled kinds for the same fail-loud
   reason; known-but-not-currently-persistable kinds (`tupleType`) are
   walked, structurally-empty kinds (`blockType` etc.) are pruned
   explicitly, and anything else — including `ChanType`, which
   `copyTypeWithRefs` never emits — hits the panic default.

MsgRun scripts never reach `saveNewPackageValuesAndTypes` (the keeper runs
them with `save=false`), and a pre-existing guard ("cannot persist object of
type defined in the private realm") independently rejects their values
escaping into realm state, so no ephemeral-package types are persisted.

## Alternatives considered

- **Save-time walk (`localTypeSaver` in `saveObject`)** — the first
  implementation of this PR: walk each to-be-persisted object's typed slots
  and `SetType` reachable local types on demand. Covers every route through
  the single `saveObject` choke point and — unlike the eager walk —
  retroactively heals packages deployed *before* the fix (their next save
  writes the missing type record). Rejected per review: it persists types
  on-demand at an unpredictable payer (whichever tx first escapes a value)
  and re-walks every saved object forever, while eager persistence pays once
  at addpkg like package-level types already do. The retroactivity advantage
  is moot if this lands before packages with escaping local types exist
  on-chain; otherwise a one-shot state migration (or temporarily keeping the
  saver as backstop) is required.
- **Predefine-time collection** (#6084, maintained as a side-by-side
  alternative): `tryPredefine` records each minted local type on the
  PackageNode (`ATTR_FUNC_LOCAL_TYPES`) and the save iterates the list —
  no save-time traversal, symmetric with the package-level loop, at the
  cost of a preprocess bookkeeping invariant (audited under debugAssert
  there). This PR keeps the walk: completeness rests on a language fact —
  every local type is a `*TypeDecl` in the fileset — rather than on
  bookkeeping.
- **`SetType` at declaration time (`OpTypeDecl`)**: runtime re-execution
  pays gas on every call of the declaring function and would fire inside
  MsgRun scripts; the static enumeration at addpkg has neither problem.
- **Resolve lazily on reload**: impossible — after restart nothing can
  reconstruct the type; the source of truth is gone.

## Consequences

- Addpkg now writes `/t/` entries for every function-local type in the
  package — including types whose values never escape (bounded store bloat,
  matching the existing behavior for package-level types that are never
  referenced). Deterministic, but a state/gas change — coordinate like other
  consensus-affecting fixes.
- Packages added before this change never had their local types persisted;
  values of those types saved *after* the upgrade still produce dangling
  refs. Deployment must either predate any such package (genesis) or include
  a migration that re-runs local-type enumeration over stored packages.
  Since #5737 is already on master, any chain state created on post-#5737
  code before this fix can already contain dangling method-value refs —
  the "fresh chain" caveat holds only if no network deploys master in the
  gap.
- Tests: `restart_local_type.txtar` is the true reproducer (fails on
  master — since #5737, the lt1 method-value route fails first); its
  `zlti` realm covers the addpkg-time escape (file-level var initializer)
  across a restart. `zrealm_localtype0/1/2/3.gno` filetests pass on master
  and act as save-side guards via `-tags debugAssert`
  (`make test.debugAssert`, not yet in CI) plus a golden pinning the
  on-the-wire bracketed `RefType`; `zrealm_localtype3.gno` additionally
  pins the save-before-finalization ordering (reordering fails under
  debugAssert).
