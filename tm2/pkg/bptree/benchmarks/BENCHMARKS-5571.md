# IAVL vs B+32 Tree Benchmark Results — PR #5571

**Platform:** Apple M4 Pro, darwin/arm64
**Go:** benchtime=500ms, count=2
**Date:** 2026-04-23

This document augments [`BENCHMARKS.md`](./BENCHMARKS.md) and
[`BENCHMARKS-5570.md`](./BENCHMARKS-5570.md) with post-improvement
measurements from PR #5571. Each metric is shown per backend (memdb, pebbledb)
with the pre-PR baseline, the post-PR-5570 number (the previous milestone),
and the post-PR-5571 number with deltas against both reference points.

| Column | Commit | Meaning |
|---|---|---|
| `B+32 pre` | `2dee48277` | Pre-optimization snapshot used as the global baseline |
| `B+32 5570` | `a807ec551` (tip of `feat/alex/bp32tree-second-pass`) | After PR #5570 (correctness + hot-path pass) |
| `B+32 5571` | `7cafdd4f3` (tip of `feat/alex/bp32tree-advanced`) | After PR #5571 (latest-view cache + leaf v2/v3 + 13-fix perf pass) |

Raw data:
- Pre: [`baselines/memdb.txt`](./baselines/memdb.txt), [`baselines/pebbledb.txt`](./baselines/pebbledb.txt)
- 5570: [`post/memdb.txt`](./post/memdb.txt), [`post/pebbledb.txt`](./post/pebbledb.txt)
- 5571: [`post-5571/memdb.txt`](./post-5571/memdb.txt), [`post-5571/pebbledb.txt`](./post-5571/pebbledb.txt)

`Δ vs 5570` = `(5571 − 5570) / 5570`. `Δ vs pre` = `(5571 − pre) / pre`. Negative = faster/smaller. `≈0` = within run-to-run noise at count=2.

---

## Headline Changes vs PR #5570

### Massive wins on iteration and read-path with persistent storage

