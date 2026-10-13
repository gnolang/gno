# PR6316: Two fewer disk flushes per block, and a restart that replays any gap

## Status

Proposed

## Context

A node applying blocks writes each one across three separate databases (the
block store, the state store and the application store) and waits for the
disk five times. In order:

| # | write | where | flush |
|---|---|---|---|
| 1 | block parts, meta, commits, then the height record | `BlockStore.SaveBlock` → `BlockStoreStateJSON.Save` (`tm2/pkg/bft/store/store.go`) | `SetSync` |
| 2 | an empty record right after it, commented "Flush" | `BlockStore.SaveBlock` | `SetSync(nil, nil)` |
| 3 | the ABCI responses for the height | `ApplyBlock` → `SaveABCIResponses` (`tm2/pkg/bft/state/store.go`) | `SetSync`, since #6085 |
| 4 | the application commit | `ApplyBlock` → `Commit` → `rootmulti.Commit` (`tm2/pkg/store/rootmulti/store.go`) | `WriteSync` |
| 5 | the state record, with the validators and consensus params info | `ApplyBlock` → `SaveState` | `SetSync` |

In consensus the WAL adds a sixth (#6085's ADR counts them). A node catching
up through block sync pays the five on every block, and the disk waits are
most of its time: see Measurements.

### What each flush protects

The restart logic (`Handshaker.ReplayBlocks`, `tm2/pkg/bft/consensus/replay.go`)
recovers from a crash by replaying. Before this change it tolerated the app
behind the store (replay through the app) and the state behind the store by
exactly one block (replay that block; against a mock app fed the saved
responses when the app had already committed it). It refused the app ahead of
the store (fatal: nothing can rebuild a block the store never kept) and the
state more than one block behind the store (panic). In consensus there is one
more dependency: the WAL's height marker is written after the block is saved,
and WAL replay refuses to run when the store lacks the block the marker
implies (`catchupReplay`).

The three databases are three files. After a power cut each file keeps some
prefix of its writes, independently of the others: pebble's log is sequential
and checked record by record, and the kernel writes back the dirty pages of
different files in no particular order. So "written before" across files
guarantees nothing; only a flush does. Within one file a flush makes every
earlier write durable too. (A process crash without a power cut loses nothing,
flushed or not; a clean stop flushes everything, pebble's `Close` syncs the
log.)

That gives each flush its verdict:

- **#1 stays.** It keeps the block on disk before the app commits it (the
  fatal case) and, in consensus, before the WAL marker.
- **#2 goes.** Same file as #1, written right after #1's flush with nothing in
  between: it makes nothing durable that was not already.
