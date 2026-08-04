# ADR: Source-Level Gas Profiler for GnoVM ("gas pprof")

## Status

Accepted — implemented.

## Summary

A source-level gas profiler for the GnoVM. A `GasMeter` decorator captures every
gas dimension (CPU, allocation, store I/O, amino, refunds) and an incremental
call-tree cursor attributes each charge to the current gno function; output is a
standard multi-dimension pprof profile viewable with `go tool pprof`. It is
driven from two surfaces: `gno test -gasprofile=<file>` (unit tests and
filetests) and the dev-only `.app/profiletx` ABCI query (reachable from the
command line via `gnokey maketx -gasprofile`, from Go via `gnoclient.ProfileTx`,
and exposed by gnodev). It is off by default and, when on, is
observation-only — it never changes gas charged or execution results. This
document records the design and the rationale; the code lives in
`gnovm/pkg/gasprof/`.

## Context

Gno developers today cannot answer the question *"where did my transaction's
gas go?"* at the level of their own code. The existing tooling profiles
adjacent things:

- **Go pprof** (`CPUPROFILE=` on the `gno` CLI, `-tags debug` HTTP server,
  node RPC profiler) attributes cost to *interpreter* functions
  (`doOpEval`, `runtime.scanobject`), not to gno functions.
- **`gnovm/pkg/benchops`** attributes wall-clock time to the 256 opcodes
  globally — no call-stack attribution, used for calibrating gas constants
  (`gnovm/cmd/calibrate`), not for profiling user code.
- **`-tags gastrace`** (`tm2/pkg/store/trace`, see `gno.land/adr/gastrace.md`)
  logs individual store-gas charges as flat text — no aggregation, no
  attribution to gno frames, store dimension only.
- **`gno test`** prints a single total (`--- GAS: n`, verbose mode only).

Nothing maps gas to gno source. PR #4557 proposed a CPU/memory profiler with
a bespoke interactive CLI; it was closed without merging and profiles time,
not gas.

Gas is the better profiling target for a smart-contract VM:

1. **Deterministic** — for a given execution the profile is exact and
   reproducible; no sampling, no noise. Every gas unit can be attributed
   to the exact frame that charged it.
2. **It is the quantity users pay for** — optimizing wall time that isn't
   gas-metered is optimizing the wrong thing.
3. **Already computed** — the VM charges gas at a small number of
   well-defined chokepoints; the profiler only *reads* amounts that are
   already being computed.

### Prior art

- **Aptos** (`--profile-gas`): flame graphs split by dimension
  (execution+IO vs. storage), plus an indented full execution trace.
  Closest existing system to this design.
