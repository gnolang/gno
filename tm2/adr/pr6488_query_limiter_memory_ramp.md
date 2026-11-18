# ADR: Sizing The Query Connection From Free Memory

## Context

PR6082 let several ABCI queries run inside the application at once, bounded by
`queryLimiter` at `GOMAXPROCS` slots. Each query carries its own allocation
budget (`maxAllocQuery`, 1.5 GB, installed twice per query), so the memory a
node may hold for queries grew with its core count: `GOMAXPROCS × 3 GB` in the
worst case. A first fix capped the bound at 4. Review rejected the hard cap:
machines that can afford more parallelism should get it, without per-machine
tuning. The integration suite had already solved the same problem for test
nodes in gnolang/gno#6061, with a budget that ramps against the machine's free
memory.

## Decision

- The memory reader from gnolang/gno#6061 moves from `tm2/pkg/testutils` to a
  new package, `tm2/pkg/meminfo` (`Info`, `Read`), so production code does not
  import a test helper package. The integration suite's `nodeBudget` now reads
  memory through it too.
- `queryLimiter` keeps its role as the `sync.Locker` the read-only connection
  holds, but its bound now moves. It starts at `min(GOMAXPROCS, 4)` and stays
  between that and `GOMAXPROCS`. As in `nodeBudget`, a quarter of the memory
  is a reserve left to everything else:
  - up one slot when callers running and waiting fill every slot and free
    memory covers the reserve plus a 1 GiB margin (`queryMemMargin`), room for
    one more query;
  - down one slot when free memory falls short of the reserve;
  - unchanged in between, which keeps it from oscillating.
  Readings are taken at most once a second, from `Lock`. Waiting callers count
  towards saturation because the reading after an `Unlock` lands in the
  waiter it woke, which always finds a slot free: counting only running
  queries, a burst that queued at once never grew the limit. A slot that a
  grown limit adds goes to the caller that grew it or to a waiter it wakes; a
  limit that shrinks takes effect as running queries return.
- Where memory cannot be read at startup (platforms other than Linux and
  macOS, or a cgroup v1 memory controller), the bound is fixed at `GOMAXPROCS`,
  as it was before this change: with no reading, the cores are the only
  measure of what the machine can handle. Machines with four cores or fewer
  are also fixed at `GOMAXPROCS`, since there is nothing to ramp to.
- There is no node setting for the bound.

## Alternatives Considered

- **Hard cap at 4.** Rejected in review: it leaves larger machines idle and
  does not stop a burst of expensive queries, it only triggers earlier.
- **Size once at startup from the worst case**,
  `clamp((available − reserve) / maxAllocQuery, …)`. Rejected: it sizes every
  machine for the rare query that spends its whole budget, needs the
  application's per-query cost threaded into tm2, and invites a config knob.
- **Share one ramp with `nodeBudget`.** Deferred: `nodeBudget`'s accounting
  for memory still owed and still returning would have to become hooks, for two
  callers. The two controllers have the same shape, so a change to pacing or to
  the saturation rule in one should be checked against the other.
- **Port `nodeBudget` whole**, including its accounting for memory a running
  worker has yet to take and for memory a cut worker has yet to return. Not
  needed here: queries are short compared with test nodes, so readings catch
  up quickly.
- **Grow while more than half the memory is free.** Rejected: on a node whose
  own state takes most of its memory, the limit would sit at the floor with
  plenty of room left for queries.
- **Fix the bound at `min(GOMAXPROCS, 4)` where memory cannot be read**, as
  `nodeBudget` does. Rejected: it is the hard cap again, on exactly the
  machines (cgroup v1 containers, the BSDs) where nothing else would lift it.

## Consequences

- A burst of maximum-cost queries can still push free memory below the
  reserve before the readings catch up; the reserve is what absorbs it.
- `queryMemMargin` is a chosen margin, not a measured query cost: tm2 does not
  know what a query costs, and the application does not tell it.
- Where memory cannot be read, the limit stays at `GOMAXPROCS`, as before this
  change.
- The limit only moves when a caller arrives, so an idle node keeps whatever
  limit it last reached until the next query.
- `sync.Cond` replaces the buffered channel, so waiting callers are no longer
  admitted strictly in arrival order.