- **#3 stays.** It keeps the responses on disk before the app commits (#6085).
  An app ahead of its responses is unrecoverable because the application store
  cannot roll back a version: bptree's `LoadVersionForOverwriting` and
  `DeleteVersionsFrom` return `ErrUnsupported` ("would leak values and nodes"),
  and ABCI has no rollback message (#6085's ADR reaches the same conclusion).
- **#4 is the application's own.** Tendermint never relies on it: an app that
  is behind is replayed from the store. Out of scope here.
- **#5 goes.** It protects nothing: the state is recomputed from #1 and #3 on
  restart. It was flushed only so that the state is never more than one block
  behind the store, which the restart logic insisted on. The other `SaveState`
  callers need no flush either: the genesis state, the state after InitChain
  and the app-version update are recomputed on every start (the application
  does not commit InitChain, so a restart before block 1 sends it again), and
  once block 1 is applied, #3's flush of the same database has made the
  InitChain-time record durable before the app committed.

## Decision

Drop #2. Make `SaveState` write without a flush, for every caller; the state
database is flushed once per block, by #3, and that flush covers every state
record written before it. The rule it relies on, "a flush makes every earlier
write to the same DB durable", is now documented once, on `dbm.DB.SetSync`
(and `DeleteSync`, `Batch.WriteSync`); the disk backends meet it, per #6085's
survey. Replace the restart logic's "at most one block behind" rule by the
general one. With app ≤ store and state ≤ store, which the remaining flushes
guarantee:

1. bring the app up to the state with the real app, without touching the
   state (as before);
2. bring the state up to the store one block at a time through `ApplyBlock`,
   which saves the responses and the state as a live node does. A height the
   app has already committed is replayed against the mock app with the
   responses saved before that commit and the app hash that followed it: the
   next block's header carries it, except for the app's latest height, whose
   hash the app gave in the handshake. The heights past the app run on the
   real app.

The old code's cases are the instances of this rule with the state at most
one block behind. The app-to-state leg also checks the first replayed block's
header against the hash the app reported (before, only later blocks were
checked), except for a fresh app, whose reported hash predates the InitChain
just sent. Per block, block sync now waits for the disk three times instead
of five (two of tendermint's flushes, one of the app's); consensus four times
instead of six.

### Write order and crash positions

Per height h, with † marking a flush: S_h† (store) · R_h† (responses) · A_h†
(app) · T_h (state, no flush) · S_h+1† · R_h+1† · …

| crash lands after | durable heights | restart |
|---|---|---|
| S_h+1† | store h+1, app h, state h or h-1 (T_h may be lost) | if the state is at h-1: mock replay of h (R_h is durable: flushed before A_h); then real replay of h+1 |
| R_h+1† | store h+1, app h, state h (R_h+1† flushed T_h too) | real replay of h+1 |
| A_h+1† | store h+1, app h+1, state h | mock replay of h+1 |
| T_h+1 | store h+1, app h+1, state h+1 or h | nothing, or the mock replay of h+1 |

T_h is durable at the latest with R_h+1†, which is in the same file, so the
state record is at most one block behind on disk whenever the node is running.
If the application's own flush is ever skipped (a possible follow-up in the
app), the app can also be any number of blocks behind; the rule above covers
that too, because every height the app has and the state lacks has its
responses on disk (R_h† precedes A_h).

## Alternatives considered

**Drop #2 only, keep #5.** Four flushes, no restart change. Rejected: half the
gain, and the restart change is what makes #5 unnecessary in the first place.

**Keep a flushing `SaveState` for the genesis and InitChain saves and add an
unflushed variant for `ApplyBlock`.** The first draft did this. Rejected on
review: those saves need no flush either (above), so a second exported
function would only move a choice to callers that have nothing to choose.

**Put the responses in the block store's database, or merge the state and
block store databases.** Then #1 covers #3: one tendermint flush per block.
Rejected for now: it changes the data layout of every existing node and needs
a migration, for one flush.

**Roll the application back on restart** (to the state's height, then replay
everything with the real app). It would allow zero per-block tendermint
flushes with a periodic barrier. Rejected: the production store has no rollback
(above) and ABCI has no message for it; it would be a different, larger change.

**Flush #1 asynchronously**: start the flush when the block is saved and wait
for it right before the app commits, so the block's execution overlaps the disk
wait. A candidate follow-up; left out to keep this change a straight-line
reordering with no concurrency.

