# PR6334: Deferred writes during block sync, with ordered flush points

## Status

Accepted; implemented in this PR. Citations with line numbers are to
origin/master `21f9abd`; names without one are identifiers, on master or
added by this PR.

## Context

A node catching up through block sync applies one block at a time and waits for
the disk five times per block, across three databases: the block store (two
flushes, one of them redundant and removed by #6316), the ABCI results, the
application commit, and the state record. On the early mainnet chain, where
blocks are nearly empty, a block took ~7.2 ms with #6308 and #6315; with every
flush disabled the same build ran at ~2.6 ms per block (devbox, 2026-10-13).
Over the whole chain a block averages ~17 ms on master (a reviewer's full
resync, without #6308 and #6315), of which the flushes are roughly a third.
At least 82% of mainnet blocks carry no transaction, so these fixed per-block
costs dominate.

Each of the four flushes that remain after #6316 protects something (#6316's
ADR): the block store's puts the block on disk before the app commits it, the
results' put them there before the app commits, the app's own makes its commit
durable, and the state record's lands before the next block can halt the node.
A governance halt panics in `BeginBlock` of the block after the halt height "so
the block at haltHeight is fully committed" (`tm2/pkg/sdk/baseapp.go:593-598`),
and the binary started after the halt starts from what is on disk
(`gno.land/pkg/gnoland/node_params.go:170-179` refuses an app below the halt
height). Removing any of them one by one breaks one of these.

This change keeps every guarantee and removes the per-block waits during block
sync by changing when things are written, not what is protected: the three
databases hold a block's writes in memory, and a flush point writes many
blocks' writes in an order that leaves them consistent after a crash at any
instant.

Prior art: no Tendermint-family node holds the block store, state and app
writes across blocks during block sync; upstream batches each block's own
writes per store (Tendermint #6018, CometBFT #1735), and the block-sync loop
still carries Tendermint's "batch saves so we dont persist to disk every
block" (`blockchain/reactor.go:367-371`). The nearest complete design is
Bitcoin Core's `FlushStateToDisk`: block files first, then the chainstate, with
a record of the flush in progress (#10148's `DB_HEAD_BLOCKS`, the new and old
tips) from which a restart rolls between the two tips and refuses any other
shape, as the flush-point record does here. Sei's asynchronous app commits
lost the block at an upgrade height (sei-chain #4287), the case
`NeedsFlushPoint` covers at the halt height. gno #6264 overlaps #6316, not
this change; CometBFT #5753, which verifies each commit once instead of twice,
is a separate saving that applies here too.

## Decision

### Deferred mode

During block sync only (`BlockchainReactor.poolRoutine`), and only with an
application that supports it (see Interface), the node defers its writes:

- **Block store:** the block's meta, parts and commits and the height record go
  to an in-memory overlay over its database. Point reads see them (`LoadBlock`
  for peers and RPC), and the in-memory `Height()` advances per block as today.
  Blocks are still saved only after they are verified, as they are applied (see
  Alternatives).
- **State database:** the results, the tx index, the validators and consensus
  params info and the state record go to an in-memory overlay. Point reads see
  them: `ApplyBlock` reads `LoadValidators(H-1)` for every block
  (`state/execution.go:325-341`), which a block a few heights earlier may have
  written.
- **Application:** rootmulti already collects a block's writes in a
  `BatchCollector` whose point reads see them
  (`tm2/pkg/db/collecting.go:155-175`); while it holds, `Commit` leaves them
  there. The bptree store reads its nodes, orphans, roots, versions and fast
  index through point reads (`tm2/pkg/bptree/nodedb.go:139-193, 304-336,
  369-389, 425-438`, `fast_index.go:91-105`).
- **Events** of each block are held (`BlockExecutor.HoldEvents`), and fired
  after the flush point that writes the block (`FireHeldEvents`).

The node opens the two tendermint databases wrapped in a new `db.HoldingDB`
(`node.initDBs`), which has a `CollectingDB` and a collector of its own: until
block sync calls `Hold(true)`, it sends every call straight to the database,
with its own sync. The handshake and consensus rely on those syncs, for
example on the synced results of #6085 (`state/store.go:214-217`). While
holding, writes and batches go to the collector, `Get` and `Has` read it first,
and `WriteHeld` writes it out. rootmulti's collector always collects, as today:
its one batch per write is what makes an application commit all-or-nothing
(`rootmulti/store.go:287-308`). The three databases hold through the same
switch, `BatchCollector.SetHolding`, which panics when turned on or off with
writes collected. rootmulti's `SetHolding(true)` therefore first writes what
its collector holds from outside a commit, such as a store's maintenance at
load (the bptree fast index's rebuild: `tm2/pkg/bptree/fast_index.go:146-148`),
which belongs to the version on disk: a window with no block then holds
nothing, and block sync can switch to consensus right after it starts, as at
every restart at the tip. Every reader, the RPC environment included (`node/node.go:862-866`),
reads through the wrappers. `BatchCollector.set` now copies with `cp`, which
keeps an empty value non-nil, as its batches already do
(`collecting.go:113-115` against `229-231`, `util.go:7-11`): an empty block's
results encode to no bytes, and a held nil reads as missing
(`state/store.go:183-185`). This also reaches consensus mode, where it changes
only one read: an empty value set directly and read back before the write now
reads as the disk already returns it after the commit.

`CollectingDB`'s iterators read only the real database
(`collecting.go:177-188`), so while its collector holds they panic with
`db.ErrIteratorOverHeldWrites` rather than miss the held writes. Nothing
iterates these databases inside a block or in `CheckTx` today: the bptree
store's own iterators walk its tree through point reads, its database iterators
run only at load and state import (`fast_index.go:155`, `nodedb.go:479`;
`nodedb.go:449` has no caller outside tests), its pruning reads through point
reads (`tm2/pkg/bptree/prune.go`), `CheckTx` runs no messages
(`baseapp.go:718`), and queries read stores built without the collector
(`rootmulti/store.go:413-447`). But `runTx` turns a panic into a failed
transaction (`baseapp.go:828-859`), and so does the VM for its messages
(`gno.land/pkg/sdk/vm/keeper.go:1435-1469`); here that would be a result no
other node has, which the next flush point could write before the next block's
header check stops the node (`state/validation.go:70-87`). So a refused
iterator is also recorded on its collector, and in deferred mode BaseApp clears
the application's record (`CommitHolder.TakeRefusedIterator`) before each
delivered transaction, `BeginBlock` and `EndBlock`, and panics after each one
if it is set (for a transaction, in `DeliverTx` after `runTx` returns,
`baseapp.go:648-655`, outside every recover). The block then fails before its
`Commit` and the process ends, since neither the local client nor block sync
recovers it (`tm2/pkg/bft/abci/client/local_client.go:91-97`,
`blockchain/reactor.go:373-377`). `CheckTx` never overlaps those calls (the
local client's mutex is shared: `tm2/pkg/bft/proxy/client.go:43`), so an
iterator there only rejects its transaction, and the record it leaves is
cleared before the block's next call. A node that hits this stops at that
block on every start, with nothing of it on disk, until the iterator is fixed;
meanwhile `fast_sync = false` lets it catch up through consensus instead
(`tm2/pkg/bft/config/config.go:286`, `node/node.go:559`).

In consensus mode every write still reaches the disk with the same flushes. A
failed write of the block store's height record is now fatal, a commit's batch
carries only the last write to each key, and the stop order below applies to
every node.

### Flush points

A flush point (`blockchain.DeferredWrites.FlushPoint`) writes, in this order,
each as one all-or-nothing write followed by a flush:

1. the block store, with the flush point's record (below),
2. the state database,
3. the application (`WriteCommits`),

then fires the held events of its blocks, in block order. A window with no new
block writes nothing.

Each overlay is cleared only after its batch is on disk:
`BatchCollector.WriteTo` drains the collector into one batch sized from the
held bytes, writes it with a sync, and only then resets the collector, which
frees its memory. `Drain` no longer clears, and skips an operation that a later
one on the same key overrides, since only the last one can reach the disk.
Until the reset, readers on other goroutines (peers, RPC, `CheckTx`) keep
finding held records in the overlay, and after it on disk. Nothing more is
needed, because no writer runs during a flush point: it runs on the block-sync
goroutine, between blocks. The application's write is a direct call, not an
ABCI message, so it runs outside the local client's mutex, which is safe for
the same reason. A failed write is fatal: the node panics and is never
recovered, because continuing would write the next window over a gap.

A flush point is taken between blocks, and only there:

- every T seconds, or when the held writes of a database exceed a byte budget
  B: the block store's, the state's or the application's (`HeldBytes`). The
  budget counts every operation, overwrites included (`BatchCollector.Size`),
  and stays well under each backend's limit for one write (pebble panics near
  4 GiB). T and B are `flushPointInterval`, one second, and `flushPointBytes`,
  64 MiB, beside `trySyncIntervalMS` (`blockchain/reactor.go:22-31`). The
  routine checks after each applied block, and on its once-a-second switch
  ticker, so that a window whose blocks stop arriving, as when every peer
  leaves, is written too. The interval runs from the end of the last flush
  point, so a slow one does not make every block take its own;
- when the application needs one (`NeedsFlushPoint`, read after each
  `ApplyBlock`): its halt height is reached, so the halt height is on disk
  before the `BeginBlock` that panics; or the block wrote a halt param
  (`RequireFlushPoint`), so a scheduled halt is on disk from the block that
  scheduled it (the upgraded binary's startup check reads it from the
  committed state: `node_params.go:135-182`);
- before switching to consensus (`blockchain/reactor.go:294-310`), in this
  order: caught up, the flush point, deferral off, `pool.Stop`, and
  `switchToConsensusFn` only if that `pool.Stop` succeeded. So no consensus
  block, whose WAL marker is synced (`consensus/state.go:1369-1386`), lands in
  an overlay, and a stop that arrives before or during that flush point does
  not start consensus: the pool stops only once
  (`tm2/pkg/service/service.go:160-165`), so after `OnStop` has stopped it the
  routine's `pool.Stop` fails and the routine returns instead.

Never after a failed block: a block that panics in the application's `Commit`
can leave part of its writes in the collector, and the base store keeps no
versions to fall back to (`tm2/pkg/store/dbadapter/store.go:91-97, 128-130`). A
failed block ends the process as today, and the window is lost with it.

The application never writes on its own, not in `Commit` and not in `Close`: an
application on disk ahead of tendermint's records is the one state the restart
cannot recover from (#6085). A stop drops the held writes, as a crash does; the
last flush point is consistent without them, and block sync fetches the window
again. (A stop that arrives while block sync switches to consensus lets the
flush point before the switch finish.) The block-sync reactor's `OnStop` waits
for its routine, if `OnStart` started one (only with fast sync on:
`reactor.go:146-156`), so that gnoland does not close the application database
under a flush point (`gno.land/cmd/gnoland/start.go:322-330`). Its main loop
watches the pool stopping instead of its own `Quit` (`reactor.go:389-390`),
which closes only after `OnStop` returns (`tm2/pkg/service/service.go:175-176`)
and so cannot end a routine that `OnStop` waits for; `OnStop` stops the pool
first (`reactor.go:160`), and today only the request goroutine watches it
(`reactor.go:266`). The p2p switch stops reactors in no
fixed order (`tm2/pkg/p2p/switch.go:240`), and a consensus reactor stopped
first leaves its never-started consensus state startable
(`tm2/pkg/service/service.go:167-171`), so the node, after closing its
private validator and its p2p transport, stops the block-sync reactor first
(the p2p switch then skips it rather than log its `ErrAlreadyStopped`), then
the p2p switch with the rest, then the RPC environment, the event switch and
the event store, which today stop before the transport and the reactors
(`node/node.go:821-837`): every node
registers the event store's listener (`node.go:262-271`), and a flush point
that completes during a stop fires its window's events, which, once the event
switch has stopped, panic on that listener's closed channel
(`tm2/pkg/events/subscribe.go:53-56`).

### What a crash leaves, and the restart

Between flush points nothing of the window is on disk, so a crash leaves all
three databases at the last flush point F, and block sync fetches the window
again. Inside a flush point:

| crash after | block store | state | app | restart |
|---|---|---|---|---|
| nothing of F' written | F | F | F | nothing to replay |
| step 1 | F' | F | F | lower the block store to F+1, then replay F+1 on the real app (existing case) |
| step 2 | F' | F' | F | replay F+1..F' on the real app without changing the state (existing case) |
| step 3 | F' | F' | F' | nothing to replay |

After a crash between steps 1 and 2 the block store is more than one block
ahead of the state, where the handshake panics today
(`consensus/replay.go:389-391`). Step 1's batch therefore also records the
flush point it belongs to, `store.FlushPoint{From: F, To: F'}`, under a key of
its own (`blockStoreFlushPoint`), which older binaries ignore. When that record
says F to F', the state is at F, the block store at F' and the application at
most at F, the handshake (`Handshaker.lowerAfterFlushPointCrash`) lowers the
block store's height to F+1 before replaying, on disk and in memory
(`BlockStore.LowerHeight`, since the handshake reads the in-memory height:
`replay.go:286`; the height record is one synced write,
`tm2/pkg/bft/store/store.go:241-255`, whose error `Save` ignored and now
checks). Any other shape, such as a partial restore, panics as today and leaves
the directory as it is. The blocks above F+1 were verified, and stay on disk, above
the height, until block sync saves those heights again (`LoadBlock` does not
check the height, so a peer asking for one is served a valid block of the
chain); block sync fetches them again, as after a crash a moment earlier. Every shape a crash can leave is then one
the handshake handles today, including a crash during its own one-block replay
(the mock app: `replay.go:427-436`) and a chain with InitialHeight > 1
(`replay.go:464`).

Events stay at most once, as today (`state/execution.go:193-199`). A crash
between flush points has fired none of the window's events, and the window is
applied again. After a crash following step 1, the handshake's replay of F+1
fires its events, and the rest come with the window fetched again. After step 2
the window is replayed without events (`ExecCommitBlock`), and after step 3 it
is not replayed, so its events are lost, as one block's are today after a crash
between `SaveState` and `fireEvents`.

Operators: a binary from before this change panics on store > state+1, so after
a crash between steps 1 and 2, start this binary once before downgrading.
Offline tools (`contribs/gnobr`, the `gnogenesis` fork) assume store ≤ state+1
and app = state, so start the node once after a crash during block sync before
running them. `gnobr` now refuses, before it trims the block store, to roll such
a directory back to a target above its state (`checkRollbackable`): it would
leave the state at its height and patch its hashes from the block after the
target, a later height's (`contribs/gnobr/main.go:83-136`). A target at or
below the state rolls back as on any directory. After a crash
inside the flush point at the halt height, the application on disk is below
the halt, so the upgraded binary refuses (`node_params.go:172-179`): start the
previous binary once, then upgrade.

### Interface

tm2 runs the application in process only
(`tm2/pkg/bft/proxy/client.go:130-151`), so the application side is an optional
Go interface, `abci.DeferredCommitter`, that `NewNode` asserts on the
application its local client creator calls (`LocalApp`, a new method of that
creator), so that only the application that answers the node's ABCI calls is
ever deferred:

```go
type DeferredCommitter interface {
	// DeferCommits makes Commit hold its writes in memory (on), or write them
	// as before (off), and reports whether the application defers. It is
	// called only between blocks, with nothing held when turned off.
	DeferCommits(on bool) bool
	// WriteCommits writes every held commit as one all-or-nothing write.
	WriteCommits() error
	// HeldBytes returns the bytes of the held writes, which the node keeps
	// under its budget.
	HeldBytes() int64
	// NeedsFlushPoint reports that the node must write everything after the
	// current block: the halt height is reached, or a flush point was required.
	NeedsFlushPoint() bool
}
```

Tendermint owns the policy, T and B, the application's bytes included; the
application reports only what correctness requires. If the application does not
implement the interface, or `DeferCommits(true)` returns false, nothing is
deferred (`DeferredWrites.Start`): deferring only tendermint's side would let
the application's commits reach disk ahead of the block store, which the
restart cannot recover from.

### Application changes (tm2/pkg/sdk, tm2/pkg/store, gno.land)

- Deferral is a BaseApp option, `sdk.AllowDeferredCommits`, off by default;
  gno.land turns it on in `NewAppWithOptions`. Other tm2 applications keep
  today's behavior: the IAVL store iterates the database inside a block (its
  fast iterator: `tm2/pkg/iavl/mutable_tree.go:240-249`,
  `tm2/pkg/iavl/nodedb.go:1029-1049`; its pruning's version lookups:
  `nodedb.go:755, 804, 846`), which would hit the panic above. gno.land mounts
  only bptree and dbadapter (`gno.land/pkg/gnoland/app.go:113-114`).
- rootmulti implements a new `store.CommitHolder`. While holding, `Commit`
  commits the stores, the metadata and the commit ID into the collector and
  makes it the `LastCommitID` without writing it; `WriteHeld` writes the
  collector as one synced batch, refreshes the query snapshot, and publishes
  the version on disk (`WrittenCommitID`), which `LoadVersion` also publishes.
  One write stays one batch: the latest-version key, the bptree roots and the
  fast-index stamp must agree on disk (`tm2/pkg/bptree/fast_index.go:238-256`).
- Queries treat as latest the version on disk. They read a snapshot of the disk
  taken after each write (`tm2/pkg/store/rootmulti/store.go:310-312, 413-447`),
  or the database itself on backends that take no snapshots (`425-431`), so in
  deferred mode they see the last flush point. `LastCommitID` stays the version
  in memory, which `Commit` counts from (`280-284`). BaseApp keeps the header
  of the last held commit and publishes it to queries (`lastBlockHeader`) in
  `WriteCommits`, with an atomic store, since that runs outside the local
  client's mutex; at `InitChain`, and at every `Commit` while not deferring, it
  publishes it as today (`baseapp.go:256-263, 366, 1032`). Store and custom
  queries default to the version on disk (today `LastBlockHeight`:
  `baseapp.go:496-498, 548-550`) and refuse a higher one; `abci_info` reports
  that version (today `cms.LastCommitID()`: `baseapp.go:329`); `Simulate` runs
  at the published header's height, which is that version, or at that version
  when the header has none (today `LastBlockHeight`: `helpers.go:83-86`). `abci_info` is also
  the handshake's source for the application's height
  (`consensus/replay.go:239`), so the version on disk is what was loaded. While
  holding, rootmulti refuses the live `.store` query that BaseApp falls back to
  (`baseapp.go:513-532`), which would answer from held writes and register a
  reader that makes pruning panic (`tm2/pkg/bptree/nodedb.go:532-540`,
  `tm2/pkg/store/bptree/store.go:144-147`). A fresh node, with nothing on disk
  until its first flush point, refuses every store and custom query and
  `Simulate` until then:
  custom queries would read an empty version 0 (`rootmulti/store.go:174-202`),
  and `Simulate`'s height-0 path would run on the live state
  (`tm2/pkg/sdk/helpers.go:83-100`). A restarted node answers at once from the
  version it loaded (`rootmulti/store.go:241-243`, `baseapp.go:209-219`).
  `CheckTx` still runs on the latest state in memory. The consensus params need
  nothing: BaseApp sets them only at load and in `InitChain` (`baseapp.go:206,
  348`).
- `NeedsFlushPoint` compares the halt height (set by governance or by the
  `halt_height` config) with the version committed in memory, not the one on
  disk, which stays below it inside a window. gno.land's
  `nodeParamsKeeper.WillSetParam`, which every write of `p:halt_height` and
  `p:halt_min_version` passes through today
  (`tm2/pkg/sdk/params/keeper.go:283-303`, `node_params.go:54-74`; `SetStruct`
  skips the hook, `keeper.go:225-244`, but nothing writes `node:p` with it),
  calls `BaseApp.RequireFlushPoint`, handed to the keeper where it is
  registered (`app.go:143`), once the value is validated and only in deliver
  mode (not in `CheckTx` or `Simulate`). A write in a transaction that later
  fails costs one extra flush point.

## Safeguard checklist

Everything that relies on what the per-block flushes protect, and how this
change keeps it (verified against the code cited):

| relies on it | today | with this change |
|---|---|---|
| restart from a crash | at most one block of replay | up to one window fetched again, or replayed from disk after a crash inside a flush point; the block store lowered when the flush-point record matches, then existing handshake cases only |
| block on disk before the app commits it | block store flushed first, per block | flush order store → state → app |
| results on disk before the app commits | `SaveABCIResponses` flush (#6085) | the app is never on disk ahead of the state |
| governance or config halt: halt height on disk | state flushed per block | flush point at the halt height, compared with the version in memory |
| upgraded binary refuses to start below the halt | the scheduled halt is in the committed state | flush point after the block that writes a halt param |
| upgraded binary started after a halt | starts from the halt height | same; after a crash inside the halt's flush point, start the previous binary once first |
| consensus: block stored before the WAL marker | unchanged | unchanged (block sync only) |
| switching to consensus | state = store | flush point first, then the same; a stop before or during that flush point does not start consensus |
| clean stop | block sync not waited for; gnoland closes the app database after the node stops; reactors stop after the event switch | the block-sync reactor stops first and waits for its routine, then the other reactors; held writes are dropped, as on a crash |
| `gnobr`, `gnogenesis` fork, older binaries | store ≤ state+1, app = state | start the node once after a crash; `gnobr` refuses a target above the state otherwise |
| a failed write (disk full) | per-block panics, partly | any flush error is fatal; nothing is written over a gap |
| peers, RPC, `CheckTx` reading during a flush | n/a | held writes are cleared only once on disk |
| an iterator over held writes | n/a | panics and is recorded; a delivered transaction, `BeginBlock` or `EndBlock` that ran one ends the process before the block's `Commit`; deferral is opt-in |
| tx event store (opt-in) | at most once | at most once; events fire after their flush point |
| RPC during block sync | sees each block, on disk | blocks and state: each block, not yet on disk (a crash takes back up to a window); app queries and events: each flush point |
| memory | per-block batches | a byte budget per database, counting every operation, under each backend's limit for one write |
| other backends | per-block flushes | one atomic batch per database per flush point, so out-of-order recovery (goleveldb) cannot split a window |

## Alternatives considered

- Removing single flushes (#6316's first version): each one protects something;
  see #6316's ADR.
- Writing downloaded blocks to the block store ahead of applying them: a block
  cannot be verified before the blocks below it are applied (the validators
  that sign block H are set by block H-2's results), and the block store serves
  peers any block it holds (`blockchain/reactor.go:195-208`) and replay trusts
  it.
- Signalling the halt or the flush through ABCI: `Commit` takes no request
  (`abci/types/application.go:22`), and a new field on a persisted response
  makes it unreadable to an older binary (amino rejects unknown fields:
  `tm2/pkg/amino/binary_decode.go:1081-1091`; `state/store.go:188-193` exits).
- A new handshake case that replays a block store more than one block ahead
  from disk: more code and tests than lowering the block store, for a crash
  inside one step of a flush point.
- Firing events per block and having the file event store skip what it holds
  after a restart: it needs the last record across file rotations and a guard
  against a store reused after a chain reset, and the first such guard turned
  the skipping off after every crash between flush points.
- Iterators that merge held writes: no current consumer, and #5431 declined
  merging for its complexity (`tm2/adr/pr5431_concurrent_queries.md:198-200`).
- Choosing deferral by store type: it cannot see a future iterator on the base
  store, whose iterators read the database directly
  (`tm2/pkg/store/dbadapter/store.go:60-66`); an option is the usual switch.
- A flush point on a clean stop: it would save fetching up to T seconds of work
  again, not worth another trigger and its tests.
- State sync (start from a verified snapshot instead of applying every block):
  a separate, larger project; nodes that apply every block still need this.

## Consequences

- During block sync no block waits for the disk; a flush point costs three
  flushes per window. Measured on a mainnet sync from genesis to height
  ~25,000 (light blocks; master and this change, two interleaved runs each, on
  one box with the same peers): 8.4-9.8 ms per block on master, 3.2-3.4 ms
  with this change, 2.6-2.9× faster, close to the ~2.6 ms measured with every
  flush disabled. Over the whole chain, where a transaction-heavy block costs
  10-15 ms of CPU, the estimate is ~1.4-1.5× against master, and ~1.8× on top
  of #6308 and #6315, which remove other per-block costs. With the opt-in file tx event store on, each of its results is still
  synced on its own (`tm2/pkg/events/subscribe.go:23-26`,
  `state/eventstore/file/file.go:86-89`), now at flush points.
- A crash or a stop during block sync costs up to one window of work again. RPC
  (`status`, `block`, `block_results`, `tx`) and peers see held blocks before
  they are on disk, so after a crash the node's height drops by up to a window,
  and those heights fail until block sync fetches them again, with the same
  data.
- During block sync, events arrive at flush points: the file event store, the
  node's first-block signal (`node/node.go:520-535`) and `broadcast_tx_commit`
  (`rpc/core/mempool.go:162`) see them up to T late. tm2 has no event
  subscription over RPC.
- Memory: up to the byte budgets of held writes (a few times that at peak: map
  keys, the batch copy), on top of the blocks the pool already holds.
- Application queries during block sync answer from the last flush point. The
  VM's code cache is shared with the live state
  (`gnovm/pkg/gnolang/store.go:271`), so a query may run code from a newer
  block over objects from the flush point; this already happens today for one
  block and for queries at older heights, and now for up to a window.
- Two stop hazards of today go away: gnoland closing the application database
  under a block in flight, and an event fired during a stop panicking on a
  closed channel.

## Tests

Every rule above has a test that fails without it, but for the node's stop
order, which needs two nodes and a stop inside a flush point to exercise, and
was checked by review.

- **Crash safety** (`bft/consensus/replay_flush_point_test.go`):
  `TestHandshakeAfterFlushPointPowerCut` syncs eight blocks with deferred
  writes in windows of four, over disks that log every write and batch, then
  cuts the power at every step of both flush points. At each cut it restarts on
  every prefix of what can survive (the writes the last flush covered, then
  each later batch whole, in order; the app at any height it last wrote or
  below), through the real handshake, and syncs on with deferred writes. Each
  restart must end with the uncut run's state, every key of both databases but
  the flush-point record, and the app's height and hash, and must fire no
  block's events a second time; in the uncut run, no block's events fire
  before its commit is on disk. The
  chain changes its validator set in the block a restart replays after a crash
  between a flush point's first two writes, changes its consensus params, has
  an empty block, and starts at height 1 or 5. `TestLowerAfterFlushPointCrash`
  checks each condition of the rewind.
- **Block sync** (`bft/blockchain/deferred_test.go`): two nodes over p2p.
  The writes are held, and on disk with nothing held at the switch to
  consensus (`TestBlockSyncDefersWrites`); an app that refuses is written per
  block (`TestBlockSyncWithoutDeferral`); a stop during the switch's flush
  point waits for it and starts no consensus
  (`TestBlockSyncStopDuringFlushPointDoesNotSwitch`); a stop while behind
  writes nothing (`TestBlockSyncStopBeforeCaughtUp`); a window whose peer
  leaves is written on the ticker (`TestBlockSyncFlushesAStalledWindow`);
  `OnStop` with fast sync off returns. Each flush-point trigger, each step's
  writes and record, and a failed write at each step panicking before the next
  (`TestDeferredWritesNeedsFlushPoint`, `TestDeferredWritesFlushPoint`,
  `TestFlushPointWriteErrorEndsTheNode`).
- **The node** (`bft/node/deferred_test.go`): a node whose app defers holds
  its block store's and its state's databases during block sync, and stops
  while behind.
- **Databases** (`db/collecting_test.go`, `db/holding_test.go`): pass-through
  until held; held writes readable until written; the batch holds only the
  last op on each key and is synced; ops kept after a failed write; an empty
  value stays non-nil; the byte count; iterators refused and recorded; holding
  switched on or off with writes collected panics. `bft/store/flushpoint_test.go`: `LowerHeight`,
  the record, the height record's write error.
  `bft/state/held_responses_test.go`: an empty block's held results load.
- **Application** (`store/rootmulti/holding_test.go`,
  `sdk/deferred_test.go`): commits held and written as one batch; the version
  on disk at load and after each write; what a store collected at load (a
  real fast-index rebuild) written before holding; database iterators
  refused; the live query refused. BaseApp defers only with the option; Info, queries,
  `Simulate` and the header answer from disk, and a restart is at the written
  version; a fresh node refuses queries and `Simulate`; the halt height in
  memory and a required flush point; an iterator in a delivered transaction,
  `BeginBlock` or `EndBlock` ends the block though a recover swallowed it,
  while one in `CheckTx` only fails its transaction and does not end the next
  call; `Close` writes nothing.
- **A real crash** (gnoland-1 on a devbox): the binary synced mainnet and was
  killed with `kill -9` mid-sync five times. One kill landed between a flush
  point's first two writes (block store at 34,751, state at 34,516): the
  restart lowered the block store to 34,517 and synced on. The chain's
  governance halt at 36,300 was on disk when the node halted: the next start
  found the application, the block store and the state at 36,300.
- **gno.land** (`gnoland/app_test.go`): `TestAppDefersCommits` runs real blocks
  with commits held, deploying a realm and calling it to write the halt
  height, which asks for a flush point that puts the halt height on disk;
  `TestAppDefersCommitsAfterRestart` starts and stops deferring on a restarted
  app with no block between, as at a restart at the tip;
  `TestNodeParamsKeeperRequiresFlushPoint` covers the hook's modes, keys and an
  invalid value. **gnobr**: `TestCheckRollbackable` covers each shape and
  target.

A mutation run breaks one rule at a time (73 mutants: each write's order, the
record, the rewind's conditions, each trigger, the stop and switch gates, the
event holding, each query's version, each refusal and record, the writes
collected at load, the option, the gno.land wiring and gnobr's check) and
checks that a test fails on each: every one does.

## AI assistance

Designed and implemented with AI assistance: a code investigation of the paths
cited, seven rounds of adversarial review of the design and seven of the
code, the review of #6316, and the mutation run above. The human author
reviews and owns it.