- **EVM**: `debug_traceTransaction`/`callTracer` (per-call gas on real
  txs) and Foundry `forge test --gas-report` / `--flamegraph` (dev-time).
  Known pitfall: pre-subtracted refunds make child calls appear to cost
  more than their parents (foundry#13262) — this design avoids that class
  of confusion explicitly (see *Refunds*).
- **Cosmos SDK / CosmWasm**: descriptor-string gas meter only; we found
  no source-level gas profiler in the Cosmos lineage (other chains such
  as NEAR and Solana offer only categorical or log-based breakdowns).

## Goals

- Attribute **every unit of gas** charged during gno execution to the gno
  call stack (function-level first; line-level as a refinement).
- Cover **all three gas dimensions**: CPU cycles, allocation, store I/O +
  amino (store is 40–70% of typical realm-call gas, per the gastrace ADR).
- Emit **standard pprof protobuf** (`profile.proto`) so the existing
  ecosystem — `go tool pprof -http`, flame graphs, `top`, `list`, diff
  profiles — works with zero custom UI.
- **No measurable impact when disabled** (one nil-check per frame
  push/pop; no per-op work). **Zero gas impact when enabled** on any
  surface where gas is already charged (the tx path, Phase 3) —
  observation only; profiling must never change gas usage or execution
  results there. The Phase-2 dev surface is the deliberate exception: it
  *adds* store metering (and routes allocation gas through the test meter)
  that plain `gno test` omits entirely, so profiled local totals include
  `store_gas`/`alloc_gas` dimensions unprofiled runs don't charge and a
  correspondingly higher `--- GAS:` total.
- Dev-first surface: `gno test`, where profiles are exactly reproducible
  run-to-run. CPU and alloc gas match on-chain costs for the same
  execution path — except `"GC"` gas, which depends on the allocator
  budget (unbounded locally vs. capped on-chain); store gas additionally
  depends on backing-store state (cache hits, dedup, depth estimator), so
  parity there is approximate by nature.

## Non-goals

- Wall-clock/CPU-time profiling (that is Go pprof's job, and PR #4557's).
- Replacing `benchops`/`calibrate` (they calibrate the gas *model*; this
  profiles gas *spent by user code*).
- A bespoke interactive CLI (pprof already provides one).
- Consensus-visible or on-by-default behavior of any kind.

## Decision

A meter-decorator plus frame-lifecycle cursor hooks, emitting standard
pprof protobuf with one value index per gas dimension.

### Verified foundation: one meter sees all three dimensions

All three gas dimensions funnel into a **single `GasMeter` instance** in
the transaction path, each with a dimension-identifying descriptor. The
tx-path timeline (verified on current master):

1. The **ante handler** creates the per-tx meter
   (`ctx.WithGasMeter(store.NewGasMeter(gasLimit))`,
   `tm2/pkg/sdk/auth/ante.go`) and charges `"txSize"` and
   `"ante verify: …"` to it.
2. The **beginTxHook** (`gno.land/pkg/gnoland/app.go` →
   `MakeGnoTransactionStore`) threads that same meter into the gno
   transaction store: `BeginTransaction` receives a `store.GasContext`
   (store I/O gas: `"DepthReadFlat"`, `"DepthSet"`, `"DepthDelete"`, …)
   and a `gasMeter` (amino encode/decode gas), both carrying
   `ctx.GasMeter()` (`gno.land/pkg/sdk/vm/keeper.go`,
   `newGnoTransactionStore`).
3. **Message handlers** construct machines with
   `GasMeter: ctx.GasMeter()`; `NewMachineWithOptions` propagates it to
   the allocator (`mm.Alloc.SetGasMeter`, machine.go), covering
   `"CPUCycles"` and `"memory allocation"`. Preprocess allocators also
   receive `ctx.GasMeter()`.

So ante, CPU, alloc, store I/O, and amino charges all hit one instance.
(The block gas meter is a separate meter charged post-hoc in baseapp and
is out of scope.)

**Consequence:** the profiler needs no per-charge-site instrumentation —
it is built from exactly two mechanisms:

### Component 1 — observing meter (dimension capture)

The `meter` type (installed via `WrapMeter`, in `gnovm/pkg/gasprof/gasprof.go`)
wraps the real `GasMeter` — a decorator implementing the same interface. It
records each charge into the profiler's current call-tree node, bucketed by
descriptor, then delegates to the wrapped meter unchanged. Consume and refund
use two different recording strategies, each chosen so the recorded amount
equals what the meter actually applied:

- **`ConsumeGas` records the passed `amount` *before* delegating.** This is
  exact because `ConsumeGas` never clamps — `amount` is always the applied
  delta. Recording first also captures the final charge on the out-of-gas
  path: the underlying meter mutates its consumed counter and *then* panics, so
  a record-after wrapper would miss that aborting charge on precisely the "why
  did my tx run out of gas" runs. (The one case where `amount` ≠ applied delta
  is int64 overflow, which panics *before* mutating and requires ~9.2e18 gas —
  terminal and unreachable in practice.)
- **`RefundGas` records the *applied delta*** (a `GasConsumed()` before/after
  snapshot), because refunds clamp to the consumed total, so the requested and
  applied amounts can differ.

**Installation is a single seam.** The wrapper is installed **once per
transaction**, where the tx meter enters gno wiring (the
beginTxHook/`MakeGnoTransactionStore` boundary for on-chain use; the
meter-construction sites in `gnovm/pkg/test` for the dev surface).
Concretely, on-chain: the wrapper replaces the meter in the sdk.Context
(`ctx.WithGasMeter(wrapper)`) at the top of `MakeGnoTransactionStore`,
*before* `newGnoTransactionStore` — the store's GasContext, the message
handlers' machines, and the preprocess allocators all read
`ctx.GasMeter()` independently, and all must see the wrapped instance. At
install time the wrapper snapshots the already-consumed amount (ante
charges pre-date it) and books it as a synthetic `(ante)` root node. Rule:
**exactly one wrapper instance per meter** — wrapping a wrapper
double-records; the installer must guard against it.

