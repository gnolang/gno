# ADR: test-13 hardfork launch

> **Note:** This ADR documents the original `chain/gnoland1`-based launch
> stack ([PR #5589](https://github.com/gnolang/gno/pull/5589)). After
> [PR #5485](https://github.com/gnolang/gno/pull/5485) (v3 valset via
> params keeper) landed on master, the master-based companion stack
> ([PR #5597](https://github.com/gnolang/gno/pull/5597)) supersedes
> this one for any new launch — see
> [`pr5597_test13_hardfork_launch_master_based.md`](./pr5597_test13_hardfork_launch_master_based.md)
> for the deltas (notably: migration sequence shrinks from 01 → 08 to
> 01 → 07 because `vm:p:valset_realm_path` is no longer a configurable
> param).

## Context

[PR #5511](https://github.com/gnolang/gno/pull/5511) shipped the
replay-engine primitives (`PastChainIDs`, `GnoTxMetadata`,
`InitialHeight`, the hardfork-aware ante handler). [PR
#5486](https://github.com/gnolang/gno/pull/5486) shipped `hf-glue`, the
testbed that drives those primitives end-to-end. [PR
#5485](https://github.com/gnolang/gno/pull/5485) shipped
`r/sys/validators/v3`, the params-keeper-based valset flow that
replaces v2's event-collector path.

This ADR is the launch playbook for **test-13**, the first concrete
chain to exercise all three: a hardfork of `gnoland1` that halts it at
a chosen height, rebuilds the genesis with its history preserved, and
resumes as `test-13` — with the v3 valset flow live from the first
post-fork block.

Unlike the referenced ADRs, this one is not proposing a gno primitive.
It records the set of operational decisions we took _as a consumer_ of
those primitives so that:

- a reviewer of the commit series understands why each rc exists;
- a future hardfork (test-14, gnoland-2.0) can reuse the same pattern;
- a later operator reading state reports can tell which post-fork
  divergence from source was intentional vs what would be a bug.

## Decision

### 1. rc-stacking branch model

Work ships as a sequence of focused rc branches rather than a single
growing one:

```
chain/gnoland1 → test13-base → test13-rc1 → rc2 → rc3 → rc4 → rc5 → rc6
```

Each rc is a strict superset of the previous and introduces one
coherent concern (see `misc/deployments/test13.gno.land/PR-DESCRIPTION.md`
for the per-rc delta list). Older rcs stay reachable on the `aeddi`
remote — a reviewer bisecting a regression between rc4 and rc5 can
check out rc4 and reproduce against it, unchanged.

Cost: the rc number becomes a load-bearing identifier in commit
messages and in external communication. That's cheap; in exchange we
get a cleaner review surface and a real history of what the launch
prep actually looked like.

### 2. Migration sequence 01 → 08

Post-history migration txs, applied to the replayed genesis in order,
each signed by the current sole-T1 member under
`--skip-genesis-sig-verification`:

| Step   | Purpose                                                                                                | Why it exists                                                                                                                                                                                                                                                                             |
| ------ | ------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **01** | `r/sys/validators/v2.NewPropRequest` — remove pre-fork validators, add the post-fork validator(s)      | Keeps v2's in-realm valset view in sync with the new `GenesisDoc.Validators`. v2 is vestigial post-#5485 (EndBlocker no longer reads it), but tooling that still queries v2 should not return a ghost of the pre-fork world.                                                              |
| **02** | `r/gov/dao/v3/impl.NewAddMemberRequest(NEW_T1_ADDR, T1, …)` signed by manfred                          | Manfred is the single T1 at halt height. Before removing him we must add the post-fork T1 so the DAO never has zero T1 members (would brick governance).                                                                                                                                  |
| **03** | `r/gov/dao/v3/impl.NewWithdrawMemberRequest(manfred, …)` — manfred proposes his own removal, votes YES | With two T1 members (manfred + NEW_T1_ADDR), one YES vote is 50% — not enough to execute. The proposal sits in `ACTIVE` state waiting for the second vote.                                                                                                                                |
| **04** | NEW_T1_ADDR votes YES on the open withdraw proposal + executes                                         | Second YES brings approval to 100%. Proposal executes, manfred is removed, NEW_T1_ADDR becomes the sole T1.                                                                                                                                                                               |
| **05** | `r/sys/params.NewSysParamStringPropRequest("vm","p","sysnames_pkgpath","")` — govDAO proposal          | v3 needs to be addpkg'd under `r/sys/*`. `r/sys/names` is enabled at halt height and rejects any addpkg whose namespace doesn't equal the caller's address (no real address satisfies `sys`). Clearing the VM param lets `checkNamespacePermission` short-circuit at `sysNamesPkg == ""`. |
| **06** | `MsgAddPackage` for `r/sys/validators/v3`, creator=manfred                                             | Deploys v3 itself. Runs while the namespace check is disabled.                                                                                                                                                                                                                            |
| **07** | Restore `vm:p:sysnames_pkgpath` to `gno.land/r/sys/names`                                              | Re-enables the namespace check so subsequent `r/sys/*` addpkgs on the chain go through the same authorisation path they did on mainnet.                                                                                                                                                   |
| **08** | Set `vm:p:valset_realm_path` to `gno.land/r/sys/validators/v3`                                         | Without this, EndBlocker reads the param as the empty string (mainnet state predates the field), picks an invalid realm path, and silently drops every v3 valset update. Can't be folded into v3's `init()` because the VM param lives outside the realm's scope.                         |

The ordering is constrained: step 01 requires v2 to still exist, steps
02–04 require a two-member DAO window, step 06 requires the namespace
check to be off and is bracketed by 05 / 07 to keep that window
minimal. Step 08 must run after step 06 (the realm has to exist before
we point the param at it). Beyond those constraints the current order
is the simplest linearisation.

### 3. Replay posture

- `--skip-genesis-sig-verification` is **required** on every validator
  at launch. Without it the first gnoland1 tx whose signature can't be
  reconstructed (e.g. because its `SignerInfo` wasn't exported) panics
  the chain at InitChain.
- `--skip-failing-genesis-txs` is **required** on every validator at
  launch. With `PanicOnFailingTxResultHandler` (the default), the
  ~2580 expected `InsufficientFundsError` failures would take the
  chain down; with `NoopGenesisTxResultHandler` they are counted and
  skipped.
- `app_state.gas_replay_mode` stays empty/`strict`. `source` mode
  (bypass the new VM gas meter for historical txs) would not recover
  the InsufficientFunds failures — those fire at
  `auth.DeductFees` _before_ the gas meter is consulted. See
  `scripts/compare-gas-modes.sh` for the A/B result; delta is zero.

### 4. Intentional divergences from source

Every post-fork state that differs from the gnoland1 state at halt
height is either recorded as intentional here or is a bug.

- **Balance top-ups** via `hf_topup_balance` in `scripts/migrate.sh`.
  Currently one entry: `g1r929wt2qplfawe4lvqv9zuwfdcz4vxdun7qh8l`
  +1 Gugnot, for storage-deposit headroom on its 7 `r/sys/*` genesis
  addpkgs. Every new top-up is logged to
  `out/TOPUP-REPORT.md`; the auditable rule is "any balance in that
  report is expected to diverge, any balance not in that report and
  not zero-on-source must match source".
- **`r/sys/validators/v2` is frozen cosmetic state.** EndBlocker
  ignores it (PR #5485). Migration 01 aims to bring v2 in sync, but it
  is known to partially apply when the removal batch hits an address
  already removed by governance history (v2's batched prop panics the
  whole batch on the first miss). Consensus is unaffected; valset
  authority is tm2 + v3.
- **Proposal IDs post-fork are contiguous with mainnet's series.**
  Mainnet at halt height had proposals up to id N; test-13 continues
  at id N+1. Our migrations create several of those ids during
  replay.
- **govDAO membership is rotated to a single post-fork T1**
  (`NEW_T1_ADDR`). Known fragility — a single-EOA T1 is the minimum
  viable and should be expanded before real production. Not a launch
  blocker for test-13 (testnet); would be for a gnoland-2.0.

### 5. Launch verification

Before calling a genesis shippable, run on the final build:

1. `make verify-reproducibility` — two independent builds produce
   identical SHA256 on the same host.
2. `make verify-txs-jsonl` — the cached historical tx set matches
   source-chain `total_txs` at halt height, plus random spot-checks.
3. `make smoketest` — `gnogenesis fork test` completes; failure count
   matches the documented baseline (2605 for the gnoland1 → test-13
   case).
4. Boot the genesis on a multi-node cluster, then:
    - `make assert-migrations` — every migration's intended effect
      landed.
    - `make state-diff` — rendered state diff against mainnet at halt
      height has no unexpected divergences.
    - `make audit-balances` — diverged accounts match the ones
      documented in `out/TOPUP-REPORT.md` (or the policy has been
      updated to accept new divergences).
    - `make audit-realm-imports` — every import in the deployed
      realms resolves against the running stdlib.

All steps are non-interactive and exit non-zero on failure, so a
simple Makefile chain (or CI job) can gate the launch build.

## Consequences

- **Silent failures are bounded.** Every divergence surface has a
  check; an operator who runs all of §5 on the build they're about to
  ship sees every documented drift line up with the check output, and
  nothing else.
- **Recovery from mid-replay kill is bootable.** The `defaultStore`
  `AddMemPackage` path is not crash-consistent (index in pebble + body
  in IAVL, two separate writes). The rc5 nil-skip in
  `PreprocessAllFilesAndSaveBlockNodes` makes a partially-persisted
  store still boot; any missed realms surface as VM errors when
  first called, not as boot-time crash loops. The proper atomic-write
  fix belongs upstream.
- **`halt_height` is one-shot-per-param.** After the halt fires the
  in-memory `BaseApp.haltHeight` is set but is not re-applied on
  process restart; operators restarting the same chain post-halt will
  resume past the halt. Intentional: after a coordinated halt the next
  action is to rebuild genesis on the new chain id, not to resume the
  old chain.
- **`/genesis` RPC serves the full ~200 MB genesis in one response.**
  Large-response handling is a known tm2 weakness; clients that
  disconnect mid-write trigger a recoverable panic on the server.
  Validators should not expose `/genesis` publicly; distribute the
  file out-of-band (GitHub release, IPFS) and point clients at that.

## Alternatives considered

- **`source` gas-replay mode** would skip the new VM's gas meter for
  historical txs. On the test-13 genesis it eliminates zero failures
  (measured with `scripts/compare-gas-modes.sh`) because the failures
  fire in the ante handler. Rejected for this launch; would add audit
  complexity with no payoff.
- **In-place `AddValidator` power update** would simplify
  `change-power.sh`. PoA's `AddValidator` panics on duplicate address
  ("validator must not be in the set already"). Rejected (upstream
  semantic change). `change-power.sh` uses an atomic remove+add in one
  proposal so the EndBlocker diff covers the transition without an
  intermediate absence.
- **Single un-stacked rc branch** would remove the rc-number
  bookkeeping. Rejected — losing the ability to bisect against older
  rcs is a real cost when debugging a late-breaking regression.
- **Fold v3 addpkg into the pre-history genesis-mode txs** (instead of
  migration 06) would avoid the sysnames disable/restore dance.
  Requires modifying the source genesis set before replay starts, and
  that set is deterministically derived from the source chain — editing
  it would desynchronise our SHA from independent rebuilds. Rejected;
  the 05/07 wrap is cheaper.
