# topaz genesis

Builds the **topaz** genesis. topaz is a **fresh chain** — not a hardfork, no historical replay.

## Why a fresh chain instead of a hardfork

topaz was initially designed as a hardfork of gnoland1 that also replayed test13's own transaction history (from [gnolang/tx-exports](https://github.com/gnolang/tx-exports/tree/main/test13.gno.land)). A full prototype of that pipeline was built and measured, and we abandoned it for three reasons:

- **The genesis ballooned to ~3 GB** (192 MB of gnoland1 replay + ~2.8 GB of test13's 1.26M txs), which every validator would have to download and verify.
- **Patching invalid transactions would take forever.** test13's txs were produced against a months-old master: current master rejects entire package families at replay time (e.g. the `package name must match the last path element` rule kills every `r/gnoswap/*/v1` deploy, cascading to ~1.2M failed calls), plus API drift in `avl`, `chain/runtime`, and others. Each root cause needs a hand-crafted patch, and the long tail is effectively unbounded.
- **Replay would take hours on every node.** The in-process audit replay took ~2 hours on a fast machine; every validator would pay a similar cost at first boot, before producing a single block.

Starting fresh trades all of that for a genesis that builds in minutes and boots in seconds.

## What topaz contains

- **Packages**: the curated `examples/` set (resolved with transitive deps) — `r/sys/...`, `r/gov/...`, `r/gnoland/{blog,wugnot,coins,boards2}/...`, `r/gnops/valopers/...`, `p/onbloc/{uint256,int256,json}`, `r/sys/validators/v3`, `r/demo/defi/grc20reg` — deployed by the deterministic `GenesisDeployer` key.
- **Governance**: aeddi (`g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m`) is the sole GovDAO T1 member, seeded by the bootstrap MsgRun, which also locks `dao.UpdateImpl`'s `AllowedDAOs` to `r/gov/dao/v3/impl`. Additional members join via GovDAO proposals.
- **Validators**: 2 founding validators (`gno-core-val-01/02`, power 60 each) in `GenesisDoc.Validators`. InitChainer seeds `r/sys/params`' `valset:current` from it, so `r/sys/validators/v3` proposals manage the set post-genesis. Each founder has a valoper profile (registered at genesis via `gnogenesis fork valoper-seed`) keyed on an operator address, giving them the operator-keyed management plane (signing-key rotation, profile edits, opt-out).
- **Namespace enforcement**: `r/sys/names.Enable` runs as a genesis MsgCall, so name-based deploy authorization is on from block 1. The call's caller field is patched to the admin address hardcoded in `r/sys/names/verifier.gno` (trusted under `--skip-genesis-sig-verification`; the private key is not needed).
- **Balances**: the 10 faucets at 1e18 ugnot (≈1T GNOT) each, plus exact-burn funding for the genesis-tx fee payers (deployer + names admin), which land at zero once the genesis txs execute. No airdrop, no inherited balances.
- **Transfers**: unrestricted — no bank lock, no unrestricted-accounts list.

Not set at genesis (defaults apply; adjustable post-genesis via GovDAO proposals, see `misc/govdao-scripts/`): CLA, minimum fee.

To run a full node and put yourself forward as a validator on topaz, see [`VALIDATOR.md`](./VALIDATOR.md).

## Quick start

The script is fully self-contained: builds the binaries from the worktree, assembles the genesis txs, measures fee-payer balances on a temp node, and verifies sha256 of every produced artifact.

```bash
./gen-genesis.sh                # full build — a few minutes
./gen-genesis.sh --no-install   # reuse previously built binaries
./gen-genesis.sh --debug        # echo every external command
```

Output: `genesis.json` at the root of this directory.

## Directory layout

```
topaz.gno.land/
├── gen-genesis.sh         # Single self-contained pipeline
├── govdao-exec.sh         # Helper for post-genesis governance ops
├── genesis.json           # Final artifact (produced by the script)
│
├── transactions/          # Per-tx directories (meta.json + optional body)
│   ├── base/
│   │   └── bootstrap/     # Bootstrap MsgRun (GovDAO T1 seed + AllowedDAOs lock)
│   └── migration/
│       └── names-enable/  # Genesis MsgCall to names.Enable
│
└── work/                  # Gitignored — generated artifacts
```

## Pipeline

`gen-genesis.sh` is a single-phase script, 9 steps:

1. Resolve script paths and tooling.
2. Verify required tools (preflight with `brew` + `apt` install hints).
3. Build binaries from source (`gno`, `gnokey`, `gnoland`, `gnogenesis`).
4. Resolve `FILTERED_PACKAGES` deps, stage them, and `addpkg` them to the genesis.
5. Add the bootstrap MsgRun from `transactions/base/bootstrap/`.
6. Add the `names.Enable` MsgCall from `transactions/migration/names-enable/`.
7. Build the valoper CSV from `INITIAL_VALSET` + `INITIAL_VALSET_OPERATORS` and add the `valopers.Register` txs (via `gnogenesis fork valoper-seed`).
8. Measure fee-payer balances via a two-pass temp-node run (measure → verify zero).
9. Add the 2 validators + balances (fee payers + 10 faucets), run `gnogenesis verify`, move `genesis.json` into place.

Every produced artifact is checked against the `CHECKSUMS_DATA` manifest embedded in the script: after the first clean build, paste the printed "not listed" lines into the heredoc to lock the build; any future run producing different bytes fails loudly.

## Transactions folder

Every entry under `transactions/` is a directory containing a `meta.json` (carries the `reason` audit field, a `kind` discriminator, and signing parameters) and optionally a body file. The `txn_dir_to_jsonl` helper in `gen-genesis.sh` converts such a directory into one tx jsonl line, signing via `gnokey` with the deterministic deployer key. `MsgCall` entries support `caller_override`: the caller field is jq-patched post-sign, which the chain trusts at genesis under `--skip-genesis-sig-verification` — used by `names-enable` to satisfy the admin gate without holding the admin key.
