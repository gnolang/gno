# IAVL vs B+32 Tree Benchmark Results

**Platform:** Apple M4 Pro, darwin/arm64
**Go:** benchtime=500ms, count=2
**Date:** 2026-04-22
**B+32 commit:** `2dee48277` (baseline, pre-optimization pass)

All benchmarks test both **memdb** (in-memory) and **pebbledb** (disk-backed) backends.
Usage: `go test -bench=. ./tm2/pkg/bptree/benchmarks/ -backend=memdb|pebbledb`

Raw data: [`baselines/memdb.txt`](./baselines/memdb.txt), [`baselines/pebbledb.txt`](./baselines/pebbledb.txt).

---

## Summary

| Category | memdb Winner | pebbledb Winner |
|---|---|---|
| GET (hit) | IAVL 1.5-2.6x | IAVL 5.5-39x |
| GET (miss) | ~tie to IAVL 2.1x | **B+32 5.5-36x** |
| HAS | **B+32 6.8-14x** | **B+32 11-213x** |
| SET (insert) | IAVL 2.7-3.5x | IAVL 3.3-4x |
| SET (update) | IAVL 2.5-4.3x | IAVL 3.3-6.3x |
| Remove | IAVL 4.6-5.2x | IAVL 4.7-7.1x |
| Iteration (full) | **B+32 5.6-6.5x** | IAVL 2.2-25x |
| Iteration (range) | **B+32 95-98x** | IAVL 1.4-24x |
| Block workload | ~tie (IAVL 5-15%) | **B+32 1.6-2.4x** |
| SaveVersion | ~tie | ~tie (IAVL 10-24% at 1K) |
| LoadVersion | **B+32 1.3-2.3x** | **B+32 1.3-2.2x** |
| Pruning | **B+32 25%** | **B+32 5.8x** |
| Membership proof | **B+32 3.0-4.9x** | **B+32 3.5-17x** |
| Non-membership proof | **B+32 2.7-4.0x** | **B+32 2.5-9.3x** |
| Memory | **B+32 37-38%** | mixed (B+32 worse at 100K) |
| Scaling GET (1M) | IAVL 1.1x | **B+32 1.8x** |

**Note on pebbledb vs goleveldb.** Prior revisions of this document used goleveldb as the
disk-backed backend. The backend has been switched to pebbledb to match gno.land's
production default. Directional results are largely the same, but absolute numbers differ:
pebbledb is typically slower than goleveldb on point reads (no in-memory value cache in
the default configuration) and writes larger files, which compresses the B+32 disk-space
advantage at 100K from ~30% to ~16%.

---

## Single Operations

### GET (hit) — known keys

| Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 1K | 53 ns | 102 ns (1.9x) | 53 ns | 401 ns (7.6x) |
| 10K | 72 ns | 159 ns (2.2x) | 72 ns | 646 ns (9.0x) |
| 100K | 110 ns | 287 ns (2.6x) | 119 ns | 4,663 ns (39x) |

IAVL uses a flat fast-node index giving O(1) GET with 0 allocs. B+32 traverses the tree
(O(log n)) — on memdb this costs 16 B / 1 alloc; on pebbledb 64-128 B / 2-3 allocs due
to per-node disk reads.

### GET (miss) — random keys not in tree

| Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 1K | 66 ns | 62 ns (**~tie**) | 341 ns | 62 ns (**B+32 5.5x**) |
| 10K | 68 ns | 110 ns (1.6x) | 909 ns | 109 ns (**B+32 8.3x**) |
| 100K | 90 ns | 192 ns (2.1x) | 6,418 ns | 178 ns (**B+32 36x**) |

On memdb, IAVL's fast-node index handles misses cheaply (24 B / 1 alloc). On pebbledb,
IAVL's miss path hits disk (88-152 B / 2-3 allocs), while B+32 rejects misses in-memory
with 0 allocs.

### HAS — existence check (known keys)

| Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 1K | 551 ns | 81 ns (**6.8x**) | 850 ns | 76 ns (**11x**) |
| 10K | 1,626 ns | 131 ns (**12x**) | 5,509 ns | 128 ns (**43x**) |
| 100K | 3,014 ns | 216 ns (**14x**) | 43,622 ns | 205 ns (**213x**) |

