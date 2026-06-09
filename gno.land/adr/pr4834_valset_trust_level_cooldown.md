# PR4834: IBC trust-level + cooldown limits on valset changes (v3)

## Context

`r/sys/validators/v3` lets GovDAO proposals change the validator set
via `NewValidatorProposalRequest`. Without the trust-level + cooldown
gate described here, nothing stops a single approved proposal from
replacing the entire valset in one shot, or from chaining many
updates in quick succession.

Two practical risks follow from that:

1. **IBC light clients break.** A light client that trusts a historical
   header verifies later headers by checking that enough of the
   trusted set's voting power signed them. If a single chain update
   removes most of the trusted set's power from the new set, no
   trusted light client can bridge from before-the-update to
   after-the-update. The bridge requires an intermediate hop the chain
   never produces.

2. **Operational instability.** Even with each update individually
   safe, a tight loop of updates makes consensus unstable and gives
   light clients no time to catch up. CometBFT itself doesn't enforce
   a minimum gap between validator-set changes — it's a policy choice
   the chain has to express.

The issue (#4829) asks for two specific rules, mirroring
ibc-go's restriction on the Tendermint trust-level fraction:

> "If we limit the valset changes to max 2/3 of the set per update,
> light clients can accept skipped updates if the trust level is
> maintained (at least 1/3 of the set is the same). The 2/3 rule can
> be easily implemented in Gno, on the contract level."

Plus a 24h cooldown between consecutive updates.

## Decision

Add two safety rules to `r/sys/validators/v3`, both checked at
proposal creation **and** re-checked at execution, both
governance-configurable through dedicated `dao.ProposalRequest`
factories.

### Trust-level rule

For every accepted valset-change proposal, the validators that were
in the baseline (the effective valset at execute time) and remain
in the new set must retain at least `trustLevelRatio` of the
**baseline's** voting power.

The check projects the CometBFT light-client rule
([`types/validation.go`](https://github.com/cometbft/cometbft/blob/main/types/validation.go))
onto a per-update guard:

| CometBFT property | This PR |
| --- | --- |
| `talliedVotingPower += val.VotingPower` where `val ∈ trustedVs` | sums `baselineByAddr[addr]` (OLD voting powers) |
| `vals.TotalVotingPower()` where `vals = trustedVs` | `baselineTotal` |
| `if got <= needed → ErrNotEnoughVotingPowerSigned` | `if left <= right → panic` (strict `>` to pass) |
| `[1/3, 2/3]` (ibc-go layer) | `trustLevelMinAllowed`/`trustLevelMaxAllowed` bounds |

The ratio is **snapshotted at proposal-creation time** and captured
in the executor closure. A same-block governance proposal that loosens
the ratio cannot retroactively relax an in-flight valset proposal —
the executor checks the snapshot, not the package-level value.

Configurable via `NewTrustLevelPropRequest(num, den, title, desc)`.
Bounds are enforced at create and re-checked at execute.

### Cooldown rule

`NewValidatorProposalRequest` rejects creation if
`time.Since(lastValsetUpdate) < valsetUpdateCooldown`. The executor
re-checks the same predicate before applying changes (binding source
of truth — a proposal can sit in GovDAO while another lands first).
`lastValsetUpdate = time.Now()` is written **after** the successful
`SetValsetProposal` call, so a panic mid-publish doesn't advance the
cooldown clock.

Default is 24h. Configurable via `NewCooldownPropRequest(seconds,
title, desc)`. Input is bounded by `cooldownMaxSeconds = 365 * 86400`
(1 year) to keep `time.Duration(seconds) * time.Second` well inside
int64 range — beyond ~9.22e9 seconds the product overflows and could
wrap to a tiny or negative value, effectively disabling the cooldown.

Unlike the trust level, the cooldown is **not** snapshotted at
creation time. Reasoning: the trust-level snapshot blocks adversarial
governance from loosening an IBC safety invariant mid-flight.
Cooldown is an operational rate-limit, not a safety invariant; the
trust-level rule still blocks any actually-unsafe valset change at
execute time. The live read also lets the test environment shorten
the cooldown and immediately use it.

### Overflow handling

Both rule checks use `math/overflow.Mulu64` for the cross-multiply
so any uint64 overflow surfaces as a deterministic panic
(`errTrustLevelOverflow`) rather than a silently-wrong comparison.

## Alternatives Considered

- **Numerator weighted by NEW voting power** (the original draft).
  Rejected because it allows a false-positive accept: a proposal that
  removes 3 of 4 baseline validators and upserts the survivor's NEW
  VP to a huge value would pass the chain check but be rejected by
  an IBC light client (which tallies baseline VPs). Regression test
  `TestNewValidatorProposalRequest_RejectsBaselineWipeViaUpsert`
  pins the correct behavior.

- **Non-strict inequality** (`pass iff retained * den >=
  total * num`). Rejected: CometBFT's `if got <= needed → error`
  fails at the boundary. Chain accept + light-client reject = same
  class of bug as the NEW-weights case.

- **No governance setter for cooldown** (the cooldown is hardcoded
  at 24h). Rejected: integration tests can't drive back-to-back
  valset transitions in real time, and a chain reasonably wants the
  ability to tune the operational rate without a binary upgrade.
  A separate `NewCooldownPropRequest` mirrors the existing
  `NewTrustLevelPropRequest` pattern.

- **Cooldown snapshot at create time** (mirror trust-level snapshot).
  Rejected: would lock in a possibly-too-long cooldown for in-flight
  proposals even after governance shrinks it. The trust-level
  snapshot is a safety lock (loosening is bad); the cooldown
  direction is opposite (loosening is intentional and welcome).

- **Configurable cooldown bounds without an upper cap.** Rejected:
  pasting a typo'd huge number could overflow `time.Duration` arithmetic
  and silently disable the rule.

## Consequences

**Positive**:
- Chain-side proposals that would break IBC light clients are
  refused before they leave the realm.
- Sequential valset thrashing is prevented by default.
- Both rules are tunable through governance.

**Negative / Tradeoffs**:
- Integration tests doing back-to-back valset operations have to
  prepend a `NewCooldownPropRequest(0, ...)` proposal — slightly more
  setup per test. Reflected in the updated `power_update` and
  `remove` txtars.
- A governance majority that ratifies both a cooldown shrink and a
  loosened trust level can still force frequent changes within the
  [1/3, 2/3] envelope. The PR doesn't (and arguably shouldn't)
  defend against adversarial governance; the [1/3, 2/3] floor is the
  IBC honest-majority floor itself.
- A `lastValsetUpdate == time.Time{}` (genesis, never updated) yields
  a huge `time.Since`, so the cooldown is skipped on first-ever
  update. Intentional — there's no prior update to gate against.

## Key files

| File | Role |
| --- | --- |
| `examples/gno.land/r/sys/validators/v3/limits.gno` | State, `trustRatio`, getters, governance proposal factories, executor helpers |
| `examples/gno.land/r/sys/validators/v3/limits_test.gno` | Unit tests for both rules + governance setters |
| `examples/gno.land/r/sys/validators/v3/proposal.gno` | Cooldown + trust-level checks wired into `NewValidatorProposalRequest` / `newValoperChangeExecutor` |
| `gno.land/pkg/integration/testdata/params_valset_trust_level_proposal.txtar` | End-to-end for `NewTrustLevelPropRequest` |
| `gno.land/pkg/integration/testdata/params_valset_proposal_{power_update,remove}.txtar` | Restructured to disable the cooldown as proposal 0 before exercising back-to-back transitions |
