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
        return ObjectID{}  // <- zero for all store-restored objects!
    }
    return oi.owner.GetObjectID()
}
```

This broke `markDirtyAncestors` in `realm.go`. When a child object (e.g., a MapValue)
is modified, `markDirtyAncestors` walks up the ownership chain to mark all ancestors
dirty so they are re-saved. But for store-restored objects, `GetOwnerID()` returned
zero, stopping the walk at the first hop. Ancestors were never re-saved, leaving stale
data in the store.

Example ownership chain:

```
Block :2  (contains RefValue{OID:3, Hash:hash3})
  +-- HeapItemValue :3  (contains RefValue{OID:4, Hash:hash4})
       +-- MapValue :4  <- modified by main()
```

Without the fix: only `:4` is re-saved with its new hash. `:3` and `:2` retain stale
bytes in the store, including stale `RefValue.Hash`, stale `ModTime`, and stale
serialized content.

### Scope of Impact

The stale ancestor hashes are a **correctness invariant violation**. Specifically:

- **No validator disagreement (no fork risk)**: All validators produce the same state
  deterministically — the bug is in the save logic, not in non-deterministic evaluation.
- **Fix changes the app hash**: Although the bug doesn't cause non-determinism, the fix
  *does* change which objects enter the save set. Escaped ancestors (e.g., the
  PackageValue at the root of the ownership chain) are now correctly re-saved with
  updated hashes. Since escaped objects are indexed in the IAVL Merkle tree, the IAVL
  root — and therefore the app hash — changes from what the buggy code produced. See
  `apphash_crossrealm38_test.go` which pins the new post-fix commitment. A network
  upgrade that includes this fix will produce a different app hash for any block that
  modifies a store-restored realm object.
- **No value retrieval impact**: objects are loaded by `ObjectID`; `RefValue.Hash` is
  written during save but never checked during load (`fillValueTV` ignores it).
- **Hash chain inconsistency**: parent objects contain `RefValue{OID, Hash}` where `Hash`
  refers to the child's old content. This violates the invariant that a parent's stored
  bytes should reflect the current state of its children.
- **Forward-compatibility risk**: the hash chain is infrastructure for future Merkle proofs
  over realm state. If proofs are implemented without this fix, they would be broken for
  any object modified after store restoration.
- **Stale ModTime/bytes**: unsaved ancestors also miss `ModTime` updates and any other
  structural changes that should have been persisted.

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
- The `RefValue.Hash` chain is now consistent: modifying a child object correctly
  propagates hash changes up through all ancestors to the package block. This maintains
  the invariant needed for future Merkle proof support over realm state.

## Key Files

| File | Change |
|------|--------|
| `gnovm/pkg/gnolang/ownership.go` | `GetOwnerID()` returns `oi.OwnerID` directly |
| `gnovm/pkg/gnolang/realm.go` | `getOwner()` uses `GetObjectSafe`, nil-checks before `SetOwner` |
| `gnovm/pkg/gnolang/realm_test.go` | `TestMarkDirtyAncestors_HashConsistency` proves the fix |
| `gnovm/tests/files/*.gno` | Updated expected outputs (ancestor re-save entries) |
| `gno.land/pkg/integration/testdata/*.txtar` | Gas adjustments for increased persistence |
