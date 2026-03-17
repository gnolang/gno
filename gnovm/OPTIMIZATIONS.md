# GnoVM Optimization Notes

## Profiling Setup

Benchmarks in `gnovm/cmd/profile/` run Gno programs through the full
`Machine.Run()` dispatch loop. Parse/preprocess happens once; only
`RunMain()` is measured.

Two benchmarks:
- **BenchmarkVM / BenchmarkVM_GasMetered** — synthetic workload: fib(20),
  sieve, struct methods, maps, closures, interfaces, type switches, pointers
- **BenchmarkContract / BenchmarkContract_GasMetered** — token contract
  simulation: transfers, allowances, balance checks (map + method heavy)

```bash
# Before/after comparison:
go test -bench=. -benchtime=10s -count=10 ./gnovm/cmd/profile/ > bench.txt
benchstat before.txt after.txt

# Profiling:
go test -bench=BenchmarkVM_GasMetered -benchtime=10s \
  -cpuprofile=cpu.prof -memprofile=mem.prof ./gnovm/cmd/profile/
go tool pprof -http=:8080 cpu.prof
```

## Profile Baseline (2026-03-16, pre-optimization)

### CPU (Apple M2)

| Category | % of total |
|---|---|
| Go runtime GC (gcDrain, gcStart, kevent) | ~41% |
| GnoVM execution (Machine.Run + children) | ~23% |
| Runtime scheduling / GC pauses | ~36% |

### Memory

**93.4% of all allocations came from `NewBlock`** — Block struct + Values slice
per function call, loop, if/switch clause.

### Server Baseline (Xeon 8168, 2 cores)

| Benchmark | ns/op |
|---|---|
| VM | 47.4ms |
| VM_GasMetered | 49.4ms |
| Contract | 2.08ms |
| Contract_GasMetered | 2.20ms |

---

## Completed Optimizations

### 1. CPU Gas Batching

Accumulate CPU gas in `Machine.cpuPending`, flush to GasMeter every ~1000 gas
instead of on every op. Reduces GasMeter interface calls ~30x.

Flush points: `OpHalt`, `GarbageCollect()`, parser callback, uverse
`consumeGas()`.

### 2. Block Arena Allocation

Pre-allocate `[]Block` arena (256 slots) per Machine. Function/loop/scope blocks
are bump-allocated from the arena. On frame return, arena slots are reclaimed in
reverse order (stack discipline). Falls back to heap if arena is full.

**Key hazard:** `EvalStatic`/`EvalStaticTypeOf` push shared static blocks via
`StaticBlock.GetBlock()`. These must NOT be arena-allocated or recycled — they
belong to the preprocessor. Only `m.newBlock()` uses the arena; `OpPopBlock`
only recycles arena blocks (detected via pointer range check).

### 3. TypedValue Slab Arena (tvArena)

Pre-allocate `[]TypedValue` slab (2048 slots) alongside the Block arena. Block
Values slices are bump-allocated from the slab instead of `make()`. Slab
position rewinds on block recycle if stack discipline holds.

Blocks whose Values grow via `ExpandWith` (switch clauses, if/else) transparently
fall back to heap — the arena slice is abandoned and Go's append allocates a new
backing array.

### 4. popCopyArgs Fast Path

For non-variadic function calls, pop args from the value stack and copy directly
into `b.Values`, skipping the intermediate `make([]TypedValue)` in
`popCopyArgs`. Variadic and deferred calls still use the old path (deferred args
are stored in `Defer.Args` and must outlive the frame).

### 5. Eliminate FuncValue.Copy on Method Dispatch

`DeclaredType.GetValueAt()` was copying the `FuncValue` on every method call just
to lazily fill the `Parent` field. Since the Machine is single-threaded and Parent
is stable once set (always the file block for package-level methods), fill Parent
on the original and skip the heap allocation.

### Combined Results (Xeon 8168)

| Benchmark | Before | After | Change |
|---|---|---|---|
| VM | 47.4ms | 28.7ms | **-39%** |
| VM_GasMetered | 49.4ms | 28.8ms | **-42%** |
| Contract | 2.08ms | 1.43ms | **-31%** |
| Contract_GasMetered | 2.20ms | 1.44ms | **-34%** |
| **geomean** | | | **-37%** |

