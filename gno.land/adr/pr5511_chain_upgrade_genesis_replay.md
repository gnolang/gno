# PR5511: Chain upgrade genesis replay

## Context

gno.land needs to support in-place chain hardforks: halt the source
chain at some height `H`, export its full state + transaction history,
and start a new chain whose genesis includes all that history so the
new chain can reach the same state by replaying it from scratch. After
replay, the new chain starts producing fresh blocks at height `H + 1`.

Historical transactions were signed against the source chain's
`chain_id`, `account_number`, and `sequence`. For signatures to verify
during replay, all three must be available in their original form — we
can't re-sign because we don't have the private keys.

The tm2-level consensus / state / app changes that enable
`InitialHeight > 1` live in
[tm2/adr/pr5511_initial_height.md](../../tm2/adr/pr5511_initial_height.md).

Earlier design iterations used a single `OriginalChainID` field
(simpler, but fragile across multi-hop upgrades). This ADR describes
the final design with `PastChainIDs` + per-tx `ChainID`.

## Decision

### `GnoTxMetadata` — per-tx replay metadata

Populated by the hardfork export tool (`gnogenesis fork generate`):

- **`Timestamp`** (`int64`) — Unix timestamp of the original block.
  When non-zero, overrides the block header time during replay.
- **`BlockHeight`** (`int64`) — original block height. When `> 0`, the
  ctx's block header height is set to this value during replay, which
  makes the ante handler treat the tx as non-genesis (full sig
  verification, real account numbers, sequences).
- **`ChainID`** (`string`) — originating chain ID. Used for per-tx
  chain-ID override during replay if `ChainID ∈ GnoGenesisState.PastChainIDs`.
- **`Failed`** (`bool`) — true if the tx had a non-zero return code on
  the source chain. Failed txs are included in the genesis for
  sequence-tracking purposes but are NOT re-executed during replay
  (re-executing could double-spend or succeed unexpectedly if a VM fix
  makes them now pass). The replay emits a non-empty `ResponseDeliverTx`
  with an error marker so indexers don't mistake the skip for success.
- **`SignerInfo`** (`[]SignerAccountInfo`) — per-signer `(Address,
  AccountNum, Sequence)`. Before each historical tx is delivered, the
  replay loop force-sets each signer's account number and pre-tx
  sequence from this. If the account doesn't exist yet,
  `auth.NewAccountWithUncheckedNumber` creates it with the specified
  number, bypassing the auto-increment counter. Uniqueness of the
  `(Address, AccountNum)` pair across all SignerInfo entries and
  balance-init accounts is enforced by `validateSignerInfo` as a
  pre-flight check in `loadAppState` (the keeper primitive does not
  re-check; its name now telegraphs the precondition).
- **`GasUsed`**, **`GasWanted`** (`int64`) — source-chain gas; used by
  `GasReplayMode="source"` and the replay report.

### `GnoGenesisState` — genesis-level replay configuration

- **`PastChainIDs`** (`[]string`) — allowlist of chain IDs from which
  historical transactions originated. Only chain IDs in this slice can
  override the context chain ID during replay. `PastChainIDs[0]` is
  also used for sig verification of genesis-mode txs (no metadata or
  `BlockHeight == 0`) when a hardfork is in progress, since those txs
  were signed against the source chain. Empty = no overrides.
  `PastChainIDs` MAY contain the current chain ID — this is valid for
  same-chain-ID hardforks (e.g. minor fork with no external identity
  change). Do NOT add validation that rejects this.
- **`InitialHeight`** (`int64`) — new chain's starting block height.
  Cross-checked against `GenesisDoc.InitialHeight` via
  `RequestInitChain.InitialHeight`; `loadAppState` rejects the genesis
  on divergence.
- **`GasReplayMode`** (`string`) — historical-tx gas metering:
  - `""` or `"strict"` (default) — new VM's gas meter is authoritative.
    Historical txs may fail if gas requirements changed between chains.
  - `"source"` — historical txs (`BlockHeight > 0`) bypass the new VM's
    gas meter via `auth.SkipGasMeteringKey`, preserving source-chain
    outcomes even when gas metering changed. Response records
    `metadata.GasUsed` for audit.

