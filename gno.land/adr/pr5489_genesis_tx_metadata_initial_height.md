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

Seven fields on `GnoTxMetadata` (populated by the hardfork export tool):

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
  unexpected behavior if VM fixes cause them to succeed). The replay emits
  a non-empty `ResponseDeliverTx` with an error marker so downstream
  consumers (indexers, explorers) don't mistake the skip for success.
- **`SignerInfo`** (`[]SignerAccountInfo`): Per-signer account metadata for
  signature verification during replay. Each entry contains:
  - `Address`: the signer's address
  - `AccountNum`: the signer's account number (stable, never changes)
  - `Sequence`: pre-tx sequence (the value used in `GetSignBytes`)

  Before each historical tx is delivered during replay, the replay loop
  force-sets each signer's account number and sequence from `SignerInfo`.
  This ensures signatures verify correctly even if prior txs diverged
  (e.g., due to VM fixes or tx deletions). If the account doesn't exist
  yet, `auth.NewAccountWithNumber` creates it with the specified number,
  bypassing the auto-increment counter.

  Sequences are determined during export via a single-pass algorithm with
  brute-force recovery: for each sender, a counter starts at 0 and
  increments per successful tx. When failed txs create ambiguity (ante-fail
  doesn't increment sequence, msg-fail does), the next successful tx's
  signature is verified against candidate sequences to resolve the gap.
- **`GasUsed`** (`int64`): Gas the tx actually consumed on the source chain.
  Used by `GasReplayMode="source"` (see below) and the replay report.
- **`GasWanted`** (`int64`): Gas the tx requested on the source chain.
  Informational; used by the replay report.

### `GnoGenesisState` extensions

Three new fields on `GnoGenesisState`:

- **`PastChainIDs`** (`[]string`): Allowlist of chain IDs from which
  historical transactions originated. Only chain IDs present in this slice
  can override the context chain ID during replay. Empty = no overrides.
  Genesis-mode txs (no metadata or `BlockHeight == 0`) that were signed
  with the source chain's chain ID verify against `PastChainIDs[0]` when
  a hardfork is in progress.
- **`InitialHeight`** (`int64`): The block height the new chain should
  start from. Must match `GenesisDoc.InitialHeight` (authoritative at the
  consensus layer); `loadAppState` cross-checks this via the
  `RequestInitChain.InitialHeight` field and rejects the genesis on
  divergence. Setting it is optional (zero = don't check); setting it is
  the simpler and recommended path for genesis-generator tools.
- **`GasReplayMode`** (`string`): Controls how historical txs are metered
  during replay:
  - `""` or `"strict"` (default) — new VM's gas meter is authoritative.
    Historical txs may fail if gas requirements changed between chains.
  - `"source"` — historical txs (`metadata.BlockHeight > 0`) bypass the
    new VM's gas meter via `auth.SkipGasMeteringKey`, preserving source-
    chain outcomes even when gas metering changed. Response records
    `metadata.GasUsed` for audit.

### `GenesisDoc.InitialHeight` (tm2)

Added to `tm2/pkg/bft/types.GenesisDoc`. When > 1, the consensus `Handshaker`
sets `state.LastBlockHeight = InitialHeight - 1` after `InitChain`, so the
first produced block has height `InitialHeight`. Validated to be non-negative.

### `RequestInitChain.InitialHeight` (tm2 ABCI)

New field on `abci.RequestInitChain`, populated by the consensus handshaker
from `GenesisDoc.InitialHeight`. Allows the app to cross-check against
`GnoGenesisState.InitialHeight`.

### `auth.SkipGasMeteringKey` (tm2)

Context key that makes `auth.SetGasMeter` install an infinite gas meter
even for non-genesis blocks. Used by `GasReplayMode="source"` to bypass
gas metering for historical txs.

### Replay report

A structured per-tx outcome report is emitted via logger at the end of
`InitChain`. Categories:
- `ok` — tx replayed successfully, gas matched source (or no source gas
  recorded)
- `ok_gas_differs` — tx succeeded but gas consumption differs from source
- `failed` — tx delivery failed during replay (detail logged per-failure)
- `skipped_failed` — tx was marked `Failed` on source, correctly skipped

Summary counts are emitted at info level; each failure also gets its own
warn line with source height, gas delta, and error. Outcomes are exposed
via `replayReport.Outcomes()` for tooling that wants to write a structured
`replay-report.json`.

### Hardfork tooling (`gnogenesis fork`)

Integrated as a subcommand of the existing `gnogenesis` CLI
(`contribs/gnogenesis/internal/fork/`). Subcommands:

- **`gnogenesis fork generate`** — reads the source (RPC URL, local
  data dir, or exported tarball), runs `bruteForceSignerSequence` to
  recover each signer's pre-tx sequence, and emits the new genesis
  populated with `PastChainIDs`, `InitialHeight`, and per-tx metadata.
- **`--patch-realm PKGPATH=SRCDIR`** (repeatable, on `generate`) —
  rewrites the genesis-mode `addpkg` tx for `PKGPATH` in-place with
  files from `SRCDIR` before writing. Source genesis on disk stays
  untouched — the patch lives only in the in-memory `GnoGenesisState`
  used for output. Motivation: you cannot re-`addpkg` post-deploy, so
  patching the original deployment tx is the only way to land a realm
  code change as part of a fork.
- **`gnogenesis fork test`** — in-process genesis replay smoke-test.

### How genesis replay works

1. `InitChain` → `loadAppState` validates `GnoGenesisState.InitialHeight`
   matches `RequestInitChain.InitialHeight` (if the app-level field is set)
   and that `GasReplayMode` is a recognised value.
2. Genesis txs **without** metadata (or `BlockHeight = 0`) → current genesis
   mode: package deploys, infinite gas, auto-account creation, no sig
   verification. Sig verification of genesis-mode txs signed with the
   source chain's chain ID is done against `PastChainIDs[0]` when a
   hardfork is in progress.
3. Genesis txs **with** `metadata.BlockHeight > 0` → normal mode: full sig
   verification, real account numbers and sequences.
4. Chain ID override applies only when all three conditions hold:
   `BlockHeight > 0` AND `metadata.ChainID != ""` AND
   `metadata.ChainID ∈ PastChainIDs`.
5. Timestamp override applies when `metadata.Timestamp != 0`.
6. If `SignerInfo` is present, each signer's account number and pre-tx
   sequence are force-set before the tx is delivered. If the account doesn't
   exist, it is created with the specified account number (via
   `NewAccountWithNumber`, which bypasses the auto-increment counter).
7. If `GasReplayMode == "source"` and `BlockHeight > 0`, the ctx carries
   `auth.SkipGasMeteringKey=true`, so `auth.SetGasMeter` installs an
   infinite gas meter for this tx. Otherwise the new VM's gas meter applies.
8. If `Failed` is true, the tx is skipped (not re-executed) and the
   `ResponseDeliverTx` carries an explicit error marker. The force-set
   from step 6 ensures the correct sequence state for the next tx. Failed
   txs are included in the genesis for sequence tracking and auditability.
9. At the end of the loop, the replay report is emitted via logger with
   summary counts and per-failure detail.
10. After `InitChain`, the consensus layer reads `GenesisDoc.InitialHeight`
    and advances `state.LastBlockHeight` so blocks start at the correct
    height.

### Key properties

- Standard genesis txs (package deployments, etc.) are unaffected.
- Unrecognised chain IDs are never silently overridden — they fail as expected.
- A genesis spanning multiple past chains works: each tx uses its own chain ID.
- All new fields use `omitempty`; existing genesis files are unaffected.
- **Same-chain-ID hardforks are supported.** `PastChainIDs` MAY contain the
  current chain ID (the one in `GenesisDoc.ChainID`). This is the right
  setup when the chain is upgraded in-place without bumping the chain ID
  (e.g. minor fork with no external-facing identity change). Historical txs
  signed with the current chain ID will still verify correctly because
  their `metadata.ChainID` matches and is allowlisted. Do NOT add a
  validation that rejects this case — it is a legitimate configuration.

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

## Bugs found and fixed (PR #5511 review)

### tm2 consensus layer (all fixed)

1. **Fast-sync broken with InitialHeight > 1.** `BlockchainReactor` started
   `BlockPool` at `store.Height()+1 = 1` instead of
   `state.LastBlockHeight+1 = InitialHeight`. Nodes attempting fast-sync
   would request non-existent blocks. Fixed: use state height when store is
   empty.

2. **Validator set / consensus params not saved at InitialHeight.** `saveState`
   only persisted validators when `nextHeight == 1`. With InitialHeight > 1,
   `LoadValidators` failed and `LoadConsensusParams` panicked when processing
   the second block. Fixed: detect first-block generically (not just height 1).

3. **`ValidateBasic` bypass via zeroed `LastBlockID`.** Any block with
   `LastBlockID.IsZero()` could skip `LastCommit` validation. Fixed: only
   skip when commit is also nil/empty (legitimate genesis), and explicitly
   reject zero `LastBlockID` with non-empty commit.

4. **`BaseApp.validateHeight` panicked with InitialHeight > 1** (PR #5540).
   Store version counter (auto-increments from 0) lags block height. First
   block at 101 while store at version 0 → `expected 2, got 102`. Fixed:
   when store version lags, accept the jump as long as height is monotonic.

5. **`BaseApp.Info()` returned store version as `LastBlockHeight`** (PR #5540).
   On restart, handshaker saw `appHeight=1` but `storeHeight=102` and tried
   to replay from height 2. Fixed: prefer persisted header height when it
   records a higher value.

6. **`BaseApp.Info()` panicked on unloaded multistore.** Fixed: guard.

### Hardfork tooling (fixed)

7. **`applyOverlay` silent no-op.** The overlay mechanism listed scripts but
   never executed them, returning success. Fixed: returns an error when
   scripts are found but execution is not implemented.

8. **JSONL serialization used `encoding/json` instead of amino.** Interface
   type info (`std.Msg`) was lost, breaking round-trip. Fixed: both writer
   (`writeTxsJSONL`) and reader (`dirSource.FetchTxs`) now use amino.

9. **`verifyGenesisFile` failure returned success.** Fixed: verification
   failure now returns an error and aborts the tool (use `--no-verify` to
   opt out).

10. **Zero unit tests for `bruteForceSignerSequence`.** Fixed: added 10
    table-driven tests covering boundaries, error cases, multiple key types,
    and tamper detection.

### App-level fixes

11. **Failed-tx `ResponseDeliverTx` was empty (looked like success).**
    Fixed: skipped failed txs now carry an explicit error marker so
    indexers and explorers can distinguish them from successful txs.

12. **`GnoGenesisState.InitialHeight` wasn't cross-checked against
    `GenesisDoc.InitialHeight`.** Fixed: added `InitialHeight` to
    `abci.RequestInitChain` and validate in `loadAppState`.

13. **No gas-change tolerance for historical txs.** If VM gas metering
    changed between chains, replayed txs may exhaust gas and fail even
    though they succeeded on the source chain. Fixed: new `GasReplayMode`
    field on `GnoGenesisState` with `"source"` option that bypasses the
    new VM's gas meter for historical txs via `auth.SkipGasMeteringKey`.

14. **No visibility into replay outcomes.** Fixed: structured replay
    report with per-tx categorization emitted via logger at end of
    `InitChain`.

### Docs infrastructure (fixed — side issue unblocking CI)

15. **Docs linter flaked on transient remote-link failures.** Added
    `staging.gno.land` and `archive.org` to the skip list; added retry
    with backoff and 15s HTTP timeout. Keeps CI green when external hosts
    are temporarily unreachable.

### Known unfixed (follow-up PRs)

16. **RPC source has no retry/resume.** A single transient error aborts the
    entire multi-block fetch. Needs exponential backoff and checkpointing.

17. **All txs accumulated in memory.** The full tx history is held in a
    single slice. Will OOM on large chains. Needs streaming to disk.

18. **`NewAccountWithNumber` has no duplicate check.** See PR comment for
    discussion; preferred approach is a pre-flight validation pass in
    `loadAppState`.

19. **`queryAccountAtHeight` silent nil.** All error paths return nil
    with no indication; flaky RPC → wrong sequence metadata.

## Open items

- ~~Account number preservation~~: **Resolved.** `SignerInfo` metadata
  records each signer's account number and pre-tx sequence. During replay,
  account state is force-set before each tx. If an account doesn't exist,
  `NewAccountWithNumber` creates it with the correct number (bypassing the
  auto-increment counter).
- ~~Replay tolerance for gas-requirement changes~~: **Resolved** via
  `GasReplayMode="source"` (item 13 above).
- ~~Replay report~~: **Resolved** via the structured logger output at end
  of `InitChain` (item 14 above).
- End-to-end test with a real chain halt → export → genesis assembly →
  new chain start: **Validated** via the hf-glue testbed
  (https://github.com/gnolang/gno/pull/5486) against gnoland1 halt@704052
  with 0 / 2715 replay failures. Full multi-validator halt test still a
  follow-up item.
