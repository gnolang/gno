# Storage Charging: Design Review

> **HISTORICAL.** This review was written against the pre-refactor storage-gas
> model (see [STORAGE_CHARGING.md](STORAGE_CHARGING.md)). The issues flagged
> below — especially P2 (read/write cost parity), P3 (no flat I/O cost), and
> P9 (depth invisible to gas) — are resolved by the gas refactor landed on
> this branch; `cache.Store` now charges a depth-aware I/O cost and flat
> constants are calibrated per-backend. See `gno.land/adr/gas_refactor.md`
> for the current design. This document is kept for historical context.

Review of the current storage gas and deposit design documented in
[STORAGE_CHARGING.md](STORAGE_CHARGING.md).

---

## What's Good

### 1. Clean separation of gas and deposits

Gas and deposits serve genuinely different purposes — one-time execution cost
vs. ongoing state rent — and they're independently parameterized. Gas can be
tuned without affecting deposit economics and vice versa.

### 2. The VM bypasses the SDK gas wrapper correctly

Using `ctx.Store()` instead of `ctx.GasStore()` avoids double-charging. The
GnoVM store has domain-specific knowledge (object vs type vs realm vs
mempackage) that the generic SDK gas config can't express, so having its own
`GasConfig` makes sense.

### 3. Incremental diff tracking via `sumDiff`

`SetObject` returns `len(hash+bz) - LastObjectSize`, so updates that grow an
object only charge for the growth. Shrinking an object gives negative diff.
This is correct incremental accounting without needing to diff the entire store.

### 4. Deposits are refundable

This creates proper incentives to clean up state. The `RestrictedDenoms` escape
hatch (redirecting refunds to `StorageFeeCollector` during token lock) is a
reasonable policy lever.

### 5. Per-realm deposit accounting

Each realm independently tracks `Deposit` and `Storage`, with a derived deposit
address. One realm's storage costs don't leak into another's accounting.

---

## Problems

### P1. MemPackages and Types are not deposit-tracked

`AddMemPackage` and `SetType` write to the backend stores and charge gas, but
they do NOT feed into `realm.sumDiff`. Only `SetObject`/`DelObject` (via realm
finalization) participate in deposit tracking.

Consequences:
- Deploying a 100KB package via `MsgAddPackage` pays gas (`GasAddMemPackage`
  8/byte = 800K gas) but locks **zero deposit** for the package source code.
  The source persists forever in `iavlStore` with no ongoing economic cost.
- Types written via `SetType` (52/byte gas) also incur no deposit.
- Package source code is likely the largest category of permanent, unaccounted
  storage on chain.

**Severity**: High. This is the biggest gap in the deposit model.

### P2. Read and write gas costs are identical

`GasGetObject` and `GasSetObject` are both 16/byte. In practice:
- Reads hit cache frequently, don't require serialization, and don't trigger
  IAVL tree mutations.
- Writes involve amino marshal, hash computation, `baseStore.Set`, and possible
  `iavlStore.Set` for escaped objects.

Read-heavy workloads (common for queries and cross-realm calls) are overcharged.
Write-heavy workloads are undercharged. For comparison, the SDK default config
makes writes ~10x reads (30 vs 3 per byte, plus 2000 vs 1000 flat cost).

**Severity**: Medium. Distorts gas costs for read-heavy vs write-heavy txs.

### P3. No flat cost component for store operations

Every operation except `DeleteObject` is purely per-byte with zero flat cost.
But every store operation has fixed overhead regardless of size: key
construction, map/tree lookup, cache check, hash computation (for `SetObject`).

A 1-byte object and a 1000-byte object pay very different gas, but the fixed
overhead is the same. Very small objects are undercharged for their fixed cost.

The SDK gas config gets this right with `ReadCostFlat: 1000` and
`WriteCostFlat: 2000`.

**Severity**: Medium. Creates an attack surface where many tiny objects cost
less gas than their true overhead.

### P4. PackageRealm gas (524/byte) may be disproportionately high

