# PR5954: Enforce the block gas price in DeliverTx

## Status

Proposed

## Context

The dynamic block gas price check ran inside `EnsureSufficientMempoolFees`,
which the auth ante handler called only during CheckTx. A proposer could bypass
its mempool and include a transaction below the consensus `LastGasPrice`; the
transaction would then execute during DeliverTx.

That function also checked validator-local `MinGasPrices`. Those values are
node configuration and cannot safely affect DeliverTx, because validators with
different local settings must produce the same consensus result.

`InitialGasPrice` is stored under `GasPriceKey` during `InitChain`. `EndBlock`
updates that value, and `LastGasPrice` returns the current consensus gas price.

## Decision

Split the checks by responsibility:

* `EnsureSufficientBlockGasPrice` compares the transaction gas price with
  `LastGasPrice` during both CheckTx and DeliverTx. It continues to use
  `std.GasPrice.IsGTE` and treats a zero or inactive block price as before.
* `EnsureSufficientMempoolFees` checks only validator-local `MinGasPrices` and
  remains CheckTx-only.
* Simulation skips both checks.

The ante wrapper installed by `NewAppWithOptions` loads `LastGasPrice` into
`GasPriceContextKey` for every `runTx` invocation (CheckTx, ReCheckTx, and
DeliverTx). Because `EndBlock` updates the stored price only after all
DeliverTx executions, every transaction in a block observes the same block gas
price.

## Alternatives considered

* Keep a single CheckTx-only validation function: rejected because it leaves
  the proposer mempool bypass open.
* Apply `MinGasPrices` in DeliverTx: rejected because node-local configuration
  would make transaction results validator-dependent.
* Add an activation height or migration: rejected because this change starts
  from a fresh testnet genesis.

## Consequences

A transaction accepted by CheckTx can fail in DeliverTx if the consensus block
gas price rises before inclusion. Transactions below the current block price
fail in ante processing before message execution or fee deduction.

After a price increase is committed, ReCheckTx uses the updated
`LastGasPrice` and may evict transactions admitted under the previous price.

A transaction rejected during DeliverTx still occupies bytes in the proposed
block without paying a fee. This matches existing handling of other ante
failures, such as insufficient funds; it is not a new block-processing policy.

Deployment on existing chains is out of scope. Before enabling this change on a
live network, the need for an activation height or migration must be
re-evaluated.

The gas price calculation itself—including fixed-scale rescaling, the canonical
denominator, and the increase/decrease curve—is unchanged.

## Validation

The existing CheckTx-only regression test is renamed to
`TestBlockGasPriceMinimumIsEnforcedInDeliverTx`, and its DeliverTx assertion is
reversed to require rejection. `TestGasPriceUpdate` also verifies that
transactions priced for the previous block are rejected after the block gas
price increases.
