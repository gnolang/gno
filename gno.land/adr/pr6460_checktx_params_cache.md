# Decode ante params once per committed height in CheckTx

## Context

The gno.land ante wrapper reads the auth params (`acck.GetParams`) and the vm
params (`vmk.GetParams` plus `ApplyToGasConfig`) on every transaction it
sees. Both reads amino-unmarshal every field and bech32-decode every
allowlist entry. Measured on a gno.land app with the real `CListMempool` at a
5000-deep mempool, the two decodes are about 305 us and 37% of a recheck,
more than signature verification. With recheck on, every pending transaction
pays them again after every block, although the params cannot have changed
between two commits from CheckTx's point of view.

## Decision

Memoize the decoded `anteParams` (auth params, vm params, derived gas config)
for `RunTxModeCheck` only, keyed on the CheckTx context's block height
(`checkTxParams` in `gno.land/pkg/gnoland/anteparams.go`).

- **Invalidation is the height.** BaseApp replaces the check state only in
  InitChain, Commit and on load, each time with the header of the block just
  committed, and CheckTx never runs a message handler, so the params store
  seen by CheckTx is constant for a given height. A params change made in
  block H reaches CheckTx through the check state of height H, which misses.
- **DeliverTx and simulate always decode.** A params change inside a block
  is seen by the next transaction of the same block exactly as before.
- **Genesis (height 0) always decodes.**
- **Gas.** Both reads already bypass the gas meter, so skipping them changes
  no gas.

## Alternatives considered

- Cache in DeliverTx too, invalidated on every params write. Saves more but
  puts a cache on the consensus path and needs a write hook in the params
  keeper. Not needed for the recheck cost.
- Cache in the params keeper keyed on raw bytes. Still pays the store reads
  and touches every module; larger change.
- An explicit Commit hook in BaseApp. Equivalent to the height key, wider API.

## Consequences

- One decoded copy of the params per app instance. Callers must treat it as
  read-only: the vm params hold slices.
- A future code path that writes params in CheckTx mode would make the
  height key stale; the type's doc comment states the invariant it relies on.
