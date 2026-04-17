# ADR: Allow xtest back-import at genesis (#4530)

## Context

Issue [gnolang/gno#4530](https://github.com/gnolang/gno/issues/4530) shows two
failure modes for a perfectly legal xtest topology — a package `importee_test`
(package name `<importee>_test`, i.e. xtest / blackbox test) that imports
`importer`, where `importer` in turn imports `importee`. This topology is
allowed by Go because xtest is a separate compilation unit.

Gno rejects it in two independent places during chain bootstrap:

1. **`gnoland start --lazy`** fails with
   `sorting packages: cycle detected: gno.land/r/issue/importee`.
2. **`loadpkg` in txtar / genesis `AddPkg`** fails with
   `importee_test.gno:7:2: could not import gno.land/r/issue/importer (unknown import path ...)`.

Root causes, verified by code inspection:

- `gnovm/pkg/packages/readpkglist.go` merges imports of all FileKinds
  (`PackageSource`, `Test`, `XTest`, `Filetest`) into the single
  `gnomod.Pkg.Imports` list used by `gnomod.PkgList.Sort()`. xtest/filetest
  edges form legitimate back-dependencies that the DFS cycle-detector mistakes
  for real cycles. `gno.land/pkg/gnoland/no_cycles_test.go` already documents
  the correct model (xtest as its own graph node) and uses its own graph
  builder, so the two paths are inconsistent today.

- `gno.land/pkg/sdk/vm/keeper.go` calls `gno.TypeCheckMemPackage`, which runs
  STEP 4 xtest and filetest passes unconditionally. At genesis the deploy
  order is derived from prod imports only, so an xtest import may reference
  a package not yet in the store and the Check fails with
  `ImportNotFoundError`.

## Decision

Two surgical, non-consensus-breaking changes:

1. **`gnovm/pkg/packages/readpkglist.go`**: exclude `FileKindXTest` and
   `FileKindFiletest` from the import-merge used for deploy-order topological
   sort. `FileKindPackageSource` and `FileKindTest` remain (internal
   `_test.gno` files share the prod package, so they cannot form back-edges
   without a Go-level cycle).

2. **`gnovm/pkg/gnolang/gotypecheck.go`**: add a new opt-in
   `TypeCheckOptions.SkipTestFileTypeCheck` field. When set, the root
   `typeCheckMemPackage` call uses `wtests = &true`, which executes the prod
   and prod+internal-test passes but skips the xtest and filetest Check
   passes. All files are still fully parsed by `GoParseMemPackage`, so syntax
   errors in test files remain reported.

3. **`gno.land/pkg/sdk/vm/keeper.go`**: set `SkipTestFileTypeCheck = true`
   **only** at genesis (`ctx.BlockHeight() == 0`). Post-genesis `AddPkg`
   behavior is byte-identical to today, preserving live-chain replay apphash.

## Alternatives considered

- **Unconditional skip at keeper**: simpler, but flips post-genesis `AddPkg`
  outcomes from reject→accept for any xtest with missing cross-package
  imports, causing different state (pkg stored + events emitted) on
  previously failing txs → different apphash on replay. Rejected.

- **Two-pass genesis (deploy all prod first, then validate test files)**:
  requires threading through the ABCI init-chainer and re-entering the VM
  keeper with all packages already deployed. Over-scoped for a targeted fix.
  Rejected.

- **Break cycles in Sort() by detecting xtest-only back-edges**: replaces
  a simple DFS with SCC analysis, and makes the sort-time and typecheck-time
  views of the graph diverge silently. More invasive and less auditable than
  narrowing the merge. Rejected.

## Consequences

- **Live chains (Betanet etc.)**: no consensus impact. Post-genesis `AddPkg`
  remains strict; `ReadPkgListFromDir` is never invoked at runtime once
  `genesis.json` is frozen.

- **New chain genesis (re-generated from `examples/`)**: apphash may differ
  from a pre-fix bootstrap if (a) an examples package has a xtest back-import
  that previously tripped the bug (currently none known) or (b) the deploy
  order differs because xtest/filetest edges were previously skewing the
  topological sort. This is the intended outcome.

- **Off-chain type-checking (`gno test`, `gno lint`, filetest runner, stdlib
  init)**: unchanged. None of these callers set
  `SkipTestFileTypeCheck`; the option defaults to false.

- **Test files on chain**: a realm deployed at genesis can now ship a
  syntactically-valid but type-incorrect xtest/filetest. Since the chain
  never executes test files, the only observer is users running `gno test`
  off-chain against the deployed package, where the error will surface
  exactly as it does today. Acceptable trade-off.

- **Internal `_test.gno` imports at genesis**: the prod+internal-test pass
  still runs at genesis (only xtest and filetest are skipped), so an
  internal test that imports a package not yet deployed will still fail.
  Intentional: internal tests share the prod compilation unit, so Go rules
  forbid back-imports anyway, and any cross-package dep of an internal test
  is typically either a stdlib or already in prod imports. If this becomes
  a real constraint for a future `examples/` package, the fix is to move
  the offending imports into an xtest file.

## Verification

- `gno.land/pkg/integration/testdata/issue_4530.txtar` reproduces the issue
  pre-fix and passes post-fix.
- `TestVMKeeperAddPackage_XTestBackImport_Genesis` in
  `gno.land/pkg/sdk/vm/keeper_test.go` pins both branches: reject at
  Height>0, accept at Height=0.
- `TestNoCycles` in `gno.land/pkg/gnoland/no_cycles_test.go` continues to
  guard against genuine cycles.
- `go test ./gno.land/pkg/integration/ -timeout 20m` passes end-to-end.

Rename this ADR file to `pr<number>_xtest_genesis_handling.md` once the PR
number is assigned.
