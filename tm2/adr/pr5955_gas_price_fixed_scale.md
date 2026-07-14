# PR5955: Canonical fixed-scale block gas price

## Context

The dynamic block gas price was stored as `1ugnot/1000gas` at its default
floor. The controller's minimum upward tick therefore changed the floor
numerator from `1` to `2`, raising the rate by 100% even when block use differed
from the target by one gas. Ethereum's 12.5% bound is a maximum full/empty-block
adjustment, not a price-resolution requirement; representation precision and
controller rate must remain separate.

## Decision

Store every nonzero consensus dynamic block gas price with the fixed gas
denominator `1_000_000`. This change targets a fresh testnet, so genesis and
configuration must provide that representation exactly; runtime and legacy
state are not rescaled. Thus the unchanged default rate is
`1000ugnot/1000000gas`, and one numerator atom is 0.1% of that floor. The
existing proportional formula, target ratio, compressor, two integer
divisions, and minimum one-atom change remain unchanged.

Genesis and governance validation reject nonzero prices with another scale,
non-positive amounts, invalid denominations, and governance denomination
changes. `gnogenesis params set` validates the resulting genesis state before
saving it. Disabled pricing uses the combined zero gas and zero amount form.
The keeper changes only the numerator and clamps decreases to the initial
floor. `std.GasPrice` remains general-purpose and fee validation continues to
use its cross-multiplication. Gnokey first adds its 5% suggested-gas buffer,
then estimates the fee for that `GasWanted` with multiply-first ceil division
using `big.Int` before applying the existing configurable fee margin.

## Alternatives considered

- Remainder consensus state would preserve sub-atom changes but add state and
  migration complexity.
- A new public fixed-point price type would duplicate `std.GasPrice` and widen
  the API change.
- Runtime ceil conversion would silently change configured rates and add
  migration behavior that a fresh chain does not need.
- Changing target ratio, compressor, or the controller curve would mix pricing
  policy with this representation-only change.

## Consequences

The `auth/gasprice` JSON shape is unchanged, but consumers must treat `gas` as
data rather than assume `1000`; the default response becomes
`{gas: 1000000, price: "1000ugnot"}`. External wallets and explorers need only
preserve ratio-based handling. The numerator remains an `int64`; a controller
increase beyond that range deterministically panics rather than wrapping or
introducing a configurable economic cap. Parameter edits through `gnogenesis`
are saved only when the complete resulting genesis state is valid.

Future target-ratio, compressor, curve, and telemetry changes are separate
work.
