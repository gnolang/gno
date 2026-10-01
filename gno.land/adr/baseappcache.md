# ADR: Single Cache Wrap with Checkpoint for BaseApp Transaction Execution

## Problem

In `BaseApp.runTx()`, two separate cache wraps are created around the block-level
multistore:

1. **Ante cache** — wraps block store, used by the ante handler (sig verify, fee deduction)
2. **Msg cache** — wraps block store, used by message execution (VM, realm finalization)

After the ante handler succeeds, its cache is flushed to the block store and discarded.
A fresh empty cache is created for message execution. This means every key read by the
ante handler (sender account, fee collector, params) must be re-read by the message
handler, incurring full IAVL depth-based gas charges even though the data is sitting
one layer up in the block-level cache.

For a simple `SaveStructToPublicRealm` call (3.4M gas total), the redundant account
re-reads alone cost ~354,000 gas (two DepthReadFlat charges at 177,000 each).

### Why Not Just Check Parent Cache?

The tx-level cache miss gas should not depend on the block-level cache state. Whether
a previous tx in the same block loaded a key should not affect this tx's gas cost.
The fix must be scoped to the single transaction's own cache layers.

## Solution

Use a **single cache wrap** for the entire transaction (ante + msgs), with a
lightweight checkpoint mechanism so ante writes survive msg failure.

### Checkpoint Mechanism on cacheStore

Add two new fields and two methods to `cacheStore`:

```go
type cacheStore struct {
    // ... existing fields ...
    checkpointCache      map[string]*cValue   // nil when no checkpoint active
    checkpointChargedGas map[string]types.Gas  // nil when no checkpoint active
}

// Checkpoint saves a shallow clone of the cache and chargedGas maps.
func (store *cacheStore) Checkpoint() {
    store.checkpointCache = maps.Clone(store.cache)
    store.checkpointChargedGas = maps.Clone(store.chargedGas)
}

// HasCheckpoint returns true if a checkpoint is active.
func (store *cacheStore) HasCheckpoint() bool {
    return store.checkpointCache != nil
}

// WriteCheckpoint restores the checkpoint snapshot, then calls Write()
// to flush only the checkpointed (ante) entries to the parent.
func (store *cacheStore) WriteCheckpoint() {
    if store.checkpointCache == nil {
        panic("WriteCheckpoint called without Checkpoint")
    }
    store.cache = store.checkpointCache
    store.chargedGas = store.checkpointChargedGas
    store.checkpointCache = nil
    store.checkpointChargedGas = nil
    store.Write()
}
```

The existing `clear()` method (called by `Write()`) must also nil out
the checkpoint fields to prevent the panic-recovery defer from
double-flushing on the success path:

```go
func (store *cacheStore) clear() {
    store.cache = make(map[string]*cValue)
    store.unsortedCache = make(map[string]struct{})
    store.sortedCache = list.New()
    store.chargedGas = make(map[string]types.Gas)
    store.checkpointCache = nil
    store.checkpointChargedGas = nil
}
```

**Why `cache` clone is safe:** `setCacheValue` always creates a new `*cValue`
(never mutates in place), so the cloned map's pointers remain valid after
subsequent Set/Delete calls by the msg handler.

**Why `chargedGas` is checkpointed:** Not strictly needed today — nothing
reads `chargedGas` after `WriteCheckpoint()` calls `Write()` then `clear()`.
But checkpointing it is cheap (`maps.Clone` of an int map) and defensive
against future changes (e.g., gas rebates on rollback).

**Fields that do NOT need checkpointing:**
- `unsortedCache` / `sortedCache` — derived from `cache`; `Write()` rebuilds
  sort order from `cache` directly, then `clear()` resets them.

### Checkpoint on cachemulti.Store

Add corresponding methods that delegate to each sub-store:

```go
func (cms Store) Checkpoint() {
    for _, store := range cms.stores {
        store.(interface{ Checkpoint() }).Checkpoint()
    }
}

func (cms Store) HasCheckpoint() bool {
    for _, store := range cms.stores {
        if store.(interface{ HasCheckpoint() bool }).HasCheckpoint() {
            return true
        }
    }
    return false
}

func (cms Store) WriteCheckpoint() {
    for _, store := range cms.stores {
        store.(interface{ WriteCheckpoint() }).WriteCheckpoint()
    }
}
```

