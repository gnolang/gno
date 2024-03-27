---
id: connect-from-go
---

# How to connect to a Gno.land chain from Go

This guide will show you how to connect to a Gno.land network from your Go apps,
using [gnoclient](../reference/gnoclient/gnoclient.md).

Gnoclient provides a simple API to sign & broadcast 

## Prerequisites
- local keybase from gnokey
- possibly a bip39 mnemonic

## Installation
Add `gnoclient` to your Go project by running the following command:

```bash
go get github.com/gnolang/gno/gno.land/pkg/gnoclient
```

## Initialize `gnoclient`

`gnoclient` can be initialized in two modes:
- with `Signer` - read & write access to the network 
- without `Signer` - read-only access to the network

### Signer

The `Signer` provides functionality to sign transactions with a Gno.land keypair.
The keypair can be generated in-memory from a BIP39 mnemonic, or can be accessed
from a local keybase.


```go

```

### without signer






Send transactions to the chain

Query the chain