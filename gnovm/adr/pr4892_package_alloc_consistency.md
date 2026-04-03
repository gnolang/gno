# PR4892: PackageValue Allocation Consistency

## Context

`PackageValue.GetShallowSize()` is used in two contexts:

1. **Store loading** (`store.go:487-489`): charges allocation for objects deserialized from store.
2. **GC recounting**: re-sums surviving object sizes after garbage collection.

The original PR added `PkgName` string bytes to `GetShallowSize()` (fixing a missing field)
and added overflow protection. A follow-up commit expanded `GetShallowSize()` to also count
`FNames`, `FBlocks`, and `fBlocksMap`. However, these fields were **never allocated** during
the creation path:

- `AddFileBlock()` appends to `FNames`/`FBlocks` and inserts into `fBlocksMap` with no
  allocation call.
- Non-main packages constructed in `PackageNode.NewPackage()` skipped `AllocatePackageValue()`
  entirely (comment: "other packages are allocated while loading from store").
- `getFBlocksMap()` had an `// XXX, pass in allocator` comment acknowledging the gap.

This created an asymmetry: GC recounting and store-loading would report larger sizes than
were ever charged during creation, violating the invariant that `GetShallowSize()` must
return the same amount that was originally allocated.

## Decision

### Shared size function

Introduced `packageValueSize(pkgName, pkgPath, fnames, fblocksLen)` as the single source of
truth for computing a `PackageValue`'s shallow memory cost. It accounts for:

- `allocPackage` (struct base + pointer + header)
- `PkgName` and `PkgPath` string content (via `allocString + allocStringByte * len`)
- `FNames` entries (string header + content per filename)
- `FBlocks` entries (`_allocValue` per interface slot)
- `fBlocksMap` entries (`_allocName + _allocPointer` per entry, derived 1:1 from `FNames`)

Both `GetShallowSize()` and the allocation paths call this function, guaranteeing consistency.

### Allocation at creation

- `NewPackageValue()` (main packages) now calls `alloc.Allocate(packageValueSize(...))`.
- `NewPackage()` for non-main packages now also calls the same allocation.
- `AddFileBlock()` takes an `*Allocator` parameter and charges the incremental cost of
  adding one file block (string for filename + interface entry + map entry).

### Overflow protection removed from GetShallowSize

`overflow.Addp`/`overflow.Mulp` calls were removed from `GetShallowSize()`. These were
unnecessary because:

- During store loading, the sizes were already validated by `Allocate()`.
- During GC recounting, sizes cannot exceed what was previously allocated.
- The shared `packageValueSize` function uses simple arithmetic, which is safe for the
  bounded values involved.

### fBlocksMap sizing based on FNames

`fBlocksMap` is a derived runtime cache (not serialized). Its entry count is always 1:1
with `FNames`. The size function uses `len(fnames)` rather than `len(fBlocksMap)` to
correctly account for map entries even when called during store loading (where `fBlocksMap`
is nil but `FNames` is populated).

## Key files

| File | Role |
|------|------|
| `gnovm/pkg/gnolang/alloc.go` | `packageValueSize()`, `GetShallowSize()`, `NewPackageValue()` |
| `gnovm/pkg/gnolang/nodes.go` | `NewPackage()` — now allocates for non-main packages |
| `gnovm/pkg/gnolang/values.go` | `AddFileBlock()` — now takes `*Allocator` for incremental allocation |
| `gnovm/pkg/gnolang/machine.go` | Call sites passing allocator to `AddFileBlock()` |
| `gnovm/pkg/gnolang/store.go:487` | Store loading path that uses `GetShallowSize()` for allocation |

## Alternatives considered

1. **Revert to `return allocPackage`**: simplest fix, but defers field-level tracking
   indefinitely and under-counts memory during store loading.
2. **Keep detailed GetShallowSize without fixing creation path**: the approach before this
   fix — creates the asymmetry the reviewer flagged.

## Consequences

- Allocation and GC recounting are now consistent for `PackageValue`.
- Gas costs increased slightly due to the additional allocations (string content for
  `PkgName`/`PkgPath` and per-file-block charges), reflected in updated test golden values.
- `AddFileBlock()` signature changed to include `*Allocator` (nil-safe, no-ops when nil).
