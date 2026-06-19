# test13 hardfork genesis

Builds the **test13** hardfork genesis from gnoland1's chain state. Versus gnoland1, test13:

- **Adds 5 packages** to the base set: `p/onbloc/{uint256,int256,json}`, `r/sys/validators/v3`, `r/demo/defi/grc20reg`.
- **Rotates the sole GovDAO T1 member** from manfred (gnoland1 inherit) to aeddi.
- **Replaces the valset** with test13's and noops every historical gnoland1 valset change 
- **Adds 10 pre-funded faucets** at 1e18 ugnot (≈1T GNOT) each and unrestricts them at genesis.
- **Rebases on current master**: gnoland1's historical txs are replayed as-is wherever they still execute; the rest are patched (39 entries in 8 groups) to absorb the breaking changes master has shipped since gnoland1.

## Quick start

The script is fully self-contained: builds the binaries from the worktree, downloads the independence-day airdrop snapshot, fetches the historical tx stream, applies patches + migrations, assembles the genesis, and verifies sha256 of every produced artifact.

```bash
./gen-genesis.sh                                  # fetch gnoland1 txs from 5 RPC endpoints in parallel — 1-2 hours
./gen-genesis.sh --source-txs-jsonl-file <path>   # use a gnoland1 amino-jsonl export — ~5 min
./gen-genesis.sh --source-txs-data-dir <path>     # use a halted gnoland1 data dir — ~5 min
```

Output: `genesis.json` (~192 MB) at the root of this directory once phase 2 succeeds.

## Directory layout

```
test13.gno.land/
├── gen-genesis.sh         # Single self-contained pipeline (both phases inline)
├── govdao-exec.sh         # Helper for post-genesis governance ops
├── genesis.json           # Final hardfork artifact (produced by phase 2)
│
├── packages/              # Project-specific realm source for batched addpkg
│   └── rotate/            # Single-use realm called via post-replay MsgCall
│
├── transactions/          # Per-tx directories, one folder per Source category
│   ├── base/
│   │   └── bootstrap/     # Bootstrap MsgRun (T1 setup, faucets, lock-down)
│   ├── migration/
│   │   ├── rotate-call/   # Post-replay MsgCall to rotate.Rotate
│   │   └── names-enable/  # Post-replay MsgCall to names.Enable
│   └── patched/           # Historical-tx overrides (39 patches in 8 groups)
│       ├── unrestrict/    │ restrict/
│       ├── set-cla/       │ set-minfee/
│       ├── gnomaze/       │ boards2-permissions/
│       ├── boards2-cascade/                # 9 patches (caller-swap + dead-letter)
│       └── validator-noops/                # 22 noop patches
│
└── work/                  # Gitignored — generated/downloaded artifacts
    ├── phase-1/           # deployer keyring, base genesis, airdrop, ...
    └── phase-2/           # cached txs.jsonl, fork-test log, intermediate genesis
```

## Pipeline architecture

`gen-genesis.sh` is a single self-contained bash script with two phases.

### Phase 1 — build the BASE genesis

1. Resolve script paths and tooling.
2. Verify required tools (preflight with `brew` + `apt` install hints).
3. Build binaries from source (`gno`, `gnokey`, `gnoland`, `gnogenesis`).
4. Generate filtered examples `addpkg` txs (88 packages + `packages/rotate`).
5. Generate the bootstrap MsgRun tx from `transactions/base/bootstrap/`.
6. Calculate deployer balances via a two-pass temp-node run (measure → verify zero).
7. Add the initial validator set.
8. Generate the `valoper-seed` migration jsonl + merge deployer/airdrop balances.
9. Append faucet balances (10 × 1e18 ugnot) + run `gnogenesis verify`.

Output: `work/phase-1/base-genesis.json`.

### Phase 2 — apply historical txs + patches + migrations

1. Resolve script paths and tooling.
2. Verify required tools + phase-1 outputs are present.
3. Build `gnokey` + `gnogenesis` (skipped under `--no-install`) and validate the chosen txs source.
4. Build the migration jsonl (`rotate-call` + `names-enable` MsgCalls signed via the unified loader) and assemble the final genesis via `gnogenesis fork generate` (base + historical + patches + migrations).
5. Audit via `gnogenesis fork test --skip-failing-genesis-txs` (requires 0 failures; any failure aborts), move the genesis to the root, and print the provenance report via `gnogenesis fork inspect`.

Output: `genesis.json` at the root of this directory.

## Transactions folder

