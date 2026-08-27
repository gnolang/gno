# PR5431: Concurrent ABCI Queries with Snapshot Isolation and Atomic Commits

## Context

Tendermint's ABCI layer creates three connections (consensus, mempool, query),
each backed by a `localClient`. Before this work every ABCI call took the same
mutex, so any `.app/simulate` request blocked consensus (`BeginBlock`,
`DeliverTx`, `EndBlock`, `Commit`) for its duration. Because `Simulate` runs a
full transaction against a `CacheContext` that is never committed, fees are
never charged, which turns unlimited simulate flooding into a zero-cost DoS
against block production.

The obvious fix — a separate `queryMtx` for the query connection — is
insufficient on its own. Once queries and consensus can run concurrently, four
latent problems become observable:

1. **Data race on `checkState`.** `Simulate` reads `app.checkState`, a live
   pointer that `Commit()` replaces via `setCheckState()`. Concurrent reads
   and writes to that pointer are a Go data race.

2. **Cross-store inconsistency.** The multistore has an IAVL sub-store
   (versioned, copy-on-write) and a `dbadapter` sub-store (unversioned,
   direct writes). A query that pins IAVL to version `N` can read `dbadapter`
   data already overwritten by block `N+1`, producing a state that mixes two
   block heights.

3. **Torn commits.** A block commit produces four separate `WriteSync()`
   calls (block header, cache flush, IAVL SaveVersion, rootmulti metadata).
   A snapshot taken between two of them observes a partially-written block.

4. **GnoVM `cacheNodes` race.** `VMKeeper.gnoStore.cacheNodes` is a plain
   `map` shared between `Simulate` (via `txLog.Get` fallthrough) and
   `DeliverTx` (via `txLog.Commit` writes) — a fatal Go data race that only
   materialises once the two paths actually run concurrently.

This ADR describes the four mechanisms landing together in PR5431 to solve
these correctly.

## Decision

### 1. Separate `queryMtx` on `localClient`

The query connection uses its own mutex, distinct from the consensus/mempool
mutex. Queries and simulate no longer serialise with block production.

### 2. `Simulate` reads a committed snapshot (not `checkState`)

`BaseApp` gains an `atomic.Pointer[headerSnapshot]` field, `lastBlockHeader`, updated in
`setCheckState()` (called from `Commit` and `InitChain`) — the only writers,
which run under the consensus mutex. Query paths read it lock-free.

`Simulate()` is rewritten to load an immutable multistore at the last
committed height rather than reading `checkState`:

```go
header := app.getLastBlockHeader()          // atomic read
cacheMS, release, err := app.cms.MultiImmutableCacheWrapWithVersion(header.Height)
defer release()
ctx := NewContext(RunTxModeSimulate, cacheMS, header, app.logger)...
return app.runTx(ctx, txBytes)
```

`handleQueryCustom` gets the same treatment: previously it read
`app.checkState.ctx.BlockHeader()`, which raced with `setCheckState`. It now
reads `getLastBlockHeader()` and calls `MultiImmutableCacheWrapWithVersion`.

**Semantic difference vs Cosmos SDK.** SDK Simulate reads `checkState` and so
sees pending mempool changes (e.g. sequence numbers updated by `CheckTx`);
our Simulate sees only committed state. CometBFT briefly shipped the same
committed-snapshot behaviour ([cometbft@869833e][cometbft-snap]) before
reverting it for compatibility, not correctness. We accept the divergence.

[cometbft-snap]: https://github.com/cometbft/cometbft/commit/869833e

### 3. DB-level snapshot for cross-store read isolation

IAVL versioning covers only the IAVL sub-store. To give queries a consistent
view across *every* sub-store, we take a **PebbleDB snapshot** — a
point-in-time MVCC view of the entire shared DB — after each block commit,
and route immutable-store reads through it.

A new interface:

```go
type Snapshot interface {
    Get([]byte) ([]byte, error)
    Has(key []byte) (bool, error)
    Iterator(start, end []byte) (Iterator, error)
    ReverseIterator(start, end []byte) (Iterator, error)
    Close() error
}
```

Implementations:

| Backend    | `NewSnapshot()`                                                    |
|------------|--------------------------------------------------------------------|
| PebbleDB   | `pebble.DB.NewSnapshot()` — real MVCC snapshot                     |
| MemDB      | copies the map under lock — tests only                             |
| Others     | returns `"snapshots not supported"` — falls back to `ImmutableDB`  |

`SnapshotDB` wraps a `Snapshot` in the full `db.DB` interface (writes panic;
`NewBatch` returns a no-op batch — IAVL constructs one eagerly in
`newNodeDB` even for read-only loads).

### 4. Snapshot lifetime — `refSnapshot` + `snapshotMu` TOCTOU guard

