# ADR-5544: Dedupe type persistence at TypeValue position

## Status

Proposed.

## Context

Every time a realm is committed, `saveNewPackageValuesAndTypes`
(`gnovm/pkg/gnolang/machine.go`) walks the package block's `TypeValue`
entries and calls `store.SetType(dt)` for each DeclaredType found, and the
block's own bytes carry a full inline copy of each TypeValue via
`copyValueWithRefs`' TypeValue case (`gnovm/pkg/gnolang/realm.go`).

Three issues compounded:

1. **`SetType` cache check was inverted.** `store.go` SetType had a comment
   saying "return if tid already known" but the code only early-returned on
   pointer MISMATCH (`tt != tt2`). For a cache hit with matching pointer it
   fell through and re-wrote the same bytes to `/t/<TypeID>` on every call.
   The accompanying `// TODO classify them and optimize.` comment
   acknowledged the intent/impl gap.

2. **Cross-package and uverse aliases were persisted redundantly.** A user
   realm with `type MyAlias = otherpkg.T` or `type MyErr = error` triggered
   `SetType` on a DeclaredType that already had an authoritative home —
   `otherpkg`'s own deploy for user types, or the in-memory uverse registry
   for uverse. The aliasing realm wrote an extra `/t/` entry (or a stale
   one) that was never read back: `fillType` resolves RefType via
   `store.GetType(tid)`, which hits `cacheTypes` first and returns the
   canonical instance.

3. **Block bytes inlined DeclaredTypes for every TypeValue entry.** The
   TypeValue case in `copyValueWithRefs` used `copyTypeWithRefs(cv.Type)`,
   which expanded each DeclaredType fully into the block blob. The
   authoritative definition at `/t/<TypeID>` was duplicated inside every
   block that referenced the type, even though every `RefType{ID}`
   resolves via the store-level type entry.

An empirical measurement (microbenchmark and integration txtar diffs) showed
the inline-at-TypeValue-position representation costs ~45 bytes for a
reference vs several hundred bytes to several KB for an inlined DeclaredType,
depending on field and method count. The savings scale with type complexity
and with the number of realms that alias common types.

## Decision

Four coordinated edits, each at its natural boundary:

1. **`machine.go` `saveNewPackageValuesAndTypes` (~line 747)** — filter
   `SetType` calls to DeclaredTypes whose `PkgPath == pv.PkgPath`:
   ```go
   if dt, ok := tvv.Type.(*DeclaredType); ok && dt.PkgPath == pv.PkgPath {
       m.Store.SetType(dt)
   }
   ```
   Own-package types get their canonical `/t/<TypeID>` entry. Cross-package
   aliases and uverse aliases are skipped — those types live elsewhere.

2. **`store.go` `SetType` cache check (~line 692)** — invert the check to
   match the comment:
   ```go
   if _, exists := ds.cacheTypes[tid]; exists {
       return  // idempotent; matches "return if tid already known"
   }
   ```
   Any already-cached TypeID is treated as fully persisted; no redundant
   backend write.

3. **`realm.go` `copyTypeWithRefs` `case *DeclaredType:` (~line 1291)** —
   remove the uverse guard. With the above changes, `copyTypeWithRefs` is
   never called on a uverse DeclaredType: SetType filters them out at the
   machine level or short-circuits on cache hit, and `refOrCopyType` at
   Layer 1 collapses any DeclaredType at a field/TypeValue position before
   reaching Layer 2.

4. **`realm.go` `copyValueWithRefs` `case TypeValue:` (~line 1444)** —
   persist the type as a reference rather than inline:
   ```go
   case TypeValue:
       return toTypeValue(refOrCopyType(cv.Type))
   ```
   Block bytes shrink from full inlined DeclaredType (hundreds of bytes to
   several KB per type) to a small `RefType{ID}` (~45 bytes). On decode,
   `fillType`'s RefType branch resolves via `store.GetType(tid)` — cache
   hit for uverse, backend hit for user types (either own-pkg from this
   deploy or cross-pkg from an earlier deploy).

A fifth supporting addition:

5. **`realm.go`: exported `PersistedTypeFormForTypeValue(typ Type) Type`.**
   Returns `refOrCopyType(typ)` so filetests (the new `// Types:` directive
   in `gnovm/pkg/test/filetest.go`) can render the on-the-wire persisted
   shape rather than the in-memory post-fillType canonical form.

