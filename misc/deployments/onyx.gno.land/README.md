# onyx genesis

Builds the **onyx** genesis. Onyx is the testnet on the mainnet line: it runs **mainnet's exact binaries** (the version in [`UPGRADES.md`](./UPGRADES.md)) and is upgraded whenever mainnet is, so mainnet's upgrades are rehearsed here first. Its genesis is mainnet's shape with a testnet's money — a fresh chain, not a hardfork of pearl.

> Launch: **2026-09-28T00:00:00Z**, chain-id `onyx-1`.
>
> The version to run today and every upgrade since launch are in [`UPGRADES.md`](./UPGRADES.md); joining as a validator is [`VALIDATOR.md`](./VALIDATOR.md).

## What onyx contains

- **Balances**: four accounts, typed in `FUNDED_ACCOUNTS` — the web faucet's dispensing account and the faucet-agent's at `1e18` ugnot each (1 trillion GNOT, pearl's ceiling, ~9x under the int64 Coin limit per account), aeddi at the same, and the gpao approvals oracle at 1,000,000 GNOT (a million approvals at gpao's default fee, so it never needs a top-up). Plus exact-burn funding for the genesis-tx fee payers — the deployer, the names admin, and every `gnomod.toml` `[addpkg] creator` in the package set — which land at zero once the genesis txs execute, or at exactly their funded amount when an address is both. **Transfers are open**: no §126 lock, no exemption list; the build asserts both.
- **Packages**: mainnet's `FILTERED_PACKAGES` set, unchanged — `r/sys/...`, `r/gov/...`, `r/gnoland/{blog,wugnot,coins,boards2}/...`, `r/gnops/valopers/...`, `p/onbloc/{uint256,int256,json}/v0`, `r/sys/validators/v0`, `r/nt/grc20reg/v0`, `p/nt/grc20/v0`, `p/nt/grc721/...` — resolved with transitive deps (test deps included) and deployed by the deterministic `GenesisDeployer` key, or by the `[addpkg] creator` a package declares.
- **Governance**: a **sole GovDAO T1 member at genesis — aeddi** (operational key), seeded with nine invitation points by the bootstrap MsgRun, which also locks `dao.UpdateImpl`'s `AllowedDAOs` to `r/gov/dao/impl/v0`. Mainnet's seed, mirrored. The build reads the seeded address back out of the bootstrap source and refuses to build unless it is funded.
- **Validator**: one founding validator — `gno-core-validator-1`, power 60, run by gno-core — with an operator-keyed valoper profile (operator: aeddi) so `r/sys/validators/v0` can manage the set post-genesis. Consensus identity from the 2026-09-25 inputs record, derived from the key at the ceremony and re-checked from the pubkey. Anyone can join afterwards: see `VALIDATOR.md`.
- **Namespaces**: mainnet's seven initial names — `gnoswap`, `onbloc`, `moul`, `aeddi`, `aib`, `samcrew`, `howl` — registered in `r/sys/users` by a genesis MsgRun through the genesis-only `r/sys/users/init.RegisterUser` path (no authority survives genesis; GovDAO can rename or delete any of them via `r/sys/namereg/v0` proposals).
- **Namespace enforcement**: `r/sys/names.Enable` runs as a genesis MsgCall, so name-based deploy authorization is on from block 1. The admin is the gnolang/multisigs `[govdao]` multisig, as on mainnet; the build asserts the configured value against the admin compiled into `r/sys/names/verifier.gno`.
- **Inert packages**: ACTIVE at genesis, exactly as on mainnet — `code_submission_policy=inert`, the gpao approvals oracle (`g1yaee4f…`) as sole `pkg_approvers` entry, `run_submitters` restricted to the seeded T1 member (so only aeddi can `maketx run`; a developer who needs it goes into `RUN_SUBMITTERS_EXTRA` by name), submission charge off. A post-genesis `MsgAddPackage` parks until gpao enables it; genesis replay is exempt.

