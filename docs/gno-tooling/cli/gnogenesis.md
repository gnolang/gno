---
id: gno-tooling-gnogenesis
---

# gnogenesis

`gnogenesis` is the gno.land blockchain `genesis.json` manipulation suite for managing genesis parameters.

#### SUBCOMMANDS

| Name        | Description                                 |
|-------------|---------------------------------------------|
| `generate`  | Generates a fresh `genesis.json`.           |
| `validator` | Validator set management in `genesis.json`. |
| `verify`    | Verifies a `genesis.json`.                  |
| `balances`  | Manages `genesis.json` account balances.    |
| `txs`       | Manages the initial genesis transactions.   |

### gnogenesis generate [flags]

Generates a node's `genesis.json` based on specified parameters.

#### FLAGS

| Name                   | Type   | Description                                                                                                                                                                                                 |
|------------------------|--------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `block-max-data-bytes` | Int    | The max size of the block data.(default: `2000000`)                                                                                                                                                         |
| `block-max-gas`        | Int    | The max gas limit for the block. (default: `100000000`)                                                                                                                                                     |
| `block-max-tx-bytes`   | Int    | The max size of the block transaction. (default: `1000000`)                                                                                                                                                 |
| `block-time-itoa`      | Int    | The block time itoa (in ms). (default: `100`)                                                                                                                                                               |
| `chain-id`             | String | The ID of the chain. (default: `dev`)                                                                                                                                                                       |
| `genesis-time`         | Int    | The genesis creation time. (default: `utc now timestamp`)                                                                                                                                                   |
| `output-path` :        | String | The output path for the `genesis.json`. If the genesis-time of the Genesis File is set to a future time, the chain will automatically start at that time if the node is online. (default: `./genesis.json`) |

### gnogenesis validator \<subcommand\> [flags]

Manipulates the `genesis.json` validator set.

#### SUBCOMANDS

| Name     | Description                                  |
|----------|----------------------------------------------|
| `add`    | Adds a new validator to the `genesis.json`.  |
| `remove` | Removes a validator from the `genesis.json`. |

#### FLAGS

| Name           | Type   | Description                                                |
|----------------|--------|------------------------------------------------------------|
| `address`      | String | The gno bech32 address of the validator.                   |
| `genesis-path` | String | The path to the `genesis.json`. (default `./genesis.json`) |

### gnogenesis validator add [flags]

Adds a new validator to the `genesis.json`.

#### FLAGS

| Name           | Type   | Description                                                     |
|----------------|--------|-----------------------------------------------------------------|
| `address`      | String | The gno bech32 address of the validator.                        |
| `genesis-path` | String | The path to the `genesis.json`. (default: `./genesis.json`)     |
| `name`         | String | The name of the validator (must be unique).                     |
| `power`        | Uint   | The voting power of the validator (must be > 0). (default: `1`) |
| `pub-key`      | String | The bech32 string representation of the validator's public key. |

```bash
gnogenesis validator add \
-address g1rzuwh5frve732k4futyw45y78rzuty4626zy6h \
-name test1 \
-pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zplmcmggxyxyrch0zcyg684yxmerullv3l6hmau58sk4eyxskmny9h7lsnz

Validator with address g1rzuwh5frve732k4futyw45y78rzuty4626zy6h added to genesis file
```

### gnogenesis validator remove [flags]

Removes a validator from the `genesis.json`.

#### FLAGS

| Name           | Type   | Description                                                 |
|----------------|--------|-------------------------------------------------------------|
| `address`      | String | The gno bech32 address of the validator.                    |
| `genesis-path` | String | The path to the `genesis.json`. (default: `./genesis.json)` |

```bash
gnogenesis validator remove \
-address g1rzuwh5frve732k4futyw45y78rzuty4626zy6h

Validator with address g1rzuwh5frve732k4futyw45y78rzuty4626zy6h removed from genesis file
```

### gnogenesis verify \<subcommand\> [flags] [\<arg\>…]

Verifies a `genesis.json`.

#### FLAGS

| Name           | Type   | Description                                               |
|----------------|--------|-----------------------------------------------------------|
| `genesis-path` | String | The path to the `genesis.json`. (default: `genesis.json`) |

### gnogenesis balances \<subcommand\> [flags] [\<arg\>…]

Manages `genesis.json` account balances.

#### SUBCOMMANDS

| Name     | Description                                            |
|----------|--------------------------------------------------------|
| `add`    | Adds the balance information.                          |
| `remove` | Removes the balance information of a specific account. |

### gnogenesis balances add [flags]

#### FLAGS

| Name            | Type   | Description                                                                                |
|-----------------|--------|--------------------------------------------------------------------------------------------|
| `balance-sheet` | String | The path to the balance file containing addresses in the format `<address>=<amount>ugnot`. |
| `genesis-path`  | String | The path to the `genesis.json` (default: `./genesis.json`)                                 |
| `parse-export`  | String | The path to the transaction export containing a list of transactions (JSONL).              |
| `single`        | String | The direct balance addition in the format `<address>=<amount>ugnot`.                       |

```bash
gnogenesis balances add \
-single g1rzuwh5frve732k4futyw45y78rzuty4626zy6h=100ugnot

1 pre-mines saved

g1rzuwh5frve732k4futyw45y78rzuty4626zy6h:{[24 184 235 209 35 102 125 21 90 169 226 200 234 208 158 56 197 197 146 186] [{%!d(string=ugnot) 100}]}ugnot
```

### gnoland balances remove [flags]

#### FLAGS

| Name           | Type   | Description                                                                                 |
|----------------|--------|---------------------------------------------------------------------------------------------|
| `address`      | String | The address of the account whose balance information should be removed from `genesis.json`. |
| `genesis-path` | String | The path to the `genesis.json`. (default: `./genesis.json`)                                 |

```bash
gnogenesis balances remove \
-address=g1rzuwh5frve732k4futyw45y78rzuty4626zy6h

Pre-mine information for address g1rzuwh5frve732k4futyw45y78rzuty4626zy6h removed
```

### gnoland txs \<subcommand\> [flags] [\<arg\>…]

Manages genesis transactions through input files.

#### SUBCOMMANDS

| Name     | Description                                       |
|----------|---------------------------------------------------|
| `add`    | Imports transactions into the `genesis.json`.     |
| `remove` | Removes the transactions from the `genesis.json`. |
| `export` | Exports the transactions from the `genesis.json`. |

### gnogenesis txs \<subcommand\> [flags] [\<arg\>…]

Manages genesis transactions through input files.

#### SUBCOMMANDS

| Name     | Description                                       |
|----------|---------------------------------------------------|
| `add`    | Imports transactions into the `genesis.json`.     |
| `remove` | Removes the transactions from the `genesis.json`. |
| `export` | Exports the transactions from the `genesis.json`. |