IAVL `Has()` always traverses the full tree (does not use the fast-node index).
B+32 `Has()` does a zero-allocation in-memory node traversal without resolving values.
On pebbledb, IAVL's tree traversal hits disk, making B+32's advantage enormous (213x at 100K).

### SET (insert) — new keys

| Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 1K | 2,255 ns | 7,929 ns (3.5x) | 2,183 ns | 8,818 ns (4.0x) |
| 10K | 2,358 ns | 8,878 ns (3.8x) | 2,395 ns | 8,573 ns (3.6x) |
| 100K | 2,862 ns | 8,544 ns (3.0x) | 33k–60k ns* | 8,657 ns (~tie/flip) |

*IAVL pebbledb/100K SetInsert shows high run-to-run variance in this sample (32,384 and
60,188 ns across the two runs), due to compaction/flush timing. B+32 is more stable at
this size because it amortizes writes across blocks.

### SET (update) — existing keys

| Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 1K | 763 ns | 3,290 ns (4.3x) | 739 ns | 4,689 ns (6.3x) |
| 10K | 1,148 ns | 3,983 ns (3.5x) | 999 ns | 4,964 ns (5.0x) |
| 100K | 1,809 ns | 4,960 ns (2.7x) | 1,797 ns | 5,991 ns (3.3x) |

On pebbledb, B+32 SET update regresses more than IAVL because B+32 must read the
existing node from disk before modifying, while IAVL's fast-node index keeps values in memory.

### Remove

| Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 1K | 1,162 ns | 6,071 ns (5.2x) | 1,033 ns | 7,372 ns (7.1x) |
| 10K | 1,232 ns | 5,890 ns (4.8x) | 1,355 ns | 6,996 ns (5.2x) |
| 100K | 1,485 ns | 6,808 ns (4.6x) | 1,608 ns | 7,772 ns (4.8x) |

Remove gap is consistent across backends. Benchmark uses batch remove/re-insert
(batch=100) to amortize timer overhead.

---

## Iteration

### Full iteration (ascending)

| Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 1K | 186 µs | 33 µs (**5.6x**) | 139 µs | 306 µs (2.2x) |
| 100K | 32.5 ms | 4.98 ms (**6.5x**) | 16.7 ms | 423 ms (25x) |

### Full iteration (descending)

| Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 1K | 183 µs | 32 µs (**5.7x**) | 141 µs | 302 µs (2.1x) |
| 100K | 27.7 ms | 4.92 ms (**5.6x**) | 18.5 ms | 427 ms (23x) |

### Range iteration (~1% of keys)

| Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 1K | 35.8 µs | 375 ns (**95x**) | 2,175 ns | 2,973 ns (1.4x) |
| 100K | 4.28 ms | 43.7 µs (**98x**) | 170.8 µs | 4.16 ms (24x) |

**Backend fundamentally changes the iteration story.** On memdb, B+32's contiguous leaf
layout and stack-based traversal give a 5.6–98x advantage. On pebbledb, B+32's
out-of-line value storage requires a separate disk read per `Value()` call (300K allocs
at 100K full iteration), making it 23–25x slower; IAVL's inline values win there.

---

## Block Workload

50% insert + 50% update, commit after each block. 100K base tree.

| Block Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 100 | 2.89 ms | 3.04 ms (IAVL 5%) | 24.04 ms | 15.03 ms (**B+32 1.6x**) |
| 500 | 10.60 ms | 12.19 ms (IAVL 15%) | 69.99 ms | 28.83 ms (**B+32 2.4x**) |

On memdb the block workload is roughly a tie (IAVL slightly ahead due to faster point
mutations). On pebbledb B+32's amortized-batched-write advantage dominates, growing
with block size (1.6x at 100 ops, 2.4x at 500 ops).

---

## Versioning

### SaveVersion (after 100 mutations)

| Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 1K | 759 µs | 787 µs (~tie) | 4.80 ms | 5.97 ms (IAVL 24%) |
| 100K | 920 µs | 1,013 µs (~tie) | 7.49 ms | 7.18 ms (~tie) |

On memdb, SaveVersion is a virtual tie. On pebbledb, IAVL is 24% faster at 1K but the
trees converge at 100K.

