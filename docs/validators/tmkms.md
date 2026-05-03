# Signing with tmkms

This guide is for operators who want to keep their validator's consensus
key out of the gnoland process and have it held by [tmkms] (or another
signer that speaks the upstream Tendermint privval protocol, like
[Horcrux]) instead.

[tmkms]: https://github.com/iqlusioninc/tmkms
[Horcrux]: https://github.com/strangelove-ventures/horcrux

## When to use this mode

Gnoland supports three privval setups. Pick one — they're mutually
exclusive.

| Mode | Key holder | Connection direction | Use when |
|---|---|---|---|
| Local file | gnoland | n/a | Dev, testnets, single-host |
| `gnokms` (remote signer) | gnokms | gnoland *dials* gnokms | You run our native signer; gnokms is on a network the validator can reach |
| `tmkms` listener | tmkms | tmkms *dials* gnoland | You already run tmkms (e.g. for other Cosmos chains), or you want the signer host to have **no inbound network surface** |

The defining property of tmkms mode is the dial direction: the
**validator listens** and the **signer dials in**. The signer machine
needs no inbound firewall rule, which is why most production Cosmos
operators run tmkms this way.

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
 │  └──────────────────┘  │         │                          │
 │   ↕                    │         └──────────────────────────┘
 │  consensus engine      │
 └────────────────────────┘
```

The validator's `SignerListenerEndpoint` accepts the inbound dial,
performs a SecretConnection mutual-auth handshake (X25519 + ChaCha20-
Poly1305 + ed25519 identities), checks the remote pubkey against an
allowlist, and then issues `PubKeyRequest` / `SignVoteRequest` /
`SignProposalRequest` over the link.

In this mode **tmkms — not gnoland — owns HRS state**: tmkms's
`consensus.json` is the authoritative double-sign gate. Gnoland's local
`priv_validator_state.json` is not used as a gate for the tmkms path;
it would be redundant at best and a misconfiguration footgun at worst.

## Configuration

In `config.toml`, under `[consensus.priv_validator.tmkms_listener]`:

```toml
[consensus.priv_validator.tmkms_listener]
# Address gnoland listens on for tmkms to dial in.
# tcp://<bind>:<port>  or  unix:///path/to/privval.sock
listen_addr = "tcp://0.0.0.0:26659"

# Hex-encoded ed25519 pubkeys of every signer instance allowed to
# connect. Each entry is 64 hex chars, optionally prefixed "ed25519:".
# Empty list is REJECTED at startup — see Security below.
allowed_kms_pubkeys = [
  "ed25519:7c4f6a1c...0000000000000000000000000000000000000000000000",
]

# Chain ID sent to the signer in PubKeyRequest / SignVoteRequest /
# SignProposalRequest. tmkms verifies its own configured chain_id
# matches and refuses to sign if it doesn't. Required.
chain_id = "test4"

# Upstream Tendermint privval dialect to speak. Today only "v0.34" is
# supported; gnoland refuses to start with any other value. Must
# match tmkms.toml's [[validator]].protocol_version. See "Protocol
# version pin" below.
protocol_version = "v0.34"

# Read/write deadline applied to the held signer connection.
timeout_read_write = "5s"

# Max time gnoland blocks at startup waiting for tmkms to dial in.
wait_for_connection_timeout = "60s"

# Per-Sign retry budget on transient errors. 0 means retry forever
# (matches cometbft's RetrySignerClient convention).
retries = 5

# Sleep between retry attempts.
retry_timeout = "1s"
```

Any non-empty `listen_addr` enables this mode. To disable, leave it
empty (the default).

`[consensus.priv_validator.remote_signer]` (gnokms) and
`[consensus.priv_validator.tmkms_listener]` are mutually exclusive —
gnoland refuses to start with both configured.

## Getting the allowlist pubkey

Each signer entity has a SecretConnection identity key (separate from
the consensus key tmkms is signing with). For tmkms, the path is set
in `tmkms.toml`:

```toml
[[providers.softsign]]
chain_ids = ["test4"]
key_format = { type = "base64" }
path = "/etc/tmkms/secrets/test4_consensus.key"

