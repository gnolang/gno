# PR5285: Correctly Retrieve Owner of an Object

## Context

When realm objects are persisted (serialized to the store), each object's `ObjectInfo`
contains two owner-related fields:

- `OwnerID ObjectID` — persisted, serialized with amino
- `owner Object` — transient pointer, NOT serialized

After deserialization, `owner` is always `nil`, but `OwnerID` retains the correct value.

The old `GetOwnerID()` implementation went through the transient pointer:

```go
func (oi *ObjectInfo) GetOwnerID() ObjectID {
    if oi.owner == nil {
        return ObjectID{}  // ← zero for all store-restored objects!
    }
    return oi.owner.GetObjectID()
}
```

This broke `markDirtyAncestors` in `realm.go`. When a child object (e.g., a MapValue)
is modified, `markDirtyAncestors` walks up the ownership chain to mark all ancestors
dirty so their hashes are re-computed. But for store-restored objects, `GetOwnerID()`
returned zero, stopping the walk at the first hop. Ancestors were never re-saved,
leaving stale hashes in the `RefValue{ObjectID, Hash}` chain — a Merkle inconsistency.

Example ownership chain:

```
Block :2  (RefValue{Hash, OID:3})
  └─ HeapItemValue :3  (RefValue{Hash, OID:4})
       └─ MapValue :4  ← modified by main()
```

Without the fix: only `:4` gets a new hash. `:3` and `:2` keep stale hashes.

## Decision

### Fix 1: Use the persisted `OwnerID` field directly

```go
func (oi *ObjectInfo) GetOwnerID() ObjectID {
    return oi.OwnerID
}
```

This is safe because `SetOwner()` always updates both `OwnerID` and `owner` in tandem,
so they are consistent for in-memory objects. For store-restored objects, `OwnerID` is
the only reliable source.

### Fix 2: Use `GetObjectSafe` in `getOwner()`

Now that `GetOwnerID()` reliably returns persisted IDs, there's a new edge case: the
referenced owner may have been deleted (e.g., a slice backing array replaced by `append`)
while the child's `OwnerID` still references it. Using `GetObjectSafe` instead of
`GetObject` prevents a panic and gracefully stops the ancestor walk.

## Alternatives Considered

1. **Hydrate `owner` pointer on deserialization** — Would require store access during
   unmarshaling, adding complexity to the deserialization path. The persisted `OwnerID`
   field already has the information we need.

2. **Lazy-load `owner` in `GetOwnerID()`** — Would require passing a `Store` to
   `GetOwnerID()`, changing the interface for all callers. Unnecessary since `OwnerID`
   is already correct.

## Consequences

- `markDirtyAncestors` now correctly walks the full ownership chain for store-restored
  objects, causing more objects to be re-saved per transaction. This increases gas
  consumption slightly (reflected in updated integration test gas values).
- Filetest expected outputs gain additional `u[oid]=` entries showing ancestors being
  re-saved with updated `ModTime` and `Hash` values.
- Merkle hash chain is now consistent: modifying a child object correctly propagates
  hash changes up through all ancestors to the package block.

## Key Files

| File | Change |
|------|--------|
| `gnovm/pkg/gnolang/ownership.go` | `GetOwnerID()` returns `oi.OwnerID` directly |
| `gnovm/pkg/gnolang/realm.go` | `getOwner()` uses `GetObjectSafe`, nil-checks before `SetOwner` |
| `gnovm/pkg/gnolang/realm_test.go` | `TestMarkDirtyAncestors_HashConsistency` proves the fix |
| `gnovm/tests/files/*.gno` | Updated expected outputs (ancestor re-save entries) |
| `gno.land/pkg/integration/testdata/*.txtar` | Gas adjustments for increased persistence |
