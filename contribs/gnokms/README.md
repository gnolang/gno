# gnokms

`gnokms` is a simple Key Management System (KMS) designed to securely manage signing keys for [gnoland](../../gno.land/cmd/gnoland) (TM2) validator nodes. Rather than storing a key in plain text on disk, a validator can run a `gnokms` server in a separate process or on a separate machine, delegating the responsibility of securely storing the signing key and using it for remote signing.

`gnokms` also aims to provide several backends, including a local [gnokey](../../gno.land/cmd/gnokey) instance, a remote HSM, or a cloud-based KMS service.

Both TCP and Unix domain socket connections are supported for communication between the validator and the `gnokms` server. TCP connections are encrypted and can be mutually authenticated using Ed25519 keypairs and an authorized keys whitelist on both sides.

## Status: alpha — not yet recommended for high-stake production

`gnokms` is functional for development, testnets, and low-stake validators, but is **not yet at feature parity with the Cosmos validator ecosystem's de-facto KMS, [tmkms](https://github.com/iqlusioninc/tmkms)**. Until the gaps below are closed, validators with material stake should plan to use tmkms once a TM2 compatibility shim lands; in the meantime, run `gnokms` only with explicit awareness of these limitations.

### Implemented

- Remote signing over TCP or Unix domain socket
- Encrypted, mutually-authenticated TCP via Tendermint SecretConnection (Ed25519 + ChaCha20-Poly1305)
- Bidirectional Ed25519 allowlist (`auth_keys.json`)
- `gnokey` keyring backend (encrypted at rest with bcrypt-derived password)
- **Signer-side double-sign protection**: persistent `(height, round, step)` state file gates every Sign request, refusing regression and refusing same-HRS-different-bytes. Equivalent to tmkms's `consensus.json` gate.

### Known limitations (vs. tmkms)

