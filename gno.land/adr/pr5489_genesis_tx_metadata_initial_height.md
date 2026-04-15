# ADR: Genesis TX Metadata and Initial Height for Chain Upgrades

## Context

Chain hard forks require replaying historical transactions in a new chain's
genesis. Historical transactions were signed with the old chain's ID; during
genesis replay the ante handler must verify signatures against the chain ID
that was in effect when the tx was originally executed.

A chain may go through multiple upgrades — a genesis could contain transactions
originating from several past chains (e.g. `gnoland1` and `gnoland-1`). Using
a single `OriginalChainID` field is fragile: it assumes all historical txs come
from one chain. Instead, we use a `PastChainIDs` allowlist and a per-tx
`ChainID` so each transaction is verified against its own originating chain ID.

## Decision

### `GnoTxMetadata` extensions

Six fields on `GnoTxMetadata` (populated by the hardfork export tool):

- **`Timestamp`** (`int64`): Unix timestamp of the original block. When
  non-zero, overrides the block header time during replay. Zero means "use
  the genesis block time" — it never clobbers the header with Unix epoch.
- **`BlockHeight`** (`int64`): Original block height of the tx. When > 0,
  the context's block header height is set to this value during replay, and
  the tx goes through the full ante handler (real sig verification, account
  numbers, sequences).
- **`ChainID`** (`string`): Originating chain ID. Used for per-tx chain ID
  override during replay if `ChainID` is in `GnoGenesisState.PastChainIDs`.
- **`Failed`** (`bool`): True if the tx had a non-zero return code on the
  source chain. Failed txs are included in the genesis for sequence tracking
  but are NOT re-executed during replay (skipped to prevent double spends or
  unexpected behavior if VM fixes cause them to succeed).
- **`SignerInfo`** (`[]SignerAccountInfo`): Per-signer account metadata for
  signature verification during replay. Each entry contains:
  - `Address`: the signer's address
  - `AccountNum`: the signer's account number (stable, never changes)
  - `Sequence`: pre-tx sequence (the value used in `GetSignBytes`)

  Before each historical tx is delivered during replay, the replay loop
  force-sets each signer's account number and sequence from `SignerInfo`.
  This ensures signatures verify correctly even if prior txs diverged
  (e.g., due to VM fixes or tx deletions).

  Sequences are determined during export via a single-pass algorithm with
  brute-force recovery: for each sender, a counter starts at 0 and
  increments per successful tx. When failed txs create ambiguity (ante-fail
  doesn't increment sequence, msg-fail does), the next successful tx's
  signature is verified against candidate sequences to resolve the gap.

### `GnoGenesisState` extensions

Two new fields on `GnoGenesisState`:

- **`PastChainIDs`** (`[]string`): Allowlist of chain IDs from which
  historical transactions originated. Only chain IDs present in this slice
  can override the context chain ID during replay. Empty = no overrides.
- **`InitialHeight`** (`int64`): Informational field for tooling. Records the
  block height the new chain should start from. The actual enforcement is at
  the consensus layer via `GenesisDoc.InitialHeight`; this field is not read
  by the app during InitChain.

### `GenesisDoc.InitialHeight` (tm2)

Added to `tm2/pkg/bft/types.GenesisDoc`. When > 1, the consensus `Handshaker`
sets `state.LastBlockHeight = InitialHeight - 1` after `InitChain`, so the
first produced block has height `InitialHeight`. Validated to be non-negative.

### How genesis replay works

1. Genesis txs **without** metadata (or `BlockHeight = 0`) → current genesis
   mode: package deploys, infinite gas, auto-account creation, no sig
   verification.
2. Genesis txs **with** `metadata.BlockHeight > 0` → normal mode: full sig
   verification, real account numbers and sequences.
3. Chain ID override applies only when all three conditions hold:
   `BlockHeight > 0` AND `metadata.ChainID != ""` AND
   `metadata.ChainID ∈ PastChainIDs`.
4. Timestamp override applies when `metadata.Timestamp != 0`.
5. If `SignerInfo` is present, each signer's account number and pre-tx
   sequence are force-set before the tx is delivered. If the account doesn't
   exist, it is created with the specified account number (via
   `NewAccountWithNumber`, which bypasses the auto-increment counter).
6. If `Failed` is true, the tx is skipped (not re-executed). The force-set
   from step 5 ensures the correct sequence state for the next tx. Failed
   txs are included in the genesis for sequence tracking and auditability.
7. After `InitChain`, the consensus layer reads `GenesisDoc.InitialHeight` and
   advances `state.LastBlockHeight` so blocks start at the correct height.

### Key properties

- Standard genesis txs (package deployments, etc.) are unaffected.
- Unrecognised chain IDs are never silently overridden — they fail as expected.
- A genesis spanning multiple past chains works: each tx uses its own chain ID.
- All new fields use `omitempty`; existing genesis files are unaffected.

## Alternatives considered

1. **Re-sign all transactions**: Requires access to all private keys. Not
   feasible.
2. **Skip sig verification entirely**: Reduces security guarantees.
3. **Single `OriginalChainID string`**: Simpler but fragile — assumes all
   historical txs come from one chain. Breaks for multi-hop upgrades
   (chain A → chain B → chain C).
4. **State-level override**: `OriginalChainID` applied uniformly to all
   historical txs. `PastChainIDs` + per-tx `ChainID` is more precise: each tx
   is verified against its own origin.

## Consequences

- Genesis files for chain upgrades will be larger (all historical txs with
  metadata).
- `InitialHeight` is enforced at the consensus layer (`GenesisDoc.InitialHeight`
  → `Handshaker` → `state.LastBlockHeight`). `GnoGenesisState.InitialHeight`
  is informational only — it is not read during `InitChain`.
- Future upgrades from `gnoland-1` to `gnoland-2` can set
  `PastChainIDs: ["gnoland1", "gnoland-1"]` to replay the full history.

## Open items

- ~~Account number preservation~~: **Resolved.** `SignerInfo` metadata
  records each signer's account number and pre-tx sequence. During replay,
  account state is force-set before each tx. If an account doesn't exist,
  `NewAccountWithNumber` creates it with the correct number (bypassing the
  auto-increment counter). Tested end-to-end against gnoland1 (2637 txs,
  0 replay failures).
- End-to-end test with a real chain halt → export → genesis assembly →
  new chain start. (Partially done: export and in-memory replay validated
  against gnoland1. Full multi-validator halt test remains.)
