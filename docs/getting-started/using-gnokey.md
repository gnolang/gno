---
id: using-gnokey
---

# Using `gnokey`

## Overview
In this tutorial, you will learn how to use the `gnokey` binary to interact with
a Gno.land chain. You will learn how to create state-changing calls, run readonly
queries without using gas, as well as create, sign, and broadcast airgapped
transactions for full security.

## Prerequisites
- **`gno`, `gnokey`, and `gnodev` installed.** Reference the
  [Local Setup](local-setup/installation.md#2-installing-the-required-tools-) guide for steps
- **A Gno.land keypair set up.** Reference the
  [Working with Key Pairs](local-setup/working-with-key-pairs.md) guide for steps

## Interacting with a Gno.land chain

`gnokey` allows you to interact with any Gno.land network, such as the
[Portal Loop](../concepts/portal-loop.md) testnet.

There are multiple ways anyone can interact with the chain:
- Transactions - state-changing calls which use gas
- ABCI queries - read-only calls which do not use gas

Both transactions and ABCI queries can be made via `gnokey`'s subcommands,
`maketx` and `query`.

## State-changing calls (transactions)

In Gno, there are four types of messages that can change on-chain state:
- `AddPackage` - adds new code to the chain
- `Call` - calls a specific path and function on the chain
- `Send` - sends coins from one address to another
- `Run` - executes a Gno script against on-chain code

A Gno.land transaction contains two main things:
- A base configuration where variables such as `gas-fee`, `gas-wanted`, and others
  are defined
- A list of messages to execute on the chain

Currently, `gnokey` supports single-message transactions, while multiple-message
transactions can be created in Go programs, supported by the
[gnoclient](../reference/gnoclient/gnoclient.md) package.

We will need some testnet coins (GNOTs) for each state-changing call. Visit the [Faucet
Hub](https://faucet.gno.land) to get GNOTs for the Gno testnets that are currently live.

Let's delve deeper into each of these message types.

### `AddPackage`

In case you want to upload new code to the chain, you can use the `AddPackage`
message type. You can send an `AddPackage` transaction with `gnokey` using the
following command:

```bash
gnokey maketx addpkg
```

To understand how to use this subcommand better, let's write a simple "Hello world"
[pure package](../concepts/packages.md). First, let's create a folder which will
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
- `-send` - a deposit amount of GNOT to send along with the transaction
- `-gas-wanted` - the upper limit for units of gas for the execution of the
  transaction
- `-gas-fee` - amount of GNOTs to pay per gas unit
- `-chain-id` - id of the chain to connect to
- `-remote` - specifies the remote node RPC listener address

The `-pkgpath` and `-pkgdir` flags are unique to the `addpkg` subcommand, while
`-broadcast`,`-send`, `-gas-wanted`, `-gas-fee`, `-chain-id`, and `-remote` are
used for setting the base transaction configuration. These flags will be repeated
throughout the tutorial.

For this specific demonstration, we will run a local Gno node using `gnodev`.
First, simply start `gnodev`:

```bash
gnodev
```

If everything went well, you should see the following output:
```bash
❯ gnodev
Accounts    ┃ I default address imported name=test1 addr=g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
Node        ┃ I pkgs loaded path="[{<your_monorepo_path> g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 }]"
Node        ┃ I node started lisn=tcp://127.0.0.1:26657 chainID=dev
GnoWeb      ┃ I gnoweb started lisn=http://127.0.0.1:8888
-- READY   ┃ I for commands and help, press `h`
```

Now we have a local Gno node listening on `127.0.0.1:26657` with chain ID `dev`,
which we can use to upload our code to.

Next, let's configure the `addpkg` subcommand. Assuming we are in the `example/p`
folder, the command will look like this:

```bash
gnokey maketx addpkg \                                                                                                                                                                                          
-pkgpath "gno.land/p/<your_namespace>/hello_world" \
-pkgdir "." \
-send "" \
-gas-fee 10000000ugnot \
-gas-wanted 8000000 \
-broadcast \
-chainid dev \
-remote "127.0.0.1:26657" \
```

Once we have added a desired namespace to upload the package to, we can specify
a keypair name to use to execute the transaction:

```bash
gnokey maketx addpkg \                                                                                                                                                                                          
-pkgpath "gno.land/p/leon/hello_world" \
-pkgdir "." \
-send "" \
-gas-fee 10000000ugnot \
-gas-wanted 200000 \
-broadcast \
-chainid dev \
-remote "127.0.0.1:26657" \
mykey
```

If the transaction was successful, you will get the following output from `gnokey`:

```
OK!
GAS WANTED: 200000
GAS USED:   117564
HEIGHT:     3990
EVENTS:     []
```

Let's analyze the output, which is standard for any `gnokey` transaction:
- `GAS WANTED: 200000` - the original amount of gas specified for the transaction
- `GAS USED:   117564` - the gas used to execute the transaction
- `HEIGHT:     3990` - the block number at which the transaction was executed at
- `EVENTS:     []` - events emitted by the transaction, in this case, none

Congratulations! You have just uploaded a pure package to your local chain.
If you wish to upload the package to a remote testnet, make sure to switch out
the `-chainid` & `-remote` values for the ones matching your desired testnet.
Find a list of all networks in the [Network Configuration](../reference/network-config.md)
section.

### `Call`

The `Call` message type is used to call any exported function on the chain.
You can send a `Call` transaction with `gnokey` using the following command:

```bash
gnokey maketx call
```

:::info `Call` uses gas

Using `Call` to call an exported function will use up gas, even if the function
does not modify on-chain state. If you are calling such a function, you can use
the [`query` functionality](#query) for a read-only call which does not use gas.

:::

For this example, we will call the `wugnot` realm, which wraps GNOTs to a
GRC20-compatible token called `wugnot`. We can find this realm deployed on the
[Portal Loop](../concepts/portal-loop.md) testnet, under the `gno.land/r/demo/wugnot`.

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
-remote "https://rpc.gno.land:443"" \
mykey
```

In this command, we have specified three main things:
- The path where the realm lives on-chain with the `-pkgpath` flag
- The function that we want to call on the realm with the `-func` flag
- The amount of `ugnot` we want to send to be wrapped, using the `-send` flag

Apart from this, we have also specified the Portal Loop chain ID, `portal-loop`,
as well as the Portal Loop remote address, `https://rpc.gno.land:443`.

To check if we actually have the `wugnot` amount that we wanted to receive, we
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

If everything was successful, we should get the following output:

```
(1000 uint64)

OK!
GAS WANTED: 2000000
GAS USED:   396457
HEIGHT:     64839
EVENTS:     []
```

At the top, you will see the output of the transaction, specifying the value and
type of the return argument.

In this case, we used `maketx call` to call a read-only function, which simply
checks the `wugnot` balance of a specific address. This is discouraged, as
`maketx call` actually uses gas. To call a read-only function without spending gas,
check out the `vm/qeval` query in the [ABCI queries section](#vmqeval).

### `Send`

We can use the `Send` message type to access the TM2 [Banker](../concepts/stdlibs/banker.md)
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
in the [ABCI queries section](#bankbalances).

### `Run`

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

In the `script.gno` file, first define the package to be `main`. Then, we cam import
the Userbook realm and define a `main()` function with no return values which will
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

#### The power of `Run`

Specifically, the above example could have been replaced with a simple `maketx call`
call. The full potential of run comes out in three specific cases:
1. Calling realm functions multiple times in a loop
2. Calling functions with non-primitive input arguments
3. Calling functions with receiver objects

Let's look at each of these cases in detail. To demonstrate, lets use the
following example realm which we will call:

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
	output := ""

	for _, f := range foos {
		output += f.String()
	}

	return output
}
```

This realm is deployed to [`gno.land/r/leon/run/examples/foo`](https://gno.land/r/leon/run/examples/foo)
on the Portal Loop testnet.

1. Calling realm functions multiple times in a loop:
```go
package main

import (
  "gno.land/r/leon/run/examples/foo"
)

func main() {
  for i := 0; i < 5; i++ {
    println(foo.Render(""))
  }
}
```

2. Calling functions with non-primitive input arguments:

Currently, `Call` only supports primitives for arguments. With `Run` these
limitations are removed - we can execute a function that takes in a struct, array,
or even an array of structs.

We are unable to call `AddFoos()` with the `Call` message type, while with `Run`,
we can:

```go
package main

import (
  "gno.land/r/leon/run/examples/foo"
  "strconv"
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

3. Calling functions with receiver objects:

```go
package main

import "gno.land/r/leon/run/examples/foo"

func main() {
	println(foo.MainFoo.String())
}
```

Finally, we can call functions that are on top-level objects, which is not possible
with the `Call` message.

## ABCI queries

Using ABCI queries you can query the state of the chain without spending any gas.
All queries need to be pointed towards a specific remote address from which
the state will be retrieved.

To send ABCI queries, you can use the `gnokey query`
subcommand, and provide it with the appropriate query.
The `query` subcommand allows us to send different types of queries to a Gno.land
network.

Below is a list of queries a user can make with `gnokey`:
- `auth/accounts/{ADDRESS}` - returns information about an account
- `bank/balances/{ADDRESS}` - returns balances of an account
- `vm/qfuncs` - returns the exported functions for a given pkgpath
- `vm/qfile` - returns the list of files for a given pkgpath
- `vm/qeval` - evaluates an expression in read-only mode on and returns the results
- `vm/qrender` - shorthand for evaluating `vm/qeval Render("")` for a given pkgpath

Let's see how we can use them.

### `auth/accounts`

We can obtain information about a specific address using this subquery. To call it,
we can run the following command:

```bash
gnokey query auth/accounts/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 -remote https://rpc.gno.land:443
```

With this, we are asking the Portal Loop network to deliver information about the
specified address. If everything went correctly, we should get the following
output:

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

The `data` field returns a `BaseAccount`, which is the main struct used in TM2 to
hold account data. It contains the following information:
- `address` - the address of the account
- `coins` - the list of coins the account owns
- `public_key` - the TM2 public key of the account, which the address is derived from
- `account_number` - a unique identifier for the account on the Gno.land chain
- `sequence` - a nonce, used for protection against replay attacks

### `bank/balances`

With this query, we can fetch balances of a specific account. To call it, we can
run the following command:

```bash
gnokey query bank/balances/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 -remote https://rpc.gno.land:443
```

If everything went correctly, we should get the following
output:

```bash
height: 0
data: "227984898927ugnot"
```

The data field will contain the coins the address owns.

### `vm/qfuncs`

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

### `vm/qfile`

With the `vm/qfile` query, we can fetch files found on a specific package path.
To specify the path we want to query, we can use the `-data` flag:

```bash
gnokey query vm/qfile -data "gno.land/r/demo/wugnot" -remote https://rpc.gno.land:443
```

The output is a string containing all exported functions for the
`wugnot` realm:

```bash
height: 0
data: gno.mod
wugnot.gno
z0_filetest.gno
```

### `vm/qeval`

`vm/qeval` allows us to evaluate a call to an exported function without using gas,
in read-only mode. For example:

```bash
gnokey query vm/qeval -remote https://rpc.gno.land:443 -data "gno.land/r/demo/wugnot
BalanceOf(\"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5\")" 
```

This command will return the `wugnot` balance of the above address without using gas.
Properly escaping quotation marks, and inputting a new line for the function
is currently required.

### `vm/qrender`

`vm/qrender` is an alias for executing `vm/qeval` on the `Render("")` function.
We can use it like this:

```bash
gnokey query vm/qrender --data "gno.land/r/demo/userbook
" -remote https://rpc.gno.land:443
```

Running this command will display the current `Render()` output of the Userbook
realm, which is also displayed by default on the [realm's page](https://gno.land/r/demo/userbook):

```bash
height: 0
data: # Welcome to UserBook!

## UserBook - Page #1:

#### User #0 - g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 - signed up at Block #0
#### User #1 - g125em6arxsnj49vx35f0n0z34putv5ty3376fg5 - signed up at Block #0
#### User #2 - g1urt7pdmwg2m6z3rsgu4e8peppm4027fvpwkmj8 - signed up at Block #0
#### User #3 - g1uf8u5jf2m9l80g0zsfq7tufl3qufqc4393jtkl - signed up at Block #0
#### User #4 - g1lafcru2z2qelxr33gm4znqshmpur6l9sl3g2aw - signed up at Block #0
---

#### Total users: 5
#### Latest signup: User #4 at Block #0
---

You're viewing page #1
```

## Making an airgapped transaction

`gnokey` provides a way to create a transaction, sign it, and later
broadcast it to a chain in an airgapped manner. With this approach, while it is
more complicated, users can get full control over the creation, signing and
broadcasting process of transactions.

Here are the steps taken in this process:
1. Fetching account information from the chain
2. Creating an unsigned transaction locally
3. Signing the transaction
4. Broadcasting the transaction

For this example, we will again use the Userbook realm on the Portal Loop testnet.

### Fetching account information from the chain

First, we need to fetch data for the account we are using to sign the transaction,
using the [auth/accounts](#authaccounts) query:

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
will need these values to sign the transaction later.

### Creating an unsigned transaction locally

To create the transaction you want, you can use the aforementioned `call` API,
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

### Signing the transaction

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
the keypair. Once we input the password, we should receive the message that the
signing was completed. If we open the `userbook.tx` file, we will be able to see
that the signature field has been populated.

We are now ready to broadcast this transaction to the chain.

### Broadcasting the transaction

To broadcast the signed transaction to the chain, we can use the `gnokey broadcast`
subcommand, giving it the path to the signed transaction:

```bash
gnokey broadcast -remote "https://rpc.gno.land:443" userbook.tx
```

In this case, we do not need to specify a keypair, as the transaction has already
been signed in a previous step and `gnokey` is only sending it to the RPC endpoint.

## Verifying a transaction's signature

To verify a transaction's signature is correct, you can use the `gnokey verify`
subcommand. We can provide the path to the transaction document using the `-docpath`
flag, provide the key we signed the transaction with, and the signature itself.
Make sure the signature is in the `hex` format.

```bash
gnokey verify -docpath userbook.tx mykey <signature>
```

## Conclusion

That's it! 🎉

In this tutorial, you've learned to use `gnokey` for interacting with a
Gno.land chain. By mastering state-changing calls, read-only queries, and airgapped
transactions, you're now equipped to manage interactions within the Gno.land
ecosystem securely and efficiently.