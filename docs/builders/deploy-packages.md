# Deploying Gno Packages to a Network

Once you've developed and tested your Gno packages locally, the next step is
deploying them to a gno.land network. This guide explains how to deploy both
realms and pure packages using `gnokey`.

## Prerequisites

Before deploying, you need:

1. A working version of your package or realm
2. A gno.land account with sufficient GNOT for gas fees
3. The `gnokey` utility installed and configured
4. (Optional) A registered namespace for deploying under your own path

In this tutorial, you will learn how to deploy Gno code to a gno.land network
via the CLI using `gnokey`. We will be reusing code from a
[previous tutorial](developing-locally/running-testing-gno.md#setup).

### A word about gas

For any state-changing call on the gno.land network, which includes code
deployment, users must pay an execution fee, commonly known as a transaction
fee. This mechanism prevents DoS attacks and is integral to most blockchain
networks.

Transaction fees on the gno.land network are paid with gno.land's native coin,
GNOT, denominated as `ugnot` (micro-GNOT, `1 GNOT = 1_000_000 ugnot`).

The transaction fee is calculated as `gas-fee * gas-wanted`, where `gas-fee` is
the current price of a unit of gas in `ugnot`, and `gas-wanted` is the total
number of gas units spent for executing the transaction.

### Getting testnet GNOT

When working with [remote networks](../resources/gnoland-networks.md), users
need to get testnet `ugnot` manually.

`ugnot` for development on remote networks can be obtained via the
[Gno Faucet Hub](https://faucet.gno.land). Select your desired network, input a
Gno address to which you want to receive GNOT, complete the captcha and request
tokens. Soon you should have a GNOT balance.

If you don't have a Gno address, check out
[Creating a key pair](developing-locally/creating-a-keypair.md), or create one
via a third-party web extension wallet, such as Adena.

## Deploying with `gnokey`

Consider the following directory structure for our `Counter` realm:

```
counter/
    ├─ gno.mod
    ├─ counter.gno
    ├─ counter_test.gno
```

Let's deploy the `Counter` realm to the
[Portal Loop](../resources/gnoland-networks.md#portal-loop) network. For this,
we can use the `gnokey maketx addpkg` subcommand, which executes a package
deployment transaction.

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

## Choosing a Package Path

When deploying to gno.land, you need to specify a package path. You have two
options:

1. **Use your registered username** - If you've registered a username, you can deploy under `gno.land/[r|p]/YOUR_USERNAME/...`
2. **Use your address namespace** - Without a username, you can deploy under `gno.land/[r|p]/YOUR_ADDRESS/...`

For more information on registering usernames and namespace ownership, see the
[Users and Teams documentation](../resources/users-and-teams.md).

## Registering a Namespace

For production packages, you'll want your own namespace:

1. Follow the [Username Registration](../resources/users-and-teams.md#registration-process) instructions
2. Once registered, deploy under `gno.land/[r|p]/YOUR_USERNAME/...`

This gives you a more human-readable package path and establishes your identity in the ecosystem.

## Understanding Deployment Parameters

- `--pkgpath` - The on-chain path where your code will be stored
- `--pkgdir` - The local directory containing your code
- `--deposit` - The amount of GNOT to deposit (typically 100 GNOT)
- `--gas-fee` - The fee per unit of gas (typically 1 GNOT)
- `--gas-wanted` - Maximum gas units for the transaction
- `--remote` - The RPC endpoint for the network
- `--chainid` - The ID of the blockchain network

For more details on gas fees and optimization strategies, see the [Gas Fees
documentation](../resources/gas-fees.md).

## Conclusion

Congratulations! If everything went as expected, you've successfully deployed a
realm to the Portal Loop network. To see it on `gnoweb` for the Portal Loop,
append `r/<your_address>/counter` to https://gno.land in your browser.

:::info

Gno code can also be deployed via the web, using the
[Gno Playground](https://play.gno.land). Deploying via the Playground requires
a third-party web extension wallet, such as Adena.

:::

[^1]: Read more about package paths [here](../resources/gno-packages.md).
[^2]: Other network configurations can be found [here](../resources/gnoland-networks.md).
[^3]: Address namespaces ([PA namespaces](../resources/gno-packages.md#package-path-structure)) are automatically granted to
users. Users can register a username using the [gno.land user registry](https://gno.land/r/demo/users),
which will grant them access to a matching namespace for that specific network.
[^4]: Automatic gas estimation is being worked on for `gnokey`. Follow progress
[here](https://github.com/gnolang/gno/pull/3330).
