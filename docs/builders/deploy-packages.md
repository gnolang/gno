# Deploying Gno code

## Prerequisites

- A Gno address in `gnokey`. For setting up `gnokey`, see
  [Installation](developing-locally/installation.md).

## Overview

In this tutorial, you will learn how to deploy Gno code to a gno.land network via
the CLI using `gnokey`. We will be reusing code from a 
[previous tutorial](developing-locally/running-testing-gno.md#setup).

### A word about gas

For any state-changing call on the gno.land network, which includes code deployment,
users must pay an execution fee, commonly known as a transaction fee. This 
mechanism prevents DoS attacks and is integral to most blockchain networks.

Transaction fees on the gno.land network are paid with gno.land's native coin, 
GNOT, denominated as `ugnot` (micro-GNOT, `1 GNOT = 1_000_000 ugnot`). 

The transaction fee is calculated as `gas-fee * gas-wanted`, where `gas-fee` is 
the current price of a unit of gas in `ugnot`, and `gas-wanted` is the total number of 
gas units spent for executing the transaction.

### Getting testnet GNOT

When working with [remote networks](../concepts/testnets.md), users need to get
testnet `ugnot` manually.

`ugnot` for development on remote networks can be obtained via the [Gno Faucet Hub](https://faucet.gno.land).
Select your desired network, input a Gno address to which you want to receive
GNOT, complete the captcha and request tokens. Soon you should have a GNOT balance.

If you don't have a Gno address, check out [Creating a key pair](developing-locally/creating-a-keypair.md),
or create one via a third-party web extension wallet, such as Adena.

## Deploying with `gnokey`

Consider the following directory structure for our `Counter` realm:

```
counter/
    ├─ gno.mod
    ├─ counter.gno
    ├─ counter_test.gno
```

Let's deploy the `Counter` realm to the [Portal Loop](../concepts/testnets.md#portal-loop) 
network. For this, we can use the `gnokey maketx addpkg` subcommand, which
executes a package deployment transaction.

We need to tell `gnokey` a couple of things:
- `pkgpath`[^1] to which we want to deploy to on the chain,
- `pkgdir` in which the package is found locally,
- `gas-fee` and `gas-wanted` values,
- the `remote` (RPC endpoint) and `chainid` of the Portal Loop network[^2], 
- that we want to broadcast the transaction, and
- the key or the address we want to use to deploy the package.

The full command would look something like the following:
```
gnokey maketx addpkg \
-pkgpath "gno.land/r/<your_address>/counter" \
-pkgdir "." \
-gas-fee 10000000ugnot \
-gas-wanted 8000000 \
-broadcast \
-chainid portal-loop \
-remote "https://rpc.gno.land:443" \
MyKey 
```

To go into more detail:
- Since we're deploying a realm, the pkgpath must start with `r/`.
- You can only deploy code within your own namespace, which is based on your address[^3].
- `gas-fee` and `gas-wanted` must be set manually. If you run into an `out of gas` 
error, try increasing the `gas-wanted` value [^4].

After entering your password, you will have successfully deployed the `Counter` 
realm to the Portal Loop network:

```
OK!
GAS WANTED: 8000000
GAS USED:   6288988
HEIGHT:     955
EVENTS:     []
TX HASH:    11fWJtYXQlyFcHY12HU1ECYs2GPo/e2z/Fdw6I8rwNs=
```

## Conclusion

Congratulations! If everything went as expected, you've successfully deployed a 
realm to the Portal Loop network. To see it on `gnoweb` for the Portal Loop,
append `r/<your_address>/counter` to https://gno.land in your browser.

If you wish to learn more about `gnokey`, check out the [gnokey developer guides](../dev-guides/gnokey/overview.md).

:::info

Gno code can also be deployed via the web, using the 
[Gno Playground](https://play.gno.land). Deploying via the Playground requires
a third-party web extension wallet, such as Adena.

:::

[^1]: Read more about package paths [here](../concepts/pkg-paths.md).
[^2]: Other network configurations can be found [here](../reference/network-config.md).
[^3]: Address namespaces ([PA namespaces](../concepts/pkg-paths.md#gno-namespaces)) are automatically granted to 
users. Users can register a username using the [gno.land user registry](https://gno.land/r/gnoland/users), 
which will grant them access to a matching namespace for that specific network.
[^4]: Automatic gas estimation is being worked on for `gnokey`. Follow progress 
[here](https://github.com/gnolang/gno/pull/3330).