### LoadVersion

| Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 1K | 5,912 ns | 4,640 ns (**B+32 1.3x**) | 5,818 ns | 4,604 ns (**B+32 1.3x**) |
| 100K | 28,862 ns | 12,755 ns (**B+32 2.3x**) | 29,941 ns | 13,364 ns (**B+32 2.2x**) |

### Multi-version creation (50 mutations/version)

| Versions | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 10 | 6.07 ms | 6.64 ms (~tie) | 68.97 ms | 55.60 ms (**B+32 1.2x**) |
| 100 | 62.96 ms | 66.96 ms (~tie) | 693.7 ms | 589.2 ms (**B+32 1.2x**) |

### Pruning (delete versions 1-50 from 100 versions, 10K base)

| | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| ns/op | 41.59 ms | 33.18 ms (**B+32 25%**) | 641.86 ms | 110.24 ms (**B+32 5.8x**) |

B+32's pruning advantage **amplifies on pebbledb** (5.8x vs 1.25x on memdb).

---

## ICS23 Proofs

### Membership proof (existence)

| Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 1K | 4,246 ns | 1,423 ns (**3.0x**) | 6,256 ns | 1,793 ns (**3.5x**) |
| 100K | 9,884 ns | 2,026 ns (**4.9x**) | 111,303 ns | 6,506 ns (**17x**) |

### Non-membership proof (non-existence)

| Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 1K | 8,035 ns | 2,971 ns (**2.7x**) | 10,181 ns | 4,085 ns (**2.5x**) |
| 100K | 16,216 ns | 4,029 ns (**4.0x**) | 197,866 ns | 21,204 ns (**9.3x**) |

B+32's proof advantage grows on pebbledb — IAVL membership-proof generation at 100K
jumps from ~10 µs (memdb) to ~111 µs (pebbledb) while B+32 only goes from 2 µs to
6.5 µs.

---

## WorkingHash (root hash computation)

| Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 1K | 47.3 µs | 82.9 µs (1.8x) | 47.5 µs | 90.7 µs (1.9x) |
| 10K | 48.9 µs | 82.8 µs (1.7x) | 51.7 µs | 90.4 µs (1.8x) |
| 100K | 58.8 µs | 85.0 µs (1.4x) | 597 µs | 87.0 µs (**B+32 6.9x**) |

At 100K on pebbledb, IAVL's hash computation walks the disk-backed tree (1260–1290
allocs) while B+32 keeps a small in-memory miniMerkle structure (~100 allocs), so
the trees flip roles and B+32 becomes 6.9x faster.

---

## Disk Space (pebbledb — always uses disk)

| Size | IAVL (MB) | B+32 (MB) | B+32 savings |
|---|---|---|---|
| 1K | 0.2044 | 0.1386 | **32%** |
| 10K | 2.501 | 1.374 | **45%** |
| 100K | 27.49 | 23.14 | **16%** |

| | IAVL (bytes/key) | B+32 (bytes/key) |
|---|---|---|
| 1K | 214 | 145 |
| 10K | 262 | 144 |
| 100K | 288 | 243 |

### Multi-version disk space (10K base)

| Versions | IAVL (MB) | B+32 (MB) |
|---|---|---|
| 10 | 2.89 | 2.37 |
| 100 | 8.97 | 19.88 |

B+32 uses less disk for few versions, but **grows ~2.2x larger at 100 versions** due
to immutable node copies per version.

---

## Memory Usage

| Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 1K | 0.88 MB (**919 B/key**) | 0.55 MB (**578 B/key**) | 0.48 MB (503 B/key) | 0.33 MB (341 B/key) |
| 10K | 7.97 MB (836 B/key) | 5.15 MB (540 B/key) | 2.55 MB (268 B/key) | 2.31 MB (242 B/key) |
| 100K | 72.63 MB (762 B/key) | 49.53 MB (519 B/key) | 25.31 MB (265 B/key) | 36.80 MB (386 B/key) |

