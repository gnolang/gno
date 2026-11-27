# ADR: Taking Exactly K Signatures On A K-of-N Multisig

## Context

A K-of-N multisignature carries one signature per set bit, and verification
used to accept any count of at least K. The sign payload does not say which
constituents sign. So when more than K of them sign, any K of their signatures
verify on their own.

That gives a relayer a rewrite no byte-level rule catches. It can drop the
spare signatures from a pending transaction and broadcast the copy. The copy is
canonical, verifies, and lands under a new hash. The original is then refused
for its used sequence, so a wallet or indexer that tracks the original hash
reports a failure for a transaction that ran. It is the same rewrite as a
stripped public key, see `pr6504_tx_signature_pubkey.md`, reached through the
multisignature instead.

Signing with spares is normal: `gnokey multisign` adds every signature file it
is given, and one gnoland-1 treasury signed a 2-of-4 batch payout with three.

## Decision

A multisignature must carry exactly K signatures. The rule lives in
`Multisignature.ValidateBasic`, which now takes the threshold key instead of
its constituent count. Both places that walk a multisignature already call it:
`PubKeyMultisigThreshold.VerifyBytes` and the ante handler's gas consumer. So
they cannot disagree, and the gas consumer refuses an over-signed level before
charging for that level's constituents. Nested multisigs follow the same rule at
every level, since both walks recurse through those same calls.

`VerifyBytes` no longer has a separate "at least K" check; `ValidateBasic`
subsumes it.

Both places also decode the multisignature with `amino.UnmarshalCanonical`
(gnolang/gno#6404). The multisignature is amino nested in
`Signature.Signature`, which gnolang/gno#6406's canonical check of the
transaction sees as opaque bytes. So a relayer could otherwise pad a varint in
it, `08 03` to `08 83 00` in the bit array, and land the same signatures under
a new hash: the same rewrite as dropping a spare, by encoding instead. Nested
levels decode through the same two calls.

`gnokey multisign` checks the combined multisignature, and every nested
multisignature a signature file brings into it, with the same `ValidateBasic`.
It refuses, leaving the tx file unchanged, when any level has more or fewer
than its K signatures. With more, the user picks which K to send.

## Alternatives considered

- *Document it for integrators*: tell them to track multisig transactions by
  sender and sequence, not hash. That leaves multisig accounts as the one
  signer whose hash a relayer can change, which every integrator would have to
  special-case.
- *Have `gnokey multisign` keep the first K silently*: it would not stop other
  clients from producing spares, and the chain would still accept them.
- *Cover which constituents sign with the sign payload*: no K-subset would
  verify, whoever made it, including the node that first receives an
  over-signed transaction. But co-signers would have to agree on the set before
  anyone signs. It would also change the sign payload of every multisig
  transaction, and with it Ledger signing, every wallet that signs one, and the
  verification of the multisig signatures already on chain.
- *A height gate*, so one binary can re-execute all of history: `VerifyBytes`
  sees no height, so the rule would have to leave `ValidateBasic` for the ante
  handler, and `VerifyBytes` and the gas consumer would disagree again, which
  is what sharing `ValidateBasic` prevents. gnolang/gno#6406, which refuses
  most of history, implemented a gate and dropped it in favour of switching
  binaries at halt heights; this follows it.

## Consequences

- **Consensus-breaking, shipped with the UTF-8 and public-key checks.** A
  multisig transaction with more than K signatures was valid and is now
  refused. It ships in the same halt-height binary switch, so blocks below the
  halt height must be run by the earlier binaries; see below.
- **One historical transaction would be refused.** Of the 91 multisig
  transactions among all 233,680 gnoland-1 transactions up to height 1,182,449,
  one carries more than K signatures: a 2-of-4 signed by three, at height
  599,791. The other 90 carry exactly K, none is nested, and every
  multisignature is canonically encoded. Tools that re-verify history with
  this library, such as `gnokey verify` or an indexer calling `VerifyBytes`,
  now report that signature as invalid although the transaction ran.
- **So history needs the earlier binary.** The UTF-8 and public-key checks
  pass every historical transaction; this rule does not.
  - A node joining gnoland-1 from genesis on this binary alone refuses that
    payout, so its state differs from 599,791 on and it halts on an AppHash
    mismatch at the next block. The mainnet joining guide's single-binary path,
    keeping the newest binary while it meets each halt's `halt_min_version`,
    no longer works across this halt: nothing gates by height, since the
    startup check reads only the halt params already in state. The release's
    ledger entry and joining notes must say to run the previous version up to
    this halt height and switch there. gnolang/gno#6406 needs the same for most
    of history; see the height gate under *Alternatives considered*.
  - A hardfork genesis replay that verifies signatures (`IsGenesisReplay` with
    `VerifyGenesisSignatures`) would refuse it too. Unlike the public-key
    check, this rule cannot exempt replay: `VerifyBytes` and the gas consumer
    see no context. Mainnet replays with `--skip-genesis-sig-verification`,
    which skips both.
- **Simulation is stricter.** `.app/simulate` runs the gas consumer always,
  and skips `VerifyBytes` unless the application's `RequireSigForSimulate`
  selects the transaction (gno.land selects those that compile code:
  `MsgAddPackage`, `MsgRun` and `MsgEnablePackage`; see `txCarriesCode`). So a
  placeholder multisignature with fewer than K entries, such as an empty
  `NewMultisig(n)`, used to pass simulation and is now refused. No client in
  this repository builds one. External wallets that estimate gas that way must
  put K entries in the placeholder at every level; a transaction
  `RequireSigForSimulate` selects still needs K real signatures, as before.
  Wherever `VerifyBytes` runs, a real under-K multisignature is now refused
  before its short level's constituents are charged, rather than after.
- **Signers with spares must pick K.** `gnokey multisign` now says so instead
  of producing a transaction the chain refuses. Other clients that assemble
  multisignatures must do the same.
- **Clients must not produce more than K signatures.** A transaction signed by
  more than K is refused at CheckTx and so is not gossiped.
- **A multisignature must be canonically encoded.** One that amino would encode
  differently is refused, as not a valid multisignature. Every multisignature
  on gnoland-1 is canonical. `gnokey multisign` writes canonical bytes and
  refuses a signature file whose nested multisignature is not, and the
  simulation `gnokey broadcast` runs zeroes one; other clients that assemble
  multisignatures must encode them as amino does.
