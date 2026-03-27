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
  restart.
- **`r/gnops/valopers` fix** (gnolang/gno#5373): Registration fee set to 0,
  unblocking validators who don't hold GNOT.
- **Namereg GovDAO whitelist** (gnolang/gno#5293): Namereg now checks GovDAO
  membership before allowing name registration.
- **Gas parameter updates** (gnolang/gno#5291, #5289, #5274): Updated gas
  schedule. Note: some previously-valid transactions may no longer be valid.

## Upgrade workflow

```
gnoland1 (running)
    │
    ├── Operators rolling-update to new binary (>= chain/gnoland1.1)
    │   using --halt-height or GovDAO proposal
    │
    ├── Chain halts at approved height
    │
    ├── Each validator runs migrate-from-gnoland1.sh  ← TODO: not yet written
    │   to produce gnoland-1/genesis.json
    │
    ├── Validators compare genesis.json SHA-256
    │   (must all match before anyone restarts)
    │
    └── Validators restart with new binary + new genesis
            chain-id: gnoland-1
```

## ⚠️  Migration script not yet written

**The migration script (`migrate-from-gnoland1.sh`) is the critical missing
piece.** Until it exists and has been tested on a dry-run, the hard fork
cannot happen.

See the TODO block inside `migrate-from-gnoland1.sh` for the full list of
what needs to be implemented.

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
