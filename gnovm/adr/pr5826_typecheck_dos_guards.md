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
| `checkTypeExpansionBound` | over-budget expansion totals | the DoS itself |

Division of labour with `Go2Gno` (which owns completeness, not cost): #6059.

## Decision: the budget bounds the per-package TOTAL, not the largest type

`validType` runs once per declared named type, so a transaction pays the **sum**.
An earlier revision capped the largest single type at 100_000 and left the sum
unbounded: thousands of just-under-cap types can share one deep chain at ~28
source bytes each, so a per-type cap `B` permits ~`B/28` nodes per byte with no
ceiling. Measured ~130µs/byte against the ~1.25µs/byte `PreprocessGasPerByte`
charges — ~100x under-priced, so a 1MB `MsgAddPackage` bought minutes of CPU.

Budget is therefore the per-package total, **20_000**, chosen so the worst
*accepted* package costs about what it pays for — which is what makes the bound
hold per **transaction**, not merely per package.

Gas is charged per source byte (`PreprocessGasPerByte = 1250`, and 1 gas ≈ 1ns on
the reference machine → ~1.25µs of CPU per byte). Bytes are a poor proxy here:
each extra `type tN struct{ a, b [0]tN-1 }` line is ~31 bytes and *doubles* the
total, so the shape reaching a budget `B` in the fewest bytes is a doubling chain,
worst case `B / (33 + 31·log2(B/14))` nodes per byte at ~25ns/node.

| budget | worst accepted | µs/byte | vs priced 1.25 |
|---|---|---|---|
| 1_000_000 | 511 bytes, ~24ms | 47.6 (measured) | **38x** |
| 20_000 | 325 bytes, ~263µs | 0.81 (measured) | **0.6x** |

At 1_000_000 the 38x multiplied across messages (see below). At 20_000 per-message
cost is bounded by per-message gas, so any number of messages is bounded by the
per-tx `GasWanted` and the block gas limit — mechanisms that already exist.

Honest code is unaffected: the largest per-package total in real code is 181, over
all stdlibs and examples including test files, pinned by
`TestHonestTypeExpansionUnderBudget` — ~110x headroom.

### Known limit: per package, not per transaction

**Multiple messages — closed by the budget above.** `Tx.Msgs` is unbounded
(`ValidateBasic` caps gas, not the message count) and baseapp dispatches every
message to the handler, so N `MsgAddPackage`/`MsgRun` messages each pay the budget
in full. At the old 1_000_000 that meant ~1400 near-budget packages in a 1MB tx ≈
tens of seconds of walk. Gas parity per message removes the amplification: the CPU
a message can buy is now under the gas it pays, so `GasWanted` and the block gas
limit bound the total.

**Transitive dependencies — still open.** One `MsgAddPackage` re-guards every
transitive user dependency, because `VMKeeper.typeCheckCache` holds only stdlibs
and is cloned per tx. Those dependencies' bytes were paid for by *earlier*
transactions, so per-message gas parity does not price them. The clean fix is to
charge gas for the computed expansion count rather than only capping it — the
guard already computes it exactly, outside `go/types`, so this needs no fork (see
*Alternatives weighed*, which rejected metering **inside** `go/types`, a different
proposal). That wants its own rate parameter and gas-fixture re-pinning, so it is
left as follow-up; the cap keeps the residual bounded meanwhile.

### Alternatives weighed

**Metering inside `go/types` instead of capping.** Rejected: `go/types` is stdlib
here, not a fork, so metering `validType` means forking a large, fast-moving
package and re-syncing it every toolchain bump, and the work spans several
internal passes with no single hook point. Note this is *not* the same as charging
gas for the count this guard already computes — that happens outside `go/types`,
needs no fork, and is the recommended fix for the transitive-dependency residual
above.

**A governance `Params` value rather than a constant.** Rejected: a param adds two
bricking modes a constant cannot have (set to 1, all deploys fail; set huge, the
DoS reopens), and changing the constant is a consensus change requiring a binary
upgrade — the right ceremony for a safety rule. But note the budget is now
*calibrated against* `PreprocessGasPerByte`, and that coupling is one-directional
and unenforced: raising the param only makes the guard more conservative (more gas
per byte for the same CPU), whereas **lowering** it breaks parity, and `Validate`
only requires it to be positive. If that param ever drops, this budget must drop
with it.

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
the budget and left `go/types` churning for tens of seconds. Gno never accepted
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
`typeExpansionBudget`. If either is relaxed, the corresponding edge must be
counted in `cost()` first.

## Open question: fatal vs. normal type-check errors

`TypeCheckMemPackage` returns guard rejections indistinguishably from ordinary
`go/types` diagnostics, so a filetest tripping a guard must pin two directives
(`// TypeCheckError:` and `// Error:`) even though preprocess only adds an
unrelated secondary error. Deploy stops on any type-check error, so this is
filetest-harness-only. Tagging fatal rejections and having `runFiletest` stop
before preprocess would need a deliberate definition of the unsupported-Gno
subset, against ~500 filetests that pin both — not a bolt-on.
