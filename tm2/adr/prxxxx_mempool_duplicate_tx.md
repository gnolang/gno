# ADR-prxxxx: One mempool entry per transaction

## Context

`resCbFirstTime` admits a checked transaction without consulting `txsMap`, and
`addTx` stores the key unconditionally. When a transaction already in the list
arrives a second time, the newer `CElement` takes over the map entry and the
older one stays in the list with nothing pointing at it.

`Update`'s own loop finds transactions through that map, so committing the
transaction removes one copy and leaves the other. The recheck `Update` starts
does reach the stranded copy: `recheckTxs` and `resCbRecheck` hand `removeTx` an
element from their own walk of `mem.txs`, and `removeTx` deletes the map key
whatever element it was given, so whichever copy recheck unlinks first takes the
other's entry.

Two conditions have to hold for the second arrival to get that far.

The cache has to miss. It is not an index. At `cache_size` 0, which
`ValidateBasic` accepts, it is `nopTxCache` and every resubmission is a miss. At
the default 10000 it is a bounded LRU, and `Update` pushes every committed
transaction into it, so it turns over on ordinary throughput and a transaction
that stays pending long enough falls out.

The application has to admit the same bytes twice. gno.land does not, at the
default `recheck` of true: `Update` re-runs `CheckTx` on everything still
pending under the lock `Commit` holds, the auth ante's sequence bump is back in
`checkState` before any peer can resubmit, and the second check fails signature
verification into `resCbFirstTime`'s `res.Error != nil` branch. At `recheck`
false, and under any tm2 application carrying no replay protection of its own,
the second check passes.

Measured with the guard removed, driving the real gno.land app behind a
`CListMempool` at `cache_size` 0 and resubmitting one signed transaction after a
block that does not carry it:

```
recheck=true    Size=1  TxsBytes=307
recheck=false   Size=2  TxsBytes=614
```

Measured with a `kvstore` application, which has no replay protection: at
`cache_size` 0, one transaction sent three times gives `Size()` 3 and `txsBytes`
33 for eleven bytes. At `CacheSize` 1000, ten blocks of a hundred transactions
the node never checked take the cache to its cap and drop a transaction that
stays listed throughout, so `Update` alone reaches the eviction, and
resubmitting then puts two elements under one key. One `ReapMaxTxs` returns that
transaction twice, and committing it leaves one copy listed with `txsMap`
holding no key for it.

Upstream fixed this in CometBFT PR 890.

## Decision

`resCbFirstTime` loads `txsMap` before it builds the entry. On a hit it records
the sending peer on the transaction already listed, then returns without adding,
notifying or logging an admission.

That load, the type assertions under it and the sender write are the same three
lines `CheckTxWithInfo` runs on a cache hit, so both sites now call
`recordSender`.

`reqResCb` still runs the external callback, so an RPC broadcast of a
transaction the mempool already holds reports the application's response
unchanged.

## Alternatives considered

* **The check in `addTx`.** `addTx` has one caller, so the guard would sit
  further from the decision it serves, and recording the sender would need a
  second way out of it. Rejected on placement.
* **`removeTx` refusing to unlink an element the map does not name.** This is
  the other half of the same hazard: unlinking a stranded copy deletes the
  surviving copy's map entry. Rejected because the guard above means no stranded
  copy is ever created, so the second guard would defend a state that cannot
  arise, not because recheck cannot reach one.

## Consequences

A resubmitted transaction is no longer counted twice in `Size()` or `txsBytes`,
and no longer reaped twice into one block.

A block carrying the same transaction twice was never a consensus fault. Neither
`Block.ValidateBasic` nor `State.ValidateBlock` checks that a block's
transactions are distinct, so every honest peer accepts such a block. The damage
is in what the node then reports. `saveTxResultIndex` keys on `tx.Hash()` alone
and `ApplyBlock` writes one record per block index, so the second copy
overwrites the first: a transaction that executed is served by `/tx?hash=` as
the failing replay beside it. `/unconfirmed_txs` counts the same transaction
twice in `Count`, `Total` and `TotalBytes`.

The second check's `GasWanted` is discarded and the listed entry keeps the value
from the first. Upstream behaves the same way.

Mempool contents only. No wire format, no storage, no hash moves.

## Riding along

Three defects that predate this change and share its surface.

* **`notifyTxEvent` closed a resolved waiter a second time.** `td.waiters` had
  no `delete` anywhere, and `fireEvents` emits one `EventTx` per block index, so
  a block carrying one transaction's bytes twice drove `notifyTxEvent` twice for
  one waiter and panicked with `close of closed channel` in `listenRoutine`'s
  own goroutine, where nothing recovers. The waiters for a transaction are now
  dropped as it resolves, so the second event finds none, and `getTxResult`
  drops its own on the timeout and quit paths, which is what its docstring
  already promised and what stops a node serving `/broadcast_tx_commit` from
  retaining every timed-out transaction's bytes for its lifetime.

  Dropping on timeout is only safe once a waiter belongs to one caller.
  `td.waiters` held one `txWaiter` per transaction and joined every later caller
  to it, and its `timeCh` is a single `time.After` channel carrying one value,
  so of two callers sharing a waiter only one ever received the timeout. That
  one removing the shared waiter left the other unreachable from
  `notifyTxEvent`, blocked on a `waitCh` nothing would close and a `quit` that
  is nil at the only call site. The map now holds one waiter per caller, all of
  them resolved with the same result, each under the timeout it asked for rather
  than the first caller's remainder.
* **`ReapMaxTxs(n)` returned n+1 transactions.** The loop tested
  `len(txs) <= maxVal`, so `/unconfirmed_txs?limit=n` reported one entry more
  than it was asked for and the result outgrew the capacity the line above
  reserves for it.
* **`addTx` and `removeTx` said their callbacks run with `mem.mtx` not held.**
  `localClient` is the only ABCI client and `completeRequest` invokes the
  callback inline, so both run under `CheckTxWithInfo`'s or `Update`'s caller's
  lock. `resCbFirstTime`'s check-then-add depends on that serialization, and the
  comment read as license to break it.

## Verification

`TestMempoolNoDuplicateAfterCacheEviction` commits ten blocks the node never
checked, resubmits the one transaction that stayed listed, and asserts a single
list element, `txsBytes` equal to that transaction alone, the second peer
recorded on it, and an empty mempool once it commits. It sets `CacheSize`
itself, since the block loop only evicts while the two divide.
`TestMempoolNoDuplicateWithoutCache` covers the `nopTxCache` path at
`cache_size` 0, where no eviction is needed, and asserts `Size()` and `txsBytes`
alone.

`TestMempoolReapMaxTxsLimit` reaps every limit from 0 to the mempool's size and
asserts the count each returns.

`TestTxDispatcher_DuplicateTxEvent` resolves a waiter, asserts it left
`td.waiters`, and sends the same transaction's event a second time.
`TestTxDispatcher_TimeoutForgetsWaiter` asserts the same after a timeout.
`TestTxDispatcher_SecondCallerSurvivesFirstTimeout` runs a 20ms caller and a 30s
caller on one transaction, lets the first time out, and asserts the second still
receives the result.

Negative control: each guard removed in turn reddens its own test, and removing
the `resCbFirstTime` guard reports two elements.
