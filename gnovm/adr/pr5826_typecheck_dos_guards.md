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

### Deriving the rate

`1 gas == 1ns` here means *reference* hardware — Intel Xeon Platinum 8168 @ 2.70GHz,
per `machine.go`'s `OpCPU*` table and `gnovm/cmd/calibrate/README.md` — so the rate
is nanoseconds per node **on that machine**, which takes two steps:

| step | source | result |
|---|---|---|
| measure the walk | `BenchmarkValidTypeWalk`, Apple M5 | 30.1 / 30.4 / 34.9 ns/node at depth 18 / 20 / 22; *marginal* rate climbs 34.3 → 40.3 from depth 22 → 26 as the working set outgrows cache |
| calibrate to the Xeon | rerun `cmd/calibrate`'s `BenchmarkAlloc` locally, compare to the shipped `bench_output_do_dedicated.txt` | Xeon 2.96x slower over 37 shared cases (median), 2.2–3.2x on small allocations — the regime resembling `validType`'s pointer chasing |

A denial of service is the large-working-set end, so price ~40, not ~30:
`40 × ~2.5 = 100`, with the spread putting the defensible range at 88–128. (The
shipped `bench_output_m2_arm64.txt` gives 2.54x for an M2 if rerunning is not an
option.) The check that ties it to something real: at 100 a whole block of gas buys
3e7 nodes, about **3s** of `validType` on reference hardware — which is what
`1 gas == 1ns` should mean for a full block.

Step 2 is the one to get wrong: skipping it and taking the development-machine
figure directly gives ~25, a ~4x under-charge, and one block would then buy ~12s of
walk. Since this charge is the only thing pricing the walk, that is the whole
defence off by 4x. The calibration factor stays the dominant uncertainty, as
`PreprocessGasPerByte` notes for itself; measuring on reference hardware removes
it.

There is **no hard ceiling** on the count, because no setting of one earns its
place:

- **Stricter than gas** and it refuses packages the sender paid for. A cap low
  enough to bite lands inside a normal budget — the 5e7 `GasWanted` routine across
  this repo's own fixtures — so it would bind only on senders funded enough to
  reach it and stay invisible to everyone below. Exactly inverted for a DoS guard.
- **More permissive than gas** and nothing above it is payable anyway
  (`GasWanted <= Block.MaxGas = 3e9`, enforced in the ante handler), so it only
  relabels an out-of-gas as something else.
- **Per-package either way**, a scope no budget has. `Tx.Msgs` is unbounded
  (`ValidateBasic` caps gas, not the message count) and one message re-checks
  every dependency it imports, whose bytes *earlier* transactions paid for:
  measured, a 55-byte package importing the tip of a 30-deep chain pulls in
  321,070 nodes against 68,750 gas of byte charges. Only a per-transaction meter
  bounds that.