`rootmulti.multiStore` holds an `atomic.Pointer[refSnapshot]`. Each snapshot
starts with `refs = 1` (the store's own reference). `Commit()` takes a fresh
snapshot after all writes complete, atomically swaps it in, and calls
`release()` on the old one — driving its refcount toward zero.

Queries `acquire()` a reference at the start of
`MultiImmutableCacheWrapWithVersion` and `release()` it in the deferred
callback. The snapshot stays alive as long as any query holds it.

**TOCTOU protection.** `atomic.Pointer.Load` is atomic in isolation, but
`Load(); acquire()` is not atomic relative to `Swap(); release()`. Without a
guard, this interleaving is legal:

```
query goroutine          Commit goroutine
──────────────────────   ────────────────────────────────
rs := Load()  (refs=1)
                         Swap(newRS)
                         old.release()  →  refs 1→0  →  snap.Close()
rs.acquire()             (snapshot already closed!)
rs.snap.Get(...)         ← use-after-free
```

A `sync.RWMutex` (`snapshotMu`) closes the window: query paths hold an
`RLock` around `Load+acquire`, `Commit()` holds the write lock around
`Swap+release`. The write-lock excludes all RLocks, so `release()` cannot
reach zero while a `Load+acquire` is in progress.

**Shutdown.** The store's initial `refs=1` is normally released only when the
next `Commit()` swaps the snapshot out. On node shutdown that swap never
happens, so PebbleDB reports a leaked snapshot at `db.Close()`. `multiStore`
implements `io.Closer`; `BaseApp.Close()` calls it before `app.db.Close()`,
draining the outstanding reference.

### 5. Atomic block commits — `BatchCollector` + `CollectingDB`

A block previously produced four separate durable writes (block header,
dbadapter cache flush, IAVL `SaveVersion` with mid-flush auto-flushes from
`BatchWithFlusher`, rootmulti metadata). A snapshot taken between them would
see a torn state, and even without snapshots a crash between them left the
DB inconsistent.

We fuse them into one `WriteSync()` by installing a wrapper at the `db.DB`
layer:

- **`BatchCollector`** — an in-memory op-log (Set/Delete queue) with an
  auxiliary `pending map[string]int` indexing the latest op per key.
- **`CollectingDB`** — wraps the real DB. `Set`/`Delete` and every batch it
  hands out route into the collector; `Write`/`WriteSync`/`Close` on those
  batches are no-ops. `Get`/`Has` consult the collector first (read-your-
  writes) then fall through to the real DB.

Every sub-store is mounted with a `CollectingDB` at `constructStore` time
(immutable/query multistores skip this — they never write). At the end of
`rootmulti.Commit()`:

```go
metaBatch := ms.collector.NewBatch()   // writes into the same collector
setCommitInfo(metaBatch, version, commitInfo)
setLatestVersion(metaBatch, version)

realBatch := ms.db.NewBatch()          // real batch on the raw DB
defer realBatch.Close()
ms.collector.Drain(realBatch)          // replay every op in order
realBatch.WriteSync()                  // ONE atomic disk flush
```

**Ordering contract.** `BaseApp.Commit()` calls
`app.deliverState.ms.MultiWrite()` immediately before `app.cms.Commit()`.
MultiWrite must precede Commit — reversing it makes IAVL SaveVersion run
against a stale in-memory tree (wrong app hash) and shifts dbadapter writes
into the next block's batch. The DB layer gives atomicity; ordering is the
caller's contract.

**Why not modify IAVL.** An earlier design considered changing IAVL's
`BatchWithFlusher` or `nodeDB` to accept an external batch. It was rejected
because our IAVL package is a vendored fork of `cosmos/iavl`, and every
modification is a permanent local divergence. The DB-layer approach handles
all four write sites uniformly (dbadapter, IAVL SaveVersion, IAVL pruning,
rootmulti metadata) without touching IAVL. `BatchWithFlusher` continues to
believe it is managing memory correctly — from its perspective it is; the
"flush" just becomes a no-op that resets its local batch pointer.

**Read-your-writes.** The initial design routed writes into the collector
without exposing them to reads through the same `CollectingDB`. This broke
any flow that writes and then reads back before the next
`rootmulti.Commit()` drain — most visibly `TestAppHashCrossrealm38`, which
deploys a package in Tx1 and imports it in Tx2 against the raw multistore.
The fix is the `pending` map: `Get`/`Has` first consult the collector
(pending Set → return value; pending Delete → return not-found), then fall
through. `Iterator`/`ReverseIterator` intentionally do not merge — no
current consumer iterates during a commit window, and merging would
significantly complicate the code.

### 6. GnoVM `cacheNodes` race — `SyncGoMap`

Stress-testing the separate-`queryMtx` configuration (8 simulate goroutines
+ 100 broadcast txs in waves, `go test -race -count=10`) revealed a race in
`VMKeeper.gnoStore.cacheNodes`. `BeginTransaction()` wraps that shared root
map with `txlog.Wrap`, producing a per-tx `txLog{source: rootGoMap, dirty:
{}}`. On a cache miss `txLog.Get` reads from `source` directly, and
`txLog.Commit` writes to `source` directly — a concurrent map read and
write.

`cacheObjects` and `cacheTypes` are not affected: transaction stores
allocate fresh maps for those and never fall through to the root store's
copies. Only `cacheNodes` uses the txlog wrapper with a shared source.

Add `SyncGoMap[K, V]` in `gnovm/pkg/gnolang/internal/txlog` — a struct
wrapping a plain map with `sync.RWMutex` that implements the same
`Map[K, V]` interface. Initialize `cacheNodes` with `NewSyncGoMap` instead
of `GoMap`. `txlog.Wrap` accepts any `Map[K, V]`, so nothing else changes.
`SyncGoMap.Iterate` takes a snapshot under `RLock` and releases the lock
before yielding — required because `txLog.Iterate` calls `source.Get(k)`
while iterating `source.Iterate()`, and holding the RLock through the lazy
iterator would deadlock under write pressure.

## Alternatives Considered

**RWMutex on `localClient` alone.** Convert the shared mutex to an RWMutex
and take an RLock for queries. Fixes contention but does not fix the
`checkState` data race or cross-store inconsistency — those live at the
application layer, not the transport layer.

**Path-sniffing in `QuerySync`.** Skip the mutex specifically for
`.app/simulate` paths. Leaks application semantics into the transport layer
and still races on `checkState`.

**Snapshot-and-release inside the mutex.** Lock briefly to snapshot state,
unlock, run simulate lock-free. Requires BaseApp to expose snapshot methods
and re-serialises the snapshot step with Commit — more invasive for the
same result.

**Per-sub-store PebbleDB snapshots.** Take one snapshot per sub-store and
combine them. Equivalent to the current approach when all sub-stores share
a DB (they do — they use key prefixes on one PebbleDB), and strictly more
complex if that ever changes.

**Modify IAVL to accept an external batch.** Fork the vendored IAVL library
so `SaveVersion`/`nodeDB` write into a caller-provided batch. Would only fix
one of the four write sites (dbadapter and rootmulti metadata still write
separately), and every modification to the fork is a permanent maintenance
tax against `cosmos/iavl`. The DB-layer approach is uniform and requires no
upstream divergence.

**VMKeeper-level RWMutex around `cacheNodes`.** Guard
`CommitGnoTransactionStore` with a write lock and simulate with a read
lock. Re-serialises simulate and DeliverTx at the keeper boundary,
defeating the point of the separate `queryMtx`.

**Deep-copy `cacheNodes` per transaction.** Snapshot the whole map into a
fresh map on every `BeginTransaction`. Correct but expensive — the cache
holds thousands of `BlockNode` entries per block.

## Consequences

- **Consensus is never blocked by queries.** Simulate and query paths hold a
  snapshot and never touch the live DB or shared mutable state.
- **Cross-store isolation is correct.** A query sees the IAVL sub-store,
  the dbadapter sub-store, and rootmulti metadata all at the same committed
  block height. Torn views across sub-stores are structurally impossible.
- **Block commits are atomic at the DB level.** Every write for a block
  lands in one `WriteSync()`; either the whole block persists or none does.
  A crash mid-commit leaves the DB at the previous version and Tendermint
  replays block N on restart.
- **Snapshot resource usage.** Each open snapshot prevents Pebble from
  compacting the key versions it covers. In practice snapshots are held for
  a single query (milliseconds) and are reference-counted, so compaction
  pressure is negligible.
- **Memory during commit.** The `BatchCollector` holds one block's writes
  in RAM before draining. Bounded by block gas limit; typical block ~2 MB.
- **memdb in tests.** `MemDB.NewSnapshot()` copies the map at snapshot
  time. O(n) in keys, tests-only.
- **Non-PebbleDB backends.** `GoLevelDB`, `BoltDB`, `LMDB`, `MDBX` return an
  error from `NewSnapshot()`; `MultiImmutableCacheWrapWithVersion` falls
  back to `ImmutableDB` over the live DB. These backends are not used in
  production.
- **Semantic change to Simulate.** Sees committed state only, not pending
  mempool state — a deliberate divergence from Cosmos SDK, aligned with
  CometBFT's briefly-shipped design.
- **Race-free.** Verified with `go test -race`, including dedicated tests:
  `TestSimulateConcurrentWithCommit` (SDK layer),
  `TestSnapshotConcurrentCommitAndQuery` (rootmulti layer),
  `TestSimulateBurstDuringCommit` (gnoclient layer, exercises GnoVM
  `cacheNodes`), and `TestCommitAtomicBatchWithCacheFlush` (verifies exactly
  one `WriteSync` per block).
