# Interacting with gno.land using gnokey

`gnokey` is the official command-line wallet and utility for interacting with
gno.land networks. It allows you to manage keys, query the blockchain, send
transactions, and deploy smart contracts. This guide will help you get started
with the essential operations.

## Installing gnokey

You can install `gnokey` through various methods:

### Option 1: Install from source

To build and install from source, you'll need:
- Git
- Go 1.22+
- Make

```bash
# Clone the repository
git clone https://github.com/gnolang/gno.git
cd gno

# Install gnokey
make install
```

### Option 2: Download prebuilt binaries

Coming soon.

## Managing key pairs

In this tutorial, you will learn how to create your Gno key pair using
[gnokey](./interact-with-gnokey.md). A key pair is required to send
transactions to the blockchain, including deploying code, interacting with
existing applications, and more.

## A word about key pairs

Key pairs are the foundation of how users interact with blockchains; and Gno is
no exception. By using a 12-word or 24-word [mnemonic phrase](https://www.zimperium.com/glossary/mnemonic-seed/)
as a source of randomness, users can derive a private and a public key.
These two keys can then be used further; a public key derives an address which is
a unique identifier of a user on the blockchain, while a private key is used for
signing messages and transactions for the aforementioned address, proving a user
has ownership over it.

Let's see how we can use `gnokey` to generate a Gno key pair locally.

## Generating a key pair

The `gnokey add` command allows you to generate a new key pair locally. Simply
run the command, while adding a name for your key pair:

```bash
gnokey add MyKey
```

After running the command, `gnokey` will ask you to enter a password that will be
used to encrypt your key pair to the disk. Then, it will show you the following
information:
- Your public key, as well as the Gno address derived from it, starting with `g1`,
- Your randomly generated 12-word mnemonic phrase which was used to derive the key pair.

:::warning Safeguard your mnemonic phrase!

A **mnemonic phrase** is like your master password; you can use it over and over
to derive the same key pairs. This is why it is crucial to store it in a safe,
offline place - writing the phrase on a piece of paper and hiding it is highly
recommended. **If it gets lost, it is unrecoverable.**

:::

`gnokey` will generate a keybase in which it will store information about your
key pairs. The keybase directory path is stored under the `-home` flag in `gnokey`.

### Gno addresses

Your **Gno address** is like your unique identifier on the network; an address
is visible in the caller stack of an application, it is included in each
transaction you create with your key pair, and anyone who knows your address can
send you [coins](../resources/gno-stdlibs.md#coin), etc.

## Making transactions

In Gno, there are four types of messages that can change on-chain state:
- `AddPackage` - adds new code to the chain
- `Call` - calls a specific path and function on the chain
- `Send` - sends coins from one address to another
- `Run` - executes a Gno script against on-chain code

A gno.land transaction contains two main things:
- A base configuration where variables such as `gas-fee`, `gas-wanted`, and others
  are defined
- A list of messages to execute on the chain

Currently, `gnokey` supports single-message transactions, while multiple-message
transactions can be created in Go programs, supported by the
[gnoclient](https://github.com/gnolang/gno/tree/master/gno.land/pkg/gnoclient) package.

We will need some testnet coins (GNOTs) for each state-changing call. Visit the [Faucet
Hub](https://faucet.gno.land) to get GNOTs for the Gno testnets that are currently live.

Let's delve deeper into each of these message types.

## `AddPackage`

In case you want to upload new code to the chain, you can use the `AddPackage`
message type. You can send an `AddPackage` transaction with `gnokey` using the
following command:

```bash
gnokey maketx addpkg
```

To understand how to use this subcommand better, let's write a simple "Hello world"
[pure package](../resources/gno-packages.md). First, let's create a folder which will
store our example code.

```bash
└── example/
```

Then, let's create a `hello_world.gno` file under the `p/` folder:

```bash
cd example
mkdir p/ && cd p
touch hello_world.gno
```

Now, we should have the following folder structure:

```bash
└── example/
│   └── p/
│       └── hello_world.gno
```

In the `hello_world.gno` file, add the following code:

```go
package hello_world

func Hello() string {
  return "Hello, world!"
}
```

We are now ready to upload this package to the chain. To do this, we must set the
correct flags for the `addpkg` subcommand.

The `addpkg` subcommmand uses the following flags and arguments:
- `-pkgpath` - on-chain path where your code will be uploaded to
- `-pkgdir` - local path where your is located
- `-broadcast` - enables broadcasting the transaction to the chain
- `-deposit` - a deposit amount of GNOT to send along with the transaction
- `-gas-wanted` - the upper limit for units of gas for the execution of the
  transaction
- `-gas-fee` - amount of GNOTs to pay per gas unit
- `-chain-id` - id of the chain that we are sending the transaction to
- `-remote` - specifies the remote node RPC listener address

The `-pkgpath`, `-pkgdir`, and `-deposit` flags are unique to the `addpkg`
subcommand, while `-broadcast`, `-gas-wanted`, `-gas-fee`, `-chain-id`, and
`-remote` are used for setting the base transaction configuration. These flags
will be repeated throughout the tutorial.

Next, let's configure the `addpkg` subcommand to publish this package to the
[Portal Loop](../resources/gnoland-networks.md) testnet. Assuming we are in
the `example/p/` folder, the command will look like this:

```bash
gnokey maketx addpkg \
-pkgpath "gno.land/p/<your_namespace>/hello_world" \
-pkgdir "." \
-deposit "" \
-gas-fee 10000000ugnot \
-gas-wanted 8000000 \
-broadcast \
-chainid portal-loop \
-remote "https://rpc.gno.land:443"
```

Once we have added a desired [namespace](../resources/users-and-teams.md) to upload the package to, we can specify a key pair name to use to execute the
transaction:

```bash
gnokey maketx addpkg \
-pkgpath "gno.land/p/examplenamespace/hello_world" \
-pkgdir "." \
-send "" \
-gas-fee 10000000ugnot \
-gas-wanted 200000 \
-broadcast \
-chainid portal-loop \
-remote "https://rpc.gno.land:443"
mykey
```

If the transaction was successful, you will get an output from `gnokey` that is
similar to the following:

```console
OK!
GAS WANTED: 200000
GAS USED:   117564
HEIGHT:     3990
EVENTS:     []
TX HASH:    Ni8Oq5dP0leoT/IRkKUKT18iTv8KLL3bH8OFZiV79kM=
```

Let's analyze the output, which is standard for any `gnokey` transaction:
- `GAS WANTED: 200000` - the original amount of gas specified for the transaction
- `GAS USED:   117564` - the gas used to execute the transaction
- `HEIGHT:     3990` - the block number at which the transaction was executed at
- `EVENTS:     []` - [Gno events](../resources/gno-stdlibs.md#events) emitted by the transaction, in this case, none
- `TX HASH:    Ni8Oq5dP0leoT/IRkKUKT18iTv8KLL3bH8OFZiV79kM=` - the hash of the transaction

Congratulations! You have just uploaded a pure package to the Portal Loop network.
If you wish to deploy to a different network, find the list of all network
configurations in the [Network Configuration](../resources/gnoland-networks.md) section.

## `Call`

The `Call` message type is used to call any exported realm function.
You can send a `Call` transaction with `gnokey` using the following command:

```bash
gnokey maketx call
```

:::info `Call` uses gas

Using `Call` to call an exported function will use up gas, even if the function
does not modify on-chain state. If you are calling such a function, you can use
the `query` functionality for a read-only call which
does not use gas.

:::

For this example, we will call the `wugnot` realm, which wraps GNOTs to a
GRC20-compatible token called `wugnot`. We can find this realm deployed on the
[Portal Loop](../resources/gnoland-networks.md) testnet, under the `gno.land/r/demo/wugnot` path.

We will wrap `1000ugnot` into the equivalent in `wugnot`. To do this, we can call
the `Deposit()` function found in the `wugnot` realm. As previously, we will
configure the `maketx call` subcommand:

```bash
gnokey maketx call \
-pkgpath "gno.land/r/demo/wugnot" \
-func "Deposit" \
-send "1000ugnot" \
-gas-fee 10000000ugnot \
-gas-wanted 2000000 \
-broadcast \
-chainid portal-loop \
-remote "https://rpc.gno.land:443" \
mykey
```

In this command, we have specified three main things:
- The path where the realm lives on-chain with the `-pkgpath` flag
- The function that we want to call on the realm with the `-func` flag
- The amount of `ugnot` we want to send to be wrapped, using the `-send` flag

Apart from this, we have also specified the Portal Loop chain ID, `portal-loop`,
as well as the Portal Loop remote address, `https://rpc.gno.land:443`.

After running the command, we can expect an output similar to the following:
```bash
OK!
GAS WANTED: 2000000
GAS USED:   489528
HEIGHT:     24142
EVENTS:     [{"type":"Transfer","attrs":[{"key":"from","value":""},{"key":"to","value":"g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"},{"key":"value","value":"1000"}],"pkg_path":"gno.land/r/demo/wugnot","func":"Mint"}]
TX HASH:    Ni8Oq5dP0leoT/IRkKUKT18iTv8KLL3bH8OFZiV79kM=
```

In this case, we can see that the `Deposit()` function emitted an
[event](../resources/gno-stdlibs.md#events) that tells us more about what
happened during the transaction.

After broadcasting the transaction, we can verify that we have the amount of `wugnot` we expect. We
can call the `BalanceOf()` function in the same realm:

```bash
gnokey maketx call \
-pkgpath "gno.land/r/demo/wugnot" \
-func "BalanceOf" \
-args "<your_address>" \
-gas-fee 10000000ugnot \
-gas-wanted 2000000 \
-broadcast \
-chainid portal-loop \
-remote "https://rpc.gno.land:443" \
mykey
```

If everything was successful, we should get something similar to the following
output:

```
(1000 uint64)

OK!
GAS WANTED: 2000000
GAS USED:   396457
HEIGHT:     64839
EVENTS:     []
TX HASH:    gQP9fJYrZMTK3GgRiio3/V35smzg/jJ62q7t4TLpdV4=
```

At the top, you will see the output of the transaction, specifying the value and
type of the return argument.

In this case, we used `maketx call` to call a read-only function, which simply
checks the `wugnot` balance of a specific address. This is discouraged, as
`maketx call` actually uses gas. To call a read-only function without spending gas,
check out the `vm/qeval` query section.

## `Send`

We can use the `Send` message type to access the TM2 [Banker](../resources/gno-stdlibs.md#banker)
directly and transfer coins from one Gno address to another.

Coins, such as GNOTs, are always formatted in the following way:

```
<amount><denom>
100ugnot
```

For this example, let's transfer some GNOTs. Just like before, we can configure
our `maketx send` subcommand:
```bash
gnokey maketx send \
-to g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 \
-send 100ugnot \
-gas-fee 10000000ugnot \
-gas-wanted 2000000 \
-broadcast \
-chainid portal-loop \
-remote "https://rpc.gno.land:443" \
mykey
```

Here, we have set the `-to` & `-send` flags to match the recipient, in this case
the publicly-known `test1` address, and `100ugnot` for the coins we want to send,
respectively.

To check the balance of a specific address, check out the `bank/balances` query
in the [Querying a network](querying-a-network.md#bankbalances) section.

## `Run`

With the `Run` message, you can write a snippet of Gno code and run it against
code on the chain. For this example, we will use the [Userbook realm](https://gno.land/r/demo/userbook),
which simply allows you to register the fact that you have interacted with it.
It contains a simple `SignUp()` function, which we will call with `Run`.

To understand how to use the `Run` message better, let's write a simple `script.gno`
file. First, create a folder which will store our script.

```bash
└── example/
```

Then, let's create a `script.gno` file:

```bash
cd example
touch script.gno
```

Now, we should have the following folder structure:

```bash
└── example/
│   └── script.gno
```

In the `script.gno` file, first define the package to be `main`. Then we can import
the Userbook realm and define a `main()` function with no return values that will
be automatically detected and run. In it, we can call the `SignUp()` function.

```go
package main

import "gno.land/r/demo/userbook"

func main() {
  println(userbook.SignUp())
}
```

Now we will be able to provide this to the `maketx run` subcommand:
```bash
gnokey maketx run \
-gas-fee 1000000ugnot \
-gas-wanted 20000000 \
-broadcast \
-chainid portal-loop \
-remote "https://rpc.gno.land:443" \
mykey ./script.gno
```

After running this command, the chain will execute the script and apply any state
changes. Additionally, by using `println`, which is only available in the `Run`
& testing context, we will be able to see the return value of the function called.

### The power of `Run`

Specifically, the above example could have been replaced with a simple `maketx call`
call. The full potential of run comes out in three specific cases:
1. Calling realm functions multiple times in a loop
2. Calling functions with non-primitive input arguments
3. Calling methods on exported variables

Let's look at each of these cases in detail. To demonstrate, we'll make a call
to the following example realm:

```go
package foo

import "gno.land/p/demo/ufmt"

var (
	MainFoo *Foo
	foos    []*Foo
)

type Foo struct {
	bar string
	baz int
}

func init() {
	MainFoo = &Foo{bar: "mainBar", baz: 0}
}

func (f *Foo) String() string {
	return ufmt.Sprintf("Foo - (bar: %s) - (baz: %d)\n\n", f.bar, f.baz)
}

func NewFoo(bar string, baz int) *Foo {
	return &Foo{bar: bar, baz: baz}
}

func AddFoos(multipleFoos []*Foo) {
	foos = append(foos, multipleFoos...)
}

func Render(_ string) string {
	var output string

	for _, f := range foos {
		output += f.String()
	}

	return output
}
```

This realm is deployed to [`gno.land/r/docs/examples/run/foo`](https://gno.land/r/docs/examples/run/foo/package.gno)
on the Portal Loop testnet.

1. Calling realm functions multiple times in a loop:
```go
package main

import (
  "gno.land/r/docs/examples/run/foo"
)

func main() {
  for i := 0; i < 5; i++ {
    println(foo.Render(""))
  }
}
```

2. Calling functions with non-primitive input arguments:

Currently, `Call` only supports primitives for arguments. With `Run`, these
limitations are removed; we can execute a function that takes in a struct, array,
or even an array of structs.

We are unable to call `AddFoos()` with the `Call` message type, while with `Run`,
we can:

```go
package main

import (
  "strconv"

  "gno.land/r/docs/examples/run/foo"
)

func main() {
  var multipleFoos []*foo.Foo

  for i := 0; i < 5; i++ {
    newFoo := foo.NewFoo(
      "bar"+strconv.Itoa(i),
      i,
    )

    multipleFoos = append(multipleFoos, newFoo)
  }

  foo.AddFoos(multipleFoos)
}

```

3. Calling methods on exported variables:

```go
package main

import "gno.land/r/docs/examples/run/foo"

func main() {
	println(foo.MainFoo.String())
}
```

Finally, we can call methods that are on top-level objects in case they exist,
which is not currently possible with the `Call` message.

## Making an airgapped transaction

`gnokey` provides a way to create a transaction, sign it, and later
broadcast it to a chain in the most secure fashion. This approach, while more
complicated than the standard approach shown [in a previous tutorial](making-transactions.md),
grants full control and provides [airgap](https://en.wikipedia.org/wiki/Air_gap_(networking))
support.

By separating the signing and the broadcasting steps of submitting a transaction,
users can make sure that the signing happens in a secure, offline environment,
keeping private keys away from possible exposure to attacks coming from the
internet.

The intended purpose of this functionality is to provide maximum security when
signing and broadcasting a transaction. In practice, this procedure should take
place on two separate machines controlled by the holder of the keys, one with
access to the internet (`Machine A`), and the other one without (`Machine B`),
with the separation of steps as follows:
1. `Machine A`: Fetch account information from the chain
2. `Machine B`: Create an unsigned transaction locally
3. `Machine B`: Sign the transaction
4. `Machine A`: Broadcast the transaction

## 1. Fetching account information from the chain

First, we need to fetch data for the account we are using to sign the transaction,
using the [auth/accounts](querying-a-network.md#authaccounts) query:

```bash
gnokey query auth/accounts/<your_address> -remote "https://rpc.gno.land:443"
```

We need to extract the account number and sequence from the output:

```bash
height: 0
data: {
  "BaseAccount": {
    "address": "g1zzqd6phlfx0a809vhmykg5c6m44ap9756s7cjj",
    "coins": "10000000ugnot",
    "public_key": null,
    "account_number": "468",
    "sequence": "0"
  }
}
```

In this case, the account number is `468`, and the sequence (nonce) is `0`. We
will need these values to sign the transaction later. These pieces of information
are crucial during the signing process, as they are included in the signature
of the transaction, preventing replay attacks.

## 2. Creating an unsigned transaction locally

To create the transaction you want, you can use the [`call` API](making-transactions.md#call),
without the `-broadcast` flag, while redirecting the output to a local file:

```bash
gnokey maketx call \
-pkgpath "gno.land/r/demo/userbook" \
-func "SignUp" \
-gas-fee 1000000ugnot \
-gas-wanted 2000000 \
mykey > userbook.tx
```

This will create a `userbook.tx` file with a null `signature` field.
Now we are ready to sign the transaction.

## 3. Signing the transaction

To add a signature to the transaction, we can use the `gnokey sign` subcommand.
To sign, we must set the correct flags for the subcommand:
- `-tx-path` - path to the transaction file to sign, in our case, `userbook.tx`
- `-chainid` - id of the chain to sign for
- `-account-number` - number of the account fetched previously
- `-account-sequence` - sequence of the account fetched previously

```bash
gnokey sign \
-tx-path userbook.tx \
-chainid "portal-loop" \
-account-number 468 \
-account-sequence 0 \
mykey
```

After inputting the correct values, `gnokey` will ask for the password to decrypt
the key pair. Once we input the password, we should receive the message that the
signing was completed. If we open the `userbook.tx` file, we will be able to see
that the signature field has been populated.

We are now ready to broadcast this transaction to the chain.

## 4. Broadcasting the transaction

To broadcast the signed transaction to the chain, we can use the `gnokey broadcast`
subcommand, giving it the path to the signed transaction:

```bash
gnokey broadcast -remote "https://rpc.gno.land:443" userbook.tx
```

In this case, we do not need to specify a key pair, as the transaction has already
been signed in a previous step and `gnokey` is only sending it to the RPC endpoint.

## Verifying a transaction's signature

To verify a transaction's signature is correct, you can use the `gnokey verify`
subcommand. We can provide the path to the transaction document using the `-docpath`
flag, provide the key we signed the transaction with, and the signature itself.
Make sure the signature is in the `hex` format.

```bash
gnokey verify -docpath userbook.tx mykey <signature>
```

# Querying a gno.land network

gno.land and `gnokey` support ABCI queries. Using ABCI queries, you can query the state of
a gno.land network without spending any gas. All queries need to be pointed towards
a specific remote address from which the state will be retrieved.

To send ABCI queries, you can use the `gnokey query` subcommand, and provide it
with the appropriate query. The `query` subcommand allows us to send different
types of queries to a gno.land network.

Below is a list of queries a user can make with `gnokey`:
- `auth/accounts/{ADDRESS}` - returns information about an account
- `bank/balances/{ADDRESS}` - returns balances of an account
- `vm/qfuncs` - returns the exported functions for a given pkgpath
- `vm/qfile` - returns package contents for a given pkgpath
- `vm/qdoc` - Returns the JSON of the doc for a given pkgpath, suitable for printing
- `vm/qeval` - evaluates an expression in read-only mode on and returns the results
- `vm/qrender` - shorthand for evaluating `vm/qeval Render("")` for a given pkgpath

Let's see how we can use them.

## `auth/accounts`

We can obtain information about a specific address using this subquery. To call it,
we can run the following command:

```bash
gnokey query auth/accounts/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 -remote https://rpc.gno.land:443
```

With this, we are asking the Portal Loop network to deliver information about the
specified address. If everything went correctly, we should get output similar to the following:

```bash
height: 0
data: {
  "BaseAccount": {
    "address": "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
    "coins": "227984898927ugnot",
    "public_key": {
      "@type": "/tm.PubKeySecp256k1",
      "value": "A+FhNtsXHjLfSJk1lB8FbiL4mGPjc50Kt81J7EKDnJ2y"
    },
    "account_number": "0",
    "sequence": "12"
  }
}
```

The return data will contain the following fields:
- `height` - the height at which the query was executed. This is currently not
  supported and is `0` by default.
- `data` - contains the result of the query.

The `data` field returns a `BaseAccount`, which is the main struct used in Tendermint2
to hold account data. It contains the following information:
- `address` - the address of the account
- `coins` - the list of coins the account owns
- `public_key` - the TM2 public key of the account, from which the address is derived
- `account_number` - a unique identifier for the account on the gno.land chain
- `sequence` - a nonce, used for protection against replay attacks

## `bank/balances`

With this query, we can fetch [coin](../resources/gno-stdlibs.md#coin) balances
of a specific account. To call it, we can run the following command:

```bash
gnokey query bank/balances/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 -remote https://rpc.gno.land:443
```

If everything went correctly, we should get an output similar to the following:

```bash
height: 0
data: "227984898927ugnot"
```

The data field will contain the coins the address owns.

## `vm/qfuncs`

Using the `vm/qfuncs` query, we can fetch exported functions from a specific package
path. To specify the path we want to query, we can use the `-data` flag:

```bash
gnokey query vm/qfuncs --data "gno.land/r/demo/wugnot" -remote https://rpc.gno.land:443
```

The output is a string containing all exported functions for the `wugnot` realm:

```json
height: 0
data: [
        {
          "FuncName": "Deposit",
          "Params": null,
          "Results": null
        },
        {
          "FuncName": "Withdraw",
          "Params": [
            {
            "Name": "amount",
            "Type": "uint64",
            "Value": ""
            }
          ],
          "Results": null
        },
        // other functions
]
```

## `vm/qfile`

With the `vm/qfile` query, we can fetch files and their content found on a
specific package path. To specify the path we want to query, we can use the
`-data` flag:

```bash
gnokey query vm/qfile -data "gno.land/r/demo/wugnot" -remote https://rpc.gno.land:443
```

If the `-data` field contains only the package path, the output is a list of all
files found within the `wugnot` realm:

```bash
height: 0
data: gno.mod
wugnot.gno
z0_filetest.gno
```

If the `-data` field also specifies a file name after the path, the source code
of the file will be retrieved:

```bash
gnokey query vm/qfile -data "gno.land/r/demo/wugnot/wugnot.gno" -remote https://rpc.gno.land:443
```

Output:
```bash
height: 0
data: package wugnot

import (
        "std"
        "strings"

        "gno.land/p/demo/grc/grc20"
        "gno.land/p/demo/ufmt"
        pusers "gno.land/p/demo/users"
        "gno.land/r/demo/users"
)

var (
        banker *grc20.Banker = grc20.NewBanker("wrapped GNOT", "wugnot", 0)
        Token                = banker.Token()
)

const (
        ugnotMinDeposit  uint64 = 1000
        wugnotMinDeposit uint64 = 1
)
...
```

## `vm/qdoc`

Using the `vm/qdoc` query, we can fetch the docs, for functions, types and variables from a specific
package path. To specify the path we want to query, we can use the `-data` flag:

```bash
gnokey query vm/qdoc --data "gno.land/r/gnoland/valopers/v2" -remote https://rpc.gno.land:443
```

The output is a JSON string containing doc strings of the package, functions, etc., including comments for `valopers` realm:

```json
height: 0
data: {
  "package_path": "gno.land/r/gnoland/valopers/v2",
  "package_line": "package valopers // import \"valopers\"",
  "package_doc": "Package valopers is designed around the permissionless lifecycle of valoper profiles. It also includes parts designed for govdao to propose valset changes based on registered valopers.\n",
  "values": [
    {
      "name": "valopers",
      "doc": "// Address -> Valoper\n",
      "type": "*avl.Tree"
    }
    // other values
  ],
  "funcs": [
    {
      "type": "",
      "name": "GetByAddr",
      "signature": "func GetByAddr(address std.Address) Valoper",
      "doc": "GetByAddr fetches the valoper using the address, if present\n",
      "params": [
        {
          "Name": "address",
          "Type": "std.Address"
        }
      ],
      "results": [
        {
          "Name": "",
          "Type": "Valoper"
        }
      ]
    }
    // other funcs
    {
      "type": "Valoper",
      "name": "Render",
      "signature": "func (v Valoper) Render() string",
      "doc": "Render renders a single valoper with their information\n",
      "params": [],
      "results": [
        {
          "Name": "",
          "Type": "string"
        }
      ]
    }
    // other methods (in this case of the Valoper type)
  ],
  "types": [
    {
      "name": "Valoper",
      "signature": "type Valoper struct {\n\tName        string // the display name of the valoper\n\tMoniker     string // the moniker of the valoper\n\tDescription string // the description of the valoper\n\n\tAddress      std.Address // The bech32 gno address of the validator\n\tPubKey       string      // the bech32 public key of the validator\n\tP2PAddresses []string    // the publicly reachable P2P addresses of the validator\n\tActive       bool        // flag indicating if the valoper is active\n}",
      "doc": "Valoper represents a validator operator profile\n"
    }
  ]
}
```

## `vm/qeval`

`vm/qeval` allows us to evaluate a call to an exported function without using gas,
in read-only mode. For example:

```bash
gnokey query vm/qeval -remote https://rpc.gno.land:443 -data "gno.land/r/demo/wugnot.BalanceOf(\"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5\")"
```

This command will return the `wugnot` balance of the above address without using gas.
Properly escaping quotation marks for string arguments is currently required.

Currently, `vm/qeval` only supports primitive types in expressions.

## `vm/qrender`

`vm/qrender` is an alias for executing `vm/qeval` on the `Render("")` function.
We can use it like this:

```bash
gnokey query vm/qrender --data "gno.land/r/demo/wugnot:" -remote https://rpc.gno.land:443
```

Running this command will display the current `Render()` output of the WUGNOT
realm, which is also displayed by default on the [realm's page](https://gno.land/r/demo/wugnot):

```bash
height: 0
data: # wrapped GNOT ($wugnot)

* **Decimals**: 0
* **Total supply**: 5012404
* **Known accounts**: 2
```

:::info Specifying a path to `Render()`

To call the `vm/qrender` query with a specific path, use the `<pkgpath>:<renderpath>` syntax.
For example, the `wugnot` realm provides a way to display the balance of a specific
address in its `Render()` function. We can fetch the balance of an account by
providing the following custom pattern to the `wugnot` realm:

```bash
gnokey query vm/qrender --data "gno.land/r/demo/wugnot:balance/g125em6arxsnj49vx35f0n0z34putv5ty3376fg5" -remote https://rpc.gno.land:443
```

To see how this was achieved, check out `wugnot`'s `Render()` function.
:::

### Gas parameters

When using `gnokey` to send transactions, you'll need to specify gas parameters:

```bash
gnokey maketx call \
  --pkgpath "gno.land/r/demo/boards" \
  --func "CreateBoard" \
  --args "MyBoard" "Board description" \
  --gas-fee 1000000ugnot \
  --gas-wanted 2000000 \
  --remote https://rpc.gno.land:443 \
  --chainid portal-loop \
  YOUR_KEY_NAME
```

For detailed information about gas fees, including recommended values and optimization strategies, see the [Gas Fees documentation](../resources/gas-fees.md).

## Conclusion

That's it! 🎉

In this tutorial, you've learned to use `gnokey` to query a gno.land
network.

