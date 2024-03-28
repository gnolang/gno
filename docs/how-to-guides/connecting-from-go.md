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

## Setup

To get started, create a new Go project. In a clean directory, run 
```bash
go mod init example
```

After this, create a new `main.go` file:

```bash
touch main.go
```

Set up your main file with the code below:

```go
package main

func main() {}
```

Finally, add the `gnoclient` package by running the following command:

```bash
go get github.com/gnolang/gno/gno.land/pkg/gnoclient
```

## Main components

The `gnoclient` package exposes a `Client` struct containing a `Signer` and 
`RPCClient` connector. `Client` exposes all available functionality for talking
to a Gno.land chain.

```go 
type Client struct {
    Signer    Signer           // Signer for transaction authentication
    RPCClient rpcclient.Client // gnolang/gno/tm2/pkg/bft/rpc/client
}
```

### Signer

The `Signer` provides functionality to sign transactions with a Gno.land keypair.
The keypair can be accessed from a local keybase, or it can be generated 
in-memory from a BIP39 mnemonic.

:::info
The keybase directory path is set with the `gnokey --home` flag. 
:::

### RPCClient

The `RPCCLient` provides connectivity to a Gno.land network via HTTP or WebSockets.


## Initialize the Signer

For this example, we will initialize the `Signer` from a local keybase: 

```go
package main

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

func main() {
	// Initialize keybase from a directory
	keybase, _ := keys.NewKeyBaseFromDir("path/to/keybase/dir")

	// Create signer
	signer := gnoclient.SignerFromKeybase{
		Keybase:  keybase,
		Account:  "<keypair_name>",     // Name of your keypair in keybase
		Password: "<keypair_password>", // Password to decrypt your keypair 
		ChainID:  "<gno_chainID>",      // id of Gno.land chain
	}
}
```

A few things to note:
- You can view keys in your local keybase by running `gnokey list`.  
- You can get the password from a user input using the IO package.

## Initialize the RPC connection & Client

You can initialize the RPC Client used to connect to the Gno.land network with
the following code:
```go
rpc := rpcclient.NewHTTP("<gno_chain_endpoint>", "/websocket")
```

A list of Gno.land network endpoints & chain IDs can be found in the [Gno RPC 
endpoints](../reference/rpc-endpoints.md#network-configurations) page. 

With this, we can initialize the `gnoclient.Client` struct: 

```go
package main

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

func main() {
	// Initialize keybase from a directory
	keybase, _ := keys.NewKeyBaseFromDir("path/to/keybase/dir")

	// Create signer
	signer := gnoclient.SignerFromKeybase{
		Keybase:  keybase,
		Account:  "<keypair_name>",     // Name of your keypair in keybase
		Password: "<keypair_password>", // Password to decrypt your keypair 
		ChainID:  "<gno_chainID>",      // id of Gno.land chain
	}

	// Initialize the RPC client
	rpc := rpcclient.NewHTTP("<gno.land_remote_endpoint>", "")
	
	// Initialize the gnoclient
	client := gnoclient.Client{
		Signer:    signer,
		RPCClient: rpc,
	}
}
```

We can now communicate with the Gno.land chain. Let's explore some of the functionality
`gnoclient` provides.

## Query account info from a chain

To send transactions to the chain, we need to know the account number (ID) and 
sequence (nonce). We can get this information by querying the chain with the
`QueryAccount` function:

```go
// Convert Gno address string to `crypto.Address`
addr, err := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // your Gno address
if err != nil {
	panic(err)
}

accountRes, unmarshalledRes, err := client.QueryAccount(addr)
if err != nil {
    panic(err)
}
```

An example output would be as follows:

```go
fmt.Println(accountRes)
// Output:
// Account:
//  Address:       g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
//  Pubkey:
//  Coins:         9999862000000ugnot
//  AccountNumber: 0
//  Sequence:      0
```

We are now ready to send a transaction to the chain.

## Sending




