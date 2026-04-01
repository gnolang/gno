# ADR: Chain Upgrade Genesis Replay

## Context

We need to support a hard fork from `gnoland1` to `gnoland-1`. The approach is to export all historical transactions from the old chain, include them in the new chain's genesis with metadata, and replay them during `InitChain`. The new chain then starts at the halted height of the old chain.

Historical transactions were signed with the old chain's ID (`gnoland1`). During genesis replay, the ante handler needs to verify these signatures using the original chain ID, not the new one.

## Decision

Extend `GnoGenesisState` with two new fields:

- **`OriginalChainID`** (`string`): The chain ID of the source chain. When set, historical transactions (those with `metadata.BlockHeight > 0`) are replayed with this chain ID in the context, allowing signature verification to succeed.
- **`InitialHeight`** (`int64`): The block height the new chain should start from after genesis replay. This corresponds to the halt height of the old chain.

Extend `GnoTxMetadata` with:

- **`BlockHeight`** (`int64`): The original block height at which the transaction was included. When greater than zero, the context's block header height is set to this value during replay.

### How it works

1. Historical txs are exported from the old chain with metadata (timestamp, block height).
2. The new genesis includes these txs along with `OriginalChainID` and `InitialHeight`.
3. During `InitChain`, the genesis tx replay loop checks each tx's metadata:
   - If `metadata.BlockHeight > 0`, the block header height is set accordingly.
   - If `metadata.BlockHeight > 0` AND `state.OriginalChainID` is set, the context's chain ID is overridden to the original chain ID.
4. The ante handler sees `BlockHeight > 0` as non-genesis, so it performs full signature verification using account numbers, sequences, and the (overridden) chain ID. This means historical tx signatures verify correctly without modification.

### Key insight

When `header.Height` is set to a non-zero value, the ante handler treats transactions as normal (not genesis), using actual account numbers/sequences and verifying signatures with `ctx.ChainID()`. By setting `ctx.WithChainID(originalChainID)`, the original signatures verify correctly.

## Alternatives considered

1. **Re-sign all transactions**: Would require access to all private keys. Not feasible.
2. **Skip signature verification entirely**: Reduces security guarantees during genesis replay.
3. **Patch the ante handler**: More invasive and harder to maintain.

## Consequences

- Genesis files for chain upgrades will be larger (containing all historical txs with metadata).
- The `InitialHeight` field is informational in the current implementation; full support requires `GenesisDoc.InitialHeight` to be propagated to the consensus layer.
- The `OriginalChainID` override only applies to txs with `BlockHeight > 0`, so standard genesis txs (package deployments, etc.) continue to work normally.
- Both new fields use `omitempty`, so existing genesis files are unaffected.
