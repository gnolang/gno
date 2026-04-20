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
func (gctx *GasContext) WillIterator()                  // ReadCostFlat (one Get-equivalent tree walk)
func (gctx *GasContext) WillIterNext(value []byte)      // IterNextCostFlat + ReadCostPerByte * len(value)
```

All methods are nil-safe: if `gctx == nil`, they are no-ops (returning 0
for methods that return `Gas`).

### DepthEstimator

Stores with depth-dependent I/O cost expose three estimates, one per
op shape. Values are fixed-point ×100 so the gas formula can work in
integer arithmetic while still representing fractional depths.

```go
type DepthEstimator interface {
    ExpectedGetReadDepth100() int64 // reads on Get: full leaf+value fetch
    ExpectedSetReadDepth100() int64 // reads on Set: leaf lookup, no value
    ExpectedWriteDepth100() int64   // writes on Set/Delete: COW path
}
```

`iavl.Store` implements all three. The three depths differ in how
much of the tree walks cost is attributable to the operation — a
Get pays the full read descent plus value-page fetch, a Set's
read-half touches only interior nodes, and the write-half depends
on the COW/path-copy amortization. See `tm2/pkg/store/iavl/store.go`
for the current formulas; they are calibrated for IAVL B+32 at
100M keys.

`dbadapter.Store` does not implement `DepthEstimator` (no depth,
flat I/O cost). `cache.Store` propagates its parent's estimator
through nested `CacheWrap()` calls so the right depths apply even
when multiple cache layers are stacked.

`tree.Size()` is consensus state — it does not change during block
execution (cache layers buffer all writes above the IAVL tree until
`Commit()`). All transactions in a block see the same depth estimate.

Governance (`vm.Params`) overlays two controls on each of the three
shapes:

- `Min{Get,Set,Write}ReadDepth100` — a floor below which the tree
  estimate is ignored. Default 0 in tm2 (use raw estimate);
  gno.land's `DefaultParams` sets 300/200/440 — calibrated for B+32
  at 100M items with 10K cache and batched 1000 mutations.
- `Fixed{Get,Set,Write}ReadDepth100` — an override that replaces
  the tree estimate entirely. Default 0 in tm2 ("no override"); 
  gno.land defaults to the same values as the Min floors so an
  empty-tree genesis starts at sane depths.

`Validate()` rejects negative values and caps each depth at
`10_000` (= 100 tree levels, well beyond any plausible B+tree /
IAVL height) to prevent governance-set absurd values from tripping
`overflow.Mulp` in `cache.Store`'s charge calc.

| Tree size | log2(size) | min_*_read_depth_100 (default) | Effective depth |
|---|---|---|---|
| 1K       | 10 | 300 (Get) / 200 (Set) / 440 (Write) | 3.0 / 2.0 / 4.4 |
| 100K     | 17 | 300 / 200 / 440 | 3.0 / 2.0 / 4.4 |
| 10M      | 24 | raw estimate prevails above the floor |
| 100M     | 27 | raw estimate prevails above the floor |

The three-depth split was introduced by commit `359308d6f`
("feat(store): split DepthEstimator into three depths (read/write)").
Earlier design iterations used a single `ExpectedDepth()` / `MinDepth`
field; those predate the split.

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

**Gas is charged at the `cache.Store` layer only.** Parents
(`iavl.Store`, `dbadapter.Store`) return un-wrapped iterators and rely
on the cache wrap above them to charge — wrapping at multiple layers
would double-count. `prefix.Store` and `immut.Store` delegate, so they
transparently receive the gas-wrapped iterator when their parent is a
`cache.Store`.

**Cost model:**
- **Seek** (iterator creation): `ReadCostFlat`. A seek tree-walks from
  the root to the first leaf — equivalent work to a flat `Get`.
  Charged **unconditionally**, even if the range turns out empty —
  the DB still did the walk.
- **Step** (each `Next()` that lands on a valid item):
  `IterNextCostFlat + ReadCostPerByte × len(value)`. The per-byte
  component uses descriptor `GasValuePerByteDesc` so traces
  distinguish iterator-returned bytes from Get-returned bytes.

The step charge fires eagerly on advance (matching the physical cost
— the backend has already fetched the page), not lazily on `Value()`.

**Reachability.** Realms cannot reach store iteration directly. Live
call sites:
1. `vm/qpaths` → `FindPathsByPrefix` (cache-wrapped iavl). Consensus
   observable only via the query meter, now bounded below.
2. `auth.IterateAccounts` via `PrefixIterator`. Threaded through, but
   today all production query contexts carry `NewInfiniteGasMeter()`
   — the charge fires with no enforcement until a future caller
   passes a bounded meter.
3. `iavl.Store` ABCI subspace query at `iavl/store.go:330` — passes
   `nil` gctx on a bare (non-cache) iavl parent, gas-free.
4. Node-startup / test helpers (`CopyFromCachedStore`,
   `populateStdlibCache`) pass `nil` gctx deliberately; comments at
   the call sites document the intent.

**Query gas metering.** All five VM query handlers — `QueryPaths`,
`QueryFuncs`, `QueryFile`, `QueryDoc`, `QueryStorage` — now install
`store.NewGasMeter(maxGasQuery)` before building the throwaway tx
store. Previously only `queryEvalInternal` did this; the others ran
against an infinite meter, making iteration gas unenforceable on
those handlers.

**`PrefixIterator` / `ReversePrefixIterator`** in
`tm2/pkg/store/types/utils.go` now take `*GasContext` as first
argument so caller-threaded gas propagates through a cache-wrapped
parent. `DiffStores` (zero callers) and the `First`/`Last` helpers in
`firstlast.go` (zero callers) are not touched — follow-up cleanup.

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

**Iterator (all stores):**
```
caller: stor.Iterator(gctx, start, end)
  cache.Store.iterator(gctx, start, end)
    parent := cache.parent.Iterator(nil, start, end)     parent gets nil — not wrapped
    cache := newMemIterator(...)                          dirty-cache overlay
    merged := newCacheMergeIterator(parent, cache)        merge parent + dirty
    return newGasIterator(gctx, merged)                   single gas wrap
      WillIterator()                                      ReadCostFlat (seek)
      if merged.Valid(): WillIterNext(merged.Value())     1st step (if any)
