---
id: using-gnoland
---

## Using `gnoland`

## Overview

In this tutorial, you will learn how to use the `gnoland` binary. This guide covers
various available subcommands and flags to help you configure, manage, and spin up
a local gno.land node.

## Prerequisites

- **Git**
- **`make` (for running Makefiles)**
- **Go 1.21+**
- **Go Environment Setup**:
    - Make sure `$GOPATH` is well-defined, and `$GOPATH/bin` is added to your `$PATH` variable.

## Installation

To install the `gnoland` binary, clone the Gno monorepo:

```bash
git clone https://github.com/gnolang/gno.git
```

After cloning the repo, go into the `gno.land/` folder, and use the existing
Makefile to install the `gnoland` binary:

```bash
cd gno.land
make install.gnoland
```

To verify that you've installed the binary properly and that you are able to use
it, run the `gnoland` command:

```bash
gnoland --help
```

If everything was successful, you should get the following output:

```bash
USAGE
  <subcommand> [flags] [<arg>...]

starts the gnoland blockchain node.

SUBCOMMANDS
  start    run the full node
  secrets  gno secrets manipulation suite
  config   gno config manipulation suite
  genesis  gno genesis manipulation suite
```

If you do not wish to install the binary globally, you can build it with the
following command from the `gno.land/` folder:

```bash
make build.gnoland
```

And finally, run it with `./build gnoland`.

Let's dive deeper into each of the subcommands and see how we can use them.

## Managing Node Secrets

The `gnoland secrets` command suite helps your manage three node secrets:
1. validator private key - `ValidatorPrivateKey`
2. node p2p key - `NodeKey`
3. validator's last sign state - `ValidatorState`

The suite allows you to initialize and verify these secrets, and also read them
via the CLI. Below are the available subcommands and their usage.

### Initializing Secrets

If you wish to configure and manage your node manually, you can choose to initialize
the validator private key, the node p2p key, and the validator's last sign state
using the `secrets init` subcommand.

```bash
gnoland secrets init [flags] [<key>]
```

Running the `init` subcommand without any flags will generate a default data
directory for node secrets. Let's see it in action:

```bash
> gnoland secrets init
Validator private key saved at gnoland-data/secrets/priv_validator_key.json
Validator last sign state saved at gnoland-data/secrets/priv_validator_state.json
Node key saved at gnoland-data/secrets/node_key.json
```

#### Flags
- `-data-dir gnoland-data/secrets` - the directory where node secrets will be saved to
- `-force` - if some secrets already exist, they will be overwritten with new ones

### Getting Secrets

To access the public values of your gno.land node secrets through the CLI, you
can use the `secrets get` subcommand:

```bash
gnoland secrets get [flags] [<key>]
```

Getting all values can be done with the following command:

```bash
❯ gnoland secrets get

[Node P2P Info]

Node ID:  g16942vugc4g98j6gqfm94tnd4qwwgk8ee4nuntd
[Validator Key Info]

Address:     g1ukvflgxqehyrzwm0ulh7a9sn6hm7505dc0mwrj
Public Key:  gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zquv55ka63cpwqx56j22fesrevem54dqdlfy2s008l4ev2cvhnnyk9tdf0t
[Last Validator Sign State Info]

Height:  0
Round:   0
Step:    0
```

#### Flags
- `-data-dir gnoland-data/secrets` - the directory from which to read the secrets

### Verifying Secrets

You can verify the existence and integrity of all node secrets using the 
`secrets verify` subcommand. The command can be run for any specific secret,
or for all of them at the same time.

```bash
gnoland secrets verify [flags] [<key>]
```

In case a specific node secret is missing or corrupt, the command will let you know.
For example, if the `node_key.json` file was missing from the secrets directory,
checking for it would result in the following output:

```bash
❯ gnoland secrets verify NodeKey 
unable to read node p2p key, unable to read data, open gnoland-data/secrets/node_key.json: no such file or directory
```

#### Flags
- `-data-dir gnoland-data/secrets` - the directory from which to read the secrets

## Node Configuration

The `gnoland config` command suite helps manage the gno.land node configuration.

### Initializing Configuration

To initialize the gno.land node configuration, you can use the following command:

```bash
gnoland config init [flags]
```

This command will initialize a configuration directory and file in the default
path. The node is highly configurable. Some of the main configuration options
which are available are listed below:
- Database which is used for storing node data
- Fast synchronization
- Local paths for node secrets
- Consensus configuration
- Mempool configuration
- Networking (P2P) configuration
- RPC configuration
- Telemetry

To view the fully detailed configuration options, take a look at the generated
`config.toml` file.

#### Flags
- `-data-dir gnoland-data/config` - the directory where the configuration file will be saved in

### Getting Node Configuration

The `config get` subcommand allows you to read specific configuration values
from the `config.toml` file.

```bash
gnoland config get [flags] <key>
```

For example, fetching the RPC listener address of the
node can be done the following way, considering that the listener address key is
`laddr`, found under the `rpc` category:

```bash
gnoland config get rpc.laddr
```

A full configuration category, such as `consensus`, can also be read via this command:

```bash
gnoland config get rpc
```

For the RPC category, the default output for this subcommand will be a struct of
all configuration fields inside it, similar to the following:

```bash
{ tcp://127.0.0.1:26657 [*] [HEAD GET POST OPTIONS] [Origin Accept Content-Type X-Requested-With X-Server-Time]  900 false 900 10s 1000000 1048576  }
```

#### Flags
- `-data-dir gnoland-data/config` - the directory from which to read the configuration

### Setting Node Configuration

The `config set` subcommand allows you to set specific configuration values
in the `config.toml` file via the CLI, instead of manually editing the file.

```bash
gnoland config set [flags] <key> <value>
```

For example, setting the node name to `my-node` using the `moniker` field can be 
done with the following command:

```bash
gnoland config set moniker my-node
```

#### Flags
- `-data-dir gnoland-data/config` - the directory in which to modify the configuration

## Configuring the Node Genesis

The node genesis file contains the initial parameters with which to set up the node.
It contains the following fields:
- Genesis time
- Chain ID
- Consensus parameters
- Genesis validator set
- Genesis balances
- Genesis transaction set

### Generating a Genesis file

To generate a genesis file, you can use the `genesis generate` subcommand:

```bash
gnoland genesis generate
```

This will initialize a default genesis file in the default path. The `genesis.json`
file will have the following data:

```json
{
  "genesis_time": "2024-06-07T13:46:01Z",
  "chain_id": "dev",
  "consensus_params": {
    "Block": {
      "MaxTxBytes": "1000000",
      "MaxDataBytes": "2000000",
      "MaxBlockBytes": "0",
      "MaxGas": "100000000",
      "TimeIotaMS": "100"
    },
    "Validator": {
      "PubKeyTypeURLs": [
        "/tm.PubKeyEd25519"
      ]
    }
  },
  "app_hash": null
}
```