Note: server numbers are from before optimizations #4 and #5; the full set
including those will show further improvement.

Total allocations reduced ~97% (153 GB → 4.2 GB in profiling run).
GC CPU overhead reduced from ~41% to ~8%.

---

## Attempted but Abandoned

### Block Values Slice Recycling (freelist)

Recycled `[]TypedValue` backing arrays via per-Machine freelist. Halved
allocations-per-block from 2 to 1, but Block struct allocation still dominated.
Superseded by the arena approach.

### SelectorFromTV (bypass PointerValue for method dispatch)

Added `SelectorFromTV` that constructs the method dispatch TypedValue directly
on the stack, bypassing the `GetPointerToFromTV` → `PointerValue{TV: &TypedValue{}}` →
`Deref()` path. This eliminated one heap allocation per method call (the
`&TypedValue{}` in PointerValue.TV) but the `&BoundMethodValue{}` allocation
still dominated since it must be heap-allocated to satisfy the `Value` interface.
Net improvement was ~1% — not worth the code duplication.

---

## Next Targets

### BoundMethodValue allocation (39% of contract allocs)

`GetPointerToFromTV` allocates `&BoundMethodValue{}` on the heap for every
method call (`values.go:1895,1932`). The BoundMethodValue is consumed immediately
by `doOpPrecall` to extract `Func` and `Receiver`, then discarded.

Eliminating the `&TypedValue{}` wrapper (attempted via SelectorFromTV) only
helped ~1% because `&BoundMethodValue{}` is the dominant cost. The real fix
would be to avoid BoundMethodValue entirely — e.g., store Func and Receiver
as two separate values on the Machine stack, or change the `doOpPrecall`
protocol to not require a `Value` interface wrapper.

### findEmbeddedFieldType for interface dispatch

`GetPointerToFromTV` calls `findEmbeddedFieldType()` at runtime for interface
method resolution (`values.go:1951`). This walks the type hierarchy each call.
Could cache the resolution result per (type, method name) pair.

### TypedValue.Copy for value receivers

`GetPointerToFromTV` VPValMethod path calls `dtv.Copy(alloc)` to copy the
receiver for value-method calls (`values.go:1887`). For primitive/small
receivers this allocates unnecessarily. Could specialize for common types.

### NewListArray (69% of remaining allocs)

User-visible array/slice allocations. These have unbounded lifetimes (can escape
via return, closure, realm persistence), so arena allocation is not safe. Possible
approaches: pool small fixed-size arrays (1-8 elements), or use a slab for
known-short-lived slices (e.g., append temporaries).

### Op dispatch overhead

`ifaceeq` (5.2% CPU from switch type assertions), `PopOp` (3.5%), `duffcopy`
(2.7%) are dispatch loop costs. A function table indexed by op code could
eliminate the interface comparisons in the switch.

---

## Dead End: Struct Allocation Arenas

With stdlib workloads, `NewStruct` (21.6%) + `NewStructFields` (15.4%) are the
largest remaining allocators. Three approaches were explored:

### Bump-only arenas (branches: `struct-fields-arena`, `struct-escape-analysis`)

Added svArena (StructValue) and sfArena (Fields slices) with bump allocation
and no reclaim — reset only on Machine.Release(). Also tried wiring arena
allocators into the Allocator via function pointers so StructValue.Copy
benefits too.

**Result:** Dead end. Without reclaim, arenas fill up quickly for any
non-trivial transaction and fall back to heap. Sizing arenas large enough
wastes memory (~2.6MB per Machine for 8K structs). The block arena works
because blocks follow stack discipline (reclaimed at frame pop); structs
don't — they can escape via return values, container storage, or realm
persistence.

### What would actually work

Struct allocation optimization requires either:
1. **Preprocessor escape analysis** — mark struct literals that provably
   don't escape, arena-allocate only those. Catches common cases (local
   loop variables) but misses dynamic patterns.
2. **Scope-based arena + runtime escape detection** — arena-allocate all
   structs, copy to heap on escape. Catches everything but requires
   intercepting every assignment/store that could cause escape.

Both are significant undertakings for ~5-10% improvement on struct-heavy
workloads.