On memdb, B+32 uses **32-36% less memory** per key (values are held in-process for
both trees, but B+32's node layout is denser).

On pebbledb, IAVL's fast-node index is paged out to disk on cold starts, so IAVL's
live heap is smaller. B+32's larger heap at 100K (38 MB vs 25 MB) reflects its
mini-merkle shadow structure plus cached leaves.

---

## Scaling (GET latency vs tree size)

| Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 1K | 54 ns | 103 ns (1.9x) | 54 ns | 409 ns (7.6x) |
| 10K | 73 ns | 164 ns (2.2x) | 73 ns | 657 ns (9.0x) |
| 100K | 120 ns | 295 ns (2.5x) | 117 ns | 4,683 ns (40x) |
| 1M | 755 ns | 862 ns (1.1x) | 10,384 ns | 5,856 ns (**B+32 1.8x**) |

On memdb, IAVL's fast-node advantage shrinks at 1M keys (1.1x gap). On pebbledb, IAVL's
fast-node cache is no longer effective at 1M and every GET hits disk (10 µs), while
B+32's shallow tree (height=4) keeps latency at 6 µs and flips the comparison.

### Scaling (SET latency vs tree size)

| Size | IAVL/mem | B+32/mem | IAVL/peb | B+32/peb |
|---|---|---|---|---|
| 1K | 821 ns | 3,288 ns (4.0x) | 753 ns | 4,741 ns (6.3x) |
| 10K | 1,102 ns | 3,862 ns (3.5x) | 1,148 ns | 5,333 ns (4.6x) |
| 100K | 1,770 ns | 5,027 ns (2.8x) | 1,851 ns | 6,059 ns (3.3x) |
| 1M | 2,937 ns | 6,421 ns (2.2x) | 125,273 ns | 6,679 ns (**B+32 19x**) |

At 1M keys on pebbledb, IAVL's per-SET latency explodes (125 µs) due to compaction
pressure and fast-node-index churn, while B+32's amortized block writes stay flat at
~6.7 µs.

---

## Backend Comparison (mixed workload, 100K keys)

70% read, 20% update, 10% insert. Commit every 500 ops.

| Backend | IAVL (ns/op) | B+32 (ns/op) | Ratio |
|---|---|---|---|
| memdb | 9,044 | 9,545 | ~tie |
| pebbledb | 57,668 | 40,429 | **B+32 30% faster** |

---

## Key Takeaways

1. **Backend choice fundamentally changes the comparison.** B+32's iteration advantage
   on memdb (6-98x faster) becomes a disadvantage on pebbledb (23-25x slower for full
   iteration) due to per-value disk reads. Choose the backend that matches your deployment.

2. **B+32 HAS is 7-213x faster** across all backends because IAVL's `Has()` traverses
   the full tree while B+32 does a zero-allocation in-memory traversal. This gap
   widens dramatically on pebbledb (213x at 100K).

3. **B+32 GetMiss flips on pebbledb**: on memdb it's a tie-to-IAVL-2.1x, but on pebbledb
   B+32 is 5.5-36x faster because IAVL's miss path hits disk while B+32 rejects misses
   in-memory.

4. **IAVL wins on single-key mutations (SET/Remove)** on both backends due to its
   fast-node index. The gap is ~3-7x on pebbledb except at 1M keys where IAVL's
   fast-node cache overflows and SET latency regresses sharply (19x slower than B+32).

5. **B+32 block workload wins on pebbledb** (1.6-2.4x faster) due to fewer, larger
   batched writes.

6. **B+32 pruning advantage amplifies on pebbledb** (5.8x vs 1.25x on memdb).

7. **ICS23 proofs are 2.5-17x faster** in B+32 across backends. On pebbledb at 100K,
   membership proofs are 17x faster (6.5 µs vs 111 µs).

8. **B+32 uses 32-38% less memory on memdb** and 16-45% less disk per key at ≤10K.
   At 100K+ keys, memory advantage depends on backend (B+32 worse at 100K pebbledb).

9. **LoadVersion is 1.3-2.3x faster** in B+32 on both backends — important for node
   startup.

10. **Multi-version disk growth remains a B+32 weakness**: 2.2x more disk than IAVL at
    100 retained versions.

11. **B+32 scales better to 1M keys on pebbledb**: GET flips to B+32 1.8x, SET flips
    to B+32 19x because IAVL hits disk-bound thrash at that size.
