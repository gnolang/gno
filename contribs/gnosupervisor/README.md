# gnosupervisor

`gnosupervisor` runs a gnoland node across the coordinated upgrades recorded in a chain's upgrade ledger (`misc/deployments/<chain>/upgrades.json`). It starts `gnoland start` on the release that runs the node's next block. When that release stops at its range's halt height, the supervisor starts the next release on the same data directory. A node joining from genesis replays the whole history with one command.

Design and scope: [`gno.land/adr/prxxxx_upgrade_supervisor.md`](../../gno.land/adr/prxxxx_upgrade_supervisor.md). This version handles **historical replay**. A live node stopping at a future halt stays up at the tip; switching it is the next step (see the ADR).

> **Status.** A gnoland-1 replay from genesis, range by range, has not been run yet; the documented joining path is still in [`misc/deployments/mainnet.gno.land/UPGRADES.md`](../../misc/deployments/mainnet.gno.land/UPGRADES.md). Replay on a non-signing node.

## Install

```shell
cd contribs/gnosupervisor
make install    # or: go install .
```

## Run

```shell
gnosupervisor run \
    --ledger   misc/deployments/mainnet.gno.land/upgrades.json \
    --genesis  genesis.json \
    --data-dir gnoland-data \
    --cache-dir gnosupervisor-cache \
    -- --log-level info
```

- `--ledger`: the chain's `upgrades.json`, from a checkout you trust. The supervisor runs what it pins and never fetches it.
- `--genesis`: the genesis file the ledger records. Its `chain_id` and sha256 must match the ledger.
- `--data-dir`: the gnoland data directory, created on first start if missing. Its `config/config.toml` (if any) must have `halt_height = 0` and `skip_upgrade_height = 0`, and use the `pebbledb` backend.
- `--cache-dir`: where release images are extracted, one directory per image digest and platform. It must be owned by the supervisor's user and writable by no one else: the supervisor runs what it holds.
- Arguments after `--` are passed to `gnoland start`. `--data-dir`, `--genesis` and `--gnoroot-dir` are set by the supervisor.
- `--continue-after-halt <H>`: go on, on the same release, after a stop at height H that the ledger does not record (a pause). See below.
- `--platform` (default: the host's), `--registry` (a mirror replacing `https://ghcr.io`), `--stop-timeout`, `--max-restarts`, `--log-format text|json`.

The supervisor takes a lock on the data directory, forwards `SIGINT`/`SIGTERM` to gnoland, and kills it after `--stop-timeout`. Logs go to stderr; gnoland's output is passed through.

## How it picks a release

For the node's last committed height `h` (read from the stopped node's state database), the release for block `h + 1` is the one `upgrades.Ledger.BlockRanges` assigns. If rolling patches were released into that range, the last one is used. On gnoland-1 that gives:

| Blocks | Release |
|---|---|
| 1 – 36300 | v1.2.0 |
| 36301 – 113000 | v1.3.0 |
| 113001 – 162200 | v1.4.0 |
| 162201 – 1183000 | v1.5.1 |
| 1183001 – 1651500 | v1.7.1 |
| 1651501 – | v1.8.0 |

When gnoland exits, the supervisor reads the committed height again:

- at the end of the range, it moves to the next release (a node catching up exits with code 2 there);
- past the end, it stops;
- after code 2 anywhere else, it stops: that is a halt the ledger does not record, which may be an upgrade missing from your copy of the ledger. The stop is recorded in `gnosupervisor.halt` in the data directory, so running again refuses too. Update the ledger, or, if it was a pause, run again with `--continue-after-halt <H>`;
- any other exit restarts the same release with a backoff, up to `--max-restarts` times in a row without progress.

A node that has left block sync before a halt stays up there instead of exiting. The supervisor polls the node's RPC (`rpc.laddr`, which must listen on TCP for this) and stops a node idle at its range's end for 2 minutes. A node at a halt the ledger does not record (read from the chain's `node:p:halt_height` at every new height) is left up, and the stop is recorded in `gnosupervisor.halt` as above. On any range, a node that commits no block for 10 minutes is logged as an error.

## Artifacts

Each release is pulled from its ledger image digest (`ghcr.io/gnolang/gno/gnoland@sha256:…`) over the registry HTTP API, with no container runtime. The index, the platform manifest and every layer are checked against their digests. Only `/usr/bin/gnoland` (a static binary) and `/gnoroot` (the stdlibs of the same commit) are extracted. A cached artifact is re-hashed before each use.

## Containers

Run the supervisor as the container's process, with gnoland as its child: no Docker socket is needed. Give the container no restart policy (see `misc/deployments/mainnet.gno.land/VALIDATOR.md`): restarts belong to the supervisor. Set `stop_grace_period` above `--stop-timeout` (60 s by default; Docker's default grace period is 10 s). Mount the data and cache directories as volumes. Replay on a non-signing node; the lock does not stop a second machine signing with the same key.

## Tests

```shell
make test                        # unit tests and a fake gnoland, no network
```

`internal/supervisor/realchain_test.go` replays a real local chain across a real GovDAO halt:

```shell
scripts/devchain.sh up           # builds two gnoland versions, halts the chain by vote, upgrades it
REALCHAIN_ENV=${TMPDIR:-/tmp}/gnosupervisor-devchain-$(id -u)/run/nodeb.env \
    go test -tags realchain -run TestRealChainReplay -v ./internal/supervisor/
scripts/devchain.sh down --purge
```
