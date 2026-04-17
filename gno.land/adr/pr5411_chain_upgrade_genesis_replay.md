# ADR: Chain Upgrade Genesis Replay

> **Superseded by [pr5489_genesis_tx_metadata_initial_height.md](pr5489_genesis_tx_metadata_initial_height.md).**
> This ADR describes the initial design with a single `OriginalChainID` field.
> The implementation evolved to use `PastChainIDs []string` with per-tx chain ID
> override and `SignerInfo` for account state. See the successor ADR for the
> current design.

## Context

We need to support a hard fork from `gnoland1` to `gnoland-1`. The approach is to export all historical transactions from the old chain, include them in the new chain's genesis with metadata, and replay them during `InitChain`. The new chain then starts at the halted height of the old chain.

Historical transactions were signed with the old chain's ID (`gnoland1`). During genesis replay, the ante handler needs to verify these signatures using the original chain ID, not the new one.

## Decision

### `GnoGenesisState` extensions

Two new fields on `GnoGenesisState`:

- **`OriginalChainID`** (`string`): The chain ID of the source chain. When set, historical transactions (those with `metadata.BlockHeight > 0`) are replayed with this chain ID in the context, allowing signature verification to succeed.
- **`InitialHeight`** (`int64`): The block height the new chain should start from after genesis replay. This corresponds to the halt height of the old chain + 1.

### `GnoTxMetadata` extensions

Three fields on `GnoTxMetadata` (populated by tx-archive export):

- **`Timestamp`** (`int64`): Unix timestamp of the original block (pre-existing field).
- **`BlockHeight`** (`int64`): The original block height at which the transaction was included. When greater than zero, the context's block header height is set to this value during replay.
- **`ChainID`** (`string`): The originating chain ID for this transaction. Informational — used by tx-archive to record provenance; the actual chain ID override during replay uses `GnoGenesisState.OriginalChainID`.

### `GenesisDoc` extension

- **`InitialHeight`** (`int64`): Added to `tm2/pkg/bft/types.GenesisDoc`. When greater than 1, the consensus `Handshaker` sets `state.LastBlockHeight = InitialHeight - 1` after `InitChain`, so the first produced block has height `InitialHeight`. Validated to be non-negative.

### How it works

1. Historical txs are exported from the old chain with metadata (timestamp, block height, chain ID).
2. The new genesis includes these txs along with `OriginalChainID` and `InitialHeight`.
3. During `InitChain`, the genesis tx replay loop checks each tx's metadata:
   - If `metadata.BlockHeight > 0`, the block header height is set accordingly.
   - If `metadata.BlockHeight > 0` AND `state.OriginalChainID` is set, the context's chain ID is overridden to the original chain ID.
   - If `metadata.BlockHeight == 0` (or no metadata), normal genesis mode applies (no chain ID override, no sig verification for package deploys).
4. The ante handler sees `BlockHeight > 0` as non-genesis, so it performs full signature verification using account numbers, sequences, and the (overridden) chain ID. This means historical tx signatures verify correctly without modification.
5. After `InitChain`, the consensus layer reads `GenesisDoc.InitialHeight` and advances `state.LastBlockHeight` so the chain starts producing blocks at the correct height.

### Key insight

When `header.Height` is set to a non-zero value, the ante handler treats transactions as normal (not genesis), using actual account numbers/sequences and verifying signatures with `ctx.ChainID()`. By setting `ctx.WithChainID(originalChainID)`, the original signatures verify correctly.

The chain ID override is intentionally guarded by both conditions (`BlockHeight > 0` AND `OriginalChainID != ""`), so:
- Standard genesis txs (package deployments, setup) are unaffected.
- Historical txs without an original chain ID use the new chain's context.

## Alternatives considered

1. **Re-sign all transactions**: Would require access to all private keys. Not feasible.
2. **Skip signature verification entirely**: Reduces security guarantees during genesis replay.
3. **Patch the ante handler**: More invasive and harder to maintain.
4. **Per-tx chain ID override** (using `GnoTxMetadata.ChainID`): We chose a state-level `OriginalChainID` instead. All historical txs in a hard fork come from the same source chain, so a single override is simpler and less error-prone.

## Consequences

- Genesis files for chain upgrades will be larger (containing all historical txs with metadata).
- `InitialHeight` is implemented end-to-end: `GenesisDoc.InitialHeight` → consensus `Handshaker` → `state.LastBlockHeight`. The chain starts producing blocks at `InitialHeight` after genesis replay.
- The `OriginalChainID` override only applies to txs with `BlockHeight > 0`, so standard genesis txs (package deployments, etc.) continue to work normally.
- All new fields use `omitempty`, so existing genesis files are unaffected.
- `GenesisDoc.InitialHeight` is validated to be non-negative.

## Open items

- Account number preservation: accounts are currently auto-assigned during balance initialization. If the old chain had different account numbers, some txs may fail replay. Workaround: ensure genesis balances are ordered so account numbers align.
- End-to-end test with a real chain halt → export → genesis assembly → new chain start.
