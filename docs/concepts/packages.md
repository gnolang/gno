---
id: packages
---

# Packages

In gno.land, packages are Gno code which is meant to be reused by other Gno code,
be it by other packages or realms. Here are some defining features of packages:
- Packages are stored on-chain under the `"gno.land/p/"` path, and can be 
written & deployed on-chain by anyone
- Packages are meant to be imported by other packages & realms
- Packages do not persist state - packages can have global variables & constants,
but any attempt to change their values will be discarded after a transaction
is completed
- Documentation for packages should be contained within package code itself,
in the form of comments, following the [Go doc standard](https://tip.golang.org/doc/comment).

To learn how to write a package,
see [How to write a simple Gno Library](../how-to-guides/simple-library.md).

## Commonly used packages

To better understand how packages work, let's look at a few commonly 
used ones. Some of the most commonly used packages live in the
[`examples`](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/demo/)
folder on the monorepo, and under the `"gno.land/p/demo"` on-chain path. 

### Package `avl`

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

### Package `ufmt`

Deployed under `gno.land/p/demo/ufmt`, the `ufmt` package is a minimal version of
the `fmt` package. From [`ufmt.gno`](https://gno.land/p/demo/ufmt/ufmt.gno):

```go
// Package ufmt provides utility functions for formatting strings, similarly
// to the Go package "fmt", of which only a subset is currently supported
// (hence the name µfmt - micro fmt).
package ufmt
```

View the package on-chain [here](https://gno.land/p/demo/ufmt), or on GitHub, 
[here](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/demo/ufmt).

### Package `seqid`
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

View the package on-chain [here](https://gno.land/p/demo/seqid), or on GitHub,
[here](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/demo/seqid).

## Packages vs Standard Libraries

Apart from packages, Gno, like Go, has standard libraries. To better
understand the difference between these two concepts, let's compare a few
specific points:
- Packages can be written and deployed by anyone at any time, while standard
libraries require thorough battle-testing and reviews by the core team & community
before being added to the language
- Standard libraries usually provide low-level necessities for the language,
while packages utilize them to create a broader range of functionality
