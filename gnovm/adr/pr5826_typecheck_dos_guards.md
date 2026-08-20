# PR #5826 — pricing the unmetered `go/types` validType walk

## Context

`TypeCheckMemPackage` runs unmetered at deploy time (`VMKeeper.AddPackage` /
`MsgRun`) and cannot be interrupted. `go/types`' `validType` walk does **not**
memoize visited types — the optimization is commented out as a workaround for
[golang/go#65711](https://github.com/golang/go/issues/65711) — so it is
exponential on value-containment fan-out: a ~40-line "doubling" chain costs ~2^40
node visits, and ~1KB of source hangs every validator. It cannot be metered
mid-flight (stdlib, non-interruptible), so it must be priced *before* `go/types`
runs, from a deterministic node count. Everything here happens between steps 3 and
4 of the type-check pipeline described in
[PR #4264 — lint/transpile](./pr4264_lint_transpile.md).

## Decision: price the walk, do not cap it

`typeExpansionCost` computes the exact node count `validType` will visit, with the
memoization `validType` lacks, which makes computing it linear. The deploy path
charges that count to `TypeCheckOptions.GasMeter` at
`typeExpansionGasPerNode = 100` before the package reaches `go/types`.

The rate is derived in two steps and an earlier revision got the second wrong:
`BenchmarkValidTypeWalk` measures 30–40 ns/node on an Apple M5 (climbing with depth
as the working set outgrows cache, and a DoS is the deep end), then
`gnovm/cmd/calibrate`'s paired output calibrates that to the Xeon reference the
`1 gas == 1ns` convention means — 2.96x slower, median over 37 shared
`BenchmarkAlloc` cases. 40 × ~2.5 = 100. The revision that shipped 25 skipped the
calibration step, a ~4x under-charge that with no ceiling behind it would have let
one block buy ~12s of walk instead of ~3s. Full derivation sits on the constant.

There is **no hard ceiling** on the count. Earlier revisions of this change had
one — first per-type at 100_000, then per-package at 1_000_000 — and it was
removed, because no setting of it earns its place:

- **Stricter than gas** and it refuses packages the sender paid for. At 1_000_000
  the charge was 2.5e7 gas at the then-rate of 25, under half the 5e7 `GasWanted`
  routine across this repo's own fixtures, so it bound only on senders funded
  enough to reach it and was invisible to everyone below — exactly inverted for a
  DoS guard.
- **More permissive than gas** and nothing above it is payable anyway
  (`GasWanted <= Block.MaxGas = 3e9`, enforced in the ante handler), so it only
  relabels an out-of-gas as something else.
- **Per-package either way**, a scope no budget has. `Tx.Msgs` is unbounded
  (`ValidateBasic` caps gas, not the message count) and one message re-checks
  every dependency it imports, whose bytes *earlier* transactions paid for:
  measured, a 55-byte package importing the tip of a 30-deep chain pulls in
  321,070 nodes against 68,750 gas of byte charges. Only a per-transaction meter
  bounds that.

What removal costs is off-chain: unmetered callers (`gno test`, `gno lint`,
gnodev) will churn on a pathological *local* package for as long as `go build`
does on the same input. Anything reached from the chain is bounded by what its
deploy could pay (≤3e7 nodes, ~3s). CI jobs that type-check `.gno` carry
`timeout-minutes: 30`, and the source has to be committed to get there, so it
fails on its author's own PR.

Consequences of having only a price:

- **The count must never under-count** (below).
- **Charged per package, before its walk.** `ConsumeGas` panics on out-of-gas and
  `go/types` re-panics non-`bailout` values, so the abort lands before the
  remaining dependencies are walked. Charging once at the end bills for CPU
  already spent, preventing nothing — the first cut did that, and
  `addpkg_typecheck_fanout.txtar` reported 3.8e14 gas instead of aborting.
  Returning an error would be *worse*: `go/types` records importer errors and
  keeps resolving the rest, leaking the CPU it meant to save.
- **A `GasMeter`, not a bespoke callback**, since `gnolang` already imports tm2's
  store and carries one (`MachineOptions.GasMeter`). `newParserCallback` is no
  counter-example: `pkg/parser` is a tm2-free fork, so *it* needs the callback.
- **Not computed at all when the meter is nil.** The count exists only to be
  charged, and resolving the dependency graph is the expensive part, so off-chain
  callers now skip it entirely.
- The rate is a gnovm constant, not a vm `Params` field: it is a measured ns/node
  rate for work gnovm performs, which is where `tokenCostFactor` and `OpCPU*`
  live. `Params` is amino-generated, so adding a field needs `go generate` plus a
  new legacy fingerprint.

`TestVMKeeperAddPackage_TypeExpansionGasCharged` compares pointer-vs-value chains
at equal declaration count (736,055 gas apart over 18 source bytes). It is the
only thing observing the charge: an `AddPackage` wiring slip during development
left it at zero and every other suite still passed.

## Decision: the cost model must over-count, never under

With no ceiling behind it, an under-counted edge is under-charging, which is the
same denial of service at a discount — and it multiplies, since a leaf at the base
of a doubling chain of depth `d` is walked 2^d times. So:

| construct | treatment | why |
|---|---|---|
| type params, `\|`, `~` | **rejected** (`checkNoUncountableGenerics`) | `cost()` cannot model them |
| dot imports | **rejected** (`checkNoDotImports`) | hide a type's expansion from `cost()` |
| predeclared names | exactly 1 | `validType` stops there |
| `.gnobuiltins.gno` `realm`/`address` | exactly 2 (`gnoBuiltinShimExpansion`) | shim injected *after* these guards run |
| imported stdlib types | `leafExpansionBound = 32` | over-counts; see below |

Both rejections are therefore **preconditions of the pricing**, not independent
niceties; `checkNoUncountableGenerics` must stay narrow (reject what `cost()`
cannot count, nothing more) with completeness left to `Go2Gno` (#6059).

The shim needs its own exact entry rather than either default. `address` is the
type of much of the stdlib's public API — `chain/runtime.Realm` is
`struct{addr address; pkgPath string}` — so scoring it 1 under-counts (4 against a
real 5), while lumping it in with `leafExpansionBound` over-charges honest code
hard: it alone moved the measured honest maximum from 431 to 8175 nodes.
`TestGnoBuiltinShimExpansion` fails if the shim grows a containment edge.

Honest code pays almost nothing: the largest per-package total over all stdlibs
and examples including test files is **431** (~43k gas), pinned by
`TestHonestTypeExpansionUnderBudget`. No gas fixture needed re-pinning and
`TestTestdata` passes unchanged. The guard scores test files that `ProdOnly`
excludes, so it can over-charge, never under.

Toolchain-upgrade check: this relies on `go/types` re-panicking non-`bailout`
values and on `check.posStack` being empty during import resolution (so the abort
prints no stderr trace). Re-verify on a Go bump.

## Decision: stdlib types are an unresolved leaf, priced at their maximum

`expansionPkgResolver` returns nil for stdlib paths, because `go/types` serves
stdlib imports from its result cache without a store read — resolving them here
would add store gas the deploy otherwise never pays. `namedCost` therefore prices
them at `leafExpansionBound = 32`, the measured max over every **exported** stdlib
type (19, `regexp.Regexp`; those are the only ones a user package can name),
pinned by `TestLeafExpansionBound`.

The margin is deliberately thin: every stdlib reference in honest code pays it, so
headroom here is a tax on ordinary deploys. It is safe as a constant only because
the set of stdlib types is fixed by the binary — stdlib source ships with the node
and cannot import user packages, so no transaction can grow it. A table
precomputed during `LoadStdlib` would price them exactly at no gas cost; not done,
since the edge is already priced above its cost.

## Decision: one shared parse cache across nested type checks

The cost model runs once per imported package, and each run re-resolved and
re-parsed its own dependencies — quadratic in the import graph. On a complete-DAG
graph, type-checking the final package: **24.9 / 63.2 / 262.8 / 633.6 ms** at
N = 20 / 40 / 80 / 120, growing 25x over a 6x rise in N where master is linear.
`gnoImporter` now owns one `expansionPkgCache` shared by every
`expansionChecker`: **6.0 / 10.3 / 22.5 / 35.4 ms** — linear, 18x faster at
N=120. Store gas is unchanged either way (`memoizingGetter` already deduplicated
`GetMemPackage`); what repeated was `GoParseMemPackage`. This is a hot-path fix,
not cold-start: dependencies are re-priced on *every* `MsgAddPackage`.

A checker's own **entry** package stays outside the cache: it is seeded with
whichever file set its caller is type-checking (test files at top level,
`MPFTest` for an `xxx_test` self-import) and must never stand in for a
dependency's prod-only sources. `TestExpansionPkgCacheSharing` pins that sharing
changes no price, in both visit orders.

**Why not skip pricing imported packages instead?** Unsound: `validType` runs on
**every** declaration of a package it checks, not only referenced ones, so a
pathological type in a dependency costs the node even if the entry package never
names it — and entry-only resolution follows only edges reachable from the entry's
own declarations. Packages deployed before this change were also never priced by
it. (Packages served from `permCache` *are* skipped, but a cached
`*types.Package` is proof `validType` already completed — not precedent.)

**Follow-up:** dependencies are still parsed twice per type check, once by the
resolver into a throwaway `token.FileSet` and once by `typeCheckMemPackage`.
Sharing that is blocked on unifying the FileSet (`go/types` needs positions in
`gimp.fset` for consensus-visible error text) and on the in-place AST mutations
`uniqueDecls`/`prepareGoGno0p9` perform, safe today only because the parses are
independent. A bounded 2x win against real hazard.

## Decision: dot imports are rejected, not counted

`namedCost` resolves an unqualified type name in the declaring package only. A
dot-imported type is named by a bare identifier, so it lands in the unresolved
case while `validType` expands it in full across the import boundary — the
cross-package hole again: written as `pkg.T` a chain is priced in microseconds,
written as `T` under `import . "pkg"` it was walked for free. Gno never accepted
dot imports, but the preprocessor's rejection runs *after* the type checker on the
deploy path, too late.

Rejected rather than counted because dot-import visibility is per **file** while
`namedCost` memoizes per `(package, name)`: two files of one package may
dot-import different packages exporting `T`, so counting needs either a per-file
memo key (weakening the memoization that keeps the model linear) or a
package-wide max. No blast radius: no `.gno` file in stdlibs, examples or testdata
uses a dot import. Gno's ban is undocumented and has no recorded rationale —
tracked in #6076 — but this guard does not depend on why.

## Alternatives weighed

**Metering inside `go/types`.** Rejected: it is stdlib here, not a fork, so
metering `validType` means forking a large, fast-moving package and re-syncing it
every toolchain bump, across several internal passes with no single hook point.
Charging a count computed outside `go/types` needs no fork.

**A governance `Params` rate.** Deferred, see above. Note that without a ceiling
a rate set too low under-charges the walk outright — as the 25 this change started
with did — so handing it to governance is a heavier decision than it was.

## Open question: fatal vs. normal type-check errors

The two syntactic rejections are indistinguishable from ordinary `go/types`
diagnostics, so a filetest tripping one must pin two directives even though
preprocess only adds an unrelated secondary error. Deploy stops on any type-check
error, so this is harness-only; splitting them needs a deliberate definition of
the unsupported-Gno subset against ~500 filetests that pin both.
