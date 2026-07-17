# Correctly reuse/count string backing bytes in alloc and GC

## Context

The allocator charges a string at creation time (`AllocateString`: header +
per-byte cost). The GC then rebuilds `alloc.bytes` from scratch by walking live
objects and summing `GetShallowSize()`. For strings this recount was wrong in
both directions:

- **Double-count on shared backings.** `s1 := s` produces two `StringValue`s
  sharing one Go backing array. `GetShallowSize()` returned header + full byte
  length for each, so after a GC the recounted total could *exceed* what was
  charged at allocation time.
- **Overcharge on slices.** `s[m:n]` went through `alloc.NewString`, charging a
  full new string even though Go string slicing shares the source's backing.
- **Missing charge on reload.** `fillTypesOfValue` did nothing for
  `StringValue`, so strings loaded from store were never charged or known to
  the allocator at all.

Any fix must also keep a sliced substring correctly counted when its source
string becomes unreachable: the slice keeps the entire backing alive, and its
data pointer differs from the source's, so identity-by-pointer-equality
(e.g. a `map[uintptr]int64` keyed by `StringData`) undercounts in exactly that
case.

## Decision

Track string *backing extents* in the allocator and recount bytes once per
backing per GC cycle:

- `Allocator.stringRanges` holds sorted, disjoint `[start, end)` extents of
  every string backing charged through the allocator. `NewString` registers
  the extent (`TrackString`); registration is idempotent via **containment**
  lookup — a pointer anywhere inside a tracked range resolves to that backing.
  Containment (not equality) is what makes the slice-whose-source-died case
  inexpressible as a bug.
- `StringValue.GetShallowSize()` now returns the header only. The GC visitor
  (`GCVisitorFn`) special-cases `StringValue`: `CountStringBytes` returns the
  **full backing length** the first time any string resolving into a range is
  visited in a cycle (stamped via `lastCycle`), and 0 on subsequent visits —
  dedup for shared backings, full charge for live-via-slice backings.
- `TypedValue.GetSlice` for strings charges the header only and shares the
  backing instead of calling `NewString`.
- `fillTypesOfValue` routes loaded `StringValue`s through
  `store.GetAllocator().NewString`, so persisted strings are both charged and
  tracked on reload.
- After each GC, `CleanupTrackedStrings` prunes ranges not visited that cycle
  (dead backings). This also bounds the window in which Go's runtime could
  recycle a tracked address to a single GC cycle.
- `Fork()` starts the child with empty tracking: the child's tx store caches
  start empty, so every string it charges is re-registered through its own
  `NewString`/`fillTypesOfValue` path. Sharing the parent's slice would be
  unsafe (query paths fork onto other goroutines; the child's cleanup would
  prune the parent's entries).
- Empty strings are never tracked: `unsafe.StringData("")` returns an
  unspecified shared sentinel that would collapse all empty strings onto one
  entry.

## Alternatives considered

1. **`map[uintptr]int64` keyed by `StringData` pointer** (earlier iteration of
   this PR) — fails when a slice outlives its source: the slice's pointer is
   not equal to the source's key, so the backing bytes are dropped from the
   recount. Range containment fixes this structurally.
2. **Visit backing bytes via `VisitAssociated`** — the backing is raw data,
   not a `Value`; there is nothing for the visitor to visit. Kept
   `StringValue.VisitAssociated` as a documented no-op and put the byte
   accounting in `GCVisitorFn` instead.
3. **Keep full-size `GetShallowSize` and dedup by value equality** — string
   equality can't distinguish shared-backing duplicates from equal copies with
   distinct backings, and costs O(len) per compare.

## Consequences

- Allocation numbers change: string-heavy workloads no longer double-count
  after GC, and string slices get cheaper (header vs full copy). Loaded
  strings are now charged, raising some numbers. Golden files (`alloc_*.gno`)
  and gas txtars (`gnokey_gasfee`, `stdlib_restart_compare`) updated.
- New filetests `alloc_13.gno` / `alloc_13a.gno` pin recounting across two GC
  cycles and shared-backing dedup; unit tests in `alloc_test.go` cover
  tracking, dedup, cleanup, slice containment, and empty strings.
- `TrackString`/`CountStringBytes` are O(log n) via binary search on the
  sorted range slice; inserts are O(n) but amortized by per-cycle pruning.
- The allocator now holds `uintptr`s into Go heap memory. They are used only
  for identity/containment (never dereferenced), and per-cycle pruning bounds
  stale-address reuse.
