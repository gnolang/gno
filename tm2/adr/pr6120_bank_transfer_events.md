# PRxxxx: Structured bank transfer events

## Status

Proposed.

## Context

Bank transfers changed balances without an event. The old commented code in the
bank keeper and handlers came from a Cosmos SDK API that tm2 does not provide, so
enabling it was not a matter of uncommenting it. This left `MsgSend`, a realm's
`banker.SendCoins`, and the `Send` envelope of `MsgCall`/`MsgRun` invisible to an
indexer unless the transfer was already described by the transaction message.

`MsgMultiSend` is deliberately out of scope. It has no producer today (no client,
CLI, or gnoclient constructs one) and is not registered with the bank Amino
package, so no multisend can currently be submitted and there is nothing for an
indexer to reconstruct. Emitting a multisend event and registering the message to
carry it is a separate change with its own transaction-encoding surface, left for
when multisend gains a producer.

tm2 events are concrete Go values implementing `abci.Event`, registered with
amino, and emitted through `sdk.Context.EventLogger`. GnoVM events already use
this path and are returned in `ResponseDeliverTx.Events`.

## Decision

Register and emit `bank.TransferEvent` for one-to-one sends:

```go
type TransferEvent struct {
    From  string    `json:"from"`
    To    string    `json:"to"`
    Coins std.Coins `json:"coins"`
}
```

Addresses are bech32 strings. The event carries no custom marshaler, so it
serializes like the other struct events already returned in transaction results
(for example `StorageDepositEvent`). In `ResponseBase.EncodeEvents` the
indexer-facing shape is
`{"from":"...","to":"...","coins":[{"denom":"ugnot","amount":7}]}`, and the amino
wire encoding tags it with its registered type `/bank.TransferEvent`.

`sendCoins` emits one event after both the debit and credit succeed. This one
point covers `MsgSend`, realm banker sends, and VM message send envelopes.

The dead handler-level module-marker comments are removed. Unrestricted sends
used for gas and storage accounting remain outside this event: the requested
public transfer paths use `SendCoins`, while `SendCoinsUnrestricted` deliberately
bypasses transfer policy.

## Alternatives considered

**Restore Cosmos string-keyed events.** Rejected because tm2 has neither the API
nor those event and attribute constants. It would introduce a second event model.

**Emit in handlers.** Rejected because realm banker sends and VM send envelopes
do not pass through bank message handlers. Keeper emission covers all requested
paths once.

**Include multisend now.** Rejected as out of scope: `MsgMultiSend` has no
producer and is not registered, so no multisend can be submitted and there is
nothing to index. Registering the message and shaping its event (a batch event vs
separate from-only/to-only events) is deferred to when multisend gains a producer.

## Consequences

The event type names and JSON fields are an indexer-facing compatibility
contract. Adding the events is consensus-visible in transaction results but does
not change balances or app state.

`EventLogger.EmitEvent` appends to an in-memory slice and charges no gas. The
additional response bytes therefore increase transaction-result size without
changing gas consumption. Failed messages do not expose their accumulated events
because the SDK discards events from a failed message transaction.
