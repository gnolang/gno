# Per-machine runtime block pool

## Context

After the byte-access fixes (#5812), `NewBlock` was the single dominant
allocation site in the VM — ~45% of remaining heap objects on the bytes
stdlib suite (≈135M: per-scope blocks from `doOpExec` for
if/for/range/switch/block, call blocks from `doOpCall`), each also allocating
a `[]TypedValue` for the block's slots. Every block was thrown away the
instant it was popped from the machine's block stack and re-allocated next
time the same scope was entered.

These blocks are short-lived and structurally identical run-to-run, which
makes them an ideal pooling target — *if* a popped block is provably dead.

## The safety invariant (why a popped block is dead)

A runtime scope/call block is **never reachable from anything that outlives
its time on the machine's block stack**, so once popped it can be zeroed and
reused. This rests on pre-existing properties of the heap-items design, not on
anything this PR introduces:

- **Closures don't reference blocks.** `doOpFuncLit` sets `FuncValue.Parent =
  nil` and copies the captured `HeapItemValue`s into `FuncValue.Captures`. A
  closure stored in realm state references heap items, never the scope block
  it was created in.
- **`&local` is heap-promoted.** The preprocessor promotes any address-taken
  or closure-captured local to a `HeapItemValue`; `GetPointerToMaybeHeapDefine`
  returns a `PointerValue` whose `Base` is the `*HeapItemValue`, never the
  `*Block`. A `PointerValue{Base: *Block}` is only produced for transient,
  non-escaping access (deref-and-discard).
- **Frames/stacktraces store indices and locations, not block pointers.**

Empirically confirmed: persisting `&localVar` from inside a `for`-block and
walking the realm object graph shows only the package block (`*PackageNode`),
file block (`*FileNode`), the `FuncValue`s, and `HeapItemValue`s — **no
runtime scope block is ever persisted**. The two block kinds that *are*
persisted (file/package) are excluded by Source type (below).

## Decision

`Machine.acquireBlock` / `releaseBlock` (machine.go) implement a per-machine
LIFO pool (`Machine.blockPool`, bounded by `blockPoolLimit = 32`), routed
through every block discard site: `OpPopBlock`, `GotoJump`,
`PopFrameAndReset`/`PopFrameAndReturn`, `PeekFrameAndContinueFor`/`Range` (all
via `releaseBlocksFrom`). Released blocks are zeroed (`*b = Block{...}`) so
they retain no references. Scope blocks (op_exec.go) and call blocks
(op_call.go `doOpCall`/`doOpReturnCallDefers`) acquire from it.

Gas and VM-GC accounting are independent of pooling: `acquireBlock` charges
`Allocator.AllocateBlock(numNames)` on both the hit and miss paths exactly as
`Allocator.NewBlock` does, and the VM-GC counts `len(b.Values)` — neither
depends on whether the `*Block` was recycled or on its slice capacity.

### Exclusions — what `releaseBlock` must never pool

`releaseBlock` skips four populations that also travel the block stack:

1. **node-owned static blocks** (Eval/RunStatement push them) — identified by
   static-block identity, `b.Source.GetStaticBlock().GetBlock() == b`;
2. **file/package blocks** (referenced by `FuncValue.Parent`) — identified by
   `Source` node type (`*FileNode`, `*PackageNode`, `RefNode`, `nil`);
3. **defer-site blocks** — `Defer.Parent` is visited by the VM-GC until the
   defer runs, so `doOpDefer` marks the block with a `noRecycle` flag;
4. **realm-attached blocks** — see the recycle guard below.

Panic unwinding (`m.Exception != nil`) skips pooling entirely as cheap
conservatism (the exception path is cold).

### Pooling during realm execution — the recycle guard

The realm-attached exclusion is `oi.ID.IsFinalized() || oi.GetIsNewReal()`.

The obvious-looking `!oi.ID.IsZero()` is wrong here. `stampPkgID` stamps every
block with the executing realm's PkgID at allocation, so under any realm every
block has a non-zero `ID` and `!IsZero()` is always true — which silently
makes the pool **inert during realm (`/r/`) execution**, recycling ~0 blocks.
(An early iteration of the PR shipped exactly that, and the pool helped only
non-realm stdlib tests.)

Only *finalized* blocks (`NewTime != 0`, set by `assignNewObjectID` at
finalize) are actually persisted, and `GetIsReal()` is defined as exactly
`ID.IsFinalized()`. `GetIsNewReal()` additionally covers the mid-transaction
window where a block has been marked reachable from the realm graph but
`assignNewObjectID` has not yet run — a window `GetIsReal()`/`IsFinalized()`
do *not* catch. Excluding "finalized or new-real" rather than "any realm
PkgID" lets the pool fire during realm execution, which is safe by the
invariant above (scope/call blocks never enter either state — they are never
in the realm graph; only file/package blocks are, and those exit at exclusion
2). A `debugAssert`-gated check in `releaseBlock` panics if a block being
pooled is still referenced as a pending `Defer.Parent`, pinning the invariant
in CI builds that enable the tag.

### `noRecycle` lives in `Block`, not `bodyStmt`

The defer-site `noRecycle` flag is a dedicated `Block` field
(`Block.setNoRecycle`/`isNoRecycle`). An earlier iteration hid it in
`bodyStmt`'s trailing padding to keep `unsafe.Sizeof(Block{})` — and the
`_allocBlock` gas constant — unchanged. That was a latent
use-after-recycle bug: a `fallthrough` re-enters the next clause body by
wholesale-reassigning `b.bodyStmt = bodyStmt{...}` (op_exec.go) *after* a
`defer` in the prior clause body already set the flag, silently clearing it —
so the switch block, still referenced as `Defer.Parent`, became eligible for
recycling.

The bug is not behaviorally observable (the deferred call resolves through
`FuncValue.Parent`/`Captures`, never `Defer.Parent`, which only feeds the
VM-GC reachability re-count, and that dedups via `GCCycle`), which is why it
survived — but it is a real invariant violation, and the `IsFinalized` guard
above removes the secondary masking that had hidden it under a realm. Moving
the flag to a real `Block` field fixes it structurally: control-flow
`bodyStmt` resets can no longer touch it, so op_exec.go's `FALLTHROUGH` needs
no special case. The cost is `sizeof(Block)` 528 → 536 and `_allocBlock`
528 → 536 (the alloc-constants init assert enforces the match). Regression
test: `gnovm/tests/files/fallthrough0.gno`, which trips the `debugAssert`
invariant without the fix.

### Uniform block size — cap 14

The pool holds **uniformly cap-14** blocks. `newPooledBlock`
(alloc.go → `newBlockWithValueCap`) allocates `make([]TypedValue, numNames,
max(numNames, blockPoolValueCap))`; `releaseBlock` only pools blocks of
`cap ≥ blockPoolValueCap`, re-sliced to exactly `[:14:14]`.

The pool is a LIFO stack and `acquireBlock` inspects only the top block, so a
too-small top forces a miss even when a larger block sits deeper. A uniform
size removes that: every demand ≤ 14 hits the top.

`14` is chosen from the Go 1.26 allocator size classes, not arbitrarily. A
`[]TypedValue` (40 B/elem) is **scannable** (TypedValue holds two interface
pointers), so a backing array over `MinSizeForMallocHeader = 512 B` gets an
8 B malloc header before size-classing. cap 13 (520 B) and cap 14 (560 B)
both land in the **576-byte size class** (576 − 8 = 568 usable = 14 slots); 15
jumps to 640. The observed block-size distribution — `numNames` gathered
across TestFiles + TestStdlibs, 443M samples — is `p50 = 0, p95 = 5,
p99 = 13, max = 35`, with a distinct 8M-block cluster at exactly 13. That
cluster forces the 576 class regardless, so cap 14 is the largest capacity
"free" within it and covers ≈ 99.3% of blocks.

A full pool pins ≤ `32 × 576 B ≈ 18 KB` of Go heap. `GarbageCollect` counts
that against the alloc budget (each pooled block by its retained capacity),
so parking a block in the pool does not appear to free its memory — the
allocation *gas* of (re)creating it is also charged on the miss path; see
below.

## Gas: recycling is cheaper than allocating

Block creation gas reflects the work actually done, and differs by path. This
is consensus-safe because it is deterministic: the per-machine pool starts
empty on every run (`Machine.Release` clears `blockPool`), so the hit/miss
sequence is a pure function of execution, identical on every validator.

- **Setup/recover CPU** (`OpCPUAcquireBlock = 100`) is charged on *both* paths
  in `acquireBlock` — the slot re-slice, `initHeapItems`, field writes,
  `stampPkgID`. Heap items, if any, are charged separately by `initHeapItems`.
- **Allocation gas** (`AllocateBlock`, which models malloc + zero-fill CPU) is
  charged *only on a miss*. A recycle reuses memory and performs no malloc, so
  charging it the allocation cost — as the original pool did — was wrong.
- A miss charges `AllocateBlock(max(numNames, 14))`: a pooled block's `Values`
  is physically sized to cap 14, so a small block costs the same malloc as a
  14-slot one (that is what is allocated).

Because block creation is now charged explicitly in `acquireBlock`, the
enclosing-op constants were re-derived to *exclude* it (they had been measured
on the always-allocating pre-pool code): `OpCPUCall` 310 → **40** and
`OpCPUReturnCallDefers` 724 → **215** (block creation was ~90% of a 0-param
call's measured cost). Scope blocks were never charged it — `OpCPUExec` was
calibrated on `EmptyStmt` — so `OpCPUAcquireBlock` is purely additive there.
Constants are anchored to the Xeon-8168 reference (1 gas = 1 ns) by ratio
against `BenchmarkOpAdd_Int`; the `Benchmark{OpAcquireBlock*,OpCallWarm,
OpReturnCallDefersWarm}` calibration benchmarks reproduce them.

Net for a 0-param call: recycle `OpCPUCall(40) + OpCPUAcquireBlock(100)` = 140;
miss adds `AllocateBlock(14)` ≈ 263. Recycling is cheaper by the avoided
malloc.

## Consensus-visible change

Two changes shift gas: the `_allocBlock` 528 → 536 bump (the `noRecycle`
field, +8 B/block) and the recycle/allocate gas split above. Goldens
regenerated under `-test.short`:

- `gas/*` and `alloc_*` filetests (e.g. `gas/const` 2343 → 2284; `alloc_*`
  `MemStats` shift as recycled blocks stop charging phantom malloc);
- five integration `GAS USED:`/fee values: `gc`, `gnokey_gasfee`,
  `simulate_gas`, `stdlib_ibc_crypto_determinism`, `stdlib_restart_compare`
  (`addpkg` gas is preprocess/storage-bound and unchanged; only `maketx call`
  runtime gas shifts).

No other goldens change.

## Benchmark proof

Ordinary gno programs run through the VM (`BenchmarkBenchdata`), this branch
vs `master`, benchstat n=6:

| | sec/op | B/op | allocs/op |
|---|--:|--:|--:|
| `fib` (geomean of params) | −33% | −96% | −75% |
| `matrix` (geomean of params) | −15% to −22% | −79% to −90% | −42% to −60% |
| `loop` (alloc-free) | +4.8% | ~ | ~ |
| **geomean (all)** | **−21.5%** | **−89%** | **−59%** |

`fib`/`matrix` allocate a block per call/scope, so the pool removes ~10× of
their allocated bytes and cuts wall-time up to a third. `loop.gno` allocates
nothing, so it sees no benefit and pays a marginal +4.8% for the pool check on
an alloc-free hot loop.

For the stdlib bytes suite (where the pool already fired pre-realm-guard):
heap objects 300M → 165M; `TestStdlibs/bytes` solo 105.2s → 94.9s. Enabling
realm-execution pooling adds `TestFiles` −17.6% and `TestStdlibs` −4.9%.

## Negative results (measured, not adopted)

- **Carry the pool across machine reuse via `machinePool`** (reused machines
  keep a warm pool): 10–25% slower on parallel workloads — extra live heap
  across GC cycles, lost cache locality — without helping machine-churn
  workloads.
- **A global `sync.Pool[*Block]`** (which uniform sizing makes clean — one
  pool, every block serves any demand ≤ 14): prototyped and A/B'd with
  benchstat (n=15). It **regresses the warm/heavy path +11.3%** (the path that
  dominates CPU/gas) and only wins short-`main` churn (−12.4%, microseconds
  against millisecond tx overhead); net geomean +0.88% time, −14% B/op, −4.6%
  allocs. `sync.Pool` Get/Put (per-P shard + atomic + interface) loses to a
  slice push/pop on a ~440M-acquire hot path, and its GC-clearing reintroduces
  cold-start misses. **The pool stays per-machine.** If short-tx throughput
  ever becomes a measured bottleneck, the next step is a hybrid (per-machine
  fast path donating to a GC-cleared global stash), gated on evidence.

## Verification

```sh
go build ./gnovm/pkg/gnolang/
go test ./gnovm/pkg/gnolang/ -run 'Files$' -test.short -count=1 -timeout=600s
go test -tags debugAssert ./gnovm/pkg/gnolang/ -run 'Files$/fallthrough0' -count=1
go test ./gno.land/pkg/sdk/vm/ -run Gas -count=1
go test ./gno.land/pkg/integration/ -run TestTestdata -count=1
go test -run=NONE -bench=BenchmarkBenchdata -benchmem -count=6 ./gnovm/pkg/gnolang/
```

`Files -test.short` shows only the pre-existing Go-toolchain type-checker
wording diffs; the `debugAssert` `Defer.Parent` invariant never fires across
the full filetest suite. vm gas, txtar, and the
closure/defer/recover/heap/goto escape-pattern filetests pass.

## Out of scope

- Sharing interface-held values when copying arrays (#5814) and per-call/per-op
  allocation reductions (#5816) — later parts of the same stack.
- Making the pool global (see negative results) — revisit only with a measured
  short-tx bottleneck.