All sub-stores in cachemulti are `*cacheStore` (from `CacheWrap()` calls on
iavl.Store, dbadapter.Store, or cacheStore), so the type assertion succeeds.

### Changes to BaseApp.runTx()

Before (simplified):

```go
// Ante: cache wrap #1
anteCtx, msCache := app.cacheTxContext(ctx)
newCtx, result, abort := app.anteHandler(anteCtx, tx, simulate)
if abort { return result }
ctx = newCtx.WithMultiStore(ms)
msCache.MultiWrite()  // flush ante, discard cache

// Msgs: cache wrap #2 (empty, cold)
runMsgCtx, msCache := app.cacheTxContext(ctx)
result = app.runMsgs(runMsgCtx, msgs, mode)
if result.IsOK() { msCache.MultiWrite() }
```

After (simplified — existing defer blocks, safety checks, gasWanted
assignment, nil anteHandler guard, and validateBasicTxMsgs are retained
but omitted here for clarity):

```go
// Single cache wrap for entire tx
txCtx, msCache := app.cacheTxContext(ctx)
newCtx, result, abort := app.anteHandler(txCtx, tx, simulate)
if abort { return result }  // discard everything

// CheckTx: flush ante writes (sequence, fees) and return.
// No msg execution happens (handler.Process is skipped for CheckTx).
if mode == RunTxModeCheck {
    msCache.MultiWrite()
    return result
}

// DeliverTx and Simulate: checkpoint ante state, then execute msgs.
// Simulate executes msgs for gas estimation (using a real gas meter).
// Its writes flow into a throwaway CacheContext from getContextForTx.
msCache.Checkpoint()

// The defer block must flush ante writes if msg execution panics
// (e.g., OutOfGasError) during DeliverTx. Without this, ante writes
// (sequence, fees) would be silently dropped.
// Only for DeliverTx — simulate panics should not flush anything.
defer func() {
    if mode == RunTxModeDeliver && msCache.HasCheckpoint() {
        msCache.WriteCheckpoint()
    }
}()

runMsgCtx := newCtx
if app.beginTxHook != nil {
    runMsgCtx = app.beginTxHook(runMsgCtx)
}

result = app.runMsgs(runMsgCtx, msgs, mode)
result.GasWanted = gasWanted

// Simulate: return after msg execution without flushing.
// The outer CacheContext (from getContextForTx) discards everything.
if mode != RunTxModeDeliver {
    return result
}

if app.endTxHook != nil {
    app.endTxHook(runMsgCtx, result)
}

if result.IsOK() {
    msCache.MultiWrite()          // flush ante + msg writes
} else {
    msCache.WriteCheckpoint()     // revert to ante-only, flush those
}
```

## Why This Works

### Gas correctness
The msg handler inherits the ante handler's read cache. Keys already loaded
(sender account, fee collector, params) are cache hits — zero gas, correctly
reflecting that no IAVL traversal occurs. Keys not yet loaded still charge
full depth-based gas on first access.

### Determinism
Both phases use the same cache within the same tx. Gas does not depend on
block-level cache state or other transactions.

### Rollback safety
- **Ante fails**: entire cache is discarded (unchanged from today).
- **Ante succeeds, msgs fail**: `WriteCheckpoint()` restores the cache to
  the ante-only state and flushes just those writes. Msg writes vanish.
- **Ante succeeds, msgs succeed**: `MultiWrite()` flushes everything.

### CheckTx correctness
CheckTx must flush ante writes (sequence increment, fee deduction) to
`checkState` for mempool replay protection. The non-Deliver path calls
`MultiWrite()` directly — no checkpoint is needed because CheckTx
skips msg execution (`handler.Process()` is not called), so there are
no msg writes to roll back.

### Simulate correctness
Simulate executes msgs (unlike CheckTx) for gas estimation. It creates
a checkpoint like DeliverTx, but returns early after `runMsgs` without
calling `MultiWrite` or `WriteCheckpoint` — the cache is simply
discarded. The panic-recovery defer also skips `WriteCheckpoint` for
simulate (`mode == RunTxModeDeliver` guard). All writes flow into a
throwaway `CacheContext()` wrapper (from `getContextForTx`), so nothing
persists. Simulate uses a real gas meter (not infinite), so the cache
warming savings directly improve gas estimation accuracy for users.