Descriptor→dimension mapping is a static table over the (small, verified)
repo-wide descriptor inventory:

| Dimension | Descriptors |
|---|---|
| `cpu` | `"CPUCycles"` (incrCPU + native gas via uverse), `"GC"`, `"parsing"`, `"ComputeMapKey"`, `"stream output"` |
| `alloc` | `"memory allocation"` |
| `store` | `"DepthReadFlat"`, `"DepthSet"`, `"DepthDelete"`, `"ReadFlat"`, `"ReadPerByte"`, `"WriteFlat"`, `"Delete"`, `"IterNextFlat"`, `"ValuePerByte"`, `"AminoEncodePerByte"`, `"AminoDecodePerByte"` |
| `other` | `"txSize"`, `"ante verify: …"`, and any descriptor not in the table |

The `refund` dimension is not descriptor-mapped: refund gas is booked
directly by the `RefundGas` path (see Component 1), not via this table.

Unknown descriptors map to `other` rather than being dropped, so new
descriptors can't silently leak gas out of the profile, and the
reconciliation invariant (below) stays intact.

### Component 2 — call-tree cursor (attribution)

The profiler maintains a call tree mirroring the gno call stack:

- **Node identity**: callee `*FuncValue` → pkgpath + function name + line
  of definition (from `fr.Func.Source.GetLocation()`) and the file
  *basename* (`fr.Func.FileName`; full paths would be non-deterministic).
  The name mirrors what `Machine.Stacktrace()` renders. Because only the
  basename is recorded, `go tool pprof -list` needs `-source_path` (see
  `docs/resources/gno-testing.md`).
- **Descend** on `PushFrameCall`; charges recorded by Component 1 accrue
  to the cursor's node as *flat* gas; cumulative gas is the subtree sum,
  computed at emission time.
