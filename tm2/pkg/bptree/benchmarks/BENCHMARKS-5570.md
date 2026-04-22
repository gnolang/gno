# IAVL vs B+32 Tree Benchmark Results — PR #5570

**Platform:** Apple M4 Pro, darwin/arm64
**Go:** benchtime=500ms, count=2
**Date:** 2026-04-22

This document augments [`BENCHMARKS.md`](./BENCHMARKS.md) with post-improvement
measurements from PR #5570. Each metric is shown per backend (memdb, pebbledb) with
the baseline B+32 number, the post B+32 number, and the % delta.

| Column | Commit | Meaning |
|---|---|---|
| `B+32-pre` | `2dee48277` (baseline, pre-optimization pass) | Before PR #5570 |
| `B+32-post` | `HEAD` of `feat/alex/bp32tree-second-pass` | After PR #5570 |

Raw data:
- Pre: [`baselines/memdb.txt`](./baselines/memdb.txt), [`baselines/pebbledb.txt`](./baselines/pebbledb.txt)
- Post: [`post/memdb.txt`](./post/memdb.txt), [`post/pebbledb.txt`](./post/pebbledb.txt)

Δ = `(post − pre) / pre`. Negative Δ = faster/smaller = improvement. `≈0` = within
run-to-run noise at count=2.

---

## Headline Changes

### Big wins (>10% improvement on at least one backend)

| Metric | memdb Δ | pebbledb Δ | Note |
|---|---|---|---|
| **Prune (10K base, 50 versions)** | **−77%** (33.18 → 7.69 ms) | **−59%** (110.2 → 45.2 ms) | 4.3x / 2.4x faster |
| **Block 500** | **−35%** (12.19 → 7.93 ms) | **−14%** (28.83 → 24.84 ms) | Larger-block advantage |
| **Block 100** | **−23%** (3.04 → 2.33 ms) | ≈0 (15.03 → 14.95 ms) | |
| **SetUpdate 1K** | **−23%** (3,290 → 2,522 ns) | **−13%** (4,689 → 4,067 ns) | Memory −35% |
| **SetUpdate 10K** | **−28%** (3,983 → 2,878 ns) | **−17%** (4,964 → 4,141 ns) | |
| **SetUpdate 100K** | **−14%** (4,960 → 4,278 ns) | **−14%** (5,991 → 5,178 ns) | |
| **Remove 1K** | **−17%** (6,071 → 5,042 ns) | **−11%** (7,372 → 6,597 ns) | Alloc 5 → 4 |
| **Remove 10K** | **−14%** (5,890 → 5,048 ns) | **−13%** (6,996 → 6,062 ns) | |
| **Remove 100K** | **−12%** (6,808 → 6,022 ns) | **−9%** (7,772 → 7,045 ns) | |
| **SetInsert 10K** | **−17%** (8,878 → 7,339 ns) | **−7%** (8,573 → 7,975 ns) | |
| **SetInsert 100K** | **−10%** (8,544 → 7,731 ns) | **−8%** (8,657 → 7,961 ns) | Memory −26% |
| **Mixed workload (BenchmarkBackends)** | **−26%** (9,545 → 7,085 ns) | **−9%** (40,429 → 36,644 ns) | Alloc −32% / −31% |
| **NonMembershipProof 1K** | **−25%** (2,971 → 2,226 ns) | **−31%** (4,085 → 2,835 ns) | |
| **NonMembershipProof 100K** | **−19%** (4,029 → 3,267 ns) | **−43%** (21,204 → 12,181 ns) | |
| **MembershipProof 1K** | **−15%** (1,423 → 1,215 ns) | **−12%** (1,793 → 1,572 ns) | |
| **MembershipProof 100K** | **−9%** (2,026 → 1,849 ns) | **−2%** (6,506 → 6,395 ns) | |
| **WorkingHash 1K** | **−10%** (82.9 → 75.0 µs) | −8% (90.7 → 83.5 µs) | Allocs 110 → 100 |
| **ScalingSet 1K** | **−23%** (3,288 → 2,535 ns) | **−19%** (4,741 → 3,846 ns) | |
| **ScalingSet 100K** | **−14%** (5,027 → 4,340 ns) | **−10%** (6,059 → 5,470 ns) | |
| **ScalingSet 1M** | **−13%** (6,421 → 5,607 ns) | **−13%** (6,679 → 5,822 ns) | |

