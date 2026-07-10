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
(`X = S{...}`) and a closure capture (heap-item slot typed `S`). A third
route — interface-bound method values (`G = i.Get`) — does *not* reproduce on
master because the eager bind resolves promotion at bind time and persists
only the embedded package-level receiver; it starts carrying the local type
once #5737's call-time dispatch lands (which persists the interface-boxed
operand). That makes this PR a prerequisite for #5737, but the fix is a
standalone correction of live corruption.

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
   already hold local-typed values at addpkg-save time.
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
   their per-tx `cacheTypes`. The type walker panics on unknown type kinds
   for the same reason; known-but-not-currently-persistable kinds
   (`tupleType`) are walked, structurally-empty kinds (`blockType` etc.) are
   pruned.

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
- The lt1 txtar case (`BoundMethodValue.Receiver`) is currently vacuous
  (eager bind flattens the receiver); it becomes live with #5737.
- Tests: `restart_local_type*.txtar` are the true reproducers (fail on
  master); `zrealm_localtype0/1/2.gno` filetests pass on master and act as
  save-side guards via `-tags debugAssert` (`make test.debugAssert`, not yet
  in CI) plus a golden pinning the on-the-wire bracketed `RefType`.
