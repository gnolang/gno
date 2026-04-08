# IAVL vs B+32 Tree Benchmark Results

**Platform:** Apple M4 Pro, darwin/arm64
**Go:** benchtime=1s, count=1
**Date:** 2026-04-08

All benchmarks test both **memdb** (in-memory) and **goleveldb** (disk-backed) backends.
Usage: `go test -bench=. ./tm2/pkg/bptree/benchmarks/ -backend=memdb|goleveldb`

---

## Summary

| Category | memdb Winner | goleveldb Winner |
|---|---|---|
| GET (hit) | IAVL 2-4x | IAVL 5-10x |
| GET (miss) | ~tie to IAVL 2.5x | **B+32 1.3-8.4x** |
| HAS | **B+32 8-15x** | **B+32 11-52x** |
| SET (insert) | IAVL 2.8-3x | IAVL 1.5-7.5x |
| SET (update) | IAVL 2.2-4.6x | IAVL 2.9-10.5x |
| Remove | IAVL 4.1-5x | IAVL 4.5-5.8x |
| Iteration (full) | **B+32 2.1-3.7x** | IAVL 1.4-9x |
| Iteration (range) | **B+32 67-92x** | IAVL 1.3-5.1x |
| Block workload | **B+32 3-14%** | **B+32 20-32%** |
| SaveVersion | ~tie | IAVL 8-23% |
| LoadVersion | **B+32 2-3.3x** | **B+32 2-3.3x** |
| Pruning | **B+32 25%** | **B+32 3.3x** |
| Membership proof | **B+32 3.7-5.1x** | **B+32 3.9-8x** |
| Non-membership proof | **B+32 2.9-3.6x** | **B+32 2.5-4.1x** |
| Memory | **B+32 40-47%** | **B+32 25-53%** |
| Scaling GET (1M) | IAVL 1.1x | IAVL 1.3x |

The backend choice fundamentally changes the comparison. B+32's iteration advantage
on memdb becomes a disadvantage on goleveldb due to per-value disk reads. Conversely,
B+32's block workload and pruning advantages grow significantly on goleveldb.

---

## Single Operations

### GET (hit) — known keys

| Size | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 1K | 55 ns | 125 ns (2.3x) | 54 ns | 405 ns (7.5x) |
| 10K | 78 ns | 189 ns (2.4x) | 79 ns | 625 ns (7.9x) |
| 100K | 239 ns | 534 ns (2.2x) | 274 ns | 2,734 ns (10.0x) |

IAVL uses a flat fast-node index giving O(1) GET with 0 allocs. B+32 traverses the tree
(O(log n)) — on memdb this costs 80 B/2 allocs; on goleveldb 264-1048 B/7-19 allocs
due to per-node disk reads.

### GET (miss) — random keys not in tree

| Size | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 1K | 72 ns | 72 ns (1.0x) | 317 ns | 72 ns (**B+32 4.4x**) |
| 10K | 74 ns | 119 ns (1.6x) | 503 ns | 141 ns (**B+32 3.6x**) |
| 100K | 102 ns | 254 ns (2.5x) | 2,050 ns | 243 ns (**B+32 8.4x**) |

On memdb, IAVL's fast-node index gives O(1) miss. On goleveldb, IAVL's miss path hits
disk (144-882 B/5-16 allocs), while B+32 rejects misses in-memory with 0 allocs.

### HAS — existence check (known keys)

| Size | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 1K | 667 ns | 83 ns (**8.0x**) | 882 ns | 81 ns (**10.9x**) |
| 10K | 2,002 ns | 129 ns (**15.5x**) | 4,542 ns | 135 ns (**33.6x**) |
| 100K | 3,864 ns | 253 ns (**15.3x**) | 15,271 ns | 291 ns (**52.5x**) |

IAVL `Has()` always traverses the full tree (does not use the fast-node index).
B+32 `Has()` does a zero-allocation in-memory node traversal without resolving values.
On goleveldb, IAVL's tree traversal hits disk, making B+32's advantage enormous (52x at 100K).

### SET (insert) — new keys

| Size | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 1K | 3,106 ns | 9,384 ns (3.0x) | 3,428 ns | 15,095 ns (4.4x) |
| 10K | 3,114 ns | 9,228 ns (3.0x) | 3,234 ns | 24,392 ns (7.5x) |
| 100K | 3,562 ns | 10,077 ns (2.8x) | 5,444 ns | 16,806 ns (3.1x) |

### SET (update) — existing keys

| Size | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 1K | 903 ns | 4,151 ns (4.6x) | 815 ns | 8,855 ns (10.9x) |
| 10K | 1,433 ns | 4,494 ns (3.1x) | 1,320 ns | 13,877 ns (10.5x) |
| 100K | 2,388 ns | 5,150 ns (2.2x) | 2,339 ns | 11,478 ns (4.9x) |