caller.Next():
  merged.Next()
  if merged.Valid(): WillIterNext(merged.Value())         IterNextFlat + perByte(value)
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
  reads (~59µs). See the _Iterator constants_ subsection below for the
  benchmark command to re-validate at 100M-key scale.

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

### Iterator constants

```go
IterNextCostFlat = 1_000   // per-step sequential leaf scan
// WillIterator uses ReadCostFlat (= 59_000) — seek does a Get-equivalent tree walk.
```

`IterNextCostFlat` is a **governance parameter** (`p:iter_next_cost_flat`);
the tm2 default above is applied via `vm.DefaultParams()` at genesis.
`Validate()` rejects zero or negative values. `ReadCostFlat` (and the
other flat/per-byte constants) remain compile-time constants today.

**Upgrade note.** A chain whose Params were serialized before this field
existed will amino-decode `IterNextCostFlat == 0`. `GetParams` does
not re-validate on read, so the chain starts, but:

- `ApplyToGasConfig` will write zero into the live `GasConfig`,
  effectively disabling iterator-step gas until Params is re-set.
- `WillSetParam` (invoked by governance proposals) re-validates the
  full Params struct, so *any* proposal on unrelated fields will panic
  with `IterNextCostFlat must be positive`.

Operators upgrading from a pre-field snapshot must either (a) reset
genesis via `DefaultParams()`, or (b) submit `p:iter_next_cost_flat =
1000` as the *first* governance proposal after upgrade — no other
proposal will land until that value is set.

**Calibration target: 100M keys on the reference hardware.** The other
storage constants (`ReadCostFlat`, `ReadCostPerByte`, etc.) are
calibrated at that scale in the Storage I/O section above; the
iterator constants should be re-validated there too.

Run:
```
go test ./gnovm/cmd/benchstore/ -bench='IterNext|IterSeek' \
    -timeout=2h -db=lmdb
```

`BenchmarkIterNext` reports `ns/op` for `Next()+Key()+Value()`.
`BenchmarkIterSeek` reports `ns/op` for opening a random-start
iterator and positioning to the first key. Both should be run at the
full `keySizes` sweep up to 100M keys; only the 100M row drives the
calibrated value, the smaller rows are sanity-check points showing
step cost stays roughly flat while seek grows with tree depth.