Not set at genesis, deliberately (chain defaults apply; adjustable post-genesis via GovDAO proposals, see `misc/govdao-scripts/`): CLA and minimum gas price. No vested accounts.

## Quick start

Build from the tree of the version onyx launches on — the [`chain/onyx`](https://github.com/gnolang/gno/releases/tag/chain%2Fonyx) tag, which sits on that version's commit on `chain/mainnet` — not from `master`: the package set is read from `examples/`, and `master`'s already differs from what the network runs. The script builds the binaries from the worktree, assembles the genesis txs, measures fee-payer balances on a temp node, and verifies sha256 of the locked build artifacts.

```bash
./gen-genesis.sh                # full build
./gen-genesis.sh --no-install   # reuse previously built binaries
./gen-genesis.sh --debug        # echo the main pipeline commands
```

Output: `genesis.json` at the root of this directory.

## Directory layout

```
onyx.gno.land/
├── gen-genesis.sh         # Single self-contained pipeline
├── govdao-exec.sh         # Helper for post-genesis governance ops
├── upgrades.json          # The upgrade ledger (source of truth)
├── UPGRADES.md            # Rendered from upgrades.json
├── VALIDATOR.md           # Joining as a validator
├── genesis.json           # Final artifact (produced by the script, gitignored)
│
├── transactions/          # Per-tx directories (meta.json + optional body)
│   ├── base/
│   │   ├── bootstrap/          # Bootstrap MsgRun (GovDAO T1 seed + AllowedDAOs lock)
│   │   └── users-preregister/  # MsgRun registering the initial namespaces
│   └── migration/
│       └── names-enable/  # Genesis MsgCall to names.Enable
│
└── work/                  # Gitignored — generated artifacts
```

## Pipeline

`gen-genesis.sh` is a single-phase script, 9 steps:

1. Resolve script paths and tooling.
2. Verify required tools and every launch value that can fail in seconds: the funded-accounts sheet (format, duplicates, int64 headroom), the bootstrap's T1 member (count, invitation points, funded), the names admin against the tree, the inert-policy values (approver listed and funded, `run_submitters` covering the T1 member).
3. Build binaries from source (`gno`, `gnokey`, `gnoland`, `gnogenesis`).
4. Resolve `FILTERED_PACKAGES` deps, stage them, and `addpkg` them to the genesis.
5. Add the bootstrap and namespace-preregistration MsgRuns from `transactions/base/`.
6. Add the `names.Enable` MsgCall from `transactions/migration/names-enable/`.
7. Build the valoper CSV from `INITIAL_VALSET` + `INITIAL_VALSET_OPERATORS` and add the `valopers.Register` tx (via `gnogenesis fork valoper-seed`).
8. Measure fee-payer balances via a two-pass temp-node run (measure → verify), gated on committed state; a fee payer that is also a funded account is merged (funded + burn).
9. Add the validator + balances, write the bank params (transfers open), reconcile the shipped account count and supply against the typed sheet and the measured burn, run `gnogenesis verify`, move `genesis.json` into place.

The locked artifacts (package list, valoper seed, tx stream, fee-payer sheet, `genesis.json`) are checked against the `CHECKSUMS_DATA` manifest embedded in the script: after the first clean build with final values, paste the printed "not listed" lines into the heredoc to lock the build; any future run producing different bytes fails loudly.

## Transactions folder

Every entry under `transactions/` is a directory containing a `meta.json` (carries the `reason` audit field, a `kind` discriminator, and signing parameters) and optionally a body file. The `txn_dir_to_jsonl` helper in `gen-genesis.sh` converts such a directory into one tx jsonl line, signing via `gnokey` with the deterministic deployer key. `MsgCall` entries support `caller_override`: the caller field is jq-patched post-sign, which the chain trusts at genesis under `--skip-genesis-sig-verification` — used by `names-enable` to satisfy the admin gate without holding the admin key. Every node runs `--skip-genesis-sig-verification` (documented in `VALIDATOR.md`), as on mainnet.