### Regressions

| Metric | memdb Δ | pebbledb Δ | Note |
|---|---|---|---|
| **LoadVersion 1K** | **+28%** (4,640 → 5,947 ns) | **+28%** (4,604 → 5,895 ns) | Allocs 18 → 27 |
| **LoadVersion 100K** | **+49%** (12,755 → 18,961 ns) | **+33%** (13,364 → 17,826 ns) | Allocs 32–36 → 41 |
| **IterationRange 1K (mem)** | **+45%** (375 → 545 ns) | +3% (2,973 → 3,063 ns) | Allocations 23 → 23, bytes 368 → 1,232 on mem |
| **IterationRange 100K (mem)** | +12% (43.7 → 49.1 µs) | ≈0 (4.16 → 4.14 ms) | |

The LoadVersion regression is the most material — alloc count increased by 9–13 per
call. If LoadVersion frequency matters to your workload (it's called on node startup and
on rollback), this is worth investigating before merge.

### No material change (within ±5% or within noise)

GetHit, GetMiss, Has, SaveVersion (memdb and pebbledb), MultiVersionCreate,
IterationFull (ascending and descending), ScalingGet (at ≤100K), disk space
(single-version and multi-version), memory usage at small sizes.

IAVL numbers are stable between baseline and post runs (as expected — the PR only
touches B+32).

---

## Full Results

Format per metric: two tables, one per backend. `B+32 pre` is pre-PR, `B+32 post (Δ)` is
post-PR with % delta.

### GET (hit)

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 53 ns | 102 ns | 100 ns (−2%) |
| 10K | 72 ns | 159 ns | 158 ns (≈0) |
| 100K | 110 ns | 287 ns | 291 ns (+1%) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 53 ns | 401 ns | 398 ns (−1%) |
| 10K | 72 ns | 646 ns | 670 ns (+4%) |
| 100K | 119 ns | 4,663 ns | 4,657 ns (≈0) |

### GET (miss)

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 66 ns | 62 ns | 59 ns (−5%) |
| 10K | 68 ns | 110 ns | 104 ns (−5%) |
| 100K | 90 ns | 192 ns | 169 ns (−12%) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 341 ns | 62 ns | 59 ns (−5%) |
| 10K | 909 ns | 109 ns | 104 ns (−4%) |
| 100K | 6,418 ns | 178 ns | 169 ns (−5%) |

### HAS

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 551 ns | 81 ns | 77 ns (−5%) |
| 10K | 1,626 ns | 131 ns | 127 ns (−3%) |
| 100K | 3,014 ns | 216 ns | 205 ns (−5%) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 850 ns | 76 ns | 75 ns (−1%) |
| 10K | 5,509 ns | 128 ns | 125 ns (−2%) |
| 100K | 43,622 ns | 205 ns | 200 ns (−3%) |

### SET (insert)

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 2,255 ns | 7,929 ns | 7,324 ns (−8%) |
| 10K | 2,358 ns | 8,878 ns | 7,339 ns (−17%) |
| 100K | 2,862 ns | 8,544 ns | 7,731 ns (−10%) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 2,183 ns | 8,818 ns | 8,061 ns (−9%) |
| 10K | 2,395 ns | 8,573 ns | 7,975 ns (−7%) |
| 100K | 33k–60k ns | 8,657 ns | 7,961 ns (−8%) |

Alloc count unchanged (10–11). Bytes/op down from ~23k to ~17k on memdb.

### SET (update)

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 763 ns | 3,290 ns | 2,522 ns (**−23%**) |
| 10K | 1,148 ns | 3,983 ns | 2,878 ns (**−28%**) |
| 100K | 1,809 ns | 4,960 ns | 4,278 ns (−14%) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 739 ns | 4,689 ns | 4,067 ns (−13%) |
| 10K | 999 ns | 4,964 ns | 4,141 ns (−17%) |
| 100K | 1,797 ns | 5,991 ns | 5,178 ns (−14%) |

