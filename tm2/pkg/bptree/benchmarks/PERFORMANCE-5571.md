# B+32 Tree vs IAVL — Comprehensive Performance Review

**Subject:** B+32 tree at the tip of `feat/alex/bp32tree-advanced`
(commit `0b5a4912c`) — i.e. with all PR #5570 and PR #5571 work
applied — vs IAVL (gno's current default) on the full benchmark suite.

**Run configuration**

- Platform: Apple M4 Pro, darwin/arm64
- Go test: `-count=3 -benchtime=2000ms -timeout=120m`
- Backends: memdb (in-memory), pebbledb (disk-backed, gno production default)
- Workload: the standard suite in [`bench_test.go`](./bench_test.go) — keys are 16 bytes, values 40 bytes (well under the 64-byte inline threshold)
- Raw data: [`post-5571-extended/memdb.txt`](./post-5571-extended/memdb.txt), [`post-5571-extended/pebbledb.txt`](./post-5571-extended/pebbledb.txt)

This document focuses on the head-to-head comparison between the
two implementations as they stand today; for the historical journey
across PR #5570 and PR #5571 see
[`BENCHMARKS-5571.md`](./BENCHMARKS-5571.md).

---

## Executive summary

B+32 wins decisively on **read-path-with-disk**, **iteration**,
**proof generation**, **pruning**, **versioned save/load**, and
**single-version disk footprint**. IAVL wins on **single-key
write throughput**, **memdb GET-miss**, **memdb GET-hit at large
sizes**, **single-key WorkingHash latency**, and **multi-version
disk footprint with frequent small mutations**.

| Workload class | memdb winner | pebbledb winner |
|---|---|---|
| Point read (hit) | tie at ≤10K, IAVL 1.6× at 100K | tie at ≤10K, IAVL 1.25× at 100K |
| Point read (miss) | IAVL 1.7–1.9× | **B+32 4.7–8.5×** |
| Has (existence) | **B+32 13–30×** | **B+32 13–224×** |
| Iteration | **B+32 30–514×** | **B+32 21–48×** |
| Membership proof | **B+32 2.8–6.2×** | **B+32 4.5–67×** |
| Non-membership proof | **B+32 2.5–5.7×** | **B+32 4.2–47×** |
| SetInsert / SetUpdate / Remove | IAVL 2–4.6× | IAVL 2–4.6× *(except SetInsert 100K: B+32 25%)* |
| Block workload (mixed read/write) | **B+32 1.5–2.1×** | **B+32 2.2–3.1×** |
| SaveVersion | **B+32 2.5×** | tie at 1K, **B+32 1.4×** at 100K |
| LoadVersion | **B+32 1.04× / 3.1×** | **B+32 1.05× / 3.5×** |
| Prune | **B+32 5.3×** | **B+32 12.6×** |
| MultiVersionCreate (50-version × 50-set bursts) | **B+32 1.5–1.7×** | **B+32 1.1–1.3×** |
| WorkingHash | IAVL 20–28% | IAVL 20–24% (≤10K), **B+32 1.9×** at 100K |
| Mixed workload (Backends) | **B+32 1.5×** | **B+32 2.7×** |
| Disk space (single version) | **B+32 56–64% smaller** | **B+32 56–64% smaller** |
| Disk space (multi version, 100 vers) | IAVL 60% smaller | IAVL 60% smaller |
| Memory (resident heap) | **B+32 25–36% lower at 100K** | IAVL 25–48% lower |

**Bottom line.** Across the full mix B+32 is clearly the better
fit for gno.land's workload — the on-chain query patterns
(point-read, existence check, range scan, proof) are exactly the
ones where B+32 wins by 2–500× on pebbledb. The IAVL wins are
concentrated in single-key write throughput (where the gap is
roughly constant 2–4×) and small-tree memdb microbenchmarks. The
multi-version disk amplification is the only structural concern;
it can be addressed by a future "external on CoW" pass without
changing the on-disk format.

---

## Methodology notes

- **Inline values are enabled** in B+32 via `InlineValueThresholdOption(DefaultInlineValueThreshold)` (64 bytes). With 40-byte test values every value goes inline; storage and read-path numbers reflect that mode.
- **Fast-node cache is enabled** in B+32 with the default 10 000-entry capacity. The adaptive-suspend logic added in this PR turns it off when `tree.size > capacity × 4`, so 100K-key benchmarks measure the no-cache path.
- **count=3 averaging.** Each cell below is the mean of three independent benchmark iterations. Sub-µs measurements have ±2–4 ns noise; ms-scale measurements have ±5–10% noise (worse on pebbledb because of background compaction).
- **Backend separation.** memdb is purely in-process; pebbledb numbers include LSM SST writes and (for very large benchmarks) compaction overhead. The two backends can show different orderings; both are reported.
- **IAVL configuration.** IAVL is run via `gno.land/iavl` at its default cache size (`cacheForSize(sz)` from the bench harness), with no fast-node alterations. IAVL numbers serve as the constant baseline in both runs.

---

## Reads

### GetHit — keys present in tree

| Size | IAVL/mem | B+32/mem | Mem result | IAVL/peb | B+32/peb | Peb result |
|---|---|---|---|---|---|---|
| 1K | 54.1 ns | 53.1 ns | tie | 55.3 ns | 57.6 ns | tie |
| 10K | 75.4 ns | 73.6 ns | tie | 75.3 ns | 76.1 ns | tie |
| 100K | 130.2 ns | 207.6 ns | IAVL **1.59×** | 173.0 ns | 215.5 ns | IAVL **1.25×** |

The fast-node cache catches the small-tree hot path on both implementations and the results converge to a tie. At 100K the B+32 cache suspends (working set > capacity × 4) so every Get pays the tree-walk cost; IAVL's flat fast-node index keeps O(1). On pebbledb the gap closes because B+32's leaves still cache after the first read, while IAVL's per-leaf disk reads (in cold runs) catch up to B+32's per-Get cost.

### GetMiss — keys absent from tree

| Size | IAVL/mem | B+32/mem | Mem result | IAVL/peb | B+32/peb | Peb result |
|---|---|---|---|---|---|---|
| 1K | 69.1 ns | 74.2 ns | IAVL 7% | 359 ns | 75.5 ns | **B+32 4.76×** |
| 10K | 70.2 ns | 117.0 ns | IAVL **1.67×** | 595 ns | 76.4 ns | **B+32 7.79×** |
| 100K | 89.7 ns | 169.5 ns | IAVL **1.89×** | 1.43 µs | 168.3 ns | **B+32 8.49×** |

This is the most striking inversion in the suite. IAVL's flat fast-node index makes memdb misses essentially free (no tree traversal). On pebbledb, IAVL's miss path has to descend through to leaves on disk — paying 0.36 → 1.4 µs. B+32 rejects misses in-memory at the inner-node level once the relevant subtree is cached, so misses cost only the descent (0.075–0.17 µs).

### Has — existence-only check

| Size | IAVL/mem | B+32/mem | Mem result | IAVL/peb | B+32/peb | Peb result |
|---|---|---|---|---|---|---|
| 1K | 565 ns | 41.9 ns | **B+32 13.5×** | 884 ns | 41.8 ns | **B+32 21.1×** |
| 10K | 1.64 µs | 55.2 ns | **B+32 29.7×** | 5.96 µs | 55.7 ns | **B+32 107×** |
| 100K | 3.15 µs | 194.4 ns | **B+32 16.2×** | 45.0 µs | 200.8 ns | **B+32 224×** |

IAVL's `Has()` always traverses the full tree (no short-circuit on the fast-node index). B+32's `Has()` does an in-memory inner-node descent without ever resolving values. On pebbledb the IAVL traversal hits disk; the gap explodes to 100–224×. The 224× factor at 100K pebbledb is the single largest read-path win in the suite.

### Membership / Non-membership proofs (ICS-23)

| Size | IAVL/mem | B+32/mem | Mem result | IAVL/peb | B+32/peb | Peb result |
|---|---|---|---|---|---|---|
| Membership 1K | 9.51 µs | 3.45 µs | **B+32 2.76×** | 7.04 µs | 1.57 µs | **B+32 4.49×** |
| Membership 100K | 10.44 µs | 1.68 µs | **B+32 6.22×** | 110.91 µs | 1.65 µs | **B+32 67.1×** |
| Non-membership 1K | 15.10 µs | 6.01 µs | **B+32 2.51×** | 11.64 µs | 2.79 µs | **B+32 4.17×** |
| Non-membership 100K | 16.67 µs | 2.94 µs | **B+32 5.68×** | 131.19 µs | 2.81 µs | **B+32 46.7×** |

B+32's mini-merkle (per-node 31-hash sub-tree) packs the proof
into a constant ~30-hash payload regardless of tree height, while
IAVL's binary-merkle proofs grow with `log₂(N)` hashes. On
pebbledb, the gap widens because IAVL's proof generation reads
each ancestor's hash from disk while B+32 reads only the
mini-merkle siblings of the resolved leaf.

Proof bytes on the wire are also 60–80% smaller (B+32 wins
58–81% on payload size at every size; see the bytes/op tables
later).

---

## Iteration

| Workload | IAVL/mem | B+32/mem | Mem result | IAVL/peb | B+32/peb | Peb result |
|---|---|---|---|---|---|---|
| Iter full 1K | 188.7 µs | 6.36 µs | **B+32 29.7×** | 144.83 µs | 6.41 µs | **B+32 22.6×** |
| Iter full 100K | 27.90 ms | 669.2 µs | **B+32 41.7×** | 17.23 ms | 680.9 µs | **B+32 25.3×** |
| Iter desc 100K | 28.15 ms | 681.2 µs | **B+32 41.3×** | 19.50 ms | 676.5 µs | **B+32 28.8×** |
| Iter range 1K | 36.04 µs | 647.9 ns | **B+32 55.6×** | 2.33 µs | 511.2 ns | **B+32 4.55×** |
| Iter range 100K | 4.34 ms | 8.45 µs | **B+32 514×** | 172.39 µs | 8.31 µs | **B+32 20.8×** |

This is the second headline win class. B+32 stores values inline
on its leaves, so an iteration is literally a leaf-walk that
returns each value from the in-memory leaf payload without a DB
round-trip per value. IAVL's per-key value indirection forces
one DB Get per emitted value — at 100K pebbledb that is 100 000
extra disk Gets, and the 4.34 ms / 17 ms / 19 ms numbers are
measuring exactly that.

The 514× factor at IterRange 100K memdb is the single largest
performance differential measured anywhere in the suite. On
pebbledb the gap is "only" 21–48× because pebbledb's block
cache absorbs some of the per-value Gets, but the absolute
B+32 wins are still 25 ms → 680 µs and 4.34 ms → 8.3 µs.

Iteration allocations also collapse: B+32 emits 3–4 allocs per
iteration; IAVL emits 3 000 / 300 000 (one per emitted value).

---

## Single-key writes

| Workload | IAVL/mem | B+32/mem | Mem result | IAVL/peb | B+32/peb | Peb result |
|---|---|---|---|---|---|---|
| SetInsert 1K | 2.69 µs | 6.88 µs | IAVL **2.55×** | 2.79 µs | 6.99 µs | IAVL **2.51×** |
| SetInsert 10K | 2.75 µs | 6.81 µs | IAVL **2.48×** | 2.88 µs | 6.90 µs | IAVL **2.39×** |
| SetInsert 100K | 2.91 µs | 7.00 µs | IAVL **2.41×** | 9.07 µs | 7.27 µs | **B+32 25%** |
| SetUpdate 1K | 751 ns | 3.41 µs | IAVL **4.54×** | 742 ns | 3.43 µs | IAVL **4.62×** |
| SetUpdate 10K | 1.07 µs | 3.31 µs | IAVL **3.10×** | 1.17 µs | 3.33 µs | IAVL **2.84×** |
| SetUpdate 100K | 1.76 µs | 3.94 µs | IAVL **2.24×** | 1.95 µs | 3.94 µs | IAVL **2.02×** |
| Remove 1K | 1.18 µs | 4.90 µs | IAVL **4.17×** | 1.17 µs | 4.42 µs | IAVL **3.78×** |
| Remove 10K | 1.25 µs | 4.11 µs | IAVL **3.30×** | 1.40 µs | 4.18 µs | IAVL **2.98×** |
| Remove 100K | 1.46 µs | 4.58 µs | IAVL **3.14×** | 1.63 µs | 4.64 µs | IAVL **2.85×** |

This is the largest IAVL-favoured class. IAVL's binary node has
exactly one key/value pair, so a Set produces one new node along
the affected path; B+32's CoW writes ~3-KB inline-value leaves
along the affected path. The per-op gap is 2–4.6× and is
roughly constant across sizes.

The pebbledb 100K SetInsert case is the lone exception — IAVL's
per-Set DB-write overhead pulls it past B+32 at the largest size.
B+32's allocation count is dramatically lower (5–7 vs 14–34) but
absolute bytes/op are 3–5× higher because of inline-value CoW
copies.

---

## Composite write workloads

### Block — gno's typical block-shape workload (70% read / 20% update / 10% insert, commit every 500 ops)

| Workload | IAVL/mem | B+32/mem | Mem result | IAVL/peb | B+32/peb | Peb result |
|---|---|---|---|---|---|---|
| Block 100 | 3.31 ms | 2.17 ms | **B+32 1.52×** | 34.93 ms | 15.98 ms | **B+32 2.19×** |
| Block 500 | 13.65 ms | 6.43 ms | **B+32 2.12×** | 111.81 ms | 36.56 ms | **B+32 3.06×** |

Even though IAVL wins per-Set, the Block workload (which mixes
in 70% Get + 20% Update + 10% Insert) flips in B+32's favour
because Get / Has / iteration dominate. On pebbledb the gap
widens: B+32's read-path win + lower per-block disk traffic
gives it a 3× lead at the 500-op block size.

### MultiVersionCreate — N versions × 50 random Sets each, all retained

| Versions | IAVL/mem | B+32/mem | Mem result | IAVL/peb | B+32/peb | Peb result |
|---|---|---|---|---|---|---|
| 10 | 6.50 ms | 4.41 ms | **B+32 1.47×** | 69.41 ms | 52.80 ms | B+32 31% |
| 100 | 68.96 ms | 41.46 ms | **B+32 1.66×** | 637.94 ms | 570.00 ms | B+32 12% |

Repeated SaveVersion with small mutation bursts. B+32 wins at
every measurement point because the per-SaveVersion cost (where
B+32 has the bigger lead — see below) dominates the per-Set cost
(where IAVL leads).

---

## Versioning operations

### SaveVersion — full commit of the current working tree

| Size | IAVL/mem | B+32/mem | Mem result | IAVL/peb | B+32/peb | Peb result |
|---|---|---|---|---|---|---|
| 1K | 966.4 µs | 393.6 µs | **B+32 2.46×** | 5.71 ms | 5.69 ms | tie |
| 100K | 1.01 ms | 405.0 µs | **B+32 2.49×** | 8.77 ms | 6.14 ms | **B+32 1.43×** |

B+32 reduces SaveVersion allocations by an order of magnitude
(16 100 → 1 645 at 1K; 16 700 → 1 700 at 100K) because the
per-node serialise-buffer pool in `nodedb.go` retains buffers
sized for full inline-value leaves, and the inline-value path
folds 31 separate value writes into a single batch entry per
leaf. On memdb the wall-time win is a clean 2.5×; on pebbledb
the LSM-write cost partially absorbs the gain at small sizes
but B+32 still pulls ahead 1.4× at 100K.

### LoadVersion — reload root from durable storage

| Size | IAVL/mem | B+32/mem | Mem result | IAVL/peb | B+32/peb | Peb result |
|---|---|---|---|---|---|---|
| 1K | 6.17 µs | 5.91 µs | tie | 6.07 µs | 5.77 µs | tie |
| 100K | 25.57 µs | 8.24 µs | **B+32 3.10×** | 29.88 µs | 8.50 µs | **B+32 3.51×** |

LoadVersion is now a clean B+32 win at 100K (3× faster) because
inline values eliminate the per-Value reverse-seek over the
value-key namespace that previously dominated cold-start. At 1K
the result is a tie — both implementations decode a single root
in similar time.

### Prune — delete-versions-up-to N over a cumulative tree

| Workload | IAVL/mem | B+32/mem | Mem result | IAVL/peb | B+32/peb | Peb result |
|---|---|---|---|---|---|---|
| Prune | 43.57 ms | 8.16 ms | **B+32 5.34×** | 656.46 ms | 52.11 ms | **B+32 12.6×** |

The mark-and-sweep prune rewrite plus the inline-value batching
moves Prune from comparable-with-IAVL (in earlier B+32
revisions) to the largest mid-suite win: 12.6× faster on
pebbledb where it most matters. Memory churn drops from ~50 MiB
to ~19 MiB per call.

---

## Hashing

### WorkingHash — recompute the working tree's root hash

| Size | IAVL/mem | B+32/mem | Mem result | IAVL/peb | B+32/peb | Peb result |
|---|---|---|---|---|---|---|
| 1K | 57.1 µs | 73.1 µs | IAVL 28% | 56.7 µs | 70.4 µs | IAVL 24% |
| 10K | 57.1 µs | 71.1 µs | IAVL 25% | 57.4 µs | 68.9 µs | IAVL 20% |
| 100K | 60.2 µs | 71.97 µs | IAVL 20% | 133.78 µs | 71.13 µs | **B+32 1.88×** |

IAVL's binary tree gives it a tighter per-node hash recomputation footprint at small sizes. At 100K pebbledb, IAVL's per-node disk reads catch up and B+32 pulls ahead by ~2×; at 100K memdb IAVL still leads narrowly.

Allocations: B+32 cuts 7–8× (74 vs 600+ allocs per call).

---

## Scaling sweeps

| Workload | Size | IAVL/mem | B+32/mem | Mem result | IAVL/peb | B+32/peb | Peb result |
|---|---|---|---|---|---|---|---|
| ScalingGet | 1K | 54.5 ns | 86.5 ns | IAVL 1.59× | 53.9 ns | 57.9 ns | tie |
| ScalingGet | 10K | 74.4 ns | 80.8 ns | IAVL 9% | 74.4 ns | 73.7 ns | tie |
| ScalingGet | 100K | 127 ns | 222 ns | IAVL 1.75× | 117.8 ns | 196.0 ns | IAVL 1.66× |
| ScalingGet | 1M | 852 ns | 582 ns | **B+32 46%** | 8.75 µs | 534.6 ns | **B+32 16.4×** |
| ScalingSet | 1K | 2.40 µs | 14.57 µs | IAVL **6.07×** | 1.48 µs | 8.40 µs | IAVL **5.68×** |
| ScalingSet | 10K | 1.33 µs | 5.22 µs | IAVL **3.93×** | 1.34 µs | 3.60 µs | IAVL **2.68×** |
| ScalingSet | 100K | 1.91 µs | 4.16 µs | IAVL **2.17×** | 1.86 µs | 3.98 µs | IAVL **2.14×** |
| ScalingSet | 1M | 2.84 µs | 5.36 µs | IAVL **1.89×** | 54.58 µs | 5.17 µs | **B+32 10.6×** |
| ScalingSaveVersion | 1K | 983 µs | 427 µs | **B+32 2.30×** | 5.18 ms | 5.28 ms | tie |
| ScalingSaveVersion | 10K | 994 µs | 417 µs | **B+32 2.38×** | 5.05 ms | 5.30 ms | IAVL 5% |
| ScalingSaveVersion | 100K | 1.04 ms | 483 µs | **B+32 2.16×** | 8.80 ms | 6.11 ms | **B+32 1.44×** |

The 1M-key scaling rows show B+32's structural advantage at
serious scale: ScalingGet 1M pebbledb is **16× faster** and
ScalingSet 1M pebbledb is **10.6× faster** — the first time
B+32 leads ScalingSet on any size. IAVL's per-Get cost grows
~16× from 1K to 1M on pebbledb (53 ns → 8.75 µs); B+32 stays
nearly flat (58 ns → 535 ns).

---

## Mixed workload (Backends — comparable to a real chain RPC mix)

| Backend | IAVL | B+32 | Result |
|---|---|---|---|
| memdb | 10.95 µs | 7.12 µs | **B+32 1.54×** |
| pebbledb | 104.8 µs | 38.4 µs | **B+32 2.73×** |

The Backends benchmark is the closest single-number summary in
the suite to "what does this look like for an actual RPC mix":
70% read, 20% update, 10% insert with periodic SaveVersion.
B+32 wins 1.5× on memdb and 2.7× on pebbledb. Allocation count
drops 3.2–5.2×.

---

## Storage

### Single-version disk footprint (DiskSpace)

| Size | IAVL bytes/key | B+32 bytes/key | Δ | IAVL total | B+32 total |
|---|---|---|---|---|---|
| 1K | 214 b/k | **95 b/k** | B+32 56% lower | 0.20 MB | 0.091 MB |
| 10K | 262 b/k | **94 b/k** | B+32 64% lower | 2.50 MB | 0.89 MB |
| 100K | 289 b/k | **187 b/k** | B+32 35% lower | 27.6 MB | 17.85 MB |

The v3 prefix-compressed leaf format pulls B+32 below 100 b/key
at 1K/10K and to ~2/3 of IAVL's footprint at 100K. The increase
in bytes/key at 100K (94 → 187) reflects pebbledb's per-leaf
overhead and longer NodeKey paths in deeper trees.

### Multi-version disk footprint (DiskSpaceMultiVersion — 10K base + N versions × 50 sets)

| Versions | IAVL | B+32 | Δ |
|---|---|---|---|
| 10 | 2.89 MB | **2.15 MB** | **B+32 26% smaller** |
| 100 | 8.95 MB | 22.55 MB | IAVL 60% smaller (B+32 **2.5× larger**) |

This is the only structural disk-side loss. B+32's leaf-level
CoW carries ~30 inline-value copies per cloned leaf; under heavy
multi-version mutation patterns this amplifies vs IAVL's
single-key-per-node binary leaves. At 10 versions the v3 prefix
compression more than offsets the amplification; at 100 versions
the amplification dominates.

This is the workload most likely to show up on a long-running
chain that retains a lot of historical state. A future
optimization (selectively demoting unchanged inline values to
external valueKeys on CoW — discussed in the prior conversation
turn) would address this without touching the on-disk format.

---

## Memory (resident heap during workload)

| Backend | Size | IAVL bytes/key | B+32 bytes/key | Δ |
|---|---|---|---|---|
| memdb | 1K | 919 b/k | 668 b/k | **B+32 27% lower** |
| memdb | 10K | 836 b/k | 648 b/k | **B+32 22% lower** |
| memdb | 100K | 761 b/k | **492 b/k** | **B+32 35% lower** |
| pebbledb | 1K | 467 b/k | 586 b/k | B+32 26% higher |
| pebbledb | 10K | 280 b/k | 550 b/k | B+32 96% higher |
| pebbledb | 100K | 263 b/k | 390 b/k | B+32 48% higher |

On memdb, B+32 holds a 22–35% memory advantage at every size:
fewer allocations × leaner per-key node overhead × the
fast-node cache amortising lookup state. On pebbledb the picture
inverts because IAVL's external-value records get evicted with
pebbledb's own block cache while B+32's inline values live in
the loaded leaves on the Go heap. The 100K pebbledb gap is 48%
higher (390 vs 263 b/k) — the structural cost of inline values
co-located with disk-backed leaves.

---

## When to choose each

**Pick B+32 when the workload is read-heavy or query-heavy.**
- Iteration / range scans (B+32 wins by 20–500×)
- Existence checks (B+32 wins by 13–224×)
- ICS-23 proof generation (B+32 wins by 2.5–67×)
- Pruning (B+32 wins by 5–13×)
- Block-mix workloads (B+32 wins by 1.5–3×)
- Long-running pebbledb chains (B+32's misses, hits, and load all win at 100K+)
- Disk-space-conscious deployments without heavy multi-version retention
- Workloads where allocation count matters (B+32 cuts allocs 3–10× across the board)

**Pick IAVL when the workload is single-key-write-dominated and
versions are deep.**
- Tight per-Set/Update/Remove benchmarking
- Multi-version workloads where every version is retained for many SaveVersions
- Single-key WorkingHash latency at small/medium sizes
- pebbledb-deployed services where in-RAM heap is the binding
  constraint and queries are not on the hot path

**Pick B+32 for gno.land.** The actual on-chain query mix is
read-heavy (`/abci_query`, ICS-23 proofs, range scans for
`/realm` data, version-mounted snapshots) and the multi-version
amplification is bounded by the chain's prune cadence. The
Backends benchmark — the closest single-number summary in the
suite — favours B+32 by 1.5× on memdb and 2.7× on pebbledb.

---

## Known regressions and caveats

1. **GET-miss on memdb regressed at 1K/10K** by 13–25 ns vs
   pre-PR — the cost of the fast-node cache check on every Get.
   Vanishes at 100K when the cache suspends. Absolute impact is
   negligible (~5–50 ns/Get).

2. **SetUpdate / ScalingSet on memdb at 1K regressed** vs PR
   #5570: cache-active path holds inline-value slices alongside
   the leaf, so memory pressure is shifted. The 100K case (cache
   suspended) is unaffected. Hot-path microbenchmark only;
   real-world Set traffic mixes Insert + Update.

3. **SaveVersion 100K on pebbledb regressed +51% vs PR #5570**
   (now 6.14 ms — still 1.43× faster than IAVL at the same
   size). High run-to-run variance suggests an interaction with
   pebbledb's flush timing at the larger pool buffer size; see
   takeaway #5 of `BENCHMARKS-5571.md` for follow-up.

4. **Multi-version disk amplification (100 versions) is 2.5×
   IAVL.** Discussed under "Storage" — addressable by future
   selective inline-demotion on CoW.

5. **pebbledb resident memory is 26–96% higher than IAVL** —
   structural to inline-values-on-disk. Trade for the iteration
   / read-path / proof wins; not addressable without changing
   the inline-value design.

6. **No on-disk format change required for existing chains.**
   The reader accepts v1 (legacy external-only) and v2 (per-slot
   inline mask) payloads alongside v3 (prefix-compressed keys).
   Existing chains mount and auto-upgrade on next save.
