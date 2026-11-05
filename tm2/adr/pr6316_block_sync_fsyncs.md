# PR6316: Drop the block store's redundant flush

## Status

Proposed

## Context

A node applying a block waits for the disk five times, across three databases:

| # | write | where | flush |
|---|---|---|---|
| 1 | the block's parts, meta and commits, then its height record | `BlockStore.SaveBlock` → `BlockStoreStateJSON.Save` (`tm2/pkg/bft/store/store.go`) | `SetSync` on the height record |
| 2 | an empty record, right after, commented "Flush" | `BlockStore.SaveBlock` | `SetSync(nil, nil)` |
| 3 | the ABCI responses | `ApplyBlock` → `SaveABCIResponses` (`tm2/pkg/bft/state/store.go`) | `SetSync`, since #6085 |
| 4 | the application commit | `rootmulti.Commit` (`tm2/pkg/store/rootmulti/store.go`) | `WriteSync` |
| 5 | the state record, with the validators and consensus params info | `ApplyBlock` → `SaveState` | `SetSync` |

In consensus the WAL adds a sixth. A node catching up through block sync paid
all five on every block until #6334, which holds block sync's writes in memory
and writes each database's share as one synced batch at a flush point, about
once a second. Since then only a live node waits for the disk on every block,
as it saves and applies each block consensus commits.

Flush 2 protects nothing. It goes to the same database as flush 1, right after
it, with no write in between, and a flush makes the database's earlier writes
durable too, so flush 1 already covers everything `SaveBlock` writes; flush 2
only writes an empty key and waits for the disk again. Upstream Tendermint
dropped the same line in v0.34: v0.33.9's `SaveBlock` ends with `// Flush` /
`bs.db.SetSync(nil, nil)`, v0.34's with "Save new BlockStoreState descriptor.
This also flushes the database." (`store/store.go`).

How each backend treats a flush, read from its code: pebbledb, the default,
writes its log in commit order, replays it up to the first lost record, and
syncs the old log before writing to a new one (pebble `db.go`, `rotateWAL`:
"Close the previous log first … syncs it to disk"), so a flush covers every
earlier write. boltdb syncs every write; lmdb and mdbx sync the whole
environment. goleveldb switches journals without syncing the old one
(goleveldb `db_state.go`, `newMem`), so there a flush does not cover writes
left in a journal it has switched away from; flush 2 changes nothing there
either, as it syncs the same journal flush 1 has just synced.

## Decision

Remove flush 2. A live node waits for the disk five times per block instead of
six. Block sync gains nothing since #6334: there flush 2's empty record is held
with the block's other writes until the flush point. Nothing else changes: the
height record is still written last and flushed, so the block store's
durability, and every case the restart handles, are as before.

## Alternatives considered

**Also drop flush 5, and let the restart rebuild a lost state record.** The
first version of this PR did this: it wrote the state record without a flush
(the next block's flush 3 covered it), reordered `ApplyBlock` so flush 3 also
covered the records the state record depends on, and generalized
`Handshaker.ReplayBlocks` to replay any gap between the state and the block
store. Crash recovery held at every write in a power-cut test, but the change
was dropped on review, for three reasons verified in the code:

- An unflushed write does not survive a process exit, not only a power cut. A
  pebble write without a sync stays in the process's memory: it does not wake
  the log writer (pebble `record/log_writer.go`, `SyncRecord`) and reaches the
  OS only with the next sync. `Node.OnStop` never closes the state or
  block-store database. So the state record would be lost on every exit: a
  kill, an out-of-memory crash, a panic, and a clean stop too.
- A governance halt panics in `BeginBlock` of the block after the halt height
  "so the block at haltHeight is fully committed" (`tm2/pkg/sdk/baseapp.go`).
  Without flush 5 the halt height's state record would be lost at every halt,
  and the binary started after it, the new one in an upgrade, would rebuild
  that state from the results the old binary saved: decoding them with its own
  types and stamping its own block version, either of which can fail where the
  old binary's state record would have been read as is. Closing the databases
  in `OnStop` would cover clean stops but not the halt: in block sync the halt
  crashes the process.
- Tools and older binaries assume the state is at most one block behind the
  block store (`contribs/gnobr`'s rollback; the old handshake panics past that);
  after any exit the state could be two behind.

**Overlap flush 5 with the next block's save.** Start the state record's flush
when `ApplyBlock` saves it and wait for it at the start of the next
`ApplyBlock`, before the block runs. The state stays durable before the next
block can halt or crash the node, as now, while in block sync its flush overlaps
the block store's flush for the next block. It needs a pending flush in
`BlockExecutor`, a wait on shutdown, and one executor across the handshake's
replays; left for a follow-up. Since #6334 block sync no longer flushes per
block, so this would now help only live nodes.

**Keep the block and the state in one database,** so that one flush covers
both: a data-layout migration for every node.

## Consequences

- One disk wait fewer per block on a live node. A reviewer timed each step
  over a full resync of gno.land mainnet (775,243 blocks, AWS t3.xlarge, gp3)
  on the first version, before #6334: `SaveBlock`, the only step this change
  touches, went from 2.21 ms to 1.23 ms per block (the first version's other
  removed flush was timed separately, in `SaveState`). Block sync now holds
  those writes instead; a live node's `SaveBlock` still writes and flushes as
  in that resync, so it should save about the same, an estimate.
- Block stores written before this change keep an empty key from flush 2;
  nothing reads it.

## Tests

- `TestSaveBlockFlushesOnlyTheHeightRecord` (`tm2/pkg/bft/store/store_test.go`):
  a saved block costs one flush, on the height record, and the height record
  is the last write. Fails on master, where the last write is flush 2's empty
  record.
- Not included, run on the change: a power-cut test that cuts the power after
  every write, flush and application commit of a four-block block-sync run,
  restarts on every prefix of each database's unflushed writes and resyncs.
  It passes, as on master, and fails with the height record's flush removed.
  It lands with the follow-up above, which needs it.

## AI assistance

Designed and implemented with AI assistance, including the analysis above and
the review rounds that cut the first version down to this change. The human
author reviewed and owns the change.
