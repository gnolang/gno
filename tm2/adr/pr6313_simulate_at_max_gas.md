# Meter a simulation at the block max gas, then hold it to GasWanted

## Context

`.app/simulate` runs a transaction through the same ante handler as CheckTx
and DeliverTx, and the ante handler meters it with
`SetGasMeter(ctx, tx.Fee.GasWanted)`. A simulation therefore stops where the
caller's guess runs out, and reports how far it got rather than what the
transaction uses.

To get the real figure a client has to raise `GasWanted` before simulating.
`GasWanted` is in the sign bytes, and for messages that carry code
(`add_package`, `run`, `enable_package`) gno.land verifies signatures on the
simulate path (`auth.AnteOptions.RequireSigForSimulate`). So `gnokey` signs a
second copy of the transaction with `GasWanted` set to the block max
(`txWithGasWanted` in `tm2/pkg/crypto/keys/client`). That costs:

- a second signature, which on a Ledger is a second prompt (#6311 removes it
  for Ledgers by giving up the gas figure instead);
- a second, validly signed transaction, handed to whichever RPC node answers
  the simulation;
- a client-side copy of the node's list of messages whose simulation is
  verified (`txNeedsSimulationSignature`), which has to be kept in sync.

Measured against gnodev, an `add_package` signed with `GasWanted` 100,000
that needs 2,637,436 gas simulates as out of gas with gas used **316,135**:
execution stopped partway through, during `DepthSet`.

## Decision

Two changes, both on the simulate path only:

1. **`auth`**: in simulate mode, when `Block.MaxGas > 0`, the ante handler
   meters at `Block.MaxGas` instead of `GasWanted`. The ante handler's own
   out-of-gas message uses the meter's limit, so running out there reads as
   exceeding the block max.
2. **`baseapp`**: after the messages run in simulate mode, if the meter's limit
   is above `GasWanted` and gas used exceeds `GasWanted`, the result becomes an
   out-of-gas error that keeps the real `GasUsed`. The out-of-gas message in
   `runTx`'s recover uses the meter's limit in simulate mode as well.

The same `add_package` now simulates as out of gas with gas used
**2,637,436**. A transaction that fits reports exactly what it did before
(2,637,446 on both nodes, measured).

### Why the verdict cannot change

Running at the block max and comparing afterwards gives the same answer as
running at `GasWanted`, because execution never reads the meter's limit:
outside the meter itself, the only reads of `Limit()` or `Remaining()` in
`tm2/pkg/sdk`, `tm2/pkg/store`, `gno.land/pkg/sdk/vm` and `gnovm/pkg/gnolang`
are the block gas meter in deliver mode and a string-writer budget in
`bounded_strings.go`. The VM consumes from `ctx.GasMeter()` and never reads
`GasWanted`. This is the same assumption `gnokey` already relies on when it
simulates its rewritten copy at the block max.

### Why the check is gated on the meter's limit

`baseapp` is generic: an ante handler may not bound gas at all, and some test
ante handlers return `GasWanted` 0 with an infinite meter. An infinite meter
reports a limit of 0, so the check only fires when an ante handler deliberately
metered past `GasWanted`, which is the case it exists for. A meter at
`GasWanted` cannot be exceeded without panicking first, so the check never
changes its result either.

### Why `MaxGas <= 0` keeps GasWanted

`-1` means no block limit. Metering a public query at "the max" would then mean
an unbounded meter, so the ante handler keeps `GasWanted` there, as before.

## Consequences

- **Not consensus.** CheckTx and DeliverTx are unchanged; the simulate branch
  runs on a context that is discarded. Nodes can upgrade one at a time.
- **Same verdict, better figure.** A transaction that does not fit its
  `GasWanted` still fails simulation, so a client that reads "simulation
  passed" as "it fits" keeps working. On failure, `GasUsed` is now the real
  figure and can exceed `GasWanted`. gnoclient's `EstimateGas` (success only)
  is unaffected; `SimulateResult` callers get the better figure.
- **No new exposure on the public query.** Anyone can already simulate at the
  block max by declaring it. What changes is that an honest simulation of a
  runaway transaction now runs to the block max instead of stopping at the
  caller's guess.
- **Signature verification is unchanged.** The signature is still checked
  against the bytes as sent; only the meter's limit moved.

### Clients

Nothing breaks for an existing client. The gain needs a client change, which
belongs in a follow-up because it touches the code #6311 changes: simulate the
transaction as signed, for every key. Against an upgraded node that gives one
signature and the real figure. Against an old node, a local key can still fall
back to a re-signed max-gas copy when the simulation runs out of gas, so no
capability detection is needed. Once enough nodes are upgraded, the copy and
`txNeedsSimulationSignature` can go.

## Alternatives considered

**Keep doing it in the client.** What #6311 does for Ledgers. It works, but
it costs a second signature and a second valid transaction for local keys, or
the gas figure for Ledgers, and keeps the client in sync with the node's list.

**An infinite meter for simulation.** Rejected: unbounded work on a public
query. The block max is the ceiling any declared transaction already has.

**Fail early instead of reporting the real figure.** That is today's
behavior, and the source of the problem.

**Do the check in the ante handler.** It runs before the messages, so it
cannot see their gas.

## Tests

- `auth.TestAnteHandlerSimulationGasLimit`: deliver meters at `GasWanted`,
  simulate at the block max, and simulate with no block max at `GasWanted`;
  `GasWanted` reported is the tx's own in each case.
- `sdk.TestSimulateMeteredPastGasWanted`: with an ante handler metering at the
  block max, a tx that fits passes with its gas used; one past `GasWanted` fails
  as out of gas with the real gas used; one past the block max runs out at the
  meter and says so.
- `sdk.TestSimulateTx`: its ante handler now declares the limit it meters at as
  `GasWanted`; it declared 0 before, which no real ante handler does.
- Both new tests fail with their production change reverted.
- Manually against gnodev, old node versus new: the measurements above.
