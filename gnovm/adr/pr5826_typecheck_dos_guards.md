# PR #5826 — pre-type-check guards for the unmetered `go/types` path

## Context

`TypeCheckMemPackage` runs unmetered at deploy time (`VMKeeper.AddPackage` /
`MsgRun`) and cannot be interrupted. `go/types` validates that named types do not
"expand" indefinitely (`validType`, src/go/types/validtype.go), and deliberately
does **not** memoize visited types — the optimization is commented out as a
workaround for [golang/go#65711](https://github.com/golang/go/issues/65711). The
walk is therefore exponential on value-containment fan-out: a ~40-line "doubling"
chain makes it visit ~2^40 nodes, so ~1KB of source hangs every validator. That
is a consensus DoS, and it cannot be metered mid-flight (stdlib black box,
non-interruptible) — it has to be bounded *before* `go/types` is invoked, by a
deterministic node count.

Three guards run in that pre-type-check step, plus one caching change that keeps
them from being quadratic:

| guard | rejects | why it must run here |
|---|---|---|
| `checkNoUncountableGenerics` | type parameters, `\|`, `~` | the shapes `cost()` cannot model |
| `checkNoDotImports` | dot imports | hides a type's expansion from `cost()` |
| `checkTypeExpansionBound` | over-budget expansion totals | the DoS itself |

The decisions below record why each is shaped the way it is, what was
deliberately left out, and the limits that remain. Naming, scope and the
`Go2Gno` division of labour are discussed further in
[#6059](https://github.com/gnolang/gno/issues/6059).


## Open question: fatal vs. normal type-check errors

`TypeCheckMemPackage()` returns two kinds of errors that today are
indistinguishable to callers:

- **normal diagnostics** — ordinary `go/types` type errors ("constant overflows
  uint16"). The Gno preprocessor re-checks the same code, so the filetest
  harness deliberately runs both and pins both (`// TypeCheckError:` for
  `go/types`, `// Error:` for preprocess) to cross-check them.
- **fatal rejections** — the package uses an unsupported construct or trips a
  DoS guard (generics/type-sets via `checkNoUncountableGenerics`, type-expansion fan-out
  via `checkTypeExpansionBound`). Proceeding is meaningless: preprocess then
  emits an unrelated secondary error (e.g. `name P not defined` for a type
  parameter), so such filetests must pin two directives for no real benefit.

The deploy path already stops on *any* type-check error (`AddPackage` →
`ErrTypeCheck`), so this only affects the filetest harness. A future change could
tag fatal rejections as a distinct error kind and have `runFiletest` stop before
preprocess for them — but only as part of a deliberate definition of the
"unsupported Gno subset" (which errors are fatal: these guards? `go1.18`
version errors? all unsupported-feature rejections?), not a bolt-on. ~500
existing filetests pin both directives, so the split must be introduced
carefully.

## Decision: stdlib types are a bounded leaf in the type-expansion guard

`checkTypeExpansionBound` follows value-containment across imports to catch a
fan-out split over several packages. For imported **stdlib** types it stops:
`expansionPkgResolver` returns nil, so a stdlib `pkg.T` is counted as a leaf (1)
rather than resolved.

Why this is safe (no bounded-factor argument needed at the call site):

- The exponential DoS vector is value-containment **fan-out**, and fan-out lives
  in **user** types, which the guard counts in full. A user doubling chain over a
  stdlib type still explodes the user-side count and trips the budget.
- A stdlib type **cannot import user packages**, so its own expansion is fixed
  and small (measured max 28 across all stdlibs, and only 19 over the *exported*
  ones, which are all a user package can name) and cannot grow with input.
- Net: the leaf under-counts only by a bounded per-reference constant
  (`real <= K_max * counted`, `K_max` = 19), so a package that passes the budget
  has bounded real validType cost — it can never hide a fan-out.

Why not fetch/count stdlib source: `go/types` answers stdlib imports from its
result cache (permCache) **without a store read**, so fetching stdlib source in
the guard would add store gas the deploy otherwise never pays (the same class of
regression as double-fetching a dependency).

Exact stdlib counting is possible without that gas: precompute a
`stdlibPkgPath -> max expansion` table during `LoadStdlib` (deterministic, at
init, so no cold/warm gas skew) and look it up per reference. It was considered
and deferred: it is cross-module (gnovm type-check API + gno.land keeper) and
buys exactness for a leaf that is already bounded-safe, with no package flipping
accept/reject at the current counts. Revisit if stdlib expansion ever grows.

## Decision: the expansion budget bounds the per-package TOTAL, not the largest type

`go/types` calls `validType` once per declared named type, so the walk a
transaction pays for is the **sum** over declarations. An earlier revision of
this guard capped the largest single type at 100_000, which left the sum
unbounded: a package can declare thousands of just-under-cap types that all
share one deep chain, at ~28 source bytes each, so a per-type cap of `B` permits
~`B/28` nodes per source byte with no ceiling on the total. Measured, that was
~130µs of `validType` per source byte, against the ~1.25µs/byte that
`Params.PreprocessGasPerByte` charges for type-check + preprocess — a ~100x
under-pricing, so a 1MB `MsgAddPackage` (the `MaxBlockTxBytes` limit) bought
minutes of unmetered CPU for ~1.25e9 gas.

The budget is therefore the per-package total, set to 1_000_000:

- **Honest code is ~4 orders of magnitude below it.** Scanning all stdlib and
  example packages including their test files (209 packages, 702 named types),
  the largest per-package total is 181 (`regexp`), and the largest single type
  anywhere is 28. `TestHonestTypeExpansionUnderBudget` pins this at a 100x
  margin, so the false-rejection argument cannot rot silently as real code grows.
- **The per-package walk is bounded at ~21ms** (~21ns/node, Apple Silicon
  go1.25), for a package that also paid byte-gas on every byte it took to reach
  the budget.

Consequence worth noting: because this bounds aggregate cost, it also caps
value-containment **depth** near 1000 — a linear chain of depth `d` totals ~`d²`
nodes, since each of the `d` types costs `O(d)`. That is not a special case
smuggled in; `validType` really does visit ~`d²` nodes for such a package. The
`deep linear chain passes` test was retuned from depth 2000 to 900 and a case
added pinning that depth 2000 is now rejected on honest arithmetic. There is no
third metric that keeps depth free while bounding cost: a per-declaration cap
leaves the sum unbounded (the defect being fixed), and capping both adds nothing
since the aggregate already dominates. Depth becomes free only if `validType`
memoizes, i.e. golang/go#65711 — already this file's revisit trigger.

### Known limit: the bound is per package, not per transaction

`MsgAddPackage` re-type-checks — and so re-guards — every transitive user
dependency, because `VMKeeper.typeCheckCache` holds only stdlibs and is
`maps.Clone`d per transaction, so user entries added during a tx are discarded.
Independently, a deploying package's prod declarations are walked twice: once for
the prod `cfg.Check` and once with its in-package tests. So the walk one tx can
buy is `~2 x (packages checked) x budget`, and the dependency count is bounded
only by what was deployed *earlier* — bytes the importing tx never paid for. A
pre-deployed chain of K near-budget packages plus one tiny importing package
costs `K x 21ms` for a tx charged on a few hundred bytes.

This is the same reasoning class as the per-type → per-package fix, one level up.
It is deliberately **not** fixed here: a per-transaction bound needs its own
constant with its own headroom (honest graphs also sum — 181 x 200 transitive
deps is ~36k) plus a compatibility argument that no existing import graph becomes
un-importable. The natural home is a running total on `gnoImporter`, which this
change already introduces as the per-transaction-scoped object. Tracked as
follow-up; the per-package bound is a strict improvement in the meantime.

### Alternatives weighed

**Gas-metering the walk instead of capping it.** Rejected: this is a validity
rule, not a price. A 2^40-node expansion is not "legal Gno that costs a lot" any
more than a dot import is — measured honest code peaks at 181 against a 1e6
budget — and pricing it would make it *purchasable*. `PreprocessGasPerByte` is
`bytes x constant`, a linear proxy that structurally cannot meter a superlinear
function; the cap's job is to bound the proxy's error, so the two are
complements. Metering would also require threading an `sdk.Context`/gas meter
into `TypeCheckMemPackage`, the same cross-module plumbing declined above for the
stdlib table — and all for a guard whose removal condition (golang/go#65711) is
already written down.

**A governance `Params` value rather than a constant.** Rejected.
`PreprocessGasPerByte` is a param *because* it is a price needing re-tuning as
hardware drifts, and it carries `Validate` bounds, amino encoding, legacy
defaulting and regression tests as the cost of that. None of it buys anything
here: the accept/reject boundary sits 4 orders of magnitude above honest code, so
no plausible governance action needs to move it, while a param would add two
bricking modes a constant cannot have (set to 1, every deploy is rejected; set
huge, the DoS reopens). Changing the constant *is* a consensus change, so it
requires a binary upgrade — the correct ceremony for a safety rule, and stronger
than a vote. Note the coupling to `PreprocessGasPerByte` is one-directional:
governance may legally raise that param up to 80x and this constant cannot
follow. Harmless at the current headroom, but it is not a calibrated pair.

Also note the stdlib-leaf under-count does **not** compound with this bound:
amplifying via stdlib leaves costs source bytes per reference, whereas the
aggregate attack needs many cheap declarations sharing one chain, so the two do
not combine. (The reachable stdlib factor is also smaller than first measured:
the worst stdlib type is `regexp.machine` at 28, but it is unexported; the worst
*exported* one, which is all a user package can name, is `regexp.Regexp` at 19.)

## Decision: one shared parse cache for the guard across nested type checks

`typeCheckMemPackage` runs once per imported package, so the guard runs once per
package too, and each run resolved and re-parsed its own dependencies. That made
dependency parsing quadratic in the import graph: measured on a complete-DAG
graph, type-checking the final package took 24.9ms / 63.2ms / 262.8ms / 633.6ms
at N = 20 / 40 / 80 / 120 — growing 25x over a 6x increase in N, where master is
linear.

`gnoImporter` now owns one `expansionPkgCache`, shared by every
`expansionChecker` it creates, holding the parsed sources and declaration index
of **resolved dependencies**. Same graph: 6.0ms / 10.3ms / 22.5ms / 35.4ms —
linear again, and 18x faster at N=120. Store gas is unaffected either way, since
`memoizingGetter` already deduplicated the `GetMemPackage` reads; what was being
repeated was `GoParseMemPackage`.

A checker's own **entry** package is deliberately held outside that cache. The
entry is seeded with whichever file set its caller is type-checking — which for a
top-level package includes test files, and for an `xxx_test` self-import is
`MPFTest`-filtered — and must never be served in place of a dependency's
prod-only sources, nor leak into the cache where a later nested check would
receive it. `TestExpansionPkgCacheSharing` pins that sharing changes no verdict,
in both visit orders.

Note this is a hot-path fix, not a cold-start one: because the keeper's
type-check cache is stdlib-only and cloned per transaction, dependencies are
re-checked and re-guarded on *every* `MsgAddPackage`, so the quadratic parse is
what production actually paid.

**Why not simply skip the guard for imported packages?** It looks attractive —
each dependency was guarded at its own deploy, and the entry guard already
crosses imports transitively with memoization — but it is unsound on three
independent grounds:

1. `go/types` runs `validType` on **every** declaration of a package it checks,
   not only the ones the importer references. So a pathological type in a
   dependency hangs the node even when the entry package never names it, and the
   entry guard — which only follows containment edges reachable from the entry's
   own declarations — cannot see it. This is the decisive one.
2. "Already guarded at its own deploy" is false for existing chain state: every
   package deployed before this guard existed was never validated by it.
3. It is false off-chain too. `TypeCheckMemPackage` is the same code path for
   `gno test`, `gno lint` and gnodev, where dependencies come from the filesystem
   and no deploy-time guard ever ran.

(The guard *is* already skipped for packages served from `permCache`, but for the
opposite reason: a cached `*types.Package` is proof `validType` already ran to
completion on it. That is not precedent for skipping.)

**Follow-up, not done here:** every dependency is still parsed twice per type
check — once by `expansionPkgResolver` into a throwaway `token.FileSet`, once by
`typeCheckMemPackage` into `gimp.fset`. Sharing that parse is blocked on
unifying the FileSet (go/types needs positions in `gimp.fset` for
consensus-visible error text) and on the in-place AST mutations `uniqueDecls` and
`prepareGoGno0p9` perform, which today are safe only because the two parses are
independent. A bounded 2x win against real hazard; the shape would be one
per-importer parse cache consumed by both, subsuming `expansionPkgCache`.

## Decision: dot imports are rejected before go/types, not counted

`checkTypeExpansionBound` resolves an unqualified type name in the **declaring
package** only (`namedCost`), and scores anything it cannot resolve — builtins,
type parameters, unresolvable qualifiers — as a leaf worth 1 node. A dot-imported
type is referenced by a bare identifier, so it lands in that leaf case while
`validType` expands it in full across the import boundary without memoizing. The
result is the cross-package hole again, reached by an unqualified name: a chain
written as `pkg.T` is rejected in microseconds, while the identical containment
shape written as `T` under `import . "pkg"` passes the budget and leaves go/types
churning inside `validType` for tens of seconds.

Gno never accepted dot imports — the preprocessor panics on them — but on the
deploy path preprocess runs *after* the type checker, so that rejection lands
only once the unmetered CPU has already been spent. `checkNoDotImports` therefore
rejects them in the same pre-type-check step as `checkNoUncountableGenerics`, for the same
reason and with the same placement argument.

Why reject rather than teach the cost model to count them:

- **Dot-import visibility is per file; the memo is per package.** `namedCost` is
  keyed `(package, name)`, which is sound exactly because an unqualified type
  name resolves in package scope. Two files of one package may dot-import
  different packages that each export `T`, so counting would need either a
  per-file memo key (weakening the memoization that makes this guard linear) or a
  package-wide max over every dot-imported package (an over-count, and a new
  false-rejection path on honest code).
- **The leaf cases stay provably safe.** "An unqualified type name resolves in
  the declaring package or is a bounded leaf" is the invariant that justifies
  every `return 1` in `cost()`/`namedCost`. Dot imports are the sole construct
  that breaks it; removing them keeps the invariant statable in one line instead
  of leaving a resolution fallback whose completeness must be re-derived on every
  future edit.
- **No blast radius.** No `.gno` file in the stdlibs, examples, or testdata uses a
  dot import (the sole occurrence, `gnovm/tests/backup/import2.gno`, is not
  referenced by any test), so nothing false-rejects.

Consequence: the guard's soundness now rests on two syntactic preconditions —
no generics/type-sets, no dot imports — both recorded in the `MAINTENANCE:` note
on `typeExpansionBudget`. If either rejection is ever relaxed, the corresponding
edge must be counted in `cost()` first.
