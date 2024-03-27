---
id: connect-from-go
---

# How to connect to a Gno.land chain from Go

This guide will show you how to read chain state, sign & broadcast transactions 
to a Gno.land network from your Go apps, using
[gnoclient](../reference/gnoclient/gnoclient.md).

- Call `Render()` on a realm
- Get account information from the chain
- Broadcast a state-changing transaction
- Evaluate an expression on a realm


## Prerequisites
- local keybase from gnokey
- possibly a bip39 mnemonic

## Installation
Add `gnoclient` to your Go project by running the following command:

```bash
go get github.com/gnolang/gno/gno.land/pkg/gnoclient
```

## Initialize `gnoclient.Client`

`gnoclient.Client` contains a `Signer` as well as a `RPCClient` connector:

```go 
type Client struct {
    Signer    Signer           // Signer for transaction authentication
    RPCClient rpcclient.Client // found in gnolang/gno/tm2/pkg/bft/rpc/client
}
```

### Signer

The `Signer` provides functionality to sign transactions with a Gno.land keypair.
The keypair can be accessed from a local keybase, or it can be generated 
in-memory from a BIP39 mnemonic.

```go
kb, _ := keys.NewKeyBaseFromDir("/path/to/keybase/dir")
signer := gnoclient.SignerFromKeybase{
    Keybase:  kb,       // keybase
    Account:  "mykey",  // name of account from keybase
    Password: "secure", // account password
}
```

:::info
The keybase directory path is set with the `gnokey --home` flag.
You can find your local keybase path from `gnokey` under the `-home` flag. 
:::

### without signer






Send transactions to the chain

Query the chain