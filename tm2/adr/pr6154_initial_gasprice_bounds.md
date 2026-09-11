# PR #6154: Bound governance gas prices and preserve their ratio

## Context

The scalar `auth:p:initial_gasprice` setter now takes effect. Unchecked
foreign denominations can reject all normal mempool fees, and int64-sized
components can overflow subsequent EndBlock price growth. The existing floor
conversion handled unequal gas units but retained the old denominator and
rounded fractional floors upward.

## Decision

Validate both scalar governance writes and complete Params writes through
`Params.Validate`: require `ugnot`, nonnegative components, positive gas for a
nonzero amount, and at most 10^12 for each component. Preserve the unset zero
value and zero-price disabled configurations. The maximum ratio permits one
million GNOT per gas and leaves over six orders of int64 headroom.

Input bounds alone cannot contain repeated congestion. Cap the dynamic
numerator at the same limit before converting the big.Int result to int64.
During decay, compare the candidate with the initial price by exact big.Int
cross-products; when it reaches the floor, adopt the complete initial price,
including Gas. Zero stored prices and disabled/degenerate targets retain their
existing behavior.

Keep GasPrice JSON object decoding atomic via the existing temporary alias;
regression tests exercise errors before and after partially decodable fields,
as well as partial and complete successful objects. Fix the existing scoped
`err` compilation error in params decoding to enable validation.

## Alternatives considered

- Parser validation would impose chain policy on the general std gas price
  parser and miss complete Params writes.
- Input bounds without a dynamic cap still overflow after repeated growth.
- Converting the floor to the stored denominator can round the effective floor
  and needlessly retain the old units.
- A test-specific genesis override is unnecessary: the actual txtar node
  construction already initializes a nonzero 1ugnot/1000gas price.

## Consequences and validation

This changes consensus price calculation and restricts accepted governance
values; validators must deploy it together. Existing stored invalid parameters
are not migrated. The cap deliberately stops further dynamic increases at
10^12; changing that economic ceiling requires a code change.

Regression tests cover rejected setters without state mutation, subsequent
EndBlocks and mempool fee acceptance, repeated saturated blocks, exact floor
adoption (including cross-products beyond int64), and atomic JSON decoding.
The governance txtar observes state before the proposal, restarts the node,
queries the effective price, rejects a transaction one ugnot below its fee
floor, and accepts one exactly at the floor.
