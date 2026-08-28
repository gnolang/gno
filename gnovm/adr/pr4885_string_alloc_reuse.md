# Correctly reuse/count string backing bytes in alloc and GC

> **Status: experimental.** Backings are identified by Go heap address
> (`unsafe.StringData`): exact byte accounting with no change to the
> `StringValue` representation or persisted format, at the price of two
> determinism hazards, both found and fixed during review (see
> *Determinism*). The fallback, if address identity ever proves fragile
> anyway, is a representation-level approach (alternatives 2/3 below).

## Context

The allocator charges a string at creation (`AllocateString`: header +
bytes); the GC rebuilds `alloc.bytes` by walking live values and summing
`GetShallowSize()`. For strings this was wrong in every direction:

- **Loaded strings were free**: `fillTypesOfValue` ignored `StringValue`,
  so strings restored from store were neither charged nor known.
- **Shared backings double-counted**: `s1 := s` yields two values over one
  backing; `GetShallowSize` returned header+bytes for each, so a GC
  recount could exceed what allocation charged.
- **Slices overcharged**: `s[m:n]` paid a full copy despite sharing the
  backing — and any fix must keep a slice counted after its *source* dies
  (the backing stays alive), which pointer-equality keying cannot express.

## Decision

Track backing extents in the allocator; recount each backing once per GC
cycle:

- `Allocator.stringRanges` holds disjoint `[start, end)` extents of every
  backing charged through `NewString`. Lookup is by **containment**, so a
  slice resolves into its source's range even after the source dies.
- The set is a treap keyed by start (`string_ranges.go`): O(log n) insert
  and lookup, O(n) per-cycle prune. A sorted slice made a many-small-
  strings tx quadratic (~29µs per `NewString` at 100k live vs ~200ns).
  Priorities derive from the key; tree shape never affects accounting.
- Each entry **pins its backing**: `stringRange` holds the string itself,
  which doubles as the extent (derived on demand — pin and extent cannot
  disagree). See *Determinism* for why the pin is required.
- `StringValue.GetShallowSize()` returns the header only; `GCVisitorFn`
  charges backing bytes via `CountStringBytes` — full backing length,
  first resolving visit per cycle, `(0, false)` after (dedup).
- `GetSlice` charges header-only and shares the backing (no `NewString`).
- `fillTypesOfValue` mints loaded `StringValue`s through
  `store.GetAllocator().NewString` — charged and tracked on reload.
- `CleanupTrackedStrings` prunes unvisited ranges after each GC (a dead
  backing is retained ≤1 cycle); `clearStringTracking` empties the set
  between messages (`ClearObjectCache`), since cycle numbering restarts
  per machine. `Reset()` keeps the ranges — the GC's recount needs them.
- `Fork()` starts empty: the child's tx store caches start empty, so it
  re-registers everything it charges; sharing the parent's set would race
  on query paths and let the child prune the parent.
- Empty strings are never tracked (`unsafe.StringData("")` is a shared
  sentinel).

### Determinism

Two hazards of address identity, both fixed:

1. **Toolchain-controlled sharing.** Whether `s1 + ""`, `string([]byte)`
   elision, or literal interning returns a shared or fresh backing varies
   by Go version. Shared → dedup to one charge; fresh → two. Since
   `runtime.MemStats()` is contract-visible and the GC verdict decides
   limit aborts, mixed-toolchain nodes could fork. **Fix:** `trackString`
   clones iff the extent overlaps a tracked range, so every `NewString`
   gets its own range on every toolchain; the common fresh-backing case
   pays no copy. (Unconditional clone was rejected: it copies twice on
   paths like concat that already produced a fresh backing.)
2. **Address recycling.** With bare `uintptr` entries, Go could free a
   dead tracked backing and recycle the address before the next prune; a
   GC-visited string never minted (formatted panic text, a recovered
   error) could land in the stale range and be charged its extent —
   dependent on Go GC timing, i.e. different across nodes (measured: 1 to
   200 false hits per 20k attempts, varying per run in one process).
   **Fix: the pin.** A tracked address can never be freed or recycled, so
   an overlap in `trackString` is always genuine sharing and a
   containment hit always refers to the tracked backing. Cost: a dead
   backing held ≤1 GC cycle, bounded by bytes already charged. Regression
   test: `TestTrackString_PinnedBackings` (fails without the pin).

### Every Gno-visible string must be minted through `NewString`

