# Sign a transaction once on a Ledger when gnokey simulates it

## Context

`gnokey maketx ... -broadcast` simulates a transaction before broadcasting it
(`-simulate test`, the default), or simulates it and stops (`-simulate only`).
Simulation raises `GasWanted` to the consensus max gas, so the estimate is not
clipped by whatever the caller guessed.

`GasWanted` is part of the signed bytes. For `add_package`, `run` and
`enable_package`, gno.land verifies signatures on the simulate path
(`auth.AnteOptions.RequireSigForSimulate`), so since #6088 `maketx` signs the
rewritten copy as well (`txNeedsSimulationSignature`). That made every one of
those transactions two signatures:

- `-simulate test`: the transaction to broadcast, and a simulation copy whose
  only use is a more precise out-of-gas message.
- `-simulate only`: the same two, though nothing is broadcast, so the first
  signature is never used.

A local key pays nothing for a signature. A Ledger pays a prompt, and the second
prompt shows the same transaction with a different gas figure. A signer who
learns that the second prompt is "the same thing again" has learned to approve
without reading, and that copy is itself a valid transaction. With the
proposed digest sign mode, where the signer compares a hash against the device
screen, it would be two comparisons.

## Decision

`planSignatures` decides what `maketx` signs. Nothing is signed that the
broadcast step does not use:

| key | `-simulate` | signs | simulates |
|---|---|---|---|
| any | `only` | the simulation copy | the copy, at max gas |
| local | `test` | the tx and the copy | the copy, at max gas |
| Ledger | `test` | the tx | the tx as signed |
| any | `skip` | the tx | nothing |

Transactions that do not need a simulation signature, and those whose
`GasWanted` is already at the max, are unchanged: one signature.

A Ledger with `-simulate test` gives up the gas figure past its own
`GasWanted`. Running out of gas still blocks the broadcast, and the node's
out-of-gas log already says to simulate with the consensus maximum. The
"suggested gas-wanted (gas used + 5%)" hint is dropped when a simulation ran at
the tx's own `GasWanted` and used all of it, because that gas used is a floor;
this also covers the existing case where the max gas fetch failed.

The broadcast tx is signed before waiting on the max gas fetch, except with
`-simulate only`, so the fetch still overlaps the device prompt.

## Alternatives considered

**Broadcast the max-gas copy and sign only it.** Rejected:
`EnsureSufficientMempoolFees` requires `gas-fee / gas-wanted` to meet the
minimum gas price, so a tx declaring the block's max gas needs a fee scaled to
it or is refused.

**One signature for every key.** Rejected: a local key would lose the precise
out-of-gas figure for nothing.

**Meter simulation at max gas on the node.** The root cause is that the
simulate path meters with `tx.Fee.GasWanted`
(`auth.SetGasMeter`). Metering a simulated tx at the block max, and reporting
gas used against the tx's own `GasWanted`, would give every key one signature
and an unclipped estimate, and remove the simulation copy from the client
entirely. It is the better long-term fix, but it changes the shared ante
handler and what `.app/simulate` reports, and the client cannot tell an
upgraded node from an old one, so this client change is needed regardless.
Worth its own issue.

## Consequences

- A Ledger signs once for every `maketx -broadcast`.
- `-simulate only` signs once for every key.
- A Ledger that runs out of gas under `-simulate test` is told so at its own
  `GasWanted`, and has to rerun with `-simulate only` for the figure.
- `txNeedsSimulationSignature` still duplicates the node's predicate; more now
  depends on it agreeing, and drift still fails loudly as a signature error.
- No node, wire or consensus change.
