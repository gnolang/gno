# PR5431: Make Simulate Use Immutable Committed Snapshot

## Context

The `.app/simulate` RPC endpoint executes full transactions via the ABCI query
path: `ABCIQuery` -> `QuerySync` -> `BaseApp.Query` -> `handleQueryApp` ->
`Simulate` -> `runTx`. Simulate wraps the result in a `CacheContext` that is
never committed, so every call is effectively free.

Tendermint's ABCI layer creates three connections (consensus, mempool, query),
each backed by a `localClient` from `localClientCreator`. To prevent query
calls from blocking consensus, the query connection was given its own mutex
(`queryMtx`), separate from the consensus/mempool mutex (`mtx`).

However, `Simulate()` read from `app.checkState`, a live pointer that is
replaced during `Commit()` via `setCheckState()`. With the query connection
running on a separate mutex, there was no synchronization between `Commit`
replacing `checkState` and `Simulate` reading it — a data race.

Additionally, the combination of free execution (fees on a `CacheContext` that
never commits) and shared mutex contention enabled a zero-cost DoS: an attacker
could flood simulate requests to block consensus from advancing.

## Decision

Make `Simulate()` read from an **immutable committed snapshot** via
`MultiImmutableCacheWrapWithVersion()`, the same pattern already used by
`handleQueryCustom` and `handleQueryStore` for their query paths.

### Changes

1. **Atomic header storage**: Add an `atomic.Value` field (`lastBlockHeader`)
   to `BaseApp`, updated in `setCheckState()` (called from `Commit` and
   `InitChain`). This provides a thread-safe way to read the last committed
   block header without holding any mutex.

2. **Rewrite `Simulate()`**: Instead of reading from `app.checkState` (via
   `getContextForTx`), load an immutable IAVL snapshot at the last committed
   height. IAVL versions are copy-on-write, so the snapshot is safe for
   concurrent reads with no synchronization. Fall back to the original
   `getContextForTx` path when `header.GetHeight() < 1` (before first commit,
   e.g. during `InitChain` or tests) where single-threaded context is
   guaranteed.

3. **Fix `handleQueryCustom`**: Replace `app.checkState.ctx.BlockHeader()`
   (same latent race) with the atomic `getLastBlockHeader()` accessor.

### Why the header needs atomic access

The `lastBlockHeader` is stored via `atomic.Value` using a `headerSnapshot`
wrapper struct (required because `atomic.Value` needs a consistent concrete
type). It is only updated in `setCheckState()`, which runs during `Commit` and
`InitChain` — both under the consensus mutex. The query path reads it without
any mutex, so the atomic guarantees visibility.

The header height from `setCheckState` matches the committed store version,
ensuring `MultiImmutableCacheWrapWithVersion(height)` always finds the
requested version.

## Alternatives Considered

### RWMutex on localClient

Change `*sync.Mutex` to `*sync.RWMutex` in `localClient`, use `RLock` for
`QuerySync`. This allows concurrent queries but does not fix the underlying
data race on `app.checkState` replacement during `Commit`. The race exists at
the application layer, not the transport layer.

### Path-sniffing in QuerySync

Skip the mutex for `.app/simulate` paths directly in
`localClient.QuerySync()`. This leaks application-level semantics into the
transport layer and still races on `checkState` reads.

### Snapshot + release pattern

Lock the mutex briefly to snapshot the state needed for simulation, then
release it and run the simulation lock-free. This requires `BaseApp` to expose
snapshot methods and is a more invasive refactor for the same result.

## Consequences

- **Simulate reads committed state**: Simulate no longer sees pending mempool
  changes (e.g. sequence number updates from `CheckTx`). This is acceptable for
  gas estimation and matches Cosmos SDK behavior.

- **Slightly slower snapshot load**: `MultiImmutableCacheWrapWithVersion` loads
  from the database rather than reading `checkState`'s in-memory cache. This is
  acceptable for query-path operations and is the same cost as all other query
  types.

