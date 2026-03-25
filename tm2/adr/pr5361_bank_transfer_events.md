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

Add three event types to `tm2/pkg/sdk/bank/events.go`, following the Cosmos SDK
convention of separate event types for different operations. All types live in the
bank module (not in `gnovm/stdlibs/chain/`) to respect the architectural layering:
`tm2` must not import `gnovm`.

### Event types

```go
// 1:1 transfers (SendCoins, SendCoinsUnrestricted)
type TransferEvent struct {
    From   crypto.Address `json:"from"`
    To     crypto.Address `json:"to"`
    Amount std.Coins      `json:"amount"`
}

// Coins leaving an account (InputOutputCoins inputs)
type CoinSpentEvent struct {
    Spender crypto.Address `json:"spender"`
    Amount  std.Coins      `json:"amount"`
}

// Coins entering an account (InputOutputCoins outputs)
type CoinReceivedEvent struct {
    Receiver crypto.Address `json:"receiver"`
    Amount   std.Coins      `json:"amount"`
}
```

Using separate types avoids zero-address sentinels in partially-populated fields
(which would serialize as `g1qqq...luuxe` and confuse indexers) and gives consumers
type-safe event discrimination.

### Emission points

- `sendCoins()`: emits `TransferEvent` with both `From` and `To` populated. Covers
  `SendCoins`, `SendCoinsUnrestricted`, and any future callers.
- `InputOutputCoins()`: emits `CoinSpentEvent` per-input and `CoinReceivedEvent`
  per-output. No `TransferEvent` is emitted because N:M multi-sends cannot be
  expressed as a single transfer.

### Gas fee transfers

`SendCoinsUnrestricted` (used for gas fee deduction in `auth/ante.go`) calls
`sendCoins()` internally, so gas fee transfers produce `TransferEvent`s. However,
note that the ante handler runs before `runMsgs()`, which creates a fresh
`EventLogger`. As a result, gas fee events are currently **not visible** in the
transaction result. Making ante-handler events visible requires changes to
`baseapp.go` and is tracked separately.

## Alternatives considered

- **Generic `chain.Event` with string attributes**: Would avoid a new type but places
  bank-level semantics in string parsing, losing type safety for indexers.

- **Single `TransferEvent` with optional empty `From`/`To` for multi-send**:
  Simpler (one type), but zero-value `crypto.Address` serializes to
  `g1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqluuxe` — a valid-looking bech32 address that
  would corrupt indexer balance tracking. Separate `CoinSpentEvent`/`CoinReceivedEvent`
  types are cleaner and match the Cosmos SDK pattern.

- **Full Cosmos SDK event set** (`CoinSpent`, `CoinReceived`, `Transfer`, `Coinbase`,
  `Burn`): Cosmos SDK emits 5 distinct event types. Gno currently has no
  minting/burning in the bank module, so `Coinbase` and `Burn` are not needed.
  Dedicated `MintEvent`/`BurnEvent` types should be introduced when those operations
  are added.

- **Suppress events for gas fees**: Would require splitting `sendCoins` into emitting
  and silent variants, adding complexity for little benefit.

## Consequences

- Indexers can track coin movements via `TransferEvent` (1:1 sends),
  `CoinSpentEvent` (debits in multi-send), and `CoinReceivedEvent` (credits in
  multi-send).
- Existing behavior is unchanged; events are additive.
- All three event types are amino-registered under the `"bank"` prefix in the bank
  module's amino package (`tm2/pkg/sdk/bank/package.go`).
- Realm-level minting/burning (`IssueCoin`/`RemoveCoin` via `SDKBanker`) calls
  `AddCoins`/`SubtractCoins` directly, bypassing `sendCoins()`, so these operations
  remain invisible to indexers. Adding events to `AddCoins`/`SubtractCoins` would cause
  double-emission on transfers. Dedicated `MintEvent`/`BurnEvent` types (or emission
  from the `SDKBanker` layer) should be introduced separately.
- Gas fee events are emitted but not yet visible in transaction results (see
  "Gas fee transfers" above). A follow-up change to `baseapp.go` is needed.
