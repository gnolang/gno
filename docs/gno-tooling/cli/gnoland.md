---
id: gno-tooling-gnoland
---

# gnoland

## Run a Gnoland Node

Start a node on the Gnoland blockchain with the following command.

```bash
gnoland
```

### **Sub Commands**
| Command   | Description       |
| --------- | ----------------- |
| `start`   | Run the full node |


### **Options**

| Name                       | Type    | Description                                                                             |
|----------------------------| ------- | --------------------------------------------------------------------------------------- |
| `chainid`                  | String  | The id of the chain (default: `dev`).                                                   |
| `genesis-balances-file`    | String  | The initial GNOT distribution file (default: `./gnoland/genesis/genesis_balances.txt`). |
| `genesis-remote`           | String  | Replacement '%%REMOTE%%' in genesis (default: `"localhost:26657"`).                     |
| `genesis-txs-file`         | String  | Initial txs to be executed (default: `"./gnoland/genesis/genesis_txs.jsonl"`).          |
| `data-dir`                 | String  | directory for config and data (default: `gnoland-data`).                                     |
| `skip-failing-genesis-txs` | Boolean | Skips transactions that fail from the `genesis-txs-file`                                |
| `skip-start`               | Boolean | Quits after initialization without starting the node.                                   |