### Panic safety
If msg execution panics (OutOfGasError, VM panic), the existing `defer`
block in `runTx` catches it and returns an error result. A new `defer`
calls `WriteCheckpoint()` if a checkpoint is active, ensuring ante writes
(sequence increment, fee deduction) are flushed even on panic. Without
this, ante writes would be silently dropped — breaking sequence tracking.

### Query path
ABCI queries (`qeval`, `qrender`, etc.) never enter `runTx`. They use
`handleQueryCustom` which creates an immutable cache wrap independently.
Only `/.app/simulate` goes through `runTx` (covered above). Queries are
completely unaffected.

### Simplicity
- Two new fields on `cacheStore` (`checkpointCache`, `checkpointChargedGas`)
- Two new methods (`Checkpoint`, `WriteCheckpoint`)
- ~15 lines changed in `baseapp.go`
- No new interfaces, no new store types, no gas formula changes

## Gas Impact

For each key read by both the ante handler and message handler, savings are:
- `DepthReadFlat` (depth100 * ReadCostFlat / 100)
- `ReadPerByte` (17 * value_length)
- `AminoDecodePerByte` (3 * value_length)

Typical savings per tx: two account re-reads (sender + fee collector) plus
VM params. Exact amount depends on IAVL tree depth.

## Trade-offs

### Failing transactions become cheaper
In the current design, a failing msg pays full read gas for every key (cold
cache). With the single-cache design, keys loaded by the ante handler are
free reads for the msg. This means intentionally-failing transactions cost
less gas, which slightly changes spam economics. This is an acceptable
trade-off since the gas savings reflect genuine computational savings.

## Implementation Notes

### Interface access
`cacheTxContext` returns `store.MultiStore`, which does not include
`Checkpoint`/`HasCheckpoint`/`WriteCheckpoint`. Define a small
`Checkpointable` interface in `store/types` and type-assert in baseapp:

```go
type Checkpointable interface {
    Checkpoint()
    HasCheckpoint() bool
    WriteCheckpoint()
}
```

### Existing runTx infrastructure to preserve
The pseudocode above is simplified. The implementation must retain:
- The nil `anteHandler` guard (`if app.anteHandler != nil`)
- Safety checks (`newCtx.IsZero()`, `abort && result.Error == nil`)
- `gasWanted = result.GasWanted` assignment from ante result
- `validateBasicTxMsgs` before cache wrapping
- PassthroughGasMeter wrapping for DeliverTx
- Block gas remaining check
- Both existing defer blocks (panic recovery + block gas meter)

### Defer ordering
The `WriteCheckpoint` defer must be registered AFTER the existing defers
(lines 761-806 in current code) so it runs FIRST (LIFO). This ensures
ante writes are flushed before the block gas meter defer runs.

### Mutex locking
Extract a `writeLocked()` helper from `Write()`. Both `Write()` and
`WriteCheckpoint()` acquire `store.mtx` then call `writeLocked()`.
All new methods (`Checkpoint`, `HasCheckpoint`, `WriteCheckpoint`)
follow the existing pattern of locking `mtx` at entry.

`clear()` nils out checkpoint fields, so on the success path
(`MultiWrite()` → `Write()` → `clear()`), the defer's
`HasCheckpoint()` returns false and `WriteCheckpoint` is skipped.
On the failure path, explicit `WriteCheckpoint()` nils the fields
before the defer runs, also skipping it. The defer only fires on
the panic path, which is its intended purpose.

### Removal of WithMultiStore(ms) revert
The current code at line 848 does `ctx = newCtx.WithMultiStore(ms)` to
revert the context's multistore to the block-level store before creating
a second cache wrap. This line is deliberately removed because the single
cache wrap is reused for msg execution — no revert is needed.

## Files to Change

1. `tm2/pkg/store/cache/store.go` — add checkpoint fields, `Checkpoint()`, `HasCheckpoint()`, `WriteCheckpoint()`; update `clear()` to nil checkpoint fields
2. `tm2/pkg/store/cachemulti/store.go` — add `Checkpoint()`, `HasCheckpoint()`, `WriteCheckpoint()` delegation
3. `tm2/pkg/sdk/baseapp.go` — restructure `runTx()` to use single cache wrap with checkpoint defer
4. `tm2/pkg/store/types/store.go` — optionally define `Checkpointable` interface

## Consensus

This changes gas accounting and must be activated at a specific block height.