Bytes/op down from ~17k to ~11k on memdb; alloc count flat (7–8).

### Remove

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 1,162 ns | 6,071 ns | 5,042 ns (−17%) |
| 10K | 1,232 ns | 5,890 ns | 5,048 ns (−14%) |
| 100K | 1,485 ns | 6,808 ns | 6,022 ns (−12%) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 1,033 ns | 7,372 ns | 6,597 ns (−11%) |
| 10K | 1,355 ns | 6,996 ns | 6,062 ns (−13%) |
| 100K | 1,608 ns | 7,772 ns | 7,045 ns (−9%) |

Alloc count drops by 1 (6 → 5 on memdb 1K/10K; 5 → 4 at 100K).

---

## Iteration

### Full iteration (ascending)

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 186 µs | 33 µs | 35 µs (+6%) |
| 100K | 32.5 ms | 4.98 ms | 5.09 ms (+2%) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 139 µs | 306 µs | 304 µs (≈0) |
| 100K | 16.7 ms | 423 ms | 421 ms (≈0) |

### Full iteration (descending)

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 183 µs | 32 µs | 34 µs (+7%) |
| 100K | 27.7 ms | 4.92 ms | 5.08 ms (+3%) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 141 µs | 302 µs | 304 µs (≈0) |
| 100K | 18.5 ms | 427 ms | 417 ms (−2%) |

### Range iteration (~1% of keys)

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 35.8 µs | 375 ns | 545 ns (**+45%**) |
| 100K | 4.28 ms | 43.7 µs | 49.1 µs (+12%) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 2,175 ns | 2,973 ns | 3,063 ns (+3%) |
| 100K | 170.8 µs | 4.16 ms | 4.14 ms (≈0) |

The memdb range regression is notable: bytes/op went from 368 → 1,232 at 1K despite
alloc count being flat (23). This looks like a materialization/copy path change — worth
a quick look before merge.

---

## Block Workload (50% insert + 50% update, 100K base)

**memdb**

| Block Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 100 | 2.89 ms | 3.04 ms | 2.33 ms (**−23%**) |
| 500 | 10.60 ms | 12.19 ms | 7.93 ms (**−35%**) |

**pebbledb**

| Block Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 100 | 24.04 ms | 15.03 ms | 14.95 ms (≈0) |
| 500 | 69.99 ms | 28.83 ms | 24.84 ms (−14%) |

B+32 alloc count drops 32-46% for block workload (28K → 19K at 100, 98K → 56K at 500).

---

## Versioning

### SaveVersion (after 100 mutations)

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 759 µs | 787 µs | 736 µs (−6%) |
| 100K | 920 µs | 1,013 µs | 881 µs (−13%) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 4.80 ms | 5.97 ms | 5.71 ms (−4%) |
| 100K | 7.49 ms | 7.18 ms | 6.74 ms (−6%) |

### LoadVersion — **regression**

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 5,912 ns | 4,640 ns | 5,947 ns (**+28%**) |
| 100K | 28,862 ns | 12,755 ns | 18,961 ns (**+49%**) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 5,818 ns | 4,604 ns | 5,895 ns (**+28%**) |
| 100K | 29,941 ns | 13,364 ns | 17,826 ns (**+33%**) |

Allocations went from 18 → 27 at 1K (+50%) and 32–36 → 41 at 100K. B+32 still wins
vs IAVL (~1.5x faster at 100K post), but the absolute regression versus the pre-PR
implementation is real. **Recommend investigating before merging.**

### Multi-version creation

**memdb**

| Versions | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 10 | 6.07 ms | 6.64 ms | 6.20 ms (−7%) |
| 100 | 62.96 ms | 66.96 ms | 63.38 ms (−5%) |

**pebbledb**

| Versions | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 10 | 68.97 ms | 55.60 ms | 54.05 ms (−3%) |
| 100 | 693.7 ms | 589.2 ms | 586.3 ms (≈0) |

