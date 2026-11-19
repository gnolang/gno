# PR0000: Give NewPebbleDB a 64 MiB block cache

## Status

Proposed

## Context

`NewPebbleDB` (`tm2/pkg/db/pebbledb/pebbledb.go`) is the registered creator for
the pebbledb backend, so it opens the app store (`gno.land/pkg/gnoland/app.go`)
and, with the default `db_backend`, the `blockstore` and `state` databases. It
passed an empty `pebble.Options`, which gives pebble's 8 MiB block cache.
`DefaultPebbleOptions()` in the same file (500 MB cache, bloom filter) is only
used by tests and benchmarks.

A pebble point lookup reads, per level it visits, an index block and then a data
block, both through the block cache. On a store holding a realm with 1,000,000
avl keys (4.0M pebble entries, 583 MiB of data in 118 sstables) the index blocks
alone are **39.4 MiB**, measured from the sstable properties. An 8 MiB cache
cannot hold the index, so random gets evict each other's index blocks before
they are reused.

## Measurement

In-process gno.land app over ABCI, signed `MsgCall` transactions against a realm
holding an avl tree of 1,000,000 keys, 100 tx per block, 100 warmup tx per
shape. Every run starts from a fresh copy of the same pre-populated store, warm
page cache, `taskset` 12 cores, GOMAXPROCS=12, 3 reps per arm with the arm order
rotated. Every arm also sets the 16 MiB memtable of #6467, so only the cache
differs. `read` reads one random key, `rmw` reads and writes one, `walk`
iterates 100 keys from a random start. n = 2000 tx per rep (1000 for `walk`).
Get time is measured around every backend `Get`/`Has` during `DeliverTx`.

| cache | shape | tx/s per rep | DeliverTx p50 ms per rep | Get ms per tx p50 | block cache hit | peak RSS |
|---|---|---|---|---|--:|--:|
| 8 MiB (default) | walk | 20, 20, 19 | 46.6, 46.3, 49.0 | 22.9, 22.9, 24.2 | 0.0% | 585 MiB |
| **64 MiB** | walk | **32, 30, 30** | **28.6, 30.9, 31.3** | **5.5, 6.0, 6.0** | 96.6% | 712 to 738 MiB |
| 256 MiB | walk | 32, 29, 28 | 29.2, 30.9, 31.5 | 5.7, 6.0, 6.1 | 97.3% | 946 to 1009 MiB |
| 1 GiB | walk | 30, 30, 30 | 29.7, 30.7, 30.8 | 5.5, 5.7, 5.8 | 97.6% | 1115 to 1118 MiB |
| 8 MiB (default) | rmw | 72, 75, 74 | 12.3, 12.2, 12.1 | 2.18, 2.13, 2.16 | 4.3% | |
| **64 MiB** | rmw | 79, 60, 74 | 11.2, 15.4, 12.2 | **0.85, 1.03, 0.90** | 73.7% | |
| 256 MiB | rmw | 81, 75, 75 | 10.7, 11.6, 11.4 | 0.76, 0.81, 0.81 | 76.0% | |
| 8 MiB (default) | read | 114, 121, 115 | 7.2, 7.5, 7.6 | 0.55, 0.57, 0.59 | 92.1% | |
| **64 MiB** | read | 121, 97, 89 | 7.2, 9.0, 9.1 | 0.45, 0.55, 0.54 | 95.7% | |

- Iteration-heavy access gains about 50% throughput: Get time falls from 48% to
  20% of `DeliverTx` wall. Point access (`rmw`) spends 60% less time in Get, but
  that was 17% of wall, so its throughput moves inside the rep-to-rep noise of
  this shared host. Single-key `read` is unchanged within noise.
- The gain is all in the first step: 256 MiB and 1 GiB do no better than
  64 MiB, consistent with the 39.4 MiB index fitting at 64 MiB.
- App hash and gas used are identical in every run. The cache only changes
  where block bytes come from.

## Decision

`NewPebbleDB` creates a 64 MiB cache, passes it in `Options.Cache`, and drops
its own reference once `pebble.Open` holds one. Nothing else changes.
`NewPebbleDBWithOpts` callers keep full control.

## Memory

The cache allocates as blocks are read, so a database that reads little stays
small. The bound is 64 MiB per database instead of 8 MiB: at most 168 MiB more
for a node's three pebble databases, in practice less because `blockstore` and
`state` read a small working set. Measured peak RSS of the benchmark process
(one database): 585 MiB at 8 MiB, 706 to 752 MiB at 64 MiB across 6 runs.

## Alternatives considered

- **256 MiB or larger**: no measured gain at 1M keys, 3 to 8x the memory.
- **Bloom filter, 10 bits per key**, on top of 64 MiB: measured on a pair of
  fully rewritten stores so every table carries filters. Get time per tx fell
  a further 5 to 10% (`walk` 4.6 to 5.1 ms down to 4.2 to 4.3 ms) and tx/s
  stayed inside the noise. It also only applies to tables written after the
  change, so an existing node would see it gradually. Not worth a second
  change today.
- **`DefaultPebbleOptions()`** (500 MB cache, 64 MiB memtable, bloom): a
  bigger memory commitment for no measured read gain over 64 MiB.
- **`MaxOpenFiles`**: the table cache hit rate was 100% in every run (118
  tables, default limit 1000). Nothing to tune at this size.
- **Avoiding the gets instead**: within a block no key was fetched twice (the
  cache store above deduplicates), and no transaction issued a backend
  iterator, so there are no repeated round trips to batch.

## Consequences

- Stores of a few million keys or less read their index from memory instead of
  re-decoding it per Get.
- The index grows with the store (about 40 MiB per million avl keys in this
  layout), so a store several times larger will outgrow 64 MiB again. Making
  the size configurable is the natural follow-up; it is left out to keep this
  change to one constant.
