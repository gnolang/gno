# ADR-003: tmkms compatibility for the privval socket

## Status

Accepted

## Context

Validator-key custody is the highest-stakes operational concern for any
staking deployment. tm2 currently has two privval options:

- **Local file** — `priv_validator_key.json` on disk next to the validator.
  Adequate for dev/test, unsuitable for any material stake.
- **gnokms (`contribs/gnokms`)** — an in-tree, tm2-native remote signer.
  Functional but alpha-tier; `contribs/gnokms/README.md` enumerates the
  gaps blocking institutional use:
  1. no reverse-dial mode (gnokms listens, the validator dials in — wrong
     posture for production where the most security-sensitive component
     should sit behind no inbound network surface);
  2. fail-open defaults (TCP listener accepts any peer that completes the
     SecretConnection handshake when no allowlist is configured; UDS
     listeners bypass the allowlist entirely);
  3. no HSM / Ledger / Fortanix / cloud-KMS backends;
  4. no threshold signing (no Horcrux equivalent);
  5. (since fixed in `0130a1bd0`) HRS double-sign authority lived in
     gnoland's `priv_validator_state.json` rather than in the signer.

Closing each of these inside gnokms is multi-month work per item, an
estimated 6+ months for full feature parity with the Cosmos validator
ecosystem's de-facto KMS:

