# Gas Refactor: Charge at the Store Boundary

**Status: Implemented**

## Problem

Gas is charged at the wrong layer. Both the tm2 `gas.Store` wrapper and the
GnoVM `defaultStore` charge gas on every API call, including cache hits. The
actual cost structure has multiple cache layers that absorb most accesses:

```
gas.Store           ← charges gas here (WRONG: includes cache hits)
  cache.Store       ← in-memory map
    iavl.Store      ← IAVL tree
      nodeDB        ← node LRU cache
        dbm.DB      ← actual disk I/O
```

Consequences:
- Cache hits are overcharged (a ~100ns map lookup is billed as if it were a
  ~67us disk read — a 670x overcharge).
- The tm2 gas constants (ReadCostFlat=1000, WriteCostFlat=2000, etc.) are
  uncalibrated, inherited from Cosmos SDK circa 2018 with no benchmarks.
- The GnoVM constants (GasSetType=52/byte, etc.) conflate amino
  serialization CPU cost with storage I/O cost into a single number.
- Two independent gas systems (tm2 for auth, GnoVM for VM ops) charge
  differently for equivalent operations, feeding the same meter.

## Design

### Core idea

Separate gas into two concerns:

| Gas type | What it measures | Where charged |
|----------|-----------------|---------------|
| **Storage I/O gas** | Disk reads/writes | At `cache.Store` (first layer with gctx) |
| **Compute gas** | Amino marshal/unmarshal, VM ops | At GnoVM store / machine |

Cache hits incur zero storage I/O gas. Amino serialization is charged
separately as compute gas regardless of whether the backing store was hit.

Storage I/O gas is estimated deterministically from the backing store's
properties (e.g., IAVL tree size) rather than measured from actual cache
hit/miss patterns, which would break consensus.

### GasContext

A new type in `tm2/pkg/store/types/gas.go`:

```go
type GasContext struct {
    Meter  GasMeter
    Config GasConfig
}

func (gctx *GasContext) WillGet()            // ReadCostFlat (non-depth fallback only;
                                             // depth stores use ConsumeGas directly)
func (gctx *GasContext) DidGet(bz []byte)    // ReadCostPerByte * len(bz)
func (gctx *GasContext) WillSet(bz []byte) Gas   // WriteCostFlat + WriteCostPerByte * len(bz); returns amount charged
func (gctx *GasContext) WillDelete() Gas         // DeleteCost; returns amount charged
func (gctx *GasContext) RefundGas(amount Gas)    // refunds previously charged gas
func (gctx *GasContext) ConsumeGas(amount Gas, descriptor string)  // charges gas
func (gctx *GasContext) WillIterator()       // (flat seek cost)
func (gctx *GasContext) WillIterNext()       // IterNextCostFlat
```

All methods are nil-safe: if `gctx == nil`, they are no-ops (returning 0
for methods that return `Gas`).

### DepthEstimator

A small interface for stores that have depth-dependent I/O cost:

```go
type DepthEstimator interface {
    ExpectedDepth() int64
}
```

`iavl.Store` implements it:

```go
func (st *Store) ExpectedDepth() int64 {
    size := st.tree.Size()
    if size <= 1 {
        return 1
    }
    return int64(bits.Len64(uint64(size))) // floor(log2(size)) + 1
}
```

`dbadapter.Store` does not implement it (no depth, flat I/O cost).

The estimator is propagated upward through `CacheWrap()`:
- `iavl.Store.CacheWrap()` → sets estimator to `self`
- `cache.Store.CacheWrap()` → propagates parent's estimator
- `prefix.Store.CacheWrap()` → propagates parent's estimator

`tree.Size()` is consensus state — it does not change during block
execution (cache layers buffer all writes above the IAVL tree until
`Commit()`). All transactions in a block see the same depth estimate.

| Size | ExpectedDepth | Height (worst) | Savings vs height |
|------|--------------|----------------|-------------------|
| 1K | 11 | 14 | 21% |
| 100K | 18 | 24 | 25% |
| 10M | 24 | 33 | 27% |
| 1B | 31 | 43 | 28% |

### Signature change

The `Store` interface methods gain `*GasContext` as the first parameter:

```go
type Store interface {
    Get(gctx *GasContext, key []byte) []byte
    Set(gctx *GasContext, key, value []byte)
    Has(gctx *GasContext, key []byte) bool
    Delete(gctx *GasContext, key []byte)
    Iterator(gctx *GasContext, start, end []byte) Iterator
    ReverseIterator(gctx *GasContext, start, end []byte) Iterator
    Write()      // no gctx — write cost already charged at Set() time
    CacheWrap() Store
}
```

Every implementation and call site is updated. Callers that don't care about
gas pass `nil` — the operation executes normally, just without charging.

