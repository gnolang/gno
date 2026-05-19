# ADR: secp256k1 validator signing keys (exploratory)

## Status

EXPLORATORY. Not for merge as-is. Open for discussion.

## Context

Today, the gno.land validator stack is implicitly ed25519-only end-to-end:

- `gnoland secrets init` only generates ed25519 priv_validator_keys
  (`tm2/pkg/bft/privval/signer/local/key.go:GenerateFileKey`).
- The privval remote-signer transport (`MakeSecretConnection` in
  `tm2/pkg/p2p/conn/secret_connection.go`) hard-types both ends to
  `ed25519.PrivKeyEd25519` and the authorized-keys allowlist in
  `RemoteSignerClientConfig.authorizedKeys()` rejects anything else.
- Per-validator vote-signature verification (`val.PubKey.VerifyBytes` in
  `tm2/pkg/bft/types/validator_set.go`) is already polymorphic via
  `crypto.PubKey`, and `DefaultValidatorParams.PubKeyTypeURLs` lists both
  ed25519 and secp256k1 as accepted schemes
  (`tm2/pkg/bft/types/params.go`). So the *consensus* path supports
  secp256k1; only the surrounding tooling does not.

Operators have expressed interest in driving validator signing from HSMs
that expose secp256k1, while other operators want to keep using
ed25519-backed remote signers (e.g. tmkms-style). In principle a single
chain can host a mixed-scheme validator set: each precommit signature is
verified against its own validator's pubkey, addresses do not collide
(ed25519 uses `SHA256-trunc20(pk)`, secp256k1 uses
`RIPEMD160(SHA256(pk))`, both 20-byte but disjoint).

This PR explores what it would take, end-to-end, to make that mix work.

## Decision

Three changes, all behind explicit opt-ins for backwards compatibility:

1. **`secrets init -key-type=ed25519|secp256k1`** (default `ed25519`).
   Adds `GenerateFileKeyOfType` / `GeneratePersistedFileKeyOfType` in
   `tm2/pkg/bft/privval/signer/local/key.go`; `GenerateFileKey` keeps
   defaulting to ed25519 so existing callers do not change behaviour.

2. **`MakeSecretConnectionAny`** in `tm2/pkg/p2p/conn`, a parallel entry
   point to `MakeSecretConnection` that accepts `crypto.PrivKey` and
   exchanges an amino-polymorphic `authSigMessageAny`. The original
   ed25519-typed `MakeSecretConnection` is untouched so the p2p layer
   (and its node-ID semantics) stays exactly as it is today. The new
   variant is wired into the privval TCP transport via
   `TCPConnConfig.SchemeAgnostic`.

3. **Polymorphic `authorized_keys` in `RemoteSignerClientConfig`**. The
   legacy `authorizedKeys()` still returns `[]ed25519.PubKeyEd25519` for
   the legacy path; a new `authorizedKeysAny()` returns
   `[]crypto.PubKey`. Operators on the scheme-agnostic path may list
   either scheme in their TOML.

## Alternatives considered

- **Make `MakeSecretConnection` itself scheme-agnostic.** Cleaner
  end-state but breaks wire-compat with existing peers immediately, and
  ripples into p2p node-ID semantics (node IDs are derived from the
  SecretConnection-authenticated key). Out of scope for an exploration
  focussed on privval.

- **HSM-direct signer (PKCS#11) bypassing SecretConnection entirely.**
  Workable but parallel infrastructure; doesn't address operators who
  want to keep the remote-signer pattern (e.g. for off-host signing) but
  with secp256k1 keys.

- **Allow only the validator signing key to be secp256k1, keep the
  privval transport ed25519-only.** Simplest scope: the validator's
  HSM-backed secp256k1 key is the on-chain identity; the host runs a
  separate ed25519 keypair for the privval-channel mutual auth. This
  *does* work today after just change (1); no protocol changes needed.
  Documented here because we believe operators may prefer it on
  pragmatic grounds and we want reviewers to push back on whether (2)
  and (3) are worth the complexity.

## Consequences

**Compatibility.** The legacy path is preserved; existing operators see
no behaviour change. The scheme-agnostic path is opt-in via
`TCPConnConfig.SchemeAgnostic` (not yet exposed in TOML config — see
"Known gaps"). The on-wire auth message format diverges between the two
paths, by design.

**Verification cost.** secp256k1 signature verification is meaningfully
slower than ed25519 (an order of magnitude in the no-cgo build; less
with cgo). Every full node verifies every commit signature each block.
Operators considering a mixed valset should benchmark using
`tm2/pkg/crypto/secp256k1/bench_test.go` against their actual valset
size before adoption.

**Light-client / IBC.** If gno.land adopts IBC, downstream light clients
will need to handle commits containing secp256k1 signatures. Most
Cosmos-style light clients do; confirm any custom relayers do too.

**Address collision risk.** Negligible. Both schemes derive 20-byte
addresses via disjoint hash chains; collisions would require a
preimage-class break.

## Known gaps (deferred)

- `TCPConnConfig.SchemeAgnostic` is not yet plumbed to the TOML config
  layer. To opt in today an operator must instantiate the privval
  client/server programmatically.
- `RemoteSignerClient.clientPrivKey` and `serverPrivKey` are still
  concretely `ed25519.PrivKeyEd25519`-typed in the struct fields; the
  legacy code path uses them. A scheme-agnostic server would need
  `WithServerPrivKeyAny(crypto.PrivKey)` (and the client counterpart).
- `MakeSecretConnection` (legacy) is unchanged. A future PR may
  consolidate both paths once the wire-format and node-ID implications
  are agreed.
- p2p Node IDs remain ed25519-only.

## Open questions

1. Is the consolidation of `MakeSecretConnection` and
   `MakeSecretConnectionAny` valuable, or is keeping them parallel
   better (since p2p and privval have very different compat constraints)?
2. Should `DefaultValidatorParams` continue to advertise secp256k1 as
   accepted on mainnet given there is no production tooling to onboard a
   secp256k1 validator today, or should it be narrowed to ed25519 until
   tooling lands?
3. Performance impact at realistic valset sizes — needs measurement
   before any policy decision.

## References

- `tm2/pkg/bft/types/params.go` — `DefaultValidatorParams.PubKeyTypeURLs`
- `tm2/pkg/p2p/conn/secret_connection.go` — legacy ed25519-typed STS
- `tm2/pkg/bft/privval/signer/local/key.go` — file-key generation
- `gno.land/pkg/gnoland/app.go:1028-1052` — valset proposal
  whole-reject when pubkey type is disallowed
