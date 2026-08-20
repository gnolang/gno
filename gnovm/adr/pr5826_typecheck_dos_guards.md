# PR #5826 — pre-type-check guards for the unmetered `go/types` path

## Context

`TypeCheckMemPackage` runs unmetered at deploy time (`VMKeeper.AddPackage` /
`MsgRun`) and cannot be interrupted. `go/types`' `validType` walk does **not**
memoize visited types — the optimization is commented out as a workaround for
[golang/go#65711](https://github.com/golang/go/issues/65711) — so it is
exponential on value-containment fan-out: a ~40-line "doubling" chain costs ~2^40
node visits, and ~1KB of source hangs every validator. It cannot be metered
mid-flight (stdlib, non-interruptible), so it must be bounded *before* `go/types`
runs, by a deterministic node count.

| guard | rejects | why it must run before go/types |
|---|---|---|
| `checkNoUncountableGenerics` | type parameters, `\|`, `~` | the shapes `cost()` cannot model |
| `checkNoDotImports` | dot imports | hide a type's expansion from `cost()` |
| `checkTypeExpansionBound` | over-ceiling expansion totals | the DoS itself |

The count `checkTypeExpansionBound` computes is also reported to the keeper, which
charges gas for it — the cap bounds one package, the charge bounds a transaction.
Division of labour with `Go2Gno` (completeness, not cost): #6059.

## Decision: the ceiling bounds the per-package TOTAL, not the largest type

`validType` runs once per declared named type, so a transaction pays the **sum**.
An earlier revision capped the largest single type at 100_000 and left the sum
unbounded: thousands of just-under-cap types can share one deep chain at ~28
source bytes each, so a per-type cap `B` permits ~`B/28` nodes per byte with no
ceiling. Measured ~130µs/byte against the ~1.25µs/byte `PreprocessGasPerByte`
charges — ~100x under-priced, so a 1MB `MsgAddPackage` bought minutes of CPU.

The ceiling is therefore the per-package total, **1_000_000**, and deliberately
*not* the primary defence — that is the per-node gas charge below. A per-package
ceiling cannot bound a transaction at all, and once the charge prices the walk
accurately there is no reason for it to also second-guess what a sender can
afford: at 1_000_000 one package's charge (2.5e7 gas) already exceeds a typical
`GasWanted`, so **gas binds first — an expensive package deploys if paid for**.

An intermediate revision set it to 20_000, so the worst accepted package's CPU
stayed under what its *bytes* paid for — the right fix while byte-gas was the only
charge, superseded by per-node charging. Reverted: it only rejected packages a
sender could afford.

It covers the two cases gas cannot:

- **Unmetered callers.** `gno test`, `gno lint` and gnodev pass no `GasMeter`, so
  nothing else stops a 2^40 walk hanging a developer's machine.
- **A broken charge.** This happened: an `AddPackage` wiring slip left the charge
  at zero and every suite still passed, because the ceiling kept the blast radius
  finite. `TestVMKeeperAddPackage_TypeExpansionGasCharged` now guards the wiring;
  the ceiling is what makes such a slip survivable.

It stays global rather than applying only when no meter is set, so both callers
reach the same verdict — a local `gno test` must never be stricter than the chain —
and so it survives a mis-wired charge.

Honest code is far below: largest per-package total over all stdlibs and examples
including test files is 181 (pinned by `TestHonestTypeExpansionUnderBudget`),
~5500x headroom, and an unmetered walk is bounded at ~25ms per package. Capping a
total also caps containment depth near 1000 (a linear chain of depth `d` totals
~`d²` nodes) — honest arithmetic, not a fan-out special case.

### Making the bound hold per transaction, not just per package

A per-package ceiling says nothing about a transaction, two ways over. `Tx.Msgs` is
unbounded (`ValidateBasic` caps gas, not the message count) and baseapp dispatches
each message to the handler, so N messages each pay the ceiling in full. And one
message re-type-checks all of its transitive dependencies, whose bytes *earlier* transactions paid
for: measured, a 55-byte package importing the tip of a 30-deep chain pulls in
321,070 nodes across its transitive dependencies against 68,750 gas of byte charges — 117x unpriced, which no
per-byte charge can fix.

Closed by charging for the count the guard already computes — exactly, linearly,
and outside `go/types`. It is charged to `TypeCheckOptions.GasMeter` once per
package and **before** that package is walked, at `typeExpansionGasPerNode = 25`
(1 gas ≈ 1ns, walk ~25ns/node). The cost of those dependencies is then tied to this tx's
`GasWanted`, which the ante handler caps at `Block.MaxGas`.

- **Per package, before the walk.** `ConsumeGas` panics on out-of-gas and
  `go/types` re-panics non-`bailout` values, so the abort lands *before* the
  remaining dependencies are walked. Charging once at the end bills for CPU already
  spent, preventing nothing — the first cut did that, and
  `addpkg_typecheck_fanout.txtar` reported 3.8e14 gas instead of the rejection.
  Returning an error would be *worse*: `go/types` records importer errors and keeps
  resolving the rest, leaking the CPU it meant to save.
- **A `GasMeter`, not a bespoke callback**, since `gnolang` already imports tm2's
  store and carries one (`MachineOptions.GasMeter`). `newParserCallback` is no
  counter-example: `pkg/parser` is a tm2-free fork, so *it* needs the callback.
- **Rejected packages are not charged** — their total is the cost *avoided*.
  `TestExpansionNotChargedWhenRejected` pins it.

The two outcomes are distinct and both covered end-to-end: over the ceiling gives
a denial-of-service rejection naming the offending type
(`addpkg_typecheck_fanout.txtar`), while unaffordable transitive dependencies gives out-of-gas
(`addpkg_typecheck_fanout_deps_gas.txtar`, where a three-line package needs
20.8M gas for the dependencies it pulls in and then deploys once its budget covers it).
gnokey reports the required gas-wanted, so the OOG case is actionable too.

No gas fixture needed re-pinning: honest totals are small (largest 181, ~4.5k gas)
and `TestTestdata` passes unchanged. The guard scores test files `ProdOnly`
excludes, so the charge can over-charge, never under. The rate is a gnovm constant
beside the ceiling rather than a vm `Params` field: it is a measured ns/node rate
for work gnovm performs, which is where `tokenCostFactor` and `OpCPU*` live, not a
governance knob. Promoting it is reasonable later — the ceiling is what makes a
governable rate safe.

Toolchain-upgrade check: this relies on `go/types` re-panicking non-`bailout`
values and on `check.posStack` being empty during import resolution (so the abort
prints no stderr trace). Re-verify on a Go bump.

### Alternatives weighed

**Metering inside `go/types`.** Rejected: it is stdlib here, not a fork, so
metering `validType` means forking a large, fast-moving package and re-syncing it
every toolchain bump, across several internal passes with no single hook point.
Charging the count this guard already computes is a different thing entirely — it
happens outside `go/types` and needs no fork.

**A governance `Params` value.** Two candidates, both deferred. The *ceiling* is a
safety rule, and a param there adds two bricking modes a constant cannot (set to 1,
all deploys fail; set huge, the DoS reopens), so a binary upgrade is the right
ceremony. The *rate* is a price and genuinely belongs in `Params` — the ceiling is
what makes handing it to governance safe, since a rate set too low then
under-charges but cannot unbound the walk. Not done here only because `Params` is
amino-generated: adding a field needs `go generate`, which AGENTS.md keeps out of
ordinary PRs, plus a new legacy fingerprint for the repricing.

## Decision: stdlib types are a bounded leaf

`expansionPkgResolver` returns nil for stdlib paths, so an imported stdlib `pkg.T`
counts as a leaf (1) instead of being resolved. Safe because fan-out lives in
**user** types, which are counted in full, and a stdlib type cannot import user
packages — its own expansion is fixed and small: measured max 28 over all stdlib
types, and only **19** over the *exported* ones, which are all a user package can
name. So the leaf under-counts by a bounded per-reference constant
(`real <= 19 * counted`) and can never hide a fan-out. It also does not compound
with the aggregate bound: amplifying through stdlib leaves costs source bytes per
reference, while the aggregate attack needs many cheap declarations sharing one
chain.

Not fetched/counted exactly because `go/types` serves stdlib imports from its
result cache without a store read, so fetching here would add store gas the
deploy otherwise never pays. A precomputed `stdlibPkgPath -> max expansion` table
built during `LoadStdlib` would avoid that gas, but is cross-module and buys
exactness for a leaf already bounded-safe. Revisit if stdlib expansion grows.

## Decision: one shared parse cache across nested type checks

The guard runs once per imported package, and each run re-resolved and re-parsed
its own dependencies — quadratic in the import graph. On a complete-DAG graph,
type-checking the final package: **24.9 / 63.2 / 262.8 / 633.6 ms** at
N = 20 / 40 / 80 / 120, growing 25x over a 6x rise in N where master is linear.
`gnoImporter` now owns one `expansionPkgCache` shared by every
`expansionChecker`: **6.0 / 10.3 / 22.5 / 35.4 ms** — linear, 18x faster at
N=120. Store gas is unchanged either way (`memoizingGetter` already deduplicated
`GetMemPackage`); what repeated was `GoParseMemPackage`. This is a hot-path fix,
not cold-start: dependencies are re-guarded on *every* `MsgAddPackage`.

A checker's own **entry** package stays outside the cache: it is seeded with
whichever file set its caller is type-checking (test files at top level,
`MPFTest` for an `xxx_test` self-import) and must never stand in for a
dependency's prod-only sources. `TestExpansionPkgCacheSharing` pins that sharing
changes no verdict, in both visit orders.

**Why not skip the guard for imported packages instead?** Unsound: `validType`
runs on **every** declaration of a package it checks, not only referenced ones,
so a pathological type in a dependency hangs the node even if the entry package
never names it — and the entry guard only follows edges reachable from the entry's
own declarations. Also, packages deployed before this guard existed were never
validated by it, and `TypeCheckMemPackage` is the same path for `gno test` /
`gno lint` / gnodev, where no deploy-time guard ever ran. (Packages served from
`permCache` *are* skipped, but a cached `*types.Package` is proof `validType`
already completed — not precedent.)

**Follow-up:** dependencies are still parsed twice per type check, once by the
resolver into a throwaway `token.FileSet` and once by `typeCheckMemPackage`.
Sharing that is blocked on unifying the FileSet (`go/types` needs positions in
`gimp.fset` for consensus-visible error text) and on the in-place AST mutations
`uniqueDecls`/`prepareGoGno0p9` perform, safe today only because the parses are
independent. A bounded 2x win against real hazard.

## Decision: dot imports are rejected, not counted

`namedCost` resolves an unqualified type name in the declaring package only and
scores anything unresolved as a leaf. A dot-imported type is named by a bare
identifier, so it lands in that leaf case while `validType` expands it in full
across the import boundary — the cross-package hole again: written as `pkg.T` a
chain is rejected in microseconds, written as `T` under `import . "pkg"` it passed
the ceiling and left `go/types` churning for tens of seconds. Gno never accepted
dot imports, but the preprocessor's rejection runs *after* the type checker on the
deploy path, too late.

Rejected rather than counted because dot-import visibility is per **file** while
`namedCost` memoizes per `(package, name)`: two files of one package may
dot-import different packages exporting `T`, so counting needs either a per-file
memo key (weakening the memoization that keeps the guard linear) or a
package-wide max (an over-count, and a new false-rejection path). Rejecting also
keeps the invariant that justifies every `return 1` in `cost()` statable in one
line. No blast radius: no `.gno` file in stdlibs, examples or testdata uses a dot
import.

Consequence: the guard's soundness rests on two syntactic preconditions — no
generics/type-sets, no dot imports — both recorded in the `MAINTENANCE:` note on
`typeExpansionCeiling`. If either is relaxed, the corresponding edge must be
counted in `cost()` first.

## Open question: fatal vs. normal type-check errors

Guard rejections are indistinguishable from ordinary `go/types` diagnostics, so a
filetest tripping a guard must pin two directives even though preprocess only adds
an unrelated secondary error. Deploy stops on any type-check error, so this is
harness-only; splitting them needs a deliberate definition of the unsupported-Gno
subset against ~500 filetests that pin both.
