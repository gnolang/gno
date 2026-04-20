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

Add `BlockEvents() []abci.Event` to `BaseApp` in `tm2/pkg/sdk/baseapp.go`.
The slice is reset at the start of each block (`BeginBlock`) and appended with
events from every successful `DeliverTx`. The EndBlocker reads it directly via
the `endBlockerApp` interface — no hook, no closure, no shared mutable state.

**Changes to `tm2/pkg/sdk/baseapp.go`:**

```go
// new field on BaseApp
blockEvents []abci.Event // reset at BeginBlock, appended by DeliverTx

// BeginBlock resets the slice (reusing capacity)
app.blockEvents = app.blockEvents[:0]

// DeliverTx appends on success
if result.IsOK() {
    app.blockEvents = append(app.blockEvents, result.Events...)
}

// exposed method
func (app *BaseApp) BlockEvents() []abci.Event { return app.blockEvents }
```

**Changes to `gno.land/pkg/gnoland/app.go`:**

```go
// endBlockerApp interface
type endBlockerApp interface {
    Logger() *slog.Logger
    BlockEvents() []abci.Event
}

// EndBlocker checks events directly — no closure flag
func EndBlocker(acck, gpk, vmk, app endBlockerApp) func(...) {
    return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
        // ... auth/gas price logic ...

        if !hasValidatorChangeEvent(app.BlockEvents()) {
            return abci.ResponseEndBlock{}
        }

        response, err := vmk.QueryEval(ctx, valRealm,
            fmt.Sprintf("%s(%d,%d)", valChangesFn, req.Height, req.Height))
        // ...
    }
}
```

The query uses `req.Height` (the current block) because:

1. `DeliverTx` runs before `EndBlock` — by the time `EndBlocker` executes,
   `BlockEvents()` already reflects all of the current block's successful txs.
2. The `deliverState` ctx passed to `EndBlocker` contains all uncommitted writes
   from the current block, so `QueryEval` sees the realm changes written during
   the current block's txs (`std.GetHeight()` equals `req.Height` during tx
   execution).
3. Tendermint persists the `ValidatorUpdates` returned from `EndBlock(N)` as
   part of block N's state. On restart Tendermint restores these from its state
   DB — no special recovery code is needed.

The in-memory `collector[T]` type and `events.go` are deleted entirely. The
`validatorEventFilter` function (which operated on `EventSwitch` events) is
replaced by `hasValidatorChangeEvent([]abci.Event) bool`.

## Alternatives Considered

### 1. `firstBlock` flag (initial fix in this PR)

On the first EndBlock after startup, unconditionally query the VM regardless
of the collector state.

**Rejected in favour of same-block:** same-block processing makes restarts
inherently safe (validator updates are persisted by Tendermint in EndBlock(N)),
eliminates the one-block delay, and removes the collector entirely. The
`firstBlock` approach was a narrow patch; same-block is the correct model.

### 2. EndTxHook + closure bool flag (intermediate approach in this PR)

Set a `hasValEvent bool` in `EndTxHook` and consume it in the EndBlocker via
a `func() bool` closure — both closures sharing the same captured variable.

**Rejected:** although the ABCI flow is single-threaded (DeliverTx always
completes before EndBlock), sharing mutable state across two closures is
conceptually fragile and would trigger the race detector if the execution model
ever changed. `BaseApp.BlockEvents()` is a clean pull model with no shared
mutable state outside of BaseApp itself.

### 3. Read `ABCIResponses(N-1)` from the Tendermint state DB

On each EndBlock, load the previous block's `ABCIResponses` from the TM state
DB (which is persisted before `fireEvents`), scan the DeliverTx events, and
decide whether to query the VM.

**Rejected:** the TM state DB is not currently accessible from the app layer
without significant plumbing. Benchmarking showed that the amino-decode step
alone costs ~50 µs/block with 1k allocs, comparable to the VM query itself.
Since same-block processing achieves both the restart fix and the one-block
improvement with no extra DB access, the ABCIResponses approach offers no
advantage.

### 4. Always query the VM (remove the gate entirely)

Remove all gating and call `GetChanges` in every `EndBlocker`.

**Rejected:** a real VM query costs ~86 µs/block and 53 KB of allocations vs
~3 ns/block for a boolean check. The boolean gate is cheap and correct.

### 5. Fire events from `EndTxHook` into the existing EventSwitch collector

Reuse the collector infrastructure but populate it via the hook rather than
the post-commit `fireEvents`.

**Rejected:** same-block processing achieves the same result more directly
without the indirection of the EventSwitch.

## Consequences

- **Positive:** validator changes are applied to consensus in the **same block**
  as the transaction, one block earlier than before.
- **Positive:** Tendermint persists EndBlock validator updates in its state DB.
  Restarts are safe without any special recovery code.
- **Positive:** the event collector, EventSwitch dependency, and `events.go` are
  removed from the EndBlocker path entirely.
- **Positive:** `BaseApp.BlockEvents()` is a general-purpose addition to tm2
  that any EndBlocker can use to inspect the current block's tx events.
- **Positive:** `endBlockerApp` no longer requires `LastBlockHeight()`, and the
  EndBlocker signature no longer takes a `func() bool` gate.
- **Testing:** a txtar integration test (`restart_validators.txtar`) verifies
  the full flow: add validator via GovDAO, restart, confirm validator appears in
  the consensus set. The `gnorpc validators` testscript command is added to
  poll the consensus validator set.
