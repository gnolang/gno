# PR5431 Follow-on: DB-Level Snapshot Read Isolation

## Context

PR5431 made `Simulate()` read from an immutable IAVL snapshot via
`MultiImmutableCacheWrapWithVersion()`. IAVL is copy-on-write and versioned, so
loading a past version is safe for concurrent reads. This fixed the data race on
`app.checkState` and removed consensus blocking.

However, a review by @thehowl identified a deeper correctness problem: **IAVL
snapshots only cover the IAVL sub-store**. The gno.land multistore also contains
a `dbadapter` sub-store (mounted under `baseKey`) for non-merkle state: the last
committed block header (`mainLastHeaderKey`) and other direct key-value data.

`dbadapter` has no versioning. It writes directly to the underlying database
with no copy-on-write semantics. `MultiImmutableCacheWrapWithVersion(N)` wraps
the IAVL store at version N in an immutable tree, but wraps the `dbadapter`
store with only an `ImmutableDB` (a panic-on-write wrapper over the **live**
database). A query pinned to IAVL version N could therefore read `dbadapter`
data that was already overwritten by block N+1 or N+2, producing a
**cross-store inconsistent view** where the two sub-stores appear to belong to
different block heights.

The current mutex-based serialization accidentally prevents this: the query
mutex shares a lock with `Commit`, so queries never overlap with the moment
writes land. The PR5431 approach of a separate `queryMtx` removes that
accidental protection without replacing it.

### The four write sites

A block commit currently produces writes to the shared PebbleDB instance across
four separate batch operations, in this order:

```
① MultiWrite()       cache.Store.Write() for each sub-store
                     dbadapter: db.NewBatch() + batch.Write()
                     IAVL:      tree.Set/Remove (in-memory only)

② SaveVersion()      iavl nodeDB: ndb.batch.WriteSync()
                     persists IAVL tree nodes for the new version

③ metadata batch     rootmulti.Commit(): batch.WriteSync()
                     writes commitInfo(N) and latestVersion=N

④ (pre-fix)          baseStore.Set(mainLastHeaderKey) directly
                     naked db.Set after Commit() returned — no batch
```

Between any two of these steps a concurrent reader sees an inconsistent DB:
e.g. between ② and ③ the IAVL nodes for version N+1 are on disk but the
metadata still claims `latestVersion=N`.

### Why a single atomic batch is not the fix

The natural solution would be to merge all four sites into one `batch.WriteSync()`
call. This is not practical without invasive changes to the IAVL library: IAVL's
`nodeDB` manages its own internal batch (`BatchWithFlusher`) and there is no API
to redirect its writes to an external batch. Threading an external batch through
`SaveVersion()` would require forking the library.

Instead, the mechanism described below achieves read isolation at the DB level
without requiring all writes to be atomic.

## Decision

### Part 1 — Order all writes before `Commit()` returns

Reorder `BaseApp.Commit()` so that every write for a block completes — and is
durable — before the function returns:

1. **Move the block header write into the deliver cache** before `MultiWrite()`.
   Previously `baseStore.Set(mainLastHeaderKey)` was called after `cms.Commit()`
   returned, as a raw direct write outside any batch. Moving it to
   `deliverState.ms.GetStore(baseKey).Set(...)` before `MultiWrite()` makes it
   flush with the rest of the block's `dbadapter` state, in the same
   `cache.Store` batch.

2. **Change the rootmulti metadata batch to `WriteSync()`**. The batch that
   writes `commitInfo` and `latestVersion` was using `batch.Write()` (async in
   some backends). Changing to `batch.WriteSync()` ensures this write is durable
   before `Commit()` returns.

After these changes the write sequence is strictly ordered and complete when
`Commit()` returns:

```
BaseApp.Commit()
  │
  ├─ deliverState.ms.GetStore(baseKey).Set(mainLastHeaderKey)  [into cache]
  │
  ├─ deliverState.ms.MultiWrite()
  │    └─ cache.Store.Write() per sub-store
  │         dbadapter → batch.Write()   ← header + all tx dbadapter state  ①
  │         IAVL      → tree.Set/Remove (in-memory)
  │
  └─ app.cms.Commit()  [rootmulti.Commit()]
       ├─ commitStores(version, stores)
       │    ├─ iavlStore.Commit()
       │    │    └─ tree.SaveVersion() → ndb.batch.WriteSync()              ②
       │    └─ dbadapterStore.Commit() → no-op
       └─ metadata batch.WriteSync()                                        ③
            writes commitInfo(N) and latestVersion=N

← Commit() returns here.  All writes for block N are on disk.
```

### Part 2 — DB-level snapshot for read isolation

Add a `Snapshot` interface to the `db.DB` abstraction:

```go
type Snapshot interface {
    Get([]byte) ([]byte, error)
    Has([]byte) (bool, error)
    Iterator(start, end []byte) (Iterator, error)
    ReverseIterator(start, end []byte) (Iterator, error)
    Close() error
}
```

PebbleDB implements this via `pebble.DB.NewSnapshot()`, which returns a
consistent point-in-time read view backed by Pebble's MVCC layer. A snapshot
opened after a set of batch commits sees exactly the state as of the last commit
— no partial writes, regardless of how many separate batches produced those
writes.

After `Commit()` returns, `rootmulti.Store` atomically replaces a stored
`*db.Snapshot` pointer with a new snapshot opened on the shared PebbleDB
instance:

```go
// inside rootmulti.Commit(), after all writes:
newSnap, _ := ms.db.NewSnapshot()
old := ms.querySnapshot.Swap(newSnap)
if old != nil {
    old.Close()
}
```

`querySnapshot` is an `atomic.Pointer[db.Snapshot]`, so the swap is
lock-free and safe for concurrent readers.