### gas.Store is removed

With `*GasContext` as an explicit parameter, `gas.Store` is no longer needed.
It previously existed to wrap a store and inject gas charging. Now the caller
passes `gctx` directly — there is nothing to wrap.

`ctx.GasStore(key)` is replaced by two explicit calls:

```go
// Before:
stor := ctx.GasStore(ak.key)
bz := stor.Get(key)

// After:
gctx := ctx.GasContext()       // builds *GasContext from ctx.GasMeter() + GasConfig
stor := ctx.Store(ak.key)
bz := stor.Get(gctx, key)
```

`ctx.GasContext()` returns `*GasContext` (or `nil` if no gas metering is
desired). This is a method on `sdk.Context`:

```go
func (c Context) GasContext() *store.GasContext {
    if c.GasMeter() == nil {
        return nil
    }
    return &store.GasContext{
        Meter:  c.GasMeter(),
        Config: store.DefaultGasConfig(),
    }
}
```

### How each layer handles gctx

All storage I/O gas is charged in `cache.Store` — the outermost layer that
holds the `gctx`. Lower layers (`iavl.Store`, `dbadapter.Store`) are pure
storage — they do not charge gas. This makes charging symmetric for reads
and writes, and keeps gas logic in one place.

**cache.Store — depth helper:**

```go
func (cs *cacheStore) expectedDepth() int64 {
    depth := int64(1)
    if cs.depthEstimator != nil {
        depth = cs.depthEstimator.ExpectedDepth()
    }
    if cs.gasConfig.MinDepth > 0 && depth < cs.gasConfig.MinDepth {
        depth = cs.gasConfig.MinDepth
    }
    return depth
}
```

Returns 1 for non-IAVL stores (no estimator). For IAVL stores, returns
`max(ExpectedDepth(), MinDepth)`. The `> 1` check in Get/Set/Delete
distinguishes IAVL paths (depth-based) from flat store paths (single
flat cost).

**cache.Store — reads:**

```go
func (cs *cacheStore) Get(gctx *GasContext, key []byte) []byte {
    if cached { return value }        // cache hit — no gas
    // cache miss — charge depth-based I/O gas, then fetch
    depth := cs.expectedDepth()
    if depth > 1 {
        gctx.ConsumeGas(Gas(depth) * gctx.Config.ReadCostFlat, "DepthReadFlat")
    } else {
        gctx.WillGet()                // flat ReadCostFlat (no depth, e.g. dbadapter)
    }
    value := cs.parent.Get(nil, key)  // nil — parent doesn't charge gas
    gctx.DidGet(value)                // ReadCostPerByte
    cs.setCacheValue(key, value, false, false)
    return value
}
```

**cache.Store — write/delete gas deduplication:**

A cached store may see multiple writes and deletes to the same key before
flushing. Gas should reflect the final operation, not every intermediate one.

The cache tracks `chargedGas map[string]Gas` — the total gas charged for
write/delete operations on each key. On every `Set` or `Delete`, it refunds
whatever was previously charged for that key, then charges for the new
operation. The last operation wins:

```go
func (cs *cacheStore) Set(gctx *GasContext, key, value []byte) {
    k := string(key)
    if prev, exists := cs.chargedGas[k]; exists {
        gctx.RefundGas(prev)
    }
    var gas Gas
    depth := cs.expectedDepth()
    if depth > 1 {
        // depth reads to find insertion point + depth writes for COW path
        // (depth writes includes the leaf node which holds the value —
        // no separate WriteCostFlat needed)
        depthGas := Gas(depth) * (gctx.Config.ReadCostFlat + gctx.Config.WriteCostFlat)
        depthGas += gctx.Config.WriteCostPerByte * Gas(len(value))
        gctx.ConsumeGas(depthGas, "IavlSet")
        gas = depthGas
    } else {
        // flat store (e.g. dbadapter): one write
        gas = gctx.WillSet(value)  // WriteCostFlat + WriteCostPerByte * len(value)
    }
    cs.chargedGas[k] = gas
    cs.setCacheValue(key, value, false, true)
}

func (cs *cacheStore) Delete(gctx *GasContext, key []byte) {
    k := string(key)
    if prev, exists := cs.chargedGas[k]; exists {
        gctx.RefundGas(prev)
    }
    var gas Gas
    depth := cs.expectedDepth()
    if depth > 1 {
        // depth reads + depth writes to remove and rebalance
        depthGas := Gas(depth) * (gctx.Config.ReadCostFlat + gctx.Config.WriteCostFlat)
        gctx.ConsumeGas(depthGas, "IavlDelete")
        gas = depthGas
    } else {
        gas = gctx.WillDelete()  // DeleteCost
    }
    cs.chargedGas[k] = gas
    cs.setCacheValue(key, nil, true, true)
}
```

