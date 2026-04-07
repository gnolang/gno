# ADR: GC Debug Tracking — Per-Type Allocation Tracking

## Context

The GnoVM allocator tracks total allocated bytes (`alloc.bytes`) but provides no
visibility into *which* value types contribute to the total. When the garbage
collector recounts live objects and the recount exceeds the pre-GC allocation
(a mismatch), there is no easy way to identify which type is under-tracked.

A recent bug (see PR #5436) showed that `PrepareNewValues` appended block items
without calling `AllocateBlockItems`, causing GC recount to exceed `alloc.bytes`.
Diagnosing this required ad-hoc instrumentation. A built-in debug facility would
make future mismatches immediately visible.

## Decision

Add per-type allocation tracking behind the `debug` build tag:

1. **`Allocate(size int64, typeName string)`** — every allocation call now carries
   a type name string. In non-debug builds, the string argument is unused and
   optimized away by the compiler (dead code after `if alloc.typeCounts != nil`).

2. **`typeCounts map[string]int64`** on `Allocator` — accumulates bytes per type
   name. Only initialized when `debug == true`.

3. **`Recount(size int64, typeName string)`** — GC recount also tracks per type.

4. **`SnapshotTypeCounts()`** — returns a copy of the current counts, used for
   before/after comparison across GC.

5. **Before/after comparison in `GarbageCollect()`** — snapshots `typeCounts`
   before GC reset, then compares with post-recount values. Any type where
   `after > before` is logged to stderr as a mismatch.

6. **`allocTypeName(v Value)`** helper in `garbage_collector.go` — extracts type
   name from a `Value` interface, handling both pointer and non-pointer receivers.

## Alternatives Considered

- **Per-type tracking always on**: rejected due to gas/performance overhead from
  the map operations on every allocation.
- **External profiling**: doesn't integrate with the allocator's logical byte
  counts and can't detect mismatch between allocation and GC recount.
- **Separate consistency check function**: considered (`CheckGCConsistency`), but
  the before/after approach in `GarbageCollect` is more natural and catches
  mismatches at the exact point they occur.

## Consequences

- **Debug builds** gain automatic mismatch detection: any allocation path that
  doesn't properly track a type will produce a stderr log line during GC.
- **Non-debug builds** have zero runtime overhead: `typeCounts` is nil, the nil
  check short-circuits, and the `typeName` string argument is not used.
- The `Allocate` and `Recount` signatures change (added `typeName` parameter),
  requiring updates to all call sites in `alloc.go`, `store.go`, and
  `garbage_collector.go`.
