# PRXXXX: Lock-Free Query Connection (Parallel Queries)

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

The machinery PR5431 added is exactly what parallel queries need:

- immutable DB snapshots pinned per query (`MultiImmutableCacheWrapWithVersion`),
- `SyncGoMap` for the VM's `cacheNodes`,
- the atomic last-block header (`getLastBlockHeader`),
- per-tx forked allocator and caches.

The shared process-wide VM type state that parallelism would otherwise expose —
the uverse TypeID / in-place `Methods` sort race — was closed separately by
[#5811](https://github.com/gnolang/gno/pull/5811) (`sealUverseTypes`), which
also hardened a batch of shared VM globals. What #5811 did not cover is the
production path where concurrent queries share the node's single
`gnoStore`/`defaultStore`; that is precisely what PR5431's snapshot work
isolates. Between the two, the remaining gap is validation rather than a
from-scratch audit.

PR5431's "Alternatives Considered" rejected an RWMutex on `localClient` because
it "does not fix the underlying data race on `checkState`". That objection was
correct at the time and no longer holds: the `checkState` race is fixed
independently by the snapshot work in that same PR.

## Decision

Remove the per-call lock from the query connection entirely.

- `localClient.mtx` becomes a `sync.Locker` rather than a `*sync.Mutex`
  (`tm2/pkg/bft/abci/client/local_client.go`). Every method body is unchanged.
- `localClientCreator.queryMtx` becomes a `noopMutex` — a `sync.Locker` whose
  `Lock`/`Unlock` do nothing (`tm2/pkg/bft/proxy/client.go`). The consensus and
  mempool connections keep their real `sync.Mutex`.

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

## Consequences

**The invariant becomes load-bearing.** "Everything reachable from
`Application.Query` and `Application.Info` is goroutine-safe" was previously an
unused property. It is now a correctness requirement, and breaking it
reintroduces a data race in production rather than a slow query.

**`SetResponseCallback` is unsynchronised on the query connection.** Its write
to `localClient.Callback` would race with concurrent `completeRequest` reads.
Nothing calls it there — documented on `NewReadOnlyABCIClient`.

**The `NewLocalClient` nil guard changed shape.** With an interface parameter,
`if mtx == nil` catches an untyped nil but not a typed nil `*sync.Mutex`, which
would panic on first use. Documented on the constructor; no in-tree caller does
this.

**Simulate's pre-first-block path is still single-threaded.** When
`getLastBlockHeader()` reports height < 1 (right after `InitChain`),
`BaseApp.Simulate` falls back to `getContextForTx`, i.e. the shared
`checkState`, and concurrent calls race on its gas meter and live stores.
Reached only before the first block is committed, so no production node serves
queries in that window, but it is now genuinely unlocked rather than
serialised. Left as-is here and flagged for a separate fix.

## Validation

`gno.land/pkg/gnoland/app_parallel_query_test.go`
(`TestParallelQueries_NWaySimulate`) is the coverage this change required and
the tree did not have. The pre-existing `TestQueryRace_FastIndexParity` pits one
query against one committer — an interleaving the old `queryMtx` already
permitted — so it can never observe two queries overlapping.

The new test runs N=8 goroutines simulating signed bank sends through the
read-only connection, in two phases: with a consensus block loop committing
underneath, and against frozen committed state. It asserts

1. **overlap actually happened**, via a gauge that counts calls inside
   `Application.Query`. The gauge is a probe application wrapped *beneath* the
   ABCI client, because that is where the mutex is held: a gauge in the test's
   own goroutines would count callers merely blocked on the mutex and report
   overlap even on fully serialised code;
2. **no data race**, under `-race`;
3. **determinism** — each querier's `GasUsed` is bit-identical across all rounds
   against frozen state.

Negative control: restoring a real mutex on `queryMtx` fails the test with
`peak in-flight = 1`.

Gas readings differ between queriers by design (different accounts mean
different address and account encodings, hence different per-byte read gas), and
drift across rounds in phase 1 (each simulate snapshots whatever height has
committed by then). Only the per-querier series against frozen state is a fixed
quantity, and that is what the determinism assertion checks.
