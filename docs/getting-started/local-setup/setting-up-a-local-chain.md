---
id: setting-up-a-local-chain
---

# Setting up a Local Chain

## Overview

In this tutorial, you will learn how to start a local Gno node (and chain!).
Additionally, you will see the different options you can use to make your Gno instance unique.

## Prerequisites

- [`gnoland` installed](local-setup.md#3-installing-other-gno-tools).

## Starting a node with a default configuration

You can start a Gno blockchain node with the default configuration by navigating to the `gno.land` sub-folder and
running the following command:

```bash
gnoland start
```

The command will trigger a chain initialization process (if you haven't run the node before), and start the Gno node,
which is ready to accept transactions and interact with other Gno nodes.

![gnoland start](../../assets/getting-started/local-setup/setting-up-a-local-chain/gnoland-start.gif)

To view the command defaults, simply run the `help` command:

```bash
gnoland start --help
```

Let's break down the most important default settings:

- `chainid` - the ID of the Gno chain. This is used for Gno clients, and distinguishing the chain from other Gno
  chains (ex. through IBC)
- `config` - the custom node configuration file
  for more details on utilizing this file
- `genesis-balances-file` - the initial premine balances file, which contains initial native currency allocations for
  the chain. By default, the genesis balances file is located in `gno.land/genesis/genesis_balances.txt`, this is also the
  reason why we need to navigate to the `gno.land` sub-folder to run the command with default settings
- `data-dir` - the working directory for the node configuration and node data (state DB)

:::info Resetting the chain

As mentioned, the working directory for the node is located in `data-dir`. To reset the chain, you need
to delete this directory and start the node up again. If you are using the default node configuration, you can run
`make fclean` from the `gno.land` sub-folder to delete the `tempdir` working directory.

:::

## Changing the chain ID

:::info Changing the Gno chain ID has several implications

- It affects how the Gno node communicates with other Gno nodes / chains
- Gno clients that communicate through JSON-RPC need to match this value

It's important to configure your node properly before launching it in a distributed network.
Keep in mind that changes may not be applicable once connected.

:::

To change the Gno chain ID, run the following command:

```bash
gnoland start --chainid NewChainID
```

We can verify the chain ID has been changed, by fetching the status of the node and seeing the
associated chain ID. By default, the node exposes the JSON-RPC API on `http://127.0.0.1:26657`:

```bash
curl -H "Content-type: application/json" -d '{
    "jsonrpc": "2.0",
    "method": "status",
    "params": [],
    "id": 1
}' 'http://127.0.0.1:26657'
```

We should get a response similar to this:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "node_info": {
      "version_set": [
        // ...
      ],
      "net_address": "g10g9r37g9xa54a6clttzmhk2gmdkzsntzty0cvr@0.0.0.0:26656",
      "network": "NewChainID"
      // ...
    }
  }
}
```

:::danger Chain ID can be set only once

Since the chain ID information is something bound to a chain, you can
only change it once upon chain initialization, and further attempts to change it will
have no effect.

:::

## Changing the node configuration

You can specify a node configuration file using the `--config` flag.

```bash
gnoland start --config config.toml
```

## Changing the premine list

You do not need to use the `gno.land/genesis/genesis_balances.txt` file as the source of truth for initial network
funds.

To specify a custom balance sheet for a fresh local chain, you can use the `-genesis-balances-file`:

```bash
gnoland start -genesis-balances-file custom-balances.txt
```

Make sure the balances file follows the following format:

```text
<address>=<balance>ugnot
```

Following this pattern, potential entries into the genesis balances file would look like:

```text
g1qpymzwx4l4cy6cerdyajp9ksvjsf20rk5y9rtt=10000000000ugnot
g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq=10000000000ugnot
```
