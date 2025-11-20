# Understanding Gno Packages

In Gno.land, code is organized into packages that are stored on-chain. This
guide explains the different types of packages, how they're organized, and how
to work with them.

## Package Types

Gno has three fundamental package types:

### Pure Packages (`/p/`)

Pure packages are stateless Gno libraries meant to be reused by other Gno
code. Here are the defining features of pure packages:
- Don't maintain state between calls
- Can be imported by both realms and other pure packages
- Are stored under paths beginning with `/p/`
- Can be written & deployed to the chain by anyone, permissionlessly
- Users cannot call functions in pure packages directly
- Documentation should be contained within package code as comments, following the [Go doc standard](https://tip.golang.org/doc/comment)

Example: `gno.land/p/nt/avl` (An AVL tree implementation)

### Realms (`/r/`)

[Realms](./realms.md) are stateful applications (smart contracts) that can:
- Maintain persistent state between transactions
- Expose functions for interaction
- Render web content
- Import pure packages and use their functionality
- Are stored under paths beginning with `/r/`

Example: `gno.land/r/demo/boards` (A discussion forum application)

For more details on realms, see the dedicated [Realms](./realms.md) documentation.

### Ephemeral Packages (`/e/`)

Ephemeral packages are temporary, user-executed code that:
- Are created dynamically when users run `gnokey maketx run`
- Have the pattern `domain/e/{user-address}/run`
- Execute in the user's realm context
- Can call both crossing and non-crossing functions
- Are not stored on-chain
- Allow complex interactions that aren't possible with simple `maketx call`

Example: `gno.land/e/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/run` (A user's run script)

For more details on ephemeral packages and the `maketx run` command, see [Interacting with gnokey](../users/interact-with-gnokey.md#run).

## Package Path Structure

A package path is a unique identifier for any package that lives on the Gno.land
blockchain. It consists of multiple parts separated with `/` and follows this
structure:

```
gno.land/[r|p|e]/[namespace]/[package-name]
          │      │          │
          │      │          └── Name of the package
          │      └── Namespace (often a username)
          └── Type (realm, pure package, or ephemeral)
```

For example:
- `gno.land/r/gnoland/home` is the gno.land home realm
- `gno.land/r/leon/hor` is the Hall of Realms
- `gno.land/p/nt/avl` is the AVL tree package
- `gno.land/e/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/run` is a user's ephemeral run script

The components of these paths are:
- `gno.land` is the chain domain. Currently, only `gno.land` is supported, but the ecosystem may expand in the future.
- `p`, `r`, or `e` declare the type of package found at the path. `p` stands for pure package, `r` represents [realm](./realms.md), and `e` represents ephemeral package.
- `demo`, `gnoland`, etc., represent namespaces as described below.
- `home`, `hof`, `avl`, `run`, etc., represent the package name found at the path.

Two important facts about package paths:
- The maximum length of a package path is `256` characters.
- A realm's address is directly derived from its package path, by using [`chain.PackageAddress()`](./gno-stdlibs.md#derivepkgaddr)

## Namespaces

Namespaces provide users with the exclusive ability to publish code under their
designated identifiers, similar to GitHub's user and organization model. For
detailed information on how to register and use namespaces,
see [Users and Teams](./users-and-teams.md).

Initially, all users are granted a default namespace with their address - a
pseudo-anonymous (PA) namespace - to which the associated address can
deploy. This namespace has the following format:
```
gno.land/{p,r}/{std.Address}/**
```

For example, for address `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`, all the
following paths are valid for deployments:

- `gno.land/p/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/mypackage`
- `gno.land/r/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/myrealm`
- `gno.land/p/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/mypackage/subpackage/package`
- `gno.land/r/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/subpackage/realm`

Apart from package names, developers can define subpackages to further organize
their code, as seen in the example above. Packages can have any varying level of
depth as long as the full package path doesn't exceed `256` characters.

### Registering a custom namespace

To register a custom namespace:

1. Register a username at `gno.land/r/gnoland/users`
2. Once registered, you can deploy packages under that namespace
3. Only you can deploy to your namespace

This prevents impersonation and name squatting, ensuring package path authenticity.

## Importing Packages

Gno packages can import other packages using standard Go import syntax:

```go
import (
    "gno.land/p/nt/avl"          // Pure package import
    "gno.land/r/demo/users"      // Realm import (access exported functions)
)
```

## Commonly Used Pure Packages

To better understand how packages work, let's look at a few commonly used ones
from the [`examples`](https://github.com/gnolang/gno/tree/master/examples/)
folder. 

### Package `avl`

Deployed under `gno.land/p/nt/avl`, the AVL package provides a tree structure
for storing data. It replaces the functionality of the native `map` in Gno, as
maps are not fully deterministic. Usage example:

```go
package myrealm

import (
	"gno.land/p/nt/avl"
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

View the package on the [Staging network](https://gno.land/p/nt/avl)
or on [GitHub](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/nt/avl).

### Package `ufmt`

Deployed under `gno.land/p/nt/ufmt`, this package is a minimal version of the
`fmt` package:

```go
// Package ufmt provides utility functions for formatting strings, similarly
// to the Go package "fmt", of which only a subset is currently supported
// (hence the name µfmt - micro fmt).
package ufmt
```

View the package on the [Staging network](https://gno.land/p/nt/ufmt) or
on [GitHub](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/nt/ufmt).

### Package `seqid`

Deployed under `gno.land/p/nt/seqid`, this package provides a simple way to
have sequential IDs in Gno:

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

View the package on the [Staging network](https://gno.land/p/nt/seqid) or
on [GitHub](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/nt/seqid).

## Exploring Deployed Packages

You can explore all deployed packages using gnoweb. For example, you can visit the
[`gnoland`](https://gno.land/r/gnoland/) namespace to see all packages that have 
been deployed there.

This provides transparency and allows you to learn from existing code.

## Building Your Own Packages

For detailed instructions on creating your own packages:

- For realms, see [Example Minisocial dApp](../builders/example-minisocial-dapp.md)
- For deployment, see [Deploying Gno Packages](../builders/deploy-packages.md)
