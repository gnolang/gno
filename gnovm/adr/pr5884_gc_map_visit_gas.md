# Meter the GC walk over map entries

## Context

GC gas is `gcVisitGas(visitCount)`. Each call to the GC visitor that reaches
an unstamped value counts one visit, priced from `gcVisitGasTable`, which
grows with heap size (29 ns/visit in cache up to 700 ns in DRAM).

`MapValue.VisitAssociated` walks the whole `MapList`: one `MapListItem`
pointer chase per entry. The visitor only counted entries whose key or value
was boxed (`V != nil`), and an `Object` only on its first reach in a cycle.
So two kinds of map cost an O(N) walk for about one visit
(gnolang/gno#5883):

- `map[int]int` and other maps of primitives, where `V == nil`;
- `map[int]X` where every value is the same `*MapValue`, `*FuncValue` or
  other `Object`, stamped after its first reach.

A 100k-entry map of either kind was charged about 200 gas per GC.

The first version of this PR counted only entries whose key and value were
both unboxed. That left the shared-object case open: 100k entries cost
232 gas.

## Decision

When the GC visitor reaches a `*MapValue`, it adds `mv.List.Size` to
`visitCount`. Every entry is counted once, whatever it holds:

- `List.Size` is O(1), so there is no second walk of the list.
- It does not depend on Go map iteration order, so it is deterministic.
- `overflow.Addp` guards the addition.
- A map still counts only on its first reach per cycle, because the cycle
  stamp is checked first. GC gas stays proportional to live heap size.

An entry's boxed key or value still counts its own visit. That is separate
work: the `MapListItem` is one heap node, and the object it points to is
another.

### Pricing

Entries are counted as full visits, not at a cheaper linear slope. A
benchmark of the bare `MapList` walk (`VisitAssociated` with a no-op visitor,
Xeon @ 2.1 GHz dev box, not the reference 8168) gave these results:

| Entries | Contiguous items | Scattered items |
|---------|------------------|-----------------|
| 1k      | 3.8 ns/entry     | 9.5 ns/entry    |
| 100k    | 4.3 ns/entry     | 56 ns/entry     |
| 1M      | 12.7 ns/entry    | 193 ns/entry    |
| 4M      | 13 ns/entry      | 230 ns/entry    |

"Scattered" is what a live chain gets, with insertions interleaved with
other allocations and realm loads. In that layout the walk is DRAM-bound
pointer chasing, the regime `gcVisitGasTable` is calibrated for. Pricing
entries like the contiguous `gcStackScanSlopeGas` would undercharge that
worst case by more than 10x. The cost of this choice is overcharging, which
is fail-safe: on a scattered heap an entry is charged about 1.6-3x its
measured cost, and on a contiguous heap much more. Recalibrating against
the reference 8168 machine is a possible follow-up.

## Alternatives considered

1. **Count only fully-unboxed entries** (the PR's first version). Rejected:
   a shared `Object` value bypasses it, and it walks the list twice
   (about 55% slower on a 1M-entry map, by one dev-box benchmark).
2. **Charge entries at a separate per-entry slope** (like
   `gcStackScanSlopeGas`). Rejected: it fits a contiguous scan, not a linked
   list. It would need a new constant calibrated on reference hardware.
3. **Count inside `MapValue.VisitAssociated`.** Rejected: `Visitor` is a
   generic `func(Value) bool` with no access to the counter. Keeping all
   gas accounting in `GCVisitorFn` keeps it in one place.

## Consequences

- **Consensus change:** GC gas rises for programs whose live heap holds
  large maps. For a 100k-entry `map[int]int` it goes from about 200 to
  about 20M gas per GC. Maps with boxed entries also pay one more visit per
  entry. Nodes must upgrade together.
- **Goldens:** no gas golden in the repo changes. The `sdk/vm` gas tests,
  integration txtars and gnovm filetests pass unchanged, because none of
  them runs GC over a large map.
- **Calibration benchmark:** `BenchmarkGCVisit` (`bench_gc_test.go`)
  builds a graph in which about 5% of objects are small maps. It now
  reports about 12% more visits for the same work, so its ns/visit figure
  drops. Keep this in mind when recalibrating `gcVisitGasTable`.
- **Follow-up, out of scope:** `ArrayValue`, `StructValue` and `Block`
  traversals still skip `V == nil` slots for free. `make([]int, 1e5)` costs
  about 7 visits. Those are contiguous scans (about 1-3 ns/slot), so the
  gap is much smaller than for maps. It should get a per-slot slope like
  `gcStackScanSlopeGas` in a separate PR, because it would touch every
  block and many goldens.
