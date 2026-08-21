# PR6082: Parallel Query Connection

Follow-up to [PR5431](./pr5431_concurrent_queries.md).

## Context

PR5431 gave the query connection its own mutex
(`proxy.localClientCreator.queryMtx`, handed to `NewReadOnlyABCIClient`), so
`.app/simulate` no longer blocks `BeginBlock`/`DeliverTx`/`EndBlock`/`Commit`.
That removed the zero-cost DoS against block production, which was the goal.

It did not make queries parallel. `localClient.QuerySync` still held `queryMtx`
for the whole `Application.Query` call, so queries were serialised one-at-a-time
among themselves. A node under simulate load still processes those simulates
strictly sequentially, on one core, regardless of how many are in flight.

The machinery PR5431 added is most of what parallel queries need:

- immutable DB snapshots pinned per query (`MultiImmutableCacheWrapWithVersion`),
- `SyncGoMap` for the VM's `cacheNodes`,
- the atomic last-block header (`getLastBlockHeader`),
- per-tx forked allocator and caches.

The rest of what they need did not exist, and the first revision of this PR
assumed it did. Two paths reachable from an ordinary RPC query raced once the
lock came off, both found by review and both reproduced with negative controls:

1. **The shared VM type graph.** `newGnoTransactionStore` forks `cacheObjects`,
   `cacheTypes` and the allocator per call, but keeps
   `cacheNodes: txlog.Wrap(ds.cacheNodes)`, so every query's store fork reaches
   the same `*PackageNode` graph and the same `Type` objects hanging off it.
   Those objects memoize on first use with an unsynchronised check-then-set:
   `FuncType.bound`, `FuncType.typeid`, `Declared`/`StructType.pkgID`, the
   effective field and method counts, `StructType.comparable`,
   `DeclaredType.methodIndex`, and `StaticBlock.nameIndex`. Two concurrent
   `vm/qeval` calls fill them together. `FuncType.BoundType` is the sharpest:
   it publishes `ft.bound` as a composite literal, so a second goroutine can
   follow that pointer and read `Params`/`Results` while the first is still
   writing them — and `TypeID`, derived from exactly those fields, is what
   `InterfaceType.VerifyImplementedBy` compares to decide whether a type
   satisfies an interface.

   [#5811](https://github.com/gnolang/gno/pull/5811) (`sealUverseTypes`) closed
   this for the process-global uverse singletons only. Its own comment states
   the assumption this PR invalidates: "Per-store types are unaffected (each is
   preprocessed by a single goroutine)."

2. **Simulate before the first block.** When `getLastBlockHeader()` reported a
   height below 1, `BaseApp.Simulate` fell back to `getContextForTx`, which
   copies `app.checkState.ctx`. `CacheContext()` forks the store but not the
   gas-meter pointer, so every concurrent simulate charged into one
   `infiniteGasMeter`, and read through one set of mutable cache stores.

   The window is reachable, which the first revision of this ADR denied. The
   node starts its RPC listeners (`node.go`) before the P2P switch and the
   consensus reactor, so the query endpoint is live before block 1 can exist;
   `gnoland start -x-early-start` widens it deliberately; and a node configured
   with `CreateEmptyBlocks=false` — both `contribs/gnodev` and
   `gno.land/pkg/integration` set it — idles in the window until a transaction
   arrives. `Commit()` republishes the header it was handed, and after
   `InitChain` that is `initHeader`, built with no `Height`, so the genesis
   commit stores a header of height 0 and the fallback stayed live until the
   first real `BeginBlock`.

PR5431's "Alternatives Considered" rejected an RWMutex on `localClient` because
it "does not fix the underlying data race on `checkState`". That objection was
correct at the time and no longer holds: the `checkState` race is fixed here, by
not reading `checkState` from the query path at all.

## Decision

Three changes: replace the query connection's lock with a bound, seal the shared
type graph, and take `checkState` off the simulate path.

### 1. The query connection takes a bound, not a lock

- `localClient.mtx` becomes a `sync.Locker` rather than a `*sync.Mutex`
  (`tm2/pkg/bft/abci/client/local_client.go`). Every method body is unchanged,
  and every one already `defer`s its `Unlock`, which is what makes a non-mutex
  Locker safe: a limiter gets its slot back even when the application panics.
- `localClientCreator.queryMtx` becomes a `queryLimiter`
  (`tm2/pkg/bft/proxy/client.go`): a `sync.Locker` backed by a buffered channel
  admitting `maxConcurrentQueries` callers at once, which is `GOMAXPROCS`. The
  consensus and mempool connections keep their real `sync.Mutex`.

Making the bound a `Locker` rather than a semaphore threaded through the client
is what keeps the diff at the seam PR5431 already established. It also makes the
degenerate value meaningful for free: a limiter of 1 *is* the mutex it replaced.

