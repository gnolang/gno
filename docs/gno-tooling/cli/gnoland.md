---
id: gno-tooling-gnoland
---

# gnoland

`gnoland` is the Gno.land blockchain client binary, which is capable of managing node working files, as well as starting the blockchain client itself.

## Gno Command Syntax Guide

### gnoland <subcommand> [flags] [<arg>...]

#### Subcommand

The gnoland command consists of various purpose-built subcommands.

#### Flags

Options of the subcommand.

- `gnoland start [-chainid]` : Allows you to configure options of the start subcommand such as the Chain ID.

#### Arg

The argument of the flag .

- `gnoland start -chainid [chain_id_value_here]` : The argument of the chainid flag.

### gnoland start [flags]

Starts the Gnoland blockchain node, with accompanying setup.

#### FLAGS

| Name | Type | Description |
| --- | --- | --- |
| `chainid` | String | The ID of the chain. (default: `dev`) |
| `data-dir` | String | The path to the node's data directory. This is an important folder. The chain may fail to start if this folder is contaminated. (default: `gnoland-data`) |
| `flag-config-path` | String | The flag config file (optional). |
| `genesis` | String | The path to the `genesis.json` file. (default: `genesis.json`) |
| `genesis-balance-file` | String | Initial distribution file. (default: `~/gno/gno.land/genesis/genesis_balances.txt`) |
| `genesis-max-vm-cycles` | Int | Sets maximum allowed vm cycles per operation. Zero means no limit. When increasing this option, the `block-max-gas` must also be increased to utilize the max cycles. (default: `100000000`) |
| `genesis-remote` | String | A replacement for `$$REMOTES%%` in genesis. (default: `localhost:26657`) |
| `genesis-txs-file` | String | Initial txs to replay. (default: ~/gno/gno.land/genesis/genesis_txs.jsonl) |
| `gnoroot-dir` | String | The root directory of the `gno` repository. (default: `~/gno`) |
| `lazy` | Boolean | Flag indication if lazy init is enabled. Generates the node secrets, configuration, and `genesis.json`. When set to `true`, you may start the chain without any initialization process, which comes in handy when developing. (default: `false`) |
| `log-format` | String | The log format for the gnoland node. (default: `console`) |
| `log-level` | String | The log level for the gnoland node. (default: `debug`) |
| `skip-failing-genesis-txs` | Boolean | Doesn’t panic when replaying invalid genesis txs. When starting a production-level chain, it is recommended to set this value to `true` to monitor and analyze failing transactions. (default: `false`) |

### gnoland genesis <subcommand> [flags] [<arg>...]

Gno `genesis.json` manipulation suite for managing genesis parameters.

#### SUBCOMMANDS

| Name        | Description                                 |
| ----------- | ------------------------------------------- |
| `generate`  | Generates a fresh `genesis.json`.           |
| `validator` | Validator set management in `genesis.json`. |
| `verify`    | Verifies a `genesis.json`.                  |
| `balances`  | Manages `genesis.json` account balances.    |
| `txs`       | Manages the initial genesis transactions.   |

### gnoland genesis generate [flags]

Generates a node's `genesis.json` based on specified parameters.

#### FLAGS

| Name | Type | Description |
| --- | --- | --- |
| `block-max-data-bytes` | Int | The max size of the block data.(default: `2000000`) |
| `block-max-gas` | Int | The max gas limit for the block. (default: `100000000`) |
| `block-max-tx-bytes` | Int | The max size of the block transaction. (default: `1000000`) |
| `block-time-itoa` | Int | The block time itoa (in ms). (default: `100`) |
| `chain-id` | String | The ID of the chain. (default: `dev`) |
| `genesis-time` | Int | The genesis creation time. (default: `utc now timestamp`) |
| `output-path` : | String | The output path for the `genesis.json`. If the genesis-time of the Genesis File is set to a future time, the chain will automatically start at that time if the node is online. (default: `./genesis.json`) |

### gnoland genesis validator <subcommand> [flags]

Manipulates the `genesis.json` validator set.

#### SUBCOMANDS

| Name     | Description                                  |
| -------- | -------------------------------------------- |
| `add`    | Adds a new validator to the `genesis.json`.  |
| `remove` | Removes a validator from the `genesis.json`. |

#### FLAGS

| Name           | Type   | Description                                                |
| -------------- | ------ | ---------------------------------------------------------- |
| `address`      | String | The gno bech32 address of the validator.                   |
| `genesis-path` | String | The path to the `genesis.json`. (default `./genesis.json`) |

### gnoland genesis validator add [flags]

Adds a new validator to the `genesis.json`.

#### FLAGS