On goleveldb, B+32 SET update regresses more than IAVL because B+32 must read the
existing node from disk before modifying, while IAVL's fast-node index keeps values in memory.

### Remove

| Size | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 1K | 1,511 ns | 7,481 ns (5.0x) | 1,144 ns | 6,646 ns (5.8x) |
| 10K | 1,781 ns | 7,257 ns (4.1x) | 1,451 ns | 6,969 ns (4.8x) |
| 100K | 1,823 ns | 7,473 ns (4.1x) | 1,710 ns | 7,698 ns (4.5x) |

Remove gap is consistent across backends. Benchmark uses batch remove/re-insert
(batch=100) to amortize timer overhead.

---

## Iteration

### Full iteration (ascending)

| Size | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 1K | 217 us | 59 us (**3.7x**) | 147 us | 299 us (2.0x) |
| 100K | 38.1 ms | 17.9 ms (**2.1x**) | 21.2 ms | 190.2 ms (9.0x) |

### Full iteration (descending)

| Size | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 1K | 223 us | 62 us (**3.6x**) | 277 us | 313 us (1.1x) |
| 100K | 38.4 ms | 19.7 ms (**1.9x**) | 21.5 ms | 184.7 ms (8.6x) |

### Range iteration (~1% of keys)

| Size | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 1K | 40 us | 595 ns (**67x**) | 2,128 ns | 2,753 ns (1.3x) |
| 100K | 5.7 ms | 62 us (**92x**) | 179 us | 908 us (5.1x) |

