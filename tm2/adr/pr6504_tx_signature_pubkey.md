# ADR: Requiring The Public Key In Every Transaction Signature

## Context

`std.Signature.PubKey` is optional once the signer's account has stored a key:
the ante handler falls back to the stored one. The sign payload does not include
the public key, so a signature verifies the same with or without it.

That gives a relayer a rewrite no byte-level rule catches. It can drop the key
from a signature in a pending transaction and broadcast the copy. The copy is
canonical, decodes to a transaction that differs only in the dropped key,
verifies, and lands under a new hash. The original is then refused for its used
sequence, so a wallet or indexer that tracks the original hash reports a failure
for a transaction that ran.

## Decision

The ante handler refuses, with `ErrInvalidPubKey`, a transaction with a
signature that has no public key. The check runs right after
`tx.ValidateBasic`, before any fee or sequence change, so a refused copy costs
the signer nothing. It covers every signature, sessions included.

Simulation and genesis are exempt, because nothing is relayed there:

- Simulation lets a client leave the key out and rely on the stored one.
- Genesis means height 0 or `IsGenesisReplay`. Replayed history carries its
  original height, so a height check alone would miss it.

Height 0 also covers CheckTx between the genesis commit and the first block,
since `InitChain` sets the check state with a header of no height. A stripped
copy can enter the mempool in that window, but DeliverTx at height 1 or more
refuses it before any fee or sequence change, and recheck evicts it.

## Alternatives considered

- *In `runTx`, beside the UTF-8 check*: the baseapp cannot see
  `IsGenesisReplay`, which lives in auth, and a height check would refuse
  replayed history that left the key out.
- *Covering the key with the signature*: a verifier given a stripped copy would
  rebuild the payload with the account's stored key, so the copy would still
  verify. Signing over the key does not force it into the tx bytes. Covering it
  would also change every sign payload, and with it Ledger signing, every
  wallet and the signatures in genesis files.
- *Refusing a key once the account has one*: that breaks gnokey and gnoclient,
  which send the key on every signature.

## Consequences

- **Consensus-breaking, shipped with the UTF-8 and multisig checks.** A
  transaction that was valid is now refused. It ships in the same halt-height
  binary switch, so blocks below the halt height keep their old binary.
- **No live client leaves the key out.** All 233,680 gnoland-1 transactions up
  to height 1,182,449 carry a key in every signature, including 429 session
  signatures.
  - The same pre-release rescan is a sanity check, not a gate.
  - A node re-executing history on the new binary would need this check gated
    in the ante handler beside it, for example through `AnteOptions`, not in
    `runTx`. On gnoland-1 no such gate would let one binary re-execute its
    blocks: the multisig check shipped with this one refuses one historical
    transaction and is deliberately not gated (see *A height gate* in
    `pr6504_multisig_exact_threshold.md`). So a node re-executing gnoland-1's
    blocks needs the earlier binary up to the halt height.
- **Clients must send the key.** A client that leaves it out must now send it.
  gnokey, gnoclient and the browser wallets seen on chain already do.
  `testutils.NewSessionTestSignature` left it out and now sends it.
- **Scope.** This closes the stripped-key rewrite only. The invalid-UTF-8 swap
  is closed by `pr6504_tx_utf8_strings.md`, and dropping spare multisig
  signatures by `pr6504_multisig_exact_threshold.md`, within the limits each
  states. gnolang/gno#6406's canonical check refuses byte-level rewrites of the
  transaction (non-minimal varints, explicit zero fields), and the multisig ADR
  decodes the multisignature nested in `Signature.Signature` canonically too.
