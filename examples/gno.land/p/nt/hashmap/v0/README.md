> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `hashmap` - Persistent map with flat storage-access cost

A persistent string-keyed map whose cost per operation does not grow with the number of entries. Entries are distributed over a fixed set of native Gno maps ("buckets"); a native map persists as a *single* object regardless of how many entries it holds, so a `Get` or `Set` loads a constant number of objects instead of the `O(log n)` an `avl.Tree` walks.

Use it for pure key-value state that is large and looked up by exact key — token ledgers, allowance tables, registries. Use `gno.land/p/nt/avl/v0` or `gno.land/p/nt/bptree/v0` when you need sorted keys, range queries, or pagination, none of which this package offers.

## Usage

```go
package myrealm

import "gno.land/p/nt/hashmap/v0"

// The zero value is NOT usable — construct with New or NewWithBuckets.
var balances = hashmap.New()

func Set(key string, value int64) {
    balances.Set(key, value)
}

func Get(key string) int64 {
    raw := balances.Get(key)
    if raw == nil {
        return 0
    }
    return raw.(int64)
}
```

For a small or very large map, size the bucket count explicitly:

```go
m := hashmap.NewWithBuckets(256) // must be a power of two
```

It also satisfies the `KV` interface of `gno.land/p/demo/tokens/grc20`, so a GRC20 ledger can be backed by it:

```go
token, ledger := grc20.NewToken("Foo", "FOO", 6, nextTokenID.Next(), cur,
    grc20.WithStorage(func() grc20.KV { return hashmap.New() }))
```

## API

```go
type Map struct{ /* unexported */ }

const DefaultBuckets = 4096

func New() *Map                     // DefaultBuckets buckets
func NewWithBuckets(count int) *Map // count must be a positive power of two

// Read
func (m *Map) Size() int
func (m *Map) Has(key string) bool
func (m *Map) Get(key string) any // nil if the key is absent
func (m *Map) Iterate(cb func(key string, value any) bool) bool

// Write
func (m *Map) Set(key string, value any) (updated bool)
func (m *Map) Remove(key string) (value any, removed bool)
```

`Get` returns `nil` for a missing key, matching `avl.Tree` and `bptree.BPTree` — use `Has` to distinguish an absent key from a key stored with a `nil` value. `Iterate` returns `true` if the callback stopped it early. `NewWithBuckets` panics if `count` is not a positive power of two.

## Sizing

The total bucket count is fixed at construction and split across a two-level directory. Larger counts shrink each leaf bucket; because the array-decode cost is paid on the small per-level arrays rather than one flat array, over-sizing is far less punishing here than it would be with a flat bucket array.

| expected entries | recommended buckets |
|---|---|
| < 1,000 | 256 |
| 1k – 1,000k | 4096 (default — flat across this whole range) |
| > ~1,000k | 16384 |

## Notes

- **Buckets live in a two-level directory**, not one flat array: `≈√B` pages of `≈√B` buckets, allocated lazily on first write. A flat `B`-slot array is itself one object that every operation must decode in full; splitting it means an operation decodes two small arrays and one bucket instead. Untouched pages cost nothing, so an empty map is cheap.
- **Measured effect.** On a cold GRC20 transfer (fresh object cache — what a validator sees after cache GC), an `avl`-backed ledger costs ~27.9M gas at 20k holders and ~35M at 1M; the same transfer over `hashmap` costs ~16.2M and ~18.8M — flat where `avl` climbs. See `gno.land/adr/pr5965_hashmap_ledger.md`.
- **No ordering and no pagination.** Iteration order is deterministic (page, then bucket, then per-bucket insertion order) and survives a store reload, but it is not sorted by key, and `Iterate` has no cursor — you can stop early but not resume, so a full paged listing costs `O(n²)`. Realms that page through their state should use `avl` or `bptree`.
- **The persisted image depends on insertion order.** Two maps holding the same entries inserted in a different order serialize differently. This is harmless for consensus, since every node replays the same operations, but the state root is not a function of the key set alone — do not build consensus logic on top of `Iterate`.
- **Bucket selection uses unkeyed SHA-256, so keys are grindable.** A party who can create entries at will can concentrate them in one bucket and raise the per-operation cost for every key that hashes there, since an operation decodes and rewrites that whole bucket. The attacker's own cost is quadratic — each insert rewrites the growing bucket — so this degrades rather than bricks, but the effect is permanent and targeted. No per-map secret is possible in a deterministic VM. Size the bucket count so entries-per-bucket stays small, and prefer `avl` where adversarial key concentration is a real threat: its `O(log n)` holds regardless of key distribution.
- **The bucket count and hash are frozen once a map is persisted.** Placement is a function of the digest and the split, so changing either relocates every key. Pick the count at construction and treat it as permanent.
- **The zero value is not usable** — `var m hashmap.Map` has no directory allocated, so `Get`, `Set`, `Has` and `Remove` panic on it (`Size` and `Iterate` happen to return empty). Always construct with `New` or `NewWithBuckets`. (This differs from `avl.Tree` and `bptree.BPTree`, whose zero values work, so it is a real difference when swapping backends.)
- Never return the live `*Map` from a realm getter: a caller can then call `Set`/`Remove` on it under your realm's authority. Return values or a read-only view instead.
- `Remove` keeps emptied buckets and pages allocated, to avoid churning the parent objects on delete-heavy workloads.
