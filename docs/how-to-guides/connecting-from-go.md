---
id: connect-from-go
---

# How to connect a Go app to Gno.land 

This guide will show you how to connect to a Gno.land network from your Go application,
using the [gnoclient](../reference/gnoclient/gnoclient.md) package.

For this guide, we will build a small Go app that will:

- Call `Render()` on a realm
- Get account information from the chain
- Broadcast a state-changing transaction
- Evaluate an expression on a realm

## Prerequisites
- A Local Gno.land keypair generated using
[gnokey](../getting-started/local-setup/working-with-key-pairs.md)

## Installation
Add `gnoclient` to your Go project by running the following command:

```bash
go get github.com/gnolang/gno/gno.land/pkg/gnoclient
```

## Main components

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

:::info
The keybase directory path is set with the `gnokey --home` flag.
You can find your local keybase path from `gnokey` under the `-home` flag. 
:::

### RPCClient

The `RPCCLient` provides connectivity to a Gno.land network via HTTP or WebSockets.


## 




Send transactions to the chain

Query the chain