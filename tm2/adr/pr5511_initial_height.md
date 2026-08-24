# PR5511: `InitialHeight` support for chain upgrades

## Context

Chain hardforks on gno.land export state and historical transactions from
the source chain and replay them in the new chain's `InitChain`. After
`InitChain`, the new chain must start producing blocks at the source
chain's halt height + 1, not at 1.

Before this PR, tm2 hard-coded "fresh chain starts at height 1" in
multiple places: block store, block pool (fast-sync), validator/consensus
param persistence, block validation, and the app's height-tracking logic.

The gno.land side of this work (genesis replay, `PastChainIDs`,
`GnoTxMetadata`, replay tooling) lives in
[gno.land/adr/pr5511_chain_upgrade_genesis_replay.md](../../gno.land/adr/pr5511_chain_upgrade_genesis_replay.md).

## Decision

### `GenesisDoc.InitialHeight` (`tm2/pkg/bft/types`)

New `int64` field. When `> 1`, the consensus `Handshaker` sets
`state.LastBlockHeight = InitialHeight - 1` after `InitChain`, so the
first produced block has height `InitialHeight`. Validated non-negative
in `GenesisDoc.ValidateAndComplete`. `omitempty` for backwards
compatibility with existing genesis files.

### `abci.RequestInitChain.InitialHeight`

New field on the ABCI struct, populated by the consensus handshaker from
`GenesisDoc.InitialHeight`. Lets the app cross-check against any
app-level `InitialHeight` field in the AppState (gno.land uses this to
reject divergent genesis files).

### `auth.SkipGasMeteringKey` (`tm2/pkg/sdk/auth`)

Context key. When set to `true`, `auth.SetGasMeter` installs an infinite
gas meter even for non-genesis blocks. Used by gno.land's
`GasReplayMode="source"` so historical txs can replay with source-chain
outcomes when gas metering has changed between chains.

### Consensus + state layer fixes

All fixes land together so `InitialHeight > 1` works end-to-end:

- **BlockPool** (`tm2/pkg/bft/blockchain/reactor.go`) — start syncing at
  `state.LastBlockHeight + 1` when the store is empty, not at
  `store.Height() + 1 = 1`.
- **`saveState`** (`tm2/pkg/bft/state/store.go`) — detect the first-block
  case generically (not just `nextHeight == 1`), so validators and
  consensus params are persisted at `InitialHeight` and can be loaded
  when processing `InitialHeight + 1`.
- **`Block.ValidateBasic`** (`tm2/pkg/bft/types/block.go`) — only skip
  `LastCommit` validation when the commit is also nil/empty (the
  legitimate genesis pattern). Explicitly reject a zero `LastBlockID`
  paired with a non-empty commit, which would otherwise bypass commit
  validation.
- **`BaseApp.validateHeight`** (`tm2/pkg/sdk/baseapp.go`) — the
  multistore version counter auto-increments from 0 and lags block
  height by `InitialHeight - 1`. Track real chain height in BaseApp
  via `lastBlockHeight`, recomputed from `multistoreVersion +
  initialHeightOffset` on every `Commit` and on restart in
  `initFromMainStore`. The offset is a chain-wide constant persisted
  in the base store under `mainInitialHeightKey` from `InitChain`.
  Strict contiguity is enforced against the real chain height (no
  permanent allow-jump branch).
- **`BaseApp.Info`** — return the persisted header's `LastBlockHeight`
  when it exceeds the store version, so on restart the handshaker
  doesn't rewind to `InitialHeight` and try to re-replay.
- **`BaseApp.Info`** — guard against a not-yet-loaded multistore to
  avoid a nil-dereference at startup.

## Alternatives considered

1. **Keep hard-coding height-1** and have the app layer translate all
   heights — unworkable, would leak fork semantics into consensus state.
2. **Track `InitialHeight` on state but not in GenesisDoc** — doesn't
   survive node restart, would need a new sidecar file.

## Consequences

- Any chain started with `GenesisDoc.InitialHeight > 1` transparently
  begins producing blocks at `InitialHeight`. All existing paths
  (fast-sync, state bootstrap, block validation, restart) work.
- Existing chains (`InitialHeight` unset or `1`) are unaffected — all
  new fields use `omitempty` and all new code paths are conditional.
- `RequestInitChain.InitialHeight` is opt-in for the app: it can be
  ignored with no downside (behavior matches pre-PR).

## Tests

New unit tests cover each fix in isolation:

- `tm2/pkg/bft/blockchain` — reactor starts pool at correct height.
- `tm2/pkg/bft/state` — validators/params saved at `InitialHeight`.
- `tm2/pkg/bft/types` — `ValidateBasic` rejects tampered blocks.
- `tm2/pkg/bft/consensus` — replay feeds `InitialHeight` to InitChain;
  `Handshaker` sets `state.LastBlockHeight` correctly.
- `tm2/pkg/sdk` — BaseApp `Info` and `validateHeight` handle
  `InitialHeight > 1`.
- `tm2/pkg/sdk/auth` — `SetGasMeter` respects `SkipGasMeteringKey`.

End-to-end validation: a production-sized hardfork genesis
(~2700 txs, `InitialHeight = 704053`) replays and boots a live node
with zero tx failures (see the gno.land ADR for details).
