---
id: connect-from-go
---

# How to connect a Go app to Gno.land 

This guide will show you how to connect to a Gno.land network from your Go application,
using the [gnoclient](../reference/gnoclient/gnoclient.md) package.

For this guide, we will build a small Go app that will:

- Get account information from the chain
- Broadcast a state-changing transaction
- Read on-chain state

## Prerequisites
- A local Gno.land keypair generated using
[gnokey](../getting-started/local-setup/working-with-key-pairs.md)

## Setup

To get started, create a new Go project. In a clean directory, run the following:
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
- `Signer` can also be initialized in-memory from a BIP39 mnemonic, using the 
[`SignerFromBip39`](../reference/gnoclient/signer.md#func-signerfrombip39) function.

## Initialize the RPC connection & Client

You can initialize the RPC Client used to connect to the Gno.land network with
the following line:
```go
rpc := rpcclient.NewHTTP("<gno_chain_endpoint>", "")
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

accountRes, _, err := client.QueryAccount(addr)
if err != nil {
    panic(err)
}
```

An example result would be as follows:

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

## Sending a transaction

A Gno.land transaction consists of two main parts:
- A set of base transaction fields, such as a gas price, gas limit, account &
sequence number,
- An array of messages to be executed on the chain.

To construct the base set of transaction fields, we can use the `BaseTxCfg` type:
```go
txCfg := gnoclient.BaseTxCfg{
    GasFee:         "1000000ugnot",                 // gas price
    GasWanted:      1000000,                        // gas limit
    AccountNumber:  accountRes.GetAccountNumber(),  // account ID
    SequenceNumber: accountRes.GetSequence(),       // account nonce
    Memo:           "This is a cool how-to guide!", // transaction memo
}
```

For calling an exported (public) function in a Gno realm, we can use the `MsgCall`
message type. We will use the wrapped ugnot realm for this example, wrapping 
`1000000ugnot` (1 $GNOT) for demonstration purposes.

```go
msg := gnoclient.MsgCall{
    PkgPath:  "gno.land/r/demo/wugnot", // wrapped ugnot realm path
    FuncName: "Deposit",                // function to call
    Args:     nil,                      // arguments in string format
    Send:     "1000000ugnot",           // coins to send along with transaction
}
```

Finally, to actually call the function, we can use `Call`:

```go
res, err := client.Call(txCfg, msg)
if err != nil {
	panic(err)
}
```

Before running your code, make sure your keypair has enough funds to send the 
transaction. 

If everything went well, you've just sent a state-changing transaction to a 
Gno.land chain!


## Reading on-chain state

To read on-chain state, you can use the `QEval()` function. This functionality
allows you to evaluate a query expression on a realm, without having to spend gas.

Let's fetch the balance of wrapped ugnot for our address:
```go
// Evaluate expression
qevalRes, _, err := client.QEval("gno.land/r/demo/wugnot", "BalanceOf(\"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5\")")
if err != nil {
	panic(err)
}
```

Printing out the result should output:
```go
fmt.Println(qevalRes)
// Output:
// (1000000 uint64)
```

To see all functionality the `gnoclient` package provides, see the gnoclient
[reference page](../reference/gnoclient/gnoclient.md).

## Conclusion

Congratulations 🎉

You've just built a small demo app in Go that connects to a Gno.land chain
to query account info, send a transaction, and read on-chain state.

Check out the full example app code [here](https://github.com/leohhhn/connect-gno/blob/master/main.go). 

To see a real-world example CLI tool use `gnoclient`,
check out [gnoblog-cli](https://github.com/gnolang/blog/tree/main/cmd/gnoblog-cli).


