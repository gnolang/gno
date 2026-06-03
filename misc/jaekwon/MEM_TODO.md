# Gas Calibration System: Model CPU Time Accurately

## Context

GnoVM gas charges three independent costs through one GasMeter: CPU (per-opcode), Memory (per-byte allocated), and GC (per-object visited). The problem: **memory gas overcharges ~10-20x relative to CPU time**.

- 1 CPU gas unit ~ 3.6ns (from benchops: OpAdd=65ns / OpCPUAdd=18)
- Allocating a 208-byte StructValue charges 208 memory gas = 624ns equivalent
- But Go's actual malloc costs ~25-40ns flat regardless of size
- ~~`VisitCpuFactor=8` guessed~~ → replaced by `gcVisitGasTable` (calibrated)

**Goal**: Make gas constants reflect actual CPU time on the host machine, with a universal `GasFactor=1000` for 0.1% precision across all gas layers, and a Makefile target for easy recalibration.

### GasFactor = 1000

All gas layers (CPU, memory, GC, store) are multiplied by `GasFactor`.
This gives 0.1% precision with ~384 days of CPU time headroom before int64 overflow.

- `GasFactorCPU` renamed to `GasFactor` (universal)
- `GasFactor = 1000`
- Applied to: CPU opcodes, memory allocations, GC visits, store operations
- 1 gas unit ~ 0.0036ns

---

## Step 1: Rename GasFactorCPU -> GasFactor and set to 1000

**Modify: `gnovm/pkg/gnolang/machine.go`**

Rename `GasFactorCPU` to `GasFactor`, change value from 1 to 1000.
Update `incrCPU()` to use `GasFactor`.

**Modify: `gnovm/pkg/gnolang/garbage_collector.go`**

Update `GasFactorCPU` reference to `GasFactor`.

**Modify: `gnovm/pkg/gnolang/alloc.go`**

Memory gas charging uses `GasFactor` (see Step 3c).

**Modify: `gnovm/pkg/gnolang/store.go`**

Scale store gas config by `GasFactor`:
```go
GasGetObject:    16 * GasFactor,   // per byte
GasSetObject:    16 * GasFactor,   // per byte
GasDeleteObject: 3715 * GasFactor, // flat
GasGetType:      10 * GasFactor,   // per byte
```

**Modify: `tm2/pkg/store/types/gas.go`**

Scale KV store defaults by `GasFactor` (or keep as-is and scale at the GnoVM layer).
Preferred: scale at GnoVM layer only, since tm2 is shared infrastructure.

---

## Step 2: Restructure allocation gas: per-alloc instead of per-byte

**Modify: `gnovm/pkg/gnolang/alloc.go`**

### 2a. Replace GasCostPerByte with GasCostPerAlloc

```go
const GasCostPerAlloc int64 = 10 // gas per heap allocation (calibrated, before GasFactor)
```

Remove `GasCostPerByte = 1`. Gas is now flat per-allocation, not per-byte.
Byte tracking remains for memory limit enforcement only.

### 2b. Add allocation count tracking to Allocator

```go
type Allocator struct {
    maxBytes   int64
    bytes      int64
    peakBytes  int64
    allocs     int64  // count of allocations
    peakAllocs int64  // high-water mark for gas charging
    collect    func() (left int64, ok bool)
    gasMeter   store.GasMeter
}
```

### 2c. Modify Allocate() gas charging

```go
func (alloc *Allocator) Allocate(size int64) {
    // ... existing memory limit check with GC (unchanged) ...
    alloc.bytes += size
    alloc.allocs++

    // Charge gas on allocation count high-water mark.
    if alloc.allocs > alloc.peakAllocs {
        if alloc.gasMeter != nil {
            delta := alloc.allocs - alloc.peakAllocs
            alloc.gasMeter.ConsumeGas(
                overflow.Mulp(overflow.Mulp(delta, GasCostPerAlloc), GasFactor),
                "memory allocation")
        }
        alloc.peakAllocs = alloc.allocs
    }
    // Track peak bytes for memory limit enforcement (no gas).
    if alloc.bytes > alloc.peakBytes {
        alloc.peakBytes = alloc.bytes
    }
}
```

### 2d. Wire allocs into GC and Reset

- `Reset()`: set `allocs = 0` (alongside `bytes = 0`)
- GC visitor already calls `alloc.Allocate(size)` per visited object, which increments `allocs` -- so after GC, `allocs` reflects the count of live objects
- `Fork()`: same pattern as bytes (not copied, intentional for query contexts)

---

## Step 3: Update VisitCpuFactor — DONE

**Modified: `gnovm/pkg/gnolang/garbage_collector.go`**

Replaced flat `VisitCpuFactor=8` with `gcVisitGasTable`: 25-entry lookup table
indexed by `log2(visitCount)`. Per-visit gas scales with heap size due to CPU
cache effects (L1→L2→L3→DRAM). Calibrated from BenchmarkGCVisit on DO Xeon 8168.
See `bench_gc_test.go` for the GC visit benchmarks.

