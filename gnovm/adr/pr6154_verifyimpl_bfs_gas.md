# Metering the interface-satisfaction BFS (preprocess and runtime)

## Status

Implemented. Verified: gnolang unit + `-race`, sdk/vm gas, and integration.
(The only `TestFiles` failures are pre-existing `go/types` golden-text diffs,
present with and without this change.)

## The problem

`InterfaceType.VerifyImplementedBy` checks satisfaction by walking the target's
embedding graph once per interface method (`findEmbeddedFieldType` →
the embedding-graph BFS): O(methods × N), where N = reachable embedded types.

This CPU was unmetered on both surfaces:

- **Preprocess (deploy: `addpkg`/`run`).** Only the flat per-byte
  `chargePreprocessGas` applied, so a wide interface + wide/deep embedding forced
  millions of type-checker cycles for a few KB of source.
- **Runtime.** Type assertions (`x.(I)`, `x, ok := …`), type switches
  (`case I:`), and Stringer/error dispatch run the same walk; only the per-method
  part (`OpCPUSlopeTypeAssertIface × methods`) was charged — the per-field walk
  was free. (Method *dispatch* already charged its walk via
  `OpCPULazyBoundResolve`; assertion/switch did not.)

The unmetered walk predates the BFS. PR #5721 additionally replaced the old O(N)
DFS lookup with a spec-compliant BFS whose same-depth duplicate check was a
linear frontier scan — O(N²) worst case. Both are fixed here.

## Severity

High-leaning (liveness DoS / gas under-pricing) — not a chain halt or fund loss.

- Gas does **not** bound this work — that is the bug. The walk is unmetered, so a
  tx pays ~O(source) gas (alloc + per-byte) while doing up to O(methods × N²) CPU
  and *succeeds* cheaply rather than hitting OOG.
- Per-tx CPU is finite (bounded by source size and `MaxStructFields`), so no one
  tx is unbounded; but many cheap high-CPU txs fill a block whose gas limit does
  not reflect the real CPU → validators fall behind → liveness degradation.
- #5721's O(N²) sharpens it (small source → heavy CPU). Pinning high vs medium
  needs measuring the CPU-per-gas amplification on reference HW.

## The decision

