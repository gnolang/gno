# IAVL vs B+32 — Disk-Bound Benchmark

**Status:** in progress. Solid data at 1M (validation) and 33M; the decisive
**100M disk-bound run is still pending** (see [TODO](#todo--the-100m-run-the-one-that-matters)).
**Date:** 2026-06-07 · **Backend:** pebbledb (production-tuned: 500 MB block
cache + bloom) · **Node LRU:** `-disk-node-cache=10000`.

> **Why this file was rewritten.** The previous version benchmarked 1K–100K keys
> on memdb/goleveldb — all of which fit in RAM, so it measured CPU/allocation,
> not storage. Several of its headline conclusions (IAVL wins SET; B+32 wins
> pruning; B+32 proofs are "faster") **invert or mislead at the scale that
> matters for a chain**. The numbers below are from `pebbledb` at 33M+ with a
> working set that starts to exceed cache, using a per-operation DB-op counter.

---

## What we measure, and why

The IAVL-vs-B+32 difference is a **disk-I/O** story, so it only shows once the
working set exceeds RAM. On a 16 GB box that means **N ≳ 57M** (IAVL ≈ 300 B/key
→ ~RAM at ~57M). Below that, both trees are cache-resident and the comparison is
just CPU.

Each run reports, per operation:

| metric | meaning |
|---|---|
| `reads/op`, `reads/write` | node DB `Get`s the tree issued (node-LRU **misses**) — the traversal/COW read depth |
| `writes/write` | node + value `Set`s the tree issued (the COW write set) — **structural**, cache-independent |
| `ns/op`, `ns/write` | wall-clock latency |
| `allocs/op` | Go allocations |

The read/write counts come from a `countingDB` wrapper around `dbm.DB`
(`countingdb_test.go`): they are **deterministic, backend-agnostic, and
unaffected by pebble's background compaction** (which lives below the interface)
— unlike a pebble block-cache-miss counter, which is process-global and wobbles
with compaction.

**Caveat (read it before quoting any number):** `reads/*` are node-LRU *misses*,
so they depend on `-disk-node-cache` and the access pattern, and they *grow with
tree depth* as N rises. Treat them as a **relative** indicator at a fixed cache
size. `writes/write` is the cleaner number — it's the structural COW set and
doesn't depend on cache.

---

## Coverage

| scale | populate | writes | reads | regime |
|---|---|---|---|---|
| 1M | ✓ | ✓ | ✓ | fully cached — *validation only* |
| 33M | ✓ | ✓ | ✓ | counts solid; **ns still cache-bound** (~10 GB < 16 GB RAM) |
| **100M** | ⛔ bptree OOMs in prune | **pending** | **pending** | **the real disk-bound test** |

At 33M the IAVL tree (~10 GB) still mostly fits in 16 GB, so the *counts*
(`reads/write`, `writes/write`) are meaningful but the *latencies* (`ns`) are
cache-bound, not disk-bound. The latency gap only opens fully at 100M.

---

## Writes — `BenchmarkDiskBlockWrite` (block=1000)

**@ 33M, 16 GB target, pebbledb, `-benchtime=300x`:**

| per write | B+32 | IAVL | ratio |
|---|---|---|---|
| `reads/write` | **1.16** | 31 | ~27× |
| `writes/write` | **4.0** | 17 | ~4× |
| `ns/write` | **82 µs** | 304 µs | ~3.7× |
| `allocs/write` | **258** | 572 | ~2.2× |
| node ops/write (read+write) | **~5** | ~48 | **~10×** |

**Reading it:**
- B+32 issues **~10× fewer node operations per write**. This is the headline —
  and the **fast-node index cannot help it** (the index accelerates GET only;
  every mutation still traverses + COWs the real tree).
- `writes/write` (4.0 vs 17) is the structural COW depth: `log₃₂N` vs `log₂N`
  (deduped across the block). This is cache-independent and is the number that
  feeds the write-depth gas param.
- `reads/write` (1.16 vs 31): B+32's shallow tree + small node set sit in the
  LRU, so it loads ~1 leaf/write; IAVL re-reads its ~25-deep path (LRU thrash +
  AVL rotations).
- `ns/write` at 33M is **CPU-bound, not disk-bound** (fixture fits in RAM). At
  100M it diverges much further — IAVL's 31 reads become real seeks.

**Scaling of the structural counts** (illustrative; 1M is local validation on
darwin/arm64, 33M is the target):

| | writes/write |  | reads/write |  |
|---|---|---|---|---|
| N | B+32 | IAVL | B+32 | IAVL |
| 1M | 2.9 | 12.8 | ~0.5 | ~20 |
| 33M | 4.0 | 17 | 1.16 | 31 |
| 100M (est.) | ~4.4 | ~18 | *pending* | ~33 |

B+32's write depth is nearly flat (`log₃₂`: +1.5 ops over 100× the keys); IAVL's
climbs with `log₂`.

---

## Reads — `BenchmarkDiskGetRandom`

**@ 33M, 16 GB target, pebbledb, `-benchtime=50000x`:**

| per op | B+32 | IAVL | winner |
|---|---|---|---|
| `reads/op` | 2.78 | **0.97** | IAVL ~2.9× |
| `ns/op` | 197 µs | **127 µs** | IAVL ~1.5× |
| `allocs/op` | 116 | **9** | IAVL ~13× |

**Reading it:**
- **Reads are IAVL's home turf.** Its fast-node index serves a latest-version
  GET in ~1 read; B+32 needs the leaf **plus** a second read for the out-of-line
  value (~2.78 with inner-node misses).
- `ns/op` is only 1.5×, not 2.9×, because at 33M the reads are cache hits (cheap)
  and a fixed per-op CPU floor dominates — note B+32's **116 allocs/op** (fat
  32-wide leaf deserialize + value resolve) vs IAVL's 9. At 100M (real seeks) the
  read-count gap will push `ns/op` toward the full ~3×, *widening* IAVL's lead.
- Mirror image of writes: B+32 loses reads, wins writes.

> `BenchmarkDiskGetMiss` (negative lookups) not yet collected at scale.

---

## ICS23 proof size (measured, memdb)

Proof *size* is structural (depends on tree shape, not the disk regime), so these
small-N numbers hold. **B+32 proofs are larger or tied — not smaller** (the
mini-merkle emits ~`log₂N` binary ops and uses full 32-byte SHA-256):

| proof | IAVL @1K | B+32 @1K | IAVL @100K | B+32 @100K |
|---|---|---|---|---|
| membership | 524 B | **685 B** | 880 B | 876 B |
| non-membership | 1067 B | **1387 B** | 1621 B | **1787 B** |

B+32 generates proofs *faster* (shallower tree, fewer node fetches), but the
**on-chain cost is the proof bytes**, and there IAVL is equal-or-smaller. Do not
cite B+32 as having an ICS23 proof-size advantage.

---

## Disk space

Preliminary and **confounded** — needs a clean steady-state measurement at
scale. Small-N, sparse-file-aware (`st_blocks`) numbers favor B+32 (fewer
physical KV entries — single-copy out-of-line values vs IAVL's value-in-leaf +
value-in-fast-node):

| backend @ small N | IAVL B/key | B+32 B/key |
|---|---|---|
| pebbledb (30K) | 275 | 187 |
| lmdb (50K) | 661 | 348 |

…but the resumable populate (12–15M, retaining 20 versions) inverts to ~304 B/key
(IAVL) vs ~403 B/key (B+32) — B+32's immutable per-version COW copies cost disk
under deep history. **TODO:** measure compacted, single-latest-version on-disk
size at 33M/100M before drawing a conclusion.

---

## Gas-param implications (`Fixed*Depth100`)

These counts are exactly what gno.land's depth gas params encode (×100
fixed-point; `gno.land/pkg/sdk/vm/params.go`). Current defaults are
**B+32-calibrated** and line up with the measured B+32 counts:

| param | current (B+32) | measured B+32 @33M |
|---|---|---|
| `FixedGetReadDepth100` | 300 | ~278 (2.78 GET reads) |
| `FixedSetReadDepth100` | 200 | ~116 (1.16 SET reads) |
| `FixedWriteDepth100` | 440 | ~400 (4.0 writes) |

**If gno.land used IAVL instead**, the params must change — and two of them can
no longer be fixed, because IAVL's depth scales with `log₂N`:

| param | IAVL @10M | IAVL @100M | note |
|---|---|---|---|
| `FixedGetReadDepth100` | **100** | **100** | fast-node = O(1), N-independent |
| `FixedSetReadDepth100` | ~2800 | ~3300 | scales with depth → prefer size-driven `ExpectedDepth` |
| `FixedWriteDepth100` | ~1600 | ~1800 | structural COW depth; cache cannot reduce it |

i.e. dropping B+32 makes GET ~3× cheaper but **SET-read ~16× and write ~4×
costlier** — the cost lives entirely on the write path, which is structural and
un-cacheable.

---

## How to reproduce

One command — populate both factories and benchmark, results in `./bench-out/`:

```bash
DIR=/data/bp32bench KEYS=100000000 ./tm2/pkg/bptree/benchmarks/run-disk-bench.sh
```