### Pruning — **major win**

**memdb**

| | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| ns/op | 41.59 ms | 33.18 ms | **7.69 ms (−77%)** |

**pebbledb**

| | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| ns/op | 641.86 ms | 110.24 ms | **45.15 ms (−59%)** |

Mark-and-sweep rewrite plus single-walk collapse drops B+32 pruning to **4.3x faster
than pre on memdb** and **2.4x faster on pebbledb**. Against IAVL at 100K pebbledb,
B+32 pruning is now 14x faster (was 5.8x in baseline).

---

## ICS23 Proofs

### Membership proof

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 4,246 ns | 1,423 ns | 1,215 ns (−15%) |
| 100K | 9,884 ns | 2,026 ns | 1,849 ns (−9%) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 6,256 ns | 1,793 ns | 1,572 ns (−12%) |
| 100K | 111,303 ns | 6,506 ns | 6,395 ns (−2%) |

Alloc count down (65 → 58 at 1K memdb; 83 → 74 at 100K memdb).

### Non-membership proof

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 8,035 ns | 2,971 ns | 2,226 ns (**−25%**) |
| 100K | 16,216 ns | 4,029 ns | 3,267 ns (−19%) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 10,181 ns | 4,085 ns | 2,835 ns (**−31%**) |
| 100K | 197,866 ns | 21,204 ns | 12,181 ns (**−43%**) |

Alloc count down from 129/165 to 112/145.

---

## WorkingHash

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 47.3 µs | 82.9 µs | 75.0 µs (−10%) |
| 10K | 48.9 µs | 82.8 µs | 75.8 µs (−8%) |
| 100K | 58.8 µs | 85.0 µs | 77.8 µs (−8%) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 47.5 µs | 90.7 µs | 83.5 µs (−8%) |
| 10K | 51.7 µs | 90.4 µs | 81.4 µs (−10%) |
| 100K | 597 µs | 87.0 µs | 81.1 µs (−7%) |

Alloc count drops from 110 → 100 on memdb 1K, 112 → 102 at 100K.

---

## Disk Space (pebbledb — always uses disk)

On-disk layout is unchanged.

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 0.2044 MB | 0.1386 MB | 0.1384 MB (≈0) |
| 10K | 2.501 MB | 1.374 MB | 1.374 MB (≈0) |
| 100K | 27.49 MB | 23.14 MB | 23.13 MB (≈0) |

### Multi-version disk space (10K base)

| Versions | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 10 | 2.89 MB | 2.37 MB | 2.36 MB (≈0) |
| 100 | 8.97 MB | 19.88 MB | 19.91 MB (≈0) |

---

## Memory Usage

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 0.88 MB | 0.55 MB | 0.55 MB (≈0) |
| 10K | 7.97 MB | 5.15 MB | 5.26 MB (+2%) |
| 100K | 72.63 MB | 49.53 MB | 50.96 MB (+3%) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 0.48 MB | 0.33 MB | 0.33 MB (≈0) |
| 10K | 2.55 MB | 2.31 MB | 3.02 MB (+31%)* |
| 100K | 25.31 MB | 36.80 MB | 38.93 MB (+6%) |

*Baseline 10K pebbledb memory had 30% run-to-run variance (2.76 and 1.86 MB across
the two samples). Post is more consistent (3.04/3.01). The apparent regression may
be measurement noise plus the mini-merkle shadow structure added in this PR.

---

## Scaling

### GET latency vs tree size

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 54 ns | 103 ns | 102 ns (≈0) |
| 10K | 73 ns | 164 ns | 159 ns (−3%) |
| 100K | 120 ns | 295 ns | 282 ns (−4%) |
| 1M | 755 ns | 862 ns | 845 ns (−2%) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 54 ns | 409 ns | 400 ns (−2%) |
| 10K | 73 ns | 657 ns | 664 ns (+1%) |
| 100K | 117 ns | 4,683 ns | 4,697 ns (≈0) |
| 1M | 10,384 ns | 5,856 ns | 5,777 ns (−1%) |

