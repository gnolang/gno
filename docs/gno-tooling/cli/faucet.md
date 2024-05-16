---
id: gno-tooling-faucet
---

# TM2 Faucet

TM2 Faucet is a versatile command-line interface (CLI) tool designed to 
effortlessly deploy a faucet server for Gno Tendermint 2 networks. The tool can
be found on the [`gnolang/faucet`](https://github.com/gnolang/faucet) repo.

## Run `faucet` Commands

### `generate`

Generates a `config.toml` file for the faucet.

-output-path ./config.toml  the path to output the TOML configuration file

#### **Options**

| Name          | Type    | Description                                    |
|---------------|---------|------------------------------------------------|
| `output-path` | String  | The path to output the TOML configuration file |

### `serve`

Starts the faucet with the default or given config.

```bash
faucet serve
```

#### **Options**

| Name             | Type   | Description                                                      |
|------------------|--------|------------------------------------------------------------------|
| `chain-id`       | String | The id of the chain (required).                                  |
| `config`         | String | the path to the command configuration file [TOML]                |
| `faucet-config`  | String | the path to the faucet TOML configuration, if any                |
| `gas-fee`        | String | he static gas fee for the transaction. Format: <AMOUNT>ugnot     |
| `gas-wanted`     | String | the static gas wanted for the transaction. Format: <AMOUNT>ugnot |
| `listen-address` | String | the IP:PORT URL for the faucet server                            |
| `mnemonic`       | String | the mnemonic for faucet keys                                     |
| `num-accounts`   | String | the number of faucet accounts, based on the mnemonic             |
| `remote`         | String | the JSON-RPC URL of the Gno chain                                |
| `send-amount`    | String | the static max send amount per drip (native currency)            |