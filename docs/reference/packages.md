# Packages / Smart Contracts

XXX: generic intro about packages
XXX: short diff between pure and realms
XXX: link to ./stdlibs.md

Apart from pure packages, Gno, like Go, has standard libraries. To better
understand the difference between these two concepts, let's compare a few
specific points:
- Pure packages can be written and deployed by anyone at any time, while standard
  libraries require thorough battle-testing and reviews by the core team & community
  before being added to the language
- Standard libraries usually provide low-level necessities for the language,
  while pure packages utilize them to create a broader range of functionality


## Realms

In gno.land, realms are entities that are addressable and identifiable by a
[Gno address](../reference/std.md#address). These can be user
realms (EOAs), as well as smart contract realms. Realms have several
properties:
- They can own, receive & send [Coins](./stdlibs/coin.md) through the
  [Banker](./stdlibs/banker.md) module
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
[reference section](../reference/std.md#realm).

### Smart Contract Realms

Often simply called `realms`, Gno smart contracts contain Gno code and exist
on-chain at a specific [package path](pkg-paths.md). A package path is the 
defining identifier of a realm, while its address is derived from it.

As opposed to [pure packages](./packages.md), realms are stateful, meaning they
keep their state between transaction calls. In practice, global variables used in realms 
are automatically persisted after a transaction has been executed. Thanks to this,
Gno developers do not need to bother with the intricacies of state management 
and persistence, like they do with other languages.

### Externally Owned Accounts (EOAs)

EOAs, or simply `user realms`, are Gno addresses generated from a BIP39 mnemonic
phrase in a key management application, such as
[gnokey](../dev-guides/gnokey/managing-keypairs.md), and web wallets, such as
[Adena](https://adena.app).

Currently, EOAs are the only realms that can initiate a transaction. They can do
this by calling any of the possible messages in gno.land, which can be 
found [here](../dev-guides/gnokey/making-transactions.md#overview).

### Working with realms

In Gno, each transaction contains a realm call stack. Every item in the stack and
its properties can be accessed via different functions defined in the `std` 
package in Gno:
- `std.GetOrigCaller()` - returns the address of the original signer of the
  transaction
- `std.PrevRealm()` - returns the previous realm instance, which can be a user realm
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
std.GetOrigCaller() => `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`
std.PrevRealm() => Realm {
    addr:    `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`
    pkgPath: ``
}
std.CurrentRealm() => Realm {
    addr:    `g17m4ga9t9dxn8uf06p3cahdavzfexe33ecg8v2s`
    pkgPath: `gno.land/r/demo/users`
}
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
std.GetOrigCaller() => `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`
std.PrevRealm() => Realm {
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
std.GetOrigCaller() => `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`
std.PrevRealm() => Realm {
    addr:    `g1dvqd8qgvavqayxklzfdmccd2eps263p43pu2c6`
    pkgPath: `gno.land/r/demo/a`
}
std.CurrentRealm() => Realm {
    addr:    `g1rsk9cwv034cw3s6csjeun2jqypj0ztpecqcm3v`
    pkgPath: `gno.land/r/demo/b`
}
```

Check out the realm reference page [here](../reference/std.md#realm).


## Pure Packages

Pure packages are Gno code meant to be reused by other Gno code, be it by other 
pure packages or realms. Here are some defining features of pure packages:
- Pure packages are stored on-chain under the `gno.land/p/` path, and can be
  written & deployed to the chain by anyone, permissionlessly
- Pure packages are meant to be imported from other packages & realms
- Users cannot call functions in pure packages directly
- Documentation for pure packages should be contained within package code itself,
  in the form of comments, following the [Go doc standard](https://tip.golang.org/doc/comment).

### Commonly used `p/` packages

To better understand how packages work, let's look at a few commonly
used ones. Some of the most commonly used packages live in the
[`examples`](https://github.com/gnolang/gno/tree/master/examples/)
folder on the monorepo, and under the `gno.land/p/demo` on-chain path.

#### Package `avl`

Deployed under `gno.land/p/demo/avl`, the AVL package provides a tree structure
for storing data. Currently, the AVL package is used to replace the functionality
of the native `map` in Gno, as maps are not fully deterministic and thus do not
work as expected in the language. Here is how using the AVL package from your
realm might look like:

```go
package myrealm

import (
	"gno.land/p/demo/avl"
)

// This AVL tree will be persisted after transaction calls
var tree *avl.Tree

func Set(key string, value int) {
	// tree.Set takes in a string key, and a value that can be of any type
	tree.Set(key, value)
}

func Get(key string) int {
  // tree.Get returns the value at given key in its raw form, 
  // and a bool to signify the existence of the key-value pair
  rawValue, exists := tree.Get(key)
  if !exists {
	  panic("value at given key does not exist")
  }
  
  // rawValue needs to be converted into the proper type before returning it
  return rawValue.(int)
}
```

View the package on the Portal Loop network [here](https://gno.land/p/demo/avl),
or on GitHub, [here](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/demo/avl).

#### Package `ufmt`

Deployed under `gno.land/p/demo/ufmt`, the `ufmt` package is a minimal version of
the `fmt` package. From [`ufmt.gno`](https://gno.land/p/demo/ufmt/ufmt.gno):

```go
// Package ufmt provides utility functions for formatting strings, similarly
// to the Go package "fmt", of which only a subset is currently supported
// (hence the name µfmt - micro fmt).
package ufmt
```

View the package on the Portal Loop network [here](https://gno.land/p/demo/ufmt),
or on GitHub, [here](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/demo/ufmt).

#### Package `seqid`

Deployed under `gno.land/p/demo/seqid`, the `seqid` package provides a simple
way to have sequential IDs in Gno. Its encoding scheme is based on the `cford32`
package. From [`seqid.gno`](https://gno.land/p/demo/seqid/seqid.gno):

```go
// Package seqid provides a simple way to have sequential IDs which will be
// ordered correctly when inserted in an AVL tree.
//
// Sample usage:
//
//	var id seqid.ID
//	var users avl.Tree
//
//	func NewUser() {
//		users.Set(id.Next().String(), &User{ ... })
//	}
package seqid
```

View the package on the Portal Loop network [here](https://gno.land/p/demo/seqid),
or on GitHub, [here](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/demo/seqid).

