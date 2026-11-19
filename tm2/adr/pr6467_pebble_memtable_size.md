# PRxxxx: 16 MiB pebble memtable

## Status

Accepted; implemented in this PR.

## Context

Every pebble database a node opens goes through the registered creator,
`NewPebbleDB`, which passed an empty `pebble.Options`: the app store
(`gno.land/pkg/gnoland/app.go`, `dbm.NewDB("gnolang", dbm.PebbleDBBackend, …)`)
and, with the default `db_backend = "pebbledb"`, the `blockstore` and `state`
databases (`tm2/pkg/bft/node/node.go`, `DefaultDBProvider`). pebble then
fills in its defaults, including a 4 MiB memtable.

`DefaultPebbleOptions()` (500 MB cache, 64 MiB memtable, bloom filter) exists
in the same file, but only benchmarks and tests call it
(`tm2/pkg/bptree/benchmarks`, `gnovm/cmd/benchstore`).

Profiling signed `MsgCall` transactions through an in-process app over ABCI
on an on-disk pebble store (100 transactions per block, each one a
read-modify-write of a random key in an avl tree of 100k or 1M keys) showed
that one block of realm writes is larger than 4 MiB. Every block therefore
rotated the memtable, flushed it to L0 and triggered L0 compactions: 100
flushes and 163 compactions in 100 blocks at 1M keys. The batch apply in
`Commit` was 10 to 13 ms per block at the median and reached 370 to 440 ms
when it queued behind a flush. The WAL fsync itself was under 1 ms.

## Decision

`NewPebbleDB` sets `MemTableSize` to 16 MiB. Nothing else changes.

## Measurements

Same binary, same pre-populated store copied fresh for each arm, arms run
back to back on the same 16 cores. App hash and gas used are identical
across every arm.

100k keys, 150 blocks of read-modify-write per arm (n=1 run per arm):

| memtable | Commit per block p50 / p90 / p99 / max ms | stddev | flushes / compactions |
|---|--:|--:|--:|
| 4 MiB (before) | 23.3 / 42.1 / 121.5 / 382.3 | 32.4 | 150 / 296 |
| 16 MiB | 20.2 / 31.0 / 49.6 / 96.0 | 9.0 | 21 / 50 |
| 64 MiB | 21.8 / 31.9 / 62.5 / 95.0 | 9.6 | 5 / 8 |

1M keys, 100 blocks per arm, two pairs with the order swapped:

| pair | memtable | tx/s | DeliverTx p50 / p90 / p99 ms | Commit p50 / p99 / max ms | flushes / compactions |
|---|---|--:|--:|--:|--:|
| 1 | 4 MiB | 66 | 13.58 / 18.58 / 29.63 | 29.6 / 317.6 / 444.4 | 100 / 163 |
| 1 | 16 MiB | 73 | 12.71 / 16.09 / 25.37 | 24.4 / 55.2 / 59.1 | 20 / 54 |
| 2 | 16 MiB | 71 | 12.93 / 16.68 / 25.55 | 25.3 / 83.3 / 165.9 | 20 / 54 |
| 2 | 4 MiB | 60 | 14.91 / 22.30 / 29.65 | 30.6 / 85.4 / 178.5 | 100 / 164 |

## Alternatives considered

- **64 MiB** (the value in `DefaultPebbleOptions`): no better than 16 MiB on
  these runs, four times the memory bound.
- **`MemTableStopWritesThreshold` 4, L0 thresholds 8/24, a 256 MiB block
  cache, each on top of 64 MiB**: within run-to-run noise of 64 MiB alone on a
  3 x 3 sweep at 100k keys.
- **Use `DefaultPebbleOptions()` in production**: 500 MB of block cache per
  database, three databases per node, for a gain these runs did not show.

## Consequences

- Memory: pebble starts each memtable at 256 KiB and doubles it only as it
  fills, so a database that writes little never reaches 16 MiB. pebble stalls
  writes once queued memtables reach `MemTableStopWritesThreshold` (2) x
  `MemTableSize`, so the bound per database goes from 8 MiB to 32 MiB, up to
  72 MiB more for the three databases of a node. Peak RSS of the benchmark
  process did not change measurably at 1M keys (518 and 520 MiB before,
  513 and 519 MiB after).
- On a short run where the memtable never rotates, the WAL fsync goes from
  0.25 ms to 3 to 5 ms per block and the batch apply halves, so the two
  cancel out. Over 100 blocks and more, rotation and WAL recycling resume and
  the net effect is the one in the tables above. Why the fsync is slower
  before the first rotation is an inference, not measured: each fsync then
  lands on a freshly preallocated WAL file instead of a recycled one.
- No change to state, app hash, or gas.