There is a bound at all because nothing else caps aggregate query work. Each
query installs its own budget — `maxAllocQuery` is 1.5 GB and `maxGasQuery` is
3e9, installed fresh per call, and `withQueryEvalMachine` takes two allocators —
and those cap one query, never the sum. Before this PR the sum *was* the
per-call cap, because the lock admitted one caller. Unbounded, the only
remaining ceiling would be the RPC listener's `MaxOpenConnections`, which
defaults to 900 and is nowhere near a machine's capacity to run 900 simultaneous
simulates.

`GOMAXPROCS` is not exposed as a setting. Query work is CPU-bound in the VM, so
the machine's parallelism is the quantity worth sizing from, and a knob with no
metric behind it (see Consequences) would be tuned blind. Making it configurable
is cheap to add later, once there is something to tune against.

An RWMutex was considered and rejected: the query connection has no writer at
all. `SetResponseCallback` is the only method that mutates `localClient` state,
the `appconn.Query` interface does not expose it, and its only two callers
(`state/execution.go`, `mempool/clist_mempool.go`) are on the other two
connections. A write lock that is never taken makes an RWMutex a slower plain
mutex, not a concurrency improvement.

Cutting per-connection rather than per-method is deliberate. Removing the lock
from `QuerySync` alone would unlock that method on the consensus and mempool
clients too — no live caller today, but `abcicli.Client` exposes `QuerySync`
publicly, so a future caller holding the consensus client would silently run
unlocked against `DeliverTx`/`Commit`. It would also leave `InfoSync` locked on
the query connection, where the lock then protects nothing.

### 2. The shared type graph is sealed on publication

`gnovm/pkg/gnolang/seal.go` generalises the walker that `sealUverseTypes`
already used, and fills every lazily-memoized cache on a type graph and the
block nodes reaching it — now including `DeclaredType.methodIndex` and
`StaticBlock.nameIndex`, which no uverse singleton is wide enough to reach but a
realm readily is.

It runs at the two points where a block node stops being private:

- `transactionStore.Write()`, over the txlog's pending entries only (a new
  `txlog.MapCommitter.Dirty`), immediately *before* `Commit` copies them into
  the parent map;
- `defaultStore.SetBlockNode`, when the store is not a transaction fork —
  genesis load and the mem-package re-run on node start. `GetBlockNodeSafe`'s
  backend-load path now routes through `SetBlockNode` rather than writing to
  `cacheNodes` itself, so publication into the shared map has one door and the
  seal sits on it.

Sealing is pure cache warming: every value it writes is the value the lazy path
would have computed on first touch. It charges no gas, touches no allocator, and
changes no struct layout or amino encoding, so it cannot move consensus. The
cost is proportional to the types a transaction just published, which that
transaction already paid to preprocess, and it is paid once on the write path
rather than repeatedly on the read path.

Alternatives considered:

- **Make the memo fields atomic or `sync.Once`.** Rejected. These are the
  hottest structs in the VM; `sync.Once` embeds a mutex, which makes
  `copyTypeWithRefs` a `copylocks` violation, and widening `FuncType`,
  `StructType` and `DeclaredType` risks the amino encoding and the allocator
  accounting for no benefit the write-once path does not already give.
- **Give query stores a private `cacheNodes`.** Rejected: it would re-load and
  re-preprocess every package on every query, trading a race for a much larger
  performance regression than the one this PR set out to remove.
- **A global lock around lazy fills.** Rejected: it puts lock traffic on the
  VM's hottest read path to protect a value that only ever gets written once.

### 3. Simulate never reads `checkState` in the reachable window

`BaseApp.Simulate` now resolves the version to simulate against as: the header
height when a real block has been committed (unchanged), and otherwise the
store's own last committed version. Both then take the same immutable snapshot,
so the pre-first-block window is served by the same isolated machinery as every
other query rather than from mutable shared state.

One case remains without a snapshot: nothing has been committed at all, between
`InitChain` and the genesis commit, where the genesis state exists only in the
uncommitted `checkState`. The consensus handshake performs both before the node
starts its RPC listeners, so no external caller is there; the in-process callers
that are (tests, tooling driving `BaseApp` directly) get
`BaseApp.preCommitSimulateMu` rather than a claim of unreachability. Serialising
a handful of startup calls costs nothing measurable, and it means the branch is
correct on its own terms instead of on an argument about who can reach it.


## Consequences

**The invariant is load-bearing, and now has enforcement.** "Everything
reachable from `Application.Query` and `Application.Info` is goroutine-safe" was
previously an unused property. It is now a correctness requirement. The
precondition is stated on both `ClientCreator` interface declarations
(`proxy` and `appconn`), not only on the concrete constructor, so an alternative
implementer reads it where they will act on it.

**`SetResponseCallback` is unsynchronised on the query connection.** Its write
to `localClient.Callback` would race with concurrent `completeRequest` reads.
Only `EchoSync`, `InfoSync` and `QuerySync` are reachable through the
`appconn.Query` wrapper, but `NewReadOnlyABCIClient` is public and returns the
raw `abcicli.Client`, so every method on it lost its lock, the mutating ones
included. Nothing calls them there — documented on `NewReadOnlyABCIClient`.

