# Lazy package file-block loading

## Context

When a `PackageValue` is loaded from the store, `defaultStore.fillPackage`
called `pv.deriveFBlocksMap(ds)`, which iterated every file in `pv.FNames` and
materialized its file block — reading and amino-decoding each one up front. A
transaction that enters only a few functions of a package (or imports a large
multi-file package and touches a fraction of it) therefore paid to hydrate file
blocks it never used.

The eager derivation was added in #4376 (`fix(gnovm): fill package value for
GetPackage and GetObject`), itself the accepted fix for a gnoswap production
panic — `"file block missing for file \"pool_type.gno\""` (#4527). That panic
came from `FuncValue.GetParent` reading `pv.fBlocksMap[fv.FileName]` directly and
panicking on a miss when a package was loaded with an unpopulated map. #4376
resolved it by guaranteeing every load path fully populated `fBlocksMap`.

The machinery to load a file block on demand (`PackageValue.GetFileBlock`, which
resolves `pv.FBlocks[i]` from a `RefValue` and caches it) already existed and
predates the eager map.

## Decision

Load file blocks lazily.

1. `fillPackage` no longer calls `deriveFBlocksMap` for multi-file packages; it
   leaves `fBlocksMap` empty. File blocks are materialized on first use.
2. `FuncValue.GetParent` (nil-`Parent` case) now calls
   `pv.GetFileBlock(store, fv.FileName)` instead of reading `fBlocksMap`
   directly and panicking. This resolves the exact site of the #4527 panic: a
   package loaded with an empty map no longer panics — it lazily loads from the
   `FBlocks` `RefValue`. The panic's root cause is *handled*, not reintroduced.
3. Eager derivation is kept on the package **creation** path
   (`RunMemPackage` → `machine.go`), where blocks are freshly built, not read
   from the store.
4. **Single-file guard**: `fillPackage` still eagerly derives when
   `len(pv.FNames) <= 1`. A package with no unused file to skip loads its one
   block on first call anyway, so laziness there only shifts *when* gas is
   charged — for no benefit. Keeping the eager path preserves master's gas for
   single-file packages (the common trivial call), confining the change to the
   multi-file case it actually optimizes.

### Determinism

Gas becomes path-dependent but remains deterministic: a transaction loads
exactly the file blocks its execution touches, identically on every node.
`PackageValue` instances are per-message (`BeginTransaction` allocates a fresh
`cacheObjects`; the keeper calls `ClearObjectCache` per message; `Write` commits
only `cacheNodes`), so a partially-materialized `fBlocksMap` never leaks across
transactions. The single-file guard branches on `len(pv.FNames)`, a serialized
field, so it is identical on every node.

## Alternatives considered

- **Pure lazy, no single-file guard** (rejected). Simpler, but single-file
  packages incur a small gas shift (+196 on trivial calls) and churn five gas
  goldens for zero benefit — there is nothing to skip in a one-file package. The
  guard removes a pointless regression and shrinks the golden footprint to one
  file.
- **Keep eager derivation** (status quo, rejected). No perf win; multi-file
  transactions keep paying to hydrate unused file blocks.
- **Cache materialized file blocks across transactions** (rejected). Would make
  gas depend on in-memory cache state, so a long-running node and a
  freshly-restarted node replaying the same transaction could charge different
  gas — a consensus fork. Per-message isolation is preserved instead.

## Consequences

- Multi-file transactions load and decode fewer objects. Measured:
  `BenchmarkPackageLoadFromStore` shows a 12-file package touched at one file
  drop from 2123 to 1213 allocs (−43%) and 102 KB to 57 KB (−44%); the
  single-file guard case is byte-identical to master (no regression). At the
  keeper level, loading the real deployed gnoswap `r/gnoswap/pool` realm graph
  (25-package closure, fetched from test13) from the store costs ~2.37M gas less.
- Single-file packages are unchanged vs master, including master's pre-existing
  behaviour of charging the one file block's allocation at `GetPackage` time
  (before the tx allocator's gas meter is wired). This change preserves that
  historical accounting rather than tightening it; closing that metering gap is
  out of scope for a perf change.
- Gas goldens: only `stdlib_restart_compare` (multi-file `strconv`/`strings`)
  changes; the other gas goldens return to master's values.
- `FuncValue.Parent` for a store-loaded package is materialized on first
  `GetParent` (unchanged timing); `FBlocks` entries remain `RefValue` after
  load, so persisted bytes and the realm-finalization crawl are unaffected.