| Name           | Type   | Description                                                     |
| -------------- | ------ | --------------------------------------------------------------- |
| `address`      | String | The gno bech32 address of the validator.                        |
| `genesis-path` | String | The path to the `genesis.json`. (default: `./genesis.json`)     |
| `name`         | String | The name of the validator (must be unique).                     |
| `power`        | Uint   | The voting power of the validator (must be > 0). (default: `1`) |
| `pub-key`      | String | The bech32 string representation of the validator's public key. |

```bash
$ gnoland genesis validator add -address g1rzuwh5frve732k4futyw45y78rzuty4626zy6h --name test1 --pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zplmcmggxyxyrch0zcyg684yxmerullv3l6hmau58sk4eyxskmny9h7lsnz
Validator with address g1rzuwh5frve732k4futyw45y78rzuty4626zy6h added to genesis file
```

### gnoland genesis validator remove [flags]

Removes a validator from the `genesis.json`.

#### FLAGS

| Name           | Type   | Description                                                 |
| -------------- | ------ | ----------------------------------------------------------- |
| `address`      | String | The gno bech32 address of the validator.                    |
| `genesis-path` | String | The path to the `genesis.json`. (default: `./genesis.json)` |

```bash
$ gnoland genesis validator remove -address g1rzuwh5frve732k4futyw45y78rzuty4626zy6h
Validator with address g1rzuwh5frve732k4futyw45y78rzuty4626zy6h removed from genesis file
```

### gnoland genesis verify <subcommand> [flags] [<arg>…]

Verifies a `genesis.json`.

#### FLAGS

| Name           | Type   | Description                                               |
| -------------- | ------ | --------------------------------------------------------- |
| `genesis-path` | String | The path to the `genesis.json`. (default: `genesis.json`) |

#### gnoland genesis balances <subcommand> [flags] [<arg>…]

Manages `genesis.json` account balances.

#### SUBCOMMANDS

| Name     | Description                                            |
| -------- | ------------------------------------------------------ |
| `add`    | Adds the balance information.                          |
| `remove` | Removes the balance information of a specific account. |

### gnoland genesis balances add [flags]

#### FLAGS

| Name            | Type   | Description                                                                                |
| --------------- | ------ | ------------------------------------------------------------------------------------------ |
| `balance-sheet` | String | The path to the balance file containing addresses in the format `<address>=<amount>ugnot`. |
| `genesis-path`  | String | The path to the `genesis.json` (default: `./genesis.json`)                                 |
| `parse-export`  | String | The path to the transaction export containing a list of transactions (JSONL).              |
| `single`        | String | The direct balance addition in the format `<address>=<amount>ugnot`.                       |

```bash
$ gnoland genesis balances add -single g1rzuwh5frve732k4futyw45y78rzuty4626zy6h=100ugnot
1 pre-mines saved

g1rzuwh5frve732k4futyw45y78rzuty4626zy6h:{[24 184 235 209 35 102 125 21 90 169 226 200 234 208 158 56 197 197 146 186] [{%!d(string=ugnot) 100}]}ugnot
```

### gnoland balances remove [flags]

#### FLAGS

| Name           | Type   | Description                                                                                 |
| -------------- | ------ | ------------------------------------------------------------------------------------------- |
| `address`      | String | The address of the account whose balance information should be removed from `genesis.json`. |
| `genesis-path` | String | The path to the `genesis.json`. (default: `./genesis.json`)                                 |

```bash
$ gnoland genesis balances remove -address=g1rzuwh5frve732k4futyw45y78rzuty4626zy6h
Pre-mine information for address g1rzuwh5frve732k4futyw45y78rzuty4626zy6h removed
```

### gnoland txs <subcommand> [flags] [<arg>…]

Manages genesis transactions through input files.

#### SUBCOMMANDS

| Name     | Description                                       |
| -------- | ------------------------------------------------- |
| `add`    | Imports transactions into the `genesis.json`.     |
| `remove` | Removes the transactions from the `genesis.json`. |
| `export` | Exports the transactions from the `genesis.json`. |

### gnoland secrets <subcommand> [flags] [<arg>…]

The gno secrets manipulation suite for managing the validator key, p2p key and validator state.

#### SUBCOMMANDS

| Name     | Description                                             |
| -------- | ------------------------------------------------------- |
| `init`   | Initializes required Gno secrets in a common directory. |
| `verify` | Verifies all Gno secrets in a common directory.         |
| `get`    | Shows all Gno secrets present in a common directory.    |

### gnoland secrets init [flags] [<key>]

Initializes the validator private key, the node p2p key and the validator's last sign state. If a key is provided, it initializes the specified key.

- Available keys
  - `ValidatorPrivateKey` : The private key of the validator, which is different from the private key of the wallet.
  - `NodeKey` : A key used for communicating with other nodes.
  - `ValidatorState` : The current state of the validator such as the last signed block.

#### FLAGS

| Name       | Type   | Description                                                     |
| ---------- | ------ | --------------------------------------------------------------- |
| `data-dir` | String | The secrets output directory. (default: `gnoland-data/secrets`) |
| `force`    | String | Overwrites existing secrets, if any. (default: `false`)         |

```bash
# force initialize all key
$ gnoland secrets init -force
Validator private key saved at gnoland-data/secrets/priv_validator_key.json
Validator last sign state saved at gnoland-data/secrets/priv_validator_state.json
Node key saved at gnoland-data/secrets/node_key.json


# force initialize a specific key type (ex: NodeKey)
$ gnoland secrets init NodeKey -force
Node key saved at gnoland-data/secrets/node_key.json
```

### gnoland secrets verify [flags] [<key>]

Verifies the validator private key, the node p2p key and the validator's last sign state. If a key is provided, it verifies the specified key value.

- Available keys: [ValidatorPrivateKey, NodeKey, ValidatorState]

#### FLAGS

| Name       | Type   | Description                                                     |
| ---------- | ------ | --------------------------------------------------------------- |
| `data-dir` | String | The secrets output directory. (default: `gnoland-data/secrets`) |

```bash
# verify all keys
$ gnoland secrets verify
Validator Private Key at gnoland-data/secrets/priv_validator_key.json is valid
Last Validator Sign state at gnoland-data/secrets/priv_validator_state.json is valid
Node P2P key at gnoland-data/secrets/node_key.json is valid


# verify a specific key type (ex: NodeKey)
$ gnoland secrets verify NodeKey
Node P2P key at gnoland-data/secrets/node_key.json is valid
```

### gnoland secrets get [flags] [<key>]

Shows the validator private key, the node p2p key and the validator's last sign state. If a key is provided, it shows the specified key value.

- Available keys: [`ValidatorPrivateKey`, `NodeKey`, `ValidatorState`]

#### FLAGS

| Name       | Type   | Description                                                     |
| ---------- | ------ | --------------------------------------------------------------- |
| `data-dir` | String | The secrets output directory. (default: `gnoland-data/secrets)` |

```bash
$ gnoland secrets get
[Node P2P Info]
Node ID:  g1lhn5zztl8vgaccper5uh99fhrdtewn6dmu8pnc

[Validator Key Info]
Address:     g1vn0jjsge9yv740kjrdgjrvvuukdncstljs90yh
Public Key:  gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq0wf2kadne5hztjch2hl38k6gcdfyzrr9c5l0awz2dw6v3ul6n654xhvua

[Last Validator Sign State Info]
Height:  0
Round:   0
Step:    0



# will return node id
$ gnoland secrets get NodeKey
# to get node id in cosmos
$ gaiad tendermint show-node-id

# will return validator address and pub key
$ gnoland secrets get ValidatorPrivateKey
# to get validator address in cosmos
$ gaiad tendermint show-address
# to get validator pub key in cosmos
$ gaiad tendermint show-validator
```

### gnoland config [subcommand] [flags]

The gno config manipulation suite for editing base and module configurations.

#### SUBCOMMANDS

| Name   | Description                             |
| ------ | --------------------------------------- |
| `init` | Initializes the Gno node configuration. |
| `set`  | Edits the Gno node configuration.       |
| `get`  | Shows the Gno node configuration.       |

### gnoland config init [flags]

Initializes the Gno node configuration locally with default values, which includes the base and module configurations.

#### FLAGS

| Name          | Type    | Description                                                                  |
| ------------- | ------- | ---------------------------------------------------------------------------- |
| `config-path` | String  | The path for the `config.toml`. (default: `gnoland-data/config/config.toml`) |
| `force`       | Boolean | Overwrites existing config.toml, if any. (default: `false`)                  |

```bash
# initialize the configuration file
$ gnoland config init
Default configuration initialized at gnoland-data/config/config.toml
```

### gnoland config set <key> <value>

Edits the Gno node configuration at the given path by setting the option specified at `<key>` to the given `<value>`.

#### FLAGS

| Name          | Type   | Description                                                                  |
| ------------- | ------ | ---------------------------------------------------------------------------- |
| `config-path` | String | The path for the `config.toml`. (default: `gnoland-data/config/config.toml`) |

:::info The `config set` command replaces the complexity of manual editing of the `config.toml` file required in Cosmos chains. :::

### gnoland config get <key>

Shows the Gno node configuration at the given path by fetching the option specified at `<key>`.

#### FLAGS

| Name          | Type   | Description                                                               |
| ------------- | ------ | ------------------------------------------------------------------------- |
| `config-path` | String | the path for the config.toml (default: `gnoland-data/config/config.toml`) |

```bash
# check the current monkier (the displayed validator name)
$ gnoland config get moniker
n3wbie-MacBook-Pro.local

# set a new moniker
$ gnoland config set moniker hello
Updated configuration saved at gnoland-data/config/config.toml


# confirm the moniker change
$ gnoland config get moniker
hello
```