### SET latency vs tree size

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 821 ns | 3,288 ns | 2,535 ns (**−23%**) |
| 10K | 1,102 ns | 3,862 ns | 3,079 ns (**−20%**) |
| 100K | 1,770 ns | 5,027 ns | 4,340 ns (−14%) |
| 1M | 2,937 ns | 6,421 ns | 5,607 ns (−13%) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 753 ns | 4,741 ns | 3,846 ns (**−19%**) |
| 10K | 1,148 ns | 5,333 ns | 4,450 ns (−17%) |
| 100K | 1,851 ns | 6,059 ns | 5,470 ns (−10%) |
| 1M | 125,273 ns | 6,679 ns | 5,822 ns (−13%) |

### SaveVersion scaling

**memdb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 758 µs | 772 µs | 734 µs (−5%) |
| 10K | 774 µs | 781 µs | 745 µs (−5%) |
| 100K | 904 µs | 958 µs | 876 µs (−9%) |

**pebbledb**

| Size | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| 1K | 4.88 ms | 5.23 ms | 5.67 ms (+8%) |
| 10K | 4.97 ms | 5.54 ms | 5.47 ms (−1%) |
| 100K | 7.22 ms | 7.10 ms | 7.02 ms (−1%) |

---

## Backend Comparison (mixed workload, 100K keys)

70% read, 20% update, 10% insert. Commit every 500 ops.

| Backend | IAVL | B+32 pre | B+32 post (Δ) |
|---|---|---|---|
| memdb | 9,044 ns | 9,545 ns | 7,085 ns (**−26%**) |
| pebbledb | 57,668 ns | 40,429 ns | 36,644 ns (−9%) |

Alloc count drops substantially: 84 → 58 on memdb (−31%), 72 → 49 on pebbledb (−32%).
On memdb B+32 flips from ~tie-with-IAVL (baseline) to 22% faster than IAVL (post).
On pebbledb, B+32's lead over IAVL grows from 30% to 36%.

---

## Takeaways Specific to PR #5570

1. **Pruning rewrite is the headline result.** Mark-and-sweep plus single-walk
   collapse cuts pruning by 77% on memdb and 59% on pebbledb. This compounds with
   the existing pruning advantage: B+32 pruning at 100K pebbledb is now **14x
   faster than IAVL**.

2. **Hot-path allocation reductions show up across write-heavy operations.**
   SetUpdate (−14 to −28%), Remove (−9 to −17%), SetInsert (−7 to −17%), and the
   mixed workload (−9 to −26%) all improve together. Bytes/op falls 26–35% on memdb
   mutations.

3. **Block workload regained parity with IAVL on memdb and widened the lead on
   pebbledb.** The 500-block case improved 35% on memdb (now faster than IAVL) and
   14% on pebbledb (from 2.4x IAVL's speed to 2.9x).

4. **Proof generation is materially faster.** Non-membership proof at 100K pebbledb
   drops 43% (21.2 µs → 12.2 µs). Membership proof at 1K drops 12–15%. Proof allocs
   fall by ~7–11 per call.

5. **LoadVersion regressed.** 27–50% slower, +9–13 allocs per call. This is the one
   metric where the PR moves in the wrong direction. B+32 still wins vs IAVL by
   1.4–1.6x, but the absolute regression should be investigated. Likely root cause
   candidates: added integrity checks in node materialization, additional bookkeeping
   for the mark-and-sweep prune metadata.

6. **IterationRange memdb regressed** at small sizes (+45% at 1K). Bytes/op tripled
   (368 → 1,232) at unchanged alloc count — points to a wider value-copy in the range
   path. Minor absolute impact (170 ns) but worth a look.

7. **On-disk layout unchanged.** No migration needed; post-PR trees are binary-
   compatible with baseline trees on disk. Memory footprint is within noise except
   at 10K pebbledb (see footnote in Memory section).

8. **IAVL numbers are stable between runs.** No changes to IAVL paths — the deltas
   above reflect B+32 changes only, not machine/environmental drift.