**Backend fundamentally changes the iteration story.** On memdb, B+32's contiguous leaf
layout and stack-based traversal gives 2-92x faster iteration. On goleveldb, B+32's
out-of-line value storage requires a separate disk read per `Value()` call (1860K allocs
for 100K keys vs IAVL's 316K), making it 5-9x slower.

---

## Block Workload

50% insert + 50% update, commit after each block. 100K base tree.

| Block Size | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 100 | 3.60 ms | 3.49 ms (**3%**) | 11.0 ms | 7.57 ms (**32%**) |
| 500 | 14.3 ms | 12.6 ms (**14%**) | 42.7 ms | 34.2 ms (**20%**) |

B+32's block advantage **grows on goleveldb** (20-32% vs 3-14% on memdb) due to
fewer, larger batched writes.

---

## Versioning

### SaveVersion (after 100 mutations)

| Size | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 1K | 967 us | 1,014 us (~tie) | 1,275 us | 1,655 us (IAVL 23%) |
| 100K | 1,105 us | 1,107 us (~tie) | 1,856 us | 2,020 us (IAVL 8%) |

On memdb, SaveVersion is a virtual tie. On goleveldb, IAVL is 8-23% faster.

### LoadVersion (goleveldb — always uses disk)

| Size | IAVL (ns/op) | B+32 (ns/op) | Ratio |
|---|---|---|---|
| 1K | 6,353 | 3,185 | **B+32 2.0x faster** |
| 100K | 16,507 | 5,005 | **B+32 3.3x faster** |

### Multi-version creation (50 mutations/version)

| Versions | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 10 | 6.36 ms | 6.43 ms (~tie) | 12.9 ms | 12.0 ms (~tie) |
| 100 | 65.5 ms | 66.8 ms (~tie) | 135 ms | 140 ms (~tie) |

### Pruning (delete versions 1-50 from 100 versions, 10K base)

| | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| ns/op | 44.6 ms | 33.3 ms (**25%**) | 187 ms | 56.4 ms (**3.3x**) |

B+32's pruning advantage **amplifies on goleveldb** (3.3x vs 1.25x on memdb).

---

## ICS23 Proofs

### Membership proof (existence)

| Size | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 1K | 4,846 ns | 1,296 ns (**3.7x**) | 5,878 ns | 1,521 ns (**3.9x**) |
| 100K | 10,360 ns | 2,016 ns (**5.1x**) | 35,600 ns | 4,438 ns (**8.0x**) |

### Non-membership proof (non-existence)

| Size | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 1K | 8,987 ns | 3,149 ns (**2.9x**) | 9,803 ns | 3,975 ns (**2.5x**) |
| 100K | 16,505 ns | 4,552 ns (**3.6x**) | 44,164 ns | 10,647 ns (**4.1x**) |

B+32's proof advantage grows on goleveldb — IAVL proof generation at 100K goes from
10us (memdb) to 36-44us (goleveldb) while B+32 only goes from 2-5us to 4-11us.

---

## Disk Space (goleveldb — always uses disk)

| Size | IAVL (MB) | B+32 (MB) | B+32 savings |
|---|---|---|---|
| 1K | 0.204 | 0.142 | **30%** |
| 10K | 2.044 | 1.410 | **31%** |
| 100K | 18.58 | 12.92 | **30%** |

| | IAVL (bytes/key) | B+32 (bytes/key) |
|---|---|---|
| 1K | 214 | 148 |
| 10K | 214 | 148 |
| 100K | 195 | 136 |

### Multi-version disk space (10K base)

| Versions | IAVL (MB) | B+32 (MB) |
|---|---|---|
| 10 | 2.43 | 2.28 |
| 100 | 5.71 | 10.08 |

B+32 uses less disk for few versions, but **grows ~1.8x larger at 100 versions** due
to immutable node copies per version.

---

## Memory Usage

| Size | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 1K | 0.88 MB | 0.48 MB (**45%**) | 0.57 MB | 0.27 MB (**53%**) |
| 10K | 8.04 MB | 4.54 MB (**44%**) | 4.09 MB | 2.63 MB (**36%**) |
| 100K | 72.6 MB | 43.4 MB (**40%**) | 35.9 MB | 27.0 MB (**25%**) |

| | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 1K | 919 B/key | 505 B/key | 601 B/key | 279 B/key |
| 10K | 843 B/key | 476 B/key | 429 B/key | 275 B/key |
| 100K | 761 B/key | 455 B/key | 377 B/key | 283 B/key |

B+32 uses less memory on both backends. On goleveldb, both trees use significantly
less heap (disk-backed nodes are not cached in memory).

---

## Scaling (GET latency vs tree size)

| Size | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 1K | 55 ns | 128 ns (2.3x) | 57 ns | 401 ns (7.0x) |
| 10K | 76 ns | 182 ns (2.4x) | 83 ns | 605 ns (7.3x) |
| 100K | 186 ns | 429 ns (2.3x) | 260 ns | 2,741 ns (10.5x) |
| 1M | 799 ns | 908 ns (1.1x) | 3,622 ns | 4,602 ns (1.3x) |

On memdb, IAVL's fast-node advantage diminishes at 1M (1.1x). On goleveldb, both
hit disk at 1M (IAVL's fast-node cache overflows), narrowing the gap to 1.3x.

## Scaling (SET latency vs tree size)

| Size | IAVL/mem | B+32/mem | IAVL/lvl | B+32/lvl |
|---|---|---|---|---|
| 1K | 821 ns | 3,705 ns (4.5x) | 804 ns | 10,921 ns (13.6x) |
| 10K | 1,148 ns | 3,655 ns (3.2x) | 1,510 ns | 11,170 ns (7.4x) |
| 100K | 1,885 ns | 4,655 ns (2.5x) | 2,203 ns | 11,742 ns (5.3x) |
| 1M | 3,029 ns | 5,891 ns (1.9x) | 8,505 ns | 13,515 ns (1.6x) |

SET gap narrows at 1M on both backends. On goleveldb, both trees hit disk for writes.

---

## Backend Comparison (mixed workload, 100K keys)

70% read, 20% update, 10% insert. Commit every 500 ops.

| Backend | IAVL (ns/op) | B+32 (ns/op) | Ratio |
|---|---|---|---|
| memdb | 10,307 | 9,939 | ~tie |
| goleveldb | 30,355 | 22,472 | **B+32 26% faster** |

---

## Key Takeaways

1. **Backend choice fundamentally changes the comparison.** B+32's iteration advantage
   on memdb (2-92x faster) becomes a disadvantage on goleveldb (5-9x slower) due to
   per-value disk reads. Choose the backend that matches your deployment.

2. **B+32 HAS is 8-52x faster** across all backends because IAVL's `Has()` traverses
   the full tree while B+32 does a zero-allocation in-memory traversal. This gap
   widens dramatically on goleveldb (52x at 100K).

3. **B+32 GetMiss flips on goleveldb**: on memdb IAVL is 2.5x faster (fast-node index),
   but on goleveldb B+32 is 8.4x faster because IAVL's miss path hits disk while B+32
   rejects misses in-memory.

4. **IAVL wins on single-key mutations (SET/Remove)** on both backends due to its
   fast-node index. The gap widens on goleveldb for SET but narrows for both at 1M keys.

5. **B+32 block workload advantage grows on goleveldb** (20-32% vs 3-14% on memdb)
   due to fewer, larger batched writes.

6. **B+32 pruning advantage amplifies on goleveldb** (3.3x vs 1.25x on memdb).

7. **ICS23 proofs are 3-8x faster** in B+32 across backends. On goleveldb at 100K,
   membership proofs are 8x faster (4us vs 36us).

8. **B+32 uses 25-53% less memory** and **30% less disk** per key.

9. **LoadVersion is 2-3.3x faster** in B+32 — important for node startup.

10. **Multi-version disk growth remains a B+32 weakness**: 1.8x more disk than IAVL at
    100 retained versions.
