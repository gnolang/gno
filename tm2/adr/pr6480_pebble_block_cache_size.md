# PR6480: Size pebble's block cache and memtable from a gnoland-1 replay

## Status

Accepted. This PR also carries #6467 (16 MiB memtable) and the part of #6265
that opens `NewPebbleDB` with `DefaultPebbleOptions()`.

## Context

`NewPebbleDB` (`tm2/pkg/db/pebbledb/pebbledb.go`) is the registered creator for
the pebbledb backend, so it opens the app store (`gno.land/pkg/gnoland/app.go`)
and, with the default `db_backend`, the `blockstore` and `state` databases. It
passed an empty `pebble.Options`: an 8 MiB block cache and 4 MiB memtables.
`DefaultPebbleOptions()` in the same file (500 MB cache, 64 MiB memtable, bloom
filters, 3 compactions) was only used by benchmarks
(`tm2/pkg/bptree/benchmarks`, `gnovm/cmd/benchstore`), so they measured a
configuration no node ran.

pebble reserves memtable memory out of the block cache: each memtable arena
calls `Cache.Reserve`, and the memtable kept for reuse after a flush keeps its
reservation. Memtables grow one doubling per rotation from 256 KiB whatever the
write rate, so every long-running database reaches 2 x `MemTableSize` reserved.
At the defaults that is all 8 MiB, and the cache then holds no blocks at all:
every point lookup reads its index and data blocks from outside the cache.

The 4 MiB memtable also rotates, flushes to L0 and triggers compactions several
times per busy block.

## Measurement: gnoland-1 replay

gnoland-1 blocks replayed through the app with gnoreplay, the pebble options of
the replay's app and state databases taken from the environment, everything
else identical:

- v1.7.1 tree, which replays history up to the v1.8.0 halt with no result
  divergence (master diverges from genesis on gas); it commits every block, as
  a node at the tip does.
- `CGO_ENABLED=0`, go1.25.9, as release binaries are built; 4 vCPU microVM; the
  database on tmpfs, so reads are page-cache hits; fsync disabled.
- Three windows, each run from a fresh copy of a snapshot that a stock replay
  from genesis left at the window's start:
  - W1, heights 250,001 to 280,000, 16,204 txs;
  - W2, heights 1,030,001 to 1,070,000, 10,499 txs, the densest stretch;
  - W3, heights 1,580,001 to 1,620,000, 7,065 txs, the largest state.
- 2 reps per arm (3 for stock), arm order reversed on the second rep;
  rep-to-rep spread mostly under 2%, up to 5% on W3.
- 0 result, gas or app hash divergences in all 69 runs.

Blocks/s relative to stock (stock: 203, 115 and 328 blocks/s), then W2's CPU
per block, peak RSS of the replay process, block cache held at the end, p99
`ApplyBlock` time, and memtable flushes:

| arm | W1 | W2 | W3 | CPU/block | RSS | cache held | p99 apply | flushes |
|---|--:|--:|--:|--:|--:|--:|--:|--:|
| stock: 8 MiB cache, 4 MiB memtable | | | | 16.1 ms | 3,016 MB | 0 MB | 113 ms | 834 |
| 64 MiB cache | +32% | +23% | +31% | 12.2 ms | 3,002 MB | 44 MB | 95 ms | 834 |
| 16 MiB memtable | +4% | +6% | +1% | 13.8 ms | 3,028 MB | 0 MB | 106 ms | 181 |
| **64 MiB cache, 16 MiB memtable** | **+35%** | **+26%** | **+30%** | **10.6 ms** | **3,044 MB** | 13 to 29 MB | **90 ms** | **182** |
| 256 MiB cache | +33% | +23% | +30% | 12.2 ms | 3,052 MB | 72 MB | 95 ms | 834 |
| 500 MiB cache | +34% | +23% | +29% | 12.2 ms | 3,058 MB | 72 MB | 94 ms | 834 |
| old `DefaultPebbleOptions()` | +36% | +28% | +28% | 9.8 ms | 3,416 MB | 54 MB | 89 ms | 44 |

- The cache gain is all in the first step: 64, 256 and 500 MiB hit at the same
  rate in every window (93.5%, 77.3%, 94.8%). Block processing never held more
  than 72 MB of blocks; its misses are first reads of tables that flushes and
  compactions just wrote.
