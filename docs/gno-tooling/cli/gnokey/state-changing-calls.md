---
id: state-changing-calls
---

# Making state-changing calls (transactions)

## Prerequisites

- **`gnokey` installed.** Reference the
  [Local Setup](../../../getting-started/local-setup/installation.md#2-installing-the-required-tools) guide for steps

## Overview

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
[gnoclient](../../../reference/gnoclient/gnoclient.md) package.

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
[pure package](../../../concepts/packages.md). First, let's create a folder which will
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
- `-chain-id` - id of the chain that we are sending the transaction to
- `-remote` - specifies the remote node RPC listener address

The `-pkgpath` and `-pkgdir` flags are unique to the `addpkg` subcommand, while
`-broadcast`,`-send`, `-gas-wanted`, `-gas-fee`, `-chain-id`, and `-remote` are
used for setting the base transaction configuration. These flags will be repeated
throughout the tutorial.

Next, let's configure the `addpkg` subcommand to publish this package to the
[Portal Loop](../../../concepts/portal-loop.md) testnet. Assuming we are in
the `example/p/` folder, the command will look like this:

```bash
gnokey maketx addpkg \                                                                                                                                                                                          
-pkgpath "gno.land/p/<your_namespace>/hello_world" \
-pkgdir "." \
-send "" \
-gas-fee 10000000ugnot \
-gas-wanted 8000000 \
-broadcast \
-chainid portal-loop \
-remote "https://rpc.gno.land:443"
```

Once we have added a desired [namespace](../../../concepts/namespaces.md) to upload the package to, we can specify
a keypair name to use to execute the transaction:

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

If the transaction was successful, you will get output from `gnokey` that is similar to the following:

```
OK!
GAS WANTED: 200000
GAS USED:   117564
HEIGHT:     3990
EVENTS:     []
TX HASH:    Ni8Oq5dP0leoT/IRkKUKT18iTv8KLL3bH8OFZiV79kM=
COMMIT DURATION:  1000ms
```

Let's analyze the output, which is standard for any `gnokey` transaction:
- `GAS WANTED:      200000` - the original amount of gas specified for the transaction
- `GAS USED:        117564` - the gas used to execute the transaction
- `HEIGHT:          3990` - the block number at which the transaction was executed at
- `EVENTS:          []` - [Gno events](../../../concepts/stdlibs/events.md) emitted by the transaction, in this case, none
- `TX HASH:         Ni8Oq5dP0leoT/IRkKUKT18iTv8KLL3bH8OFZiV79kM=` - the hash of the transaction
- `COMMIT DURATION: 1000ms` - the time from the transaction submission to being committed on-chain

Congratulations! You have just uploaded a pure package to the Portal Loop network.
If you wish to deploy to a different network, find the list of all network 
configurations in the [Network Configuration](../../../reference/network-config.md) section.

## `Call`

The `Call` message type is used to call any exported realm function.
You can send a `Call` transaction with `gnokey` using the following command:

```bash
gnokey maketx call
```

:::info `Call` uses gas

Using `Call` to call an exported function will use up gas, even if the function
does not modify on-chain state. If you are calling such a function, you can use
the [`query` functionality](./querying-a-network.md) for a read-only call which
does not use gas.

:::

For this example, we will call the `wugnot` realm, which wraps GNOTs to a
GRC20-compatible token called `wugnot`. We can find this realm deployed on the
[Portal Loop](../../../concepts/portal-loop.md) testnet, under the `gno.land/r/demo/wugnot` path.

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
GAS WANTED:       2000000
GAS USED:         489528
HEIGHT:           24142
EVENTS:           [{"type":"Transfer","attrs":[{"key":"from","value":""},{"key":"to","value":"g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"},{"key":"value","value":"1000"}],"pkg_path":"gno.land/r/demo/wugnot","func":"Mint"}]
TX HASH:          Ni8Oq5dP0leoT/IRkKUKT18iTv8KLL3bH8OFZiV79kM=
COMMIT DURATION:  1000ms
```

In this case, we can see that the `Deposit()` function emitted an 
[event](../../../concepts/stdlibs/events.md) that tells us more about what
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
GAS WANTED:       2000000
GAS USED:         396457
HEIGHT:           64839
EVENTS:           []
TX HASH:          gQP9fJYrZMTK3GgRiio3/V35smzg/jJ62q7t4TLpdV4=
COMMIT DURATION:  1000ms
```

At the top, you will see the output of the transaction, specifying the value and
type of the return argument.

In this case, we used `maketx call` to call a read-only function, which simply
checks the `wugnot` balance of a specific address. This is discouraged, as
`maketx call` actually uses gas. To call a read-only function without spending gas,
check out the `vm/qeval` query in the [Querying a network](./querying-a-network.md#vmqeval) section.

## `Send`

We can use the `Send` message type to access the TM2 [Banker](../../../concepts/stdlibs/banker.md)
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
in the [Querying a network](./querying-a-network.md#bankbalances) section.

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

## Conclusion

That's it! 🎉

In this tutorial, you've learned to use `gnokey` for sending multiple types of 
state-changing calls to a gno.land chain. 
