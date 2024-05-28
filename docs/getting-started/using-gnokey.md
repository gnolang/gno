---
id: using-gnokey
---

# Using `gnokey`

## Overview
In this tutorial, you will learn how to use `gnokey`, a tool used for 


which are
required for interacting with the Gno.land blockchain. You will understand what
mnemonics are, how they are used, and how you can make interaction seamless with
Gno.

## Prerequisites
- **`gno`, `gnokey`, and `gnodev` installed.** Reference the
  [Local Setup](local-setup/installation.md#2-installing-the-required-tools-) guide for steps

## Keypairs


## Interacting with a Gno.land chain

`gnokey` allows you to interact with any Gno.land chain, such as the Portal Loop. 

There are multiple ways anyone can interact with the chain:
- Transactions - state-changing calls which use up gas
- ABCI queries - read-only calls which do not use up gas

Both transactions and ABCI queries can be used via `gnokey`'s subcommands,
`maketx` and `query`.

## State-changing calls (transactions)

In Gno, there are three types of messages that can change on-chain state:
- `AddPackage` - adds code to the chain
- `Call` - calls a specific path and function on the chain
- `Run` - executes a Gno script against on-chain code

A Gno.land transaction contains two main things: 
- A base configuration where variables such as `gas-fee`, `gas-wanted`, and others
are defined
- A list of messages to execute on the chain 

Currently, `gnokey` supports single-message transactions, while multiple-message
transactions can be created in Go programs, supported by the
[gnoclient](../reference/gnoclient/gnoclient.md) package.

We will need some testnet GNOTs for each state-changing call. Visit the [Faucet
Hub](https://faucet.gno.land) to get GNOTs for the currently live Gno testnets.

Let's delve deeper into each of these messages.

### `AddPackage`

In case you want to upload new code to the chain, you can use the `AddPackage` 
message type. You can send an `AddPackage` transaction with `gnokey` using the 
following command:

```bash
gnokey maketx addpkg
```

To understand how to use this subcommand better, let's write a simple "Hello world"
[pure package](../concepts/packages.md). First create a folder which will store 
our example code.

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

In the `hello_world.gno` file, add define the following code:

```go
package hello_world

func Hello() string {
	return "Hello, world!"
}
```

We are now ready to upload this packge to the chain. To do this, we must set the
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

For this demonstration, we will run a local Gno node using `gnodev`. First, simply
start `gnodev`:

```bash
gnodev
```

If everything went well, you should see the following output:
```bash
❯ gnodev
Accounts    ┃ I default address imported name=test1 addr=g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
Node        ┃ I pkgs loaded path="[{<your_monorepo_path> g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 }]"
Node        ┃ I node started lisn=tcp://127.0.0.1:36657 chainID=dev
GnoWeb      ┃ I gnoweb started lisn=http://127.0.0.1:8888
--- READY   ┃ I for commands and help, press `h`
```

Now we have a local Gno node listening on `127.0.0.1:36657` with chain ID `dev`,
which we can use to upload our code to.

Next, let's configure the `addpkg` subcommand. Assuming we are in the `example/p` 
folder, the command will look like this:

```bash
gnokey maketx addpkg \                                                                                                                                                                                          
--pkgpath "gno.land/p/<your_namespace>/hello_world" \
--pkgdir "." \
--send "" \
--gas-fee 10000000ugnot \
--gas-wanted 8000000 \
--broadcast \
--chainid dev \
--remote "127.0.0.1:36657" \
```

Once we have added a desired namespace to upload the package to, we can specify
a keypair name to use to execute the transaction:

```bash
gnokey maketx addpkg \                                                                                                                                                                                          
--pkgpath "gno.land/p/leon/hello_world" \
--pkgdir "." \
--send "" \
--gas-fee 10000000ugnot \
--gas-wanted 200000 \
--broadcast \
--chainid dev \
--remote "127.0.0.1:36657" \
dev
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

Congratulations! You have just uploaded a pure package to the chain.

### `Call`

You can call any exported function on the chain using the `call` message type. 
You can send a `Call` transaction with `gnokey` using the following command:

```bash
gnokey maketx call
```

:::info `call` uses gas

Using `call` to call an exported function will use up gas, even if the function
does not modify on-chain state. If you are calling such a function, you can use
the [`query` functionality](#query) for a read-only call which does not use gas.

:::

For this example, we will call the `wugnot` realm, which wraps GNOTs to a 
GRC20-compatible token called `wugnot`. We can find this realm deployed on the 
 [Portal Loop](../concepts/portal-loop.md) testnet, under the `gno.land/r/demo/wugnot` 

We will wrap `100ugnot` into the equivalent in `wugnot`. To do this, we can call
the `Deposit()` function. As previously, we will configure the `maketx call`
subcommand:

```bash
gnokey maketx call \
--pkgpath "gno.land/r/demo/wugnot" \
--func "Deposit" \
--send "1000ugnot" \
--gas-fee 10000000ugnot \
--gas-wanted 2000000 \
--broadcast \
--chainid portal-loop \
--remote https://rpc.gno.land:443 \
dev
```

In this command, we have specified three main things:
- The path where the realm lives on-chain with the `-pkgpath` flag
- The function  that we want to call on the realm with the `-func` flag
- The amount of `ugnot` we want to deposit to wrap using the `-send` flag

Apart from this, we have also specified 

### ABCI queries

## `query`