### Part 3 — Thread the snapshot through query/simulate paths

`MultiImmutableCacheWrapWithVersion(version)` is updated to acquire the current
snapshot and use it as the DB for loading the immutable store, instead of the
live database:

```go
func (ms *multiStore) MultiImmutableCacheWrapWithVersion(version int64) (types.MultiStore, func(), error) {
    snap := ms.querySnapshot.Load()
    snapDB := dbm.NewSnapshotDB(snap)         // read-only DB backed by the snapshot
    ims := &multiStore{db: snapDB, ...}
    ims.LoadVersion(version)
    release := func() { /* snapDB released, snapshot ref counted down */ }
    ...
    return cachemulti.New(stores, keysByName), release, nil
}
```

The `release` function is deferred by callers (`Simulate`, `handleQueryCustom`)
and closes the caller's reference to the snapshot once the query completes.
Pebble reference-counts snapshots internally; the underlying snapshot is not
freed until all callers have released it.

### Why this works

The correctness argument rests on two properties:

**Property 1 — writes complete before the snapshot is taken.**
Part 1 ensures `Commit()` only returns after all three batch writes (dbadapter,
IAVL nodes, metadata) are durable. The snapshot is taken immediately after.
There is no window in which the snapshot could be taken between two of the
batch writes.

**Property 2 — a PebbleDB snapshot is a consistent point-in-time view.**
Pebble implements MVCC: each key write is tagged with a sequence number.
`NewSnapshot()` records the current maximum sequence number. All reads through
the snapshot see only writes with a lower sequence number — i.e. all writes that
completed before the snapshot was opened, and none that came after. This is true
regardless of whether those writes arrived in one batch or three separate
batches, because by the time the snapshot is opened all three batches have
already committed.

**Combined**: a query that acquires a snapshot sees a state where:
- the dbadapter sub-store (including the block header) reflects block N
- the IAVL sub-store reflects block N
- the metadata (`commitInfo`, `latestVersion`) reflects block N

It is impossible to see a mix of N and N+1 data, because the snapshot was taken
after all of N's writes and before any of N+1's writes have started (block
commits are serialized by the consensus mutex).

### What is not atomic and why it does not matter

The three batch writes (①②③) are still separate `WriteSync()` calls. They are
not a single atomic operation at the DB level. A crash between ① and ③ would
leave the DB in a partially-written state. This is the same durability
guarantee as before this change — crash recovery is handled by the existing
`LoadLatestVersion` logic, which validates `commitInfo` against each sub-store's
`LastCommitID()` on startup. The isolation improvement applies only to
**concurrent readers during normal operation**, not to crash recovery.

## Alternatives Considered

### Single atomic batch across all write sites

Route all writes — IAVL tree nodes, dbadapter flushes, and metadata — through a
single `db.Batch` commited with one `WriteSync()`. This would give true
atomicity and make the snapshot unnecessary.

Not chosen for two concrete reasons:

1. **IAVL's `BatchWithFlusher` auto-flushes mid-`SaveVersion()`**. When a tree
   is large, `BatchWithFlusher.Set()` checks the batch size on every write and
   calls `batch.Write()` whenever the `flushThreshold` is exceeded — committing
   partial writes to PebbleDB before `SaveVersion()` even returns. A
   write-capturing shim would need to intercept every one of those intermediate
   flushes, buffer them, and replay them in a final batch — effectively
   implementing a write-ahead log on top of the DB layer.

2. **IAVL's `nodeDB` owns its batch**. The `BatchWithFlusher` is created inside
   `newNodeDB` and is not exposed. Routing its writes to an external batch
   requires either modifying `nodeDB` to accept an externally-provided batch, or
   replacing the DB instance it holds with a capturing proxy — both requiring
   changes to `tm2/pkg/iavl`.

The complexity and maintenance burden outweigh the benefit, given that Pebble
snapshots already provide the read isolation we need without requiring
atomicity.

### Keep the consensus mutex for queries

Revert PR5431 and go back to serializing all queries with `Commit`. This
prevents the cross-store inconsistency trivially and matches Cosmos SDK
behavior.

Not chosen because it reintroduces the DoS vector (flood simulate to block
consensus) and the data race on `checkState` that PR5431 fixed.

### Per-store snapshot (one snapshot per sub-store)

Take a separate Pebble snapshot for each sub-store and combine them. This is
equivalent to the current approach when all sub-stores share one DB instance
(which they do — they use key prefixes on the same PebbleDB), but fails to
generalize if sub-stores ever migrate to separate DB instances.

Not chosen because it is strictly more complex than one snapshot on the shared
DB, with no benefit given the current architecture.

## Consequences

- **Correct cross-store isolation**: queries and simulate always observe both
  sub-stores (IAVL and dbadapter) at the same committed block height.

- **No consensus blocking**: query and simulate paths hold a snapshot reference
  and never touch the live DB or any shared mutable state. Long-running queries
  do not delay `Commit`.

- **Snapshot resource usage**: each open snapshot prevents Pebble from
  compacting the key versions it covers. In practice, snapshots are held for the
  duration of a single query (milliseconds) and are reference-counted, so the
  compaction impact is negligible.

- **memdb in tests**: `MemDB.NewSnapshot()` copies the map at snapshot time.
  This is O(n) in the number of keys but is only used in tests; production nodes
  run PebbleDB.

- **Non-PebbleDB backends**: `GoLevelDB`, `BoltDB`, `LMDB`, and `MDBX` return
  `errors.New("snapshots not supported")` from `NewSnapshot()`. These backends
  are not used in production. If a node is started with one of them,
  `MultiImmutableCacheWrapWithVersion` will fall back to the previous
  `ImmutableDB`-wrapped live-DB behaviour and log a warning.
