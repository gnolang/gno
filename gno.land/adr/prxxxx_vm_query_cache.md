# ADR: Answer repeat VM queries from the last result

## Status

Proposed

## Context

Every `vm/q*` query re-runs from scratch. Rendering a realm page means
loading its objects out of the store, rebuilding them in the VM, and
evaluating `Render`. None of that work is kept, so the thousandth view of a
page costs exactly what the first one did.

Profiling a single render of a realm holding a few thousand objects showed:

- ~224,000 allocations and ~9.6 MB of garbage per render
- about 43% of CPU time in the garbage collector
- ~19.2 µs per stored object, plus ~217 µs fixed cost

Six identical renders in a row reused nothing.

This is also what makes `abci_query` cheap to abuse. Queries are
unauthenticated and unmetered by the fee market, and `localClient.QuerySync`
holds a single application mutex for the whole call, so repeated expensive
queries serialise behind each other and crowd out real work.

Two facts make a cache possible:

1. Every `vm/q*` endpoint is read-only.
2. A query's answer is fixed by two things — the state version it reads,
   chosen by `req.Height`, and the context it runs in. `handleQueryCustom`
   builds that context from the *latest* block header, whatever `req.Height`
   asks for.

So for a fixed chain tip, the same request always has the same answer.

## Decision

Add a cache in front of `vmHandler.Query`, in `gno.land/pkg/sdk/vm/query_cache.go`.

**Keyed on the request, scoped to the tip.** The key covers `req.Height`,
`req.Path` and `req.Data`. The chain tip is deliberately *not* in the key:
when the tip moves, the whole map is dropped instead. That is what makes the
cache easy to reason about — a hit can only ever come from the same tip that
produced it, so there is no per-entry invalidation to get wrong.

**The key is length-prefixed, not separated.** `req.Path` is caller-supplied
and may contain any byte, so a separator is ambiguous. With a NUL separator,
path `vm/q` with data `\x00x` and path `vm/q\x00` with data `x` produce the
same key, and one query would be answered with another's result. Each
variable-length part is therefore preceded by its length.

**Bounded.** At most 1024 entries and 32 MB total, and no single entry over
1 MB. Reaching a bound stops new entries for the rest of the block; nothing
is evicted early, because the map dies at the next block anyway.

**Copies in and out.** A hit returns a copy of the byte fields, so no caller
can edit what the next hit returns.

**Nothing is stored for a panicking query.** The store happens after the
endpoint switch returns normally, not in a `defer`. A `defer` runs while
panicking too, and the response is still empty at that point — the next
identical query would then get an empty success in place of the panic.
`queryRender` panics on malformed input, so this path is reachable.

Deterministic *error* responses are cached. They are the correct answer for
that request at that tip, and caching them also blunts a cheap way to make
the node redo failing work.

## Alternatives considered

**Per-entry invalidation.** Track which objects a query read and drop only
the affected entries. Much more precise, and much easier to get wrong — a
missed dependency serves stale state. Dropping everything each block costs
one re-evaluation per distinct query per block and cannot go stale.

**Cache in gnoweb or the RPC layer.** Would help page views but not other
clients, and the node still burns the CPU when anyone bypasses it. The cost
is in the VM, so the fix belongs there.

**LRU with eviction.** Unnecessary. Entries live at most one block, so a
simple bound plus a full drop is enough.

## Consequences

- Repeat queries within a block become nearly free. A measured repeat render
  went from 19.4 ms to 248 ns, with allocations down to 3.
- Memory grows by at most 32 MB per node.
- A caller issuing *distinct* expensive queries gets no benefit from this,
  and the underlying `abci_query` exposure remains. A wall-clock query
  timeout, a semaphore in place of the single mutex, and rate limiting are
  the follow-ups. The real repair is the per-object render cost itself.
- Historical queries (`req.Height` set) are keyed apart from current ones but
  are still dropped when the tip moves, so they are re-evaluated more often
  than strictly necessary. That is accepted for the simpler invalidation
  rule.