- **Ascend** inside `PopFrame` itself — *after* `maybeFinalize` runs
  (the realm-finalization step in the return sequence that persists a
  crossing realm's objects), so those finalize-time store writes attribute
  to the frame whose realm is being finalized.

Cost: O(1) per *call* (calls are orders of magnitude rarer than ops),
plus one nil-check per frame push/pop when profiling is off.

**Cursor integrity — the cursor must survive unbalanced unwinding.** The
VM does not guarantee one pop per push:

- `PopUntilLastReviveFrame` truncates any number of call frames in a
  single slice assignment (machine.go) when a panic unwinds to a *revive
  frame* — a panic-recovery boundary that can drop several call frames at
  once, rather than one return at a time.
- Machine aborts (unhandled gno panics, alloc-limit panics, out-of-gas
  panics) escape the op loop entirely and reach `Release()`, which resets
  `m.Frames` wholesale — no pops at all. Failed txs are a primary
  profiling target, not an edge case.

Therefore the cursor does **not** rely on balanced push/pop events. It
tracks call frames (frames with `Func != nil`, i.e. `IsCall()`)
exclusively, driven by two kinds of hook:

- **`Enter` / `Pop`** — the O(1) common path. `PushFrameCall` descends the
  cursor one node; `PopFrame` (which `PopFrameAndReturn` and
  `PopFrameAndReset` route through) ascends it one node. Both are guarded
  by `IsCall()`.
- **`SyncDepth(n)`** — the bulk path. `PopUntilLastReviveFrame` truncates
  several call frames at once (revive/panic-recovery unwinding), so the
  cursor re-syncs absolutely: it ascends until its depth matches the
  machine's post-truncation call-frame count.

`GotoJump` and `PopUntilLastCallFrame` trim only non-call (loop/block)
frames — the cursor does not track those, so they need no hook (see the
comments at those call sites). At `Release()` the cursor resets to the
root. **Emission tolerates a non-root cursor**: attribution happened at
charge time, so the tree is valid even when the tx aborted mid-call.

**Sub-machines** (preprocess machines spawned mid-execution, the
keeper's sequential run machines): each Machine has its *own* `Frames`
stack, so sub-machine frame events must not drive the parent's cursor.
Sub-machine charges are attributed to the parent's current node (they
share the meter, hence the wrapper), optionally under a synthetic child
marking the boundary (e.g. `(preprocess)`). Per-machine cursor segments
are a possible refinement, not required for correctness.

**Charges outside any frame** (tx finalization / store writes at commit,
preprocess/type-check gas, package init) attribute to the synthetic
`(root)` node directly, since no call frame is on the cursor when they are
charged. The one exception booked as a distinct synthetic child today is
`(ante)`: the on-chain path snapshots `GasConsumed()` at meter-install time
and books it via `Profiler.Book("(ante)", …)` so the tree reconciles with
the meter. Real-world realm txs spend heavily in finalize-time store
writes; those currently show up under `(root)` rather than a dedicated
`(finalize)` node. Finer synthetic labels (`(finalize)`, `(preprocess)`,
`(init)`, per-file preprocess) are an open refinement, not yet implemented.

**GC**: gas charged during a GC sweep (descriptor `"GC"`) attributes to
the frame that triggered the sweep ("trigger pays") — same semantics as
Go CPU profiles attributing GC to the allocating stack. Documented, not
hidden.

### Dimensions and output schema

One pprof profile, **multiple value indices** (Aptos's
execution-vs-storage split, done the pprof-native way):

```
sample_type = [
  {type: "cpu_gas",    unit: "gas"},
  {type: "alloc_gas",  unit: "gas"},
  {type: "store_gas",  unit: "gas"},
  {type: "other_gas",  unit: "gas"},   // ante + unmapped descriptors
  {type: "refund_gas", unit: "gas"},   // positive magnitudes, separate index
  {type: "total_gas",  unit: "gas"},   // cpu+alloc+store+other (gross, pre-refund)
]
default_sample_type = "total_gas"      // set explicitly, not by position
```

- `go tool pprof -sample_index=store_gas -http=:8080 gas.pprof` → storage
  flame graph.
- pprof `Function`/`Location`/`Line` map to gno pkgpath / file / line.
  Function-level attribution ships first; line-level (recording
  `m.Lastline` — the VM's currently-executing source line — at charge
  time) is a follow-up that fits the same schema and unlocks
  `pprof list <fn>` annotation.

### Refunds

gno's cache store issues gas **refunds** on write dedup (`RefundGas`,
descriptor `"Refund"`). Netting refunds into the tree is how Foundry
ended up with "child costs more than parent" confusion, and negative
flat values render badly in flame graphs.

**Decision: refunds are a separate, positive-valued dimension**
(`refund_gas`), attributed to the node whose action triggered the refund.
All flame-graph-facing indices stay non-negative; net cost is
`total_gas − refund_gas`, and both are separate pprof value indices so
`go tool pprof -sample_index=` can show either.

### Safety invariants (consensus)

1. **Observation only.** The wrapper delegates to the real meter with
   unchanged semantics; the profiler never calls `ConsumeGas`/`RefundGas`
   itself. Charged gas with profiling on ≡ charged gas with profiling
   off, by construction.
2. **Off = inert.** When not profiling, the real meter is installed
   directly (no wrapper, no indirection) and cursor hooks are behind a
   single `m.gasProfiler != nil` check per frame push/pop/release. No build
   tag required — the hooks are per-call, not per-op, unlike benchops.
3. **Never on a validator path by default.** Activation is explicit:
   a `MachineOptions.GasProfiler` option / CLI flag. Phase 3's node
   tracer follows the existing `Unsafe*` RPC convention (dev nodes only).
4. **Reconciliation.**
   `sum(tree: cpu+alloc+store+other) − sum(tree: refund) ==
   meter.GasConsumed()`. This holds by construction: the `(ante)` node's
   value equals the install-time `GasConsumed()` snapshot and lives inside
   the tree's `other` bucket, and recording captures the applied
   (post-clamp) amounts. Tests assert it end-to-end, including for aborted
   (out-of-gas) runs; a debug-build assertion at emission time is a
   possible future hardening.

### Surfaces (phased)

