---
id: validators-start-a-new-gno-chain-and-validator
---

# Start a New Gno Chain and a Validator

## 1. Initialize the configurations (required)

- Initializes the Gno node configuration locally with default values, which includes the base and module configurations
- A required step in order for the Gno.land node to function correctly.

```bash
gnoland config init -config-path gnoland-data/config/config.toml
```

## 2. Initialize the secrets (required)

- initializes the validator private key, the node p2p key and the validator's last sign state
- A required step which may otherwise prevent blocks from being produced.

```bash
gnoland secrets init -data-dir gnoland-data/secrets
```

:::tip

A moniker is a human-readable username of your validator node. You may customize your moniker with the following command:

```bash
gnoland config set moniker node01 -config-path gnoland-data/config/config.toml
```

:::

## 3. Set the rpc connection address (required for connecting with other nodes)

- A configuration to connect with the RPC service (port 26657) when an external client (i.e. a dApp like Adena Wallet) communicates with the chain (transaction request, block height check, etc.).

```bash
gnoland config set rpc.laddr "tcp://0.0.0.0:26657" -config-path gnoland-data/config/config.toml

# similar behavior for cosmos validator
# gaiad tx staking create-validator `--node string (default:tcp://localhost:26657)`
```

## 4. Set the validator private key (optional)

- Set path of the validator private key. A default value exists. When using a separate secrets folder, you must set the path to the respective location.

:::tip

The key file path is relative by default.

:::

```bash
gnoland config set priv_validator_key_file secrets/priv_validator_key.json -config-path gnoland-data/config/config.toml
```

:::

info validator private key is one of secrets that centeralized within `<data-dir>/secrets`, it can be replaced or regenerated with `gnoland secrets init ValidatorPrivateKey --force`

:::

## 5. Set the validator state (optional)

- Set path of the validator state. A default value exists. When using a separate secrets folder, you must set the path to the respective location.

:::tip

The key file path is relative by default.

:::

```bash
gnoland config set priv_validator_state_file secrets/priv_validator_state.json -config-path gnoland-data/config/config.toml
```

:::

info validator state is one of secrets that centeralized within `<data-dir>/secrets`, it can be replaced or regenerated with `gnoland secrets init ValidatorState --force`

:::

## 6. Set the node id (optional)

- Set path of the node id. A default path exists. When using a separate secrets folder, you must set the path to the respective location.

:::tip

The key file path is relative by default.

:::

```bash
gnoland config set node_key_file secrets/node_key.json -config-path gnoland-data/config/config.toml
```

:::info

node is is one of secrets that centeralized within `<data-dir>/secrets`, it can be replaced or regenerated with `gnoland secrets init NodeID --force`

:::

## 7. Generate the genesis file (required)

- When the chain starts, the first block will be produced after all of the content inside the genesis file is executed.

```bash
gnoland genesis generate
```

## 8. Add a validator (required)

- Add an initial validator. Blocks will not be produced if the chain is started without an active validator.

```bash
# check the secrets file generated in step (2)
$ gnoland secrets get -data-dir gnoland-data/secrets
[Node P2P Info]
Node ID:  g19d8x6tcr2eyup9e2zwp9ydprm98l76gp66tmd6

[Validator Key Info]
Address:     g1lnha5yem9dmj0yfzysfqsnvrm6j2ywshq83qdf
Public Key:  gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpleysamt23ar025757uepld60xztnw7ujc3gwtjuy4pwv6z9sh4g284h3q

[Last Validator Sign State Info]
Height:  0
Round:   0
Step:    0

# add the validator to the genesis file using the address and the public key in the Validator Key Info section
$ gnoland genesis validator add \
  -address g1lnha5yem9dmj0yfzysfqsnvrm6j2ywshq83qdf \ # address of validator
  -name node01 \ # name of validator
  -power 10 \ # voting power of validator
  -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpleysamt23ar025757uepld60xztnw7ujc3gwtjuy4pwv6z9sh4g284h3q # public key of validator
```

## 9. Start the chain

```bash
gnoland start -data-dir ./gnoland-data -genesis ./genesis.json
```
