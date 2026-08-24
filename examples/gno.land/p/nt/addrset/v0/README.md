> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `addrset` - Set of addresses

A set of blockchain addresses kept in sorted order, backed by `gno.land/p/nt/bptree/v0`, with a read-only view type for safe cross-realm exposure.

It mirrors the `gno.land/p/moul/addrset` API on a B+ tree backing. A B+ tree packs many entries per persisted node, so a stored address costs ~0.9 KB against the one-node-per-entry AVL backing's ~2.0 KB — 2.2x asymptotically, 1.6x at 10 entries, with insert gas ~2.1x lower. Prefer this package when the set is part of persisted realm state.

## Usage

```go
package myrealm

import "gno.land/p/nt/addrset/v0"

// The zero value is an empty, usable set.
var members addrset.Set

func Join(cur realm, addr address) bool {
    return members.Add(addr) // true if newly added
}

func IsMember(addr address) bool {
    return members.Has(addr)
}

func List(offset, count int) []address {
    var out []address
    members.IterateByOffset(offset, count, func(addr address) bool {
        out = append(out, addr)
        return false // return true to stop early
    })
    return out
}
```

Expose the set to other realms through the read-only view, never the `*Set` itself:

```go
func Members() *addrset.ReadonlySet {
    return members.Readonly()
}
```

## API

```go
type Set struct{ /* unexported */ }

// Read
func (s *Set) Size() int
func (s *Set) Has(addr address) bool
func (s *Set) IterateByOffset(offset, count int, cb func(addr address) bool)
func (s *Set) ReverseIterateByOffset(offset, count int, cb func(addr address) bool)

// Write
func (s *Set) Add(addr address) bool    // true if newly added
func (s *Set) Remove(addr address) bool // true if it was present

// Read-only view
func (s *Set) Readonly() *ReadonlySet
func NewReadonlySet(s *Set) *ReadonlySet

type ReadonlySet struct{ /* unexported */ }

func (r ReadonlySet) Size() int
func (r ReadonlySet) Has(addr address) bool
func (r ReadonlySet) IterateByOffset(offset, count int, fn func(addr address) bool) (stopped bool)
func (r ReadonlySet) ReverseIterateByOffset(offset, count int, fn func(addr address) bool) (stopped bool)
```

Iteration is in sorted address order; `ReverseIterateByOffset` counts its offset from the end. Callbacks return `true` to stop early. `Set`'s iterators have no return value, while `ReadonlySet`'s report whether iteration was stopped that way.

## Notes

- **Do not mutate the set from inside an iteration callback** — no `Add` or `Remove`. The B+ tree mutates in place; the AVL-backed `p/moul/addrset` tolerated this because its copy-on-write nodes made the walk independent of later writes.
- **Do not copy a non-zero `Set` by value.** The copies share live tree nodes while their roots and sizes diverge. Copies of the AVL-backed set were independent snapshots; these are not.
- **Return `ReadonlySet`, never `*Set`, across a realm boundary.** A `*Set` handed to another realm can be mutated with `Add`/`Remove` under your realm's authority. `ReadonlySet` exposes no mutators and holds the `*Set` in an unexported field, so a foreign realm can neither reach it nor call through it.
- `ReadonlySet` is a live handle, not a snapshot: reads through it always reflect the set's current contents.
- There is deliberately no `Tree()` escape hatch, so the backing store cannot leak.
- The zero value of `Set` is an empty, usable set — no constructor needed.
