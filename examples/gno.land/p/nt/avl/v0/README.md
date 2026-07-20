> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `avl` - Gas-efficient AVL tree

A self-balancing AVL tree for storing key-value data in Gno realms. Each node is persisted as a separate object, so operations only load `O(log n)` nodes from storage instead of the entire collection.

## Usage

```go
package myrealm

import "gno.land/p/nt/avl/v0"

// Persisted across transactions.
var tree avl.Tree

func Set(key string, value int) {
    tree.Set(key, value)
}

func Get(key string) int {
    // Get returns nil for an absent key. A stored nil value looks the same,
    // so use Has when you must tell absent from present-but-nil.
    raw := tree.Get(key)
    if raw == nil {
        panic("not found")
    }
    return raw.(int)
}

// Iterate a bounded key range, stopping early when possible. Iterating
// the whole tree with ("", "") loads every node (O(n) storage reads);
// for large or user-growable trees, paginate with the pager subpackage.
func ListRange(start, end string) {
    tree.Iterate(start, end, func(key string, value any) bool {
        // return true to stop early
        return false
    })
}
```

## API

```go
type Tree struct{ /* unexported */ }

func NewTree() *Tree

// Read
func (t *Tree) Size() int
func (t *Tree) Has(key string) bool
func (t *Tree) Get(key string) (value any) // nil if the key is absent
func (t *Tree) GetByIndex(index int) (key string, value any)
func (t *Tree) Iterate(start, end string, cb IterCbFn) bool
func (t *Tree) ReverseIterate(start, end string, cb IterCbFn) bool
func (t *Tree) IterateByOffset(offset, count int, cb IterCbFn) bool
func (t *Tree) ReverseIterateByOffset(offset, count int, cb IterCbFn) bool

// Write
func (t *Tree) Set(key string, value any) (updated bool)
func (t *Tree) Remove(key string) (value any, removed bool)

type IterCbFn func(key string, value any) bool

type ITree interface { /* same shape as Tree's methods */ }
```

The zero value of `Tree` is a usable empty tree. `Get` returns `nil` for an absent key, so use `Has` to distinguish a stored `nil` value from a missing one. `Iterate` uses `[start, end)` (start inclusive, end exclusive); empty strings mean unbounded. Callbacks return `true` to stop early.

## Notes

- `avl.Tree` and `bptree` (`gno.land/p/nt/bptree/v0`) expose the same `ITree` interface; bptree swaps AVL balancing for a B+ layout with better cache locality. `seqid` (`gno.land/p/nt/seqid/v0`) generates ordered keys usable in either.
- Never return the live `*Tree` from a realm getter: a caller can then call `Set`/`Remove` on it under your realm's authority (readonly taint does not block method dispatch). Return values, copies, or a read-only `rotree` view.

## Subpackages

- `gno.land/p/nt/avl/v0/pager` - pagination helper for trees and lists.
- `gno.land/p/nt/avl/v0/rotree` - read-only view of a `Tree`.

## Why AVL over Map?

In Gno, the choice between `avl.Tree` and `map` is about how data is persisted.

**Maps** are stored as a single monolithic object. Accessing *any* value loads the *entire* map. A map with 1,000 entries loads all 1,000 on every read.

**AVL trees** store each node as a separate object. Accessing a value loads only the nodes along the search path — `~log2(n)`. A tree with 1,000 entries loads ~10 nodes; a tree with 1,000,000 entries still loads only ~20.

### Storage comparison (1,000 entries)

**Map:**

```
Object :4 = map{
  ("0" string):("123" string),
  ("1" string):("123" string),
  ...
  ("999" string):("123" string)
}
```
- `map["100"]` loads object `:4` — all 1,000 pairs.
- Gas cost proportional to total map size.

**AVL tree:**

```
Object :6  = Node{key="4",   height=10, size=1000, left=:7,  right=...}
Object :9  = Node{key="2",   height=9,  size=334,  left=:10, right=...}
Object :11 = Node{key="14",  height=8,  size=112,  left=:12, right=...}
Object :13 = Node{key="12",  height=6,  size=46,   left=:14, right=...}
Object :15 = Node{key="11",  height=5,  size=24,   left=:16, right=...}
Object :17 = Node{key="102", height=4,  size=13,   left=:18, right=...}
Object :19 = Node{key="100", height=3,  size=5,    left=:30, right=...}
Object :31 = Node{key="101", height=1,  size=2,    left=:32, right=...}
Object :33 = Node{key="100", value="123", height=0, size=1}
```
- `tree.Get("100")` loads ~10 objects (the search path only).
- Gas cost proportional to `log2(n)`.

## Further reading

- [Why should you use an AVL tree instead of a map?](https://howl.moe/posts/2024-09-19-gno-avl-over-maps/)
- [Berty's AVL scalability report](https://github.com/gnolang/hackerspace/issues/67) - testing up to 20M entries
- [Effective Gno - Prefer avl.Tree over map](https://docs.gno.land/resources/effective-gno#prefer-avltree-over-map-for-scalable-storage)
- [Wikipedia - AVL tree](https://en.wikipedia.org/wiki/AVL_tree)
