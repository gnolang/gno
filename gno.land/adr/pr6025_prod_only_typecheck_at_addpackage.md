# ADR: AddPackage type-checks production files only

## Status

Proposed. Consensus-breaking; requires a chain relaunch or upgrade boundary.

## Context

`AddPackage` stored a package's `_test.gno` / `_filetest.gno` files **and
type-checked them** as part of transaction execution. Resolving a test file's
stdlib imports goes through `TypeCheckOptions.TestGetter`, which merged the
chain-stored stdlib package with a "test overlay" read at runtime from
`$GNOROOT/gnovm/tests/stdlibs/<pkgPath>` — a directory on each node operator's
local filesystem.

That put node-local, unreplicated, unverified state into a gas-metered
consensus computation, with three demonstrated failure modes:

1. **Gas depended on node process history.** The overlay lookup memoised in a
   process-lifetime map, so a cache hit skipped a gas-metered store read. The
   first such deploy in a process paid ~4.9M gas and later ones paid nothing.
   A node restart, an ABCI `Simulate`, or an earlier deploy all changed the
   price. **This forked topaz-1 at block 301381**: canonical committed success
   at 17,447,245 gas; nodes executing the block themselves computed 20,007,348
   against a 20,000,000 limit, went out of gas, and halted on an AppHash
   mismatch.
2. **Missing overlay tree → outcome fork.** With `$GNOROOT/gnovm/tests/stdlibs`
   absent, the same transaction fails with `ErrTypeCheck` where other nodes
   succeed — a divergence on transaction *outcome*, not merely gas.
3. **Operator injection.** The lookup joins `$GNOROOT/gnovm/tests/stdlibs/` with
   *any* stdlib path, not just the seven that exist upstream. An operator who
   creates, say, `gnovm/tests/stdlibs/strconv/inject.gno` makes their node
   **accept a transaction the rest of the network rejects**. Reproduced.

The type-check of test files buys nothing on chain. The result is discarded —
`_, err = gno.TypeCheckMemPackage(memPkg, opts)` — only the verdict is used as a
deploy gate. And the chain can never run these files: `runMemPackage` demotes
`MP*All` to `MP*Prod`, importers filter `MPFProd`, and `Run` forces `MPUserProd`.
Symbols declared in a deployed `_test.gno` are unreachable via `vm/qeval`, from
an importing package, and from `maketx run` — verified end-to-end.

