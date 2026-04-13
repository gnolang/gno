# Bitset Package

Package implements an arbitrary-size bit set, also known as bit array.

Bit sets are useful when you need a compact way to track large sets of boolean flags,
such as permissions, feature toggles, or membership sets, using significantly less
memory than a slice of booleans while also supporting fast bulk operations across
entire sets.

## API

- `New(size uint64) BitSet` - Returns a new BitSet pre-allocated for `size` bits

## Usage

[embedmd]:# (filetests/readme_filetest.gno go)
```go
package main

import "gno.land/p/jeronimoalbi/bitset"

const (
	PermRead   = 0
	PermWrite  = 1
	PermDelete = 2
	PermAdmin  = 3
)

func main() {
	var perms bitset.BitSet
	perms.Set(PermRead)
	perms.Set(PermWrite)

	println("Can read:", perms.IsSet(PermRead))
	println("Can write:", perms.IsSet(PermWrite))
	println("Can delete:", perms.IsSet(PermDelete))
	println("Is admin:", perms.IsSet(PermAdmin))
	println("Permissions set:", perms.Len())
}

// Output:
// Can read: true
// Can write: true
// Can delete: false
// Is admin: false
// Permissions set: 2
```
