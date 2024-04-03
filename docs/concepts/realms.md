---
id: realms
---

# Realms
In Gno.land, realms are entities that are addressable and identifiable by a 
[Gno address](../reference/standard-library/std/address.md). These can be user 
realms (EOAs), as well as smart contract realms. Realms have several 
properties:
- They can own, receive & send [Coins](./standard-library/coin.md) through the
[Banker](./standard-library/banker.md) module
- They can be part of a transaction call stack, as a caller or a callee
- They can be with or without code - EOAs, or smart contracts

Realms are represented by a `Realm` type in Gno:
```go
type Realm struct {
    addr    Address // Gno address in the bech32 format
    pkgPath string  // realm's path on-chain
}
```
The full Realm API can be found under the 
[reference section](../reference/standard-library/std/realm.md).

## Smart Contract Realms
Often simply called `realms`, Gno smart contracts contain Gno code and exist
on-chain at a specific package path. A package path is the defining identifier
of a realm, while its address is derived from it.

### On-chain paths
Since Gno.land is built for full transparency and auditability, all on-chain Gno
code is open-sourced. You can view realm code by simply going to its path in
your web browser. For example, to take a look at the `gno.land/r/demo/users` realm,
used for user registration, by visiting
[`gno.land/r/demo/users`](https://gno.land/r/demo/users/users.gno).

:::info
Depending on the network, the realm domain might change. Currently, 
the `gno.land/` domain (and all of its subdomains, such as `r/`) is pointing to
the [Portal Loop](./portal-loop.md) testnet endpoint, which is subject 
to change. To view realms on the `test3` network (depr.), prepend `test3` to 
the domain: [`test3.gno.land/r/demo/users`](https://test3.gno.land/r/demo/users).
:::

[//]: # (Learn more about package paths & allowed namespaces [here].)

To learn how to actually write a realm,
see [How to write a simple Gno Smart Contract](../how-to-guides/simple-contract.md).

## Externally Owned Accounts (EOAs)
EOAs, or simply `user realms`, are Gno addresses generated from a BIP39 mnemonic
phrase in a key management application, such as
[`gnokey`](../gno-tooling/cli/gnokey.md), and [Adena](https://adena.app).

## Working with realms
In Gno, each transaction consists of a call stack, which is made up of `frames`.
A single frame is a unique realm in the call stack. Every frame and its properties 
can be accessed via different functions defined in the `std` package in Gno:
- `std.GetOrigCaller()` - returns the address of the original signer of the
transaction
- `std.GetOrigPkgAddr()` - returns the address of the first caller (entry point) 
in a sequence of realm calls
- `std.PrevRealm()` - returns the previous realm instance, which can be a user realm
or a smart contract realm
- `std.CurrentRealm()` - returns the instance of the realm that has called it
- `std.CurrentRealmPath()` - shorthand for `std.CurrentRealm().PkgPath()`
- `std.GetCallerAt()` - returns the n-th caller's address, going back in
the transaction call stack

Let's look at return values of these functions in two distinct situations:
- EOA calling a realm
- EOA calling a realm, in turn calling another realm





User realms are recognizable by the fact that their package path is empty.
This can be checked by calling the `IsUser()` method on the realm object. 

