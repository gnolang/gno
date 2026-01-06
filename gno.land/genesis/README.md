# gno.land genesis

This directory contains the genesis configuration files for the Gnoland blockchain.

## Genesis Balances (`genesis_balances.txt`)

Plain text file with one balance entry per line:

```
<bech32_address>=<amount>ugnot # optional comment
```

Example:
```
g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=10000000000000ugnot # test1
```

## Genesis Transactions (`genesis_txs.jsonl`)

JSONL file containing transactions to execute at genesis. Each line is an Amino JSON-encoded transaction.

For comprehensive format documentation, see the [gnogenesis README](../../contribs/gnogenesis/README.md#genesis-transaction-sheet-format-reference).

## Tools

Use the [`gnogenesis`](../../contribs/gnogenesis/) CLI tool to manage genesis files:

```bash
# Add transactions from a sheet
gnogenesis txs add sheets ./txs.jsonl

# Add packages from a directory
gnogenesis txs add packages ./examples --deployer-address=<address>

# Verify genesis.json integrity
gnogenesis verify --genesis-path ./genesis.json
```
