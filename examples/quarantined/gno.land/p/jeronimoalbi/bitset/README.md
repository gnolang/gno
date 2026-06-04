# Bitset Package

Package implements an arbitrary-size bit set, also known as bit array.

Bit sets are useful when you need a compact way to track large sets of boolean flags,
such as permissions, feature toggles, or membership sets, using significantly less
memory than a slice of booleans while also supporting fast bulk operations across
entire sets.

## API

- `New(size uint64) BitSet` - Returns a new BitSet pre-allocated for `size` bits
- `BitSet.Set(i uint64)` - Turn on the bit at position `i`
- `BitSet.Clear(i uint64)` - Turn off the bit at position `i`
- `BitSet.ClearAll()` - Turn off all bits
- `BitSet.Compact()` - Reclaim memory by removing trailing zero words
- `BitSet.IsSet(i uint64) bool` - Check whether the bit at position `i` is set
- `BitSet.Size() int` - Return the total number of bits currently allocated
- `BitSet.Len() int` - Return the number of set bits
- `BitSet.And(other BitSet)` - In-place intersection with another set
- `BitSet.Or(other BitSet)` - In-place union with another set
- `BitSet.Xor(other BitSet)` - In-place symmetric difference with another set
- `BitSet.Equal(other BitSet) bool` - Check whether two sets have the same bits set
- `BitSet.String() string` - Return a binary (MSB-first) representation of the set
- `BitSet.PaddedString() string` - Return a zero-padded binary (MSB-first) representation of the set

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