This was never a deliberate decision. `36dc78f81` (*feat(gnovm): refactor
mempackage type*, #4463) is the only commit ever to touch `memPkg.Type` in
`keeper.go`. It set `MPUserAll` and, in the same diff, flipped
`issue_2763.txtar` from expecting success to expecting failure, replacing the
comment *"addpkg does not add \*_test.gno files"*. The commit body is entirely
about the gnovm type taxonomy. For roughly 14 months before it, `AddPackage`
type-checked production files only.

## Decision

Add `ProdOnly` to `TypeCheckOptions`. When set, `TypeCheckMemPackage` stops
after the production type-check pass, skipping the with-tests, `xxx_test` and
`_filetest` passes — the only three that consult `TestGetter`. The chain sets it
at `AddPackage` and `Run`.

The plumbing already existed: `typeCheckMemPackage` takes a `wtests *bool`, and
a pointer to false already meant "stop after production". `ProdOnly` exposes it.

**The production pass checks exactly the set that runs.** `GoParseMemPackage`
buckets `_filetest.gno` into `tgofs` and `xxx_test`-package files into `_gofs`,
so `gofs` holds production files plus same-package `_test.gno`; `filterTests`
strips the latter. The resulting set is exactly `MPFProd` — what the VM
executes. No file can therefore run without having been type-checked.

Skipping the later passes cannot change the *production* verdict either — but
not because a larger file set can only add errors. It can remove them: a test
file may define a symbol that resolves an otherwise-undefined production
reference. The guarantee comes from **ordering**, not monotonicity. The
production-only pass runs first and any error returns immediately (`gotypecheck.go`,
"Fail early: there's no point checking the others"), so a failing production
verdict never reached the later passes and they could not have rescued it.
Production code therefore cannot borrow declarations from test files: such a
reference is undefined and rejected, before and after this change.

What is no longer caught is errors arising only from the combination — chiefly a
test file redeclaring a production symbol (`issue_2763`) — where the production
code is valid on its own and is what runs.
`addpkg_testfile_typecheck.txtar` pins both directions.

**Every file is still parsed.** `GoParseMemPackage` runs over the full
`MPUserAll` mempackage before any type-checking, so a syntax error anywhere —
including in a `_test.gno` — still rejects the deploy. Test files are still
stored, still split into the `#allbutprod` sibling, and still served by
`vm/qfile` and gnoweb. (`vm/qdoc` parses them but exposes nothing from them —
`pkgData.testFiles` is appended to and never read — so it can only *break* on
them, which is precisely why the parse must stay.)

With no path left to the overlay, `testStdlibCache` / `testStdlibGetter` and the
keeper's `gnoenv.RootDir()` call are deleted. The net change is **~20 fewer
lines**.

### Why not filter the mempackage instead

`TypeCheckMemPackage(MPFProd.FilterMemPackage(memPkg), opts)` looks equivalent
and is not: it never *parses* test files. `vm/qdoc` fully parses stored
`_test.gno` files, so an unparseable one would deploy successfully and then
break `qdoc` — and therefore gnoweb's doc page — for that package permanently,
since the bytes are immutable in state. The parse is load-bearing.
`addpkg_testfile_typecheck.txtar` pins this.

## Consequences

- **Consensus-breaking, twice over.** Packages rejected today (a test file that
  fails type-check) now deploy, and deploy gas for a package with a `_test.gno`
  drops ~86% (measured 20.6M → 2.86M). Not replay-compatible: a historical
  `AddPackage` that ran out of gas would now succeed, flipping
  `ABCIResult.Error`, which *is* committed. Needs a relaunch or an upgrade
  boundary — as does every other option considered, since none reproduces the
  price block 301381 committed.
- **`$GNOROOT` leaves the consensus path.** All three failure modes above are
  closed structurally rather than contained. Verified with a `panic()` probe at
  the overlay getter: zero hits across the entire `TestTestdata` suite, over
  genesis loads, restarts, addpkg, run, call and queries.
- **Unmetered validator CPU decreases** — two to three fewer `go/types` passes
  per deploy. `chargePreprocessGas` is unchanged and still charges over all
  `.gno` bytes including tests, so it is now slightly conservative (its doc
  comment is updated to say so).
- **Lost guarantee:** "this deployed package's tests compile" no longer holds.
  The chain could never act on it — it cannot run them. It remains enforced
  locally by `gno test` / `gno lint`, which are unaffected (`ProdOnly` defaults
  false). The practical gap is that `gnokey maketx addpkg` performs no
  client-side type-check, so the chain was the only thing telling a deployer
  their tests compiled. Worth wiring `gno lint` into the deploy flow separately.
- **Import gating over test files is lost too**, which is a governance rule
  rather than a compile-time nicety and so is worth stating separately from the
  bullet above. Resolving a test file's imports is what produced
  `ImportNotFoundError`, `ImportDraftError` and `ImportPrivateError`; with those
  passes skipped, a `_test.gno` may now name an import path that does not exist,
  or a draft or private package, and the deploy succeeds. Reproduced
  post-genesis: a package whose `_test.gno` imports `gno.land/r/demo/draftrealm`
  deploys where it previously failed with `ImportDraftError`. The import is inert
  — the file can never run, and `doc.pkgData.imports()` reads production files
  only, so `vm/qdoc` is unaffected (verified across a restart) — but the deploy
  gate no longer enforces it.
- **Genesis is included**: `ProdOnly` applies at height 0, so a broken test file
  in `examples/` would ship without the chain objecting. That belongs in CI
  (`gno test` / `gno lint` over `examples/`); confirm CI covers it.

## Alternatives considered

**Make the overlay getter gas-transparent** — perform the metered read
unconditionally and memoise only the disk read. Implemented and reviewed first.
It fixes failure mode 1 and neither 2 nor 3: with it applied, a deploy still
fails on a node missing the overlay tree, and an injected overlay is still
accepted (both reproduced). It also leaves the machinery wired in, one careless
edit from regressing.

**Embed the overlays with `//go:embed`** — closes 2 and 3 without a consensus
break, and survived adversarial review. Rejected here only because this change
deletes the machinery outright; embedding would be the right answer if the
type-check of test files had to be kept. Note it would make
`gnovm/tests/stdlibs/**/*.gno` consensus-critical binary content in a directory
named "tests", with no CI or CODEOWNERS signal.

**Load the overlays into chain state at genesis** — architecturally consistent
with production stdlibs, but state-breaking, and it makes realm-forging test
helpers (`testing_unsafe.gno` overrides `getRealm`) chain state requiring
airtight isolation from the production import path.

**Strip `_test.gno` at deploy** — same determinism win, but loses gnoweb's test
sidebar, `vm/qfile` listing, and the ability to read a deployed realm's tests.
Test files are a genuine transparency feature.

## Testing

- `addpkg_testfile_typecheck.txtar` — an unparseable `_test.gno` is still
  rejected (the guard against the filter-based implementation); one that parses
  but fails type-check is accepted; a broken *production* file is still
  rejected.
- `addpkg_testfile_restart_gas.txtar` — two identical-shape deploys straddling
  a `gnoland restart` must report identical gas. This pins the *property* that
  failed on topaz-1, not the mechanism, so any future cache placed in front of a
  metered read on the AddPackage path trips it.
- `issue_2763.txtar` — rewritten. A `_test.gno` redeclaring a production symbol
  now deploys, and `vm/qeval` proves the production symbol is what the chain
  resolves, while `vm/qfile` proves the test file is still stored.
- `TestTypeCheckMemPackage_prodOnlyPassGating` — pins which passes `ProdOnly`
  gates, by putting a type error in one bucket at a time: a same-package
  `_test.gno` (`gofs`), a `package xxx_test` file (`_gofs`) and a `_filetest.gno`
  (`tgofs`). Each must be reported with `ProdOnly` false and skipped with it
  true. This guards the three-state `wtests` mapping: collapsing it to
  `wtests := !opts.ProdOnly` sends the default to `&true`, which stops after the
  with-tests pass and silently drops the `xxx_test` and `_filetest` passes for
  `gno test` and `gno lint`. That collapse was applied and reverted during this
  PR's review, and no existing test caught it — the two `_gofs`/`tgofs` cases
  fail under it.

Verification set from AGENTS.md, all passing:

- `go test ./gno.land/pkg/sdk/vm/ -run Gas`
- `go test ./gno.land/pkg/integration/ -run TestTestdata`
- `go test ./gnovm/pkg/gnolang/ -run Files -test.short`

## Related, not addressed here

- `app.go` reads production stdlib *source* off the operator's disk into chain
  state at genesis — same class, but genesis-only, so it fails loudly from block
  1 rather than forking a live chain.
- `defaultStore.stdlibKeyBytes` and `defaultStore.cacheNodes` are process-
  lifetime caches consulted before gas-metered reads. Neither is currently
  breakable, but both rest on unenforced invariants.
- `mempackage.go` does `goodFiles = append(goodFiles, ".go")` where `goodFiles`
  is a list of exact filenames and `goodFileXtns` is the extension list, so
  stdlib `.go` native files are silently dropped from mempackages. Compare the
  correct `goodFileXtns = append(goodFileXtns, ".go")` in
  `ValidateMemPackageAny`.
