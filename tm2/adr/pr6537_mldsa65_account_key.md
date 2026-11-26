# ML-DSA-65 as an account key type

## Context

Every account key tm2 accepts, ed25519 and secp256k1, rests on the elliptic
curve discrete logarithm problem. A break of ECDLP, classical or quantum,
exposes the private key of every account whose public key is on chain, which
is every account that has signed once. There is no fallback key type today.

## Decision

Add `tm2/pkg/crypto/mldsa65`, a `crypto.PubKey` and `crypto.PrivKey` pair for
ML-DSA-65 (FIPS 204, category 3), built on `github.com/cloudflare/circl`
v1.6.1. That version is the newest that adds no transitive bump to
`golang.org/x/*`; v1.6.5 would have raised five of them.

- The private key is the 32-byte FIPS 204 seed, expanded on use. The public
  key is the 1952-byte packed key, a fixed array like the ed25519 type, so
  amino and genproto2 treat it the same way. Signatures are 3309 bytes.
- The address is `tmhash.SumTruncated` of the raw public key, as for the other
  keys.
- Signing is deterministic with an empty context string. Hedged signing can be
  adopted later without touching verification.
- The amino package path is new (`github.com/gnolang/gno/tm2/pkg/crypto/mldsa65`)
  and is registered in `misc/genproto2`, whose output is committed.
- `auth.Params` gains `SigVerifyCostMLDSA65`, default 1400, and the ante
  handler charges it per ML-DSA-65 signature. The figure scales the ed25519
  cost by measured `VerifyBytes` time (ed25519 33 us, secp256k1 113 us,
  ML-DSA-65 76 us, so 590 x 76 / 33 is about 1360). The size of signature and
  key, 5261 bytes, is charged separately through `TxSizeCostPerByte`.

### The new param on chains that already run

The field is appended last in `Params`: amino numbers fields by position, so
inserting it beside the other costs would have renumbered every later field
of params already stored.

A genesis file or a stored params struct written before the field existed
reads back with a zero there. Left alone, that zero fails `Validate` at
`InitGenesis`, panics every later auth param proposal (`WillSetParam`
revalidates the whole struct), and would make ML-DSA-65 verification free.
`Params.withDefaults` fills the default in `GetParams`, `ValidateGenesis` and
`InitGenesis`. No block before this change can hold an ML-DSA-65 key, the
type being unregistered, so the default changes no replayed result.

## Alternatives considered

- Hybrid-only (a composite ed25519 plus ML-DSA key): stronger during the
  transition, but a new composite type and verification rule. A 2-of-2
  `PubKeyMultisigThreshold` of the two already expresses it.
- SLH-DSA (SPHINCS+): more conservative assumptions, but signatures of 7 to
  50 KB.
- Falcon: smaller signatures, but floating point signing and no final FIPS yet.

## Out of scope

Validator keys (`ValidatorParams.PubKeyTypes`), the p2p handshake, a message
that rotates an existing account to a new key, hybrid multisig guidance, and
tx and block size limits sized for 5 KB signatures. Whether
`multisig.ValidateStructure`'s size bounds admit ML-DSA-65 constituents is not
checked here.

## Consequences

A fresh account can hold its funds behind a lattice key. Each such signature
costs about 52 600 gas in tx bytes at the default `TxSizeCostPerByte` of 10,
plus 1400 for verification: roughly 25 times an ed25519 signature.