Realm metadata is read and written on every realm-touching transaction
(`processStorageDeposit` calls `GetPackageRealm` then `SetPackageRealm`). At
524/byte this is 32x more expensive than object get/set (16/byte).

The realm struct is small (a few hundred bytes), so in absolute terms this is
~50K-100K gas per realm-touching transaction just for bookkeeping. This could
disproportionately penalize cheap transactions (e.g., a simple state increment).

**Severity**: Medium. Needs calibration data to confirm whether 524/byte
reflects actual cost or is an outlier.

### P5. Escaped object IAVL writes are not separately gas-charged

When `SetObject` processes an escaped object (`store.go:676-680`), it performs
two backend writes:
- `baseStore.Set(key, hash||bz)` — the full object (gas-charged via
  `GasSetObject × len(bz)`)
- `iavlStore.Set(oid, hash)` — the escape index (not gas-charged)

IAVL writes are more expensive than flat key-value writes due to tree
rebalancing, but the second write gets no additional gas.

**Severity**: Low-medium. Escaped objects are a subset of all objects, and the
IAVL value is small (32 bytes), but the tree operation cost is non-trivial.

### P6. `DeleteObject` flat cost ignores object size

Deleting a 10-byte object costs the same gas (3715) as deleting a 100KB object.
The backend `Delete` cost may vary with key/tree size, and more importantly this
creates an asymmetry:

- Create 100KB object: pays `16 × 100,000 = 1,600,000` gas
- Delete same object: pays `3,715` gas (and gets deposit back)

A contract could create-then-delete large objects as a way to do cheap I/O
(1.6M gas to write, 3.7K gas to delete, deposit refunded).

**Counterargument**: This may be intentional — making deletes cheap encourages
state cleanup. But the 430x asymmetry at 100KB seems too large.

**Severity**: Low-medium. The gas is still burned on creation, so the total
round-trip cost is dominated by the write.

### P7. Gas is charged on amino bytes, deposits on hash+amino bytes

Gas for `SetObject` is `16 × len(bz)` where `bz` is the amino bytes. The
deposit diff is `len(hash) + len(bz) - LastObjectSize`, where `LastObjectSize`
includes the hash (20 bytes, `HashSize`).

Gas and deposits measure slightly different byte quantities per object. This is
a ~20-byte systematic discrepancy per object — negligible for large objects but
up to ~50% error for very small ones.

**Severity**: Low. Mostly a consistency issue.

### P8. `AddMemPackage` index write is not gas-charged

`AddMemPackage` charges `8 × len(bz)` once but performs two writes:
- `baseStore.Set(idxkey, pkgPath)` — the index entry (not gas-charged)
- `iavlStore.Set(pathkey, bz)` — the full package (gas-charged)

The index write is small but free.

**Severity**: Low. Negligible in practice.

### P9. IAVL tree depth is invisible to gas metering

IAVL tree operations (Get/Set/Delete) traverse O(log n) nodes where n is the
total number of keys. Each level may require a disk seek for cold paths. As the
chain grows, the real cost of these operations increases logarithmically, but
gas stays constant.

**Store layout** (`gno.land/pkg/gnoland/app.go:102-103`):

| Store key | Backend | Used for |
|-----------|---------|----------|
| `mainKey` | IAVL tree | auth accounts, bank balances, escaped object hashes, MemPackages |
| `baseKey` | Flat DB (dbadapter) | objects, types, block nodes, realm metadata |

Both stores sit on top of the backing DB, which defaults to **PebbleDB**
(`tm2/pkg/bft/config/config.go:338`) — a pure-Go LSM-tree (CockroachDB's
Pebble). `goleveldb` and `boltdb` are also supported.

The `baseStore` is a `dbadapter` — a thin wrapper that maps Get/Set/Delete
directly to the backing DB with no additional tree layer. However, because the
backing DB is an LSM-tree (not a hash table), reads are not truly O(1): they
traverse memtable → L0 → L1 → ... → Ln, mitigated by bloom filters and block
cache. Writes are O(1) amortized (append to WAL + memtable) but trigger
background compaction. In practice, PebbleDB point lookups are fast and
predictable for typical key counts, but read amplification grows as the DB
accumulates more levels — especially under write-heavy workloads where
compaction falls behind.

