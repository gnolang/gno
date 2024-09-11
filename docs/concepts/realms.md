---
id: realms
---

# Realms

In gno.land, realms are entities that are addressable and identifiable by a 
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

As opposed to [packages](./packages.md), realms are stateful, meaning they keep
their state between transactions calls. In practice, global variables in the
realm code are automatically persisted after a transaction has been executed,
resulting in the fact that developers do not need to bother with the intricacies 
of state management and persistence, like they do with other languages.

### On-chain paths

Since gno.land is built for full transparency and auditability, all on-chain Gno
code is open-sourced. You can view realm code with an ABCI query, or simply going
to its path in your web browser. For example, to take a look at the
`gno.land/r/demo/users` realm, used for user registration, by visiting
[`gno.land/r/demo/users`](https://gno.land/r/demo/users/users.gno).

:::info
Depending on the network, the realm domain might change. Currently,
the `gno.land/` domain (and all of its subdomains, such as `r/`) is pointing to
the [Portal Loop](./portal-loop.md) testnet endpoint, which is subject
to change. To view realms on the `test3` network (depr.), prepend `test3` to
the domain: [`test3.gno.land/r/demo/users`](https://test3.gno.land/r/demo/users).
:::

To learn how to write a realm, see [How to write a simple Gno Smart Contract](../how-to-guides/simple-contract.md).

## Externally Owned Accounts (EOAs)

EOAs, or simply `user realms`, are Gno addresses generated from a BIP39 mnemonic
phrase in a key management application, such as
[`gnokey`](../gno-tooling/cli/gnokey.md), and [Adena](https://adena.app).

Currently, EOAs are the only realms that can initiate a transaction. They can do
this by calling any of the possible messages in gno.land, such as 
[Call](../gno-tooling/cli/gnokey.md#call),
[AddPackage](../gno-tooling/cli/gnokey.md#addpkg),
[Send](../gno-tooling/cli/gnokey.md#send), or Run.

## Working with realms

In Gno, each transaction consists of a call stack, which is made up of `frames`.
A single frame is a unique realm in the call stack. Every frame and its properties 
can be accessed via different functions defined in the `std` package in Gno:
- `std.GetOrigCaller()` - returns the address of the original signer of the
transaction
- `std.PrevRealm()` - returns the previous realm instance, which can be a user realm
or a smart contract realm
- `std.CurrentRealm()` - returns the instance of the realm that has called it
- `std.CurrentRealmPath()` - shorthand for `std.CurrentRealm().PkgPath()`
- `std.GetCallerAt()` - returns the n-th caller's address, going back in
the transaction call stack

Let's look at return values of these functions in two distinct situations:
1. EOA calling a realm
2. EOA calling a sequence of realms

### 1. EOA calling a realm

Take these two actors in the call stack:
```
EOA:
    addr: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
    pkgPath: "" // empty as this is a user realm

Realm A:
    addr: g17m4ga9t9dxn8uf06p3cahdavzfexe33ecg8v2s
    pkgPath: gno.land/p/demo/users
    
        ┌─────────────────────┐      ┌─────────────────────────┐
        │         EOA         │      │         Realm A         │
        │                     │      │                         │
        │  addr:              │      │  addr:                  │
        │  g1jg...sqf5        ├──────►  g17m...8v2s            │
        │                     │      │                         │
        │  pkgPath:           │      │  pkgPath:               │
        │  ""                 │      │  gno.land/p/demo/users  │
        └─────────────────────┘      └─────────────────────────┘
```

Let's look at return values for each of the methods:
```go
std.GetOrigCaller() => `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`
std.PrevRealm() => Realm {
    addr:    `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`
    pkgPath: ``
}
std.CurrentRealm() => Realm {
    addr:    `g17m4ga9t9dxn8uf06p3cahdavzfexe33ecg8v2s`
    pkgPath: `gno.land/r/demo/users`
}
std.CurrentRealmPath() => `gno.land/r/demo/users`
std.GetCallerAt(1) => `g17m4ga9t9dxn8uf06p3cahdavzfexe33ecg8v2s`
std.GetCallerAt(2) => `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`
std.GetCallerAt(3) => error
```

### 2. EOA calling a sequence of realms

Take these three actors in the call stack:
```
EOA:
    addr: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
    pkgPath: "" // empty as this is a user realm

Realm A:
    addr: g1dvqd8qgvavqayxklzfdmccd2eps263p43pu2c6
    pkgPath: gno.land/p/demo/a
    
Realm B:
    addr: g1rsk9cwv034cw3s6csjeun2jqypj0ztpecqcm3v
    pkgPath: gno.land/p/demo/b

┌─────────────────────┐   ┌──────────────────────┐   ┌─────────────────────┐
│         EOA         │   │       Realm A        │   │       Realm B       │
│                     │   │                      │   │                     │
│  addr:              │   │  addr:               │   │  addr:              │
│  g1jg...sqf5        ├───►  g17m...8v2s         ├───►  g1rs...cm3v        │
│                     │   │                      │   │                     │
│  pkgPath:           │   │  pkgPath:            │   │  pkgPath:           │
│  ""                 │   │  gno.land/p/demo/a   │   │  gno.land/p/demo/b  │
└─────────────────────┘   └──────────────────────┘   └─────────────────────┘
```

Depending on which realm the methods are called in, the values will change. For
`Realm A`:
```go
std.GetOrigCaller() => `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`
std.PrevRealm() => Realm {
    addr:    `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`
    pkgPath: ``
}
std.CurrentRealm() => Realm {
    addr:    `g1dvqd8qgvavqayxklzfdmccd2eps263p43pu2c6`
    pkgPath: `gno.land/r/demo/a`
}
std.CurrentRealmPath() => `gno.land/r/demo/a`
std.GetCallerAt(1) => `g1dvqd8qgvavqayxklzfdmccd2eps263p43pu2c6`
std.GetCallerAt(2) => `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`
std.GetCallerAt(3) => error
```

For `Realm B`:
```go
std.GetOrigCaller() => `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`
std.PrevRealm() => Realm {
    addr:    `g1dvqd8qgvavqayxklzfdmccd2eps263p43pu2c6`
    pkgPath: `gno.land/r/demo/a`
}
std.CurrentRealm() => Realm {
    addr:    `g1rsk9cwv034cw3s6csjeun2jqypj0ztpecqcm3v`
    pkgPath: `gno.land/r/demo/b`
}
std.CurrentRealmPath() => `gno.land/r/demo/b`
std.GetCallerAt(1) => `g1rsk9cwv034cw3s6csjeun2jqypj0ztpecqcm3v`
std.GetCallerAt(2) => `g1dvqd8qgvavqayxklzfdmccd2eps263p43pu2c6`
std.GetCallerAt(3) => `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`
std.GetCallerAt(4) => error
```

## `Render`

A notable feature of non-EOA realms is the ability to have a render function. A
render function allows the developer of the realm to choose how to render the 
state of the realm by returning a custom-made valid Markdown string. It also 
allows the developer to define different renders depending on the `path` argument:

```go
package demo

func Render(path string) string {
	if path == "" {
		return "# Hello Gnopher!"
	}
	
	return "# Hello" + path
}
```

:::info
You can see the `Render` function in action by visiting the 
[home page of gno.land](https://gno.land/) - it is actually the render of 
`r/gnoland/home` realm. The same is true for the
[gno.land Blog](https://gno.land/r/gnoland/blog), and most other pages on the domain.
:::