This gives exact charging — the gas consumed for a key always reflects the
final operation, regardless of how many intermediate writes/deletes occurred:

Example with IAVL-backed store (depth estimator present, R=ReadCostFlat,
W=WriteCostFlat, Wb=WriteCostPerByte):

| Sequence | Action | Running total for key |
|---|---|---|
| Set("k", 100b) | charge | depth\*(R+W) + 100\*Wb |
| Set("k", 500b) | refund prev, charge | depth\*(R+W) + 500\*Wb |
| Delete("k") | refund prev, charge | depth\*(R+W) |
| Set("k", 200b) | refund prev, charge | depth\*(R+W) + 200\*Wb |

Example with dbadapter-backed store (no depth estimator):

| Sequence | Action | Running total for key |
|---|---|---|
| Set("k", 100b) | WillSet(100b) | WriteCostFlat + 100\*Wb |
| Set("k", 500b) | refund prev, WillSet(500b) | WriteCostFlat + 500\*Wb |
| Delete("k") | refund prev, WillDelete() | DeleteCost |
| Set("k", 200b) | refund prev, WillSet(200b) | WriteCostFlat + 200\*Wb |

`WillSet` and `WillDelete` return the amount of gas charged. `RefundGas`
decreases gas consumed by that amount. The refund is always safe — it
exactly matches what was previously returned for that key. The
`GasMeter.RefundGas` implementation should floor at zero to guard against
edge cases.

`Write()` keeps its current signature (no `gctx` parameter). It does not
charge gas — the write cost was already charged at `Set()` time. `Write()`
just moves bytes between layers:

```go
func (cs *cacheStore) Write() {
    for key, value := range dirty {
        cs.parent.Set(nil, key, value)  // no gas — already paid at Set() time
    }
    cs.chargedGas = make(map[string]Gas) // reset — stale entries would cause wrong refunds
}
```

**Multi-layer cache invariant.** In gno.land, two cache layers exist per
transaction (tx-scoped wrapping block-scoped). Write gas is charged only
at the outermost layer — when the user calls `Set(gctx, ...)`. When
`Write()` flushes to the inner cache via `parent.Set(nil, ...)`, the inner
cache's `chargedGas` is never populated (gctx is nil, no charging occurs).
When the inner cache later calls `Write()`, it also passes `nil`. This is
correct — gas is charged exactly once per key, at the layer where the
user-facing `Set()` occurred.

**Transaction isolation.** Each transaction pays full depth-based I/O gas
as if it were the only transaction in the block. If tx #1 loads a key into
the block-scoped cache (cache layer 1), and tx #2 reads the same key, tx
#2's cache layer 2 misses, charges full depth-based read gas, then finds
the value in cache layer 1 without hitting disk. This is a slight
overcharge relative to actual I/O, but is the correct design:

- Gas is deterministic and **independent of transaction ordering** within
  a block. If gas depended on what prior transactions loaded, reordering
  transactions would change their gas costs.
- Block producers cannot manipulate gas by choosing transaction order.
- Gas estimation (simulate mode) is reliable — it does not depend on what
  other transactions will execute before yours.
- Each transaction is independently gas-accountable.

**iavl.Store — pure storage, no gas logic:**

```go
func (st *Store) Get(gctx *GasContext, key []byte) []byte {
    value, _ := st.tree.Get(key)
    return value  // gctx ignored — gas is charged by cache.Store above
}

func (st *Store) Set(gctx *GasContext, key, value []byte) {
    st.tree.Set(key, value)  // gctx ignored
}
```

`iavl.Store` implements `DepthEstimator` so that `cache.Store` above it
can estimate I/O cost. It does not charge gas itself.

**dbadapter.Store — pure storage, no gas logic:**

```go
func (dsa Store) Get(gctx *GasContext, key []byte) []byte {
    return dsa.db.Get(key)  // gctx ignored — gas is charged by cache.Store above
}

func (dsa Store) Set(gctx *GasContext, key, value []byte) {
    dsa.db.Set(key, value)  // gctx ignored
}
```

`dbadapter.Store` does not implement `DepthEstimator`. The `cache.Store`
above it falls back to flat cost (1x `ReadCostFlat` / `WriteCostFlat`).

### Iterators

`Iterator` and `ReverseIterator` gain `*GasContext` as the first parameter:

```go
type Store interface {
    Iterator(gctx *GasContext, start, end []byte) Iterator
    ReverseIterator(gctx *GasContext, start, end []byte) Iterator
    // ...
}
```

The returned `Iterator` stores the `gctx` and calls
`gctx.WillIterator()` on creation (flat seek cost) and
`gctx.WillIterNext()` on each `Next()` call. The per-step cost is charged
at the point of iteration (matching the current `gasIterator` behavior,
minus the wrapper).

