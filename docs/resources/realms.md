# Realms

In gno.land, realms are entities that are addressable and identifiable by a
[Gno address](./gno-stdlibs.md#address). These can be user
realms (EOAs), as well as smart contract realms. Realms have several
properties:
- They can own, receive & send [Coins](./gno-stdlibs.md#coin) through the
  [Banker](./gno-stdlibs.md#banker) module
- They can be part of a transaction call stack, as a caller or a callee
- They can be with or without code - smart contracts, or EOAs

Realms are represented by a `Realm` type in Gno:
```go
type Realm struct {
    addr    Address // Gno address in the bech32 format
    pkgPath string  // realm's path on-chain
}
```

The full Realm API can be found under the
[reference section](./gno-stdlibs.md).

### Smart Contract Realms

Often simply called `realms`, Gno smart contracts contain Gno code and exist
on-chain at a specific [package path](gno-packages.md). A package path is the
defining identifier of a realm, while its address is derived from it.

As opposed to [pure packages](./gno-packages.md#pure-packages-p), realms are
stateful, meaning they keep their state between transaction calls. In practice,
global variables used in realms are automatically persisted after a transaction
has been executed. Thanks to this, Gno developers do not need to bother with the
intricacies of state management and persistence, like they do with other
languages.

### Externally Owned Accounts (EOAs)

EOAs, or simply `user realms`, are Gno addresses generated from a BIP39 mnemonic
phrase in a key management application, such as
[gnokey](../users/interact-with-gnokey.md), and web wallets, such as
[Adena](../users/third-party-wallets.md).

Currently, EOAs are the only realms that can initiate a transaction. They can do
this by calling any of the possible messages in gno.land, which can be
found [here](../users/interact-with-gnokey.md#making-transactions).

### Working with realms

In Gno, each transaction contains a realm call stack. Every item in the stack and
its properties can be accessed via different functions defined in the `std`
package in Gno:
- `std.OriginCaller()` - returns the address of the original signer of the
  transaction
- `std.PreviousRealm()` - returns the previous realm instance, which can be a user realm
  or a smart contract realm
- `std.CurrentRealm()` - returns the instance of the realm that has called it

Let's look at the return values of these functions in two distinct situations:
1. EOA calling a realm
2. EOA calling a sequence of realms

#### 1. EOA calling a realm

Take these two actors in the call stack:
```
EOA:
    addr: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
    pkgPath: "" // empty as this is a user realm

Realm A:
    addr: g17m4ga9t9dxn8uf06p3cahdavzfexe33ecg8v2s
    pkgPath: gno.land/r/demo/users

        ┌─────────────────────┐      ┌─────────────────────────┐
        │         EOA         │      │         Realm A         │
        │                     │      │                         │
        │  addr:              │      │  addr:                  │
        │  g1jg...sqf5        ├──────►  g17m...8v2s            │
        │                     │      │                         │
        │  pkgPath:           │      │  pkgPath:               │
        │  ""                 │      │  gno.land/r/demo/users  │
        └─────────────────────┘      └─────────────────────────┘
```

Let's look at return values for each of the methods, called from within `Realm A`:
```
std.OriginCaller() => `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`
std.PreviousRealm() => Realm {
    addr:    `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`
    pkgPath: ``
}
std.CurrentRealm() => Realm {
    addr:    `g17m4ga9t9dxn8uf06p3cahdavzfexe33ecg8v2s`
    pkgPath: `gno.land/r/demo/users`}
```

#### 2. EOA calling a sequence of realms

Take these three actors in the call stack:
```
EOA:
    addr: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
    pkgPath: "" // empty as this is a user realm

Realm A:
    addr: g1dvqd8qgvavqayxklzfdmccd2eps263p43pu2c6
    pkgPath: gno.land/r/demo/a

Realm B:
    addr: g1rsk9cwv034cw3s6csjeun2jqypj0ztpecqcm3v
    pkgPath: gno.land/r/demo/b

┌─────────────────────┐   ┌──────────────────────┐   ┌─────────────────────┐
│         EOA         │   │       Realm A        │   │       Realm B       │
│                     │   │                      │   │                     │
│  addr:              │   │  addr:               │   │  addr:              │
│  g1jg...sqf5        ├───►  g17m...8v2s         ├───►  g1rs...cm3v        │
│                     │   │                      │   │                     │
│  pkgPath:           │   │  pkgPath:            │   │  pkgPath:           │
│  ""                 │   │  gno.land/r/demo/a   │   │  gno.land/r/demo/b  │
└─────────────────────┘   └──────────────────────┘   └─────────────────────┘
```

Depending on which realm the methods are called in, the values will change. For
`Realm A`:
```
std.OriginCaller() => `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`
std.PreviousRealm() => Realm {
    addr:    `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`
    pkgPath: ``
}
std.CurrentRealm() => Realm {
    addr:    `g1dvqd8qgvavqayxklzfdmccd2eps263p43pu2c6`
    pkgPath: `gno.land/r/demo/a`
}
```

For `Realm B`:
```
std.OriginCaller() => `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`
std.PreviousRealm() => Realm {
    addr:    `g1dvqd8qgvavqayxklzfdmccd2eps263p43pu2c6`
    pkgPath: `gno.land/r/demo/a`
}
std.CurrentRealm() => Realm {
    addr:    `g1rsk9cwv034cw3s6csjeun2jqypj0ztpecqcm3v`
    pkgPath: `gno.land/r/demo/b`
}
```

For more information about realms and how they fit into the gno.land ecosystem,
see the [Package Path Structure](./gno-packages.md#package-path-structure)
documentation.

To learn how to develop your own realms, check out the
[Anatomy of a Gno Package](../builders/anatomy-of-a-gno-package.md) and
[Example Minisocial dApp](../builders/example-minisocial-dapp.md) guides.
