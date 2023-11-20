---
id: deploy
---

# How to deploy a Realm / Package

## Overview

This guide shows you how to deploy any realm or package to the Gno chain. Deployment is be done by utilizing `gnokey`'s `maketx addpkg` API.

:::info
Regardless of whether you're deploying a realm or a package, you will be using `gnokey`'s `maketx addpkg` - the usage of `maketx addpkg` in both cases is identical. 
:::

## Prerequisites

- **Have `gnokey` installed**
- **Have access to a `gnoland` node (local or remote)**
- **Have generated a keypair with `gnokey` & funded it with `gnot`**
- **Have a Realm or Package ready to deploy**

## Deploying

To illustrate deployment, we will use a realm. Consider the following folder structure:

```
counter-app/
├─ r/
│  ├─ counter/
│  │  ├─ counter.gno
```

We would like to deploy the realm found in `counter.gno`. To do this, open a terminal at `counter-app/` and use the following `gnokey` command:

```bash
gnokey maketx addpkg \
--pkgpath "gno.land/r/demo/counter" \
--pkgdir "./r/counter" \
--gas-fee 10000000ugnot \
--gas-wanted 800000 \
--broadcast \
--chainid dev \
--remote localhost:26657 \
MyKey
```

Let's analyze all of the flags in detail:
- `--pkgpath` - path where the package/realm will be placed on-chain
- `--pkgdir` - local path where the package/realm is located
- `--gas-wanted` - the upper limit for units of gas for the execution of the transaction - similar to Solidity's gas limit
- `--gas-fee` - similar to Solidity's gas-price
- `--broadcast` - broadcast the transaction on-chain
- `--chain-id` - id of the chain to connect to - local or remote
- `--remote` - `gnoland` node endpoint - local or remote
- `MyKey` - the keypair to use for the transaction

:::info
As of October 2023, `--gas-fee` is fixed to 1gnot (10000000ugnot), with plans to change it down the line.
:::

Next, confirm the transaction with your keypair passphrase. If deployment was successful, you will be presented with a message similar to the following:

```
OK!
GAS WANTED: 800000
GAS USED:   775097
```
Depending on the size of the package/realm, you might need to increase amount given in the `--gas-wanted` flag to cover the deployment cost.

## Conclusion

That's it 🎉

You have now successfully deployed a realm/package to a Gno.land chain. 