**Save several blocks before flushing once** (amortize #1). Rejected: a block
can only be verified against the state after the previous block was applied,
so blocks cannot be verified ahead and saved in bulk.

**Skip the application's flush (#4) while catching up.** Safe for consistency
(an app that is behind is replayed) at the price of a longer restart after a
power cut during catch-up. A separate change in the application, not in tm2.

## Consequences

- The state record is no longer durable per block. After a power cut the
  restart replays the blocks it lacks, as it already did for the app; each
  replayed block costs what applying it costs.
- The restart no longer panics on a store more than one block ahead of the
  state; it replays. The two remaining sanity checks stay: the app ahead of
  the store is an error, the state ahead of the store a panic. A block store
  from another chain still fails, through `ValidateBlock` on the first
  replayed block. A block store copied from another node of the same chain
  onto a fresh state and app directory, which used to hit the panic, now
  syncs from it.
- After a power cut, a data directory written by this code can have its state
  two or more blocks behind the store, which binaries before this change refuse
  to start on ("StoreBlockHeight > StateBlockHeight + 1"). Starting the new
  binary once repairs it. On-disk formats are unchanged in both directions.
- `SaveState` still discards write errors; #6085's ADR lists that as its own
  follow-up.
- An external caller that relied on `SaveState` flushing (none in the
  repository; `contribs/gnobr` patches a state offline and closes the
  database, which flushes) now gets the write flushed by the next
  `SaveABCIResponses` or by `Close`.
- When the app is many blocks behind the state, the no-mutation replay loads
  the validators from the database for each height (#6308 only skips the load
  for the state's own height), so that leg replays at the old rate. Pre-existing.

## Tests

- `TestSaveBlockFlushesOnlyTheHeightRecord` (`tm2/pkg/bft/store/store_test.go`):
  a saved block costs one flush, on the height record, and that record is the
  last write, so the flush covers the block. Fails on the old code (the last
  write is the empty flush).
- `TestApplyBlockFlushesABCIResponsesBeforeAppCommit`
  (`tm2/pkg/bft/state/execution_test.go`, #6085's test extended): the responses
  record is the only flush `ApplyBlock` issues on the state database, and the
  durable state is one block behind right after. Fails on the old code (two
  flushes).
- `TestHandshakeReplay{All,Some,One,None}` (`tm2/pkg/bft/consensus/replay_test.go`)
  gain two shapes. Mode 3: the state two blocks behind the store, with the app
  behind the state (two real replays after catching the app up), between (a
  mock replay with the header's app hash, then a real one) or level with the
  store (two mock replays). Mode 4: the state back at genesis, with the app
  anywhere from fresh (a block store copied onto fresh state and app
  directories: six real replays) to level with the store (six mock replays).
  The old code panics on all of them (`StoreBlockHeight (6) > StateBlockHeight
  + 1 (5)`). The expected replay count is derived from the built state's and
  app's heights rather than listed per mode, and the final app hash is checked
  against the state record at the store's height.
- #6085's tests still pass: the responses are flushed before the app commits,
  and a failed responses write stops the block.

## Measurements

Four syncs of gno.land mainnet from genesis to height ~31k (the empty-block
range after the genesis transactions), on a 4-vCPU DigitalOcean droplet (NVMe,
pebbledb), peered to the same sentry and public peers, back to back, from
binaries built on v1.5.0 + #6308 + #6315, so the CPU side matches master's
sync path. The rate is blocks per second between heights ~3.4k and ~30k,
sampled every 10 s from `/status`; the ~200 s each run spends at height 1
while the block pool fills is excluded.

| build | blocks/s | ms per block |
|---|---|---|
| v1.5.0 + #6308 + #6315 | 136 | 7.3 |
| + this change (no empty flush, state record unflushed) | 196 | 5.1 |
| + the application's commit unflushed too (not in this change) | 202 | 5.0 |
| + every pebble write unflushed (the ceiling; not crash-safe) | 447 | 2.2 |

So this change makes light blocks about 1.4× faster to sync (7.3 → 5.1 ms per
block). The two reference binaries measured 138 and 390 blocks/s on another
run the day before, so run-to-run noise is about ±15% (other jobs share the
box); the ordering held both times. Leaving the application's own flush out
made no measurable difference on this disk. In transaction-heavy ranges a
block costs 10–15 ms of CPU, so the saving is the same ~2 ms per block but a
smaller share.

## AI assistance

Designed and implemented with AI assistance: the per-flush analysis above, the
restart rule, the tests and this record; the diff and the record were revised
through review rounds (one of which removed the first draft's second
`SaveState` variant and moved the durability rule onto the `dbm.DB` contract).
The human author reviewed and owns the change.
