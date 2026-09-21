# ADR: Valset Updates via VM Params Keeper (v3)

## Context

### The bug we're actually fixing

The v1/v2 valset flow uses an **in-memory event collector** in `EndBlocker`:
the realm fires `ValidatorAdded`/`ValidatorRemoved` events during `DeliverTx`,
the collector buffers them, and at the end of the block `EndBlocker` checks
the buffer to decide whether to call back into the VM (`GetChanges(from, to)`)
and emit consensus updates.

This is wrong in two ways that bite in production:

1. **Lost-on-restart (#5469).** The event collector is process-local. If the
   node shuts down between `DeliverTx` (which already committed the realm-side
   changes to chain state) and `EndBlocker` running for that block, the
   collector is empty after restart. `EndBlocker` sees no events, skips the VM
   query, and the changes — already on chain — never propagate to consensus.
   Until the *next* validator change happens, the consensus valset silently
   diverges from on-chain truth.

2. **Stale-heights diff (#5556 discussion).** When valset changes happen in
   block N and the affected validator goes offline before block N+1,
   `EndBlocker` of block N+1 calls `GetChanges(from, to)` with heights that
   don't capture the right set of changes. The result: changes are computed
   against the wrong reference state and either applied incorrectly or
   dropped entirely.

Both are surface symptoms of the same architectural mistake: **state that
consensus depends on is being driven by ephemeral in-memory signaling
instead of durable chain state.**

### Prior attempts

- **v1**: Valset changes emitted as events, caught by `EndBlocker`. Lost on
  restart.
- **v2**: Events still gate the EndBlocker; on hit, it calls back into the VM
  via `GetChanges(from, to)`. Same lost-on-restart class of bug, plus the
  stale-heights diff issue, plus regex parsing of typed VM response strings.
- **#5469 firstBlock-flag workaround**: forces `GetChanges(lastHeight, lastHeight)`
  on the first block after startup. Recovers on restart, but still queries
  for changes already applied to consensus (one-block delay), and the rest of
  the time still depends on the in-memory collector.
- **#5556 EndTxHook approach**: scans tx events synchronously during
  `DeliverTx` and queries `GetChanges(req.Height, req.Height)` in the same
  block. Closer, but still couples `EndBlocker` to `VMKeeperI` and still
  routes through event-then-query indirection.

All three v1/v2 patches are tactical fixes for the same root issue: the
chain layer is reading derived state via VM callbacks instead of reading
authoritative state directly.

## Decision

Drop event-driven valset signaling entirely. **State that consensus needs
lives in chain state (params), not in events.** This mirrors the standard
Cosmos SDK idiom for chain↔node interaction.

1. The valset realm (`r/sys/validators/v3`) writes changes directly into the
   VM params keeper under realm-scoped keys.
2. `EndBlocker` reads those keys from the params keeper, computes the diff
   between `valset_prev` and `valset_new`, and propagates the changes to
   consensus.

Because params are durable chain state, restart-safety is structural: a
shutdown between `DeliverTx` and `EndBlocker` doesn't lose the pending flag
or the proposed valset. There's no in-memory bridge to drop.

### Params keys (prefix: `vm:p:`)

Standard `vm:p:*` namespace, same as every other vm-module config param.
No realm-scoped indirection: the chain reads from a fixed location, and
only `gno.land/r/sys/params` (via the privileged `sys/params` stdlib) can
write to it.

| Key            | Written by         | Read by    | Description                                                           |
|----------------|--------------------|------------|-----------------------------------------------------------------------|
| `valset_dirty` | realm + EndBlocker | EndBlocker | Flag: realm sets `true` on change; EndBlocker sets `false` after apply |
| `valset_new`   | realm              | EndBlocker | `[]string` — each entry is `"<bech32-pubkey>:<decimal-power>"`         |
| `valset_prev`  | EndBlocker (init: realm) | EndBlocker | Same shape as `valset_new`; the previously applied set          |

Power 0 means removal. Address derives from pubkey on parse; not stored
separately. The `pubkey:power` split keeps a one-line, trivially-parseable
encoding without piling on parallel lists.

### Realm-side writes

The realm imports `gno.land/r/sys/params` and calls thin wrappers
(`SetValsetProposal`, `InitValsetPrev`) that in turn use `sys/params`.
This matches the convention used by other privileged sys realms (halt,
fee_collector). The realm never touches `chain/params` directly.

### Valset diff

A `ValidatorUpdates.UpdatesFrom(v2)` method on `tm2/pkg/bft/abci/types`
computes the minimal diff between two validator sets:
- Additions: in v2 but not prev.
- Removals: in prev but not v2 (emitted with `Power=0`).
- Power changes: in both but with different power.

Output is sorted to give a deterministic order independent of map iteration.

### Validation

`abci.ParseValidatorUpdate(s)` parses entries of the form
`"<pubkey>:<power>"` into `ValidatorUpdates`. Used by both `WillSetParam`
(write path) and `EndBlocker` (read path). It rejects invalid bech32
pubkeys, missing separators, negative powers, and `int64`-overflowing
powers.

The `EndBlocker` additionally filters consensus updates by allowed pubkey
type (per `ConsensusParams.Validator.PubKeyTypeURLs`).

## Alternatives Considered

- **Keep v2 approach**: Simpler for the realm (no params awareness) but
  requires `EndBlocker` to call back into the VM. Rejected because of the
  coupling and fragility.
- **ABCI events with typed payloads**: Would require extending the GnoVM's
  event system with typed values. More invasive; params keeper already exists.
- **Configurable valset realm path**: Earlier draft made `valset_realm_path`
  a vm param so different realms could write to different keyspaces.
  Dropped: only one realm at a time can hold the privilege of writing the
  valset, and that authority is already controlled at the `sys/params`
  layer (which restricts writers). Indirection earned no flexibility.
- **Two parallel typed lists per slot** (`pubkeys[]` + `powers[]`): same
  goal as the chosen format (no bespoke encoding), but doubled the key
  count and added a length-parity invariant. The single
  `"<pubkey>:<power>"` per entry parses with a 1-line `strings.Cut` and
  keeps the slot self-describing.

## Consequences

**Positive**:
- `EndBlocker` no longer needs a `VMKeeperI` reference.
- No regex parsing of VM responses.
- Validation happens at write time (fail fast).
- Standard `vm:p:*` keyspace; consistent with other vm config params.

**Negative / Tradeoffs**:
- The realm must go through `gno.land/r/sys/params` to write valset state.
- The param keys must be kept in sync between the gno helper and `app.go`
  (covered by `TestValsetConstsDoNotDrift`).
- Existing v2 realm/chain state is not migrated (v3 is a fresh start).
