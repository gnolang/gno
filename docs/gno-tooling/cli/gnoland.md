---
id: gno-tooling-gnoland
---

# gnoland

`gnoland` is the gno.land blockchain client binary, which is capable of managing
node working files, as well as starting the blockchain client itself.

### gnoland start [flags]

Starts the Gnoland blockchain node, with accompanying setup.

#### FLAGS

| Name                       | Type    | Description                                                                                                                                                                                                                                      |
|----------------------------|---------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `chainid`                  | String  | The ID of the chain. (default: `dev`)                                                                                                                                                                                                            |
| `data-dir`                 | String  | The path to the node's data directory. This is an important folder. The chain may fail to start if this folder is contaminated. (default: `gnoland-data`)                                                                                        |
| `flag-config-path`         | String  | The flag config file (optional).                                                                                                                                                                                                                 |
| `genesis`                  | String  | The path to the `genesis.json` file. (default: `genesis.json`)                                                                                                                                                                                   |
| `genesis-balance-file`     | String  | Initial distribution file. (default: `~/gno/gno.land/genesis/genesis_balances.txt`)                                                                                                                                                              |
| `genesis-max-vm-cycles`    | Int     | Sets maximum allowed vm cycles per operation. Zero means no limit. When increasing this option, the `block-max-gas` must also be increased to utilize the max cycles. (default: `100000000`)                                                     |
| `genesis-remote`           | String  | A replacement for `$$REMOTES%%` in genesis. (default: `localhost:26657`)                                                                                                                                                                         |
| `genesis-txs-file`         | String  | Initial txs to replay. (default: ~/gno/gno.land/genesis/genesis_txs.jsonl)                                                                                                                                                                       |
| `gnoroot-dir`              | String  | The root directory of the `gno` repository. (default: `~/gno`)                                                                                                                                                                                   |
| `lazy`                     | Boolean | Flag indication if lazy init is enabled. Generates the node secrets, configuration, and `genesis.json`. When set to `true`, you may start the chain without any initialization process, which comes in handy when developing. (default: `false`) |
| `log-format`               | String  | The log format for the gnoland node. (default: `console`)                                                                                                                                                                                        |
| `log-level`                | String  | The log level for the gnoland node. (default: `debug`)                                                                                                                                                                                           |
| `skip-failing-genesis-txs` | Boolean | Doesn’t panic when replaying invalid genesis txs. When starting a production-level chain, it is recommended to set this value to `true` to monitor and analyze failing transactions. (default: `false`)                                          |

### gnoland secrets \<subcommand\> [flags] [\<arg\>…]

The gno secrets manipulation suite for managing the validator key, p2p key and
validator state.

#### SUBCOMMANDS

| Name     | Description                                             |
|----------|---------------------------------------------------------|
| `init`   | Initializes required Gno secrets in a common directory. |
| `verify` | Verifies all Gno secrets in a common directory.         |
| `get`    | Shows all Gno secrets present in a common directory.    |

### gnoland secrets init [flags] [\<key\>]

Initializes the validator private key, the node p2p key and the validator's last
sign state. If a key is provided, it initializes the specified key.

Available keys:

- `validator_key` : The private key of the validator, which is different from the private key of the wallet.
- `node_id` : A key used for communicating with other nodes.
- `validator_state` : The current state of the validator such as the last signed block.

#### FLAGS

| Name       | Type   | Description                                                     |
|------------|--------|-----------------------------------------------------------------|
| `data-dir` | String | The secrets output directory. (default: `gnoland-data/secrets`) |
| `force`    | String | Overwrites existing secrets, if any. (default: `false`)         |

```bash
# force initialize all key
gnoland secrets init -force

Validator private key saved at gnoland-data/secrets/priv_validator_key.json
Validator last sign state saved at gnoland-data/secrets/priv_validator_state.json
Node key saved at gnoland-data/secrets/node_key.json


# force initialize a specific key type (ex: NodeKey)
gnoland secrets init node_key -force
Node key saved at gnoland-data/secrets/node_key.json
```

### gnoland secrets verify [flags] [\<key\>]

Verifies the validator private key, the node p2p key and the validator's last
sign state. If a key is provided, it verifies the specified key value.

