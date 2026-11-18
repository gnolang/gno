# gnoland-1 — Hard Fork of gnoland1

`gnoland-1` is the upgraded successor to `gnoland1`. It is produced via a
coordinated hard fork at a governance-approved halt height.

## Chain ID change

| Old      | New        |
|----------|------------|
| gnoland1 | gnoland-1  |

The hyphen was added to make the chain ID upgrade-compatible — `gnoland1`
could not support chain upgrades that preserve the chain ID cleanly because
the naming convention conflated chain identity with version. `gnoland-1` is
the permanent base name; future upgrades increment a suffix on a sub-release
tag (e.g., `chain/gnoland-1.1`).

## What changed

This hard fork bundles the following upgrades in one shot:

- **`r/sys/params` halt height** (gnolang/gno#5368): GovDAO can now vote to
  halt the chain at a specific block and enforce a minimum binary version on
  restart. _(awaiting merge)_
- **`r/gnops/valopers` fee = 0**: Registration fee was set to 0 via a GovDAO
  transaction on gnoland1; preserved in genesis replay. No code change needed.
- **Namereg GovDAO whitelist** (gnolang/gno#5293): Namereg now checks GovDAO
  membership before allowing name registration. ✅ _merged_
- **GovDAO scripts** (gnolang/gno#5375): Updated scripts for the new chain. ✅ _merged_

**Not confirmed for this hard fork** (need explicit sign-off from Jae):
- Gas parameter updates (gnolang/gno#5291, #5289, #5274)

## Upgrade workflow

Approach: **Scenario A — genesis tx-replay with InitialHeight preservation**.
All historical txs from gnoland1 are exported with their original block heights
and timestamps, then assembled into the genesis for gnoland-1. The new chain
starts at `initial_height = halt_height + 1`, preserving height continuity.

```
gnoland1 (running)
    │
    ├── [gnoland1.2] Operators rolling-update with halt_height config
    │
    ├── Chain halts at GovDAO-approved height
    │
    ├── Each validator runs migrate-from-gnoland1.sh  ← NOT YET IMPLEMENTED
    │     - tx-archive exports all txs with block height + timestamp
    │     - genesis-assemble produces genesis.json for gnoland-1
    │       (chain_id=gnoland-1, initial_height=halt+1, original_chain_id=gnoland1)
    │
    ├── Validators compare genesis.json SHA-256
    │   (must all match before anyone restarts)
    │
    └── Validators restart with new binary + new genesis
            chain-id: gnoland-1, starts at height halt+1
```

## ⚠️  Migration script not yet written

**The migration script (`migrate-from-gnoland1.sh`) is the critical missing
piece.** Until it exists and has been tested on a dry-run on test12, the
hard fork cannot happen.

Blockers:
- `tx-archive genesis-assemble` command (companion to gnolang/gno#5411)
- `tx-archive` offline export from block store (no live node required)
- Jae's tm2 `GenesisDoc.InitialHeight` port (hard blocker for gnolang/gno#5411)
- test12 dry-run: full halt → export → genesis-assemble → restart

See the TODO block inside `migrate-from-gnoland1.sh` for details.

Dry-run target: test12 (see gnoland1/govdao-scripts/ for tooling).

## GovDAO scripts

The govdao scripts in `govdao-scripts/` are identical to those in
`../gnoland1/govdao-scripts/` but default to `CHAIN_ID=gnoland-1`.

All scripts default to `GNOKEY_NAME=moul`, `CHAIN_ID=gnoland-1`, and
`REMOTE=https://rpc.gno.land:443`. Override via env vars.

```bash
./govdao-scripts/add-validator-from-valopers.sh ADDR
./govdao-scripts/add-validator.sh ADDR PUBKEY [POWER]
./govdao-scripts/rm-validator.sh ADDR
./govdao-scripts/unrestrict-account.sh ADDR [ADDR...]
```

## Config

Copy `config.toml` and edit the `# Change me` fields:

```shell
mkdir -p gnoland-data/config
cp config.toml gnoland-data/config/config.toml
grep -n "Change me" gnoland-data/config/config.toml
```
