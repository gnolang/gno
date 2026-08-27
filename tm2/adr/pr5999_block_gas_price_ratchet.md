# PR5999: The block gas price must not ratchet, panic, or halt the node

## Status

Proposed

## Context

`GasPriceKeeper.calcBlockGasPrice` recomputes the minimum gas price at the end of
every block. It derives a target from the block gas limit,
`targetGas = MaxGas * TargetGasRatio / 100`, divides by that target in both the
increase and the decrease branch, and multiplies the price by a ratio that has no
upper bound. Three shapes of configuration and load break it.

**`Block.MaxGas == -1`, the "no gas bound" sentinel.** Nothing on this path maps
it to "unbounded", unlike `BaseApp.getMaximumBlockGas` and `NewAnteHandler`.
`big.Int.Div` floors, so the target is `-1`, every `gasUsed >= 0` compares above
it and takes the increase branch, and there the intermediate quotient is negative
so the min-1 floor clamps it to `+1`. The price rises by 1 on every block,
including idle ones, and the decrease direction is unreachable.

**`Block.MaxGas * TargetGasRatio < 100`**, which at the default ratio of 70 means
`MaxGas` 0 or 1. The target is 0 and the first non-empty block divides by zero
inside `EndBlock`, halting the node. `MaxGas == 0` is the other "unbounded"
spelling: `getMaximumBlockGas` turns it into an infinite gas meter, so blocks do
consume gas and the division is reached.

Both pass `ValidateConsensusParams`, which only rejects `MaxGas < -1`, and both
are one JSON field away from a genesis the tooling produces:
`ConsensusParams.Update` copies a non-nil `Block` whole rather than field by
field, so a genesis whose `consensus_params.Block` omits `MaxGas` validates with
`MaxGas == 0` and no default backfill.

**Sustained congestion.** At the shipped parameters (`MaxGas` 3e9, ratio 70,
compressor 10) a completely full block raises the price by about 4.29%, so from a
price of 1 the `IsInt64` check fires after 1,002 full blocks and `EndBlock`
panics on every node at the same height. `Params.Validate` also accepts a
`TargetGasRatio` of 1 with a `GasPricesChangeCompressor` of 1, which reaches the
same point after nine full blocks.

A fourth, smaller shape: `Params.Validate` accepts an `InitialGasPrice` of 0, and
a stored price of 0 is read at the top of `calcBlockGasPrice` as "dynamic pricing
disabled". A chain whose floor is 0 therefore walks down to 0 and can never rise
again.

## Decision

Three changes, all inside `calcBlockGasPrice` except the last.

1. **A non-positive target returns the price unchanged.** The guard sits after
   the target is computed and before either branch is chosen, so a chain with a
   positive target reaches exactly the code that ran before. It covers the `-1`
   sentinel and the rounding case in one condition, because both mean the same
   thing: there is no congestion signal to price against.
2. **The decrease floor is `max(InitialGasPrice, 1)`.** The price can still decay
   to the configured floor, but never into the absorbing 0 state.
3. **The int64 overflow clamps instead of panicking.** The price saturates at
   `math.MaxInt64` and the node keeps producing blocks. Only the increase branch
   can produce a non-int64 value: every decrease result lands in
   `[1, max(lastPrice, initialPrice)]`, both ends of which are already int64.
4. **`UpdateGasPrice` logs at `ERROR` when the new price is the ceiling.** At that
   price no transaction wanting more than 1,000 gas can pay the minimum fee, and
   the skip-when-unchanged write means the price stops moving and the telemetry
   hook stops firing, so nothing else in the node reports the state.

## Alternatives considered

- **Map `MaxGas` -1 and 0 to "unbounded" here, the way `getMaximumBlockGas`
  does.** Rejected: an unbounded limit has no target, so the controller has
  nothing to steer towards. Freezing the price is what "no target" means; picking
  an arbitrary substitute target would invent a congestion signal.
- **Reject `MaxGas` 0 and -1 in `ValidateConsensusParams`.** Rejected as the fix
  for this PR: it makes an existing genesis invalid, and a node should not halt
  on a configuration the validator accepted. Worth doing separately as a
  tooling-level guard.
- **Keep the panic and cap the price at a governance parameter.** A policy
  ceiling below the int64 one is a real design question, and the `XXX` in the
  increase branch still asks it. It is not this PR: a policy cap needs a new
  parameter, a migration and a governance decision, while the int64 ceiling is a
  property of the type and must be handled either way.
- **Return an error from `calcBlockGasPrice`.** Rejected: it runs in `EndBlock`,
  where there is no caller that can do anything but panic.

## Consequences

- A chain with no usable target keeps the price it starts with, which on such a
  chain is the configured `InitialGasPrice`.
- A chain under sustained congestion reaches a price no transaction can pay
  rather than halting. That state is not absorbing: with the mempool rejecting
  everything, blocks go idle and the price returns to the floor in 407 blocks at
  the shipped parameters. It is a bounded outage where the previous behavior was
  a coordinated crash.
- The change is consensus-visible exactly where the previous code panicked. For
  every configuration with a positive target, a non-zero initial price and a
  valid compressor, the function is output-identical to before: a 91,584-case
  differential against the merge base classifies all 19,585 divergent cases into
  those three intended causes, with none unexplained, and the merge base panics
  in all 14,976 cases where the new code does not.
- A rolling upgrade is safe on gno.land's own configuration. A `MaxGas` 0 chain
  is already halted on the old binary, so there is no live network to fork; a
  `MaxGas` -1 chain keeps producing blocks with a climbing price, so mixed
  binaries there would diverge and it needs a coordinated restart. No
  configuration in the tree sets -1.
- Setting `auth:p:initial_gasprice` to a zero-amount price no longer disables
  dynamic pricing by letting the price decay to 0.

## Tests

`tm2/pkg/sdk/auth/keeper_test.go`:

- `TestCalcBlockGasPriceUnboundedMaxGas`: `MaxGas` -1 does not ratchet on idle
  blocks; `MaxGas` 0 and 1 do not panic at any usage; the smallest `MaxGas` with
  a positive target still adjusts the price.
- `TestCalcBlockGasPriceZeroInitialPrice` and
  `TestCalcBlockGasPriceFloorAboveOne`: the floor is 1 when the initial price is
  0, and the initial price itself when that is higher.
- `TestCalcBlockGasPrice/int64 overflow caps instead of panicking`,
  `/sustained congestion settles at the cap` and `/decreasing stays in range`:
  the clamp is reached, keeps the denominator and denom, returns to the floor
  under a bounded number of idle blocks, and is unreachable from the decrease
  branch even at an initial price of `math.MaxInt64`.
- `TestUpdateGasPriceCeilingLogs`: the ceiling is logged, and nothing is logged
  below it.

Each of these fails against the previous implementation, by panic or by
assertion, except the floor-above-one case, which guards the floor against a
future simplification.
