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
- **Go Environment Setup**: Ensure you have Go set up as outlined in the [Go official installation documentation](https://go.dev/doc/install) for your environment

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

## Running a Local Node

Using the `gnoland start` command allows you to run a local gno.land node.
The node can be started in two main ways:
- With `lazy` mode
- With manual configuration

Let's see how we can start the node lazily for the quickest setup.

### Lazy Node Initialization

The simplest way to spin up a local node is to use the following command:

```bash
gnoland start --lazy
```

![gnoland-start-lazy](../assets/getting-started/using-gnoland/gnoland-start-lazy.gif)

This command will generate all necessary files for you, including the `genesis.json`
file and node secrets, which include the validator & node private keys.
By default, the node will start listening on `localhost:26657`.

#### Flags
- `-chainid dev` - the ID of the chain.
- `-data-dir gnoland-data` - the path to the node's data directory.
- `-flag-config-path ...` - the flag config file (optional).
- `-genesis genesis.json` - the path to the `genesis.json`.
- `-genesis-balances-file <path_to_genesis_balances_file>` - initial distribution file.
- `-genesis-max-vm-cycles 100000000` - set maximum allowed VM cycles per operation. Zero means no limit.
- `-genesis-remote localhost:26657` - replacement for '%%REMOTE%%' in genesis.
- `-genesis-txs-file <path_to_genesis_txs_file>` - initial transactions to replay.
- `-gnoroot-dir <path_to_your_repo_dir>` - the root directory of the gno repository.
- `-lazy=false` - flag indicating if lazy init is enabled. Generates the node secrets, configuration, and `genesis.json`.
- `-log-format console` - log format for the gnoland node.
- `-log-level debug` - log level for the gnoland node.
- `-skip-failing-genesis-txs=false` - don't panic when replaying invalid genesis transactions.

### Manual configuration

For manual configuration of the node, two main steps are required:
- Setting up node secrets
- Creating a genesis file

To see how to create and set up these files, check out the sections below. You
can also check out the [Setting up a Local Chain](../gno-infrastructure/setting-up-a-local-chain.md)
guide in the [Gno Infrastructure section](../gno-infrastructure/gno-infrastructure.md).

## Managing Node Secrets

The node secrets can be managed using the `gnoland secrets` command suite.
This command suite helps you control three node secrets:
1. validator private key - `ValidatorPrivateKey`
2. node p2p key - `NodeKey`
3. validator's last sign state - `ValidatorState`

The suite allows you to initialize and verify these secrets, and also read them
via the CLI. Below are the available subcommands and their uses.

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

Let's see how we can manipulate these values.

### Generating a Genesis file

To generate a genesis file, you can use the `genesis generate` subcommand:

```bash
gnoland genesis generate [flags]
```

Without any flags, this command will initialize a default genesis file in the
default path. The `genesis.json` file will contain the following data:

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

Below are the flags that we can use to customize the initial `genesis.json` file.

#### Flags

- `-block-max-data-bytes` - sets the max size of the block data, in bytes
- `-block-max-gas` - sets the max gas limit for the block
- `-block-max-tx-bytes` - sets the max size of the block transaction
- `-block-time-iota` - sets the block time iota (in ms)
- `-chain-id` - sets the ID of the chain
- `-genesis-time` - sets the genesis creation time
- `-output-path` - sets the output path for the genesis.json file

### Manage the Genesis Validator Set

You can manage the genesis validator set via the CLI by using the 
`genesis validator add` and `genesis validator remove` subcommands.

To add a new validator to genesis, the following parameters need to be configured:
- `-address` - the gno bech32 address of the validator
- `-name` - the name of the validator (must be unique)
- `-power` - the voting power of the validator (must be > 0)
- `-pub-key` - the bech32 string representation of the validator's public key

An example command for adding a validator to the genesis set might look like this:

```bash
gnoland genesis validator add \
-address g1c56yp4k38dl2637zrwc7zqrzt95gtzj8q8d9q0 \
-pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpu2rqkx2sqq67d54psa77n2ksgmktzy8mhvscayjujchjwwcde3ls2trg7 \
-name validator1 \
-power 100
```

If the parameters are correct, the following confirmation output will be given:

```bash
Validator with address g1c56yp4k38dl2637zrwc7zqrzt95gtzj8q8d9q0 added to genesis file
```

To remove a specific validator from the genesis set, you can use the
`genesis validator remove` subcommand:

```bash
gnoland genesis validator remove --address g1c56yp4k38dl2637zrwc7zqrzt95gtzj8q8d9q0
```

If a validator with the given address was found, it will be removed from the genesis:

```bash
Validator with address g1c56yp4k38dl2637zrwc7zqrzt95gtzj8q8d9q0 removed from genesis file
```

### Managing Genesis Balances

By using the `genesis balances` subcommand you can control the initial balances
of specific addresses. This subcommand allows you to add & remove new balances,
as well as export the balances to a separate file.

#### Adding New Balances

To add new balances to the `genesis.json` file, use the following command:

```bash
gnoland genesis balances add [flags]
```

The flags for this subcommand give us multiple ways to manage the balance information:
- `-balance-sheet` - read balances from a separate file containing addresses in the format `<address>=<amount>ugnot`
- `-parse-export` - path to the transaction export containing a list of transactions (JSONL)
- `-single` - directly add a new balance item with the format `<address>=<amount>ugnot`

For example, you can add a single balance item with the following command:

```bash
gnoland genesis balances add --single g1c56yp4k38dl2637zrwc7zqrzt95gtzj8q8d9q0=1000ugnot
```

If the command was successful, the following output will be displayed:

```bash
1 pre-mines saved
```

#### Exporting Balances

To export the balances from the `genesis.json` file, use the following command:

```bash
gnoland genesis balances export [flags] <output-path>
```

For example, we can export the previously added balance item(s) to a `balances.txt`
file with the following command:

```bash
gnoland genesis balances export ./balances.txt
```

This will generate a `balances.txt` file with the following contents:

```text
g1c56yp4k38dl2637zrwc7zqrzt95gtzj8q8d9q0=1000ugnot
```

#### Removing Balances

To remove the balance information of a specific account from the `genesis.json` 
file, use the following command:

```bash
gnoland genesis balances remove [flags]
```

For example, we can remove the previously added item with the following command:

```bash
gnoland genesis balances remove --address g1c56yp4k38dl2637zrwc7zqrzt95gtzj8q8d9q0
```

If the command was successful, the following output will be displayed:

```bash
Pre-mine information for address g1c56yp4k38dl2637zrwc7zqrzt95gtzj8q8d9q0 removed
```

### Managing Genesis Transactions

By using the `genesis txs` subcommand you can manipulate the transactions to be
included in the `genesis.json` file. Transaction data that can be imported into
the `genesis.json` file is generated by the [`tx-archive`](https://github.com/gnolang/tx-archive) tool.

#### Adding a Transaction

To import transactions into the `genesis.json` file, use the following command:

```bash
gnoland genesis txs add [flags] <tx-file>
```

For example, we can add the following transaction found inside a 
`transactions.tx` file:

```json
{"msg":[{"@type":"/vm.m_call","caller":"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj","send":"","pkg_path":"gno.land/r/demo/users","func":"Invite","args":["g1thlf3yct7n7ex70k0p62user0kn6mj6d3s0cg3\ng1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5\n"]}],"fee":{"gas_wanted":"2000000","gas_fee":"1000000ugnot"},"signatures":[{"pub_key":{"@type":"/tm.PubKeySecp256k1","value":"AmG6kzznyo1uNqWPAYU6wDpsmzQKDaEOrVRaZ08vOyX0"},"signature":"njczE6xYdp01+CaUU/8/v0YC/NuZD06+qLind+ZZEEMNaRe/4Ln+4z7dG6HYlaWUMsyI1KCoB6NIehoE0PZ44Q=="}],"memo":""}
```

Running the subcommand with the path of the file will result in the addition of 
the transaction to the genesis transaction set:

```bash
gnoland genesis txs add transactions.tx 
```

If the command was successful, the following output will be displayed:
```bash
Saved 1 transactions to genesis.json
```

To find more example genesis transactions, take a look at [this file](https://github.com/gnolang/gno/blob/master/gno.land/genesis/genesis_txs.jsonl).

#### Exporting Transactions

To export the transactions from the `genesis.json` file, use the following command:

```bash
gnoland genesis txs export [flags] <output-path>
```

For example, we can export the previously added transaction to a `export.tx` file 
with the following command:

```bash
gnoland genesis txs export ./export.tx
```

This will generate a `export.tx` file with the following contents:

```json
{"msg":[{"@type":"/vm.m_call","caller":"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj","send":"","pkg_path":"gno.land/r/demo/users","func":"Invite","args":["g1thlf3yct7n7ex70k0p62user0kn6mj6d3s0cg3\ng1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5\n"]}],"fee":{"gas_wanted":"2000000","gas_fee":"1000000ugnot"},"signatures":[{"pub_key":{"@type":"/tm.PubKeySecp256k1","value":"AmG6kzznyo1uNqWPAYU6wDpsmzQKDaEOrVRaZ08vOyX0"},"signature":"njczE6xYdp01+CaUU/8/v0YC/NuZD06+qLind+ZZEEMNaRe/4Ln+4z7dG6HYlaWUMsyI1KCoB6NIehoE0PZ44Q=="}],"memo":""}
```

#### Removing Transactions

To remove transactions from the `genesis.json` file, you can use the following command,
passing in the `hex` encoded transaction hash:

```bash
gnoland genesis txs remove <tx-hash>
```

### Verifying a Genesis file

Using the `genesis verify` subcommand you can verify that a `genesis.json` file
exists and is not corrupted.

```
gnoland genesis verify [flags]
```

#### Flags

- `-genesis-path` - the path to the `genesis.json` file to verify

## Conclusion

That's it 🎉

This tutorial covered the essential steps to install, configure, and manage a
gno.land node using the `gnoland` binary. By using various subcommands and 
flags, you can effectively set up and maintain your gno.land node. 
For more details, refer to the `--help` option with each subcommand.