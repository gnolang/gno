# PRXXXX: Make the ABCI responses crash-recovery record durable

## Status

Proposed

## Context

`BlockExecutor.ApplyBlock` persists a block in this order
(`tm2/pkg/bft/state/execution.go`):

1. `SaveABCIResponses(db, H, …)` — the record that exists so a crash between the
   application commit and the state save can be recovered from
   (`tm2/pkg/bft/state/store.go`).
2. `saveTxResultIndex` for each tx — the by-hash lookup index.
3. `blockExec.Commit` → `proxyApp.CommitSync()` — the application store commit,
   one atomic fsynced batch (`tm2/pkg/store/rootmulti/store.go`, `WriteSync`).
4. `SaveState(db, state)` — ends in `db.SetSync`
   (`tm2/pkg/bft/state/store.go`).

The block store was already flushed before `ApplyBlock` runs
(`tm2/pkg/bft/store/store.go`, trailing `SetSync`).

Step 1 was written with `db.Set`. The `DB` interface documents no durability
guarantee for `Set`, and the two WAL-based backends implement it as an unsynced
write: pebbledb passes `pebble.NoSync` (`tm2/pkg/db/pebbledb/pebbledb.go`) and
goleveldb omits `Sync` from its write options
(`tm2/pkg/db/goleveldb/go_level_db.go`). pebbledb is the production default
(`DBBackend: "pebbledb"`, `tm2/pkg/bft/config/config.go`). On those backends the
next write that flushes the state store's log is step 4, so the record whose only
purpose is surviving a crash *before* step 4 only reached the disk *because of*
step 4. In the window between step 3 and step 4 it provided no protection at all,
which is the sole window it was written for.

(The transaction-backed backends — boltdb, lmdb, mdbx — commit a transaction per
`Set` and are durable either way; boltdb's `SetSync` is literally `return
bdb.Set(key, value)`. The bug is specific to the WAL-based backends, which
includes the default.)

A process death in that window leaves:

```
block store = H   (flushed before ApplyBlock)
application = H   (flushed at step 3)
state       = H-1 (step 4 lost)
responses(H)      (step 1 lost with it)
```

`Handshaker.ReplayBlocks` (`tm2/pkg/bft/consensus/replay.go`) handles this as its
`storeBlockHeight == stateBlockHeight+1`, `appBlockHeight == storeBlockHeight`
case — note there is a second, unrelated `appBlockHeight == storeBlockHeight`
case under `storeBlockHeight == stateBlockHeight`, which is the healthy "we're all
synced up" path. The skewed case replays H against a mock application that serves
the *stored* responses, because the real application must not execute a height it
has already committed. `LoadABCIResponses` is the only source for them, and there
is no fallback, so the node fails on every subsequent start with `error during
handshake: error on replay: Could not find results for height #H` and cannot
recover without operator surgery on the data directory.

This was observed in production on a non-validator node stopped during fast sync
(app = store = H, state = H-1, responses for H absent). It crash-looped until it
was restored from a snapshot.

## Decision

Write the record with `db.SetSync`.

That restores the invariant the ordering assumes: the responses for H are on
disk before the application commits H, so every position a crash can land in
maps onto a handshake case that recovers.

| Crash position | Resulting heights | Handshake path |
|---|---|---|
| before the responses are flushed | app = state = H-1, store = H | replay H against the real application |
| after them, before the application commit | app = state = H-1, store = H | replay H against the real application |
| after the application commit, before the state save | app = store = H, state = H-1 | replay H against the mock application, responses guaranteed present |

## Alternatives considered

**Flush the state store explicitly before `proxyApp.CommitSync()` instead.**
Same fsync count and the same guarantee, but it puts the requirement in the
caller and leaves the exported function's documented contract still unmet for
any other caller. Rejected as the more fragile placement.

**Keep the per-height record unsynced and add a separate synced key, as the
v0.34–v0.38 upstream line does.** Rejected because the condition that scheme
exists to serve does not hold in tm2. Upstream split the record so that the
per-height copy could be **thrown away**: #9090 (closing #8028, "Flag to save
ABCIResults", and #8946, "Optimize ABCI Response Information stored as part of
state.db to validate latest height") added a config flag "which enables
discarding of abci responses […] so the node operator can decide if they would
like to save or discard the responses for efficiency purposes". Once the
per-height copy is optional, crash recovery needs a key that cannot be
discarded — hence `lastABCIResponseKey`.

tm2 has no such flag, and cannot have one: the per-height responses are
load-bearing for the RPC tx-results endpoint. `Tx()` in
`tm2/pkg/bft/rpc/core/tx.go` resolves a tx hash to a height through the tx
result index and then reads `LoadABCIResponses` at that height, so the
per-height key must exist and must be complete for every height a node serves.
There is no discard or prune path for it anywhere in tm2.

Both halves of that come from the same commit: #1546 added the endpoint that
made the per-height key undiscardable *and* removed the flush that made it
durable. Since the key has to be there regardless, making it durable is the
minimal correct design; adding a second synced key would introduce a redundant
second source of truth for data the first copy is already obliged to hold.

The upstream history, for the record:

- Tendermint v0.33 (and earlier) wrote the per-height key with `SetSync`
  (`state/store.go` at v0.33.9). That is the code tm2 forked.
- Tendermint #9090, backported as #9159 (merged 2022-08-11, merge commit
  `fbd754b4ded5612b5031d09c275c276221cee398`, base `v0.34.x`) downgraded the
  per-height key to `Set` — but *simultaneously* introduced
  `lastABCIResponseKey`, written with `SetSync` under the comment "We always save
  the last ABCI response for crash recovery". The per-height copy became a
  discardable query index; the synced single key took over the recovery duty.
  Verified present at v0.34.24 (`Set` at line 460, `SetSync(lastABCIResponseKey)`
  at line 476) and unchanged through the v0.37 and v0.38 lines (v0.37.15 lines
  468 and 484; v0.38.17 likewise).
