> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `bptree` - Mutable B+ tree

A mutable, in-place B+ tree for storing key-value data in Gno realms. Exposes the same `ITree` interface as `gno.land/p/nt/avl/v0` but uses a B+ tree internally — fewer pointer dereferences per operation and better cache locality, with a configurable fanout.

## Usage

```go
package myrealm

import "gno.land/p/nt/bptree/v0"

// Zero value is usable (fanout 32). Persisted across transactions.
var tree bptree.BPTree

func Set(key string, value int) {
    tree.Set(key, value)
}

func Get(key string) int {
    raw := tree.Get(key)
    if raw == nil {
        panic("not found")
    }
    return raw.(int)
}

func RangeAsc(start, end string) {
    tree.Iterate(start, end, func(key string, value any) bool {
        // return true to stop early
        return false
    })
}
```

For a different fanout, use a constructor:

```go
tree := bptree.NewBPTreeN(64) // fanout 64
```

## API

```go
type BPTree struct{ /* unexported */ }

func NewBPTree32() *BPTree            // fanout 32
func NewBPTreeN(fanout int) *BPTree   // panics if fanout < 4

// Read
func (t *BPTree) Size() int
func (t *BPTree) Has(key string) bool
func (t *BPTree) Get(key string) (value any) // nil if the key is absent
func (t *BPTree) GetByIndex(index int) (key string, value any)
func (t *BPTree) Iterate(start, end string, cb IterCbFn) bool
func (t *BPTree) ReverseIterate(start, end string, cb IterCbFn) bool
func (t *BPTree) IterateByOffset(offset, count int, cb IterCbFn) bool
func (t *BPTree) ReverseIterateByOffset(offset, count int, cb IterCbFn) bool

// Write
func (t *BPTree) Set(key string, value any) (updated bool)
func (t *BPTree) Remove(key string) (value any, removed bool)

type IterCbFn func(key string, value any) bool

type ITree interface { /* same shape as BPTree's methods */ }
```

The zero value of `BPTree` is a usable empty tree (fanout 32). `Iterate` uses `[start, end)` (start inclusive, end exclusive); `ReverseIterate` uses `[start, end]` (both inclusive). Empty strings mean unbounded. Callbacks return `true` to stop early. `GetByIndex` panics on out-of-range indices.

The tree must not be modified during iteration (no `Set` or `Remove` from the callback).

## Subpackages

- `gno.land/p/nt/bptree/v0/list` - ordered list built on top of `BPTree`.
- `gno.land/p/nt/bptree/v0/pager` - pagination helper for trees and lists.
- `gno.land/p/nt/bptree/v0/rotree` - read-only view of a `BPTree`.

## Notes

- API and semantics match `gno.land/p/nt/avl/v0` exactly — `""` is a valid key, `Get` returns `nil` for a missing key (use `Has` to distinguish a stored `nil`), and `Remove` returns `(nil, false)`.
- Never return the live `*BPTree` from a realm getter: a caller can then call `Set`/`Remove` on it under your realm's authority. Return values, copies, or a read-only `rotree` view.
- Sequential keys from `seqid` (`gno.land/p/nt/seqid/v0`) pair well here: monotonic inserts hit the append-optimized split path.
- Fanout must be `>= 4`. Higher fanouts mean shallower trees and fewer object loads per lookup, at the cost of larger individual node objects.
- Each node (leaf or inner) is persisted as a separate object, so reads only load the `O(log n)` nodes on the search path — same storage-efficiency benefit as `avl`.
- No sibling pointers or `first`/`last` shortcuts: iteration uses an ephemeral stack to keep every persisted node at ref-count 1 (avoids Gno's object-escape penalty).
