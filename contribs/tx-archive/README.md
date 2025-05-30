<h2 align="center">⚛️ Tendermint2 Backup / Restore ⚛️</h2>

## Overview

The `tx-archive` is a powerful command-line interface (CLI) tool designed to streamline the process of backing up and
restoring transaction data from Tendermint2 chains. It does so by utilizing the
efficient [jsonl](https://jsonlines.org/) format.

## Backup

With the `backup` subcommand, users can perform the following actions:

- Create backups of transactions from a running Tendermint2 node.
- Define backup limits by specifying start and end block numbers.
- Enable live backups of a running Tendermint2 node using the `--watch` flag, allowing for the capture of incoming
  transactions in real-time

Options available for backup:

```bash
USAGE
  backup [flags]

Runs the chain backup service

FLAGS
  -from-block 1                   the starting block number for the backup (inclusive)
  -legacy=false                   flag indicating if the legacy output format should be used (tx-per-line)
  -output-path ./backup.jsonl     the output path for the JSONL chain data
  -overwrite=false                flag indicating if the output file should be overwritten during backup
  -remote http://127.0.0.1:26657  the JSON-RPC URL of the chain to be backed up
  -to-block -1                    the end block number for the backup (inclusive). If <0, latest chain height is used
  -watch=false                    flag indicating if the backup should append incoming tx data
```

**Note**: this backup tool uses `amino.MarshalJSON` to package backup data, so make sure your client can understand
Amino JSON.

## Restore

With the `restore` subcommand, users can perform the following actions:

- Restore (replay) transactions from an input file.
- Set up live restore (replay) tracking, allowing the tool to monitor changes to the input file.

Options available for restore:

```bash
USAGE
  restore [flags]

Runs the chain restore service

FLAGS
  -input-path string              the input path for the JSONL chain data
  -legacy=false                   flag indicating if the input file is legacy amino JSON
  -remote http://127.0.0.1:26657  the JSON-RPC URL of the chain to be backed up
  -watch=false                    flag indicating if the restore should watch incoming tx data
```

## Formats

### Standard

The standard format wraps transaction data (`std.Tx`) into a structure associated with a block number:

```go
package types

type TxData struct {
	Tx       std.Tx `json:"tx"`
	BlockNum uint64 `json:"blockNum"`
}
```

### Legacy

The "legacy" format contains only the transaction data (`std.Tx`) without any accompanying block information. It is used
by existing Gno archive repositories like [tx-exports](https://github.com/gnolang/tx-exports):

```go
package std

type Tx struct {
	Msgs       []Msg       `json:"msg" yaml:"msg"`
	Fee        Fee         `json:"fee" yaml:"fee"`
	Signatures []Signature `json:"signatures" yaml:"signatures"`
	Memo       string      `json:"memo" yaml:"memo"`
}
```

Please ensure you choose the appropriate format depending on your use case and compatibility requirements.