An honest note on store gas: plain `gno test` calls
`BeginTransaction(tcw, tcw, nil, nil)` (gnovm/pkg/test/test.go) — nil
GasContext and nil amino meter — so it charges **no** store gas, and its
totals are unchanged by this work. Store gas is wired **only under
`-gasprofile`** (Phase 2, unit-test surface). `gno run` still constructs
its machine with no `GasMeter` at all and is not yet a profiling surface.
These are wiring facts about the surfaces, not the profiler design.

| Phase | Deliverable | Notes |
|---|---|---|
| 0 (done) | CPU-only vertical slice (spike, throwaway) | Validated frame-identity legibility, cursor resync, and pprof rendering on a real package before hardening. |
| 1 (done) | `gno test -gasprofile=<file>`: cpu + alloc dimensions | Meter decorator + incremental cursor; wraps the test `NewInfiniteGasMeter()` and rewraps the allocator's meter; writes a 6-value pprof. Store columns present but zero. |
| 2 (done) | Store dimension on `gno test` | `gasProfileStoreMeters` wires a `GasContext` (`DefaultGasConfig`) + amino meter, both backed by a wrapper around the **shared** profiler, into the test store's `BeginTransaction` — **only when profiling** (normal `gno test` totals unchanged). Because the shared store and per-test machines hold separate meters, store gas rides its own infinite meter and records to the shared profiler at the current cursor. Measured: store is ~99% of gas on a store-heavy package, mostly at `(root)` (package load/persist charged outside test frames). **Filetests** also covered: they share one meter between store and machine, so the meter is wrapped *before* the store is built and the cursor enabled via `MachineOptions.GasProfiler` (filetest.go). (`gno run` still has no meter — not a profiling surface.) |
| 3a (done, engine) | Node/keeper tx tracer — engine | `vm.WithGasProfile(ctx)` marks a ctx; `MakeGnoTransactionStore` wraps `ctx.GasMeter()` once (all dimensions share it on-chain) and books the pre-install consumed as `(ante)`. `MachineOptions.GasProfiler` drives the cursor on the top-level tx machines (Call/Run/AddPackage); preprocess sub-machines inherit the wrapped meter but not the cursor, so their gas attributes to the parent. Verified end-to-end: a real Call captures cpu+alloc+**store** and reconciles exactly with the tx meter. Off unless `WithGasProfile` is called. |
| 3b (done, query) | Node/keeper tx tracer — surface | `.app/profiletx` ABCI query returns a pprof profile of a tx's gas. tm2 stays profiler-free: `BaseApp.Simulate` gained variadic `ContextFn`s and a nil-by-default `TxProfiler` hook; `AppOptions.EnableGasProfiler` (off by default; dev nodes only) registers a closure that runs `Simulate` with a `WithGasProfile` ctxFn and returns `WritePprof` bytes. Verified end-to-end: query a Call tx → gzipped pprof naming the function + cpu/store dimensions; rejected when disabled. Ergonomics wired: `gnoclient.ProfileTx`, gnodev enables it by default, and `gnokey maketx -gasprofile <file>` (all subcommands, same flag name as `gno test -gasprofile`) signs the tx and writes the pprof instead of broadcasting, reporting explicitly that the tx was not sent. |
| 3.1 (not pursued) | Historical tx re-execution | A faithful `debug_traceTransaction` (reproduce a past tx's *actual* gas) is **not achievable** in gno: the VM/gas rules live in the binary, not versioned per height, so replaying an old tx with the current binary applies today's rules to old state. gno's own hardfork/genesis replay confirms this — it bypasses the gas meter (`GasReplayMode="source"` → `SkipGasMeteringKey`) and compares against *recorded* source gas rather than reproducing it. Historical *state* is retrievable within the prune window, but faithful trace would need a per-height ruleset (a chain-design change, out of scope). What we built is the `debug_traceCall` analog: trace under current rules. |

### Code layout

```
gnovm/pkg/gasprof/gasprof.go  # the whole profiler: call tree + cursor,
                              # the `meter` decorator (WrapMeter), descriptor
                              # table, and hand-encoded pprof emission
gnovm/pkg/gnolang/machine.go  # MachineOptions.GasProfiler, cursor hooks in
                              # PushFrameCall/PopFrame/PopUntilLastReviveFrame/
                              # Release, Attach/SetGasProfilerCursor (nil-guarded)
gnovm/pkg/test/               # gno test -gasprofile wiring (test.go + filetest.go)
gnovm/cmd/gno/test.go         # -gasprofile flag
gno.land/pkg/sdk/vm/keeper.go # WithGasProfile + meter wrap at MakeGnoTransactionStore
gno.land/pkg/gnoland/app.go   # AppOptions.EnableGasProfiler + .app/profiletx closure
tm2/pkg/sdk/                  # Simulate ctxFns + TxProfiler hook + profiletx query
tm2/pkg/crypto/keys/client/   # gnokey maketx -gasprofile flag (signTx + ProfileTx)
gno.land/pkg/gnoclient/       # ProfileTx client method
```

## Alternatives considered

- **Revive PR #4557 (CPU/mem time profiler + custom CLI).** Profiles the
  wrong quantity (nondeterministic time vs. deterministic gas) and
  builds/maintains a bespoke UI that pprof provides for free.
- **Per-charge-site hooks instead of a meter wrapper.** More invasive
  (touches machine.go, alloc.go, store.go, cache store), easy to miss a
  site, and unnecessary given the verified single-meter wiring. The
  wrapper plus delta recording makes reconciliation structural.
- **Extend `-tags gastrace` text logs with stack info.** Text logs don't
  aggregate; store-only; no ecosystem tooling.
- **Extend `benchops`.** Opcode-indexed, per-op timing focus, build-tag
  UX, no call attribution; its file exporter is already half-orphaned
  (cmd/benchops was removed in favor of cmd/calibrate).
- **Custom JSON/flamegraph output.** pprof protobuf is strictly more
  capable (multi-dimension switching, diff profiles, `list`, web UI) at
  lower cost.
- **Netting refunds as negative flat gas.** Rejected — see *Refunds*.

## Questions resolved during implementation

1. **Frame-identity legibility** — the Phase-0 spike confirmed closures,
   deferred calls, bound methods, and init functions render legibly.
   Identity is the fully-qualified name built from `FuncValue`
   (`stacktraceFuncName` + pkgpath), the same as `Machine.Stacktrace()`.
2. **pprof emission dependency** — hand-encoded `profile.proto` (a ~40-line
   protobuf writer in `gasprof.go`); no new dependency, validated against
   `go tool pprof`.
3. **Line-level granularity** — not built. Attribution is function-level;
   line-level (recording `m.Lastline` per charge) remains a follow-up that
   fits the same pprof schema.
4. **Test-surface `GasConfig`** — the dev surface uses tm2's
   `DefaultGasConfig`. Because the test base store is flat (no IAVL depth
   estimator) and chains can override gas params, local store gas is an
   approximation, documented as such (good for relative comparison, not
   exact on-chain prediction).
5. **Third-party viewers** — `go tool pprof` handles the multi-valued
   profile natively (the primary target). speedscope and others are less
   uniform with multiple sample types; not a blocker.
6. **Historical tx re-execution** — not pursued; a faithful
   `debug_traceTransaction` is not achievable in gno (see the "3.1" row in
   the Surfaces table).

## Consequences

**Positive**
- First source-level gas profiler we know of in the Cosmos/tm2 lineage;
  comparable to Aptos's `--profile-gas`.
- Deterministic, exact profiles — every gas unit accounted for
  (including ante and finalize residuals), reconciled against the meter.
- No custom UI to maintain; full pprof ecosystem (flame graphs, diff
  profiles between contract versions, per-line annotation) for free.
- Foundation for a Phase-3 "why did this mainnet tx cost X" node tracer.

**Negative / risks**
- Four small permanent hooks in machine.go (`PushFrameCall`, `PopFrame`,
  `PopUntilLastReviveFrame`, `Release`), all nil-guarded.
- The descriptor→dimension table is hand-maintained: a new gas descriptor
  not added to it falls into `other` (reconciliation still holds; only the
  store/cpu split would be off). A test pins the current mapping so a
  regression fails loudly.
- Meter-wrapper indirection during profiled runs only; profiled runs are
  slower (acceptable — dev tool).
- Sub-machine and unwind edge cases were the likeliest source of
  implementation bugs; mitigated by depth-based resync, `Release` reset,
  and the reconciliation invariant, and covered by adversarial review at
  each phase.