1. **No reverse-dial mode.** `gnokms` listens; the validator dials in. tmkms convention is the inverse — the signer dials the validator, leaving the signer host with no inbound network surface. The current direction means the most security-sensitive component is a network listener, which is the wrong posture for production. See `priv_validator_listen_addr` in upstream Tendermint.
2. **Defaults fail open.** With no `auth_keys.json`, a TCP listener accepts any peer that completes the SecretConnection handshake (warning logged, but not enforced). Unix-socket listeners bypass the allowlist entirely. Production deployments must run `gnokms auth generate` and configure the allowlist explicitly. The README's `unix:///tmp/gnokms.sock` example is dev-only — `/tmp` is world-traversable on most systems; pick a private directory with `0700` perms.
3. **No HSM or hardware-key backend.** Only the `gnokey` keyring backend is available today. Issues are open for [Ledger](https://github.com/gnolang/gno/pull/5088), [YubiHSM 2](https://github.com/gnolang/gno/issues/3236), and cloud KMS backends, but none are merged.
4. **No threshold signing.** No equivalent of [Horcrux](https://github.com/strangelove-ventures/horcrux). `gnocrux` is [proposed](https://github.com/gnolang/hackerspace/issues/76) but not yet implemented.
5. **No panic recovery in connection handler.** A single malformed message from an authenticated peer can panic and kill the gnokms process. Acceptable on a hardened single-validator host with restart supervision; suboptimal otherwise.
6. **Single-connection serial accept.** No concurrent connections; no read deadline. A connected peer that stalls mid-request blocks all signing until the connection is dropped. Slowloris-shaped DoS.
7. **Operator hardening not documented.** No systemd template (`NoNewPrivileges`, `ProtectSystem=strict`, `LimitCORE=0`, `MemoryDenyWriteExecute`, etc.); no `kernel.yama.ptrace_scope=2` recommendation; no incident-response guidance. The keyring password lives in process memory for the entire runtime — a root-on-host or core-dump-capable attacker recovers the key. Treat the gnokms host as crown jewels.

### Recommended posture

- **Dev/test/local chains:** `gnokms` is fine. Use it.
- **Public testnets, low-stake validators:** `gnokms` is acceptable if you (a) configure the mutual-auth allowlist, (b) avoid `/tmp` socket paths, (c) harden the host with systemd settings, and (d) treat `priv_validator_state.json` and `signer_state.json` as crown-jewels (do not restore from snapshots without a known-good HRS marker).
- **Mainnet, institutional stake:** prefer running upstream tmkms once the TM2 compatibility shim lands. The compatibility work is planned and bounded but not yet shipped. Until then, accept the limitations above explicitly or wait. Running `gnokms` with material stake without the limitations addressed is an unforced operational risk.

### Tracking

The roadmap to feature parity is tracked across:
- Reverse-dial topology
- YubiHSM 2 / Ledger / cloud-KMS-signing backends
- Threshold signing (`gnocrux`)
- tmkms-compat path: lets unmodified upstream tmkms speak to gno.land

### Flowchart

```text
                                                            ┌─────────────────────┐
                                                            │                     │
                                              ┌─────────────┤ Cloud-based backend │
                                              │             │                     │
                                              │             └─────────────────────┘
                                              │
                                              │
                                              │
┌───────────────────┐                 ┌───────┴───────┐     ┌─────────────────────┐
│                   │                 │               │     │                     │
│ gnoland validator │◄─── UDS/TCP ───►│ gnokms server ├─────┤    gnokey backend   │
│                   │                 │               │     │                     │
└───────────────────┘                 └───────┬───────┘     └─────────────────────┘
                                              │
                                              │
                                              │
                                              │             ┌─────────────────────┐
                                              │             │                     │
                                              └─────────────┤     HSM backend     │
                                                            │                     │
                                                            └─────────────────────┘
```

## Getting Started

### Using `gnokms` with a gnoland validator

**Note:** The only supported backend for now is [gnokey](../../gno.land/cmd/gnokey), so the following instructions will use it.

1. Generate a signing key using [gnokey](../../gno.land/cmd/gnokey) if you do not already have one.
2. Start a `gnokms` server with the [gnokey](../../gno.land/cmd/gnokey) backend using:

```shell
$ gnokms gnokey '<key_name>' -listener '<listen_address>'
# <key_name> is the name of the key generated in step 1.
# <listen_address> is the address on which the server should listen (e.g., 'tcp://127.0.0.1:26659' or 'unix:///tmp/gnokms.sock').
```

3. Set the `gnokms` server address in the gnoland validator config using:

```shell
$ gnoland config set consensus.priv_validator.remote_signer.server_address '<gnokms_server_address>'
Updated configuration saved at gnoland-data/config/config.toml
```

### Genesis

When launching the `gnokms` server (e.g. step 2 from the previous section), it should display JSON containing validator information that is compatible with a genesis file. Example:

```shell
$ gnokms gnokey test1
Enter password to decrypt the key
2025-02-26T17:30:25.340+0100 INFO  Validator info:
Genesis format:
{
  "address": "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
  "pub_key": {
    "@type": "/tm.PubKeySecp256k1",
    "value": "A+FhNtsXHjLfSJk1lB8FbiL4mGPjc50Kt81J7EKDnJ2y"
  },
  "power": "10",
  "name": "gnokms_remote_signer"
}
Bech32 format:
  pub_key: gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0skzdkmzu0r9h6gny6eg8c9dc303xrrudee6z4he4y7cs5rnjwmyf40yaj
  address: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
```

If you need to manually edit a genesis file to include these info, you can copy and paste the `Genesis format` part of the output. If it better suits your needs, you can also use the `Bech32 format` part in conjunction with the [gnogenesis](../gnogenesis) command:

```shell
$ gnogenesis validator add \
--address g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 \
--pub-key gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0skzdkmzu0r9h6gny6eg8c9dc303xrrudee6z4he4y7cs5rnjwmyf40yaj \
--name gnokms_remote_signer \
--power 10 \
--genesis-path <path_to_genesis_file>
```

### Mutual TCP Authentication

In the case of a TCP connection, the connection is encrypted. It can also be mutually authenticated to ensure an additional level of security (recommended outside of a testing or development context).

1. Generate a random keypair and an empty whitelist on the server side using:

```shell
$ gnokms auth generate
Generated auth keys file at path: "/home/gnome/.config/gnokms/auth_keys.json"
```

2. Note the public key of the `gnokms` server displayed by the command:

```
$ gnokms auth identity
Server public key: "<gnokms_public_key>"
```

3. On the client side, add the `gnokms` server’s key to the validator’s whitelist using:

```shell
$ gnoland config set consensus.priv_validator.remote_signer.tcp_authorized_keys '<gnokms_public_key>'
Updated configuration saved at gnoland-data/config/config.toml
```

4. Note the validator’s public key displayed by the command:

```shell
$ gnoland secrets get node_id.pub_key
"<validator_public_key>"
```

5. On the server side, add the node’s key to the `gnokms` server whitelist using:

```shell
$ gnokms auth authorized add '<validator_public_key>'
Public key "<validator_public_key>" added to the authorized keys list.
```

### Signer-side double-sign protection

`gnokms` persists a state file recording the `(height, round, step)` of the last Sign request. Every subsequent Sign is gated against this state: regressions are refused and same-HRS requests are refused unless the SignBytes are byte-identical (idempotent retransmit). This prevents the most common slashing scenarios — operator restoring from a VM snapshot, dual-validator misconfiguration, validator-side `priv_validator_state.json` corruption — from cascading into a double-sign at the signer.

Default location: `<user-config-dir>/gnokms/signer_state.json` (e.g., `~/.config/gnokms/signer_state.json` on Linux). Override with `-state-file <path>`.

**Operator notes:**
- Treat this file as crown-jewel. Restoring an older copy can cause a double-sign because gnokms will think it can sign at a lower HRS than the validator has already committed.
- Do NOT co-locate with the validator's `priv_validator_state.json` — the two state files protect against different scenarios and should be on different volumes (ideally different hosts).
- Place on persistent storage. If the file lives on tmpfs and gnokms restarts, the gate resets to H=0 and the next request silently bypasses the protection. Reject ephemeral mounts at deployment time.
- If the file is missing at startup, gnokms creates an empty one. This is correct for a fresh validator key; it is **wrong** if you're recovering an existing validator and have lost the state — in that case you must wait long enough for the chain to advance past the validator's last-committed height before signing again.
