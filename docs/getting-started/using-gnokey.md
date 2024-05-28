---
id: using-gnokey
---

# Using `gnokey`

## Overview
In this tutorial, you will learn how to use `gnokey`, a tool used for 


which are
required for interacting with the Gno.land blockchain. You will understand what
mnemonics are, how they are used, and how you can make interaction seamless with
Gno.

## Prerequisites
- **`gno` & `gnokey` installed.** Reference the
  [Local Setup](local-setup/installation.md#2-installing-the-required-tools-) guide for steps

## Keypairs


## Interacting with a Gno.land chain

`gnokey` allows you to interact with any Gno.land chain, such as the Portal Loop. 

There are multiple ways anyone can interact with the chain:
- Transactions - state-changing calls which use up gas
- ABCI queries - read-only calls which do not use up gas

Both transactions and ABCI queries can be used via `gnokey`'s subcommands,
`maketx` and `query`.

## State-changing calls

In Gno, there are three types of messages that can change on-chain state:
- `AddPackage` - adds code to the chain
- `Call` - calls a specific path and function on the chain
- `Run` - executes a Gno script against on-chain code

A Gno.land transaction contains two main things: 
- A base configuration where variables such as `gas-fee`, `gas-wanted`, and others
are defined
- A list of messages to execute on the chain 

Currently, `gnokey` supports single-message transactions, while multiple-message
transactions can be created in Go programs, supported by the
[gnoclient](../reference/gnoclient/gnoclient.md) package.

Let's delve deeper into each of these messages.

### `AddPackage`

In case you want to upload new code to the chain, you can use the `AddPackage` 
message type. You can send an `AddPackage` transaction with `gnokey` using the 
following command:

```bash
gnokey maketx addpkg
```

To understand how to use this subcommand better, let's create a folder with some
example code. First create an empty directory somewhere on disk:



The `addpkg` subcommmand takes the following parameters:
- `--pkgpath` - on-chain path where your code will be uploaded to
- `--pkgdir` - local path where your is located
- `--gas-wanted` - the upper limit for units of gas for the execution of the
transaction
- `--gas-fee` - similar to Solidity's `gas-price`
- `--broadcast` - broadcast the transaction on-chain
- `--chain-id` - id of the chain to connect to, in our case the local node, `dev`
- `--remote` - specify node endpoint, in our case it's our local node
- `Dev` - the keypair to use for the transaction



gnokey maketx addpkg \                                                                                                                                                                                          
--pkgpath "gno.land/r/leon/v5/memeland" \
--pkgdir "." \
--gas-fee 10000000ugnot \
--gas-wanted 8000000 \
--broadcast \
--chainid portal-loop \
--remote https://rpc.gno.land:443 \
mykey