1. **O(N²) → O(N), and one walk per check.** The BFS is now `embedWalk`: the
   linear same-depth duplicate scan is replaced by O(1) lookups in the walk's
   `seen` map (each type's index in the level under construction, or -1 once
   that level is scanned), and the walk is expanded lazily and **shared by all
   method lookups of one `checkImplementedBy` call** — the graph is built once
   and re-scanned per name, instead of rebuilt from the root per method (the
   report's "no caching between method lookups"). Expansion is billed once per
   type reached and each lookup per entry it scans (see 2), so the metering
   stays deterministic and tracks the work actually done (128×128 shape: see
   the benchmark figures in the PR). The walk keeps the old lookup's two
   allocation-free exits: a depth-0 hit, and a root that exposes no struct to
   expand (`rootSt == nil`), which returns before `seen`/`levels` are built.
   The latter is load-bearing rather than cosmetic — it is the primitive-root
   case, and `Fprint` probes Stringer and then error on *every* printed value,
   so dropping it cost 4 allocations and 344 B per probe (282 ns vs 113 ns for
   one probe on `string`) on the hottest formatting path in the chain, with no
   gas change to make it visible.
2. **Meter the walk** via `checkImplementedBy(gm, …)` → `chargeCPUGas(gm, …)`
   (a no-op when `gm` is nil): `OpCPUSlopeTypeAssertIface` (349) per method,
   `OpCPUSlopeEmbedExpand` (200) per embedded type the walk expands (once per
   check), `OpCPUSlopeEmbedScan` (25) per expanded entry each name lookup
   scans, and `OpCPUSlopeEmbedTrailHop` (135) per hop of a found name's trail
   (`buildEmbeddedTrail`, ~depth², once per hit). One body serves preprocess
   (gm = preprocess meter) and runtime (gm = `m.GasMeter`). Charging upfront
   from a one-shot type count was rejected: the walk early-exits at the
   shallowest providing level, so it does not visit every type; metering inside
   bills exactly what is traversed.

   **Calibration.** The three slopes are fitted on a depth × width × hit/miss ×
   methods grid (`BenchmarkOpEmbedWalk`) and converted to reference-ns with a
   machine factor measured directly, not by ratio to another constant:
   table ÷ `ns/op(pure)` over seven flat ops marked `ok` in
   `cmd/calibrate/op_bench_analysis.txt` (Add, Eql, IfCond ×2, SwitchClause,
   SwitchClauseCase ×2) gives 1.87–2.17 on the dev box; 2.1 is used. Each slope
   is taken at the deep end of the grid (`MaxEmbedDepth = 8`) and rounded up;
   in reference-ns the tool reads expand 161 (→ 200), scan 24.2 (→ 25; the
   per-level overhead lands per entry when a level is one type wide, which a
   wide-shape fit alone would miss), trail 131.5 per hop (→ 135). The trail is
   billed as its own per-hop
   term rather than folded into the scan slope because it scales with depth,
   not width: a single scan slope chosen at the deep end would overprice wide,
   shallow shapes ~10×, while a per-hop term prices depth where it occurs. All
   three remain floors: the 128×128 check charges 497K gas against ≈305K
   reference-ns of work, the deep chain (8×1, 16 methods) 27.7K against ≈18.8K.

   Two earlier drafts were replaced. A single 813-per-(method × type) charge,
   sized for per-method re-expansion, overpriced the shared walk ~10×. Its
   successor (560/75) was ratio-scaled off `OpCPUSlopeTypeAssertIface`, whose
   own fit predates the O(1) `methodIndex` lookup and so no longer measures
   what it did — the ×8.1 "hardware factor" that produced was mostly that
   staleness (review by @thehowl); the direct machine factor above replaces it.
   `OpCPUSlopeTypeAssertIface` itself is left as is (out of scope), noting that
   on the same box its per-method work measures ≈43 ns ≈ 90 reference-ns.

   The fit is reproducible by tooling: `cmd/calibrate/gen_analysis.py` has a
   section (2b) that reads the three slopes from `BenchmarkOpEmbedWalk` and
   compares them to the constants, and takes an optional measured-ns-per-
   reference-ns argument for non-reference hardware; the dev-box run behind
   these values is checked in as `cmd/calibrate/embedwalk_bench_m5_arm64.txt`.
   A reference-box run of `-bench='BenchmarkOp'` now regenerates them alongside
   every other op constant.

### Meter plumbing (and why not a global)

The meter is the tx's own — from the store's preprocess allocator (preprocess) or
`m.GasMeter` (runtime). An earlier attempt used a package-level
`var preprocessGasMeter`; that is wrong because preprocess runs concurrently (the
ABCI query connection has its own mutex, so `vm/qeval`/`vm/qrender` → `Preprocess`
run alongside `DeliverTx`). A global then **races** (confirmed by `-race`) and
**cross-bills** a query's work to a tx → non-deterministic `DeliverTx`, a
consensus hazard. So `gm` is threaded explicitly:

- **Preprocess**: every assignability helper that can reach the walk —
  `mustAssignableTo`, `checkAssignableTo`, `BinaryExpr.AssertCompatible`,
  `checkCompatibility` — takes the `store` its callers already hold (as
  `RangeStmt`/`AssignStmt.AssertCompatible` always did), and the meter is read
  once, at the single point of use, via `preprocessGasMeterOf(store)` in
  `checkAssignableTo`'s interface branch. No path can skip billing by passing a
  nil meter, and no call site has to know a meter exists. Paths covered:
  `checkOrConvertType` (`var x I = S{}`), the `CallExpr` interface-conversion
  branch and `convertConst` (`I(S{})`), `AssignStmt.AssertCompatible`
  (`a, b = f()`, `v, ok := m[k]`, type-assert-to-interface),
  `RangeStmt.AssertCompatible` (ranging a `map[I]V`), the embedded-varg argument
  loop, `specifyType` (generic native-func inference),
  `parseMultipleAssignFromOneExpr`, `BinaryExpr.AssertCompatible` (EQL/NEQ —
  `S{} == i` reaches the same walk), and `SelectorExpr` resolution.

  One redundancy had to go with it. For the `len(Lhs) == len(Rhs)` ASSIGN form
  (`x = S{}`), `AssignStmt.AssertCompatible` ran `mustAssignableTo` over every
  pair and then preprocess ran `checkOrConvertType` over every `Rhs[i]`, which
  does the same `mustAssignableTo` against the same `lt`. Free before, that
  became a doubled walk *and* a doubled charge: measured over a 32-method
  interface satisfied through 32 embedded types, `x = S{}` cost 103,813 gas per
  statement against 59,837 for `var x I = S{}`. `AssertCompatible` now only
  validates the left values (`evalAssignLhsType`, which is what its "assert
  valid left value" comment describes) and leaves the pair check to
  `checkOrConvertType`, which is reached for every such pair — it is that
  branch's `default:` case, and `AssignStmt.AssertCompatible` has no other
  caller. `evalAssignLhsType` is `evalStaticTypeOf` plus the blank-identifier
  and `assertValidAssignLhs` handling, so both sites derive the same `lt`.
- **Runtime** (gm = `m.GasMeter`): `doOpTypeAssert1/2`, `doOpTypeSwitch`,
  `Fprint` Stringer/error dispatch, `doOpStaticTypeOf`, both the bind-time
  (`getPointerToFromTV` VPInterface) and call-time (`resolveLazyBound` →
  `resolveInterfaceTrail`, whose per-hop `getPointerToFromTV` can re-enter the
  walk on a `VPInterface` hop) embedding walks of a lazy interface method value
  (whose `OpCPULazyBoundResolve` charge is per-hop and did not scale with the
  per-hop walk width), and the error-interface
  check in result formatting (`stringifyJSONResults`/`tryGetError` in
  `gno.land/.../convert.go`, via `IsErrorType(gm, …)`/`ImplError(gm)`). The
  last runs on a live, gas-metered machine after a successful query: its only
  production caller is `QueryEvalJSON` (the `vm/qeval_json` ABCI query, billed
  against the `maxGasQuery` meter), which passes no `*FuncType`. `MsgCall` does
  **not** reach this path — it formats results with `TypedValue.String()` →
  `ProtectedSprint`, which never dispatches Stringer/error — so the signature-
  based `IsErrorType` branch is exercised only by tests today. `m` is nil only
  on the marshal-only path where the value cannot be an error, guarded
  accordingly. Both metered checks sit inside `tryGetError`, *after* its
  `defer`/`recover`: each can panic `OutOfGasError`, and the whole point of
  that recover is that a late OOG degrades to "no `@error` field" rather than
  discarding the already-marshaled `Results` payload.
  A full sweep of `findEmbeddedFieldType` callers confirms the
  only remaining nil-meter site is the debugger's out-of-band expression
  evaluator, which is not a gas-metered path. (Two further unmetered
  satisfaction checks live in `debugAssertEqualityTypes` — `isImplementedBy`
  for `==`/`!=` operand typing — but they run only under the `if debug {…}`
  guard, which is compiled off on-chain, so they are not attacker-reachable.)
  The concrete assertion path folds its former upfront
  `m.incrCPU(349 × methods)` into `checkImplementedBy` (no double charge); the
  non-concrete fast-fail path keeps `m.incrCPU` (no walk runs). Both paths
  paying the per-method amount exactly once is pinned by
  `TestRuntime_TypeAssert_PerMethodChargedOnce`.

### No exported nil-metered wrappers

The first draft kept `VerifyImplementedBy`, `IsImplementedBy`, `IsErrorType`,
`ImplError` and `GetPointerToFromTV` as exported, meter-less conveniences
delegating with `gm = nil`. They were removed: a call site added later would
silently walk unmetered with nothing at the call saying so. `IsErrorType(gm, t)`
and `(*TypedValue).ImplError(gm)` now take the meter explicitly; the
package-private `checkImplementedBy` / `isImplementedBy` /
`getPointerToFromTV` are the only routes into the walk, so a `nil` meter can
only be written inside the package, where it is visible (tests, the debugger's
out-of-band evaluation).

## Consequences

- Deploy-time satisfaction is now gas-bounded on every path that reaches the
  walk — the assignment form and the conversion (`I(S{})`) and multi-assign
  (`a, b = f()`) forms alike (pinned by `TestPreprocess_IfaceImpl_*`,
  including the `ConversionMetered`/`MultiAssignMetered` regression tests, plus
  `wide_embed_verifyimpl.txtar`, whose tight-budget deploy OOGs only with the
  walk charge; the call-time
  lazy-bound walk additionally has the readable end-to-end
  `lazy_bound_verifyimpl.txtar`, whose tight `-gas-wanted` OOG fails without
  the expansion charge). All unit tests live in `iface_impl_walk_gas_test.go`, one
  file with a preprocess and a runtime section. The preprocess
  fixtures use a 1-method interface satisfied only by the last of 64 embedded
  types, so the per-method charge (349 × 20 checks ≈ 7K gas) stays far under
  the budget and only the walk ((200 + 25) × 64 × 20 + 20 × 135 ≈ 300K gas)
  can trip it — deleting the per-type expansion charge alone fails these
  tests, not just deleting all metering. `TestRuntime_TrailHop_Charged` pins
  the per-hop term on an 8-deep chain.
- **Runtime raises on-chain gas** for assertions/switches/formatting over
  embedding-satisfied types: the per-method amount is unchanged, only the
  previously-free walk is added. Pinned by the `TestRuntime_*` tests,
  incl. `ScalesWithIterations` (isolates the per-assertion runtime charge from
  the one-time preprocess cost).
- **`m.Cycles`** no longer counts the per-method assertion cost on the concrete
  path (moved to the meter). It is telemetry-only (`logTelemetry`), not
  gas/consensus — observability-only.

## Known gap, tolerated for now

The Go type-checker pass (`gotypecheck.go`, `go/types`) also verifies
interface satisfaction for the same source and is not gas-metered either.
`go/types` caches method sets, so its cost per assignment scales with the
method count, not methods × embedded fields — a far smaller amplification than
the one fixed here — and it runs under the same per-byte preprocess gas. It is
left as-is for now and recorded here so the gap is known; if it ever needs
closing, the fix is a bound or charge around `TypeCheckMemPackage`, not in the
Gno preprocessor.

## Follow-ups

- **Calibration.** The three `OpCPUSlopeEmbed*` constants are dev-box fits
  converted with a measured machine factor (see The decision, 2); a run of
  `BenchmarkOpEmbedWalk` on the gas-table reference hardware would confirm them
  directly, and `OpCPUSlopeTypeAssertIface` (349) deserves a refit of its own
  since `methodIndex`.