### Full call path

**IAVL stores (auth keeper) — read:**
```
auth keeper: stor.Get(gctx, key)
  cache.Store.Get(gctx, key)
    HIT  → return value                         no gas
    MISS →
      depth = depthEstimator.ExpectedDepth()
      gctx.ConsumeGas(depth * ReadCostFlat)      depth flat reads
      value = parent.Get(nil, key)                parent doesn't charge
        iavl.Store.Get(nil, key)                  pure storage, no gas
      gctx.DidGet(value)                          ReadCostPerByte
```

**IAVL stores (auth keeper) — write:**
```
auth keeper: stor.Set(gctx, key, value)
  cache.Store.Set(gctx, key, value)
    refund prev chargedGas[key] if exists
    depth = depthEstimator.ExpectedDepth()
    gctx.ConsumeGas(depth * (R + W) + Wb * len(value))       depth reads + depth writes + per-byte
    buffer in cache                                           no parent call
```

**Has (all stores):**

`cache.Store.Has(gctx, key)` calls `self.Get(gctx, key)` — gas is charged
via the Get path. No separate Has gas logic.

**Non-IAVL stores (GnoVM baseStore via dbadapter) — read:**
```
Gno: ds.baseStore.Get(gctx, key)
  cache.Store.Get(gctx, key)
    HIT  → return value                         no gas
    MISS →
      depthEstimator == nil
      gctx.WillGet()                             1x ReadCostFlat
      value = parent.Get(nil, key)
        dbadapter.Store.Get(nil, key)             pure storage, no gas
      gctx.DidGet(value)                          ReadCostPerByte
```

Callers without gas (tests):
```
stor.Get(nil, key)               // same path, all gas calls are no-ops
```

### GnoVM integration

The VM keeper passes `gctx` into the Gno transaction store:

```go
func (vm *VMKeeper) newGnoTransactionStore(ctx sdk.Context) gno.TransactionStore {
    base := ctx.Store(vm.baseKey)
    iavl := ctx.Store(vm.iavlKey)
    gctx := ctx.GasContext()
    gasMeter := ctx.GasMeter()
    return vm.gnoStore.BeginTransaction(base, iavl, gctx, gasMeter)
}
```

The Gno `defaultStore` holds the `gctx` and passes it through on every
`baseStore.Get(gctx, key)`, `baseStore.Set(gctx, key, bz)`,
`iavlStore.Get(gctx, key)`, and `iavlStore.Set(gctx, key, bz)` call.
Storage I/O gas is charged by the `cache.Store` layer above iavl/dbadapter
automatically. Note that Gno's `SetObject`/`SetType`/`SetPackageRealm`
call `baseStore.Set(gctx, key, bz)` eagerly during execution — but
`baseStore` is a `cache.Store`, so these writes are buffered in the cache
(not written to the underlying iavl/dbadapter until `Write()` flushes).
Gas is charged at the `cache.Store.Set()` call, when `gctx` is live.
`tree.Size()` on the underlying IAVL tree is unaffected by these buffered
writes — it remains constant throughout the block until `Commit()`.

The GnoVM `defaultStore` continues to charge its own gas via `consumeGas()`,
but the per-operation constants (`GasGetObject`, `GasSetType`, etc.) are
replaced by two universal amino constants defined in `tm2/pkg/amino`:

```go
// tm2/pkg/amino/gas.go
const (
    GasEncodePerByte int64 = 3   // ~2.8 ns/byte (Binary2/genproto2 amino marshal)
    GasDecodePerByte int64 = 3   // ~2.8 ns/byte (same slope assumed for unmarshal)
)
```

These constants come from `gnovm/adr/STORAGE_CHARGING_AMINO_HEURISTIC.png`
which benchmarks amino Binary2 (genproto2) marshal at **2.8 ns/byte + 427 ns
flat**. The per-byte slope is used; the flat component is small relative to
typical serialized sizes. The same slope is used for decode (unmarshal)
pending separate benchmarks. These constants assume a future migration to
amino2 (Binary2); reflect-based amino is ~12.9 ns/byte but will be replaced.

The GnoVM `GasConfig` imports these as defaults:

```go
type GasConfig struct {
    // Storage I/O — charged via GasContext at cache.Store level
    // (not in this config; lives in store.GasConfig)

    // Amino compute — charged at the Gno store level
    GasAminoEncode int64  // per byte, default from amino.GasEncodePerByte
    GasAminoDecode int64  // per byte, default from amino.GasDecodePerByte
}
```

