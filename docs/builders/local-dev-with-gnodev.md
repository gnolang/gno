# Running a local dev node


## Installation

To install `gnodev`, simply clone the monorepo, and run `make install`:

```
git clone git@github.com:gnolang/gno.git
cd gno && make install
```

## Overview

In this tutorial, you will learn how to run up a local development node with
`gnodev`. By spinning up a local Gno.land
node, users can simulate the blockchain environment locally on their machines,
allowing them to easily see how their code behaves before deploying it to a
remote Gno.land network.

This tutorial will show you how to use gnodev, a local development solution stack
offering a built-in Gno.land node with a hot-reload feature for packages and 
realms, as well as a built-in instance of [gnoweb](../users/explore-with-gnoweb.md).

## Primary features

Apart from providing a built-in Gno.land node and a `gnoweb` instance, `gnodev`
also provides an array of other useful features. Let's explore the three most
prominent ones:
1. Automatic package deployment
2. Premining balances
3. Hot reload

If you're familiar with the features above, jump to the [practical example
section](#practical-example).

`gnodev` also provides many useful features such as loading genesis transactions,
resolving packages from remote networks, modifying the built-in node parameters,
etc. Check out the full `gnodev` developer guide for more information.

### 1. Automatic deployment

`gnodev` automatically deploys your contracts to the built-in node, making
them readily accessible via `gnoweb`. This means that developers do not need to
manually deploy their contracts during local development.

Packages and realms are deployed with a default Gno address[^1], which can be changed
via the `-deploy-key` flag.

#### Detecting package paths

If the current working directory contains a `gnomod.toml` file, `gnodev` deploys the
package to the `pkgpath` specified inside. Check out [this page](../resources/configuring-gno-projects.md)
for more info.

If no `gnomod.toml` file is found, `gnodev` searches for a `.gno` file containing a
package name and deploys it under `gno.land/r/dev/<pkgname>`.

#### Deploying example packages

In addition to your working directory, `gnodev` automatically deploys all packages
and realms located in the [examples/ folder](https://github.com/gnolang/gno/tree/master/examples)
from the monorepo it was installed from. This makes all packages in the `examples/`
folder available for use during development. `gnodev` also provides the option
to resolve packages from a remote testnet, which can be set via the `-resolver` flag.

### 2. Premining balances

`gnodev` automatically detects your Gno keys from the local `gnokey` keybase, and
pre-mines a large amount of testnet GNOT to all of your addresses, which can
then be used for testing applications.

You can verify the balance of your addresses by pressing `A` when `gnodev` is running:

```
Accounts    ┃ I (2) known keys
            ┃   table=
            ┃   │ KeyName  Address                                   Balance
            ┃   │ test1    g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5  10000000000000ugnot
            ┃   │ MyKey    g1q4q3uegdnq9rsvf3xgxydr3yqd2v6w2tww5920  10000000000000ugnot
```

This simplifies development by removing the need to manually acquire testnet GNOT.
This is not the case for remote testnets, where users must obtain testnet GNOT
from faucets, such as the ones found on [faucet.gno.land](https://faucet.gno.land).

### 3. Hot reload

`gnodev` watches the current working directory for any changes that happen within
your code, and automatically reloads the built-in node, while trying to replay
previous transactions to maintain the state of your smart contracts between
code changes.

Directory watching, as well as transaction replaying, can be disabled with the
`-no-watch` and `-no-replay` flags, respectively.

With the main features of `gnodev` out of the way, let's dive into a practical
example.

## Practical example

Let's use the local file structure we set up in the [previous tutorial](anatomy-of-a-gno-package.md):

```
counter/
    ├─ gnomod.toml
    ├─ counter.gno
```

Let's go into the `counter` folder and run `gnodev`:

```bash
cd counter
gnodev
```

You should receive an output similar to the following:

```bash
❯ gnodev
Loader      ┃ I guessing directory path path=gno.land/r/example/counter dir={your_pwd}
Accounts    ┃ I default address imported name=test1 addr=g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
Node        ┃ I packages paths=[gno.land/r/example/counter]
Event       ┃ I sending event to clients clients=0 type=NODE_RESET event=&{}
GnoWeb      ┃ I gnoweb started lisn=http://127.0.0.1:8888
--- READY   ┃ I for commands and help, press `h` took=1.391020125s
```

By opening the `gnoweb` listener address, [`http://localhost:8888`](http://127.0.0.1:8888),
we should see the render of our `counter` realm:

```
Current counter value: 0
```

### Modifying `Render()`

Let's modify the `Render()` function inside `counter.gno` as follows, importing
the `strconv` package:

```go
func Render(_ string) string {
	return "My amazing counter value: " + strconv.Itoa(count)
}
```

`gnodev` will automatically detect the change in the file and reload both the node
and `gnoweb`. The render of our realm will then change:

```
My amazing counter value: 0
```

### Interacting with the realm

To interact with our `counter` realm, let's create a simple transaction calling
the `Increment()` function with `gnokey`, using the key we created in the
[in with `gnokey`](../users/interact-with-gnokey.md#generating-a-key-pair).
Running the following command in your terminal will execute the transaction:

```
gnokey maketx call \
-pkgpath "gno.land/r/example/counter" \
-func "Increment" \
-args "42" \
-gas-fee 1000000ugnot \
-gas-wanted 20000000 \
-broadcast \
{MYKEY}
```

After entering the keypair password, you should get a response similar to this:

```
Enter password.
(42 int)

OK!
GAS WANTED: 20000000
GAS USED:   126933
HEIGHT:     203
EVENTS:     []
TX HASH:    k+WuKgPpoAg+EcR2EnzqxeWqUXB4KhOhg3l6zthSy0I=
```

Looking at the render of our realm, we'll see that the value of the counter
has increased, as expected:

```
My amazing counter value: 42
```

:::info

The above section showcases a simple `gnokey` command that will execute a
transaction executing `Increment(42)` in the `counter` realm, which lives on the
`gno.land/r/example/counter` package path on the local node.

A detailed explanation how to use `gnokey` will be provided in an
upcoming tutorial.

:::

After running gnodev, you can access several components:

1. A local version of [gnoweb](../users/explore-with-gnoweb.md)
2. A local blockchain instance for testing
3. The web-based gnodev UI to monitor your node

[^1]: The default deployer address is `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`,
a.k.a. `test1` - the mnemonic phrase for this address is publicly known.
