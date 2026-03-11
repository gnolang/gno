# ADR-001: Multi-Denomination Gas Fees

## Status

Implemented

## Context

The `auth` module's `GasPriceKeeper` supports only a single gas fee denomination. This limits chains built on tm2 to accepting fees in one token. Chains that accept multiple tokens via IBC or other mechanisms have no way to let users pay gas fees in those tokens.

## Decision

Extend the `auth` module to support multiple fee denominations. Each denomination has its own independently tracked and dynamically adjusted gas price.

### Storage

Gas prices are stored per-denom using a prefix key:
```
GasPriceKeyPrefix = "gasPrice:"   // e.g. "gasPrice:ugnot", "gasPrice:uphoton"
```

### Parameters

`InitialGasPrices []std.GasPrice` replaces the former singular `InitialGasPrice`. This defines:
- Which denominations are accepted for gas fees
- The floor price for each denom (dynamic price cannot drop below this)

Validation rejects duplicate denoms, empty denoms, and non-positive gas values.

### GasPriceKeeper

- `SetGasPrice(ctx, gp)`: Stores a gas price keyed by `gasPrice:<denom>`.
- `LastGasPrices(ctx) []std.GasPrice`: Returns all tracked denoms via prefix iterator.
- `UpdateGasPrice(ctx)`: Called in EndBlock. First seeds any new denoms from `InitialGasPrices` that aren't yet in the store, then adjusts each existing denom's price independently using the EIP-1559-inspired formula: `newPrice = lastPrice + lastPrice*(gasUsed - targetGas) / (targetGas * compressor)`.

New denoms added to `InitialGasPrices` via governance are seeded before the `gasUsed > 0` guard, so they become active immediately on the next EndBlock regardless of whether the block has transactions.

### Ante Handler

`EnsureSufficientMempoolFees` finds the matching denom in the block gas prices slice and checks that the transaction's fee meets or exceeds it. Then checks the node's `minGasPrices` as before.

Transactions must pass both the consensus-level block gas price (per-denom) and the node's local `minGasPrices` config. Node operators must configure `minGasPrices` for all accepted denoms.

### Genesis

`InitChainer` accepts `[]std.GasPrice` and seeds each into the `GasPriceKeeper` store.

### Gas Price Format

Gas prices use the existing multi-denom format supported by `std.ParseGasPrices`:
```
"1ugnot/1000gas;1uphoton/1000gas"
```

### Design Properties

- **Independent price tracking**: Each denom has its own store entry and adjusts independently.
- **Shared gas pressure**: All denoms see the same block gas usage signal. If a block is full, prices rise for all denoms regardless of which denom was used to pay. Block space is fungible.
- **Governance-extensible**: New denominations can be added post-genesis by updating `InitialGasPrices`. The `GasPriceKeeper` automatically seeds new denoms on the next EndBlock.
- **Denom removal**: Removing a denom from `InitialGasPrices` removes its price floor, allowing it to decay toward zero over time. The store entry persists. Full removal would require a `DeleteGasPrice` mechanism if needed in the future.

## Consequences

### Positive

- Chains built on tm2 can accept gas fees in any number of denominations.
- Each denom's price adjusts independently, preventing any single denom from being artificially cheap.
- Minimal code change surface — the core formula and ante handler logic are preserved.

### Negative

- Node operators must configure `minGasPrices` for all accepted denominations.
- All denoms share the same gas pressure signal, which could feel unintuitive.

## References

- [PR #2838: Dynamic gas price keeper](https://github.com/gnolang/gno/pull/2838) — original single-denom implementation
- `std.ParseGasPrices` — existing multi-denom format parser
- EIP-1559 — inspiration for the dynamic gas price adjustment formula
- [gno.land/adr/adr-001-photon-gas-fees.md](../../gno.land/adr/adr-001-photon-gas-fees.md) — gno.land-specific application of this feature
