# ADR: Genesis TX Metadata and Initial Height for Chain Upgrades

## Context

Chain hard forks require replaying historical transactions in a new chain's genesis. Historical transactions were signed with the old chain's ID; during genesis replay the ante handler must verify signatures against the chain ID that was in effect when the tx was originally executed.

A chain may go through multiple upgrades — a genesis could contain transactions originating from several past chains (e.g. `gnoland1` and `gnoland-1`). Using a single `OriginalChainID` field is fragile: it assumes all historical txs come from one chain. Instead, we use a `PastChainIDs` allowlist and a per-tx `ChainID` so each transaction is verified against its own originating chain ID.

## Decision

### `GnoGenesisState` extensions

Two new fields on `GnoGenesisState`:

- **`PastChainIDs`** (`[]string`): Allowlist of chain IDs from which historical transactions in this genesis originated. Only chain IDs present in this slice can be used for the chain ID override during replay.
- **`InitialHeight`** (`int64`): The block height the new chain should start from after genesis replay. This corresponds to the halt height of the old chain + 1.

### `GnoTxMetadata` extensions

Three fields on `GnoTxMetadata` (populated by tx-archive export):

- **`Timestamp`** (`int64`): Unix timestamp of the original block (pre-existing field).
- **`BlockHeight`** (`int64`): The original block height at which the transaction was included. When greater than zero, the context's block header height is set to this value during replay, and the tx goes through the normal ante handler (full sig verification).
- **`ChainID`** (`string`): The originating chain ID for this transaction. Used for the per-tx chain ID override during replay if `ChainID` is in `GnoGenesisState.PastChainIDs`.

### `GenesisDoc` extension

- **`InitialHeight`** (`int64`): Added to `tm2/pkg/bft/types.GenesisDoc`. When greater than 1, the consensus `Handshaker` sets `state.LastBlockHeight = InitialHeight - 1` after `InitChain`, so the first produced block has height `InitialHeight`. Validated to be non-negative.

### How it works

1. Historical txs are exported from the old chain with metadata (timestamp, block height, chain ID).
2. The new genesis includes these txs along with `PastChainIDs` (the allowlist) and `InitialHeight`.
3. During `InitChain`, the genesis tx replay loop checks each tx's metadata:
   - If `metadata.BlockHeight > 0`, the block header height is set to `metadata.BlockHeight`.
   - If `metadata.BlockHeight > 0` AND `metadata.ChainID != ""` AND `metadata.ChainID` is in `state.PastChainIDs`, the context's chain ID is overridden to `metadata.ChainID` for that tx's sig verification.
   - If `metadata.BlockHeight == 0` (or no metadata), normal genesis mode applies (no chain ID override, no sig verification for package deploys).
4. The ante handler sees `BlockHeight > 0` as non-genesis, performing full signature verification using account numbers, sequences, and the (possibly overridden) chain ID.
5. After `InitChain`, the consensus layer reads `GenesisDoc.InitialHeight` and advances `state.LastBlockHeight` so the chain starts producing blocks at the correct height.

### Key insight

The override is guarded by three conditions: `BlockHeight > 0` AND `metadata.ChainID != ""` AND `metadata.ChainID ∈ PastChainIDs`. This means:
- Standard genesis txs (package deployments, setup) are unaffected.
- Historical txs with an unrecognised chain ID are not silently overridden — they fail as expected.
- A genesis spanning multiple past chains works correctly: each tx uses its own chain ID.

## Alternatives considered

1. **Re-sign all transactions**: Would require access to all private keys. Not feasible.
2. **Skip signature verification entirely**: Reduces security guarantees during genesis replay.
3. **Single `OriginalChainID` field**: Simpler but fragile — assumes all historical txs come from one chain. Breaks for multi-hop upgrades (chain A → chain B → chain C).
4. **State-level override (old design)**: `OriginalChainID` applied to all historical txs regardless of their actual origin. `PastChainIDs` + per-tx `ChainID` is more precise and more extensible.

## Consequences

- Genesis files for chain upgrades will be larger (containing all historical txs with metadata).
- `InitialHeight` is implemented end-to-end: `GenesisDoc.InitialHeight` → consensus `Handshaker` → `state.LastBlockHeight`. The chain starts producing blocks at `InitialHeight` after genesis replay.
- The chain ID override only applies to txs satisfying all three conditions, so standard genesis txs continue to work normally.
- All new fields use `omitempty`, so existing genesis files are unaffected.
- `GenesisDoc.InitialHeight` is validated to be non-negative.
- Future upgrades from `gnoland-1` to `gnoland-2` can include `PastChainIDs: ["gnoland1", "gnoland-1"]` to replay the full history.

## Open items

- Account number preservation: accounts are currently auto-assigned during balance initialization. If the old chain had different account numbers, some txs may fail replay. Workaround: ensure genesis balances are ordered so account numbers align.
- End-to-end test with a real chain halt → export → genesis assembly → new chain start.
