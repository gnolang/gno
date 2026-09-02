# PRxxxx: Structured bank transfer events

## Status

Proposed.

## Context

Bank transfers changed balances without an event. The old commented code in the
bank keeper and handlers came from a Cosmos SDK API that tm2 does not provide, so
enabling it was not a matter of uncommenting it. This left `MsgSend`,
`MsgMultiSend`, a realm's `banker.SendCoins`, and the `Send` envelope of
`MsgCall`/`MsgRun` invisible to an indexer unless the transfer was already
described by the transaction message.

tm2 events are concrete Go values implementing `abci.Event`, registered with
amino, and emitted through `sdk.Context.EventLogger`. GnoVM events already use
this path and are returned in `ResponseDeliverTx.Events`.

## Decision

Register and emit `bank.TransferEvent`:

```go
type TransferEvent struct {
    From   string    `json:"from"`
    To     string    `json:"to"`
    Amount std.Coins `json:"amount"`
}
```

Addresses are bech32 strings. The event carries no custom marshaler, so it
serializes like the other struct events already returned in transaction results
(for example `StorageDepositEvent`). In `ResponseBase.EncodeEvents` the
indexer-facing shape is
`{"from":"...","to":"...","amount":[{"denom":"ugnot","amount":7}]}`, and the amino
wire encoding tags it with its registered type `/bank.TransferEvent`.

`sendCoins` emits one event after both the debit and credit succeed. This one
point covers `MsgSend`, realm banker sends, and VM message send envelopes.

`InputOutputCoins` emits a debit event with only `From` after each successful
input subtraction, followed by a credit event with only `To` after each
successful output addition. This matches the GRC20 burn and mint convention, so
an indexer can update both sides using events alone without inventing a
one-to-one mapping. Inputs and outputs each retain their slice order.

The dead handler-level module-marker comments are removed. Unrestricted sends
used for gas and storage accounting remain outside this event: the requested
public transfer paths use `SendCoins`, while `SendCoinsUnrestricted` deliberately
bypasses transfer policy.

## Alternatives considered

**Restore Cosmos string-keyed events.** Rejected because tm2 has neither the API
nor those event and attribute constants. It would introduce a second event model.

**Emit one multisend event containing every input and output.** This preserves
the complete relation but creates another public event type and makes ordinary
transfer tracking different from `MsgSend`.

**Assign each output the first input as sender.** Rejected because it records
false provenance for N:M multisends. Separate from-only debit and to-only credit
events preserve every balance change without claiming a relationship between
individual inputs and outputs.

**Emit in handlers.** Rejected because realm banker sends and VM send envelopes
do not pass through bank message handlers. Keeper emission covers all requested
paths once.

## Consequences

The event type name and its three JSON fields are an indexer-facing compatibility
contract. Adding the event is consensus-visible in transaction results but does
not change balances or app state.

`EventLogger.EmitEvent` appends to an in-memory slice and charges no gas. The
additional response bytes therefore increase transaction-result size without
changing gas consumption. Failed messages do not expose their accumulated events
because the SDK discards events from a failed message transaction.