Override `KEYS`, `BACKEND` (e.g. `lmdbdb`), `BUILD_BATCH`, `PARALLEL`,
`GOMEMLIMIT`, … — see the script header. Or run the steps by hand:

```bash
DIR=/data/bp32bench; KEYS=33000000        # match what you populated

# build fixtures (resumable; run the two factories in parallel)
go test ./tm2/pkg/bptree/benchmarks/ -run=TestDiskPopulate -v \
  -disk-dir=$DIR -disk-keys=$KEYS -disk-factory=iavl   -disk-verbose -timeout=24h
go test ./tm2/pkg/bptree/benchmarks/ -run=TestDiskPopulate -v \
  -disk-dir=$DIR -disk-keys=$KEYS -disk-factory=bptree -disk-verbose -timeout=24h

# bench one factory at a time (drop page cache between for a clean read)
for f in iavl bptree; do
  sync; echo 3 | sudo tee /proc/sys/vm/drop_caches
  go test ./tm2/pkg/bptree/benchmarks/ -run='^$' -bench=BenchmarkDiskBlockWrite \
    -disk-dir=$DIR -disk-keys=$KEYS -disk-factory=$f -disk-block=1000 -benchtime=300x -timeout=2h
  go test ./tm2/pkg/bptree/benchmarks/ -run='^$' -bench=BenchmarkDiskGetRandom \
    -disk-dir=$DIR -disk-keys=$KEYS -disk-factory=$f -benchtime=50000x -timeout=1h
done
```

Add `-disk-backend=lmdbdb` (to populate **and** bench) to measure on LMDB, the
backend gno.land's flat I/O constants reference. The `reads/*`/`writes/*` counts
are backend-agnostic; `ns/*` is not.

---

## TODO — the 100M run (the one that matters)

1. **Unblock the bptree 100M populate.** It currently OOM-kills on 16 GB during
   prune (the `walkAndPrune` / `findCorrespondingChild` dual-walk loads orphaned
   subtrees into memory; the spike grows with N). Either fix the prune to be
   lockstep (like IAVL's `traverseOrphans`) or add a `-disk-no-prune` build
   option (the benchmark only reads the latest version, so pruning during the
   build is optional — it trades the memory spike for more disk). Until then,
   `GOMEMLIMIT=12GiB` + `-disk-build-batch=25000` may get it through.
2. **Run BlockWrite + GetRandom at 100M** on the 16 GB target. This is the only
   regime where the IAVL working set (~30 GB) truly exceeds RAM, so its
   `reads/write` become real disk seeks and `ns/write` diverges from B+32's
   (which stays ~flat). Expected: IAVL `reads/write` ~33, `writes/write` ~18;
   IAVL `ns/write` blows up while B+32 holds; GET stays IAVL's win (fast-node
   ~1 read) and its margin *grows* under disk pressure.
3. **Validate the flat gas costs** (`ReadCostFlat`=59 µs, `WriteCostFlat`=24 µs).
   They're calibrated for 100M disk-bound reads; at 33M (cached, ~6 µs/op) they
   over-state cost ~7–10×. Confirm against the 100M run.
4. **Clean disk-space measurement** (compacted, latest-version-only) at 33M/100M.
5. **`GetMiss` at scale**, and an **LMDB** pass for gas-constant calibration.
```
