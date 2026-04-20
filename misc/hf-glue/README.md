# misc/hf-glue — HIGHLY EXPERIMENTAL hardfork testbed

> ⚠️ **DO NOT MERGE. DO NOT USE IN PRODUCTION.** ⚠️
>
> Throwaway integration harness for the hardfork-replay mechanism.
> Depends on (now split into) these PRs:
>
> - [#5511](https://github.com/gnolang/gno/pull/5511) `feat/genesis-replay-upgrade3` — replay engine: `PastChainIDs`, `GnoTxMetadata.{BlockHeight,ChainID,Failed,SignerInfo}`, `InitialHeight`
> - [#5540](https://github.com/gnolang/gno/pull/5540) hardfork-replay improvements — `tm2/sdk` `InitialHeight > 1` fixes, genesis-mode `PastChainIDs[0]` override, `hardfork --patch-realm`
> - [#5533](https://github.com/gnolang/gno/pull/5533) `contribs/tx-archive` — hardfork-replay metadata, `SignerInfo` brute-force resolver, progress log, gas-replay report
> - [#5376](https://github.com/gnolang/gno/pull/5376) gnoland-1 chain config
> - [#5368](https://github.com/gnolang/gno/pull/5368) govDAO halt-height — fuels the `--patch-realm` demo
>
> Findings land in the PRs above, **not** on this branch.

## What this gives you

One command pulls a full source chain and replays it into a single-validator
fork that runs in Docker, serves RPC + gnoweb, and can optionally ship realm
upgrades inside the fork (via `--patch-realm`).

```
┌──────────────┐  1) base genesis  ┌───────────────────────────────┐
│ GitHub       │ ─────────────────►│                               │
│ release      │                   │ misc/hardfork:                │  3) hardfork
│ gnoland1.0   │                   │   assemble genesis.json       │ ───────────►┌──────────┐
└──────────────┘                   │   + PastChainIDs              │             │ Docker   │
                                   │   + InitialHeight             │             │ gnoland  │
┌──────────────┐  2) historical    │   + SignerInfo per tx         │             │ (single  │
│ rpc.gno.land │ ─ txs (batched) ─►│   + single local validator    │             │ validator│
│ (gnoland1)   │  contribs/        │   + optional --patch-realm    │             │ + gnoweb)│
└──────────────┘  tx-archive       └───────────────────────────────┘             └──────────┘
                                                                                      │
                                                                                      ▼
                                                                             http://localhost:26657 (RPC)
                                                                             http://localhost:8888  (gnoweb)
```

End-to-end tested against gnoland1 (halt @ 704052): 2 637 historical txs, 192 MB
output genesis, **0 / 2715 tx failures** on replay, 1:1 render parity vs. prod
gno.land for sampled realms.

## Requirements

- Go 1.24+ (for building `hardfork` / `tx-archive` locally)
- Docker + `docker compose`
- `jq`, `curl`, `bash`

## Quick start

```bash
cd misc/hf-glue

# 1. Pull base genesis (GitHub release by default) + all historical txs from RPC
#    → out/source/config/genesis.json + out/source/txs.jsonl
#    Then assemble the hardfork genesis → out/genesis.json
make fetch       # ~12 min on full gnoland1

# 2. Generate a fresh single-validator identity and patch the genesis to use it.
#    Also writes a config.toml binding RPC + p2p to 0.0.0.0.
make init

# 3. Start the node + gnoweb in Docker.
make up

# RPC at http://localhost:26657
# gnoweb at http://localhost:8888

# Tail logs
make logs

# Post txs against the fork from another terminal
gnokey maketx ... -remote http://localhost:26657 -chainid gnoland-1

# Stop but keep state
make down

# Wipe db + WAL + last-signed state (keeps genesis + keys — lets you re-replay)
make reset-db

# Nuclear reset (wipe everything, including out/)
make reset
```

### Picking a different source chain

Everything is env-driven:

```bash
SOURCE=http://rpc.test11.testnets.gno.land:443 \
RPC_URL=http://rpc.test11.testnets.gno.land:443 \
ORIGINAL_CHAIN_ID=test11 \
CHAIN_ID=test11-hf \
make fetch init up
```

| Variable | Default | Meaning |
|---|---|---|
| `SOURCE` | `https://github.com/gnolang/gno/releases/download/chain/gnoland1.0/genesis.json` | Base genesis. Direct `.json` URL, RPC endpoint (`/genesis` is fetched + unwrapped), or local file. |
| `RPC_URL` | `https://rpc.gno.land` | RPC endpoint used by `tx-archive` to pull historical blocks. |
| `ORIGINAL_CHAIN_ID` | `gnoland1` | Source chain ID — goes into `PastChainIDs` so historical tx sigs verify. |
| `CHAIN_ID` | `gnoland-1` | New chain ID after the fork. |
| `HALT_HEIGHT` | *(auto)* | Block to stop pulling at. Empty = RPC's current latest at start time. |
| `VALIDATOR_NAME` | `hf-glue-local` | Name baked into the single-validator entry in the genesis. |
| `PATCH_REALMS` | `gno.land/r/sys/params=$REPO/examples/gno.land/r/sys/params` | Space-separated `PKGPATH=SRCDIR` entries. Rewrites matching genesis-mode addpkg txs in-memory with files from the given dir — lets you deliver realm upgrades as part of the fork. Empty to disable. |
| `NODE_DIR`, `TXS_JSONL` | *(unset)* | Only used by `make fetch-from-dir` — local data dir alternative to the RPC pull. |

## Delivering a realm upgrade inside the fork

Default `PATCH_REALMS` swaps `r/sys/params` with the repo's examples copy. After
merging [#5368](https://github.com/gnolang/gno/pull/5368), that copy gains
`halt.gno` with `NewSetHaltRequest` — so the fork boots with the govDAO halt
mechanism available, without the realm ever having been redeployed on-chain:

```
$ curl -sG "http://localhost:26657/abci_query?path=%22vm%2Fqfile%22" \
      --data-urlencode "data=gno.land/r/sys/params" \
  | jq -r '.result.response.ResponseBase.Data' | base64 -d
fee_collector.gno
gnomod.toml
halt.gno          ← added by --patch-realm
params.gno
unlock.gno
```

The source genesis on disk stays pristine — patches are applied only during
hardfork assembly, in memory.

## Make targets

| target | what |
|---|---|
| `make fetch` | Pull base genesis + all blocks, assemble `out/genesis.json` |
| `make fetch-from-dir` | Alt: assemble from a local gnoland data dir (requires `NODE_DIR` + `TXS_JSONL`) |
| `make init` | Generate validator secrets + `config.toml` + patch genesis to single validator |
| `make up` | Docker compose up (gnoland + gnoweb) |
| `make down` | Stop containers, keep state |
| `make logs` | Tail `gnoland` logs |
| `make status` | Print `/status` JSON |
| `make reset-db` | Wipe DB + WAL + last-signed state (lets you re-replay without nuking keys) |
| `make reset` | Nuclear — wipe all of `out/` |
| `make smoketest` | In-memory replay via `hardfork test --verbose` (no Docker, no persisted state) |
| `make replay-log` | Same as smoketest, tee full log to `out/replay.log` + summary |
| `make report-replay` | Build categorized `out/REPLAY-REPORT.md` from replay log |
| `make check-state` | Compare running node vs `gno.land` prod, write `out/STATE-REPORT.md` |
| `make reports` | Full pipeline: `replay-log` + `report-replay` + `check-state` |
| `make gen-local-genesis` | (Alternative) Rebuild gnoland1 base genesis locally via `misc/deployments/gnoland1/gen-genesis.sh` instead of downloading |

## Files

| path | purpose |
|---|---|
| `Makefile` | Entrypoint targets above |
| `docker-compose.yml` | `gnoland` + `gnoweb` services, Dockerfile `target=all` image |
| `scripts/fetch.sh` | 3-stage: download genesis, run `tx-archive backup`, assemble with `hardfork genesis` |
| `scripts/fetch-from-dir.sh` | Local-dir alternative (no RPC) |
| `scripts/init-node.sh` | `gnoland secrets init` + `config init` + rewrite validator via `fixvalidator` |
| `scripts/gen-local-genesis.sh` | Calls `misc/deployments/gnoland1/gen-genesis.sh` if you want to rebuild rather than download |
| `scripts/replay-log.sh` | In-process replay + log capture |
| `scripts/report-replay.sh` | Build `REPLAY-REPORT.md` from log |
| `scripts/check-state.sh` | Local vs prod comparison report |
| `fixvalidator/` | Tiny Go helper that overwrites the genesis validator set with a single local key |
| `out/` | *(gitignored)* all generated artifacts — genesis, secrets, node data, reports |

## Status (halt @ 704052, full gnoland1 chain)

What's been validated end-to-end on this branch:

- [x] Account numbers / sequences preserved via `SignerInfo` brute-force resolver
- [x] Historical tx signatures verify via `PastChainIDs` allowlist
- [x] `InitialHeight > 1` handled across consensus, state, store, SDK (`BaseApp.validateHeight`, `BaseApp.Info`, `saveState`)
- [x] First block produced at `InitialHeight` exactly (704053)
- [x] Node restarts from persisted state cleanly
- [x] Genesis-mode + historical tx replay: **0 / 2715 failures** (one unrelated `r/sys/txfees` storage-deposit failure not caused by the replay itself)
- [x] Chain-ID switch `gnoland1` → `gnoland-1` verified on `/status` + on every historical-tx sig verification
- [x] Realm parity vs. prod: `r/sys/names`, `r/sys/users`, `r/gov/dao` (+`:proposals`), `r/gnoland/blog`, `r/gnoland/coins`, `r/gnoland/wugnot` all ✅
- [x] Manfred's account: `account_num=3096261`, `sequence=31` — matches production exactly
- [x] Delivering a realm upgrade as part of the fork via `--patch-realm` (demo: `r/sys/params` gains `halt.gno` from [#5368](https://github.com/gnolang/gno/pull/5368))
- [x] `contribs/tx-archive` pulls the full chain in ~12 min with progress logging

Things the testbed does **not** cover:

- Multi-validator scenarios (we run a single local validator)
- `--skip-failing-genesis-txs` is still enabled; a few gnogenesis txs with
  `msg.Creator ≠ signing key` fail the pubkey-address check and are skipped.
  Production gnoland1 uses the same flag for the same reason.
- Parameter-set preservation (e.g. `valoper` gas-fee=0) is not explicitly
  asserted — relies on `app_state` carrying over.

## Relation to `hardfork test`

`hardfork test` (in `misc/hardfork`) does an in-memory smoke-test — node runs,
replays in RAM, exits. Perfect for CI. **This testbed is the opposite**:
persistent disk state, real Docker node, keeps running, accepts txs, exposes
gnoweb — meant for a human to poke at.

## Reproducing end-to-end

```bash
cd misc/hf-glue
make fetch && make init && make up

# Wait ~30s for replay to finish and the first block to commit.
curl -s http://localhost:26657/status | jq '.result.sync_info.latest_block_height'
# → 704053+

# Verify the patched realm
curl -sG "http://localhost:26657/abci_query?path=%22vm%2Fqfile%22" \
     --data-urlencode "data=gno.land/r/sys/params" \
  | jq -r '.result.response.ResponseBase.Data' | base64 -d | grep halt.gno
# → halt.gno

# Generate reports
make reports
ls -la out/*.md
```