**`NewLocalClient` no longer accepts nil.** The old nil guard substituted a
fresh mutex for an untyped nil but not for a typed nil `*sync.Mutex`, so it
turned one kind of programming error into a silent success and left the other
panicking later. No in-tree caller passed nil; the guard and the paragraph
explaining its asymmetry are both gone, and any nil now fails the same way.

**Query work competes for cores with block execution.** Bounded, but real, and
the bound is not currently adjustable: a validator that also serves heavy RPC
can only lower it by lowering `GOMAXPROCS` for the whole process.

**Sealing adds work to the transaction commit path.** Proportional to the types
each transaction publishes, and unmetered. A package with a large type graph
pays a walk over it once, at deploy, having already paid gas to preprocess it.

**Not addressed here.** Two items the review raised are deliberately left:

- There is no in-flight query counter and no query-latency histogram, so the
  quantity this PR bounds is not yet observable, which is also why the bound is
  fixed at `GOMAXPROCS` rather than exposed as a setting.
- Every custom query and simulate still builds a fresh immutable multistore per
  call (`MultiImmutableCacheWrapWithVersion` → `immutableAtVersion` →
  `LoadVersion`). The PR 6018 review measured that cost and signed it off on the
  grounds that "the query connection serializes on one mutex, so the cost caps
  throughput on every query path". That premise is exactly what this PR removes:
  serialised, it was a throughput ceiling; bounded-parallel, N concurrent
  queries each pay it at once. The fix that review named — one cached immutable
  multistore per snapshot generation — is now a follow-up rather than an
  optimisation, with the `GOMAXPROCS` bound holding the line in the meantime.

## Validation

`gno.land/pkg/gnoland/app_parallel_query_test.go`:

- `TestParallelQueries_NWaySimulate` runs N=8 goroutines simulating signed bank
  sends through the read-only connection, in two phases: with a consensus block
  loop committing underneath, and against frozen committed state. It asserts
  **overlap actually happened**, via a gauge counting calls inside
  `Application.Query` — the gauge is a probe application wrapped *beneath* the
  ABCI client, because that is where the lock is held, and a gauge in the test's
  own goroutines would count callers merely blocked on it — **no data race**
  under `-race`, and **determinism**, each querier's `GasUsed` matching a serial
  baseline.
- `TestParallelQueries_PreFirstBlockSimulate` commits genesis and stops there,
  which is the state a `CreateEmptyBlocks=false` node serves from, then fires
  concurrent simulates into that window.

`gno.land/pkg/gnoland/app_parallel_vmquery_test.go`
(`TestParallelVMQueries`) covers what the simulate tests cannot: `.app/simulate`
routes to the bank handler and never enters the GnoVM, while `vm/qeval` and
`vm/qrender` go through `handleQueryCustom` into `gno.Machine`, preprocess and
the shared type graph — the machinery this PR's invariant actually names. Its
realm is shaped to reach every lazily-memoized cache in one query: a
concrete-to-interface conversion preprocessed fresh per call, a type with more
than `methodIndexThreshold` methods, more than `nameIndexThreshold` top-level
names, a struct used as a map key, and an anonymous struct built inside a body.

`tm2/pkg/bft/proxy/client_test.go` pins the connection contract structurally, in
a package that previously had no tests at all, with no database, no VM and no
genesis: the read-only connection admits concurrent callers (reintroducing a
per-call mutex deadlocks the test rather than slowing it), the bound is
enforced, a bound of 1 serialises, the mutating connections still serialise, and
a query in flight does not block a consensus call.

**Ordering is part of these tests, not an accident of how they were written.**
Every cache at issue is filled on first touch, so a serial baseline taken
*before* the concurrent phase warms the whole shared graph single-threaded and
leaves the concurrent phase reading already-populated fields — the test then
passes against racy code. Both the VM test and phase 2 of the simulate test run
the concurrent phase first, against a cold graph, and take the serial baseline
afterwards as the value oracle. The comments say so at both sites.

**`-race` now runs in CI.** It previously could not: the test workflow uses
`-covermode=set`, which the toolchain refuses to combine with `-race`, and no
workflow passed the flag. `.github/workflows/ci-race.yml` runs the concurrency
packages under `-race -covermode=atomic`. It is deliberately narrow rather than
`./...`, which would cost far more than the signal is worth.

Negative controls, all confirmed: restoring a real mutex on the query connection
fails the overlap assertion with `peak in-flight = 1`; reverting the seal makes
`TestParallelVMQueries` fail under `-race` on `FuncType.BoundType`,
`FuncType.TypeID` and `DeclaredType.GetPkgID`; reverting the `Simulate` change
makes `TestParallelQueries_PreFirstBlockSimulate` fail on a single shared
`infiniteGasMeter`.
