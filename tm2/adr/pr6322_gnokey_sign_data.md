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
document whose single message is `sign/MsgSignData {signer, data, origin, purpose}`,
with `account_number`, `sequence` and `gas` at `"0"`, an empty fee and an empty
memo, sorted by key. ADR-036's message is `{signer, data}`; `origin` and
`purpose` are added (see 2). The format is specified in
`docs/resources/sign-data.md`. The payload is pinned against signatures from
`@cosmjs/crypto`, Adena's signing stack, so a service handles a proof from
gnokey or Adena the same way.

A node rebuilds a transaction's sign bytes from the decoded transaction, and
amino writes every message there as `{"@type": …}`. A document whose message is
`{"type", "value"}` is no transaction's sign document, so a proof can never be
broadcast. With a zero fee, that `"@type"` is the only difference between a
proof's bytes and a transaction's, so `TestSignDataPayloadIsNoTransactionPayload`
pins it from both sides. A transaction's messages all carry `"@type"`, even one
shaped exactly like the sign-data message. The sign-data message carries none.

### 2. Bind a proof to a service and a use

ADR-036 and Keplr's `signArbitrary` sign `{chain, signer, data}` only. If two
services both hand out a bare random nonce as the challenge, a proof made for
one passes at the other. The message therefore also carries:

- `origin`, the service the proof is for. A wallet takes it from the
  requester, as Adena takes the chain id from the connected network, never
  from the request: for a web page, the page's origin.
- `purpose`, what the proof is for, as the service names it, such as
  `username/Reserve`.

Both are required and returned in `SignedData` for the service to compare. The
field was first named `url`; it is `origin` because it holds an origin and
nothing else, which the name makes hard to get wrong.

`SignDataPayload` refuses both unless they are in canonical form, so that
signing and verifying enforce the same rule:

- **`origin`** is written as a browser serializes a page's origin:
  `scheme://host[:port]`, lowercase, the host an ASCII DNS name (`xn--` for an
  internationalized one), an IPv4 address or a bracketed IPv6 address in its
  shortest form, the port only when not the scheme's default, nothing after,
  at most 256 bytes. Any non-empty text would let one wallet sign
  `https://example.com` and another `https://example.com/app?x=1` for the same
  service, and a verifier comparing strings would refuse a valid proof. A
  wallet uses the browser's own serialization, and gets this form. A custom
  scheme with an authority (`com.example.game://callback`) is accepted, for a
  requester that is not a web page.
- **`purpose`** is 1 to 64 ASCII letters, digits, and `.`, `_`, `/` or `-`.
  Every wallet shows it before signing, so a line break, control or bidi
  character, or a look-alike letter would let a requester reshape the approval
  screen in all of them. Loosening the set later is compatible, tightening it
  is not.

Both rules keep the document ASCII outside `data`, which is base64. They sit inside the message, not at the document's root.
The root's keys are fixed: the Ledger Cosmos app checks them against an
allowlist (see `signDocFee`). Adena already ships the document without them, so
adding them is a format change now. After the Ledger app and more wallets ship
it, the same change would cost more.

### 3. Keep the chain id

ADR-036 leaves `chain_id` empty. Here it is required and signed, as Adena does,
so a proof made on one network is refused on another.

### 4. Verify against rebuilt bytes, and only the canonical document

A verifier gets the document back as a JSON object and may reserialize it in any
key order, spacing or string escaping. `std.VerifySignData` reads its values,
rebuilds the canonical bytes with `SignDataPayload`, and checks the signature
over those. It takes only a secp256k1 key, checks that it is the signer's, and
returns `{ChainID, Origin, Purpose, Signer, Data}`, leaving the comparison with
what it expects to the caller.

Only the JSON rendering may differ: key order, spacing and string escapes. The
document, re-encoded sorted by key and compacted as the payload is, must equal
the rebuilt bytes, so a missing or extra field, a key in another case, a number
for a string, a null for the empty fee or non-canonical base64 in the data is
refused. No object in the document or in the proof
around it may repeat a key either: Go keeps a repeated key's last value and
other parsers its first, so a service reading the document with another parser
would see other values than the ones verified.