- **[tmkms](https://github.com/iqlusioninc/tmkms)** — Iqlusion's Rust
  KMS, in production across most major Cosmos chains for ~7 years, with
  YubiHSM 2 / Ledger / Fortanix DSM / softsign / cloud-KMS backends,
  one external security audit (NCC Group), and a documented Horcrux
  integration for threshold signing.

Cosmos validator operators already run tmkms — familiar runbooks,
familiar systemd templates, familiar incident response.

The strategically cheaper alternative is to make **unmodified upstream
tmkms** able to sign for a tm2 validator. End state: the operator adds
one `[[validator]]` block to their `tmkms.toml`. No code changes on the
tmkms side; all the work is in tm2. Effort: ~3–5 weeks of tm2-side work
versus ~6 months of gnokms hardening — and tmkms's existing audits and
hardware backends carry over for free.

The tmkms wire protocol diverges from gnokms in three ways that need to
be addressed in tm2:

1. **Dial direction is inverted.** tmkms dials the validator (the
   validator listens). The validator's process holds the inbound port;
   the signer host has no inbound network surface.
2. **The signer canonicalizes.** In tm2 today the validator builds
   `CanonicalVote` sign-bytes locally and sends opaque bytes to gnokms.
   Upstream Tendermint protocol sends a structured `Vote` and the KMS
   canonicalizes using its own `chain_id` config — preventing the
   validator from tricking the KMS into signing for a different chain.
3. **HRS state lives at the signer.** tmkms's `consensus.json` is the
   double-sign gate. Restoring a stale copy enables double-signing.

(2) and (3) are *better* security posture than gnokms's current model.

## Decision

Add a tmkms compatibility shim to tm2 in a new package
`tm2/pkg/bft/privval/upstream/` that speaks upstream Tendermint v0.34's
privval socket protocol. Operators enable it via a new config block
`[consensus.priv_validator.tmkms_listener]`, mutually exclusive with the
existing gnokms `[remote_signer]` block. Both gnokms's and tm2's chain
p2p wire formats are otherwise unchanged.

Three guiding principles:

1. **Mirror cometbft's `privval/` package structure.** Type names, file
   layout, method signatures, default timeouts. Goal: a reviewer who
   knows cometbft's privval can read tm2's tmkms-compat code and
   recognize every part. Reduces audit surface to "what differs from
   upstream" rather than "is the whole thing correct."
2. **Use protoc-generated types directly at the wire boundary.** The
   `upstreampb` sibling package is generated from upstream Tendermint
   v0.34's `privval/types.proto`, `canonical.proto`, `crypto.proto`,
   `types.proto`. Wire I/O uses these types via
   `google.golang.org/protobuf/proto`; conversion to/from tm2's
   chain-internal types happens at the application edge.
3. **Pin to a single dialect.** Today `protocol_version = "v0.34"` is
   the only accepted value. tmkms supports both v0.34 (deprecated but
   accepted) and v0.38 (default). Pinning forces operators to set the
   same value on both sides — silent canonical-bytes drift is the
   feared failure mode and we'd rather refuse to start.

### Implementation history (all shipped on `feat/jae/gnokms-hrs`)

| Commit | Description |
|---|---|
| `910b5ad71` | amino `binary:"varint"` tag — opt-in plain protobuf varint instead of zigzag |
| `1d3210de2` | Re-tag CanonicalProposal/PartSetHeader for upstream byte-compat |
| `86bbae90d` | Upstream-shaped Vote/Proposal types |
| `413ed9016` | Three-layer wire-compat test suite vs upstream |
| `e34a4f5cd` | Route privval-protocol wire I/O through protoc-generated upstreampb |
| `7e070ab1a` | SignerListenerEndpoint + base endpoint (port of cometbft v0.39.1) |
| `9e507a157` | SignerClient + RetrySignerClient + socket listener |
| `aaff0fc87` | Wire upstream listener mode into NewPrivValidatorFromConfig |
| `2e84471f9` | Security hardening for tmkms-compat path (six fixes) |
| `edb0de5bf` | Pin protocol version to v0.34 |
| `1c48ce60e` | SecretConnection byte-compat verification |
| `c401b23fb` | tmkms-compat SecretConnection (port of cometbft v0.34 STS, Merlin-bound) |
| `ea10ad550` | tmkms binary integration test |
| `68282b930` | Wire fixes surfaced by the real-tmkms test |
| `6a7674c9a` | Cross-link contribs/gnokms/README.md to the operator doc |

## Architecture

```
 ┌────────────────────────┐         ┌──────────────────────────┐
 │ gnoland (validator)    │         │ tmkms (signer host)      │
 │                        │         │                          │
 │  ┌──────────────────┐  │         │  ┌────────────────────┐  │
 │  │ TCPListener      │ ◄┼─dials──┼─ │ tmkms validator    │  │
 │  │ (allowlist)      │  │         │  │ block (chain_id,   │  │
 │  │                  │  │         │  │ secret_key)        │  │
 │  │ SignerListener-  │  │         │  └────────────────────┘  │
 │  │ Endpoint         │  │         │                          │
 │  │   ↕              │  │         │     consensus.json       │
 │  │ SignerClient     │  │         │   (HRS double-sign gate) │
 │  │   ↕              │  │         │                          │
 │  │ RetrySignerClient│  │         │                          │
 │  └──────────────────┘  │         │                          │
 │   ↕                    │         └──────────────────────────┘
 │  consensus engine      │
 └────────────────────────┘
```

The validator listens on `tcp://…:26659` (or a unix socket). tmkms dials
in. Mutual SecretConnection auth (X25519 + ChaCha20-Poly1305 + ed25519
identities, Merlin-bound STS handshake), allowlist check on the remote
ed25519 pubkey, then `PubKeyRequest` / `SignVoteRequest` /
`SignProposalRequest` exchanges thereafter.

### Two SecretConnection implementations

The chain-p2p `tm2/pkg/p2p/conn/secret_connection.go` is **unchanged** —
amino-shaped `authSigMessage`, no Merlin transcript. Touching it is a
chain wire-format break.

The tmkms-listener path uses a **separate** copy of cometbft v0.34.34's
secret_connection.go in `tm2/pkg/bft/privval/upstream/secret_connection.go`,
adapted to tm2 types but byte-identical to upstream:

- Merlin-transcript-bound challenge derivation (`gtank/merlin`)
- `AuthSigMessage` protobuf with `PublicKey` oneof
- HKDF-SHA256 derivation, ChaCha20-Poly1305 framing, X25519 DH

Verification tests (`secret_connection_compat_test.go`) pin both
encodings: the chain-p2p one as a known-divergence canary (ensuring an
accidental change can't sneak in without a chain-wide review), the
listener one as a positive match against upstream byte-for-byte.

### Security model

| Concern | Defense |
|---|---|
| Slashing via HRS state regression | Authority moved to tmkms's `consensus.json` (operator runbook treats it as crown-jewel; never restore from snapshot). |
| Signer swap mid-session (TCP reset + race a different tmkms instance) | Connection-generation counter on the endpoint. SignerClient re-fetches and verifies the signer's pubkey on every reconnect; mismatch refuses to sign. |
| Compromised signer that rewrites `(Height, Round, BlockID)` in the response | Echo verification: signer may only fill in `Signature` (and canonicalize `Timestamp`); other fields must echo what we sent. Signature is rejected on mismatch. |
| Fail-open allowlist | `ValidateBasic` refuses an empty `allowed_kms_pubkeys` when the listener mode is enabled. |
| Local-user privilege escalation via UDS socket | `os.Chmod(0600)` on the socket immediately after `bind(2)`. |
| Listener leak on Init failure | `endpoint.Stop()` (not `sc.Close()`) drains goroutines and releases the bound port. |
| Protocol-dialect drift | Pinned to `v0.34`; `ValidateBasic` refuses any other value. |

### Wire-format requirements (learned from real-tmkms integration test)

The integration test surfaced two non-obvious requirements pinned in
`translator_pb.go`:

- **`Timestamp` must always be present.** `tendermint-rs` raises
  `MissingTimestamp` on nil; zero-valued `time.Time` serializes as the
  protobuf year-0001 timestamp, which both sides canonicalize
  identically.
- **`PartSetHeader` must always be present.** `tendermint-rs`'s
  `BlockId::TryFrom` rejects `part_set_header is None`; empty values
  (Total=0, Hash=nil) are accepted as `Some(default)` (`tag len=0`).

A residual amino-vs-proto canonical divergence exists for default-valued
embedded messages (amino skips them; upstream proto encodes
`tag len=0`). Harmless in real consensus where every block has parts.
Pinned by the wire-compat tests; will fail loud if ever exercised.

## Consequences

### Positive

- Validators can run unmodified upstream tmkms with all its hardware
  backends — YubiHSM 2, Ledger, Fortanix DSM, softsign, cloud KMS.
- Threshold signing via Horcrux comes for free (Horcrux speaks the same
  upstream privval protocol).
- Better security posture than gnokms by construction: HRS authority at
  the signer, signer canonicalizes with its own `chain_id`, signer host
  has no inbound network surface.
- Chain wire format is unchanged. tm2's chain-p2p SecretConnection is
  untouched. Adopting tmkms-compat is opt-in per validator.
- Reduces gnokms's role to dev/test; operators with material stake have
  a clear, documented path.

### Negative

- Two SecretConnection implementations to maintain. The
  upstream-compat one is a near-verbatim port of cometbft v0.34.34, so
  the maintenance cost is mostly tracking upstream changes (mitigated by
  pinning v0.34 and only updating when forced).
- New module dep: `github.com/gtank/merlin v0.1.1` (pure-Go Merlin
  transcript) and its transitive `github.com/mimoo/StrobeGo`. Apache-2.0
  licensed, small surface.
- `protocol_version = "v0.34"` is deprecated upstream — tmkms 0.15
  warns "will be a hard error in next release." Future tmkms releases
  may drop v0.34 and force a v0.38 migration; this is a tracked
  follow-up, not a blocker.
- One amino-vs-proto canonical divergence remains for default-valued
  embedded messages. Documented; won't regress without the canary
  failing.

### Neutral

- gnokms remains in-tree as a dev/test convenience. Its README is
  cross-linked to `docs/validators/tmkms.md` for operators with stake.
- Operators must explicitly set `protocol_version = "v0.34"` on both
  sides. Mismatches fail loud at startup.

## Verification

- Three-layer wire-compat test suite in
  `tm2/pkg/bft/privval/upstream/upstreamwire_test.go` (protowire walks,
  hand-rolled spec encoders, stdlib protobuf round-trip via generated
  upstreampb).
- SecretConnection internal contracts pinned in
  `tm2/pkg/p2p/conn/secret_connection_compat_test.go` (HKDF info string,
  output split, nonce layout, frame layout).
- SecretConnection wire-format match in
  `tm2/pkg/bft/privval/upstream/secret_connection_compat_test.go`
  (positive match for the listener-path encoding; canary for the
  chain-p2p divergence).
- End-to-end test against a real tmkms 0.15.0 binary in
  `tm2/pkg/bft/privval/upstream/tmkms_integration_test.go` (gated by
  `//go:build tmkms_integration`). CI builds tmkms from source via
  `cargo install` (`.github/workflows/ci-tmkms-integration.yml`) and
  exercises PubKey + SignVote (h=1, h=2) + SignProposal (h=3).

## References

- Operator guide: [`docs/validators/tmkms.md`](../../docs/validators/tmkms.md)
- gnokms README (dev/test path):
  [`contribs/gnokms/README.md`](../../contribs/gnokms/README.md)
- Upstream cometbft v0.34.34 source files this implementation mirrors:
  - `cometbft/privval/signer_listener_endpoint.go`
  - `cometbft/privval/signer_endpoint.go`
  - `cometbft/privval/signer_client.go`
  - `cometbft/privval/retry_signer_client.go`
  - `cometbft/privval/socket_listeners.go`
  - `cometbft/p2p/conn/secret_connection.go`
- Upstream protobuf schema: tendermint/proto/tendermint/{privval,types,
  crypto}/types.proto at v0.34
- tmkms: <https://github.com/iqlusioninc/tmkms>
- Horcrux: <https://github.com/strangelove-ventures/horcrux>
