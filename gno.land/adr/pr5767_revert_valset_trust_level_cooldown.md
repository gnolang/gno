# Revert chain-side valset trust-level and cooldown limits

## Context

PR #4834 added two checks to `r/sys/validators/v3`:

1. A validator-set update had to retain enough voting power from the
   previous set to satisfy an IBC light-client trust level.
2. Successful validator-set updates had to be separated by a
   governance-configurable cooldown, defaulting to 24 hours.

The checks were based on an incorrect ownership assumption. Tendermint
light-client verification distinguishes adjacent and non-adjacent
headers. `VerifyNonAdjacent` checks that enough trusted voting power
signed the later header. `VerifyAdjacent` instead verifies that the
new header's validator hash matches the previous header's committed
next-validator hash and does not apply the trust-level overlap check.

When a relayer cannot verify a skipped update directly, it can submit
intermediate headers until verification reaches adjacent heights. A
chain-side overlap restriction therefore rejects validator-set
transitions that the light-client protocol permits.

The 24-hour cooldown is also not required by the IBC or CometBFT
specification. In particular, it prevents future AtomOne Interchain
Security (ICS) provider-validator-set mirroring from applying updates
to gno.land when AtomOne changes its validator set without a matching
cooldown constraint. This is separate from IBC light-client
verification.

Relevant references:

- [ICS-07 Tendermint Client](https://github.com/cosmos/ibc/blob/main/spec/client/ics-007-tendermint-client/README.md)
- [CometBFT light verifier](https://github.com/cometbft/cometbft/blob/main/light/verifier.go)

## Decision

Remove the chain-side trust-level and cooldown checks from
`r/sys/validators/v3`.

This removes:

- the trust-level overlap check from validator-set proposal execution;
- the cooldown checks from proposal creation and execution;
- the governance proposal factories for changing those values;
- the dedicated unit and integration tests;
- test-only cooldown-disable proposals from integration scenarios.

Keep validator-set validation added independently of these limits,
including operator deduplication, KeepRunning opt-out enforcement,
signing-key re-resolution, and the empty-set liveness check.

## Alternatives Considered

- **Remove only the cooldown.** Rejected because the trust-level
  overlap restriction still moves relayer responsibility into chain
  governance and rejects protocol-valid adjacent transitions.
- **Keep the settings but default them to disabled.** Rejected because
  the settings expose unsupported policy knobs and retain unnecessary
  code paths.
- **Require relayers to update every block.** Rejected because the
  light-client protocol already supports skipped updates and
  intermediate-header bisection.

## Consequences

Validator sets may change without a chain-enforced cooldown and may be
fully replaced between adjacent blocks. Relayers are responsible for
submitting intermediate headers when a skipped light-client update
cannot satisfy the configured trust level directly.

The validator realm becomes smaller and no longer exposes
`GetTrustLevel`, `NewTrustLevelPropRequest`, `GetCooldown`, or
`NewCooldownPropRequest`.