### Sequence recovery algorithm (`gnogenesis fork generate`)

Account numbers come from one RPC call per address at halt height —
they're stable, never change once assigned.

Sequences are harder: they advance through both genesis-mode txs and
successful historical txs, but failed txs sometimes consume a sequence
(msg-fail) and sometimes don't (ante-fail). We can't tell the two apart
without re-verifying the signature, and the tx bytes don't carry a
"sequence used" field.

The tool uses a single-pass algorithm with buffered brute-force:

1. **Initialisation**: for each signer, query their current sequence
   at halt height (`finalSeq`). Brute-force their first successful
   historical tx's signature against `[0, finalSeq]` — the matching
   value is the signer's starting counter (typically `0`, or `N` if
   they had `N` genesis-mode txs).

2. **Forward pass**, block-ordered:
   - **Successful tx, no pending failures**: assign current counter as
     pre-tx sequence, increment counter.
   - **Failed tx**: buffer it.
   - **Successful tx after failed txs from same signer**: brute-force
     the successful tx's signature against `[counter, counter + len(buffer)]`.
     The matching value is this tx's pre-tx sequence; work backwards to
     assign sequences to each buffered failed tx (ante-fails keep the
     counter, msg-fails consume one).

3. **Trailing failures** (no subsequent success): query sequence at
   halt height and diff against the last known counter.

Signature verification is offline — `GetSignBytes` only needs
`chain_id`, `account_number`, `sequence`, fee, msgs, memo, all of
which are available from the tx bytes + metadata.

### Genesis replay flow

1. `InitChain` → `loadAppState` validates
   `state.InitialHeight == req.InitialHeight` (if set) and that
   `GasReplayMode` is recognised.
2. Genesis-mode txs (no metadata or `BlockHeight == 0`) → current
   genesis behaviour: package deploys, infinite gas, auto-account
   creation. Sig-verify against `PastChainIDs[0]` when a hardfork is
   in progress (these txs were signed with the source chain ID).
3. Historical txs (`BlockHeight > 0`) → full ante handler. For each:
   a. Ctx's block header height is overridden to `metadata.BlockHeight`.
   b. If `metadata.ChainID ∈ PastChainIDs`, ctx's chain ID is
      overridden for sig verification.
   c. If `metadata.Timestamp != 0`, ctx's time is overridden.
   d. If `SignerInfo` is present, each signer's account number and
      pre-tx sequence are force-set.
   e. If `GasReplayMode == "source"`, ctx carries
      `auth.SkipGasMeteringKey` so `auth.SetGasMeter` installs an
      infinite gas meter for this tx.
   f. If `Failed`, skip `Deliver` and emit an error-marker response.
      Otherwise `Deliver` normally.
4. At end of loop, the replay report is emitted via logger: summary
   counts (`ok` / `ok_gas_differs` / `failed` / `skipped_failed`) and
   per-failure detail.
5. Consensus advances `state.LastBlockHeight` to
   `GenesisDoc.InitialHeight - 1` so the next block is produced at
   `InitialHeight`.

### Replay report

`replayReport` accumulates per-tx outcomes and emits via the `slog`
logger at the end of `InitChain`. Categories:
- `ok` — succeeded, gas matched source (or no source gas recorded).
- `ok_gas_differs` — succeeded but gas consumption differs from source.
- `failed` — delivery failed (detail logged per-failure).
- `skipped_failed` — marked `Failed` in metadata, correctly skipped.

Summary counts at info level; each failure gets its own warn line with
source height, gas delta, and error. `replayReport.Outcomes()` exposes
the data for tooling that wants to write a structured
`replay-report.json`.

### Tooling — `gnogenesis fork`

Integrated into the existing `gnogenesis` CLI as a subcommand group
(`contribs/gnogenesis/internal/fork/`):

- **`gnogenesis fork generate`** — reads the source (RPC URL / local
  dir / tarball), runs sequence recovery, emits a ready-to-replay
  genesis with `PastChainIDs`, `InitialHeight`, and per-tx metadata.
