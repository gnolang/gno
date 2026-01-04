# Using `gnoclient`

This guide will show you how to connect to a gno.land network from your Go application,
using the [gnoclient](https://gnolang.github.io/gno/github.com/gnolang/gno/gno.land/pkg/gnoclient.html)
package.

For this guide, we will build a small Go app that will:

- Get account information from the chain
- Broadcast a state-changing transaction
- Read on-chain state with ABCI queries

## Prerequisites

- A local gno.land keypair generated using gnokey

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
to a gno.land chain.

```go
type Client struct {
    Signer    Signer           // Signer for transaction authentication
    RPCClient rpcclient.Client // gnolang/gno/tm2/pkg/bft/rpc/client
}
```

### Signer

The `Signer` provides functionality to sign transactions with a gno.land keypair.
The keypair can be accessed from a local keybase, or it can be generated
in-memory from a BIP39 mnemonic.

:::info
The keybase directory path is set with the `gnokey --home` flag.
:::

### RPCClient

The `RPCCLient` provides connectivity to a gno.land network via HTTP or WebSockets.

## Initialize the Signer

For this example, we will initialize the `Signer` from a local keybase:

```go
package main

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

func main() {
	// Initialize keybase from a directory
	keybase, _ := keys.NewKeyBaseFromDir(gnoenv.HomeDir()) // default keybase path
 
	// Create signer
	signer := gnoclient.SignerFromKeybase{
		Keybase:  keybase,
		Account:  "<keypair_name>",     // Name of your keypair in keybase
		Password: "<keypair_password>", // Password to decrypt your keypair
		ChainID:  "<gno_chainID>",      // id of gno.land chain
	}
}
```

A few things to note:
- You can view keys in your local keybase by running `gnokey list`.
- You can get the password from a user input using the IO package.
- `Signer` can also be initialized in-memory from a BIP39 mnemonic, using the
  [`SignerFromBip39`](https://gnolang.github.io/gno/github.com/gnolang/gno@v0.0.0/gno.land/pkg/gnoclient.html#SignerFromBip39)
  function.

## Initialize the RPC connection & Client

You can initialize the RPC Client used to connect to the gno.land network with
the following line:
```go
rpc, err := rpcclient.NewHTTPClient("<gno.land_remote_endpoint>")
if err != nil {
    panic(err)
}
```

A list of gno.land network endpoints & chain IDs can be found in the
[Gno.land Networks](../../../../docs/resources/gnoland-networks.md) page.

With this, we can initialize the `gnoclient.Client` struct:

```go
package main

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
)

func main() {
	// Initialize keybase from a directory
	keybase, _ := keys.NewKeyBaseFromDir(gnoenv.HomeDir()) // default keybase path

	// Create signer
	signer := gnoclient.SignerFromKeybase{
		Keybase:  keybase,
		Account:  "<keypair_name>",     // Name of your keypair in keybase
		Password: "<keypair_password>", // Password to decrypt your keypair
		ChainID:  "<gno_chainID>",      // id of gno.land chain
	}

	// Initialize the RPC client
	rpc, err := rpcclient.NewHTTPClient("<gno.land_rpc_endpoint>")
	if err != nil {
		panic(err)
	}

	// Initialize the gnoclient
	client := gnoclient.Client{
		Signer:    signer,
		RPCClient: rpc,
	}
}
```

We can now communicate with the gno.land chain. Let's explore some of the functionality
`gnoclient` provides.

## Query account info from a chain

To send transactions to the chain, we need to know the account number (ID) and
sequence (nonce). We can get this information by querying the chain with the
`QueryAccount` function:

```go   
// Getting account info
account, err := client.Signer.Info()
if err != nil {
    panic(err)
}

// Querying an account
address := account.GetAddress()
accountRes, _, err := client.QueryAccount(address)
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

A gno.land transaction consists of two main parts:
- A set of base transaction fields, such as a gas price, gas limit, account &
  sequence number,
- An array of messages to be executed on the chain.

To construct the base set of transaction fields, we can use the `BaseTxCfg` type:
```go
txCfg := gnoclient.BaseTxCfg{
    GasFee:         "1000000ugnot",                 // gas price
    GasWanted:      10000000,                       // gas limit
    AccountNumber:  accountRes.GetAccountNumber(),  // account ID
    SequenceNumber: accountRes.GetSequence(),       // account nonce
    Memo:           "This is a cool how-to guide!", // transaction memo
}
```

For calling an exported (public) function in a Gno realm, we can use the `MsgCall`
message type. We will use the wrapped ugnot realm for this example, wrapping
`1000000ugnot` (1 GNOT) for demonstration purposes.

```go
import (
	...
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/std"
)
```

```go
pkgpath := "gno.land/r/gnoland/wugnot"
msg := vm.MsgCall{
    Caller:  addr,                                                    // address of the caller (signer)
    PkgPath: pkgpath,                                                 // wrapped ugnot realm path
    Func:    "Deposit",                                               // function to call
    Args:    nil,                                                     // arguments in string format
    Send:    std.Coins{{Denom: ugnot.Denom, Amount: int64(1000000)}}, // coins to send along with transaction
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
gno.land chain!

## Reading on-chain state

To read on-chain state, you can use the `QEval()` function. This functionality
allows you to evaluate a query expression on a realm, without having to spend gas.

Let's fetch the balance of wrapped ugnot for our address:
```go
// Evaluate expression
expr := fmt.Sprintf("BalanceOf(\"%s\")", address.String())
qevalRes, _, err := client.QEval(pkgpath, expr)
if err != nil {
    panic(err)
}
```

The result should contain a similar output when printed:
```go
fmt.Println(qevalRes)
// Output:
// (1000000 uint64)
```

To see all functionality the `gnoclient` package provides, see the gnoclient
[gnoclient reference](https://gnolang.github.io/gno/github.com/gnolang/gno/gno.land/pkg/gnoclient.html).

## Full code

```go
package main

import (
	"fmt"
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func main() {
	// Initialize keybase from a directory
	keybase, _ := keys.NewKeyBaseFromDir(gnoenv.HomeDir())

	// Create signer
	signer := gnoclient.SignerFromKeybase{
		Keybase:  keybase,
		Account:  "<keypair_name>",     // Name of your keypair in keybase
		Password: "<keypair_password>", // Password to decrypt your keypair
		ChainID:  "<gno_chainID>",      // id of gno.land chain
	}

	// Initialize the RPC client
	rpc, err := rpcclient.NewHTTPClient("<gno.land_rpc_endpoint>")
	if err != nil {
		panic(err)
	}

	// Initialize the gnoclient
	client := gnoclient.Client{
		Signer:    signer,
		RPCClient: rpc,
	}

	// Get account info
	account, err := client.Signer.Info()
	if err != nil {
		panic(err)
	}

	address := account.GetAddress()
	// Querying an account
	accountRes, _, err := client.QueryAccount(address)
	if err != nil {
		panic(err)
	}

	// Sending a tx
	txCfg := gnoclient.BaseTxCfg{
		GasFee:         "1000000ugnot",                 // gas price
		GasWanted:      10000000,                       // gas limit
		AccountNumber:  accountRes.GetAccountNumber(),  // account ID
		SequenceNumber: accountRes.GetSequence(),       // account nonce
		Memo:           "This is a cool how-to guide!", // transaction memo
	}

	pkgpath := "gno.land/r/gnoland/wugnot"
	msg := vm.MsgCall{
		Caller:  address,                                                 // address of the caller (signer)
		PkgPath: pkgpath,                                                 // wrapped ugnot realm path
		Func:    "Deposit",                                               // function to call
		Args:    nil,                                                     // arguments in string format
		Send:    std.Coins{{Denom: ugnot.Denom, Amount: int64(1000000)}}, // coins to send along with transaction
	}

	fmt.Println("--------------------------- SENDING TX")
	fmt.Printf("Calling %s.Deposit(\"1000000ugnot\")\n", pkgpath)
	res, err := client.Call(txCfg, msg)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Events: ")
	fmt.Println(res.DeliverTx.Events)
	fmt.Println("--------------------------- QUERYING STATE")

	// Using ABCI queries
	expr := fmt.Sprintf("BalanceOf(\"%s\")", address.String())
	qevalRes, _, err := client.QEval(pkgpath, expr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Evaluating %s.%s\n", pkgpath, expr)
	fmt.Printf("Result: %s", qevalRes)
}
```

## Conclusion

Congratulations 🎉

You've just built a small demo app in Go that connects to a gno.land chain
to query account info, send a transaction, and read on-chain state.

To see a real-world example CLI tool use `gnoclient`,
check out [gnoblog-cli](https://github.com/gnolang/blog/tree/main/cmd/gnoblog-cli).