An untracked string recounts as its 48-byte header regardless of length —
an undercount, the failure mode that matters for a memory cap. Entry
points audited: `alloc.NewString`, `Go2GnoValue` (stdlib native returns),
`fillTypesOfValue` (loads). Confirmed tracked: concat (`addAssign`),
`string(x)` conversions; `GetSlice`/`TypedValue.Copy` share tracked
backings via containment. Fixed in this PR: `convert.go` `MsgCall` string
args (attacker-controlled, previously free and untracked) and the
`uverse.go` realm-handle constructors (~150 gas per crossing now).
`*FuncValue` captures need no walk — each is a `*HeapItemValue` persisted
as its own object, re-tracked on load.

Deliberately untracked (thanks to the pin, an undercount only — a false
containment hit is impossible):

- Literals/const-folded strings: charged once at preprocess against the
  preprocess budget; same bytes for every execution.
- `typedString`/`typedRuntimeError` panic values: short VM-internal text,
  reachable only via `recover()`.

Follow-up (not string-related): byte-array call args in `convert.go` skip
`AllocateDataArray`; the GC recount catches those (arrays are Objects).

Dedup happens inside the visitor, keyed on the backing — independent of
which root reaches a string. No root-specific string handling should ever
be added.

## Alternatives considered

1. **Ref/owner flag on `StringValue`** ([#5082]): slices charge fixed
   overhead, owners charge bytes. A per-value flag cannot express backing
   identity: `s1 := s` still double-counts (two "owners"), and a ref
   whose owner dies drops the backing from the count entirely — the two
   corners this PR's containment lookup exists for.
2. **Backing-ID representation** (`StringValue{str, id, extent}`): a
   VM-assigned serial carried in the value. Deterministic by construction,
   no addresses, no pin; dedup by ID. The sound successor to (1) and the
   long-term alternative — rejected here for the representation and
   serialization migration it requires.
3. **`StringValue{*ArrayValue}`**: the backing becomes a real VM Object,
   so GC dedup (`LastGCCycle`) and slice sharing fall out of existing
   machinery. Heaviest at runtime: every string drags a heap object plus
   indirection (~230B header vs 48B), and every native op must
   materialize a Go string. Same migration cost as (2).
4. **`map[uintptr]int64` keyed by pointer equality** (earlier iteration):
   fails when a slice outlives its source — the slice's pointer is not a
   key. Containment fixes this structurally.
5. **`weak.Pointer` instead of the pin** (prototyped, benchmarked): store
   `weak.Make(backing)`; a dead weak pointer marks the entry stale, so
   recycled addresses miss deterministically with no memory retained.
   Killed by `weak.Make` throwing an *unrecoverable* runtime fatal on
   non-heap pointers — and rodata reaches `NewString` from user code
   (`lit + ""` can return the literal operand). The only guard is cloning
   every mint: measured 2.3–4.8× slower `NewString`, +14–22% GC pass.
6. **Death notifications** (`SetFinalizer`/`AddCleanup`): `AddCleanup`
   gives no ordering guarantee (memory can be reused before the cleanup
   runs — the race survives); `SetFinalizer` keeps the object alive an
   extra cycle (i.e. it *is* a pin), rejects interior pointers, and both
   fire on another goroutine, requiring locks in the consensus path.
7. **Mint every visited string + linter** (no pin): route panic texts
   through tracking and ban raw `StringValue(` conversions by lint; then
   every visited string provably resolves to its own or its live source's
   range. Sound today, but the guarantee degrades from construction to
   audit — one future `//nolint` or reflection-built value silently
   reopens the hole. Kept as a complement, not a substitute.
8. **Visit backing bytes via `VisitAssociated`**: nothing to visit — the
   backing is raw data, not a `Value`. Kept as a documented no-op.
9. **Dedup by string equality**: cannot distinguish a shared backing from
   equal copies, and costs O(len) per compare.

## Consequences

- Allocation numbers change: no post-GC double-count, slices cheaper
  (header vs full copy), loaded strings now charged. Golden `alloc_*`
  filetests and gas txtars updated; realm-handle strings and `MsgCall`
  string args now cost what they use.
- `trackString`/`CountStringBytes` are O(log n); the prune is O(n). Each
  tracked string costs one ~48B treap node (the pinning string doubles as
  the extent) of Go heap not charged to the allocator — bounded by
  strings charged, so not amplifiable. `string_ranges_test.go`
  model-checks the treap; `BenchmarkNewStringTracked`/
  `BenchmarkGCStringPass` are the regression reference (pin vs pre-pin:
  parity).
- The allocator holds `uintptr`s only for identity/containment (never
  dereferenced) plus the pinned strings that keep them valid.
- Strings whose backing Go chose to share pay one extra copy at
  `NewString`; fresh backings — the common case — pay nothing.

[#5082]: https://github.com/gnolang/gno/pull/5082
