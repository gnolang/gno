# ADR: test-13 hardfork launch — master-based rc series

## Context

This ADR is the master-based companion to
[`pr5589_test13_hardfork_launch.md`](./pr5589_test13_hardfork_launch.md).
The two cover the same launch — gnoland1 → test-13 with history
preserved, v3 valset live from the first post-fork block — but they
target different bases:

- **PR [#5589](https://github.com/gnolang/gno/pull/5589)** stacks on
  `chain/gnoland1`. That base is what gnoland1 actually shipped with;
  cherry-picks of merged upstream PRs ride along to bring its core
  fixes forward without picking up later refactors.
- **This stack ([PR #5597](https://github.com/gnolang/gno/pull/5597))**
  stacks on `master`. By the time this is being prepared, every
  cherry-pick in #5589's rc1 has independently landed on master, and
  master also carries the gas-storage refactor ([#5415](https://github.com/gnolang/gno/pull/5415)),
  the gas-model recalibration ([#5291](https://github.com/gnolang/gno/pull/5291)),
  type-persistence dedup ([#5544](https://github.com/gnolang/gno/pull/5544)),
  bptree ([#5475](https://github.com/gnolang/gno/pull/5475)), account
  sessions ([#5307](https://github.com/gnolang/gno/pull/5307)), and
  others. Building on master gives test-13 those improvements at the
  cost of a smaller-but-non-zero set of new launch-readiness fixes —
  see §1 below.

Both stacks produce a bootable test-13 genesis from the same
gnoland1-history input. They differ in the resulting node behaviour
(notably gas accounting at the SDK layer) and in the migration
mechanics required to make the new VM accept the source genesis.

The launch playbook itself — migration sequence 01 → 07, replay
posture (`-skip-genesis-sig-verification` + `-skip-failing-genesis-txs`),
intentional divergences from source, launch verification gates,
alternatives considered — is identical to #5589's modulo step 08
(see §2 below). Everything in [`pr5589_test13_hardfork_launch.md`
§2 through §5 plus Consequences and Alternatives](./pr5589_test13_hardfork_launch.md)
applies here unchanged except for that step. This document only
records the decisions that are **specific to master-basing** the
stack.

## Decision

### 1. Master adds a new launch-readiness gap; rc5-master closes it

The chain/gnoland1 base predates [#5415](https://github.com/gnolang/gno/pull/5415)
(gas storage refactor) and the related gas-model work. Its
`vm.params` schema has six fields, all carried in the gnoland1
release genesis. After fork, those six pass through unchanged.

Master's `vm.params` schema added seven gas-storage fields
(`min_get_read_depth_100`, `min_set_read_depth_100`,
`min_write_depth_100`, three `fixed_*_depth_100`, and
`iter_next_cost_flat`). The post-refactor `Params.Validate()` rejects
`iter_next_cost_flat == 0` — and that is exactly what the gnoland1
release genesis decodes to under the post-refactor struct, because
the field doesn't exist in the source. Without intervention,
`make migrate` produces a genesis that panics at
`InitChainer: standard libraries loaded` on every validator.

Two further mechanical issues compound this:

- The post-refactor gas meter rejects ~2580 historical txs with
  `InsufficientFundsError` from `auth.DeductFees`. The default
  `gas_replay_mode == "" / "strict"` exposes them; `"source"` is
  the documented escape hatch. The chain/gnoland1 base never had
  to pick a mode because no field existed.
- Migration `01_reset_valset` is a single-proposal-multi-removal:
  one missing validator panics the whole proposal. On the
  chain/gnoland1 base this isn't a launch blocker because the
  migration's whole-batch failure is silently absorbed by
  `--skip-failing-genesis-txs`; on master the same silent absorption
  leaves v2 with the original 7 validators (visible to tooling, not
  to consensus) and the operator only finds out via
  `assert-migrations` post-launch.

[`chain/test13-rc5-master`](https://github.com/aeddi/gno/tree/chain/test13-rc5-master)
addresses all three so the rest of the stack matches the
chain/gnoland1 stack's "regenerate genesis → boot cluster → assert
migrations" workflow without manual JSON patching:

- **`fix(gnogenesis): default gas-storage params and gas_replay_mode
  in hardfork genesis`** — when the source `vm.params` has every new
  field at zero (the pre-refactor signature), `buildHardforkGenesis`
  populates them from `vm.DefaultParams()` and sets `gas_replay_mode
  = "source"`. An operator who has tuned any one of the seven fields
  is left alone (no partial defaulting); a non-empty `gas_replay_mode`
  is preserved. Covered by 4 unit tests in
  `contribs/gnogenesis/internal/fork/generate_test.go`.
- **`fix(deployments/gnoland-1): harden 01_reset_valset against v2
  state drift`** — the migration callback consults live
  `valr.IsValidator(addr)` at proposal-execution time and only emits
  removals for validators actually present. A drift between the
  build-time RPC snapshot and the post-replay v2 state is now a
  no-op on the missing entries instead of a whole-proposal panic.
- **`fix(gnovm/store): body-first AddMemPackage ordering +
  skip-don't-panic in IterMemPackage`** — `AddMemPackage` writes the
  iavlStore body, then the baseStore index slot, then the counter.
  A SIGKILL between any two writes leaves the store in a state where
  `IterMemPackage` simply doesn't see the half-write (instead of
  yielding a counter that points at a missing body). The matching
  `IterMemPackage` change yields `nil` on inconsistency rather than
  panicking — the existing consumer-side skip in
  `PreprocessAllFilesAndSaveBlockNodes` then keeps the node bootable.
  Side effect: replay walltime drops from ~12 minutes to ~36 seconds
  on the test-13 genesis (the order change avoids a pebble write
  amplification pattern that the previous order triggered). Covered
  by 3 unit tests in `gnovm/pkg/gnolang/store_test.go`.

There is also a `feat(gnogenesis): add --skip-failing-genesis-txs
and --skip-genesis-sig-verification flags to fork test` commit so
that `make smoketest` matches what production validators are
configured for — the in-process replay used to fail on the same
~2580 absorbed-by-cluster failures, which made the smoketest output
read as "every build is broken" and obscured real failures.

### 2. Branch numbering

The rc-master series is renumbered relative to #5589's rc series so
that "the layer that will become a no-op once the upstream PR
lands" is always the topmost rc:

| #5589 layer            | This stack's equivalent | Notes                                                                                 |
| ---------------------- | ----------------------- | ------------------------------------------------------------------------------------- |
| `chain/test13-base`    | `chain/test13-base-master` | 5486 hf-glue testbed + master                                                       |
| `chain/test13-rc1`     | _(absorbed into master)_ | All cherry-picks landed upstream; nothing to layer separately                         |
| `chain/test13-rc4`     | `chain/test13-rc1-master`  | Tier-1 audits (audit-balances, state-diff, verify-txs-jsonl, assert-migrations)     |
| `chain/test13-rc5`     | `chain/test13-rc2-master`  | Tier-2 resilience (verify-reproducibility, partial-mpkg-skip, val-ops, assert fix)   |
| `chain/test13-rc6`     | `chain/test13-rc3-master`  | Tier-3 audits (audit-realm-imports, compare-gas-modes, repro-doc fix, ADR import)   |
| `chain/test13-rc2/rc3` | `chain/test13-rc4-master`  | v3 deploy / migration plumbing only — the v3 realm itself now lives in master via #5485 |
| _(new)_                | `chain/test13-rc5-master`  | Master-specific launch fixes (this ADR's §1)                                         |

After #5485 landed on master, `rc4-master` was rebased to drop the
redundant v3 cherry-pick (`a0f175de5`), the eager-eval fix
(`c994c89bd` — master already evaluates the change function eagerly
in `NewProposalRequest`), and the `vm:p:valset_realm_path` migration
step (`2f7155303` — the configurable `ValsetRealmPath` field was
removed from `vm.Params` in the merged version, with the realm path
hardcoded in `r/sys/params/valset.gno` instead). What remains is the
test-13-specific deploy plumbing: the `addpkg r/sys/validators/v3`
migration step (mainnet has no v3 yet), its `sysnames` disable/restore
wrap, and the `add-validator.sh` operator script (adapted to the
renamed `valr.NewProposalRequest(fn, title, description)` signature).
The migration sequence is therefore 01 → 07 on the master-based
stack vs 01 → 08 on #5589's chain/gnoland1 stack.

### 3. Operational consequences vs the chain/gnoland1 stack

- **Smoketest baseline**: the chain/gnoland1 stack documents 2605
  expected failures (#5589 §5.3). The master stack measures 2604
  with rc5-master active (one fewer because the hardened migration 01
  now succeeds where it previously contributed one of the 2605).
- **Replay walltime**: ~36 s on master vs ~13 min on chain/gnoland1
  on identical hardware (M3 Pro, gno-cluster docker). The delta is
  the body-first AddMemPackage ordering — the chain/gnoland1 stack
  doesn't need it because it doesn't have the gas-storage refactor
  that exposed the write-amplification pattern.
- **Boot from `make migrate` output is hands-off**: no `jq` patch
  required. `make migrate && make init && cp out/genesis.json …`
  produces a bootable file.
- **Multi-validator cluster boot reaches consensus**: rc4-master on
  the master stack stuck at the genesis block (813643). With
  rc5-master, the chain advances past genesis (verified to 813665+
  on a 4-node cluster).

### 4. What still belongs upstream

The rc5-master fixes are scoped to make the launch work — they are
not the right long-term shape for any of the three concerns:

- **Cross-substore atomicity in `AddMemPackage`** — body-first
  ordering closes most of the crash window but baseStore and
  iavlStore have separate WALs, so a kill between the two flushes
  can still leave a half-write. The proper fix is a multistore
  batch commit. Tracked alongside the original
  [`b15ffde6e` follow-up](https://github.com/aeddi/gno/commit/b15ffde6e).
- **Migration drift handling in `r/sys/validators/v2`** — the realm
  itself should not panic on `removeValidator(missing)`. Fixing it
  in-realm would let migration 01 stay declarative. Out of scope for
  test-13 launch (touches a realm we want frozen).
- **Pre-refactor params auto-defaulting** — `buildHardforkGenesis`
  now does this for the seven gas-storage fields by hand. A proper
  solution is `vm.Params.UnmarshalJSON` (or amino post-decode hook)
  applying defaults for any newly-introduced field; that lets every
  consumer of the type benefit, not just the hardfork tool.

## Status

Implemented for the master-based test-13 launch path.
Validated end-to-end on a 4-node gno-cluster boot (replay completes,
chain advances past genesis, all 7 migration txs succeed, post-fork
state matches `assert-migrations` expectations). See PR
[#5597](https://github.com/gnolang/gno/pull/5597) for the rc-by-rc
delta.