### 5. The proof's JSON lives in std

The proof is Adena's `SignArbitrary` answer:
`{signed, signature: {pubKey: {typeUrl, value}, signature}}`.
`std.NewSignDataProof` writes it and `std.VerifySignDataProof` reads it. The
reader accepts only a `/tm.PubKeySecp256k1` key of the right size and then calls
`VerifySignData`. A Go service checks a proof in one call, and the writer and
the reader cannot drift apart.

### 6. Two gnokey commands

- `gnokey sign-data -chainid <id> -origin <origin> -purpose <purpose>
  (-data <text> | -data-path <file>) <key>` prints the proof, or writes it with
  `-output-document`. `-chainid`, `-origin` and `-purpose` are required: none is
  assumed. A file is signed as it is, a final
  newline included. Only secp256k1 keys can sign. A Ledger signs with whichever
  device is plugged in, so the signature is checked against the key's public
  key before the proof is written.
- `gnokey verify-data -chainid <id> -origin <origin> -purpose <purpose>
  (-data <text> | -data-path <file>) -proof-path <file|-> <address or key-name>`
  requires every expectation and
  fails unless the proof matches all of them. The proof holds the public key, so
  an address needs no key in the keybase.

## Alternatives considered

- **A gno-specific document, or a prefixed raw text** (as Ethereum's
  `personal_sign`). Proofs from gnokey and Adena would differ, and services would
  have to handle two formats.
- **An empty chain id, as ADR-036 specifies.** A proof would be good on every
  network.
- **ADR-036's `{signer, data}` alone**, as Adena's first `SignArbitrary` and
  Keplr do. Binding a proof to a service would rest on every service writing
  its name into the challenge, which nothing checks.
- **`origin` and `purpose` as optional fields.** A service that leaves them out is
  back to the case above.
- **Any non-empty text in `origin`, normalized by the verifier.** Every verifier
  would have to normalize the same way, and a URL parser's idea of equal URLs
  differs from one library to the next. Refusing all but one form puts the work
  on the wallet, which already has the browser's serialization.
- **A full URL in `origin`, path included.** A page's path is not who the
  service is: one service signs users in from many pages.
- **Printable ASCII, or any text without control characters, in `purpose`.**
  Either still lets a requester write a sentence into the approval screen. A
  name is what a service needs; the text the user reads goes in `data`.
- **EIP-712-style typed data**, with a domain and a schema per message. It
  binds proofs as well, but every wallet and the Ledger app would need to hash
  and display typed structs. Two flat text fields give the same binding.
- **Accepting any document that decodes to the right values.** It is what the
  first version did. No signer or text could be forged, but the documented
  shape was not enforced, and a repeated key could be read differently by
  another verifier.
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

- A proof is bound to a chain, a service and a use, not to a time or a session.
  Replay protection is the service's job: a challenge carries its own nonce,
  expiry and session, and the service accepts each one once. It keys on the challenge, not on the proof's
  bytes: the envelope around the document is read leniently (its base64 skips
  newlines, its keys match in any case), so one proof has many byte forms.
  `VerifySignDataProof`'s doc says so.
- Verifying costs time and memory in proportion to the proof's size; a service
  bounds what it accepts.
- Multisig and non-secp256k1 keys cannot make a proof. `VerifySignData` takes a
  `secp256k1.PubKeySecp256k1`, not a `crypto.PubKey`, so a service cannot hand
  it one either: a multisig address would otherwise verify.
- The format is shared with Adena. Adena's released `SignArbitrary` signs the
  document without `origin` and `purpose`. Its proofs are refused here until it
  signs the new one. From then on, a change on either side must be made on both,
  and the pinned vectors catch a divergence.
- Both shapes will be in use while wallet releases roll out. They are told
  apart by `origin`, which only the new message carries, and the docs say so:
  a service that has to accept the old one binds it through its challenge text,
  and stops once its users have moved on.
- `std` gains `SignDataPayload`, `VerifySignData`, `SignedData`, `SignDataProof`
  (with `SignDataSignature` and `SignDataPubKey`), `SignDataPubKeyType`,
  `NewSignDataProof` and `VerifySignDataProof`.