- The 16 MiB memtable is what lifts the hit rate further, to 97.2%, 92.9% and
  97.9% next to the 64 MiB cache: it cuts W2's flushes from 834 to 182 and its
  compaction volume from 5.7 GB to 1.3 GB, so fewer new tables are read cold.
  It also cuts CPU per block by 8%, 13% and 4% over the 64 MiB cache alone.
- The old `DefaultPebbleOptions()` (500 MB cache, 64 MiB memtable, bloom, 3
  compactions) is within noise of 64 MiB + 16 MiB on throughput (+0.4, +1.4,
  -2.2 points) and costs 290 to 520 MB more RSS per window, mostly from its
  64 MiB memtables.
- Bloom filters, 10 bits per key, measured on fully compacted copies of the
  snapshots so every table carries filters: Get time fell 7 to 10% next to the
  64 MiB cache, blocks/s moved -0.2%, +0.8% and +1.2%. The filters take 4.9 to
  6.3 MiB per store.
- The app store's index blocks: 10.4 MiB at height 250k, 12.7 MiB at 1.03M,
  13.2 MiB at 1.58M, 13.5 MiB at 1.65M (629 MiB of data, 8.9M entries).

## Measurement: synthetic workloads

Earlier measurements on in-process apps driven over ABCI, before the replay:

- Block cache, 1M avl keys, 100 tx per block, every arm with 16 MiB memtables:
  64 MiB raised iteration-heavy tx/s from 20 to 30-32 and cut Get time per tx
  by 60% for read-modify-write; 256 MiB and 1 GiB did no better.
- Memtable, 100 read-modify-write tx per block: 16 MiB cut flushes from 100 to
  20 per 100 blocks and Commit p99 from 85-318 ms to 55-83 ms at 1M keys;
  64 MiB did no better than 16 MiB. Mainnet blocks are lighter, and the replay's p99
  apply moved from 113 ms to 106 ms with the memtable alone.

App hash and gas used were identical in every run of both.

## Decision

`DefaultPebbleOptions()` returns a 64 MiB block cache and 16 MiB memtables,
nothing else, and `NewPebbleDB` opens with it, dropping its own cache reference
once `pebble.Open` holds one. Benchmarks that call `DefaultPebbleOptions()` now
measure what nodes run. `NewPebbleDBWithOpts` callers keep full control.

## Memory

pebble counts memtables against the cache, so a database holds at most 64 MiB
of blocks and memtables together, against 8 MiB before: at most 168 MiB more
for a node's three databases. Release binaries are built with
`CGO_ENABLED=0`, where pebble allocates both on the Go heap, so with
`GOGC=100` the heap target can grow by up to twice that. The cache fills only
as blocks are read: the replay process's peak RSS moved by +28 to +140 MB over
stock.

## Alternatives considered

- **Larger caches (256 MiB, 500 MiB, 1 GiB)**: no measured gain, in the replay
  or the synthetic runs. Nodes serving heavy query traffic, not measured here,
  may want more; a configuration option is the follow-up for them.
- **The old `DefaultPebbleOptions()`**: no throughput gain over this change,
  290 to 520 MB more RSS, and a 1 GB heap target per database at its ceiling.
- **64 MiB memtable**: four times the memory of 16 MiB for no measured gain.
- **Bloom filters**: at most 1% blocks/s, and only for tables written after the
  change.
- **An 88 MiB cache, to keep 56 MiB for blocks next to 16 MiB memtables**: the
  32 MiB left is more than the replay ever held (29 MB) and more than twice the
  app store's index.
- **3 concurrent compactions**: not measured on its own; with 16 MiB memtables
  compaction volume is a quarter of what it was.

## Consequences

- Block processing with per-block commits runs 26 to 35% faster on mainnet's
  history.
- 32 MiB of the cache is left for blocks. The app store's index grew 3.1 MiB
  between heights 250k and 1.65M, about eight weeks of blocks, so at that rate
  it reaches 32 MiB in about a year; a configuration option is the natural
  follow-up for larger stores.
- No change to state, app hash or gas. The on-disk format is unchanged, so
  older binaries open databases written by this one.
- Not measured: block sync's deferred writes on master (#6334), whose batches
  are larger than a block; fsync cost; the blockstore database; query load.
