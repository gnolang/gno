# misc/hf-glue — HIGHLY EXPERIMENTAL hardfork testbed

> ⚠️ **DO NOT MERGE. DO NOT USE IN PRODUCTION.** ⚠️
>
> This directory exists **only** on the `moul/hf-glue-experimental` branch,
> which is the *glue* PR combining [#5411](https://github.com/gnolang/gno/pull/5411)
> (genesis-replay mechanism) and [#5376](https://github.com/gnolang/gno/pull/5376)
> (gnoland-1 chain config) so the hardfork flow can be smoke-tested end-to-end in Docker.
>
> The contents of this folder are throwaway tooling — we are using this branch to
> find gaps in #5411/#5376 and then coordinate fixes **back upstream**, not to ship
> a new blessed workflow.

## What this gives you

A one-command local hardfork: fetch state from a running `gnoland1` (or any source
that `hardfork genesis` accepts), rewrite the validator set to a single key you
control, and run the resulting chain in Docker with its state **persisted on disk**
so you can keep posting txs against it.

```
┌──────────────┐   fetch    ┌──────────────────┐   rewrite val   ┌──────────┐
│ rpc.gno.land │ ─────────► │ genesis.json     │ ──────────────► │ Docker   │
│  (gnoland1)  │            │ + historical txs │                 │ gnoland  │
└──────────────┘            │ + original chain │                 │ single   │
                            │   id for sig vfy │                 │ valdtr   │
                            └──────────────────┘                 └──────────┘
```

## Requirements

- Go 1.24+ (for building `hardfork` / `gnogenesis` locally)
- Docker + `docker compose`
- `jq`, `bash`

## Quick start

```bash
# 1. Fetch state from gnoland1 (defaults to the live RPC) and build genesis.
make fetch

# 2. Generate a single-validator key pair locally, patch genesis to use it.
make init

# 3. Start the node in docker (persists state in ./data/).
make up

# 4. Tail logs.
make logs

# 5. From another terminal, post txs against it:
gnokey maketx ... -remote http://localhost:26657 -chainid gnoland-1

# 6. Stop but keep data:
make down

# 7. Nuclear reset (remove all state + generated genesis):
make reset
```

### Picking a different source chain

Everything is parameterized via environment variables, set in `Makefile.env`
or on the command line:

```bash
SOURCE=http://rpc.test11.testnets.gno.land:443 \
ORIGINAL_CHAIN_ID=test11 \
CHAIN_ID=test11-hf \
make fetch init up
```

| Variable             | Default                        | Meaning |
|----------------------|--------------------------------|---------|
| `SOURCE`             | `http://rpc.gno.land:26657`    | RPC URL / local data dir / exported tarball the `hardfork` tool understands |
| `ORIGINAL_CHAIN_ID`  | `gnoland1`                     | Source chain ID (used for historical signature verification) |
| `CHAIN_ID`           | `gnoland-1`                    | New chain ID after the fork |
| `HALT_HEIGHT`        | *(auto-detect)*                | Height to stop pulling at; empty = latest |
| `VALIDATOR_NAME`     | `hf-glue-local`                | Name baked into the single validator entry |

## Files

| Path                     | Purpose |
|--------------------------|---------|
| `Makefile`               | Entrypoint targets: `fetch`, `init`, `up`, `down`, `logs`, `reset` |
| `scripts/fetch.sh`       | Runs the `misc/hardfork` tool against `$SOURCE` to produce `out/genesis.json` |
| `scripts/init-node.sh`   | Runs `gnoland secrets init` + rewrites validator set in `out/genesis.json` |
| `docker-compose.yml`     | Single `gnoland` service using root `Dockerfile` target `all` |
| `out/`                   | *(gitignored)* generated artifacts — genesis, secrets, node data |

## Known gaps / what we are hunting for

This testbed exists so we can find and file issues against #5411 / #5376. Early
suspects (to be confirmed by running it):

- [ ] Account numbers / sequences preserved across chain ID switch? (see 5411 open items)
- [ ] Auth genesis state carried from source chain, or reset by hardfork tool?
- [ ] Historical tx signatures verify against `original_chain_id` all the way
      through genesis replay — no `chainID` leakage?
- [ ] `GenesisDoc.InitialHeight` correctly picked up so `state.LastBlockHeight`
      is set before the first new block?
- [ ] First block after replay produced at `InitialHeight` exactly?
- [ ] `valoper` fee = 0 carried over via param preservation?
- [ ] Replay time on a real gnoland1 dataset (tune `timeout` in `hardfork test`)

Findings get reported back to #5411 / #5376 as issues/review comments, **not**
fixed on this branch.

## Relation to `hardfork test`

`hardfork test` (added in #5411) does an in-memory smoke-test — node runs, replays
in RAM, exits. That is perfect for CI. **This testbed is the opposite**: persistent
disk state, real Docker node, keeps running, accepts txs so the human can poke it.