- gno #1546 (`3d1d26cbb`, "feat(tm2): store tx results and add endpoint to query
  them") took the downgrade **without** the compensating synced key. That is
  precisely this bug: tm2 ended up with neither upstream's synced per-height key
  nor upstream's synced last-response key.
- CometBFT v1.0 uses a single per-height `SetSync` and marks
  `lastABCIResponseKey` deprecated (v1.0.1 `state/store.go`: `SetSync` at line
  763, the key marked `// DEPRECATED` at line 67 with a read-only compatibility
  path). That is the same design as this change.

**Add a fallback to the replay path for stores that are already skewed.**
Rejected as unsound. Rebuilding the state for H needs `EndBlock`'s validator
updates and consensus-param updates and the results hash
(`updateState`, `tm2/pkg/bft/state/execution.go`). None are derivable from the
block, and H is the store tip, so there is no H+1 header to cross-check a guess
against. Assuming "no validator change" would silently produce a wrong
validator set — a fork, which is worse than refusing to start.

Recovering an already-skewed data directory requires rolling the application
store back one version so the handshake takes the real-application path. The
version history is retained by `rootmulti`, but the handshaker reaches the
application only over ABCI, which has no rollback message, so this belongs in an
offline command rather than in the handshake. Upstream agrees: Tendermint added a
`rollback` command in v0.34.14 and CometBFT still ships it
(`cmd/cometbft/commands/rollback.go`), documented as overwriting "a state at
height n with the state at height n - 1" and noting that "The application should
also roll back to height n - 1" — and that without `--hard` block n is kept, so
restarting re-executes it against the application, which is exactly the
real-application replay path. Left as follow-up.

**Drain in-flight block application on shutdown.** `BlockchainReactor.OnStop`
stops the block pool without waiting for `poolRoutine`, which may be inside
`ApplyBlock`, and the consensus reactor's `conS.Wait()` drain is skipped during
fast sync. A catching-up node is inside `ApplyBlock` almost continuously, which
is why a routine stop hit this window in production. Worth fixing, but it only
narrows the window; it does not make any single crash position recoverable.
Left as follow-up, out of scope here.

## Consequences

One additional fsync per block on the state store. In steady-state consensus the
per-block path goes from four fsyncs to five: the block store
(`tm2/pkg/bft/store/store.go`), the consensus WAL
(`tm2/pkg/bft/consensus/state.go`), the application store
(`tm2/pkg/store/rootmulti/store.go`), and `SaveState`, plus this one — against a
5 s default commit interval (`TimeoutCommit`,
`tm2/pkg/bft/consensus/config/config.go`). Bulk handshake replay is unaffected,
because `ExecCommitBlock` does not save responses. Fast sync pays the extra fsync
per block, which is precisely the path where the production failure occurred.
Paying it is not optional: without a flush the record cannot do the one job it
exists for.

`SaveABCIResponses` discards the error from `SetSync`, matching the convention of
every other write in the file. So the invariant is conditional on the write
succeeding rather than unconditional. This is a deliberately unchanged residual:
an I/O failure severe enough to fail this sync would also fail the application
store's `WriteSync` a moment later, and that one panics
(`tm2/pkg/store/rootmulti/store.go`, `panic("rootmulti: Commit() failed: …")`),
so a real disk failure crashes loudly instead of silently producing the skew.
Threading an error return out of `SaveABCIResponses` would change the signature
of an exported function and every caller, and is left out of a fix meant to be
minimal.

The tx result index (step 2) stays unsynced. It is a lookup index, not a
recovery record, and `ApplyBlock` rewrites it when the block is replayed.

Data directories already in the skewed state are not repaired by this change.
It prevents new ones; recovering an existing one still means restoring from a
validated snapshot, or the offline rollback command noted above.

## Tests

`TestSaveABCIResponsesIsDurable` (`tm2/pkg/bft/state/store_test.go`) pins the
exported function's contract.

`TestApplyBlockFlushesABCIResponsesBeforeAppCommit`
(`tm2/pkg/bft/state/execution_test.go`) covers the ordering through
`ApplyBlock`: it snapshots what a restarting process would find at the instant
the application commits, and asserts the state is still at H-1 (so the snapshot
really is inside the window) and that the responses for H are readable from it.

Both use `unsyncedWriteDB` (`tm2/pkg/bft/state/helpers_test.go`), which models
the durability behaviour of the WAL-based backends — an unsynced write survives
only if a later synced write flushes the log — rather than faking fsync
semantics. Both save a non-empty payload and assert on its contents, so they
cannot pass on the present-but-empty case. Both fail before the change with the
production error, `Could not find results for height #1`.

## AI assistance

Implemented with AI assistance: the root-cause analysis was done against the
bricked data directory, the fix was driven test-first, and the diff and this
record were revised through several review rounds (which corrected an
overbroad claim about backend durability and sharpened the upstream history
above). The human author reviewed and owns the change.
