# Benchops & Gas Metering: Issues and Concerns

## Context

The GnoVM gas metering system has four layers: CPU opcodes, memory allocation, GnoVM store, and KV store. The benchops system measures wall-clock time per operation to calibrate gas constants. After refactoring benchops to a gap-free SwitchOpCode model, a full audit reveals issues ranging from critical security vulnerabilities to calibration concerns.

---

## CRITICAL: Security Vulnerabilities

### C1. Peak-bytes allocation bypass via GC cycling
**Files:** `gnovm/pkg/gnolang/alloc.go:133-163`, `garbage_collector.go:56`

Gas is only charged when `alloc.bytes > alloc.peakBytes`. GC resets `bytes` to 0 but NOT `peakBytes`. Attack:
1. Allocate X bytes → pay X gas
2. Trigger GC → `bytes=0`, `peakBytes=X`
3. Allocate X again → `bytes=X`, but `X <= peakBytes` → **0 gas**
4. Repeat forever

User-callable `runtime.GC()` (`gnovm/stdlibs/runtime/runtime.go`) makes this trivially exploitable.

### C2. Fork() drops gasMeter and peakBytes
**File:** `gnovm/pkg/gnolang/alloc.go:123-131`

```go
func (alloc *Allocator) Fork() *Allocator {
    return &Allocator{
        maxBytes: alloc.maxBytes,
        bytes:    alloc.bytes,
        // gasMeter: NOT COPIED → nil
        // peakBytes: NOT COPIED → 0
    }
}
```
Forked allocators (used in transaction stores via `store.go:227`) have nil gasMeter → all memory allocations are gas-free.

### C3. No gas metering during preprocessing/parsing
**File:** `gnovm/pkg/gnolang/preprocess.go`

`Preprocess()` does type-checking, import resolution, and declaration processing with NO gas charging. An attacker can submit code with deeply nested types or long import chains to consume unbounded CPU at zero gas cost.

---

## HIGH: Gas Model Flaws

### H1. Suspicious/unvalidated allocation size constants
**File:** `gnovm/pkg/gnolang/alloc.go:29-82`

Multiple constants marked `XXX` — guessed, not measured:
- `_allocBase = 24` — "defensive... XXX"
- `_allocBigint = 200` — XXX
- `_allocBigdec = 200` — XXX
- `_allocType = 200` — XXX
- `_allocAny = 200` — XXX
- `allocMapItem = _allocTypedValue * 3` — XXX (why 3?)

If these are wrong, memory gas is systematically over/under-charged.

### H2. CPU opcode constants are unvalidated
**File:** `gnovm/pkg/gnolang/machine.go:1110-1230`

Many constants marked XXX or with "Todo benchmark this properly":
- `OpCPUEnterCrossing = 100 // XXX`
- `OpCPUCallNativeBody = 424 // Todo benchmark this properly`
- `OpCPUReturnAfterCopy = 38 // XXX`
- `OpCPUStaticTypeOf = 100` — comment says "arbitrary number"

### H3. Store gas config: read == write for objects
**File:** `gnovm/pkg/gnolang/store.go:126-137`

```go
GasGetObject: 16,  // per byte
GasSetObject: 16,  // per byte — SAME as read!
```
Writes are more expensive than reads (disk sync, Merkle proof updates). At the KV layer, writes ARE more expensive (WriteCostFlat=2000 vs ReadCostFlat=1000), but the GnoVM layer doesn't reflect this.

### H4. DeleteObject uses flat cost, ignores object size
**File:** `gnovm/pkg/gnolang/store.go:136, 703`

`GasDeleteObject: 3715` — flat cost regardless of object size. Deleting a 1MB object costs the same gas as deleting a 1-byte object.

### H5. Native function gas gaps
**File:** `gnovm/pkg/gnolang/uverse.go`

Only `print`/`println` have explicit gas metering. Other native functions (`append`, `copy`, `make`, `delete`) rely only on the outer OpCPU charge, not on the actual work done (e.g., copying N elements).

---

## MEDIUM: Benchops Measurement Issues

### M1. ~~runtime.GC() in StartNative distorts timing~~ DONE
**File:** `gnovm/pkg/benchops/bench.go:124`

Fixed: `finalizeCurrent()` now runs before `runtime.GC()` so GC time is not attributed to any opcode.

### M2. Export format: 4-byte uint32 duration cap (4.29s)
**File:** `gnovm/pkg/benchops/exporter.go:39`

Any operation exceeding ~4.29 seconds causes `log.Fatalf`. Real-world store operations with disk I/O could hit this.

### M3. ~~FinishRun exports per-run averages, losing intra-run variance~~ DONE
**File:** `gnovm/pkg/benchops/exporter.go:76-91`

Fixed: export format expanded to 14 bytes with (totalDuration, totalSize, count). Stats computes proper weighted averages and per-run stddev.

### M4. ~~Package loading contaminates store measurements~~ DONE
**File:** `gnovm/cmd/benchops/run.go`

Fixed: all packages loaded before `bm.Init()` creates the exporter. Export functions guard against nil fileWriter so package-loading measurements are silently discarded.

### M5. Benchmark contracts have limited coverage
- Opcodes: doesn't cover all 256 possible opcodes, misses edge cases
- Storage: fixed 1KB writes only, no variable sizes
- Native: only tests print with 3 sizes

---

## LOW: Minor Issues

### L1. GC visitor `VisitCpuFactor = 8` is approximate
**File:** `garbage_collector.go:14` — "TODO: more accurate benchmark"

### L2. GC silent failure returns `(-1, false)` with no diagnostics
**File:** `garbage_collector.go:70-95`

### L3. Stats division by zero if an opcode has 0 records
**File:** `gnovm/cmd/benchops/stats.go:182` — `float64(len(crs))` unguarded

---

## Files to modify (for fixes)

| File | Issues |
|------|--------|
| `gnovm/pkg/gnolang/alloc.go` | C1, C2, H1 |
| `gnovm/pkg/gnolang/garbage_collector.go` | C1, L1, L2 |
| `gnovm/pkg/gnolang/store.go` | H3, H4 |
| `gnovm/pkg/gnolang/machine.go` | H2 |
| `gnovm/pkg/gnolang/uverse.go` | H5 |
| `gnovm/pkg/gnolang/preprocess.go` | C3 |
| `gnovm/pkg/benchops/bench.go` | M1 |
| `gnovm/pkg/benchops/exporter.go` | M2, M3 |
| `gnovm/cmd/benchops/run.go` | M4 |
| `gnovm/cmd/benchops/stats.go` | L3 |
