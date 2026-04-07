# gnobr — gno block rollback

Roll back a gnoland node to a target height and replay blocks locally. No network access needed.

## Usage

```bash
gnobr --data-dir gnoland-data --drop-after <height> [--app-hash <hex>]
```

### Flags

| Flag | Description |
|---|---|
| `--data-dir` | Path to gnoland data directory (default: `gnoland-data`) |
| `--drop-after` | Keep blocks up to this height, drop everything after |
| `--app-hash` | Override the app hash written into state.db (hex); auto-detected if omitted |
| `--dry-run` | Show what would be done without modifying anything |

### What it does

1. Trims `blockstore.db` — removes all blocks after the target height
2. Patches `state.db` — sets LastBlockHeight, LastBlockID, LastBlockTime, LastBlockTotalTx, LastResultsHash, and AppHash to match the target block
3. Wipes `gnolang.db` — forces the app to replay from genesis
4. Wipes WAL
5. Resets `priv_validator_state.json`

On restart, gnoland's Handshaker sees `appHeight=0` but `storeHeight=N` and `stateHeight=N`, runs InitChain, then replays all N blocks from the local blockstore.

## Normal rollback (same binary)

When rolling back due to a non-determinism bug or AppHash divergence with the same binary:

```bash
gnobr --data-dir gnoland-data --drop-after <height>
```

gnobr auto-detects the correct app hash from block `<height+1>`'s header in the blockstore (that header carries the committed state hash after `<height>`). No `--app-hash` flag needed.

## App hash breaking changes

When the gnoland binary changes in a way that affects state computation (new gas model, VM changes, store refactor, etc.), replaying the same blocks with the new binary produces a **different app hash** than the original chain.

gnobr auto-detects the **old** app hash from the blockstore, but the new binary computes a different one. Use a two-pass approach to find the new hash:

### Pass 1: replay and extract the new hash

```bash
gnobr --data-dir gnoland-data --drop-after <height>

# Start gnoland — it will replay all blocks then panic:
#   "state.AppHash does not match AppHash after replay.
#    Got <NEW_HASH>, expected <OLD_HASH>"
#
# The "Got" value is the correct hash for the new binary.
```

### Pass 2: patch with the correct hash

```bash
# --app-hash is required here: block <height+1> was already trimmed by pass 1,
# so auto-detect is unavailable.
gnobr --data-dir gnoland-data --drop-after <height> --app-hash <NEW_HASH>

# Restart gnoland — replay succeeds, node is ready.
```

The second pass is fast since gnobr is idempotent — it skips blockstore trimming if already done and only patches state.db.

## Build

```bash
cd contribs/gnobr
go build -o gnobr .
```
