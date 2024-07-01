---
id: validators-connect-to-and-existing-gno-chain
---

# Connect to an Existing Gno Chain

## Overview

In this tutorial, you will learn how to start a local Gno node and connect to an existing Gno chain (like a testnet).

## Prerequisites

- **Git**
- **`make` (for running Makefiles)**
- **Go 1.22+**
- **Go Environment Setup**: Ensure you have Go set up as outlined in
  the [Go official installation documentation](https://go.dev/doc/install) for your environment

## 1. Initialize the node directory

To initialize a new Gno.land node working directory (configuration and secrets), make sure to
follow [Step 1](./setting-up-a-new-chain.md#1-generate-the-node-directory-secrets--config) from the
chain setup tutorial.

## 2. Obtain the `genesis.json` of the remote chain

The genesis file of target chain is required in order to initialize the local node.

:::info

The genesis file will
be [easily downloadable from GitHub](https://github.com/gnolang/gno/issues/1836#issuecomment-2049428623) in the future.

For now, obtain the file by

1. Sharing via scp or ftp
2. Fetching it from `{chain_rpc:26657}/genesis` (might result in time-out error due to large file sizes)

:::

## 3. Confirm the validator information of the first node.

```bash
gnoland secrets get node_id

{
    "id": "g17h5t86vrztm6vuesx0xsyrg90wplj9mt9nsxng",
    "p2p_address": "g17h5t86vrztm6vuesx0xsyrg90wplj9mt9nsxng@0.0.0.0:26656"
}
```

### Public IP of the Node

You need the IP information about the network interface that you wish to connect from external nodes.

If you wish to only connect from nodes in the same network, using a private IP should suffice.

However, if you wish to connect from all nodes without any specific limitations, use your public IP.

```bash
curl ifconfig.me/ip # GET PUBLIC IP

# 1.2.3.4 # USE YOUR OWN PUBLIC IP
```

## 4. Configure the `persistent_peers` list

We need to configure a list of nodes that your validators will always retain a connection with.
To get the local P2P address of the current node (these values should be obtained from remote peers):

```bash
gnoland secrets get node_id.p2p_address

"g17h5t86vrztm6vuesx0xsyrg90wplj9mt9nsxng@0.0.0.0:26656"
```

We can use this P2P address value to configure the `persistent_peers` configuration value

```bash
gnoland config set p2p.persistent_peers "g19d8x6tcr2eyup9e2zwp9ydprm98l76gp66tmd6@1.2.3.4:26656"
```

## 5. Configure the seeds

We should configure the list of seed nodes. Seed nodes provide information about other nodes for the validator to
connect with the chain, enabling a fast and stable initial connection. These seed nodes are also called _bootnodes_.

:::warning

The option to activate the Seed Mode from the node is currently missing.

:::

```bash
gnoland config set p2p.seeds "g19d8x6tcr2eyup9e2zwp9ydprm98l76gp66tmd6@1.2.3.4:26656"
```

## 6. Start the node

Now that we've set up the local node configuration, and added peering info, we can start the Gno.land node:

```shell
gnoland start \
--genesis ./genesis.json \
--data-dir ./gnoland-data
```

That's it! 🎉

Your new Gno node should be up and running, and syncing block data from the remote chain.