## Testing

- **New `// Types:` filetest directive** (`gnovm/pkg/test/filetest.go`).
  Prints each declared type's name + TypeID + persisted form. Gives
  reviewers and future readers a visual anchor for the wire shape.

- **Three new realm filetests** (`gnovm/tests/files/`):
  - `alias_uverse_realm.gno` — `type MyErr = error`
  - `alias_selfpkg_realm.gno` — self-package `type MyAlias = T`
  - `alias_crosspkg_realm.gno` — cross-package `type TestingT = uassert.TestingT`

  All three show the compact `{ "ID": "..." }` form in the `// Types:`
  block. Under the old code these would have been hundreds of lines of
  inlined DeclaredType JSON.

- **Existing integration goldens (`gno.land/pkg/integration/testdata/`)
  updated** for the five txtars that pin gas-used and storage-delta
  accounting: `issue_4983`, `restart_gas`, `storage_deposit`,
  `storage_deposit_bank_send`, `storage_deposit_collector`. Every delta
  traces to the same single cause: smaller persisted byte counts.

- Full `./gnovm/pkg/gnolang/` filetest suite passes. Full
  `./gno.land/...` tree passes.

## Alternatives considered

- **Only edit #4 (TypeValue → `refOrCopyType`)** and keep the uverse guard.
  Does not eliminate the redundant `/t/.uverse.<Name>` backend writes that
  `SetType(gErrorType)` triggers (the guard catches copyTypeWithRefs but
  doesn't prevent the outer amino marshal + Set). Edits #1 and #2 are
  needed to stop the actual backend writes.

- **Only edit #1 (`machine.go` filter)** without the inverted cache check.
  For `type T struct; type MyAlias = T` in the same package, both block
  entries match the filter and both call `SetType(T_dt)` — without edit
  #2, the second call re-writes the same bytes.

- **Skip edit #3 (keep the uverse guard).** Still safe; the guard becomes
  unreachable but harmless. Removed for clarity; a comment replaces it
  documenting why the branch doesn't special-case uverse.

## Consequences

### Positive

- Eliminates redundant `/t/<TypeID>` writes for cross-package aliases.
- Eliminates spurious `/t/.uverse.<Name>` writes entirely.
- Shrinks persisted block bytes for any block that holds a TypeValue.
- Removes two inconsistent special cases (uverse guard, inverted
  `SetType` cache check).
- Resolves the `TODO classify them and optimize` on `SetType`.

### Negative (consensus-breaking)

- **Gas-used values change** for any tx that involves SetType or block
  persistence of TypeValues. Measured ranges: a few hundred to ~70k gas
  less per AddPackage tx.

- **IAVL merkle leaves change** for any realm whose block contains a
  TypeValue that ends up in an escaped object. Object bytes are hashed
  at `store.go:~588-594` and the hash is written to `iavlStore`. TypeValue
  payload changed → object bytes changed → hash changed → leaf changed.

Both are strictly deterministic across nodes running the same version.
But existing chain state cannot replay under the new code and produce
the same gas/merkle values as the old chain recorded. **This change is
only shippable for a new chain launching from genesis, or as part of a
coordinated hard fork / consensus-version gate.**

Backward decode is safe: new `fillType` handles both RefType and
inlined-DeclaredType shapes in block bytes, and stale `/t/.uverse.<Name>`
entries written by old code are never read (cache hit wins). So there's
no lock-in risk — a node can read old-format state without modification,
but any tx re-executed will produce new-format output.

### Neutral

- `aminoCache` is populated less often from `SetType` (since fewer calls
  reach the write block). Verified as a pure perf cache keyed by bytes
  hash; no correctness dependency.

- Other callers of `SetType`: `store.go:350` (pkgGetter path, runs for
  every block value with TypeKind) naturally benefits from the new cache
  check. `store_test.go` tests are unaffected.

## Follow-ups

- File a companion issue documenting that this is a hard-fork-class
  change so gnoland1 or whichever fresh-genesis launch adopts it.

- Consider whether `machine.go:747-759` should also skip types whose
  `PkgPath` is empty (pre-defined / anonymous) — current code doesn't
  generate such DeclaredTypes but a future change might, and the filter
  would silently drop them.
