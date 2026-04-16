# misc/hf-glue — HIGHLY EXPERIMENTAL hardfork testbed

> ⚠️ **DO NOT MERGE. DO NOT USE IN PRODUCTION.** ⚠️
>
> This directory is throwaway tooling for smoke-testing the hardfork flow
> end-to-end in Docker. It depends on
> [#5511](https://github.com/gnolang/gno/pull/5511) (chain hardfork mechanism v3:
> `PastChainIDs`, `SignerInfo` metadata, `InitialHeight` fixes) and
> [#5376](https://github.com/gnolang/gno/pull/5376) (gnoland-1 chain config).
>
> Findings get reported back upstream, **not** fixed on this branch.

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
| `SOURCE`             | `https://rpc.betanet.testnets.gno.land` | RPC URL / local data dir / exported tarball the `hardfork` tool understands |
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

Items addressed by #5511 (should now work):

- [x] Account numbers / sequences preserved — `SignerInfo` metadata carries per-signer state
- [x] Historical tx signatures verify via `PastChainIDs` allowlist (no `chainID` leakage)
- [x] `GenesisDoc.InitialHeight` correctly handled across consensus, state, store, SDK

Still to validate end-to-end in Docker:

- [ ] Full genesis replay completes without `--skip-failing-genesis-txs`
- [ ] First block after replay produced at `InitialHeight` exactly
- [ ] New txs can be posted against the forked chain with the new chain ID
- [ ] `valoper` fee = 0 carried over via param preservation
- [ ] Replay time on a real gnoland1/betanet dataset (feasibility)
- [ ] Node can restart from persisted state after replay

## Relation to `hardfork test`

`hardfork test` (added in #5411) does an in-memory smoke-test — node runs, replays
in RAM, exits. That is perfect for CI. **This testbed is the opposite**: persistent
disk state, real Docker node, keeps running, accepts txs so the human can poke it.
