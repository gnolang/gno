---
id: querying-a-network
---

# Querying a gno.land network

## Prerequisites

- **`gnokey` installed.** Reference the
  [Local Setup](../../../getting-started/local-setup/installation.md#2-installing-the-required-tools) guide for steps

## Overview

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
- `vm/qfile` - returns the list of files for a given pkgpath
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

The `data` field returns a `BaseAccount`, which is the main struct used in [TM2](../../../concepts/tendermint2.md)
to hold account data. It contains the following information:
- `address` - the address of the account
- `coins` - the list of coins the account owns
- `public_key` - the TM2 public key of the account, from which the address is derived
- `account_number` - a unique identifier for the account on the gno.land chain
- `sequence` - a nonce, used for protection against replay attacks

## `bank/balances`

With this query, we can fetch [coin](../../../concepts/stdlibs/coin.md) balances
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

## `vm/qeval`

`vm/qeval` allows us to evaluate a call to an exported function without using gas,
in read-only mode. For example:

```bash
gnokey query vm/qeval -remote https://rpc.gno.land:443 -data "gno.land/r/demo/wugnot.BalanceOf(\"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5\")" 
```

This command will return the `wugnot` balance of the above address without using gas.
Properly escaping quotation marks for string arguments is currently required.

// only primitive types are supported?

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

## Conclusion

That's it! 🎉

In this tutorial, you've learned to use `gnokey` to query a gno.land
network.
