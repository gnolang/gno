# ADR-001: $PHOTON Gas Fees for gno.land Launch

## Status

Implemented

## Context

At gno.land launch, $GNOT will be non-transferrable. However, through ICS (Interchain Security) with atom.one, $PHOTON will be transferrable via IBC. Users who bridge $PHOTON to gno.land need a way to submit transactions without holding $GNOT.

## Decision

Use the tm2 multi-denomination gas fee support (see [tm2/adr/adr-001-multi-denom-gas-fees.md](../../tm2/adr/adr-001-multi-denom-gas-fees.md)) to accept $PHOTON as a gas fee token alongside $GNOT.

### Genesis Configuration

The default genesis gas prices are configured in `gno.land/pkg/gnoland/genesis.go`. For launch, this should include both ugnot and uphoton:
```
"1ugnot/1000gas;1uphoton/1000gas"
```

The `gnogenesis` CLI tool supports setting gas prices in the genesis file:
```
gnogenesis params set auth.initial_gasprices "1ugnot/1000gas;1uphoton/1000gas"
```

### Node Operator Requirements

Node operators must configure `--min-gas-prices` to include all accepted denominations. For example:
```
--min-gas-prices "1ugnot/1000gas;1uphoton/1000gas"
```

Nodes that only configure ugnot will reject transactions paying in uphoton at the mempool level.

### Limitations at Launch

- **Gas fees**: $PHOTON holders can pay gas fees. This is what this ADR addresses.
- **Storage deposits**: Realm deployment and package additions charge per-byte storage deposits in ugnot only. $PHOTON-only users cannot deploy realms until they acquire $GNOT.

### Post-Launch: GnoSwap Integration

After launch, whitelisting GnoSwap (Onbloc's AMM DEX) pool addresses and seeding $GNOT liquidity will allow $PHOTON holders to swap PHOTON for GNOT and become full gno.land users. This is outside the scope of this ADR.

## Consequences

### Positive

- $PHOTON holders can submit transactions on gno.land at launch without needing $GNOT.
- Additional IBC denominations can be added via governance without code changes.

### Negative

- $PHOTON-only users cannot deploy realms or add packages until they acquire $GNOT (via GnoSwap post-launch).
- Node operators need to update their configuration to include uphoton in `--min-gas-prices`.

## References

- [tm2/adr/adr-001-multi-denom-gas-fees.md](../../tm2/adr/adr-001-multi-denom-gas-fees.md) — tm2 protocol-level multi-denom gas fee support
- [GnoSwap](https://gnoswap.io) — planned PHOTON/GNOT swap venue post-launch
