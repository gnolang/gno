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

1. **Eager persistence at addpkg, enumerated at predefine time**: local
   `DeclaredType`s are minted in exactly one place (`tryPredefine` ->
   `declareWith`; the TRANS_LEAVE `*TypeDecl` case copies *into* that
   instance), so `tryPredefine` records every function-local type (blank
   decls never mint one — the early return precedes `declareWith`) on the
   `PackageNode` as `ATTR_FUNC_LOCAL_TYPES` — the attribute
   bag is the established home for derived, non-serialized preprocess
   products (same machinery as `ATTR_REF_ELEM_TYPE`); the typed accessor
   pair `FuncLocalTypes`/`AddFuncLocalType` keeps call sites cast-free.
   `saveFuncLocalTypes` then just iterates the list and `SetType`s —
   fully symmetric with the package-level loop, which likewise iterates
   state (block slots) that predefine populated; no save-time AST
   traversal on the main path. The entire type-storage cost lands at
   addpkg with the deployer, and transaction saves stay free of type
   writes. The save runs *before* `FinalizeRealmTransaction` because
   file-level var initializers may already hold local-typed values at
   addpkg-save time (pinned by `zrealm_localtype3.gno` under debugAssert);
   its preconditions (live `*Block`, `*PackageNode` source) are invariants
   and panic if broken. Completeness invariant (documented on
   `AddFuncLocalType`): every code path that mints a function-local
   DeclaredType must append. It is machine-checked, not trusted:
   `assertFuncLocalTypesComplete` (debugAssert) re-runs the fileset AST
   walk at addpkg and panics on any disagreement with the collection,
   including duplicates (unit-tested by
   `TestAssertFuncLocalTypesCompleteFires`); the `SetObject` assert below
   is the second, independent tripwire. Preprocess contexts that never
   save (MsgRun, queries, boot re-preprocess) rebuild the attribute
   harmlessly without reading it.
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

**Why local types save before finalization but package-level types after —
and why the assert is scoped to function-local refs.** A declared type's
persisted record embeds its methods' FuncValue hashes (`copyTypeWithRefs` →
`copyMethods` → `toRefValue`), and those hashes exist only once
`FinalizeRealmTransaction` has saved the objects; objects in turn embed type
refs. The codebase breaks this object↔type cycle by writing objects first,
package-level type records second — so within the addpkg tx, refs to the
new package's own package-level types are legitimately unresolvable at
`SetObject` time, and a broad "every RefType must resolve" assert there is
impossible (verified empirically: moving the package-level `SetType` loop
before finalization panics with "non-escaped object should not have zero
hash"). Function-local types are the one class exempt from the cycle —
methods attach exclusively to package-level named types, so local type
records embed no object refs — which is what makes both the pre-finalize
`saveFuncLocalTypes` ordering and the func-local-scoped `SetObject` assert
sound. The full invariant would only be assertable at tx-commit time, a
heavier mechanism deliberately not built here.

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
- **Save-time AST walk (`Transcribe` over the fileset in
  `saveFuncLocalTypes`)** — the shape #5935 ships: enumerate local types at
  addpkg by walking every file for `*TypeDecl` nodes. Equivalent output and
  cost class (once per addpkg); its completeness rests on a language fact
  ("every local type is a TypeDecl in the fileset") rather than on the
  predefine bookkeeping invariant above, at the price of the only
  `Transcribe` call outside the preprocess layer and an extra full AST
  pass. This PR exists as the side-by-side comparison of the two shapes.
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
  across a restart. `zrealm_localtype0/1/2.gno` pass on master (verified:
  the bracketed `RefType` IDs in object bytes come from the live type and
  were always correct — the bug's missing artifact is the `/t/` record,
  which Realm goldens cannot see); `zrealm_localtype3.gno` fails on master
  — a second, in-process release-build reproducer: its addpkg-time escape
  is reloaded by main's own transaction, panicking with "unexpected type
  with id ...". localtype0's Realm golden pins the on-the-wire bracketed
  `RefType` encoding; localtype3's bare one pins that main persists
  nothing (its escape happened at addpkg); localtype1/2 carry none —
  goldens cannot detect the missing `/t/` record, so hundreds of lines of
  hash churn buy no regression power. All four act as save-side tripwires under
  `-tags debugAssert` (`make test.debugAssert`, not yet in CI; verified:
  neutering `saveFuncLocalTypes` fails all four at save time with the
  dangling-ref assert, while untagged only localtype3 and the txtar
  fail). The `zltsh CallB` scenario turned out to persist no trace of S2
  at all (the type name is preprocess-folded into a constTypeExpr, not
  captured) — it pins TypeID stability of a type re-minted from source
  after restart, not a persistence route.
  `TestAssertFuncLocalTypesCompleteFires` pins the bookkeeping audit.
