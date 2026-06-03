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

Introduced `packageValueSize(pkgName, pkgPath, fnames)` as the single source of
truth for computing a `PackageValue`'s shallow memory cost. It accounts for:

- `allocPackage` (`_allocHeap` + `unsafe.Sizeof(PackageValue{})`), which already
  includes the inline `PkgName`/`PkgPath` string **headers**
- `PkgName`/`PkgPath` **backing bytes only** (`allocStringData` = heap overhead +
  content); the 16-byte headers are deliberately not re-counted here because they
  are already part of the struct sizeof above (re-adding `allocString` would
  double-count them)
- per `FNames` filename (`fileBlockEntrySize`): the filename's `FNames` slot header
  + backing bytes (`allocStringSize`), the `FBlocks` interface slot (16), and the
  `fBlocksMap` key header (16) + value `*Block` pointer (8). These headers live in
  the slice backing arrays / map buckets, not in the struct, so they are counted.

Both `GetShallowSize()` and the allocation paths call this function, guaranteeing consistency.

### Allocation at creation

- `NewPackageValue()` (main packages) now calls `alloc.Allocate(packageValueSize(...))`.
- `NewPackage()` for non-main packages now also calls the same allocation.
- `AddFileBlock()` takes an `*Allocator` parameter and charges the incremental cost of
  adding one file block (string for filename + interface entry + map entry).

### Overflow protection retained in GetShallowSize

`overflow.Addp`/`overflow.Mulp` are kept in `packageValueSize`, `fileBlockEntrySize`, and the
`GetShallowSize()` implementations as a defense-in-depth measure. While `Allocate()` already
performs an overflow-checked addition against `maxBytes`, the size computation itself happens
before that check, so wrapping the arithmetic here guarantees a deterministic panic on overflow
rather than a silent wrap that could mis-report a size. The cost is negligible for the bounded
values involved.

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
- `AddFileBlock()` signature changed to include `*Allocator` and charges the incremental
  file-block cost; the lint path passes `fallbackAllocator` so nothing is charged there.

## Known asymmetry (lint path)

`PreprocessFiles()` (`machine.go`) is the lint/import-only path: the resulting `PackageValue`
is never persisted and runs without a transaction. It passes `fallbackAllocator` to
`AddFileBlock()`, so the per-file-block cost is not charged to the machine's allocator there.
`fallbackAllocator` has no gas meter and a `MaxInt64` budget, so it neither charges gas nor
throttles; it is master's vehicle for a valid-but-non-charging allocator, used here because
master's `Allocate` is not nil-safe (the original PR relied on a nil `*Allocator` no-op, which
master removed). Because no later `GetShallowSize()`/recount path observes the uncharged bytes
for this throwaway value, the "single source of truth, no asymmetry" invariant still holds for
every path that can persist or GC a `PackageValue`; the lint path is the one narrow, intentional
exception.
