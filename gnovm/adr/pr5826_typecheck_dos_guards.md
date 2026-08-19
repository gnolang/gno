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

The count `checkTypeExpansionBound` computes is also reported to the keeper, which
charges gas for it — the cap bounds one package, the charge bounds a transaction.

Division of labour with `Go2Gno` (completeness, not cost): #6059.
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
total, so the worst shape is a doubling chain: `B / (33 + 31·log2(B/14))` nodes
per byte at ~25ns/node.

| budget | worst accepted | µs/byte | vs priced 1.25 |
|---|---|---|---|
| 1_000_000 | 511 bytes, ~24ms | 47.6 (measured) | **38x** |
| 20_000 | 325 bytes, ~263µs | 0.81 (measured) | **0.6x** |

Honest code is unaffected: the largest per-package total in real code is 181, over
all stdlibs and examples including test files, pinned by
`TestHonestTypeExpansionUnderBudget` — ~110x headroom.

### Making the bound hold per transaction, not just per package

A per-package cap does not bound a transaction, in two ways.
**Multiple messages.** `Tx.Msgs` is unbounded (`ValidateBasic` caps gas, not the
message count) and baseapp dispatches each to the handler, so N messages each pay
the budget in full — at the old 1_000_000, ~1400 near-budget packages in a 1MB tx.
Closed by the budget above: per-message CPU is now under per-message gas.

**Transitive dependencies.** One message re-type-checks its whole closure, whose
bytes were paid for by *earlier* transactions. Measured: a 55-byte package
importing the tip of a 30-deep chain pulls in 321,070 closure nodes against 68,750
gas of byte charges — 117x unpriced. Byte parity cannot fix this.

Closed by charging for the count the guard already computes — exactly, linearly,
and outside `go/types`, so no fork is involved. It goes to
`TypeCheckOptions.ChargeExpansion` once per package and **before** that package is
walked; the keeper's callback consumes `nodes x 25` gas (1 gas ≈ 1ns, walk
~25ns/node). Closure cost is then tied to this tx's `GasWanted`, and the ante
handler rejects any `GasWanted` above `Block.MaxGas`, so one tx cannot exceed the
block ceiling.

- **Per package, not once at the end.** The gas meter panics on out-of-gas and
  `go/types` re-panics anything that is not its own `bailout`, so the abort
  propagates out of `cfg.Check` *before* the remaining dependencies are walked.
  Charging once at the end would bill for CPU already spent, preventing nothing.
  `TestExpansionChargedPerPackage` pins one charge per package plus the
  mid-closure abort.
- **A callback, not a gas meter**, so gnovm keeps no tm2 dependency —
  `gnovm/pkg/parser`'s `ParseFile2` does the same for per-token gas.
- **Rejected packages are not charged** — their total is the cost *avoided*.
  Charging it prices prevented work and turns the informative rejection into an
  out-of-gas, which the first implementation did: `addpkg_typecheck_fanout.txtar`
  reported 3.8e14 gas instead of the expected message.
  `TestExpansionNotChargedWhenRejected` pins the fix.
- **The cap stays**: the rate is calibrated in ns/node on one machine and can be
  wrong on another, whereas the cap cannot, so a mis-estimated rate leaves the
  worst package under-charged but still *bounded*.

No gas fixture needed re-pinning (honest totals are small — largest 181, ~4.5k gas
— and `TestTestdata` passes unchanged). The guard scores test files `ProdOnly`
excludes, so the charge can over-charge, never under. The rate is a constant, not a
`Params` field, only because `Params` is amino-generated and adding a field needs
`go generate`, which AGENTS.md keeps out of ordinary PRs.

### Alternatives weighed

**Metering inside `go/types`.** Rejected: `go/types` is stdlib
here, not a fork, so metering `validType` means forking a large, fast-moving
package and re-syncing it every toolchain bump, and the work spans several
internal passes with no single hook point. Note this is *not* the same as charging
gas for the count this guard already computes — that happens outside `go/types`,
needs no fork, and is the recommended fix for the transitive-dependency residual
above.

**A governance `Params` value for the budget.** Rejected: a param adds two bricking
modes a constant cannot (set to 1, all deploys fail; set huge, the DoS reopens),
and changing a constant is a consensus change needing a binary upgrade — the right
ceremony for a safety rule. Note the budget is *calibrated against*
`PreprocessGasPerByte`, one-directionally and unenforced: raising that param only
makes the guard more conservative, whereas **lowering** it breaks parity and
`Validate` only requires it positive. If it drops, this budget must drop too.

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
filetest-harness-only, and splitting them off needs a deliberate definition of the
unsupported-Gno subset against ~500 filetests that pin both.