[[validator]]
chain_id = "test4"
addr = "tcp://<gnoland-bind>:26659"
secret_key = "/etc/tmkms/secrets/kms-identity.key"
protocol_version = "v0.34"
reconnect = true
```

Get the hex-encoded pubkey from `secret_key` and put it in
`allowed_kms_pubkeys`. tmkms exposes the matching command — see
[tmkms's docs][tmkms-pubkey] — typically:

```
$ tmkms init --pubkey-only /etc/tmkms/secrets/kms-identity.key
```

For Horcrux running multiple cosigners, list one entry per cosigner.

[tmkms-pubkey]: https://github.com/iqlusioninc/tmkms/blob/main/README.md

## TCP vs UDS

| | TCP | UDS |
|---|---|---|
| Connection | Network | Local filesystem socket |
| Encryption | SecretConnection (X25519 + ChaCha20-Poly1305) | None — kernel-level isolation |
| Mutual auth | ed25519 allowlist | Filesystem perms (gnoland chmods `0600`) |
| Use when | Signer is on a different host | Signer is on the same host |

For UDS, gnoland sets the socket to mode `0600` after `bind(2)` so a
local non-root user can't reach the SecretConn handshake stage at all.
You should still keep the parent directory's perms tight (e.g. only
the gnoland service user has search bit on it).

## Protocol version pin

Both sides of the privval socket must agree on the dialect. The
upstream Tendermint privval protocol changed shape between v0.34
(used by the Cosmos Hub for years) and v0.38+ (added new fields that
alter canonical sign-bytes). gnoland's `upstreampb` types are wired
to **v0.34** and only that value is accepted in
`tmkms_listener.protocol_version`.

In `tmkms.toml`, set `protocol_version = "v0.34"` on the matching
`[[validator]]` block. A future gnoland release adding v0.38 support
will accept new values; until then anything else is rejected at
startup with a clear error.

This is deliberate: silently misencoding a vote because the dialect
drifted would produce a signature tmkms thinks is valid but the chain
rejects (or vice versa) — a hard-to-debug consensus failure mode.
We'd rather fail loud at startup.

## Security notes

- **Empty allowlist is refused at startup.** A misconfigured firewall
  plus an attacker who can mint an ed25519 keypair would otherwise be
  enough to substitute the signer over TCP. This is enforced in
  `ValidateBasic`; you can't accidentally ship without an allowlist.
- **Identity is re-verified on every reconnect.** If the signer's
  pubkey changes between connections (e.g. a different tmkms instance
  raced into the listener slot), the next sign call refuses with
  `signer pubkey changed across reconnect`.
- **Signer-echoed fields are rebuilt locally.** Only `Signature` (and
  the canonicalized `Timestamp`) come from the signer's response —
  `Height`, `Round`, `BlockID` etc. are checked against what we sent
  and the response is rejected on mismatch. A compromised signer
  cannot redirect a vote to a different block.
- **HRS authority lives in tmkms.** Treat `consensus.json` (or the
  Horcrux equivalent) as the canonical double-sign gate. Never restore
  a stale copy of it.

## SecretConnection: tmkms-compat vs chain p2p

gnoland uses two distinct SecretConnection implementations:

- **chain p2p** (`tm2/pkg/p2p/conn/secret_connection.go`) — pre-Merlin
  STS handshake, amino-encoded `AuthSigMessage`. Internal to gnoland
  validators talking to each other; unchanged.
- **tmkms listener** (`tm2/pkg/bft/privval/upstream/secret_connection.go`)
  — direct port of cometbft v0.34's Merlin-bound STS handshake with
  protobuf `AuthSigMessage` (`PublicKey` oneof + signature). Used
  only on the listener path the signer dials in to.

Phase 6 byte-compat verification confirms the listener-path
implementation is wire-identical to upstream Tendermint v0.34
(see `secret_connection_compat_test.go` —
`TestUpstreamSecretConnection_AuthSigMessage_MatchesUpstream` and
`TestUpstreamSecretConnection_SelfHandshake`). The chain-p2p
divergence is pinned by `TestSecretConnectionWire_AuthSigMessage_KnownDivergence`
so an accidental "fix" to the chain path can't sneak in without a
chain-wide review (changing chain p2p bytes is a hard fork).

## Wire-format requirements (learned from Phase 7)

The integration test against a real tmkms binary surfaced two
non-obvious wire requirements; they're encoded in
`translator_pb.go::VoteToProto / ProposalToProto / blockIDToProto`:

- **Timestamp must always be present.** A SignVoteRequest with a
  missing (nil) Timestamp triggers tendermint-rs's `MissingTimestamp`
  error and tmkms drops the connection. Zero-valued time
  (`time.Time{}`) serializes as the protobuf year-0001 timestamp;
  both sides canonicalize that case identically.
- **PartSetHeader must always be present.** A BlockID with a missing
  (nil) part_set_header field triggers tendermint-rs's
  `InvalidPartSetHeader: part_set_header is None`. Empty values
  (Total=0, Hash=nil) are accepted as long as the field is encoded
  (`tag len=0`).

There is one remaining tm2-internal divergence: amino's
canonicalization (used by `Vote.SignBytes`) omits a default-valued
embedded message, while upstream proto encodes it as `tag len=0`.
The gap surfaces only when a vote carries an empty PartSetHeader,
which never happens in real consensus traffic — every block has
parts. Real-network signatures from a tmkms-using validator verify
correctly on tm2 nodes because both sides hit the populated-header
case. The integration test populates PartSetHeader explicitly to
exercise that path.

## Verifying compat with a real tmkms binary

The repo ships a build-tagged Go integration test that orchestrates
a real tmkms binary against the gnoland listener path. CI runs it
via `.github/workflows/ci-tmkms-integration.yml` (only when the
upstream privval code or the workflow itself changes — it builds
tmkms from source and isn't cheap).

To run locally:

```sh
# Install tmkms once (Rust toolchain required).
cargo install tmkms --version 0.15.0 --features softsign --locked

