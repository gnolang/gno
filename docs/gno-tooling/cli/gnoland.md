---
id: gno-tooling-gnoland
---

# `gnoland`

## Run a gno.land node

`gnoland` is the Gno.land blockchain client binary. `gnoland` is capable of
managing node working files, as well as starting the blockchain client itself.

```bash
gnoland
```

### **Sub Commands**
| Command   | Description                    |
|-----------|--------------------------------|
| `start`   | Run the full node              |
| `secrets` | Gno secrets manipulation suite |
| `config`  | Gno config manipulation suite  |
| `genesis` | Gno genesis manipulation suite |

## `gnoland start`

Start the full blockchain node.

### **Options**

| Name                       | Type    | Description                                                                 |
|----------------------------|---------|-----------------------------------------------------------------------------|
| `chainid`                  | String  | The ID of the chain                                                         |
| `config-path`              | String  | The node TOML config file path (optional)                                   |
| `data-dir`                 | String  | The path to the node's data directory                                       |
| `flag-config-path`         | String  | The flag config file (optional)                                             |
| `genesis`                  | String  | The path to the genesis.json                                                |
| `genesis-balances-file`    | String  | Initial distribution file                                                   |
| `genesis-max-vm-cycles`    | Integer | Set maximum allowed VM cycles per operation. Zero means no limit.           |
| `genesis-remote`           | String  | Replacement for '%%REMOTE%%' in genesis                                     |
| `genesis-txs-file`         | String  | Initial txs to replay                                                       |
| `gnoroot-dir`              | String  | The root directory of the Gno repository                                    |
| `log-format`               | String  | Log format for the gnoland node                                             |
| `log-level`                | String  | Log level for the gnoland node                                              |
| `skip-failing-genesis-txs` | Boolean | Don't panic when replaying invalid genesis txs                              |
| `skip-start`               | Boolean | Quit after initialization, don't start the node                             |
| `tx-event-store-path`      | String  | Path for the file tx event store (required if event store is 'file')        |
| `tx-event-store-type`      | String  | Type of transaction event store                                             |


## `gnoland secrets`

Manages node secrets.

### **Sub Commands**
| Command  | Description                                            |
|----------|--------------------------------------------------------|
| `init`   | initializes required Gno secrets in a common directory |
| `verify` | verifies all Gno secrets in a common directory         |
| `get`    | shows all Gno secrets present in a common directory    |

### `gnoland secrets init`






