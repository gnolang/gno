# PR6322: Sign and verify a text to prove control of an address off-chain

## Status

Proposed

## Context

A service that needs to know an address belongs to its visitor (reserving a
name, linking an account, gating a page) hands out a challenge and asks for a
signature over it. In the browser, Adena answers with `SignArbitrary`. A user
whose key lives in gnokey, or in a Go wallet built on tm2, had no way to answer,
and a Go service had nothing to check an answer with.

Signing the challenge's raw bytes is not an option: a key that signs whatever
bytes a service supplies can be made to sign a transaction's sign document.

## Decision

### 1. Sign the document Adena's SignArbitrary signs

`std.SignDataPayload` builds ADR-036's off-chain document: a transaction sign
document whose single message is `sign/MsgSignData {signer, data}`, with
`account_number`, `sequence` and `gas` at `"0"`, an empty fee and an empty memo,
sorted by key. These are the bytes Adena signs, pinned against signatures from
`@cosmjs/crypto` (Adena's signing stack) and in Adena's `sign-arbitrary.spec.ts`,
so a service handles a proof from either the same way.

`sign/MsgSignData` is registered nowhere with amino, so no node can decode the
document as a transaction: a proof can never be broadcast.

### 2. Keep the chain id

ADR-036 leaves `chain_id` empty. Here it is required and signed, as Adena does,
so a proof made on one network is refused on another.

### 3. Verify against rebuilt bytes, with a fixed shape

A verifier gets the document back as a JSON object and may reserialize it in any
key order. `std.VerifySignData` reads its values, refuses any field or value
outside the fixed shape, rebuilds the canonical bytes with `SignDataPayload` and
checks the signature over those. It also checks that the public key is the
signer's. It returns `{ChainID, Signer, Data}` and leaves the comparison with
what the caller expects to the caller.

### 4. The proof's JSON lives in std

The proof is Adena's `SignArbitrary` answer:
`{signed, signature: {pubKey: {typeUrl, value}, signature}}`.
`std.NewSignDataProof` writes it and `std.VerifySignDataProof` reads it. The
reader accepts only a `/tm.PubKeySecp256k1` key of the right size and then calls
`VerifySignData`. A Go service checks a proof in one call, and the writer and
the reader cannot drift apart.

### 5. Two gnokey commands

- `gnokey sign-data -chainid <id> (-data <text> | -data-path <file>) <key>`
  prints the proof, or writes it with `-output-document`. `-chainid` is
  required: the network is never assumed. A file is signed as it is, a final
  newline included. Only secp256k1 keys can sign.
- `gnokey verify-data -chainid <id> (-data <text> | -data-path <file>)
  -proof-path <file|-> <address or key-name>` requires every expectation and
  fails unless the proof matches all of them. The proof holds the public key, so
  an address needs no key in the keybase.

## Alternatives considered

- **A gno-specific document, or a prefixed raw text** (as Ethereum's
  `personal_sign`). Proofs from gnokey and Adena would differ, and services would
  have to handle two formats.
- **An empty chain id, as ADR-036 specifies.** A proof would be good on every
  network.
- **Verifying the bytes as received.** Any reserialization by the service, such
  as reading `signed` as an object and writing it back, would break the
  signature.
- **A `verify-data` that only reports whether the signature is valid.** A
  signature over another chain, an old challenge or another address would read
  as a valid proof to someone checking by hand.
- **Expectation checks inside `std`**, as a `SignedData.Expect` method or folded
  into `VerifySignDataProof`. A service often matches a proof against a set of
  outstanding challenges, or looks the signer up after verifying, so a fixed
  three-value check would not suit it. The CLI, which does know all three, makes
  them.
- **A flag on `gnokey verify`.** That command checks a transaction against a key
  in the keybase; a proof is self-contained and carries its own key.
- **Keeping the proof's JSON in the CLI package.** Every Go verifier would
  re-implement the key type and size checks.

## Consequences

- A proof is not bound to a time or a use. Replay protection is the service's
  job: a challenge carries its own nonce and expiry, and the service accepts
  each one once.
- Multisig and non-secp256k1 keys cannot make a proof.
- The format follows Adena's. A change on either side must be made on both; the
  pinned vectors catch a divergence.
- `std` gains `SignDataPayload`, `VerifySignData`, `SignedData`, `SignDataProof`
  (with `SignDataSignature` and `SignDataPubKey`), `SignDataPubKeyType`,
  `NewSignDataProof` and `VerifySignDataProof`.