Until the 100M-key numbers are in, the defaults above are chosen
conservatively: step at 1_000 gas (~1 µs) and seek at 59_000 gas
(one flat Get). Both may over-charge warm-cache hits but should not
under-charge cold-cache reads at the target scale.

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
8. Three-depth governance — six `p:{min,fixed}_{get_read,set_read,write}_depth_100`
   params (split out by commit `359308d6f` from the earlier single-`MinDepth`
   design)
9. PkgID flag nibble — IsStdlib/IsImmutable/IsInternal bits
10. Immutable package guard — skip refcount mutations in DidUpdate
11. Stdlib byte cache — gas-free stdlib object reads
12. Debug cleanup — all instrumentation removed
13. Iterator gas at cache layer — `gasIterator`, `WillIterator` /
    `WillIterNext(value)`, 5 VM query handlers bounded by `maxGasQuery`
14. `IterNextCostFlat` governance parameter (`p:iter_next_cost_flat`)
15. ADR update — this section

## Follow-ups

Items identified during implementation that are intentionally deferred
out of this PR. Capture as issues when merging:

1. **100M-key iterator benchmark calibration.** Run
   `BenchmarkIterNext` / `BenchmarkIterSeek` at the 100M reference scale
   on LMDB and adjust `IterNextCostFlat` if needed. Defaults in this PR
   are conservative placeholders chosen to over-charge warm-cache steps
   rather than under-charge cold-cache.

2. **Dedicated `DeleteCost` benchmark.** Currently `DeleteCost` is set
   equal to `ReadCostFlat` on the premise "delete must find the key"
   — validate with a direct benchmark once LMDB delete timings are in
   hand.

3. **`ReadCostPerByte` / `WriteCostPerByte` re-validation on local
   NVMe.** The current slopes come from networked SSD (Xeon 8358); local
   NVMe has faster sequential I/O and the per-byte costs may be lower.
   Re-run the value-size sweep and re-calibrate.

4. **`GasDecodePerByte` dedicated benchmark.** Currently assumed equal
   to `GasEncodePerByte` (~2.8 ns/byte); amino unmarshal may have a
   different slope.

5. **Remove zero-caller helpers.** `tm2/pkg/store/types/utils.go`
   `DiffStores` and `tm2/pkg/store/firstlast.go`'s `First`/`Last` have
   no production callers today. Both currently pass `nil` gctx, so
   any future caller that wired a bounded meter to them would iterate
   gas-free. Delete or adopt.

6. **gnoweb OOG handling.** `vm/qpaths` now returns OOG on large
   prefixes under `maxGasQuery`. Long-lived gnoweb clients
   (`gno.land/pkg/gnoweb/client.go`) should handle the error
   gracefully — today they may surface a bare internal error.

7. **`qpaths_oog.txtar` DoS-validation fixture.** Deferred because a
   txtar exercising the bounded meter needs ~1.6M paths
   (`maxGasQuery / per-step gas`). Either make `maxGasQuery`
   test-overridable or accept the gap.

8. **Re-calibrate constants after any backend migration.** Current
   defaults are calibrated for LMDB; if gno.land runs on PebbleDB or
   MDBX in production the values may over- or under-charge.

9. **`Params.Validate` test coverage.** `params_test.go` should
   include failing cases for negative depth values and zero / huge
   `IterNextCostFlat` to lock in the current rejection semantics.

10. **`/subspace` ABCI query bounding.** `iavl.Store.Query` at the
    `/subspace` branch (`iavl/store.go:329-341`) materializes the full
    prefix into memory before amino-marshalling. Pre-existing, but
    worth a simple hard-cap on result count or total byte size now
    that the rest of the store path is metered.

11. **`handleQueryCustom` meter bounding.** Custom-querier paths
    (auth/bank/params) run under `InfiniteGasMeter`. Single-Get
    queries are small; `auth.IterateAccounts` through a custom route
    is the exposure. Either bound or document the decision.

12. **`immut.Store` DepthEstimator forwarding.** Historical-query
    `MultiImmutableCacheWrapWithVersion` currently charges flat
    `ReadCostFlat` instead of the depth-based cost because
    `immut.Store` doesn't forward `DepthEstimator` to the cache wrap
    above it. Either forward it or document historical queries as
    intentionally flat.
