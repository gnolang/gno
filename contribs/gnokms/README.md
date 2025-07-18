# gnokms

`gnokms` is a simple Key Management System (KMS) designed to securely manage signing keys for [gnoland](../../gno.land/cmd/gnoland) (TM2) validator nodes. Rather than storing a key in plain text on disk, a validator can run a `gnokms` server in a separate process or on a separate machine, delegating the responsibility of securely storing the signing key and using it for remote signing.

`gnokms` also aims to provide several backends, including a local [gnokey](../../gno.land/cmd/gnokey) instance, a remote HSM, or a cloud-based KMS service.

Both TCP and Unix domain socket connections are supported for communication between the validator and the `gnokms` server. TCP connections are encrypted and can be mutually authenticated using Ed25519 keypairs and an authorized keys whitelist on both sides.

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
