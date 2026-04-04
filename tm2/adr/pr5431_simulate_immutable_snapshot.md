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