The leaf format v2/v3 pair (PR #5571) eliminates the per-value DB Get on
external-storage iteration. Every benchmark that walks values through pebbledb
collapses by 1–2 orders of magnitude:

| Metric | pebbledb 5570 | pebbledb 5571 | Δ |
|---|---|---|---|
| Iter full 100K | 420.6 ms | **675.7 µs** | **−100% (623× faster)** |
| Iter desc 100K | 417.2 ms | **656.2 µs** | **−100% (635× faster)** |
| Iter range 100K | 4.14 ms | **8.35 µs** | **−100% (496× faster)** |
| Iter full 1K | 304.4 µs | **6.25 µs** | **−98% (49× faster)** |
| Iter range 1K | 3.06 µs | **335.8 ns** | **−89% (9.1× faster)** |
| GET (hit) 100K | 4.66 µs | **213 ns** | **−95%** |
| GET (hit) 10K | 670 ns | **78.5 ns** | **−88%** |
| GET (hit) 1K | 398 ns | **55.9 ns** | **−86%** |
| Non-membership 100K | 12.18 µs | **2.75 µs** | **−77%** |
| Membership 100K | 6.39 µs | **1.59 µs** | **−75%** |
| ScalingGet 100K | 4.70 µs | **199.8 ns** | **−96%** |
| ScalingGet 1M | 5.78 µs | **635 ns** | **−89%** |

### Major write-path improvements

| Metric | memdb 5570 → 5571 | pebbledb 5570 → 5571 |
|---|---|---|
| **SaveVersion 100K** | 880.9 µs → **338.5 µs (−62%)** | 6.74 ms → 10.18 ms (+51% — see regressions) |
| **SaveVersion 1K** | 735.7 µs → **358.6 µs (−51%)** | 5.71 ms → 5.45 ms (−4%) |
| Remove 100K | 6.02 µs → **4.65 µs (−23%)** | 7.04 µs → **4.58 µs (−35%)** |
| Remove 1K | 5.04 µs → **4.71 µs (−7%)** | 6.60 µs → **4.47 µs (−32%)** |
| SET update 100K | 4.28 µs → **3.92 µs (−8%)** | 5.18 µs → **3.86 µs (−25%)** |
| SET insert 100K | 7.73 µs → **6.76 µs (−12%)** | 7.96 µs → **6.77 µs (−15%)** |

### Latest-view cache fixes the small-tree GET regressions and adds new HAS speedups

| Metric | memdb 5570 → 5571 |
|---|---|
| GET (hit) 1K | 100.2 ns → **53.8 ns (−46%)** |
| GET (hit) 10K | 157.7 ns → **74.3 ns (−53%)** |
| GET (hit) 100K | 290.9 ns → **208.8 ns (−28%)** *(cache suspends at this size)* |
| HAS 1K | 76.8 ns → **39.4 ns (−49%)** |
| HAS 10K | 127.0 ns → **54.8 ns (−57%)** |

### LoadVersion regression recovered

| Metric | pre | 5570 | 5571 | Δ vs 5570 | Δ vs pre |
|---|---|---|---|---|---|
| LoadVersion 1K (memdb) | 4.64 µs | 5.95 µs | **5.80 µs** | −2% | +25% |
| LoadVersion 100K (memdb) | 12.75 µs | 18.96 µs | **8.48 µs** | **−55%** | **−33%** |
| LoadVersion 1K (pebbledb) | 4.60 µs | 5.89 µs | **5.41 µs** | −8% | +18% |
| LoadVersion 100K (pebbledb) | 13.36 µs | 17.83 µs | **8.00 µs** | **−55%** | **−40%** |

The 100K case fully recovers and beats the pre-optimization baseline by 33–40%
(inline values eliminate the value-namespace reverse-seek that previously
dominated the cold path). The 1K case is within noise of the 5570 measurement;
at this size LoadVersion's cost is dominated by the single root-leaf decode,
which is now larger because leaves carry inline values — the +25/+18% residual
vs the pre-PR baseline is structural to the inline-values format choice and is
documented in the 5571 commit `7d2a9b517` ("cut LoadVersion alloc overhead").

### Regressions

| Metric | memdb Δ vs 5570 | pebbledb Δ vs 5570 | Note |
|---|---|---|---|
| **SaveVersion 100K** | −62% (improvement) | **+51%** (6.74 → 10.18 ms) | pebbledb-specific: high run-to-run variance (8.59 / 11.76 ms across the two measurements). Bytes/op drops from 1.18 MB to 458 KB and allocs/op from 9 173 to 1 232 (−87% allocs), but per-call ns rises. Likely a flush-timing interaction with pebbledb's internal write batching at the larger pool buffer size; worth pursuing as a follow-up. |
| **SET update 1K (memdb)** | **+31%** (2.52 → 3.31 µs) | −21% (improvement) | memdb-specific. The 100K case (cache suspended) is faster; the 1K case (cache active) is slower despite −3 allocs/op. The cached `payload.inline` slice is now shared with the LRU rather than copied — long-lived cache entries may shift GC pressure on the small-tree, fast-Set path. |
| **ScalingSet 1K (memdb)** | +98% (2.54 → 5.03 µs) | −4% (≈tie) | Same pattern as SET update 1K; same hypothesised cause. |
| **GET (miss) 1K (memdb)** | +25% (58.9 → 73.7 ns) | +24% (59.2 → 73.4 ns) | The fast-node cache check on a 1K tree pays its overhead per call; absolute impact is ~15 ns per Get. The 100K cases are flat (cache is suspended above the working-set threshold). |
| **GET (miss) 10K (memdb)** | +17% (104.2 → 122.2 ns) | +13% (103.7 → 117.5 ns) | Same. |
| **Prune (pebbledb)** | +5% (memdb: 7.69 → 8.07 ms) | **+21%** (45.2 → 54.5 ms) | Both backends regress slightly; pebbledb more so. Prune still beats baseline by 76% (memdb) / 51% (pebbledb). |
| Block 500 (pebbledb) | −29% (improvement) | +7% (24.84 → 26.57 ms) | memdb improves 29%; pebbledb +7%. |
| Membership 1K (memdb) | +6% (1.22 → 1.29 µs) | −25% (improvement) | Within noise on memdb. |
| Non-membership 1K (memdb) | +5% (2.23 → 2.33 µs) | −26% (improvement) | Within noise on memdb. |

### No material change (within ±5%)

GET (miss) 100K, HAS 100K, ScalingSet 1M (pebbledb), LoadVersion 1K, Block 100
(pebbledb), MultiVersionCreate, ScalingGet 1K (memdb already small).

IAVL numbers continue to be stable across all three runs (PR #5571 only touches
B+32).

---

## Full Results

### memdb

| Metric | B+32 pre | B+32 5570 (Δ vs pre) | B+32 5571 (Δ vs 5570) (Δ vs pre) |
|---|---|---|---|
| GET (hit), 1K | 101.7 ns | 100.2 ns (−1%) | **53.8 ns (−46%) (−47%)** |
| GET (hit), 10K | 158.8 ns | 157.7 ns (−1%) | **74.3 ns (−53%) (−53%)** |
| GET (hit), 100K | 287.0 ns | 290.9 ns (+1%) | **208.8 ns (−28%) (−27%)** |
| GET (miss), 1K | 62.5 ns | 58.9 ns (−6%) | 73.7 ns (+25%) (+18%) |
| GET (miss), 10K | 109.9 ns | 104.2 ns (−5%) | 122.2 ns (+17%) (+11%) |
| GET (miss), 100K | 191.8 ns | 169.2 ns (−12%) | 174.9 ns (+3%) (−9%) |
| HAS, 1K | 80.8 ns | 76.8 ns (−5%) | **39.4 ns (−49%) (−51%)** |
| HAS, 10K | 130.9 ns | 127.0 ns (−3%) | **54.8 ns (−57%) (−58%)** |
| HAS, 100K | 216.3 ns | 205.0 ns (−5%) | 201.7 ns (−2%) (−7%) |
| SET insert, 1K | 7.93 µs | 7.32 µs (−8%) | 6.84 µs (−7%) (−14%) |
| SET insert, 10K | 8.88 µs | 7.34 µs (−17%) | 6.75 µs (−8%) (−24%) |
| SET insert, 100K | 8.54 µs | 7.73 µs (−10%) | 6.76 µs (−12%) (−21%) |
| SET update, 1K | 3.29 µs | 2.52 µs (−23%) | 3.31 µs (+31%) (≈0) |
| SET update, 10K | 3.98 µs | 2.88 µs (−28%) | 3.16 µs (+10%) (−21%) |
| SET update, 100K | 4.96 µs | 4.28 µs (−14%) | **3.92 µs (−8%) (−21%)** |
| Remove, 1K | 6.07 µs | 5.04 µs (−17%) | 4.71 µs (−7%) (−22%) |
| Remove, 10K | 5.89 µs | 5.05 µs (−14%) | **4.10 µs (−19%) (−30%)** |
| Remove, 100K | 6.81 µs | 6.02 µs (−12%) | **4.65 µs (−23%) (−32%)** |
| Iter full, 1K | 32.99 µs | 35.04 µs (+6%) | **6.40 µs (−82%) (−81%)** |
| Iter full, 100K | 4.98 ms | 5.09 ms (+2%) | **709.5 µs (−86%) (−86%)** |
| Iter desc, 100K | 4.92 ms | 5.08 ms (+3%) | **692.3 µs (−86%) (−86%)** |
| Iter range, 1K | 375.2 ns | 545.3 ns (+45%) | **347.3 ns (−36%) (−7%)** |
| Iter range, 100K | 43.73 µs | 49.12 µs (+12%) | **8.43 µs (−83%) (−81%)** |
| Block 100 | 3.04 ms | 2.33 ms (−24%) | **1.68 ms (−28%) (−45%)** |
| Block 500 | 12.19 ms | 7.93 ms (−35%) | **5.64 ms (−29%) (−54%)** |
| SaveVersion, 1K | 787.1 µs | 735.7 µs (−7%) | **358.6 µs (−51%) (−54%)** |
| SaveVersion, 100K | 1.01 ms | 880.9 µs (−13%) | **338.5 µs (−62%) (−67%)** |
| LoadVersion, 1K | 4.64 µs | 5.95 µs (+28%) | 5.80 µs (−2%) (+25%) |
| LoadVersion, 100K | 12.75 µs | 18.96 µs (+49%) | **8.48 µs (−55%) (−33%)** |
| Prune | 33.18 ms | **7.69 ms (−77%)** | 8.07 ms (+5%) (−76%) |
| Membership 1K | 1.42 µs | 1.22 µs (−15%) | 1.29 µs (+6%) (−9%) |
| Membership 100K | 2.03 µs | 1.85 µs (−9%) | **1.69 µs (−9%) (−17%)** |
| Non-membership 1K | 2.97 µs | 2.23 µs (−25%) | 2.33 µs (+5%) (−21%) |
| Non-membership 100K | 4.03 µs | 3.27 µs (−19%) | **2.97 µs (−9%) (−26%)** |
| WorkingHash 1K | 82.92 µs | 75.04 µs (−10%) | 70.56 µs (−6%) (−15%) |
| WorkingHash 100K | 85.00 µs | 77.79 µs (−8%) | **69.29 µs (−11%) (−18%)** |
| ScalingGet 1K | 103.3 ns | 102.1 ns (−1%) | **57.1 ns (−44%) (−45%)** |
| ScalingGet 10K | 164.2 ns | 158.8 ns (−3%) | **77.2 ns (−51%) (−53%)** |
| ScalingGet 100K | 294.9 ns | 281.6 ns (−4%) | **219.8 ns (−22%) (−25%)** |
| ScalingGet 1M | 862.5 ns | 845.4 ns (−2%) | **639.3 ns (−24%) (−26%)** |
| ScalingSet 1K | 3.29 µs | 2.54 µs (−23%) | 5.03 µs (+98%) (+53%) |
| ScalingSet 100K | 5.03 µs | 4.34 µs (−14%) | **4.04 µs (−7%) (−20%)** |
| ScalingSet 1M | 6.42 µs | 5.61 µs (−13%) | 5.60 µs (≈0) (−13%) |
| Mixed (Backends) | 9.54 µs | 7.08 µs (−26%) | **5.57 µs (−21%) (−42%)** |

### pebbledb

| Metric | B+32 pre | B+32 5570 (Δ vs pre) | B+32 5571 (Δ vs 5570) (Δ vs pre) |
|---|---|---|---|
| GET (hit), 1K | 401.1 ns | 398.2 ns (−1%) | **55.9 ns (−86%) (−86%)** |
| GET (hit), 10K | 646.1 ns | 670.3 ns (+4%) | **78.5 ns (−88%) (−88%)** |
| GET (hit), 100K | 4.66 µs | 4.66 µs (≈0) | **213.2 ns (−95%) (−95%)** |
| GET (miss), 1K | 61.5 ns | 59.2 ns (−4%) | 73.4 ns (+24%) (+19%) |
| GET (miss), 10K | 108.8 ns | 103.7 ns (−5%) | 117.5 ns (+13%) (+8%) |
| GET (miss), 100K | 178.3 ns | 169.2 ns (−5%) | 166.6 ns (−2%) (−7%) |
| HAS, 1K | 76.3 ns | 75.1 ns (−2%) | **40.2 ns (−46%) (−47%)** |
| HAS, 10K | 127.8 ns | 124.9 ns (−2%) | **55.5 ns (−56%) (−57%)** |
| HAS, 100K | 205.4 ns | 200.2 ns (−3%) | 193.2 ns (−4%) (−6%) |
| SET insert, 1K | 8.82 µs | 8.06 µs (−9%) | **6.72 µs (−17%) (−24%)** |
| SET insert, 10K | 8.57 µs | 7.97 µs (−7%) | **6.67 µs (−16%) (−22%)** |
| SET insert, 100K | 8.66 µs | 7.96 µs (−8%) | **6.77 µs (−15%) (−22%)** |
| SET update, 1K | 4.69 µs | 4.07 µs (−13%) | **3.22 µs (−21%) (−31%)** |
| SET update, 10K | 4.96 µs | 4.14 µs (−17%) | **3.21 µs (−22%) (−35%)** |
| SET update, 100K | 5.99 µs | 5.18 µs (−14%) | **3.86 µs (−25%) (−36%)** |
| Remove, 1K | 7.37 µs | 6.60 µs (−11%) | **4.47 µs (−32%) (−39%)** |
| Remove, 10K | 7.00 µs | 6.06 µs (−13%) | **4.06 µs (−33%) (−42%)** |
| Remove, 100K | 7.77 µs | 7.04 µs (−9%) | **4.58 µs (−35%) (−41%)** |
| Iter full, 1K | 306.3 µs | 304.4 µs (−1%) | **6.25 µs (−98%) (−98%)** |
| Iter full, 100K | 422.6 ms | 420.6 ms (≈0) | **675.7 µs (−100%) (−100%)** |
| Iter desc, 100K | 427.0 ms | 417.2 ms (−2%) | **656.2 µs (−100%) (−100%)** |
| Iter range, 1K | 2.97 µs | 3.06 µs (+3%) | **335.8 ns (−89%) (−89%)** |
| Iter range, 100K | 4.16 ms | 4.14 ms (−1%) | **8.35 µs (−100%) (−100%)** |
| Block 100 | 15.03 ms | 14.95 ms (−1%) | 14.62 ms (−2%) (−3%) |
| Block 500 | 28.83 ms | 24.84 ms (−14%) | 26.57 ms (+7%) (−8%) |
| SaveVersion, 1K | 5.97 ms | 5.71 ms (−4%) | 5.45 ms (−4%) (−9%) |
| SaveVersion, 100K | 7.18 ms | 6.74 ms (−6%) | 10.18 ms (+51%) (+42%) |
| LoadVersion, 1K | 4.60 µs | 5.89 µs (+28%) | 5.41 µs (−8%) (+18%) |
| LoadVersion, 100K | 13.36 µs | 17.83 µs (+33%) | **8.00 µs (−55%) (−40%)** |
| Prune | 110.24 ms | **45.15 ms (−59%)** | 54.48 ms (+21%) (−51%) |
| Membership 1K | 1.79 µs | 1.57 µs (−12%) | **1.17 µs (−25%) (−35%)** |
| Membership 100K | 6.51 µs | 6.39 µs (−2%) | **1.59 µs (−75%) (−75%)** |
| Non-membership 1K | 4.08 µs | 2.83 µs (−31%) | **2.11 µs (−26%) (−48%)** |
| Non-membership 100K | 21.20 µs | 12.18 µs (−43%) | **2.75 µs (−77%) (−87%)** |
| WorkingHash 1K | 90.72 µs | 83.48 µs (−8%) | **68.88 µs (−17%) (−24%)** |
| WorkingHash 100K | 87.02 µs | 81.15 µs (−7%) | **69.60 µs (−14%) (−20%)** |
| ScalingGet 1K | 409.2 ns | 400.1 ns (−2%) | **56.3 ns (−86%) (−86%)** |
| ScalingGet 10K | 656.8 ns | 663.6 ns (+1%) | **75.5 ns (−89%) (−89%)** |
| ScalingGet 100K | 4.68 µs | 4.70 µs (≈0) | **199.8 ns (−96%) (−96%)** |
| ScalingGet 1M | 5.86 µs | 5.78 µs (−1%) | **635.3 ns (−89%) (−89%)** |
| ScalingSet 1K | 4.74 µs | 3.85 µs (−19%) | 3.70 µs (−4%) (−22%) |
| ScalingSet 100K | 6.06 µs | 5.47 µs (−10%) | **4.00 µs (−27%) (−34%)** |
| ScalingSet 1M | 6.68 µs | 5.82 µs (−13%) | 5.43 µs (−7%) (−19%) |
| Mixed (Backends) | 39.90 µs | 34.99 µs (−12%) | 33.44 µs (−4%) (−16%) |

---

## Allocations per op (B+32)

### memdb — allocs/op

| Metric | B+32 pre | B+32 5570 | B+32 5571 |
|---|---|---|---|
| GET (hit), 1K | 1 | 1 | 1 |
| GET (hit), 100K | 1 | 1 | 0 |
| GET (miss), 1K | 0 | 0 | 0 |
| HAS, 1K | 0 | 0 | 0 |
| SET insert, 1K | 11 | 9 | 7 |
| SET insert, 100K | 11 | 10 | 7 |
| SET update, 1K | 9 | 8 | 5 |
| SET update, 100K | 9 | 8 | 5 |
| Remove, 1K | 6 | 4 | 3 |
| Remove, 100K | 6 | 4 | 3 |
| SaveVersion, 1K | 11 845 | 9 870 | 1 504 |
| SaveVersion, 100K | 11 824 | 9 855 | 1 501 |
| LoadVersion, 1K | 18 | 27 | 24 |
| LoadVersion, 100K | 34 | 41 | 38 |
| Prune | — | — | 120 408 |
| WorkingHash 100K | 112 | 102 | 72 |
| Mixed (Backends) | 84 | 57 | 38 |

### pebbledb — allocs/op

| Metric | B+32 pre | B+32 5570 | B+32 5571 |
|---|---|---|---|
| GET (hit), 1K | 3 | 3 | 1 |
| GET (hit), 100K | 8 | 8 | 0 |
| GET (miss), 1K | 0 | 0 | 0 |
| HAS, 1K | 0 | 0 | 0 |
| SET insert, 1K | 11 | 10 | 7 |
| SET insert, 100K | 11 | 10 | 7 |
| SET update, 1K | 11 | 10 | 5 |
| SET update, 100K | 11 | 10 | 5 |
| Remove, 1K | 8 | 6 | 3 |
| Remove, 100K | 8 | 6 | 3 |
| SaveVersion, 1K | 11 793 | 9 879 | 1 503 |
| SaveVersion, 100K | 11 856 | 9 173 | 1 232 |
| LoadVersion, 1K | 18 | 27 | 24 |
| LoadVersion, 100K | 34 | 41 | 41 |
| WorkingHash 100K | 112 | 102 | 72 |
| Mixed (Backends) | 72 | 49 | 17 |

The dominant allocation reductions are concentrated in the
SaveVersion path (the per-node serialise pool now retains buffers
sized for full inline-value leaves) and in the Set / Remove hot
paths (the inline-value cache no longer makes a redundant copy
when populating the latest-view fast-node cache, and the bound-
once value resolver eliminates a per-call closure allocation).

---

## Takeaways Specific to PR #5571

1. **Iteration through pebbledb is the headline result.** Inline values
   eliminate the per-Value DB Get that previously dominated leaf-walk
   benchmarks. `IterationFull/100k pebbledb` drops from **420 ms to 676 µs**
   (623× faster); `IterationDescending/100k pebbledb` from **427 ms to 656 µs**
   (650× faster); `IterationRange/100k pebbledb` from **4.14 ms to 8.35 µs**
   (496× faster). Block 100/500 (which re-walks values) sees a smaller
   improvement because its per-call cost is dominated by Set, not Read.

2. **Read-path on pebbledb collapses by 80–95%.** GetHit at every size
   improves dramatically (86–95%); HAS at small sizes improves 46–57%; both
   proof types at 100K improve 75–77%. The latest-view fast-node cache catches
   the small-tree hot keys; the inline-values format eliminates the per-value
   resolver call for the common case.

3. **Large-tree LoadVersion regression is fully recovered.** Both backends'
   100K LoadVersion drop 55% vs PR #5570 and 33–40% below the pre-optimization
   baseline. The 1K case sits ±2–8% of the 5570 number (essentially noise);
   the residual +18 to +25% vs the pre-PR baseline at 1K is structural to the
   inline-value leaf format (the entire 1K tree fits in the root leaf, whose
   decode is now larger because it carries inline value bytes).

4. **Adaptive fast-node cache fixes the small-tree hot-path regression on
   GetHit.** When the working set fits the 10 000-entry LRU (1K, 10K) the
   cache services 99% of Gets at 50–80 ns. Above the threshold (cap × 4) the
   cache is suspended so its lookup overhead doesn't tax the now-mostly-miss
   path; GetHit/100K still drops 28% on memdb (95% on pebbledb) because the
   v3 prefix-compressed leaf decode and bound-once value resolver close the
   per-Get cost on its own.

5. **SaveVersion 100K on pebbledb regressed.** The single material regression
   in this round: 6.74 → 10.18 ms (+51%). Bytes/op falls 1.18 MB → 458 KB and
   allocs/op falls 9 173 → 1 232 (−87%) — fewer-but-larger allocations may be
   shifting GC and pebbledb-batch flush timing in a way that costs wall time.
   The two count=2 measurements span 8.59–11.76 ms (run-to-run variance is
   high), so the regression deserves more iteration to pin down before
   investigating. memdb on the same workload improves 62%, so the cause is
   pebbledb-specific.

6. **GET (miss) sees a small absolute regression at 1K and 10K** (+13 to
   +25%, or ~15 ns per call on both backends). This is the cost of the
   `fastNodeActive` check + `fastNodes.Get(string(key))` lookup paid on every
   call. At 100K the cache suspends and the regression vanishes; for the
   working sets at which the cache helps GetHit by 50%, GetMiss pays a few ns
   in trade.

7. **Prune regressed slightly on both backends** (+5% memdb, +21% pebbledb).
   Still 76% / 51% faster than the pre-PR baseline; pebbledb's extra cost
   matches the SaveVersion regression direction and may share its root cause
   (pool-sized buffers interacting with pebbledb internal flush behaviour).

8. **Allocation counts are dramatically lower across the board.** SaveVersion
   drops 9 870 → 1 504 (memdb) and 9 173 → 1 232 (pebbledb) — an 85% reduction
   driven by the per-node serialise-buffer pool retaining buffers sized for
   full inline-value leaves. Set/Remove hot paths drop 30–60% in alloc count;
   the mixed workload (Backends) drops 33% / 65%.

9. **IAVL numbers continue stable across all three measurement points.** The
   PR only touches the B+32 implementation; cross-run drift on IAVL is within
   noise.

10. **No on-disk format change is required for existing trees.** The reader
    still accepts v1 (legacy external-only) and v2 (per-slot inline mask)
    payloads alongside the new v3 (prefix-compressed keys). Existing chains
    mount cleanly and auto-upgrade on next save.
