--
id: using-gnokey
--

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
- `Send` - sends coins from one address to another
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
-- READY   ┃ I for commands and help, press `h`
```

Now we have a local Gno node listening on `127.0.0.1:36657` with chain ID `dev`,
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

We will wrap `1000ugnot` into the equivalent in `wugnot`. To do this, we can call
the `Deposit()` function. As previously, we will configure the `maketx call`
subcommand:

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
main
```

In this command, we have specified three main things:
- The path where the realm lives on-chain with the `-pkgpath` flag
- The function  that we want to call on the realm with the `-func` flag
- The amount of `ugnot` we want to deposit to wrap using the `-send` flag

Apart from this, we have also specified the Portal Loop chain ID, `portal-loop`,
as well as the Portal Loop remote, `https://rpc.gno.land:443`.

Chain IDs and remote addresses can be found in the 
[Network Configuration](../reference/network-config.md) page.

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
dev
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

We can use the `Send` message type to access the TM2 banker directly and transfer
coins from one Gno address to another. 

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
-chainid dev \
-remote "127.0.0.1:26657" \
dev
```

Here, we have set the `-to` & `-send` flags to match the recipient, in this case
the publicly-known `test1` address, and `100ugnot` for the coins we want to send,
respectively.

To check the balance of a specific address, check out the `bank/balances` query
in the [ABCI queries section](#bankbalances).

### `Run`



## ABCI queries

ABCI queries are available on Gno.land chains. todo add more info

for all queries, we can specify a remote address to ask for information.

### `query`

The query subcommand allows us to send different types of queries to a Gno.land
network.

Below is a list of queries a user can make with `gnokey`:
- `auth/accounts/{ADDRESS}` - returns information about an account
- `bank/balances/{ADDRESS}` - returns balances of an account
- `vm/qfuncs` - returns the exported functions for a given pkgpath
- `vm/qfile` - returns the list of files for a given pkgpath
- `vm/qeval` - evaluates an expression in read-only mode on and returns the results 
- `vm/qrender` - shorthand for evaluating `vm/qeval Render("")` for a given pkgpath

Let's see how we can use them.

#### `auth/accounts`

We can obtain information on a specific address using this subquery. To call it,
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
- `height` - the height at which the query was executed. This is currently not supported.
- `data` - contains the result of the query.

The `data` field returns a `BaseAccount`, which is the main struct used in TM2 to
hold account data. It contains the following information:
- `address` - the address of the account
- `coins` - the list of coins the account owns
- `public_key` - the TM2 public key of the account, which the address is derived from
- `account_number` - a unique identifier for the account on the Gno.land chain
- `sequence` - a nonce, used for protection against replay attacks

#### `bank/balances`

With this query, we can fetch balances of a specfic account. To call it, we can
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

#### `vm/qfuncs`

Using the `vm/qfuncs` query, we can fetch exported functions from a specific package
path. To specify the path we want to query, we can use the `-data` flag:

```bash
gnokey query vm/qfuncs --data "gno.land/r/demo/wugnot" -remote https://rpc.gno.land:443
```

The output is a JSON-formatted string containing all exported functions for the
`wugnot` realm:

```json
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

#### `vm/qfile`

With the `vm/qfile` query, we can fetch files found on a specific package path.
To specify the path we want to query, we can use the `-data` flag:

```bash
gnokey query vm/qfile -data "gno.land/r/demo/wugnot" -remote https://rpc.gno.land:443
```

The output is a JSON-formatted string containing all exported functions for the
`wugnot` realm:

```bash
height: 0
data: gno.mod
wugnot.gno
z0_filetest.gno
```

#### `vm/qeval`

`vm/qeval` allows us to evaluate a call to an exported function without using gas,
in read-only mode. For example:

```bash
gnokey query vm/qeval -remote https://rpc.gno.land:443 -data "gno.land/r/demo/wugnot
BalanceOf(\"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5\")" 
```

This command will return the `wugnot` balance of the above address without using gas.
Properly escaping quotation marks, and inputting the newline for the function
is currently required.

#### `vm/qrender`

`vm/qrender` is an alias for executing `vm/qeval` on the `Render("")` function.
We can use it like this:

```bash
gnokey query vm/qrender -remote https://rpc.gno.land:443 -data "gno.land/r/demo/wugnot"
// not working?
```

#### `vm/qrender`

`vm/qrender` is an alias for executing `vm/qeval` on the `Render("")` function.
We can use it like this:

```bash
gnokey query vm/qrender -remote https://rpc.gno.land:443 -data "gno.land/r/demo/wugnot"
// not working?
```


