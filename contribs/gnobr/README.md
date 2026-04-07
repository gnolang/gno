# gnobr — gno block rollback

Roll back a gnoland node to a target height and replay blocks locally. No network access needed.

## Usage

```bash
gnobr --data-dir gnoland-data --drop-after <height> --app-hash <hex>
```

### Flags

| Flag | Description |
|---|---|
| `--data-dir` | Path to gnoland data directory (default: `gnoland-data`) |
| `--drop-after` | Keep blocks up to this height, drop everything after |
| `--app-hash` | Hex-encoded app hash to write into state.db |
| `--dry-run` | Show what would be done without modifying anything |

### What it does

1. Trims `blockstore.db` — removes all blocks after the target height
2. Patches `state.db` — sets LastBlockHeight, LastBlockID, LastBlockTime, LastBlockTotalTx, LastResultsHash, and AppHash to match the target block
3. Wipes `gnolang.db` — forces the app to replay from genesis
4. Wipes WAL
5. Resets `priv_validator_state.json`

On restart, gnoland's Handshaker sees `appHeight=0` but `storeHeight=N` and `stateHeight=N`, runs InitChain, then replays all N blocks from the local blockstore.

## When the app hash is known

If all validators agree on the app hash (e.g. the chain halted due to a non-determinism bug but the correct hash is known), pass it directly:

```bash
gnobr --data-dir gnoland-data --drop-after 352921 \
  --app-hash 311BB98564478295A971AE5481D8F58B12A6D945A63883330BD0A2788A873132
```

## App hash breaking changes

When the gnoland binary changes in a way that affects state computation (new gas model, VM changes, store refactor, etc.), replaying the same blocks with the new binary produces a **different app hash** than the original chain. In this case, you can't know the correct hash upfront.

Use a two-pass approach:

### Pass 1: replay and extract the new hash

```bash
# Run gnobr without --app-hash (or with any placeholder)
gnobr --data-dir gnoland-data --drop-after <height> \
  --app-hash 0000000000000000000000000000000000000000000000000000000000000000

# Start gnoland — it will replay all blocks then panic:
#   "state.AppHash does not match AppHash after replay.
#    Got <NEW_HASH>, expected 0000..."
#
# The "Got" value is the correct hash for the new binary.
```

### Pass 2: patch with the correct hash

```bash
# Re-run gnobr with the hash from the panic message
gnobr --data-dir gnoland-data --drop-after <height> \
  --app-hash <NEW_HASH>

# Restart gnoland — replay succeeds, node is ready.
```

The second pass is fast since gnobr is idempotent — it skips blockstore trimming if already done and only patches state.db.

## Build

```bash
cd contribs/gnobr
go build -o gnobr .
```
