# ADR: Valset Updates via VM Params Keeper (v3)

## Context

This PR introduces the third iteration of on-chain validator set management
(`r/sys/validators/v3`). Previous iterations:

- **v1**: Valset changes emitted as events, caught by `EndBlocker`.
- **v2**: Events triggered a VM query in `EndBlocker` to scrape on-chain state
  via `GetChanges(from, to)` — still event-driven.

Both v1 and v2 required the `EndBlocker` to:
1. Listen to on-chain events via an event collector (`collector[validatorUpdate]`).
2. Call back into the GnoVM to fetch the actual changes (v2), or parse event
   payloads (v1).

Problems with those approaches:
- **Coupling**: The `EndBlocker` needed a `VMKeeperI` reference solely to query
  the valset realm.
- **Fragility**: Regex-based parsing of typed GnoVM response strings.
- **Indirection**: An event triggers a VM query, which returns data that was
  already computed on-chain.

## Decision

Replace the event-based approach with a **params-keeper-based** approach:

1. The valset realm (`r/sys/validators/v3`) writes changes directly into the
   VM params keeper under realm-scoped keys.
2. `EndBlocker` reads those keys from the params keeper, computes the diff
   between `valset_prev` and `valset_new`, and propagates the changes to
   consensus.

### Params keys (prefix: `vm:gno.land/r/sys/validators/v3:`)

| Key                    | Written by  | Read by     | Description                                |
|------------------------|-------------|-------------|--------------------------------------------|
| `new_updates_available`| realm       | EndBlocker  | Flag: set true when valset changed         |
| `valset_new`           | realm       | EndBlocker  | Serialized proposed valset                 |
| `valset_prev`          | EndBlocker  | EndBlocker  | Serialized previously applied valset       |

Serialization format: `<bech32-address>:<bech32-pubkey>:<uint64-voting-power>`

### Valset diff

A new `ValidatorUpdates.UpdatesFrom(v2)` method on `tm2/pkg/bft/abci/types`
computes the minimal diff between two validator sets:
- Additions: in v2 but not prev.
- Removals: in prev but not v2 (emitted with `Power=0`).
- Power changes: in both but with different power.

### Validation

`WillSetParam` in `VMKeeper` validates `valset_new` updates at write time,
ensuring each entry is well-formed (address/pubkey match, valid power).

The `EndBlocker` still filters out updates with disallowed pubkey types.

### Active valset realm path

The realm path is configurable via `vm:p:valset_realm_path` (default:
`gno.land/r/sys/validators/v3`). This allows future upgrades without changing
the `EndBlocker` code.

## Alternatives Considered

- **Keep v2 approach**: Simpler for the realm (no params awareness) but
  requires `EndBlocker` to call back into the VM. Rejected because of the
  coupling and fragility.
- **ABCI events with typed payloads**: Would require extending the GnoVM's
  event system with typed values. More invasive; params keeper already exists.

## Consequences

**Positive**:
- `EndBlocker` no longer needs a `VMKeeperI` reference.
- No regex parsing of VM responses.
- Validation happens at write time (fail fast).
- Realm path is configurable.

**Negative / Tradeoffs**:
- The realm must know about the params keeper API.
- The param keys must be kept in sync between the realm and `app.go`.
- Existing v2 realm/chain state is not migrated (v3 is a fresh start).
