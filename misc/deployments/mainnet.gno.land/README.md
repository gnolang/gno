# mainnet genesis (WIP)

Builds the **gno.land mainnet** genesis. Mainnet is a **fresh chain** — not a hardfork of betanet (gnoland1) — whose balances come from the audited [gnolang/independence-day](https://github.com/gnolang/independence-day) allocation.

> **Status: WORK IN PROGRESS — not launchable.** Chain-id, genesis time, the validator set (all four entries are throwaway keys), three T1 member keys, the vesting list, the inert-package params, and the transfer policy are all open — grep `TODO(mainnet)` in this folder for the authoritative list. `CHECKSUMS_DATA` stays unlocked until every value is final, and the independence-day pin moves with that repo until the allocation freeze.

## What mainnet contains

- **Balances**: the independence-day allocation — ~3,262,457 accounts totalling ~1,333,000,000 GNOT (ATOM/ATONE airdrops, investor buckets, treasuries, public sale; see that repo's README for the bucket table) — downloaded by pinned-commit URL and sha256-verified at build time. Plus exact-burn funding for the genesis-tx fee payers: they land at zero post-genesis, or at exactly their allocation when an address is both (the collision gnoland1 left unresolved is handled by summing burn on top of allocation). **No faucets.**
- **Governance**: the seven GovDAO T1 members from the gnolang/multisigs `[govdao]` section (aeddi's operational key overriding accounts.csv; three more key choices pending — see the bootstrap file), seeded by the bootstrap MsgRun, which also locks `dao.UpdateImpl`'s `AllowedDAOs` to `r/gov/dao/v3/impl`.
- **Validators**: 4 founding validators planned — Gnocore, OnBloc, Samourai-Coop, Berty — one each, power 60 (one dark = one quarter lost, above the halt boundary). All keys/operators are `TODO(mainnet)` placeholders pending each org's ceremony.
- **Namespace enforcement**: `r/sys/names.Enable` runs as a genesis MsgCall, so name-based deploy authorization is on from block 1 (admin address confirmation pending — `TODO(mainnet)`).
- **Vested accounts**: `TODO(mainnet)` — the §132 investors-vesting bucket (150M GNOT, 24 months) is the known candidate; the mechanism (balance-sheet vesting syntax, continuous or cliff) is inherited from the pearl builder and already exercised there.
- **Inert packages**: to be ACTIVE at genesis (`code_submission_policy=inert`) — approvers, run-submitters, and the submission charge are `TODO(mainnet)`; genesis replay is exempt so the genesis deploys still execute.
- **Transfers**: `TODO(mainnet)` — restricted vs open is undecided (gnoland1 launched locked; the fresh testnets launched open).

Not set at genesis (defaults apply; adjustable post-genesis via GovDAO proposals, see `misc/govdao-scripts/`): CLA (`TODO(mainnet)`: decide), minimum fee (`TODO(mainnet)`: mainnet likely wants a real minimum gas price at genesis).

## Quick start

The script builds the binaries from the worktree, downloads and verifies the allocation sheet (cached after the first run), assembles the genesis txs, measures fee-payer balances on a temp node, and verifies sha256 of the locked build artifacts.

```bash
./gen-genesis.sh                # full build
./gen-genesis.sh --no-install   # reuse previously built binaries
./gen-genesis.sh --debug        # echo the main pipeline commands
```

Output: `genesis.json` at the root of this directory (large: ~3.26M balance entries).

## Directory layout

```
mainnet.gno.land/
├── gen-genesis.sh         # Single self-contained pipeline
├── govdao-exec.sh         # Helper for post-genesis governance ops
├── genesis.json           # Final artifact (produced by the script)
├── allocation_balances.txt.gz  # Cached independence-day sheet (gitignored)
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
8. Download + sha256-verify the allocation sheet, enforce the overlap rules (fee payer ∩ allocation → summed; vested ∩ anything → rejected), then measure fee-payer balances via a two-pass temp-node run (measure → verify zero), gated on committed state.
9. Add the validators + balances (fee payers + vested, then the allocation sheet last), run `gnogenesis verify`, move `genesis.json` into place.

The locked artifacts (package list, valoper seed, tx stream, `genesis.json`) are checked against the `CHECKSUMS_DATA` manifest embedded in the script: after the first clean build with final values, paste the printed "not listed" lines into the heredoc to lock the build; any future run producing different bytes fails loudly.

## Transactions folder

Every entry under `transactions/` is a directory containing a `meta.json` (carries the `reason` audit field, a `kind` discriminator, and signing parameters) and optionally a body file. The `txn_dir_to_jsonl` helper in `gen-genesis.sh` converts such a directory into one tx jsonl line, signing via `gnokey` with the deterministic deployer key. `MsgCall` entries support `caller_override`: the caller field is jq-patched post-sign, which the chain trusts at genesis under `--skip-genesis-sig-verification` — used by `names-enable` to satisfy the admin gate without holding the admin key. `TODO(mainnet)`: decide whether mainnet keeps the caller-patch/dummy-signature pattern (every node must run `--skip-genesis-sig-verification` forever) or runs a proper signing ceremony for the admin txs.
