# ADR: secp256k1 validator signing keys (exploratory)

## Status

SUPERSEDED by [`pr5949_remove_secp256k1_validators.md`](./pr5949_remove_secp256k1_validators.md),
which landed on master and removed `secp256k1.PubKeySecp256k1` from
`validatorPubKeyTypeURLs` and `DefaultValidatorParams()`. A chain now rejects a
secp256k1 validator at consensus-params validation, so the plumbing explored
here has no reachable configuration.

The deciding constraint turned out not to be the one this exploration set out to
measure. #5949's reason is the **IBC GNO light client**: its signature
verification and validator-set encoding are built around ed25519, so a chain
admitting a secp256k1 validator would produce headers that light client cannot
verify. That is a correctness/bridgeability argument, and it does not depend on
how fast secp256k1 verification is. Open question 3 below has since been
answered anyway, and the answer was "affordable" — which is exactly why it is
worth recording that performance was not what closed this.

Kept open as a record of what the change would have required, and because two
pieces of it stand on their own regardless of the validator key policy — see
Consequences.

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
slower than ed25519, and every full node verifies every commit signature each
block. This was written as an unquantified worry; it is now measured end-to-end
at the `VerifyCommit` level rather than per-signature — see Open questions 3.
The short version: a few secp256k1 validators cost almost nothing, an
all-secp256k1 valset costs ~3.4x, and neither is prohibitive.

**Light-client / IBC.** This section originally read "most Cosmos-style light
clients handle secp256k1; confirm any custom relayers do too" — a note to
follow up on. It is the whole story. The Gno IBC light client's signature
verification and validator-set encoding are built around ed25519, so a
secp256k1 validator makes the chain unbridgeable, and #5949 removed the key
type from the validator allow-list for that reason. Recorded here as a lesson:
the item flagged as a "confirm later" was the blocker, while the item flagged
as needing measurement before any decision turned out not to matter.

**What stands on its own.** Two pieces of this branch are independent of the
validator key-type policy and could be salvaged if wanted:

- `tm2/pkg/bft/types/validator_set_mixed_bench_test.go` — a `VerifyCommit`
  benchmark parameterized on valset size, which is useful for sizing the
  ed25519-only valset too.
- `secrets_init`'s explicit key-type flag and validation, which currently
  accepts one value; the plumbing makes a future key type a config change
  rather than a code change.

The scheme-agnostic `MakeSecretConnectionAny` and the privval wiring are not in
that category — they exist only to carry a non-ed25519 validator key.

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
2. ~~Should `DefaultValidatorParams` continue to advertise secp256k1 as
   accepted on mainnet given there is no production tooling to onboard a
   secp256k1 validator today, or should it be narrowed to ed25519 until
   tooling lands?~~ **Answered by #5949: narrowed to ed25519**, and not for the
   tooling reason — see Status.
3. ~~Performance impact at realistic valset sizes — needs measurement
   before any policy decision.~~ **Answered**, by @D4ryl00 in
   [#5687](https://github.com/gnolang/gno/pull/5687#issuecomment-3110915073).
   Their `BenchmarkMixedScheme_VerifyCommit` is now in-tree
   (`tm2/pkg/bft/types/validator_set_mixed_bench_test.go`, cherry-picked from
   their fork) so the numbers are reproducible here rather than only on a
   branch:

   ```
   go test ./tm2/pkg/bft/types -run '^$' \
     -bench '^BenchmarkMixedScheme_VerifyCommit$' -benchmem -benchtime=3s
   ```

   Full-node `VerifyCommit`, Apple M1 (re-run at this merge, `-benchtime=200x`;
   matches their reported figures within noise):

   | valset | all ed25519 | 1 secp | ~1/3 secp | all secp |
   |---:|---:|---:|---:|---:|
   | 10  | 0.54 ms | 0.55 ms | 0.77 ms |  1.54 ms |
   | 50  | 2.31 ms | 2.40 ms | 3.95 ms |  7.64 ms |
   | 100 | 4.43 ms | 4.54 ms | 7.99 ms | 15.34 ms |
   | 180 | 8.11 ms | 8.18 ms | 14.56 ms | 27.62 ms |

   Cost scales linearly in valset size and in the secp share. A handful of
   secp256k1 validators is close to free (180 validators, one secp: +0.9%); an
   all-secp256k1 valset is ~3.4x an all-ed25519 one.

   Validator-side signing latency, measured separately by @D4ryl00 with a
   scenario-23 single-validator run (each validator only signs its own
   messages, so this does not grow with valset size), avg ms:

   | phase | ed25519 | secp256k1 |
   |---|---:|---:|
   | proposal  | 0.167 | 0.242 |
   | prevote   | 0.099 | 0.189 |
   | precommit | 0.078 | 0.165 |

   Conclusion: sub-millisecond on the signing side, and tens of milliseconds
   at the worst end of verification — real but affordable. Performance was
   therefore never the blocker.
4. Does anything in the IBC light client's ed25519 assumption (the #5949
   reason) have a bounded fix, or is ed25519-only a permanent property of the
   validator set? Not investigated here.

## References

- `tm2/pkg/bft/types/params.go` — `DefaultValidatorParams.PubKeyTypeURLs`
- `tm2/pkg/p2p/conn/secret_connection.go` — legacy ed25519-typed STS
- `tm2/pkg/bft/privval/signer/local/key.go` — file-key generation
- `gno.land/pkg/gnoland/app.go:1028-1052` — valset proposal
  whole-reject when pubkey type is disallowed
