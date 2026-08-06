# ADR: Soft Version Warnings for Package Versioning

## Context

Gno packages support versioning via path suffixes (`gno.land/p/demo/avl/v0`,
`v1`, `v2`, etc.). There is no enforcement that versions are deployed
sequentially — a user could deploy `v0` then `v5`, skipping `v1` through `v4`.

On-chain enforcement of sequential versioning was considered but rejected
because **IBC package synchronization delivers packages in arbitrary order**.
If chain A deploys `v1`, `v2`, `v3` and syncs to chain B, the packages may
arrive as `v3`, `v1`, `v2`. Enforcing sequential deployment would break
cross-chain package sync.

However, large version gaps are almost always accidental (typos, copy-paste
errors), and since **on-chain packages cannot be deleted**, deploying `v100`
when you meant `v1` permanently pollutes the namespace.

## Decision

Implement **soft warnings** in tooling (CLI and web UI) without any on-chain
enforcement. The chain remains fully permissive.

### New ABCI query: `vm/qlatestversion`

A new query endpoint that, given a base package path (e.g.
`gno.land/p/demo/avl`), returns:

```json
{"latest": "v5", "first_missing": "v1", "missing": 3}
```

- `latest` — highest deployed version string
- `first_missing` — first gap in the version sequence (omitted if none)
- `missing` — count of missing versions between v0 and latest

The missing count is computed as `(latest + 1) - deployedCount`, which is
O(deployed) rather than O(latest), making it safe even with large version
numbers.

### New helper: `ParseVersionSuffix()`

Added to `gnovm/pkg/gnolang/mempackage.go`. Extracts the version number from
a package path: `gno.land/p/demo/avl/v2` → `("gno.land/p/demo/avl", 2, true)`.
Handles integer overflow gracefully by returning `ok=false`.

### gnokey `maketx addpkg` warnings

When deploying a versioned package (v1+):

- **Informational warning** (stderr): printed when the previous version
  (v(N-1)) doesn't exist on-chain. Does not block deployment.
- **Hard block** (error, requires `--force`): when the version gap exceeds 5.
  This catches accidental large version numbers since on-chain packages can't
  be deleted. Network/RPC errors are silently ignored so offline usage works.

## Alternatives Considered

- **On-chain enforcement:** Rejected due to IBC ordering constraint. Packages
  synced via IBC arrive in arbitrary order; enforcing sequential deployment
  would break this.
- **Listing all missing versions in the response:** Rejected to avoid huge
  responses. Only the count and first missing version are returned.
- **CI lint for `examples/`:** Out of scope — already being addressed in
  another PR.
- **Blocking all gaps (not just > 5):** Too aggressive. Small gaps (1-5) get
  a warning but are allowed since they may be intentional (e.g., reserving
  version numbers).

## Consequences

- **Positive:** Users get immediate feedback about version gaps before and
  after deployment, reducing accidental namespace pollution.
- **Positive:** The chain remains fully permissive, preserving IBC
  compatibility and allowing intentional non-sequential versioning.
- **Positive:** The `--force` flag provides an escape hatch for legitimate
  large version gaps.
- **Trade-off:** The `vm/qlatestversion` query adds a new ABCI endpoint to
  maintain. However, it follows existing patterns (`vm/qstorage`,
  `vm/qpaths`) and is straightforward.
