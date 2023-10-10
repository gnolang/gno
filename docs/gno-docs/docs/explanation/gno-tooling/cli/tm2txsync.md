---
id: gno-tooling-tm2txsync
---

# tm2txsync

`tm2txsync` is used for backing up a local node's transactions.

## Import Transaction Data To (or Export It From) a Node

You may import or export transaction data with the following command.

```bash
tm2txsync {SUB_COMMAND}
```

#### **Subcommands**

| Name     | Description                |
| -------- | -------------------------- |
| `export` | Exports txs from the node. |
| `import` | Imports txs to the node.   |

### Import

#### **Options**

| Name     | Type   | Description                                                       |
| -------- | ------ | ----------------------------------------------------------------- |
| `remote` | String | The Remote RPC in `addr:port` format (default: `localhost:26657`) |
| `in`     | String | The input file path (default: `txexport.log`)                     |

### Export

#### **Options**

| Name     | Type    | Description                                                       |
| -------- | ------- | ----------------------------------------------------------------- |
| `remote` | String  | The Remote RPC in `addr:port` format (default: `localhost:26657`) |
| `start`  | Int64   | Starting height (default: `1`)                                    |
| `tail`   | Int64   | Start at LAST - N.                                                |
| `end`    | Int64   | End height (optional)                                             |
| `out`    | String  | The output file path (default: `txexport.log`)                    |
| `quiet`  | Boolean | Quiet mode.                                                       |
| `follow` | Boolean | Keep attached and follow new events.                              |