What having no ceiling costs is off-chain: unmetered callers (`gno test`, `gno lint`,
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
- **The count→gas conversion clamps.** `cost()` saturates at `math.MaxUint64`, and
  `int64(math.MaxUint64)` is `-1`, so converting straight through would charge
  `-100` gas — a *refund* for the worst package a sender could submit.
  `expansionGas` clamps at `math.MaxInt64` instead, making an unrepresentable count
  simply unaffordable; `TestExpansionGas` pins the boundary.
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
would add store gas the deploy otherwise never pays. `namedCost` prices them at
`leafExpansionBound = 32`, the measured max over every **exported** stdlib type
(19, `regexp.Regexp`; those are the only ones a user package can name), pinned by
`TestLeafExpansionBound`. The margin is thin on purpose: every stdlib reference in
honest code pays it, so headroom is a tax on ordinary deploys. Safe as a constant
only because stdlib source ships with the binary and cannot import user packages,
so no transaction can grow it.

Resolving stdlib for an exact count is what the estimate replaces, and it is not
free the way resolving a user dependency is. For a user dependency `go/types` will
call `GetMemPackage` itself, so `memoizingGetter` deduplicates the guard's fetch to
nothing. For stdlib it never does: `ImportFrom` returns from `permCache` before
reaching the getter, so a fetch here would be a store read — and a re-parse — that
no deploy otherwise pays, on every deploy, since every package imports stdlib. The
exactness would buy no safety either, only a slightly smaller over-charge: 32
against a real 19, worth ~25k gas on the largest honest package.

Two ways to get exactness for free, both deferred:

- **A table precomputed during `LoadStdlib`.** That path already parses and
  type-checks every stdlib package (`keeper.go`, `TypeCheckMemPackage` per
  `stdlibs.InitOrder()`), discarding the ASTs afterwards; scoring them there and
  keeping a `map[pkgPath]uint64` rides on a parse that already happens. Retaining
  the ASTs instead is the wrong trade — `LoadStdlibCached`'s own comment gives
  memory and cold start as why normal nodes avoid holding that.
- **Score from the cached `*types.Package`.** `permCache` already holds it, and its
  `Scope()` exposes the `*types.Named` whose `Underlying()` carries exactly the
  containment edges `validType` walks — no source, no parse, no store read, and
  `gimp.permCache` is already in hand where the resolver runs. The catch is that
  the *price* would then depend on a cache's contents. On chain that cache is
  always warm, so it is deterministic in practice, but making a consensus-visible
  charge a function of cache state is the classic shape of a consensus bug. A
  compile-time constant has no such coupling, which is why it wins for now.

## Decision: one shared parse cache across nested type checks

The cost model runs once per imported package, and each run re-resolved and
re-parsed its own dependencies — quadratic in the import graph. Type-checking the
final package of a complete-DAG graph took **24.9 / 63.2 / 262.8 / 633.6 ms** at
N = 20 / 40 / 80 / 120 (25x growth over a 6x rise in N, where master is linear);
with one `expansionPkgCache` shared by every `expansionChecker`, **6.0 / 10.3 /
22.5 / 35.4 ms** — linear, 18x faster at N=120. Store gas is unchanged
(`memoizingGetter` already deduplicated `GetMemPackage`); `GoParseMemPackage` was
what repeated. A hot path, not cold start: dependencies are re-priced on *every*
`MsgAddPackage`.

A checker's own **entry** package stays outside the cache: it is seeded with
whichever file set its caller is checking (test files at top level, `MPFTest` for
an `xxx_test` self-import), so it must never stand in for a dependency's prod-only
sources. `TestExpansionPkgCacheSharing` pins that sharing changes no price, in both
visit orders.

**Why not skip pricing imported packages instead?** Unsound: `validType` runs on
**every** declaration of a package it checks, so a pathological type in a
dependency costs the node even if the entry package never names it — and entry-only
resolution follows only edges reachable from the entry's own declarations. Packages
deployed before this change were also never priced. (`permCache` hits *are*
skipped, but a cached `*types.Package` is proof `validType` already completed.)

**Follow-up:** dependencies are still parsed twice per type check, once by the
resolver and once by `typeCheckMemPackage`. Sharing that is blocked on unifying the
FileSet (`go/types` needs positions in `gimp.fset` for consensus-visible error
text) and on the in-place AST mutations `uniqueDecls`/`prepareGoGno0p9` perform,
safe today only because the parses are independent. A bounded 2x win.

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

**Fork `go/types` and meter `validType` from inside.** This is the strongest
alternative, since this repo already does exactly that one layer up: `gnovm/pkg/parser` is a fork of
`go/parser` (5,460 lines) whose reason to exist is a metering hook,
`ParserCallback func(tok token.Token, nestedLevel int)`, driven by
`newParserCallback` in `go2gno.go`. Fork-and-hook is the established answer here to
an unmetered stdlib pass, not a novelty.

And metering **measures** where this change **predicts**, which would delete the
entire soundness apparatus below: `cost()` mirroring validType's edges, both
syntactic rejections *in their cost-guard role*, `leafExpansionBound`,
`gnoBuiltinShimExpansion`, and — since nothing would need pre-parsing — the shared
parse cache and its quadratic-parsing fix too. The invariant a future maintainer can
silently break, "`cost()` must never under-count", would simply not exist.

Rejected anyway, on size and on shipping risk. `go/types` is **34,545 lines across
107 files**, 6.3x the parser fork, and it is a whole type system rather than a token
stream, so re-syncing it on each Go bump is qualitatively harder. That inverts the
"don't add a layer" argument: the fork *is* the larger layer, only made of code we
did not write. This guard is ~450 lines whose lifetime is bounded by `go/types`' own
— when the type checker goes, one file is deleted — so neither option accretes and
disposal is symmetric; the only question is which is cheaper to hold in the interim.
Separately, a forked type checker makes consensus-visible error text ours to keep
byte-stable, which is a large review surface for a security fix that should land
quickly.

If `go/types` removal lands sooner than expected this stays the right call, because
the guard is deleted rather than migrated. If it slips, the fork would have been the
expensive thing to be holding — `pkg/parser` shows how long these live.

**A governance `Params` rate.** Deferred. With no ceiling behind it, a rate set
too low under-charges the walk outright, so handing it to governance is a heavier
decision than it looks: the calibration below is the only thing keeping the price
honest.

## Open question: this prices one pass, and assumes the rest are near-linear

`validType` is now priced by structure. Everything else in type-check and
preprocess is priced by `PreprocessGasPerByte` (1250 gas/byte, charged at
`MsgAddPackage` and `MsgRun`), and that is only correct for work whose cost is
roughly proportional to source size. **Nobody has audited the other passes against
that assumption.** This change fixes the hole that was found and demonstrated; it
does not establish that there are no others.

The criterion for needing structural pricing is not "is this pass expensive" but
"can its cost be approximated from bytes". Note the threshold is lower than it
looks:

| pass complexity in source size | is byte gas enough? |
|---|---|
| linear | yes |
| **quadratic** | **no** — `MaxTxBytes` bounds source, but 1MB of O(n²) work is ~1e12 operations against the ~1.25e9 gas its bytes buy |
| exponential | catastrophic; this was that case |

So a merely quadratic pass would already be badly under-priced. Two things reduce
the exposure today: the generics rejection also removes instantiation-driven
blowup, which is the other known super-linear behaviour in `go/types`; and
`PreprocessGasPerByte`'s own comment already flags host calibration as its
dominant uncertainty, so it is not claimed to be tight.

Unverified candidates, listed so the next person starts somewhere rather than from
scratch: untyped constant arithmetic (shift-driven `big.Int` growth, which
`go/types` bounds but by an unchecked-here margin); method-set and interface
satisfaction checks, plausibly quadratic in method count; and `Identical` on deeply
nested structural types. What this change contributes beyond the one fix is a
reusable shape for any of them — compute a deterministic quantity, charge it before
the pass runs.

## Open question: fatal vs. normal type-check errors

The two syntactic rejections are indistinguishable from ordinary `go/types`
diagnostics, so a filetest tripping one must pin two directives even though
preprocess only adds an unrelated secondary error. Deploy stops on any type-check
error, so this is harness-only; splitting them needs a deliberate definition of
the unsupported-Gno subset against ~500 filetests that pin both.