- **No consensus blocking**: With the query connection on a separate mutex and
  `Simulate` no longer accessing shared mutable state, simulate requests cannot
  block `BeginBlock`, `DeliverTx`, `EndBlock`, or `Commit`.

- **Race-free**: Verified with `go test -race` including a dedicated concurrent
  simulate+commit test (`TestSimulateConcurrentWithCommit`).

---

## Follow-on: GnoVM `cacheNodes` data race

### Context

PR5431 correctly isolated the IAVL/KV store layer via `MultiImmutableCacheWrapWithVersion`. However, stress-testing the separate-queryMtx configuration (8 simulate goroutines + 100 broadcast txs in waves, `go test -race -count=10`) revealed a second race at the GnoVM store layer.

`VMKeeper` holds a single `vm.gnoStore` (`*defaultStore`) that lives for the node lifetime. Its `cacheNodes` field (`txlog.Map[Location, BlockNode]`) is initialized as a plain `GoMap[Location, BlockNode]` — a bare Go map with no synchronization.

When `BeginTransaction()` is called it wraps the root map with `txlog.Wrap(ds.cacheNodes)`, producing a per-transaction `txLog{source: rootGoMap, dirty: {}}`. The `dirty` map is private to each transaction. But `source` is the **shared root GoMap**, accessed concurrently by:

- Goroutine A (Simulate, `queryMtx`): on a `cacheNodes` cache miss, `txLog.Get()` falls through to `source.Get(k)` — a direct read on the root GoMap (`txlog.go`, `txLog.Get`).
- Goroutine B (DeliverTx commit, `mtx`): `transactionStore.Write()` → `txLog.Commit()` → `source.Set(k, v)` for every dirty node — a direct write on the root GoMap (`txlog.go`, `txLog.Commit`).

A concurrent map read and map write is a fatal Go data race. `cacheObjects` and `cacheTypes` are not affected — transaction stores allocate fresh empty maps for those and never fall through to the root store's copies. Only `cacheNodes` uses the txlog wrapper pattern with a shared source.

### Decision

Add `SyncGoMap[K, V]` to `gnovm/pkg/gnolang/internal/txlog` — a struct wrapping a plain map with `sync.RWMutex` that implements the same `Map[K, V]` interface. Initialize the root store's `cacheNodes` with `NewSyncGoMap[Location, BlockNode]()` instead of `GoMap`.

`txlog.Wrap()` accepts any `Map[K, V]`, so `BeginTransaction()` continues to work unchanged: each transaction gets a `txLog{source: *SyncGoMap}`, and all reads/writes flowing through the wrapper's source pointer go through the mutex.

`SyncGoMap.Iterate()` takes a snapshot under `RLock` and releases the lock before yielding. This is required because `txLog.Iterate()` calls `b.source.Get(k)` while iterating `b.source.Iterate()` — holding `RLock` for the full duration of the lazy iterator would cause a deadlock under write pressure on the same goroutine.

### Alternatives considered

**VMKeeper-level `RWMutex`**: Guard `CommitGnoTransactionStore` with a write lock and the simulate path with a read lock. This re-serializes simulate and DeliverTx at the keeper boundary, defeating the purpose of the separate `queryMtx`.

**Snapshot on `BeginTransaction`**: Deep-copy `cacheNodes` into a fresh map for each transaction instead of wrapping it. Correct but expensive — the cache may contain thousands of BlockNodes per block.

### Consequences

- The root `cacheNodes` map acquires a read lock on every `txLog.Get()` source fallthrough and a write lock on every `txLog.Commit()` flush. BlockNode lookups are read-heavy (one write per new node, many reads per execution), so `RWMutex` contention is low in practice.
- Verified with `go test -race -run TestSimulateBurstDuringCommit -count=10` (`gno.land/pkg/gnoclient`), which runs 8 simulate goroutines concurrently with 100 broadcast txs and previously reproduced the race in 2/10 runs.