# From the repo root:
go test -tags=tmkms_integration -count=1 -v \
    ./tm2/pkg/bft/privval/upstream/...
```

The test (`TestTmkmsIntegration_FullSigningFlow`) generates fresh
ed25519 keys, writes a tmkms.toml that pins
`protocol_version = "v0.34"`, spawns `tmkms start` against an
ephemeral listener, and asserts:

- `PubKey()` reported by tmkms matches the consensus key in
  softsign,
- `SignVote` at heights 1 and 2 round-trip with valid signatures
  (verified against the consensus pubkey),
- `SignProposal` at height 3 round-trips with a valid signature.

If the test fails, check tmkms's stdout/stderr in the test logs —
the most common failure modes are protocol-version skew (set
`protocol_version = "v0.34"` on both sides) and chain_id
mismatches.

## Operational checklist

- [ ] One signer mode enabled in `config.toml` (gnokms OR tmkms, not both).
- [ ] `chain_id` in `tmkms_listener` matches both the genesis chain ID
      and the `chain_id` in tmkms's `[[validator]]` block.
- [ ] `protocol_version = "v0.34"` set on both sides.
- [ ] `allowed_kms_pubkeys` lists the hex pubkey of every tmkms
      identity (or every Horcrux cosigner) expected to connect.
- [ ] If using TCP, `listen_addr` is reachable from the signer host
      and not exposed publicly (firewall the inbound port to the
      signer's source IP).
- [ ] If using UDS, the parent directory is on local disk (not a
      shared filesystem) and only the gnoland service user can
      traverse it.
- [ ] `consensus.json` (tmkms HRS state) is durable and backed up;
      restoring an older copy enables double-signing.
- [ ] tmkms is started and reachable BEFORE gnoland — gnoland's
      `Init` blocks for `wait_for_connection_timeout` waiting for the
      dial-in.
