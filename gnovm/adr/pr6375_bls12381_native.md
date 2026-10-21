# crypto/bls12381: BLS signature verification native

## Context

An IBC light client of Ethereum (the Altair sync protocol) verifies, for every
update, a BLS12-381 aggregate signature of the sync committee (512 members on
mainnet and Sepolia) over the attested beacon header. Pairing arithmetic in
interpreted Gno would cost orders of magnitude more gas than a block allows.

`gnark-crypto`, already a dependency (for `crypto/bn254` and
`crypto/cometbls`), implements BLS12-381 including RFC 9380 hash-to-curve.

## Decision

Add `crypto/bls12381` with one native, `FastAggregateVerify(pubkeys, msg,
sig []byte) bool`, implementing the IETF draft's FastAggregateVerify for the
proof-of-possession ciphersuite Ethereum uses
(`BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_`): compressed 48-byte G1 public
keys concatenated, a compressed 96-byte G2 signature. Every point is decoded
with its subgroup check, the identity is rejected as a public key and as the
aggregate of the public keys (KeyValidate applies to the key CoreVerify is
handed: keys summing to the identity, `pk` and `-pk`, would otherwise verify
the identity signature over any message), and the check is a single
two-pairing product. The signature is subgroup-checked only, as CoreVerify
specifies: with a non-identity key the identity signature cannot pass.

Two obligations the native cannot check fall on the caller, stated in the
imperative in the package doc: every individual key must have had its possession
proven (the POP ciphersuite is rogue-key forgeable otherwise, and there is no
`PopVerify` here; for `FastAggregateVerifyExcluding` this applies to the
committee's member keys, not the aggregate itself; Ethereum sync committee
keys qualify because each validator's deposit signature over a message bound
to its key was checked with `bls.Verify`, a different operation than the
draft's `PopVerify` with the same assurance), and for
`FastAggregateVerifyExcluding` every excluded key must be a verified member
of the committee behind the aggregate (otherwise `excluded` can be chosen so
that the aggregate minus it is the attacker's own key).

The Gno wrapper rejects wrong lengths before calling the native, so malformed
input is not copied into Go first (see `native_input_bounds.md`).

Gas: a `SizeLenBytes` row on `pubkeys`. Measured on Apple M4 Pro: 0.88ms base
(hash-to-curve plus pairing) and about 40µs per key (decompression and
subgroup check, 0.83µs/byte). The row scales these by 2.3x toward the Xeon
Silver 4114 the other IBC crypto draft rows were measured on, and is marked
draft: the whole table is regenerated on the reference machine before any
consensus-relevant deployment.

A second native, `FastAggregateVerifyExcluding(aggregate, excluded, msg,
sig)`, verifies the same equation for the signers of a committee from the
committee's aggregate key minus the keys of the members that did not sign. An
Ethereum `SyncCommittee` commits to its `aggregate_pubkey` (the sum of its
keys, computed by the state transition) next to the keys, and sync committee
participation is typically above 95%, so a light client decodes a handful of
keys per update instead of ~500. With it the light client never needs the
committee's key list on chain: updates carry only the non-signers' keys with
Merkle branches to the committee's pubkeys root. Same per-key pricing as
`fastAggregateVerify`, on the excluded bytes.

## Alternatives considered

- **EIP-2537-style primitives** (G1/G2 add, MSM, pairing, map-to-curve) with
  the verification written in Gno. More general, but each step crosses the
  native boundary and the Gno side still decompresses and aggregates 512 keys.
- **A native taking pre-decompressed keys**, letting the client cache them per
  period. Cuts the per-key cost, but needs realm storage for 512 uncompressed
  keys per period; subtracting the non-signers from the committed aggregate
  needs neither.

## Consequences

- A 512-key verification costs about 47M gas in the draft table, well under
  the block limit.
- Adding a stdlib changes genesis state (stdlib sources are committed), so the
  pinned app hash in `apphash_crossrealm38_test.go` moves.
