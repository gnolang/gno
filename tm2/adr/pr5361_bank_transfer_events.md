# Bank Module Transfer Events

## Context

The bank module was ported from Cosmos SDK (commit `9bbdf7117`, 2021-08-31) with all
event emission code commented out because the Cosmos SDK event API (`EventManager`,
`NewEvent`, `NewAttribute`) was not ported to Gno.

The Gno event system was built later in two phases:
- `EventLogger` and `std.Emit` — PR [#1653](https://github.com/gnolang/gno/pull/1653) (2024-04-30)
- Typed ABCI events (`StorageDepositEvent`, etc.) — PR [#4630](https://github.com/gnolang/gno/pull/4630) (2025-09-02)

The VM module adopted the event system, but the bank module was never updated. This
makes it impossible for indexers to track balance changes via events.

Ref: https://github.com/gnolang/gno/issues/5344

## Decision

Add a single `TransferEvent` type to `gnovm/stdlibs/chain/emit_event.go` (alongside
existing `StorageDepositEvent` and `StorageUnlockEvent`) and emit it from the bank
keeper whenever coins are transferred.

### Event type

```go
type TransferEvent struct {
    From   crypto.Address `json:"from"`
    To     crypto.Address `json:"to"`
    Amount std.Coins      `json:"amount"`
}
```

### Emission points

- `sendCoins()`: emits with both `From` and `To` populated. Covers `SendCoins`,
  `SendCoinsUnrestricted` (gas fees), and any future callers.
- `InputOutputCoins()`: emits per-input (only `From` set) and per-output (only `To` set).

### Gas fee transfers emit events

`SendCoinsUnrestricted` (used for gas fee deduction in `auth/ante.go`) calls
`sendCoins()` internally, so gas fee transfers automatically produce events.
This matches Cosmos SDK behavior where `SendCoins` always emits regardless of caller.

## Alternatives considered

- **Generic `chain.Event` with string attributes**: Would avoid a new type but places
  bank-level semantics in string parsing, losing type safety for indexers.

- **Multiple event types like Cosmos SDK** (`CoinSpent`, `CoinReceived`, `Transfer`):
  Cosmos SDK emits 5 distinct event types for coin operations:

  | Event | Attributes | Purpose |
  |-------|-----------|---------|
  | `coin_spent` | spender, amount | Low-level: coins left an account |
  | `coin_received` | receiver, amount | Low-level: coins entered an account |
  | `transfer` | sender, recipient, amount | High-level: send correlation |
  | `coinbase` | minter, amount | Minting |
  | `burn` | burner, amount | Burning |

  The separation of `coin_spent`/`coin_received` from `transfer` serves different
  indexer needs and handles asymmetric operations (minting has no sender, burning has
  no recipient). However, Gno currently has no minting/burning in the bank module, so
  a single `TransferEvent` with optional empty `From`/`To` is sufficient. If Gno adds
  minting or burning later, dedicated event types should be introduced at that point
  rather than overloading `TransferEvent`.

- **Suppress events for gas fees**: Would require splitting `sendCoins` into emitting
  and silent variants, adding complexity for little benefit.

## Consequences

- Indexers can track all coin movements (sends, multi-sends, gas fees) via `TransferEvent`.
- Existing behavior is unchanged; events are additive.
- The `TransferEvent` is amino-registered under the `"tm"` prefix, consistent with
  other chain events.
- Realm-level minting/burning (`IssueCoin`/`RemoveCoin` via `SDKBanker`) calls
  `AddCoins`/`SubtractCoins` directly, bypassing `sendCoins()`, so these operations
  remain invisible to indexers. Adding events to `AddCoins`/`SubtractCoins` would cause
  double-emission on transfers. Dedicated `MintEvent`/`BurnEvent` types (or emission
  from the `SDKBanker` layer) should be introduced separately.