Every entry under `transactions/` is a directory containing a `meta.json` (always — carries the `reason` field surfaced in `gnogenesis fork inspect` plus a `kind` discriminator) and optionally one or more body files (a sibling `body.gno`, or a `pkg/` subdir for multi-file packages). The `txn_dir_to_jsonl` helper in `gen-genesis.sh` converts any such directory into one `AnnotatedTx` jsonl line. The three subdirectories track three of the four `GnoTxMetadata.Source` categories the chain tags every genesis tx with (`historical` is the fourth, fed from gnoland1 RPC / cache at run time).

### `base/`

Genesis-time txs prepended to the assembled genesis. One entry today: `bootstrap/` — the test13 bootstrap MsgRun. Seeds the deployer as sole T1, unrestricts the 10 faucets + the gnoland1 airdrop faucet + the GovDAO multisig, locks the bank, swaps in manfred as the permanent T1, restricts `dao.UpdateImpl` `AllowedDAOs` to `r/gov/dao/v3/impl` + `r/test13/rotate`, and self-ejects the temporary deployer.

### `patched/`

Per-tx overrides applied to gnoland1's historical tx stream. Each patch is matched against the source by `(block_height, signer_info[0].address, signer_info[0].sequence)`; `gnogenesis fork generate` fails fast on duplicate or unmatched keys. All 8 groups, summarised:

| Group                 | #   | Cause                                                                                                                                                                                                             | Technique                                                                                |
| --------------------- | --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| `unrestrict`          | 3   | Post-#5669 API drift on `params.ProposeAddUnrestrictedAcctsRequest` — needs `cur realm` param + `cross(cur)` + `dao.NewVoteRequest` factory                                                                       | rewrite body                                                                             |
| `restrict`            | 1   | Same drift on the `Remove` variant                                                                                                                                                                                | rewrite body                                                                             |
| `set-cla`             | 1   | Same drift on `cla.ProposeNewCLA`                                                                                                                                                                                 | rewrite body                                                                             |
| `set-minfee`          | 1   | Same drift on `proposal.ProposeNewMinFeeProposalRequest`                                                                                                                                                          | rewrite body                                                                             |
| `gnomaze`             | 1   | User realm `g19p3yzr…/gnomaze`: `runtime.CurrentRealm` moved to `runtime/unsafe`; drop `cur` from helpers that don't need it                                                                                      | rewrite body (multi-file `pkg/`)                                                         |
| `boards2-permissions` | 1   | User realm `g1hy6zry…/boards2/permissions/v1`: `runtime/unsafe` split + `banker.NewBanker(BankerTypeReadonly)` → `banker.NewReadonlyBanker()`                                                                     | rewrite body (multi-file `pkg/`)                                                         |
| `boards2-cascade`     | 9   | PR #5280 removed `g16jpf0…` from `boards2/v1` `initRealmPermissions`; historical ops by that address can't pass the permission check on test13                                                                   | 7 caller-swap (rewrite `MsgCall.Caller` to the GovDAO T1 multisig) + 2 dead-letter noops |
| `validator-noops`     | 22  | 17 historical `add_validator`/`rm_validator` MsgRuns target the vestigial `r/sys/validators/v2` realm (unseeded on test13 by design) + 5 historical `valopers.Register` MsgCalls fail the post-#5285 squat guard | noop MsgRun                                                                              |

The single biggest cascade comes from `unrestrict` h1950: without that patch the gnoland1 airdrop faucet stays restricted, and 2,553 subsequent faucet `MsgSend`s fail with `RestrictedTransferError`. One patch resolves the entire cascade.

### `migration/`

Post-replay genesis-mode MsgCalls — they run after the historical replay and are tagged `Source="migration"`. Two entries today:

- `rotate-call/` — MsgCall to `gno.land/r/test13/rotate.Rotate()`. Swaps the sole T1 from manfred to the test13 multisig via direct memberstore writes, then self-ejects from `AllowedDAOs`.
- `names-enable/` — MsgCall to `gno.land/r/sys/names.Enable()`. Turns on the v3 names module (gnoland1 left it disabled).

Both use `caller_override` to appear as the GovDAO T1 multisig at execution time (the only account funded at migration-replay time).

## Packages

`packages/` holds project-specific realm source that's bundled into the batched `addpkg` alongside the filtered examples in phase-1 step 4 (the example tooling topo-sorts so deps land first). One entry today:

- `rotate/` — single-use realm deployed at base genesis. Its `Rotate()` function is called once via `transactions/migration/rotate-call/`; the realm self-ejects from `AllowedDAOs` after the swap. Direct memberstore writes — proposal-flow rotation isn't workable across the genesis-mode replay boundary.