This replaces the current 8 per-operation constants with 2 amino constants.
Every store read charges `GasAminoDecode * len(bz)` on cache miss (where
amino unmarshal actually happens). Every store write charges
`GasAminoEncode * len(bz)` (where amino marshal happens). The cost is the
same regardless of whether the data is an object, type, realm, or
mem-package — amino cost per byte is amino cost per byte.

| Operation | Current | New: amino compute | New: I/O (via GasContext) |
|-----------|---------|-------------------|--------------------------|
| GetObject (16/byte) | mixed | GasAminoDecode/byte | depth\*ReadFlat + ReadPerByte |
| SetObject (16/byte) | mixed | GasAminoEncode/byte | depth\*(ReadFlat+WriteFlat) + WritePerByte |
| GetType (5/byte) | mixed | GasAminoDecode/byte | depth\*ReadFlat + ReadPerByte |
| SetType (52/byte) | mixed | GasAminoEncode/byte | depth\*(ReadFlat+WriteFlat) + WritePerByte |
| GetPackageRealm (524/byte) | mixed | GasAminoDecode/byte | depth\*ReadFlat + ReadPerByte |
| SetPackageRealm (524/byte) | mixed | GasAminoEncode/byte | depth\*(ReadFlat+WriteFlat) + WritePerByte |

For non-IAVL stores (depth=1, no estimator), Set charges
`WriteCostFlat + WritePerByte` and Get charges `ReadCostFlat + ReadPerByte`.

The current per-operation asymmetries (SetType 52x vs GetType 5x,
PackageRealm 524x vs Object 16x) disappear. Those differences were
artifacts of conflating I/O cost with amino cost — the I/O component
varied by access pattern, inflating some constants relative to others.
With I/O factored out, amino encode/decode cost per byte is uniform.

### Cache hit behavior changes

**Before (current):**
```
Gno GetObject cache hit:   GasGetObject * len(bz) charged    (WRONG)
Gno GetObject cache miss:  GasGetObject * len(bz) charged    (same cost!)
```

**After:**
```
Gno GetObject cache hit:   no gas (already deserialized, already in memory)
Gno GetObject cache miss:  amino compute gas + I/O gas (depth-based)
```

The GnoVM store's `consumeGas` calls should be moved to the cache-miss path:

```go
func (ds *defaultStore) GetObject(oid ObjectID) Object {
    if oo, exists := ds.cacheObjects[oid]; exists {
        return oo  // cache hit — no amino, no I/O, no gas
    }
    bz := ds.baseStore.Get(ds.gctx, key)  // I/O gas charged by cache.Store
    amino.MustUnmarshal(bz, &obj)
    ds.consumeGas(ds.gasConfig.GasAminoDecode * int64(len(bz)), "AminoDecodePerByte")
}
```

## Changes by package

### tm2/pkg/amino/gas.go (new)
- Define `GasEncodePerByte = 3` and `GasDecodePerByte = 3`
  (see Calibration section)

### tm2/pkg/store/types/gas.go
- Add `GasContext` struct and methods
- `WillSet`/`WillDelete` return `Gas` (the amount charged)
- `RefundGas` on `GasContext` calls `Meter.RefundGas`
- Add `RefundGas(amount Gas, descriptor string)` to `GasMeter` interface
  (floors at zero). Implementations needed:
  - `basicGasMeter`: decrement `consumed`, floor at zero
  - `infiniteGasMeter`: decrement `consumed`, floor at zero
  - `passthroughGasMeter`: refund on both `Base` and `Head` meters
- Add `DepthEstimator` interface

### tm2/pkg/store/types/store.go
- Add `*GasContext` as first parameter to `Get`/`Set`/`Has`/`Delete` in
  the `Store` interface

### tm2/pkg/store/gas/
- Delete. No longer needed. `ctx.GasContext()` replaces it.

### tm2/pkg/sdk/context.go
- Add `GasContext()` method returning `*store.GasContext`
- Remove `GasStore()` method (or deprecate)

### tm2/pkg/sdk/auth/keeper.go
- Replace `ctx.GasStore(ak.key)` with `ctx.Store(ak.key)` + `ctx.GasContext()`
- Pass `gctx` as first arg to all `stor.Get`/`Set`/`Delete` calls

