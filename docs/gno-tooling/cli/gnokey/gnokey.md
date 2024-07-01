---
id: gnokey
---

# `gnokey`

## Overview

In this tutorial section, you will learn how to use the `gnokey` binary.
`gnokey` allows you to interact with a Gno.land chain - you will learn how manage your keypairs, create
state-changing calls, run readonly queries without using gas, as well as create, 
sign, and broadcast airgapped transactions for full security.

// TODO FIX 

## Prerequisites

- **`gno`, `gnokey`, and `gnodev` installed.** Reference the
  [Local Setup](../../../getting-started/local-setup/installation.md#2-installing-the-required-tools-) guide for steps
- **A Gno.land keypair set up.** Reference the
  [Working with Key Pairs](working-with-key-pairs.md) guide for steps

## Interacting with a Gno.land chain

`gnokey` allows you to interact with any Gno.land network, such as the
[Portal Loop](../../../concepts/portal-loop.md) testnet.

There are multiple ways anyone can interact with the chain:
- Transactions - state-changing calls which use gas
- ABCI queries - read-only calls which do not use gas

Both transactions and ABCI queries can be made via `gnokey`'s subcommands,
`maketx` and `query`.


TODO FIX