- **`--patch-realm PKGPATH=SRCDIR`** (repeatable, on `generate`) —
  rewrites the genesis-mode `addpkg` tx for `PKGPATH` in-place with
  files from `SRCDIR` before writing. The source genesis on disk is
  untouched. This is the only way to land a realm code change as part
  of a fork (you can't re-`addpkg` post-deploy).
- **`gnogenesis fork test`** — in-process `InitChain` smoke-test
  against a genesis.json.

A chain-specific wrapper (`misc/deployments/gnoland-1/generate-genesis.sh`)
hardcodes the gnoland1→gnoland-1 chain IDs and delegates to the CLI.

## Alternatives considered

1. **Re-sign all transactions** — requires access to all private keys.
   Not feasible.
2. **Skip sig verification entirely** — reduces security guarantees.
3. **Single `OriginalChainID string`** — simpler but fragile; assumes
   all historical txs come from one chain, breaks for multi-hop
   upgrades (chain A → B → C). `PastChainIDs` + per-tx `ChainID`
   handles the multi-hop case cleanly.
4. **Absorb `hardfork` tool as a standalone CLI in `misc/`** —
   original design, but it's really a genesis-manipulation tool, so
   it lives with its siblings under `gnogenesis`.

## Consequences

- Genesis files for chain upgrades are large (all historical txs with
  metadata). ~192 MB for gnoland1 → gnoland-1 at halt height 704052.
- `GnoGenesisState.InitialHeight`, `GnoGenesisState.GasReplayMode`,
  and all the `GnoTxMetadata` fields use `omitempty`; existing genesis
  files are unaffected.
- Future chain A → B → C upgrades can set
  `PastChainIDs: ["A", "B"]` to replay both predecessors' histories.

## Bugs found and fixed during review

### App layer

- **Failed-tx `ResponseDeliverTx` was empty (looked like success)** —
  now carries an explicit error marker so indexers can distinguish.
- **`state.InitialHeight` wasn't cross-checked against
  `GenesisDoc.InitialHeight`** — `RequestInitChain.InitialHeight` plumbed
  through, `loadAppState` validates match.

### Tooling (`gnogenesis fork`)

- **`applyOverlay` silent no-op** — listed scripts but didn't execute
  them, returned success. Entire overlay mechanism removed (dead code).
- **JSONL export used `encoding/json` instead of amino** — lost
  interface types (`std.Msg`) on round-trip. Both writer and reader
  now use amino.
- **`verifyGenesisFile` failure returned success** — now aborts
  (opt out with `--no-verify`).
- **Zero unit tests for `bruteForceSignerSequence`** — added 10
  table-driven tests.

### Docs linter (side fix to unblock CI)

- Added `staging.gno.land` + `archive.org` to skip list, added retry
  with backoff and HTTP timeout so transient external-host failures
  don't block unrelated PRs.

### `InitChainerConfig.StrictReplay` — opt-in fail-closed boot

Defaults to `false` for backwards compatibility. Hardfork operators set
it to `true` so any non-skipped genesis tx failure aborts `InitChain`
with an error instead of letting the chain boot in a corrupted state
(AppHash diverging from source under `GasReplayMode="strict"`). Skipped
txs (`metadata.Failed = true`, intentional source-chain failures) do
not count as failures. The `replayReport.FailedCount()` accessor exposes
the underlying tally for tooling that wants to gate on it externally.

A complementary BaseApp fix surfaces the real `loadAppState` error
through to the operator: when `InitChainer` returns
`ResponseInitChain.Error`, `BaseApp.InitChain` now returns it cleanly
instead of falling through to a misleading `validators count mismatch`
panic.

## Known unfixed (follow-up PRs)

1. **RPC source has no retry/resume.** A single transient error aborts
   the entire multi-block fetch. Needs exponential backoff +
   checkpointing.
2. **All txs accumulated in memory.** Full tx history is held in a
   single slice — will OOM on large chains. Needs streaming writer.
3. **`queryAccountAtHeight` silent nil.** All error paths return nil
   with no indication; flaky RPC → wrong sequence metadata.

## Validation

End-to-end test via the hf-glue testbed
([#5486](https://github.com/gnolang/gno/pull/5486)): production-sized
hardfork genesis (~192 MB, 2715 historical txs,
`InitialHeight = 704053`) replays with **0 tx failures** and boots a
live `gnoland-1` node producing fresh blocks.