Available keys: [`validator_key`, `node_id`, `validator_state`]

#### FLAGS

| Name       | Type   | Description                                                     |
|------------|--------|-----------------------------------------------------------------|
| `data-dir` | String | The secrets output directory. (default: `gnoland-data/secrets`) |

```bash
# verify all keys
gnoland secrets verify
Validator Private Key at gnoland-data/secrets/priv_validator_key.json is valid
Last Validator Sign state at gnoland-data/secrets/priv_validator_state.json is valid
Node P2P key at gnoland-data/secrets/node_key.json is valid


# verify a specific key type (ex: NodeKey)
gnoland secrets verify node_key
Node P2P key at gnoland-data/secrets/node_key.json is valid
```

### gnoland secrets get [flags] [\<key\>]

Shows the validator private key, the node p2p key and the validator's last sign
state. If a key is provided, it shows the specified key value.

Available keys: [`validator_key`, `node_key`, `validator_state`]

#### FLAGS

| Name       | Type   | Description                                                     |
|------------|--------|-----------------------------------------------------------------|
| `data-dir` | String | The secrets output directory. (default: `gnoland-data/secrets)` |

```bash
gnoland secrets get

{
    "validator_key": {
        "address": "g14j4dlsh3jzgmhezzp9v8xp7wxs4mvyskuw5ljl",
        "pub_key": "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqaqle3fdduqul4slg6zllypq9r8gj4wlfucy6qfnzmjcgqv675kxjz8jvk"
    },
    "validator_state": {
        "height": 0,
        "round": 0,
        "step": 0
    },
    "node_id": {
        "id": "g17h5t86vrztm6vuesx0xsyrg90wplj9mt9nsxng",
        "p2p_address": "g17h5t86vrztm6vuesx0xsyrg90wplj9mt9nsxng@0.0.0.0:26656"
    }
}

# will return node id info
gnoland secrets get node_id

# to get node id in cosmos
# gaiad tendermint show-node-id

# will return validator address and pub key
gnoland secrets get validator_key
# to get validator address in cosmos
# gaiad tendermint show-address

# to get validator pub key in cosmos
# gaiad tendermint show-validator
```

### gnoland config [subcommand] [flags]

The gno config manipulation suite for editing base and module configurations.

#### SUBCOMMANDS

| Name   | Description                             |
|--------|-----------------------------------------|
| `init` | Initializes the Gno node configuration. |
| `set`  | Edits the Gno node configuration.       |
| `get`  | Shows the Gno node configuration.       |

### gnoland config init [flags]

Initializes the Gno node configuration locally with default values, which
includes the base and module configurations.

#### FLAGS

| Name          | Type    | Description                                                                  |
|---------------|---------|------------------------------------------------------------------------------|
| `config-path` | String  | The path for the `config.toml`. (default: `gnoland-data/config/config.toml`) |
| `force`       | Boolean | Overwrites existing config.toml, if any. (default: `false`)                  |

```bash
# initialize the configuration file
gnoland config init

Default configuration initialized at gnoland-data/config/config.toml
```

### gnoland config set \<key\> \<value\>

Edits the Gno node configuration at the given path by setting the option
specified at `\<key\>` to the given `\<value\>`.

#### FLAGS

| Name          | Type   | Description                                                                  |
|---------------|--------|------------------------------------------------------------------------------|
| `config-path` | String | The path for the `config.toml`. (default: `gnoland-data/config/config.toml`) |

:::info
The `config set` command replaces the complexity of manually editing the `config.toml` file.
:::

### gnoland config get \<key\>

Shows the Gno node configuration at the given path by fetching the option
specified at `\<key\>`.

#### FLAGS

| Name          | Type   | Description                                                               |
|---------------|--------|---------------------------------------------------------------------------|
| `config-path` | String | the path for the config.toml (default: `gnoland-data/config/config.toml`) |

```bash
# check the current monkier (the displayed validator name)
gnoland config get -r moniker
n3wbie-MacBook-Pro.local

# set a new moniker
gnoland config set moniker hello
Updated configuration saved at gnoland-data/config/config.toml


# confirm the moniker change
gnoland config get -r moniker
hello
```
