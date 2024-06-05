---
id: running-a-validator
---

# Start a New Gno Chain and a Validator

## 1. Initialize the configurations (required)

```bash
gnoland config init
```

## 2. Initialize the secrets (required)

```bash
gnoland secrets init
```

## 3. Set a moniker (optional)

```bash
gnoland config set moniker node01
```

## 4. Set the rpc connection address (required for connecting with other nodes)

```bash
gnoland config set rpc.laddr "tcp://0.0.0.0:26657"

# similar behavior for cosmos validator
# gaiad tx staking create-validator `--node string (default:tcp://localhost:26657)`
```

## 5. Set the validator private key (required)

:::tip

The key file path is relative by default.

:::

```bash
gnoland config set priv_validator_key_file secrets/priv_validator_key.json
```

## 6. Set the validator state (required)

:::tip

The key file path is relative by default.

:::

```bash
gnoland config set priv_validator_state_file secrets/priv_validator_state.json
```

## 7. Set the node key (required)

:::tip

The key file path is relative by default.

:::

```bash
gnoland config set node_key_file secrets/node_key.json
```

## 8. Set the genesis file (required)

```bash
gnoland genesis generate
```

## 9. Add a validator (required)

```bash
# check the secrets file generated in step (2)
$ gnoland secrets get
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
$ gnoland genesis validator add -address g1lnha5yem9dmj0yfzysfqsnvrm6j2ywshq83qdf -name node01 -power 10 -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpleysamt23ar025757uepld60xztnw7ujc3gwtjuy4pwv6z9sh4g284h3q
```

## 10. Start the chain

```bash
gnoland start
```

:::warning

Slashing policies for downtimes, double signing or any manipulation using the validator membership will be introduced in the near future.

Setting up the Sentry Node Architecture minimizes the risks of DDoS targetting the validator node. Configuring the `unconditional_peer_ids` is essential. Refer to [this issue](https://github.com/gnolang/gno/issues/1923) for further information. :::

# Connect to an Existing Gno Chain

## 1. Initialize the configurations (required)

```bash
gnoland config init
```

## 2. Initialize the secrets (required)

```bash
gnoland secrets init
```

## 3. Set a moniker (optional)

```bash
gnoland config set moniker node02
```

## 4. Set the rpc connection address (required for connecting with other nodes)

```bash
gnoland config set rpc.laddr "tcp://0.0.0.0:26657"
```

## 5. Set the validator private key (required)

:::tip

The key file path is relative by default.

:::

```bash
gnoland config set priv_validator_key_file secrets/priv_validator_key.json
```

## 6. Set the validator state (required)

:::tip

The key file path is relative by default.

:::

```bash
gnoland config set priv_validator_state_file secrets/priv_validator_state.json
```

## 7. Set the node key (required)

:::tip

The key file path is relative by default.

:::

```bash
gnoland config set node_key_file secrets/node_key.json
```

## 8. Obtain the genesis file of the chain to connect to

:::info

The genesis file will be [easily downloadable from GitHub](https://github.com/gnolang/gno/issues/1836#issuecomment-2049428623) in the future.

For now, obtain the file by

1. Sharing via scp or ftp
2. Getting from `{chain_rpc:26657}/genesis` (might result in time-out error due to large file size)

:::

```bash
## TODO: Add link to download the file from GitHub
```

## 9. Add the new validator to existing chain

::: info

This step is currently unavailable. It will be supported in the future after complete implementation of validator set injection with the `r/sys/val` realm.

:::

```bash
## TODO: Add a new validator
```

## 10. Confirm the validator information of the first node.

```bash
# Node ID
$ gnoland secrets get NodeKey

[Node P2P Info]
Node ID:  g19d8x6tcr2eyup9e2zwp9ydprm98l76gp66tmd6

# The Public IP of the Node
$ curl ifconfig.me/ip
1.2.3.4 # USE YOUR OWN IP
```

## 11. Configure the persistent_peers list

Configure a list of nodes that your validators will always retain a connection with.

```bash
$ gnoland config set p2p.persistent_peers "g19d8x6tcr2eyup9e2zwp9ydprm98l76gp66tmd61.2.3.4:26656"
```

## 12. Configure the seeds

Configure the list of seed nodes. Seed nodes provide information about other nodes for the validator to connect with the chain, enabling a fast and stable initial connection.

:::info

This is an option to configure the node set as the Seed Mode. However, the option to activate the Seed Mode from the node is currently missing.

:::

```bash
gnoland config set p2p.seeds "g19d8x6tcr2eyup9e2zwp9ydprm98l76gp66tmd6@1.2.3.4:26656"
```

## 13. Start the second node

```bash
gnoland chain start
```

## Results After Starting the Chain and Two Nodes

![1st_node](../assets/validator/running-a-validator/1st_node.png) The 1st node at height 12263.

![2nd_node](../assets/validator/running-a-validator/2nd_node.png) The 2nd node at height 12263 (synced with the 1st node)