The `iavlStore` adds a second layer of O(log n) traversal on top of the same
backing DB. Every IAVL Get/Set/Delete walks the balanced binary tree, loading
nodes from the DB at each level. The IAVL tree exposes `Height()` and `Size()`
(`tm2/pkg/iavl/immutable_tree.go:121-140`), but nothing reads them for gas.
The `gas.Store` wrapper (`tm2/pkg/store/gas/store.go`) charges flat + per-byte,
treating the tree as if it were a hash map. The GnoVM store doesn't use the
`gas.Store` wrapper at all (it calls `ctx.Store()` not `ctx.GasStore()`).

In summary, both stores have sub-O(1) read characteristics that grow with data
size, but the `iavlStore` is substantially worse due to the extra tree traversal
layer on top of the LSM lookup.

**Scale of the problem:**

| Tree keys | Height | Nodes traversed per op | Effect |
|-----------|--------|----------------------|--------|
| 1M | ~20 | ~20 | Fits in node cache, fast |
| 100M | ~27 | ~27 | Cold paths hit disk per level |
| 1B | ~30 | ~30 | Significant disk I/O per op |

Writes are worse: IAVL Set traverses down + rebalances up + persists new nodes.
Each modified node requires a disk write. The node cache masks this for hot
paths, but cold paths (infrequently accessed packages, old escaped objects) hit
disk at every level.

The practical effect: validators absorb the growing real cost as the chain
matures, since gas revenue doesn't increase to match.

**Possible mitigations:**
1. **Governance-adjusted flat costs**: periodically increase flat costs via
   param proposals as the tree grows. Simple but coarse and reactive.
2. **Tree-height-aware gas**: charge `baseCost + perLevelCost × height` using
   the already-available `root.subtreeHeight`. Self-adjusting.
3. **Pessimistic amortization**: set constants for the expected tree size at
   some future horizon (e.g., 10 years). Overcharges early, correct later.

**Severity**: Medium-high. Not urgent on a young chain, but becomes a structural
subsidy from validators to users as state grows. Should be addressed before
the tree reaches ~100M keys.

### P10. `SetObject` diff may over-count if `LastObjectSize` is stale

`store.go:614`: `diff := int64(len(hash)+len(bz)) - o2.(Object).GetObjectInfo().LastObjectSize`

For objects loaded from cache rather than backend, `LastObjectSize` could be
zero if the object was modified in-memory without a prior `SetObject` persisting
its size. This would make `diff` equal to the full object size instead of the
true delta, over-charging the deposit.

**Severity**: Low. In practice, objects loaded from backend have
`LastObjectSize` set at load time (`store.go:481`), and newly created objects
correctly start at 0. But the invariant is implicit and fragile.

---

## Summary of Priorities

| # | Issue | Severity | Fix complexity |
|---|-------|----------|---------------|
| P1 | MemPackages/Types escape deposit tracking | High | Medium — need to add diff tracking for these store paths |
| P9 | IAVL tree depth invisible to gas metering | Medium-high | Medium — tree-height-aware gas or pessimistic constants |
| P2 | Read = Write gas cost | Medium | Low — split into separate read/write constants |
| P3 | No flat cost component | Medium | Low — add flat base to each operation |
| P4 | PackageRealm 524/byte may be too high | Medium | Low — needs benchmark calibration |
| P5 | Escaped IAVL writes uncharged | Low-medium | Low — add gas charge in escape path |
| P6 | Delete flat cost vs create per-byte asymmetry | Low-medium | Low — add per-byte or size-tiered delete cost |
| P7 | Gas vs deposit byte measurement mismatch | Low | Low — use same size for both |
| P8 | AddMemPackage index write uncharged | Low | Trivial |
| P10 | LastObjectSize staleness risk | Low | Medium — needs invariant enforcement |
