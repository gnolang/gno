# PRxxxx: Deferred writes during block sync, with ordered flush points

## Status

Proposed (design; not implemented). Citations are to origin/master `21f9abd`.

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

This design keeps every guarantee and removes the per-block waits during block
sync by changing when things are written, not what is protected: the three
databases hold a block's writes in memory, and a flush point writes many
blocks' writes in an order that leaves them consistent after a crash at any
instant.

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
  (`tm2/pkg/db/collecting.go:155-175`); in deferred mode it is not drained at
  `Commit`. The bptree store reads its nodes, orphans, roots, versions and fast
  index through point reads (`tm2/pkg/bptree/nodedb.go:139-193, 304-336,
  369-389, 425-438`, `fast_index.go:91-105`).
- **Events** of each block are held, and fired after the flush point that
  writes the block (below).

The two tendermint overlays reuse `CollectingDB`, which rootmulti already uses,
with a switch that only they have. Off, every call goes straight to the
database with its own sync, as today; the handshake and consensus rely on that,
for example on the synced results of #6085 (`state/store.go:214-217`). On, in
deferred mode only, writes are held. rootmulti's collector always collects, as
today: its one batch per write is what makes an application commit
all-or-nothing (`rootmulti/store.go:287-308`). The overlays wrap the databases
where the node opens them, so every reader, the RPC environment included
(`node/node.go:862-866`), reads through them. `BatchCollector.set` must keep an
empty value non-nil, as its batches already do (`collecting.go:113-115` against
`229-231`, `util.go:7-11`): an empty block's results encode to no bytes, and a
held nil reads as missing (`state/store.go:183-185`). The fix also reaches
consensus mode, where it changes nothing, since rootmulti reads nothing back
from the collector before writing it out.

