# PR5377: In-Place Block-Replay Migration

## Context

When the gno.land chain needs a hard fork — for example to change `r/sys/params`,
upgrade GovDAO, or alter gas schedules — the application state DB (`gnolang.db`)
may no longer be compatible with the new binary.  Operators need a way to rebuild
that state deterministically after a coordinated chain halt, without losing the
full transaction history or requiring clients to switch chain IDs.

The chosen strategy is **Scenario A** (in-place block-replay):

- All validators halt at an agreed block height (via `--halt-height` or a
  GovDAO-voted halt).
- Each validator runs the *new* binary with `--migrate`.
- The binary replays all blocks 1→N from the existing block store using the new
  ABCI logic, building a fresh application state.
- The new state DB atomically replaces the old one; the TM state DB and block
  store are untouched.
- The node then restarts normally; the handshaker sees `appHeight == stateHeight`
  and skips replay.

## Decision

### `RunInPlaceMigration` in `gno.land/pkg/gnoland/migration.go`

The public API is a single function:

```go
func RunInPlaceMigration(cfg MigrationConfig) error
```

`MigrationConfig` carries:

| Field | Purpose |
|-------|---------|
| `DataRootDir` | Node data root (parent of `db/`) |
| `GenesisPath` | Path to the new genesis.json |
| `DBBackend` | Backend for block store / state DB (default: PebbleDB) |
| `GenesisOverlay` | Optional hook to patch the genesis doc before replay |
| `NewApp` | Factory that creates the ABCI app given a `dbm.DB` |
| `Logger` | Progress logging |

**Steps performed:**

1. Guard: refuse if `gnolang.bak.db` already exists (previous migration left a
   backup that was never cleaned up or migration is being run twice).
2. Open existing `blockstore.db` (read-only) to determine `haltHeight`.
3. Open existing `state.db` (read-only) — used only to supply validator-set
   information to `BeginBlock` during replay via `getBeginBlockLastCommitInfo`.
4. Load genesis doc; apply `GenesisOverlay` if provided.
5. Create `gnolang-mig.db` via `NewApp`.
6. Send `InitChainSync` to the fresh app.
7. For each block `h = 1..haltHeight`: call `sm.ExecCommitBlock`, which sends
   `BeginBlock + DeliverTx* + EndBlock + Commit` to the ABCI app.
8. On success: `gnolang.db → gnolang.bak.db`, `gnolang-mig.db → gnolang.db`.
9. On failure: remove the partial `gnolang-mig.db`; original state is untouched.

### `--migrate` flag on `gnoland start`

The migration runs *before* the normal app and node are created, so no DB handle
is held open during the swap.  After migration completes, `execStart` continues
normally with the freshly rebuilt `gnolang.db`.

### `GenesisOverlay` hook

Chain-specific migration scripts (e.g., the `gnoland-1` hard fork) can supply a
`GenesisOverlay` to modify the genesis doc before replay — rename chain ID, inject
new governance params, etc. — without embedding migration logic in this package.

### Which DBs are replaced and which are not

| DB | Action |
|----|--------|
| `gnolang.db` (app state) | Replaced by `gnolang-mig.db`; original kept as `gnolang.bak.db` |
| `state.db` (TM consensus state) | Unchanged — `LastBlockHeight`, validators, etc. remain valid |
| `blockstore.db` (block history) | Unchanged — the source of truth for replay |

The TM state DB does not need rebuilding because:
- `LastBlockHeight` in `state.db` already equals `haltHeight` when migration runs.
- Validators are saved there by normal operation; we only *read* them during replay.
- After migration, `appHeight == stateHeight`, so the handshaker takes the "no
  replay needed" path.

## Alternatives considered

### Scenario B — export-then-genesis (height reset)

Export the final state from `gnoland1`, produce a new `genesis.json`, launch
`gnoland2` from height 0.

Rejected because:
- All clients must switch chain ID — worse UX.
- Block heights are reset; any on-chain logic depending on height behaves differently.
- Indexers and explorers lose continuity.

### Migration binary separate from the node

A standalone `gnoland migrate` subcommand was considered instead of a flag on
`gnoland start`.  The flag approach was chosen because it reuses the existing
app-factory wiring in `execStart` (stdlibs dir, pruning config, gas prices) without
duplicating it.

### Auto-detect migration from chain state

The migration could be triggered automatically when `node:p:halt_height` (from
gnolang/gno#5368) matches the block store height, without requiring `--migrate`.
Auto-detection is deferred: the params-based halt height is not yet merged, and
an explicit operator flag is safer for the first deployment (no risk of
accidental replay on a normal restart).

## Key files

| File | Role |
|------|------|
| `gno.land/pkg/gnoland/migration.go` | `RunInPlaceMigration`, `MigrationConfig` |
| `gno.land/pkg/gnoland/migration_test.go` | Guard condition unit tests |
| `gno.land/cmd/gnoland/start.go` | `--migrate` flag; wires `MigrationConfig.NewApp` |
| `tm2/pkg/bft/state/execution.go` | `ExecCommitBlock` — used for per-block replay |
| `tm2/pkg/bft/consensus/replay.go` | `replayBlocks` — reference for the replay pattern |

## Consequences

- Operators can rebuild app state after a halt with a single flag: `gnoland start --migrate`.
- The block store and TM state DB are never written to, so rollback is trivial:
  remove `gnolang.db` and rename `gnolang.bak.db` back.
- Txs that were valid under the old binary may fail under the new one (e.g. gas
  schedule changes); this is expected and must be communicated to operators.
- Migration time is proportional to chain height and ABCI throughput — for a
  long-running chain this may take significant time.  The smoke-test tooling
  (Tool 1 in the chain-upgrade task) should be used to benchmark before
  triggering on a live chain.
- `GenesisOverlay` is the intended extension point; chain-specific migration
  scripts should not modify `migration.go` itself.
