# pearl genesis

Builds the **pearl** genesis. pearl is a **fresh chain** — not a hardfork, no historical replay.

> **Status: one launch value outstanding.** The ceremony keys, operators, genesis time, faucet accounts, peers, and the ten vested test accounts are final. Outstanding: the `PKG_APPROVERS` oracle address. pearl launches under the `inert` code-submission policy, and `gen-genesis.sh` refuses to build a genesis without an approver (step 1, before anything is compiled) rather than shipping a chain that parks deploys nobody can activate. `CHECKSUMS_DATA` is locked on the first clean build after that address is pasted in.

## What pearl contains

- **Packages**: the curated `examples/` set (resolved with transitive deps) — `r/sys/...`, `r/gov/...`, `r/gnoland/{blog,wugnot,coins,boards2}/...`, `r/gnops/valopers/...`, `p/onbloc/{uint256,int256,json}`, `r/sys/validators/v3`, `r/demo/defi/grc20reg` — deployed by the deterministic `GenesisDeployer` key (packages carrying a `gnomod.toml` `[addpkg] creator` are deployed under that creator instead). The resolution includes test-only dependencies (`-test-dep`): packages ship on-chain with their `_test.gno` files, and the test deps are deployed alongside so those files keep their imports resolvable on-chain (`MsgAddPackage` itself type-checks production files only).
- **Governance**: aeddi (`g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m`) is the sole GovDAO T1 member, seeded by the bootstrap MsgRun, which also locks `dao.UpdateImpl`'s `AllowedDAOs` to `r/gov/dao/v3/impl`. Additional members join via GovDAO proposals.
- **Validators**: 3 founding validators (`gno-core-validator-1/-2/-3`, power 60 each) in `GenesisDoc.Validators` — at 3 × 60, one validator dark is exactly the one-third halt boundary, accepted for launch until partners join. InitChainer seeds `r/sys/params`' `valset:current` from it, so `r/sys/validators/v3` proposals manage the set post-genesis. Each founder has a valoper profile (registered at genesis via `gnogenesis fork valoper-seed`) keyed on an operator address, giving them the operator-keyed management plane (signing-key rotation, profile edits, opt-out).
- **Namespace enforcement**: `r/sys/names.Enable` runs as a genesis MsgCall, so name-based deploy authorization is on from block 1. The call's caller field is patched to the admin address hardcoded in `r/sys/names/verifier.gno` (trusted under `--skip-genesis-sig-verification`; the private key is not needed).
- **Code submission**: pearl launches under the **`inert`** policy (`code_submission_policy`, from #6088). A post-genesis `MsgAddPackage` from anyone is *parked* — stored without a type-check, without `init()`, not importable — and becomes callable only once an address in `pkg_approvers` sends `MsgEnablePackage`; see [Approval oracle](#approval-oracle). Genesis content is exempt by construction, and the launch rests on it: the keeper's parking branch requires `BlockHeight() > 0` and `InitChain` runs at height 0, so all 85 packages deploy live. Pinned by `TestInertPolicyAtGenesisDeploysLive` (`gno.land/pkg/gnoland`) and confirmed on the built artifact — `vm/qinertpaths` empty, `vm/qpaths` listing all 85. A parked submission also returns `ok`, so the strict-replay report alone does not prove liveness; the queries do. Submitting costs a flat `inert_submission_charge` of 1 GNOT paid to the approver, so submissions fund the approvals they cause — not levied on genesis content, which is why it cannot perturb the fee-payer measurement.
- **MsgRun allowlist**: `run_submitters` is **armed** — aeddi (`g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m`) and the relayer (`g1z437dpuh5s4p64vtq09dulg6jzxpr2hd4q8r5x`). This is what makes the review gate mean something: `inert` defers only `MsgAddPackage`'s type-check, while `MsgRun` type-checks and *executes* arbitrary source immediately under every policy value, in an ephemeral `gno.land/e/<caller>/run` realm. Two consequences to know before launch. Ordinary users cannot `gnokey maketx run` on pearl. And because GovDAO proposal creation is MsgRun-only (a `ProposalRequest` carries a func value, which `MsgCall` cannot marshal), **every future GovDAO member has to be added to this list too**, or they can join the DAO and still not be able to propose — nine of the twelve `misc/govdao-scripts/` commands are `maketx run`. Genesis txs are exempt (every InitChain tx bypasses the code-policy ante), so the bootstrap MsgRun is unaffected. Post-genesis the key is *not* settable through `r/sys/params`' generic factories: it is reserved for `ProposeSetRunSubmitters`, which can also delegate add-only management to one other realm.
- **Balances**: 3 faucet accounts at 1e18 ugnot (≈1T GNOT) each — the web faucet's dispensing account, the faucet-agent's dispensing account, and an operator reserve — plus the `PKG_APPROVERS` oracle at 1e18 (it pays the gas for every `MsgEnablePackage`, so an unfunded approver stalls the policy on the first submission), plus exact-burn funding for the genesis-tx fee payers — the deployer, the names admin, and every `gnomod.toml` `[addpkg] creator` address in the package set — which land at zero once the genesis txs execute. No airdrop, no inherited balances.
- **Vested accounts**: the `VESTED_ACCOUNTS` entries, created at genesis as vesting accounts via the balance-sheet vesting syntax (`addr=coins;vesting=coins,start,end[;type=delayed]`). Continuous schedules unlock linearly between start and end; delayed schedules are a cliff. The unvested remainder is spendable immediately.
- **Transfers**: unrestricted — no bank lock, no unrestricted-accounts list.

Not set at genesis (defaults apply; adjustable post-genesis via GovDAO proposals, see `misc/govdao-scripts/`): CLA, minimum fee, and `default_deposit` — `100000000ugnot` (100 GNOT, ≈1 MB of state at the 100ugnot/byte storage price) since #6088 lowered it from 600M; it is a per-package ceiling, not a charge.

To run a full node and put yourself forward as a validator on pearl, see [`VALIDATOR.md`](./VALIDATOR.md).

## Approval oracle

Under `inert`, nothing submitted after genesis becomes callable until an approver enables it. That approver is [`contribs/gpao`](../../../contribs/gpao) — it watches blocks, re-runs the type-check and preprocess off-chain under a wall-clock budget, and broadcasts `MsgEnablePackage` for what passes.

**pearl does not deploy new code without it running.** If the oracle is down, nothing breaks and nothing is lost — submissions accumulate parked, and activate whenever it comes back. But no new package goes live in the meantime, so treat it as launch infrastructure alongside the RPC and faucet, not as an optional extra.

```bash
gpao \
  --remote https://rpc.pearl.testnets.gno.land \
  --chain-id pearl-1 \
  --key <approver-key> \
  --status-listen 127.0.0.1:8546
```

Operational notes:

- **The key** is the `PKG_APPROVERS` entry, funded at genesis. It can activate any parked package and disable any active one, which makes it the most consequential key on the chain after governance — and it sits unattended on a daemon, unlocked from `$GPAO_PASSWORD`. Keep it to this one job.
- **Why the charge goes to the approver.** `MsgEnablePackage` runs the submitted package's `init()` on the *oracle's* transaction and gas meter, so the submitter chooses the work and the oracle pays for it. Unpriced, that is a cheap way to exhaust the oracle's `--max-spend` and stop approvals for everyone; `inert_submission_charge` prices it and routes it back to the payer.
- **Serve `--status-listen`.** A refusal is the oracle's knowledge alone: the chain can say a package is parked, never why this one was not approved. Without the status endpoint a submitter pays the charge and hears nothing.
- **The oracle is untrusted for correctness.** The chain re-runs `TypeCheckMemPackage` at enable and rejects ill-typed code regardless, and the namespace check already applied at submit. gpao only decides *which* parked packages are proposed for activation, and when.

## Quick start

The script is fully self-contained: builds the binaries from the worktree, assembles the genesis txs, measures fee-payer balances on a temp node, and verifies sha256 of the locked build artifacts.

```bash
./gen-genesis.sh                # full build — a few minutes
./gen-genesis.sh --no-install   # reuse previously built binaries
./gen-genesis.sh --debug        # echo the main pipeline commands
```

Output: `genesis.json` at the root of this directory.

## Directory layout

```
pearl.gno.land/
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

1. Resolve script paths and tooling, and validate the code-submission launch parameters. This runs before anything is compiled, and refuses the combinations that cannot be repaired on a live chain: `inert` with no approver, an armed `run_submitters` missing the GovDAO T1 seed (cross-checked against the bootstrap tx, not against a comment), a charge above the on-chain cap, and a charge with no real collector.
2. Verify required tools (preflight with `brew` + `apt` install hints).
3. Build binaries from source (`gno`, `gnokey`, `gnoland`, `gnogenesis`).
4. Resolve `FILTERED_PACKAGES` deps, stage them, and `addpkg` them to the genesis. The code-submission vm params are written here, by the same helper that writes them into every other genesis this script generates — step 8 asserts the measurement genesis's vm params equal the shipping genesis's, since a measurement taken under different fee-governing params would be wrong.
5. Add the bootstrap MsgRun from `transactions/base/bootstrap/`.
6. Add the `names.Enable` MsgCall from `transactions/migration/names-enable/`.
7. Build the valoper CSV from `INITIAL_VALSET` + `INITIAL_VALSET_OPERATORS` and add the `valopers.Register` txs (via `gnogenesis fork valoper-seed`).
8. Measure fee-payer balances via a two-pass temp-node run (measure → verify zero). Readiness gates on committed state, and a fee payer reading zero in the measure pass aborts the build.
9. Add the validators + balances (fee payers + faucet accounts + the approver + vested accounts), run `gnogenesis verify`, move `genesis.json` into place.

The locked artifacts (package list, valoper seed, tx stream, `genesis.json`) are checked against the `CHECKSUMS_DATA` manifest embedded in the script: after the first clean build, paste the printed "not listed" lines into the heredoc to lock the build; any future run producing different bytes fails loudly.

## Transactions folder

Every entry under `transactions/` is a directory containing a `meta.json` (carries the `reason` audit field, a `kind` discriminator, and signing parameters) and optionally a body file. The `txn_dir_to_jsonl` helper in `gen-genesis.sh` converts such a directory into one tx jsonl line, signing via `gnokey` with the deterministic deployer key. `MsgCall` entries support `caller_override`: the caller field is jq-patched post-sign, which the chain trusts at genesis under `--skip-genesis-sig-verification` — used by `names-enable` to satisfy the admin gate without holding the admin key.
