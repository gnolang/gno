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

## Known issue: SecretConnection auth-sig wire incompat

Phase 6 byte-compat verification surfaced a **2-byte divergence**
between tm2's SecretConnection AuthSigMessage encoding and upstream
Tendermint v0.34's. The ephemeral-pubkey exchange and the HKDF /
nonce / frame layers all match upstream byte-for-byte, but the
post-DH authentication message diverges:

- Upstream wraps the ed25519 pubkey in a `PublicKey` oneof
  (`AuthSigMessage{pub_key: PublicKey{ed25519: ...}, sig: ...}`).
- tm2 emits the ed25519 pubkey directly under field 1 of the
  unexported `authSigMessage{Key, Sig}` struct via amino, with no
  oneof wrapper.

**Effect**: gnoland in tmkms-listener mode completes the ephemeral
key exchange, derives matching keys, and then deadlocks at the
auth-sig step — tmkms's protobuf decoder can't parse our amino-shaped
bytes (and vice versa). End-to-end signing against a real tmkms
binary will not work until this is addressed.

**Why we haven't fixed it yet**: the encoding is shared with tm2's
node-to-node p2p SecretConnection. Changing it is a chain-wide wire
break (post-mainnet hard fork). The right fix is an upstream-compat
handshake variant used only on the tmkms-listener path; that work is
tracked separately and is not in scope for the verification phase.

The divergence is pinned in
`tm2/pkg/bft/privval/upstream/secret_connection_compat_test.go` —
`TestSecretConnectionWire_AuthSigMessage_KnownDivergence` will fail
loud the moment anyone fixes the encoding (intentionally or not),
forcing a chain-wide review before the change lands.

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