`CollectingDB`'s iterators read only the real database
(`collecting.go:177-188`), so while writes are held across blocks (deferred
mode) they panic rather than miss them. Nothing iterates these databases inside
a block or in `CheckTx` today: the bptree store's own iterators walk its tree
through point reads, its database iterators run only at load and state import
(`fast_index.go:155`, `nodedb.go:479`; `nodedb.go:449` has no caller outside
tests), `CheckTx` runs no messages (`baseapp.go:718`), and queries read stores
built without the collector (`rootmulti/store.go:413-447`). But `runTx` turns a
panic into a failed transaction (`baseapp.go:828-859`), and so does the VM for
its messages (`gno.land/pkg/sdk/vm/keeper.go:1435-1469`); here that would be a
result no other node has, which the next flush point could write before the
next block's header check stops the node (`state/validation.go:70-87`). So the
refused iterator is also recorded on its collector: the tendermint overlays
refuse while their switch is on, and rootmulti's collector while a flag of its
own, set by `DeferCommits`, is on. In deferred mode BaseApp clears the
application's record, through a method the multistore offers, before each
delivered transaction, `BeginBlock` and `EndBlock`, and panics after each one
if it is set (for a transaction, in `DeliverTx` after `runTx` returns,
`baseapp.go:648-655`, outside every recover). The block then fails before its
`Commit` and the process ends, since neither the local client nor block sync
recovers it (`tm2/pkg/bft/abci/client/local_client.go:91-97`,
`blockchain/reactor.go:373-377`). `CheckTx` never overlaps those calls (the
local client's mutex is shared: `tm2/pkg/bft/proxy/client.go:43`), so an
iterator there only rejects its transaction. A node that hits this stops at
that block on every start, with nothing of it on disk, until the iterator is
fixed; meanwhile `fast_sync = false` lets it catch up through consensus instead
(`tm2/pkg/bft/config/config.go:286`, `node/node.go:559`).

In consensus mode nothing changes: every write goes to disk with its flushes
exactly as today.

### Flush points

A flush point writes, in this order, each as one all-or-nothing write followed
by a flush:

1. the block store,
2. the state database,
3. the application,

then fires the held events of its blocks, in block order.

Each overlay is cleared only after its batch is on disk: `Drain` stops clearing
(`collecting.go:63-64`) and `Reset` runs after the write. Until then readers on
other goroutines (peers, RPC, `CheckTx`) keep finding held records in the
overlay, and after it on disk. Nothing more is needed, because no writer runs
during a flush point: it runs on the block-sync goroutine, between blocks. The
application's write is a direct call, not an ABCI message, so it runs outside
the local client's mutex, which is safe for the same reason. A failed write is
fatal: the node panics and is never recovered, because continuing would write
the next window over a gap.

A flush point is taken between blocks, and only there:

- every T seconds, or when the held writes of an overlay exceed a byte budget
  B. The budget counts every operation, since the collector keeps each
  overwrite (`collecting.go:116-119`), and stays well under each backend's
  limit for one write (pebble panics near 4 GiB). T and B are constants beside
  `trySyncIntervalMS` (`blockchain/reactor.go:22-31`);
- when the application needs one (`NeedsFlushPoint`, read after each
  `ApplyBlock`): it holds more than B; its halt height is reached, so the halt
  height is on disk before the `BeginBlock` that panics; or the block wrote a
  halt param, so a scheduled halt is on disk from the block that scheduled it
  (the upgraded binary's startup check reads it from the committed state:
  `node_params.go:135-182`);
- before switching to consensus (`blockchain/reactor.go:294-310`), in this
  order: caught up, the stop check below, the flush point, the stop check
  again, deferral off, `pool.Stop`, then `switchToConsensusFn`. So no consensus
  block, whose WAL marker is synced (`consensus/state.go:1369-1386`), lands in
  an overlay, and a stop that arrives during the flush point does not start
  consensus.

Never after a failed block: a block that panics in the application's `Commit`
can leave part of its writes in the collector, and the base store keeps no
versions to fall back to (`tm2/pkg/store/dbadapter/store.go:91-97, 128-130`). A
failed block ends the process as today, and the window is lost with it.

The application never writes on its own, not in `Commit` and not in `Close`: an
application on disk ahead of tendermint's records is the one state the restart
cannot recover from (#6085). A stop drops the held writes, as a crash does; the
last flush point is consistent without them, and block sync fetches the window
again. The block-sync reactor's `OnStop` waits for its routine, if `OnStart`
started one (only with fast sync on: `reactor.go:146-156`), so that gnoland
does not close the application database under a flush point
(`gno.land/cmd/gnoland/start.go:322-330`). Its main loop gains a case for the
pool stopping (`OnStop` stops the pool first: `reactor.go:160`; today only the
request goroutine watches it: `reactor.go:266`), since `Quit` closes only after
`OnStop` returns (`tm2/pkg/service/service.go:175-176`). The loop checks it
before and after the flush point that precedes the switch to consensus, whose
caught-up test ignores the pool's state (`pool.go:165-184`); the routine stops
the pool itself only after the second check (today at `reactor.go:300`). The
p2p switch stops reactors in no fixed order (`tm2/pkg/p2p/switch.go:240`), and
a consensus reactor stopped first leaves its never-started consensus state
startable (`tm2/pkg/service/service.go:167-171`), so the node stops the
block-sync reactor first (the p2p switch then skips it rather than log its
`ErrAlreadyStopped`: `tm2/pkg/p2p/switch.go:240-243`), then the p2p switch with
the rest, then the RPC environment, the event switch and the event store, where
today the event switch and the event store stop first (`node/node.go:821-837`):
every node registers the event store's listener (`node.go:262-271`), and a
flush point that completes during a stop fires its window's events, which, once
the event switch has stopped, panic on that listener's closed channel
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
flush point it belongs to (F and F', under a key of its own, which older
binaries ignore). When that record says F to F', the state is at F, the block
store at F' and the application at most at F, the node lowers the block store's
height to F+1 before replaying, on disk and in memory (a new `BlockStore`
method: the handshake reads the in-memory height, `replay.go:286`; the height
record is one synced write, `tm2/pkg/bft/store/store.go:241-255`, whose error
`Save` ignores today and must check). Any other shape, such as a partial
restore, panics as today and leaves the directory as it is. The blocks above
F+1 were verified, and block sync fetches them again, as after a crash a moment
earlier. Every shape a crash can leave is then one the handshake handles today,
including a crash during its own one-block replay (the mock app:
`replay.go:427-436`) and a chain with InitialHeight > 1 (`replay.go:464`).

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
running them. `gnobr` gains a check, made before it trims the block store, that
refuses when `state.LastBlockHeight+1 < blockStore.Height()`: on such a
directory it would write a later height's hashes into an earlier state
(`contribs/gnobr/main.go:83-136`). After a crash inside the flush point at the
halt height, the application on disk is below the halt, so the upgraded binary
refuses (`node_params.go:172-179`): start the previous binary once, then
upgrade.

### Interface

tm2 runs the application in process only
(`tm2/pkg/bft/proxy/client.go:130-151`), so the application side is an optional
Go interface that `NewNode` asserts on `config.LocalApp`, as gnoland already
asserts it to call `SetHaltHeight` (`gno.land/cmd/gnoland/start.go:288-293`):

```go
// DeferredCommitter is an application that can hold its commits in memory
// across blocks and write them when tendermint says so.
type DeferredCommitter interface {
	// DeferCommits, with maxHeldBytes > 0, makes Commit hold its writes in
	// memory; with 0, Commit writes them, as today. It reports whether the
	// application defers.
	DeferCommits(maxHeldBytes int64) bool
	// WriteCommits writes every held commit as one all-or-nothing write.
	WriteCommits() error
	// NeedsFlushPoint reports that a flush point must follow the current
	// block: the halt height is reached, a halt param was written, or more
	// than maxHeldBytes is held.
	NeedsFlushPoint() bool
}
```

Tendermint owns the policy (T and B); the application reports only what
correctness or its budget requires. If the application does not implement the
interface, or `DeferCommits` returns false, nothing is deferred: deferring only
tendermint's side would let the application's commits reach disk ahead of the
block store, which the restart cannot recover from.

### Application changes (tm2/pkg/sdk, tm2/pkg/store, gno.land)

- Deferral is a BaseApp option, off by default; gno.land turns it on. Other tm2
  applications keep today's behavior: the IAVL store iterates the database
  inside a block (its fast iterator: `tm2/pkg/iavl/mutable_tree.go:240-249`,
  `tm2/pkg/iavl/nodedb.go:1029-1049`; its pruning's version lookups:
  `nodedb.go:755, 804, 846`), which would hit the panic above. gno.land mounts
  only bptree and dbadapter (`gno.land/pkg/gnoland/app.go:113-114`).
- `rootmulti.Commit` splits into holding (commit the stores, the metadata and
  the commit ID in memory) and writing (drain into one batch, `WriteSync`,
  reset, refresh the query snapshot). One write stays one batch: the
  latest-version key, the bptree roots and the fast-index stamp must agree on
  disk (`tm2/pkg/bptree/fast_index.go:238-256`).
- Queries treat as latest the version on disk. They read a snapshot of the disk
  taken after each write (`tm2/pkg/store/rootmulti/store.go:310-312, 413-447`),
  or the database itself on backends that take no snapshots (`425-431`), so in
  deferred mode they see the last flush point. rootmulti keeps `LastCommitID`
  as the version in memory, which `Commit` counts from (`280-284`), and
  publishes a second commit ID, the version on disk, at every write and at
  load. BaseApp publishes the header it gives queries (`lastBlockHeader`) with
  it, and at `InitChain` as today (`baseapp.go:262, 366`). Today
  `setCheckState` does both, at every `Commit` (`baseapp.go:256-263, 1032`); in
  deferred mode `Commit` still resets the check state but no longer publishes
  the header, which the write publishes with an atomic store, since it runs
  outside the local client's mutex. Queries, `Simulate` and `abci_info` default
  to the version on disk (today `LastBlockHeight`: `baseapp.go:496-498,
  548-550`) and refuse a higher one. `abci_info` is also the handshake's source
  for the application's height (`consensus/replay.go:239`), so the version on
  disk is what was loaded. In deferred mode the `.store` query never falls back
  to the live store (`baseapp.go:513-532`), which would answer from held writes
  and register a reader that makes pruning panic
  (`tm2/pkg/bptree/nodedb.go:532-540`,
  `tm2/pkg/store/bptree/store.go:144-147`). A fresh node, with nothing on disk
  until its first flush point, refuses every query and `Simulate` until then:
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
  calls a new `BaseApp.RequireFlushPoint`, handed to the keeper where it is
  registered (`app.go:143`), in deliver mode (not in `CheckTx` or `Simulate`).
  A write in a transaction that later fails costs one extra flush point.

## Safeguard checklist

Everything that relies on what the per-block flushes protect, and how this
design keeps it (verified against the code cited):

| relies on it | today | with this design |
|---|---|---|
| restart from a crash | at most one block of replay | up to one window fetched again, or replayed from disk after a crash inside a flush point; the block store lowered when the flush-point record matches, then existing handshake cases only |
| block on disk before the app commits it | block store flushed first, per block | flush order store → state → app |
| results on disk before the app commits | `SaveABCIResponses` flush (#6085) | the app is never on disk ahead of the state |
| governance or config halt: halt height on disk | state flushed per block | flush point at the halt height, compared with the version in memory |
| upgraded binary refuses to start below the halt | the scheduled halt is in the committed state | flush point after the block that writes a halt param |
| upgraded binary started after a halt | starts from the halt height | same; after a crash inside the halt's flush point, start the previous binary once first |
| consensus: block stored before the WAL marker | unchanged | unchanged (block sync only) |
| switching to consensus | state = store | flush point first, then the same; a stop during that flush point does not start consensus |
| clean stop | block sync not waited for; gnoland closes the app database after the node stops; reactors stop after the event switch | the block-sync reactor stops first and waits for its routine, then the other reactors; held writes are dropped, as on a crash |
| `gnobr`, `gnogenesis` fork, older binaries | store ≤ state+1, app = state | start the node once after a crash; `gnobr` refuses otherwise |
| a failed write (disk full) | per-block panics, partly | any flush error is fatal; nothing is written over a gap |
| peers, RPC, `CheckTx` reading during a flush | n/a | held writes are cleared only once on disk |
| an iterator over held writes | n/a | panics and is recorded; a delivered transaction, `BeginBlock` or `EndBlock` that ran one ends the process before the block's `Commit`; deferral is opt-in |
| tx event store (opt-in) | at most once | at most once; events fire after their flush point |
| RPC during block sync | sees each block, on disk | blocks and state: each block, not yet on disk (a crash takes back up to a window); app queries and events: each flush point |
| memory | per-block batches | a byte budget per overlay, counting every operation, under each backend's limit for one write |
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
  flushes per window. Expected (estimate from the measurements above): ~2.6 ms
  instead of ~7.2 ms per light block; over the whole chain ~1.4-1.5× against
  master, and ~1.8× on top of #6308 and #6315, which remove other per-block
  costs. With the
  opt-in file tx event store on, each of its results is still synced on its own
  (`tm2/pkg/events/subscribe.go:23-26`, `state/eventstore/file/file.go:86-89`),
  now at flush points.
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

## Tests (plan)

- The power-cut test from #6316's review, extended: overlays that hold writes,
  flush points, a crash after every write and inside every flush point, the
  block store lowered after a crash between steps 1 and 2 (including
  InitialHeight > 1 and a crash during the one-block replay), the application
  holding commits, halts by governance and by config (including one inside a
  window, and a halt param written inside a window), stops, the switch to
  consensus, and `gnobr` refusing a directory left by a crash; with mutants for
  each rule (the app written first, the state before the block store, two
  batches for one application write, the application writing on `Close` or at
  its halt by itself, a flush point after a failed block, held writes cleared
  before they are on disk, the halt compared with the height on disk, the
  switch to consensus before the flush point or with deferral still on, the
  block store lowered without a matching flush-point record).
- In deferred mode, an iterator in a delivered transaction, through a VM
  message (`MsgCall`, `MsgRun`) included, or in `BeginBlock` or `EndBlock`,
  ends the process before the block's `Commit`; one in the ante handler during
  `CheckTx` rejects only its transaction, including a `CheckTx` that runs
  between two delivered transactions of a block.
- `Simulate` inside a window answers from the last flush point.
- A stop during the flush point before the switch to consensus does not start
  consensus, whichever reactor the p2p switch stops first; the switch itself
  still happens when no stop comes.
- RPC reads held blocks and results, an empty block's included, through the
  node's handles.
- Queries answer from the last flush point; a fresh node refuses every query
  and `Simulate` until its first flush point; a restarted node answers at once;
  in consensus mode, `Simulate` right after `InitChain` works as today.
- Events fire only after step 3, in block order; no result is stored twice in
  the file tx event store across crashes and restarts.
- `RequireFlushPoint` is not called from `CheckTx` or `Simulate`.
- `OnStop` returns with fast sync off and during a flush point.
- A flush error stops the node; a restart then recovers from the previous flush
  point.

## AI assistance

Designed with AI assistance from a code investigation of the paths cited, seven
rounds of adversarial review, and the review of #6316. The human author reviews
and owns it.
