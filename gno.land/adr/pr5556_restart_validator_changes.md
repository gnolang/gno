# ADR: Recover Validator Changes After Node Restart

## Context

The `EndBlocker` in `gno.land/pkg/gnoland/app.go` previously relied on an
in-memory event collector to decide whether to query the VM for validator set
changes. The collector listened on the `EventSwitch` for `validatorUpdate`
events fired by `fireEvents` (in `tm2/pkg/bft/state/execution.go`) **after**
each block was committed. When events were present the EndBlocker would call
`GetChanges(lastHeight, lastHeight)` on `r/sys/validators/v2` and forward the
updates to Tendermint's consensus layer.

**The bug:** `fireEvents` is called after `SaveABCIResponses` and `SaveState`,
so the collector is populated *between* block N-1 being committed and block N
being processed. On node restart the in-memory collector is empty, so
`EndBlocker(N)` returned early and never applied the validator changes from
block N-1 — even though those changes were already committed to the realm.

This was confirmed by an integration test: a validator added via GovDAO proposal
and verified in the realm (`IsValidator` returns `true`) disappeared from the
consensus set after a restart.

Additionally, the collector introduced an inherent **one-block delay**: events
from block N-1's txs were consumed only in block N's EndBlocker. The
`GetChanges(N-1, N-1)` query fetched changes already committed one block earlier.

## Decision

Replace the `EventSwitch`-based collector with **same-block event detection via
`EndTxHook`**. During each block's execution, `BaseApp.runTx` calls
`EndTxHook(ctx, result)` after each successful transaction. The hook now checks
`result.Events` for validator add/remove events from `r/sys/validators/v2` and
sets an in-closure boolean flag.

The `EndBlocker` receives a `func() bool` that reads and resets the flag. When
the flag is set it queries the VM for **the current block**:

```go
hasValEvent := false

baseApp.SetEndTxHook(func(ctx sdk.Context, result sdk.Result) {
    if result.IsOK() {
        vmk.CommitGnoTransactionStore(ctx)
        if !hasValEvent {
            hasValEvent = hasValidatorChangeEvent(result.Events)
        }
    }
})

baseApp.SetEndBlocker(EndBlocker(
    func() bool {
        had := hasValEvent
        hasValEvent = false
        return had
    },
    acck, gpk, vmk, baseApp,
))
```

Inside the `EndBlocker`, the query uses `req.Height` (the current block) instead
of `app.LastBlockHeight()` (the last committed block):

```go
response, err := vmk.QueryEval(ctx, valRealm,
    fmt.Sprintf("%s(%d,%d)", valChangesFn, req.Height, req.Height))
```

This is correct because:
1. `EndTxHook` fires during `DeliverTx`, *before* `EndBlock` — so by the time
   `EndBlocker` runs, `hasValEvent` already reflects the current block's txs.
2. The `deliverState` ctx passed to `EndBlocker` contains all the uncommitted
   writes from the current block, so `QueryEval` sees the realm changes from
   the current block's txs (`std.GetHeight()` equals `req.Height` during tx
   execution).
3. Tendermint persists the `ValidatorUpdates` from `EndBlock(N)` as part of
   block N's ABCIResponses. On restart Tendermint restores these from its state
   DB — no special recovery code is needed.

The `firstBlock` flag introduced in an earlier version of this fix is removed.
The in-memory `collector[T]` type and `events.go` are deleted entirely. The
`validatorEventFilter` function (which operated on `EventSwitch` events) is
replaced by `hasValidatorChangeEvent` (which operates directly on
`[]abci.Event`).

## Alternatives Considered

### 1. `firstBlock` flag (initial fix in this PR)

On the first EndBlock after startup, unconditionally query the VM regardless
of the collector state.

**Rejected in favour of same-block:** same-block processing makes restarts
inherently safe (validator updates are persisted by Tendermint in EndBlock(N)),
eliminates the one-block delay, and removes the collector entirely. The
`firstBlock` approach was a narrow patch; same-block is the correct model.

### 2. Read `ABCIResponses(N-1)` from the Tendermint state DB

On each EndBlock, load the previous block's `ABCIResponses` from the TM state
DB (which is persisted before `fireEvents`), scan the DeliverTx events, and
decide whether to query the VM.

**Rejected:** the TM state DB is not currently accessible from the app layer
without significant plumbing. Benchmarking showed that the amino-decode step
alone costs ~50 µs/block with 1k allocs, comparable to the VM query itself.
Since same-block processing achieves both the restart fix and the one-block
improvement with no extra DB access, the ABCIResponses approach offers no
advantage.

### 3. Always query the VM (remove the gate entirely)

Remove all gating and call `GetChanges` in every `EndBlocker`.

**Rejected:** a real VM query costs ~86 µs/block and 53 KB of allocations vs
~3 ns/block for a boolean check. The boolean gate is cheap and correct.

### 4. Fire events from `EndTxHook` into the existing EventSwitch collector

Reuse the collector infrastructure but populate it via the hook rather than
the post-commit `fireEvents`.

**Rejected:** same-block processing achieves the same result more directly
without the indirection of the EventSwitch.

## Consequences

- **Positive:** validator changes are applied to consensus in the **same block**
  as the transaction, one block earlier than before.
- **Positive:** Tendermint persists EndBlock validator updates in its state DB.
  Restarts are safe without any special recovery code.
- **Positive:** the event collector and EventSwitch dependency are removed from
  the EndBlocker path; `events.go` is deleted.
- **Positive:** `endBlockerApp` no longer requires `LastBlockHeight()`.
- **Testing:** a txtar integration test (`restart_validators.txtar`) verifies
  the full flow: add validator via GovDAO, restart, confirm validator appears in
  the consensus set. The `gnorpc validators` testscript command is added to
  poll the consensus validator set.