---

## Step 4: Create calibration benchmarks

**New file: `gnovm/pkg/gnolang/calibrate/calibrate_test.go`**

Standard Go benchmark tests measuring raw Go runtime costs:

### 4a. Flat allocation cost
Benchmark `new(StructValue)`, `new(ArrayValue)`, `new(Block)`, etc.
- Pre-populates ~100MB of live GnoVM-style objects for realistic GC pressure
- `runtime.GC()` + `b.ResetTimer()` before measuring
- Allocates in tight loop, stores pointers to prevent dead-store elimination

### 4b. Per-byte marginal cost
Compare `make([]byte, 64)` vs `make([]byte, 4096)` vs `make([]byte, 65536)`.

### 4c. GC visit cost — DONE
Build realistic object graph, time `GCVisitorFn` traversal. Measures ns/visit.
Implemented in `bench_gc_test.go` with 100 to 10M objects. Results used to
calibrate `gcVisitGasTable` in Step 3.

**New file: `gnovm/pkg/gnolang/calibrate/prepopulate.go`**

`PrePopulateMemory(sizeMB int) []interface{}` -- realistic GnoVM heap for GC pressure.

---

## Step 5: Create constant generator tool

**New file: `gnovm/pkg/gnolang/calibrate/cmd/generate/main.go`**

Parses `go test -bench` output, computes calibrated constants:

```
cpuBaseNs = 3.6  (overridable via -cpubase flag)

GasCostPerAlloc = round(allocAvgNs / cpuBaseNs)  // before GasFactor
VisitCpuFactor  = round(visitNs / cpuBaseNs)       // before GasFactor
```

Prints Go source for manual copy into alloc.go / garbage_collector.go.

---

## Step 6: Makefile target

**Modify: `gnovm/Makefile`**

```makefile
.PHONY: calibrate.gas
calibrate.gas:
	@mkdir -p .tmp
	@echo "=== Allocation benchmarks ==="
	go test -bench=BenchmarkAlloc -benchmem -count=10 -timeout=10m \
		./pkg/gnolang/calibrate/ | tee .tmp/calibrate_alloc.txt
	@echo "=== GC visit benchmarks ==="
	go test -bench=BenchmarkGCVisit -benchmem -count=10 -timeout=10m \
		./pkg/gnolang/calibrate/ | tee .tmp/calibrate_gc.txt
	@echo "=== Calibrated constants ==="
	go run ./pkg/gnolang/calibrate/cmd/generate \
		-alloc .tmp/calibrate_alloc.txt \
		-gc .tmp/calibrate_gc.txt \
		-cpubase 3.6
```

---

## Step 7: Update golden test values

- `gnovm/tests/files/gas/*.gno` (3 files)
- `gnovm/tests/files/alloc_*.gno` (~9 files)
- `gno.land/pkg/integration/testdata/gc.txtar`

Run: `go test pkg/gnolang/files_test.go -test.short --update-golden-tests`

---

## Files to create

| File | Purpose |
|------|---------|
| `gnovm/pkg/gnolang/calibrate/calibrate_test.go` | Allocation + GC benchmarks |
| `gnovm/pkg/gnolang/calibrate/prepopulate.go` | Memory pre-population for realistic GC pressure |
| `gnovm/pkg/gnolang/calibrate/cmd/generate/main.go` | Parse benchmarks -> emit calibrated constants |

## Files to modify

| File | Change |
|------|--------|
| `gnovm/pkg/gnolang/machine.go` | Rename GasFactorCPU -> GasFactor, set to 1000 |
| `gnovm/pkg/gnolang/alloc.go` | GasCostPerAlloc, allocs/peakAllocs, GasFactor in gas formula |
| `gnovm/pkg/gnolang/garbage_collector.go` | Update VisitCpuFactor, use GasFactor |
| `gnovm/pkg/gnolang/store.go` | Scale store gas config by GasFactor |
| `gnovm/Makefile` | Add `calibrate.gas` target |
| `gnovm/tests/files/gas/*.gno` | Update golden gas values |
| `gnovm/tests/files/alloc_*.gno` | Update golden alloc values |
| `gno.land/pkg/integration/testdata/gc.txtar` | Update expected gas |

## Verification

1. `make calibrate.gas` -- runs benchmarks, prints calibrated constants
2. Update constants in alloc.go + garbage_collector.go
3. `go test pkg/gnolang/files_test.go -test.short --update-golden-tests`
4. `go test ./pkg/gnolang/... -count=1` -- all tests pass
5. Verify gas ratios: struct allocation ~10,000 gas (10 * 1000), OpAdd ~18,000 gas (18 * 1000)

## Note on consensus

Changing gas constants is consensus-breaking. Must be gated behind a chain upgrade version.
