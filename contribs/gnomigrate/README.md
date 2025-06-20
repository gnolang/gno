# `gnomigrate`: CLI Gno legacy data migration tool.

## Overview

`gnomigrate` is a CLI tool designed to migrate Gno legacy data formats to the new standard formats used in Gno
blockchain.

## Features

- **Transaction Migration**: Converts legacy `std.Tx` transactions to the new `gnoland.TxWithMetadata` format.

## Installation

### Clone the repository

```shell
git clone https://github.com/gnolang/gno.git
```

### Navigate to the project directory

```shell
cd contribs/gnomigrate
```

### Build the binary

```shell
make build
```

### Install the binary

```shell
make install
```

## Migrating legacy transactions

The `gnomigrate` tool provides the `txs` subcommand to manage the migration of legacy transaction sheets.

```shell
gnomigrate txs --input-dir <input_directory> --output-dir <output_directory>
```

### Flags

- `--input-dir`:  Specifies the directory containing the legacy transaction sheets to migrate.
- `--output-dir`:  Specifies the directory where the migrated transaction sheets will be saved.

### Example

```shell
gnomigrate txs --input-dir ./legacy_txs --output-dir ./migrated_txs
```

This command will:

- Read all `.jsonl` files from the ./legacy_txs directory, that are Amino-JSON encoded `std.Tx`s.
- Migrate each transaction from `std.Tx` to `gnoland.TxWithMetadata` (no metadata).
- Save the migrated transactions to the `./migrated_txs` directory, preserving the original directory structure.