### tm2/pkg/store/cache/store.go
- Update `Get`/`Set`/`Has`/`Delete` signatures to accept `*GasContext`
- Add `depthEstimator DepthEstimator` field, set via `SetDepthEstimator()`
- On read miss: charge `depth * ReadCostFlat` (or 1x if no estimator) +
  `ReadCostPerByte`, then call `parent.Get(nil, key)` (parent doesn't charge)
- Add `chargedGas map[string]Gas` field for write/delete deduplication
- `Set()`/`Delete()` refund previous charge for same key, then charge
  depth-based + flat + per-byte (last operation wins)
- `clear()` must reset `chargedGas` alongside existing cache state
- `Write()` unchanged — passes `nil` to parent (already paid)
- `CacheWrap()` propagates `depthEstimator` to child cache

### tm2/pkg/store/iavl/store.go
- Update `Get`/`Set`/`Has`/`Delete` signatures to accept `*GasContext`
  (parameter is ignored — iavl.Store is pure storage)
- Implement `DepthEstimator` interface:
  `ExpectedDepth() = bits.Len64(uint64(tree.Size()))` (floor(log2(size)) + 1)
- `CacheWrap()` sets `depthEstimator` on the returned cache.Store

### tm2/pkg/store/dbadapter/store.go
- Update `Get`/`Set` signatures to accept `*GasContext` (parameter is
  ignored — dbadapter.Store is pure storage, gas charged by cache above)

### tm2/pkg/store/prefix/store.go
- Update `Get`/`Set`/`Has`/`Delete`/`Iterator` signatures to accept
  `*GasContext`, pass through to parent
- `CacheWrap()` propagates `depthEstimator` from parent (if parent
  implements `DepthEstimator`, or if prefix.Store carries one)

### tm2/pkg/store/immut/store.go
- Update `Get`/`Has`/`Iterator` signatures to accept `*GasContext`, pass
  through to parent

### tm2/pkg/iavl/ (mutable_tree.go, immutable_tree.go, nodedb.go)
- No changes needed. Gas is estimated and charged at the `cache.Store`
  level via `DepthEstimator`. The tree internals are not modified.

### tm2/pkg/db/ (all backends)
- No changes needed. The DB interface is unchanged.

### tm2/pkg/sdk/bank/keeper.go
- Mechanical: pass `nil` (or `gctx` if bank should charge I/O gas) to
  store calls made via auth keeper

### tm2/pkg/sdk/params/keeper.go
- Mechanical: pass `nil` or `gctx` to store `Get`/`Set` calls

### gno.land/pkg/sdk/vm/keeper.go
- Pass `ctx.GasContext()` into `gnoStore.BeginTransaction()`
- Set `MinDepth = 12` on the GasConfig (from governance params)

### gnovm/pkg/gnolang/store.go
- Accept `*GasContext` in `BeginTransaction()`
- Pass `gctx` through all `baseStore.Get(gctx, key)` calls
- Move `consumeGas` calls to cache-miss paths
- Replace 8 per-operation gas constants with `GasAminoEncode`/`GasAminoDecode`
  (defaults from `tm2/pkg/amino`)

### All other callers
- Mechanical: add `nil` as first argument to every `Get`/`Set`/`Has`/`Delete`
  call. The compiler flags every site.

## Migration

Two commits:

**Commit 1: Signature change (mechanical refactor)**

1. Add `GasContext` type and `DepthEstimator` interface to
   `tm2/pkg/store/types/gas.go`.
2. Add `GasContext()` method to `sdk.Context`.
3. Change the `Store` interface: add `*GasContext` as first parameter to
   `Get`/`Set`/`Has`/`Delete`/`Iterator`/`ReverseIterator`.
4. Update every implementation and call site. All callers pass `nil`.
5. Auth keeper: replace `ctx.GasStore()` with `ctx.Store()`, pass `nil`
   as gctx. **Auth storage gas is temporarily removed** (restored in
   commit 2).
6. Delete `tm2/pkg/store/gas/` package.
7. All tests pass with updated gas-wanted values where needed.

**Commit 2: Wire up gas at the cache.Store boundary (behavior change)**

1. Add `RefundGas` to `GasMeter` interface and all implementations
   (`basicGasMeter`, `infiniteGasMeter`, `passthroughGasMeter`).
2. Implement `DepthEstimator` on `iavl.Store`. Add `depthEstimator`
   field + propagation to `cache.Store`, `prefix.Store`.
3. Add depth-based gas charging and `chargedGas` deduplication to
   `cache.Store.Get`/`Set`/`Delete`.
4. Auth keeper: pass `ctx.GasContext()` instead of `nil`.
   **Auth keeper gas is restored**, now charged at the cache.Store level.
   Cache hits become free.
5. VM keeper: pass `ctx.GasContext()` into `BeginTransaction()`.
   **VM transactions now incur I/O gas.**
6. Move GnoVM `consumeGas` calls to cache-miss paths. Re-label constants
   as amino compute costs.
7. Set constants:
   - `store.GasConfig` (tm2 layer): `ReadCostFlat = 59_000`,
     `WriteCostFlat = 24_000` (see calibration notes below).
   - `amino.GasEncodePerByte = 3`, `amino.GasDecodePerByte = 3`
     (see amino calibration notes below).
8. Update gas-wanted values in integration tests.

## Calibration

### Storage I/O constants

```go
// store.DefaultGasConfig()
ReadCostFlat:      59_000   // ~59µs per random read at 100M keys
WriteCostFlat:     24_000   // ~24µs per write (amortized, batch=1000) at 100M keys
ReadCostPerByte:   17       // ~17 ns/byte (LMDB overflow page reads at 100KB values)
WriteCostPerByte:  14       // ~14 ns/byte (LMDB overflow page writes at 100KB values)
DeleteCost:        59_000   // same as ReadCostFlat (delete requires finding the key)
IterNextCostFlat:  1_000    // ~1µs per iteration step (sequential leaf scan)
MinDepth:          0        // tm2 default: no floor (use actual depth estimate)
```

gno.land overrides `MinDepth`:

```go
// gno.land default (set in vm keeper or params)
MinDepth: 12   // floor for IAVL depth estimate (~4K keys equivalent)
```

`MinDepth` is a **floor** on the depth estimate used for gas calculation.
It prevents gas from being unrealistically cheap for small or empty IAVL
trees, which have real overhead not captured by `log2(size)` alone. With
`MinDepth = 12`, even a 1-key tree pays `12 * ReadCostFlat` per read
rather than `1 * ReadCostFlat`.

`MinDepth` is a **governance parameter** in gno.land — adjustable via
param proposals without a code change. This allows the community to raise
the floor if operational costs increase or lower it if gas is too
expensive for small realms.

The depth used for gas is: `max(ExpectedDepth(), MinDepth)`.

These constants are calibrated for **LMDB** from `gnovm/cmd/benchstore`:

- `ReadCostFlat = 59_000`: random Get latency at 100M keys (~59µs).
  From local NVMe benchmarks.
- `WriteCostFlat = 24_000`: SetOverwrite latency at 100M keys, batch=1000,
  amortized per key (~24µs). From local NVMe benchmarks.
- `ReadCostPerByte = 17`: per-byte I/O cost for reads at large value sizes
  (100KB values, 1M keys, ~95 GB on disk). From the value size sweep
  (`BenchmarkValueSizeGet`).
- `WriteCostPerByte = 14`: per-byte I/O cost for writes at large value
  sizes (100KB values, 1M keys, ~95 GB on disk). From the value size
  sweep (`BenchmarkValueSizeSet`).

- `DeleteCost = 59_000`: estimated as equal to `ReadCostFlat` — a delete
  must find the key (same cost as a read). TODO: validate with dedicated
  delete benchmarks in `gnovm/cmd/benchstore`.
- `IterNextCostFlat = 1_000`: estimated at ~1µs per step — LMDB leaf
  pages are linked, so sequential iteration is much cheaper than random
  reads (~59µs). TODO: validate with dedicated iterator benchmarks.

**Caveats:**
- The flat costs (59K/24K) are from **local NVMe** benchmarks. The per-byte
  costs (17/14) are from **networked SSD** (Xeon 8358) benchmarks and
  should be re-validated on local NVMe using the value size sweep in
  `gnovm/cmd/benchstore`. Local NVMe has faster sequential I/O, so the
  per-byte costs may be lower.
- The current default database backend is **PebbleDB**, not LMDB. A future
  migration to LMDB is planned. These constants are calibrated for LMDB
  and may overcharge or undercharge on PebbleDB. The constants should be
  re-validated after the LMDB migration.

### Amino compute constants

```go
// amino.GasEncodePerByte / amino.GasDecodePerByte
GasEncodePerByte = 3   // ~2.8 ns/byte
GasDecodePerByte = 3   // ~2.8 ns/byte (same slope assumed)
```

From `gnovm/adr/STORAGE_CHARGING_AMINO_HEURISTIC.png`: amino Binary2
(genproto2) marshal benchmarks show **2.8 ns/byte + 427 ns flat**. The
per-byte slope is used; the flat component is small relative to typical
serialized sizes. The same slope is assumed for unmarshal pending separate
benchmarks. These constants assume amino2 (Binary2); the current
reflect-based amino is ~12.9 ns/byte but will be replaced.

## Observability bonus

The `GasContext` pattern makes it trivial to count actual DB hits per
transaction without changing gas logic:

```go
type GasContext struct {
    Meter      GasMeter
    Config     GasConfig
    ReadCount  int64  // optional: count DB reads
    WriteCount int64  // optional: count DB writes
    ReadBytes  int64  // optional: total bytes read from DB
}
```

This replaces the benchops instrumentation for storage operations with
production-available metrics.

## What this does NOT change

- **CPU/opcode gas**: VM instruction gas is unchanged.
- **Allocation/GC gas**: unchanged.
- **Storage deposits**: unchanged (charged at the realm layer, orthogonal).
- **Block gas limit**: unchanged.
- **Gas meter / passthrough meter**: unchanged, except `RefundGas` added
  to the `GasMeter` interface for write deduplication.
- **IAVL tree internals**: unchanged (no gctx threading through tree/nodeDB).
- **dbm.DB interface**: unchanged.

## PkgID Flag Nibble

The first nibble (4 bits) of `PkgID.Hashlet` is reserved for package
classification flags, set during `PkgIDFromPkgPath()`:

- bit 0 (0x80): `IsStdlib` — standard library package
- bit 1 (0x40): `IsImmutable` — stdlib or `/p/` package
- bit 2 (0x20): `IsInternal` — internal package path
- bit 3 (0x10): reserved (always 0)

The remaining 156 bits are the truncated SHA-256 hash. Methods:
`PkgID.IsStdlibPkg()`, `PkgID.IsImmutablePkg()`, `PkgID.IsInternalPkg()`.
O(1) bitwise checks, no map lookups or path string parsing.

These flags enable efficient runtime decisions about package mutability
and stdlib caching without needing reverse mappings from hash to path.

## Immutable Package Guard

When a realm references a stdlib or `/p/` package object (e.g.,
`ref = strconv.Itoa`), the realm finalization previously mutated the
object's `RefCount`, `IsEscaped`, and `ModTime`, then re-persisted it
to the store. This was unnecessary — immutable package objects are never
deleted, so refcount tracking serves no purpose from other realms.

In `DidUpdate`, `incRefCreatedDescendants`, and `decRefDeletedDescendants`,
all mutations are skipped when:

```go
co.GetObjectID().PkgID.IsImmutablePkg() && co.GetObjectID().PkgID != rlm.ID
```

The `PkgID != rlm.ID` check allows mutations during the package's own
initialization (via throwaway realm where `rlm.ID == PkgID`). For objects
with zero ObjectID (not yet real), `IsImmutablePkg()` returns false, so
the guard doesn't apply.

This eliminates spurious writes to stdlib/`p`-package objects during
realm transactions.

## Stdlib Byte Cache

Stdlib object reads previously charged 59K+ gas per cache miss at the
tx-scoped cache layer, even though stdlib objects are immutable
infrastructure loaded at every node start. Profiling showed the
RegisterUser transaction spent 408M gas (80%) on 252 stdlib object reads.

### Design

`defaultStore.stdlibKeyBytes map[string][]byte` caches raw amino bytes
for all stdlib objects. In `loadObjectSafe`, stdlib objects are read from
the byte cache instead of `baseStore.Get(gctx)`, skipping I/O gas:

```go
var hashbz []byte
if oid.PkgID.IsStdlibPkg() {
    hashbz = ds.stdlibKeyBytes[key]  // from byte cache — no I/O gas
}
if hashbz == nil {
    hashbz = ds.baseStore.Get(ds.gctx, []byte(key))  // charges I/O gas
}
```

Each transaction amino-unmarshals its own copy from cached bytes — no
shared mutable objects, no aliasing risk. If a VM bug mutates a stdlib
object, the corruption is contained to one transaction. Amino decode gas
(3/byte) is still charged — it's real CPU work.

### Population

`PopulateStdlibCache(paths)` iterates the `baseStore` with per-package
prefix iterators on `"oid:<pkgid_hex>:"` for each stdlib path. Called:

- From `VMKeeper.Initialize()` on restart (after `PreprocessAllFilesAndSaveBlockNodes`)
- From `InitChainerConfig.loadStdlibs()` at genesis (after `CommitGnoTransactionStore`)

The map is shared via `BeginTransaction` (shared reference). Stdlib
objects are immutable after genesis, so bytes never change.

### Safety

The byte cache approach was chosen over an object cache because:
- Each tx gets its own deserialized copy (no shared mutable state)
- If a VM bug mutates a stdlib object, it doesn't persist beyond the tx
- No aliasing between per-tx `cacheObjects` and shared cache

## Implementation Commits

1. ADR — this document
2. `*GasContext` on Store interface — ~490 call sites updated
3. cache.Store gas charging — DepthEstimator, chargedGas deduplication
4. GnoVM store threaded — `ds.gctx` through all baseStore/iavlStore calls
5. Amino constants — 8 per-operation → 2 universal encode/decode
6. gas.Store deleted — `GasStore()` removed from sdk.Context
7. Calibrated constants — LMDB benchmarks applied to DefaultGasConfig
8. MinDepth governance — `p:min_depth` param, default 12 for gno.land
9. PkgID flag nibble — IsStdlib/IsImmutable/IsInternal bits
10. Immutable package guard — skip refcount mutations in DidUpdate
11. Stdlib byte cache — gas-free stdlib object reads
12. Debug cleanup — all instrumentation removed
13. ADR update — this section
