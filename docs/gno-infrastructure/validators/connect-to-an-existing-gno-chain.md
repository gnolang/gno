---
id: validators-connect-to-and-existing-gno-chain
---

# Connect to an Existing Gno Chain

## 1. Initialize the configurations (required)

```bash
gnoland config init -config-path gnoland-data/config/config.toml
```

## 2. Initialize the secrets (required)

```bash
gnoland secrets init -data-dir gnoland-data/secrets
```

:::tip

Set a new moniker to distinguish your new node from the existing one.

```bash
gnoland config set moniker node02 -config-path gnoland-data/config/config.toml
```

:::

## 3. Obtain the genesis file of the chain to connect to

- The genesis file of target chain is required to communicate.

:::info

The genesis file will
be [easily downloadable from GitHub](https://github.com/gnolang/gno/issues/1836#issuecomment-2049428623) in the future.

For now, obtain the file by

1. Sharing via scp or ftp
2. Getting from `{chain_rpc:26657}/genesis` (might result in time-out error due to large file size)

:::

```bash
## TODO: Add link to download the file from GitHub
```

## 4. Confirm the validator information of the first node.

```bash
# Node ID
$ gnoland secrets get NodeKey -data-dir gnoland-data/secrets

[Node P2P Info]
Node ID:  g19d8x6tcr2eyup9e2zwp9ydprm98l76gp66tmd6

# The Public IP of the Node

You need the IP information about the network interface that you wish to connect from external nodes.

If you wish to only connect from nodes in the same network, using a private IP should suffice.

However, if you wish to connect from all nodes without any specific limitations, use your public IP.

$ curl ifconfig.me/ip # GET PUBLIC IP
1.2.3.4 # USE YOUR OWN PUBLIC IP
```

## 5. Configure the persistent_peers list

Configure a list of nodes that your validators will always retain a connection with.

```bash
$ gnoland config set p2p.persistent_peers "g19d8x6tcr2eyup9e2zwp9ydprm98l76gp66tmd6@1.2.3.4:26656" -config-path gnoland-data/config/config.toml
```

## 6. Configure the seeds

Configure the list of seed nodes. Seed nodes provide information about other nodes for the validator to connect with the
chain, enabling a fast and stable initial connection.

:::info

This is an option to configure the node set as the Seed Mode. However, the option to activate the Seed Mode from the
node is currently missing.

:::

```bash
gnoland config set p2p.seeds "g19d8x6tcr2eyup9e2zwp9ydprm98l76gp66tmd6@1.2.3.4:26656" -config-path gnoland-data/config/config.toml
```

## 7. Start the second node

```bash
gnoland start -data-dir ./gnoland-data -genesis ./genesis.json
```
