# ADR: Simulating Without A Broadcastable Signature

## Context

gnokey estimates gas by sending the transaction to the `.app/simulate` query
of whichever RPC the user configured. It signed the transaction first and sent
it with its signature attached, so that RPC received a transaction it could
broadcast, even for `--simulate only`, a failed simulation, or
`gnokey broadcast -dry-run`.

The RPC already sees the transaction contents during simulation and would see
the signed transaction during an actual broadcast. The extra exposure is the
signature from a transaction the user did not intend to send. Raising
`GasWanted` for simulation does not close it: a malicious RPC can refuse the
maximum-gas query and receive the original transaction, or try likely original
values against the signature.

## Decision

A simulated transaction carries a valid signature only when the chain needs
one. gno.land verifies signatures during simulation only for code-bearing
messages (`add_package`, `run`, `enable_package`; see
`RequireSigForSimulate`). Every other transaction is simulated with its
signature bytes zeroed.

The placeholder preserves the public key when present, the session address,
and the encoded length. A multisig also preserves its structure because the
gas consumer decodes and walks it. This keeps the inputs to signature and
per-byte gas accounting unchanged while making the transaction unbroadcastable.

The blanking happens where the simulation bytes are built, in
`buildSimulationTxBytes`. That one place covers both entry points:
`gnokey maketx`, and `gnokey broadcast -dry-run` for transactions signed
elsewhere.

## Alternatives considered

- *Strip the signatures entirely*: the tx becomes shorter, so simulation
  under-charges the per-byte gas and an estimate taken this way can run out of
  gas when delivered. Multisig gas cannot be computed from an empty signature.
- *Sign only after simulating*: avoids producing a signature for `maketx`, but
  does nothing for `broadcast -dry-run`, whose input is already signed, so
  blanking would still be needed.
- *Bind the signature to the chain or add an expiry*: changes the sign bytes,
  a consensus change, out of proportion to this.

## Consequences

For sends, calls and other non-code messages, the simulating RPC holds nothing
it can broadcast.

The rule that decides which messages need a signature is still duplicated
between gnokey and the chain. If the two drift, the failure is loud: a newly
gated message fails simulation with a signature error.
