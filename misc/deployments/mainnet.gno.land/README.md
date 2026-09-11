# mainnet genesis (WIP)

Builds the **gno.land mainnet** genesis. Mainnet is a **fresh chain** — not a hardfork of betanet (gnoland1) — whose balances come from the audited [gnolang/independence-day](https://github.com/gnolang/independence-day) allocation.

> **Status: WORK IN PROGRESS — not launchable.** Genesis time, the vesting list and the inert-package params are all open — grep `TODO(mainnet)` in this folder for the authoritative list. `CHECKSUMS_DATA` stays unlocked until every value is final, and the independence-day pin moves with that repo until the allocation freeze.

## What mainnet contains

- **Balances**: the independence-day allocation — 3,262,454 accounts totalling 1,332,999,998.328067 GNOT at the current pin (ATOM/ATONE airdrops, investor buckets, treasuries, public sale; see that repo's README for the bucket table) — downloaded by pinned-commit URL and sha256-verified at build time, then reconciled against the shipped genesis at step 9.5. Plus exact-burn funding for the genesis-tx fee payers: they land at zero post-genesis, or at exactly their allocation when an address is both (the collision gnoland1 left unresolved is handled by summing burn on top of allocation). **No faucets.**
- **Governance**: the seven GovDAO T1 members from the gnolang/multisigs `[govdao]` section (every key confirmed with its owner; aeddi's is his operational key rather than the accounts.csv one), seeded by the bootstrap MsgRun, which also locks `dao.UpdateImpl`'s `AllowedDAOs` to `r/gov/dao/v3/impl`. The build reads those addresses back out of the bootstrap source and refuses to produce a genesis unless each one holds a balance: with no faucet and no transferable supply, a member seeded without funds could never pay for a proposal. Any nonzero balance qualifies — fees are collected via the bank's unrestricted path, so neither a vesting schedule nor the §126 lock stops a member from paying gas.
- **Validators**: 4 founding validators — Gnocore, OnBloc, Samourai Crew, Berty — one each, power 60 (one dark = one quarter lost, safely below the one-third halt boundary). All four are real: each org's ceremony consensus pair plus its operator address, every consensus address cross-checked by deriving it from the pubkey. Per-org signer requirements remain `TODO(mainnet)`.
- **Namespace enforcement**: `r/sys/names.Enable` runs as a genesis MsgCall, so name-based deploy authorization is on from block 1. The admin is the gnolang/multisigs `[govdao]` 4-of-7 multisig, and the build asserts the configured value against the admin actually compiled into `r/sys/names/verifier.gno`, so a master-side change fails the build instead of panicking a temp node.
- **Vested accounts**: `TODO(mainnet)` — the §132 investors-vesting bucket (150M GNOT, 24 months) is the known candidate; the mechanism (balance-sheet vesting syntax, continuous or cliff) is inherited from the pearl builder and already exercised there.
- **Inert packages**: to be ACTIVE at genesis (`code_submission_policy=inert`) — approvers, run-submitters, and the submission charge are `TODO(mainnet)`; genesis replay is exempt so the genesis deploys still execute.
- **Transfers**: **locked at genesis**, per Constitution §126 (*"$GNOT will not be transferrable initially except for whitelisted addresses"*). `bank.params.restricted_denoms=["ugnot"]` plus the 71-address exemption list fetched from `gnolang/independence-day` (`mkgenesis/unrestricted.txt`, same sha256 treatment as the balance sheet; the two pins are temporarily split across commits — see the `TODO(mainnet)` at the pin definitions). Applied at step 9.3. Both knobs are required: with `restricted_denoms` empty the exemption list is inert.

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
├── allocation_balances.txt.gz   # Cached independence-day sheet (gitignored)
├── allocation_unrestricted.txt  # Cached §126 exemption list (gitignored)
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
2. Verify required tools, fetch and sha256-verify the allocation sheet and the §126 exemption list, and assert the bootstrap's T1 members are funded — everything that can fail in seconds, before the ten-minute build.
3. Build binaries from source (`gno`, `gnokey`, `gnoland`, `gnogenesis`).
4. Resolve `FILTERED_PACKAGES` deps, stage them, and `addpkg` them to the genesis.
5. Add the bootstrap MsgRun from `transactions/base/bootstrap/`.
6. Add the `names.Enable` MsgCall from `transactions/migration/names-enable/`.
7. Build the valoper CSV from `INITIAL_VALSET` + `INITIAL_VALSET_OPERATORS` and add the `valopers.Register` txs (via `gnogenesis fork valoper-seed`).
8. Enforce the overlap rules (fee payer ∩ allocation → summed; vested ∩ anything → rejected), then measure fee-payer balances via a two-pass temp-node run (measure → verify zero), gated on committed state.
9. Add the validators + balances (fee payers + vested, then the allocation sheet last), apply the §126 transfer lock, reconcile the shipped account count and supply against the pinned sheet and the measured burn, run `gnogenesis verify`, move `genesis.json` into place.

The locked artifacts (package list, valoper seed, tx stream, `genesis.json`) are checked against the `CHECKSUMS_DATA` manifest embedded in the script: after the first clean build with final values, paste the printed "not listed" lines into the heredoc to lock the build; any future run producing different bytes fails loudly.

## Transactions folder

Every entry under `transactions/` is a directory containing a `meta.json` (carries the `reason` audit field, a `kind` discriminator, and signing parameters) and optionally a body file. The `txn_dir_to_jsonl` helper in `gen-genesis.sh` converts such a directory into one tx jsonl line, signing via `gnokey` with the deterministic deployer key. `MsgCall` entries support `caller_override`: the caller field is jq-patched post-sign, which the chain trusts at genesis under `--skip-genesis-sig-verification` — used by `names-enable` to satisfy the admin gate without holding the admin key. `TODO(mainnet)`: decide whether mainnet keeps the caller-patch/dummy-signature pattern (every node must run `--skip-genesis-sig-verification` forever) or runs a proper signing ceremony for the admin txs.
