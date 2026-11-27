# Refresh redeployed types in `SetType`

## Context

During a private realm redeployment, loading the old `PackageValue` can load
its declared types from the backend into `cacheTypes`. `ClearObjectCache` does
not clear that cache. The redeployed package can then produce a new
`DeclaredType` with the same name-based `TypeID` and a different layout.

`SetType` previously returned whenever `cacheTypes` contained the `TypeID`, so
the backend retained the old definition. A fresh store could consequently
resolve persisted objects with a stale type whose fields or methods no longer
matched the redeployed package.

## Decision

`SetType` skips persistence only when the cached type is the same in-memory
instance as the supplied type:

```go
cached, exists := ds.cacheTypes[tid]
if exists && cached == tt {
	return
}
```

The first deployment remains a cache miss. Repeated registration of the same
type within one package remains idempotent. A redeployment supplies a different
instance, so it replaces the backend entry and then replaces the cached entry.

## Alternatives considered

Remove the early return completely. This would add duplicate backend writes
for the same instances reached through uverse and package-getter paths. It
would also invalidate `copyTypeWithRefs`'s invariant that uverse
`DeclaredType` instances installed with `SetCacheType` do not enter its
declared-type copy path through `SetType`. The narrower identity check keeps
that behavior while allowing a different instance to refresh the definition.

## Consequences

- A redeployment that supplies a different type instance encodes and writes
  the type again, so its gas use can increase.
- The type backend is outside the app-hash leaves, so overwriting the same type
  key does not directly change an app-hash leaf.
- Different gas use can still change whether a gas-bound transaction succeeds,
  so this behavior change requires a coordinated upgrade.
- Same-instance registrations retain their previous no-write behavior.
