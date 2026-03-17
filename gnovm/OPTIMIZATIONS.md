# GnoVM Optimization Notes

## Profiling Setup

Benchmarks in `gnovm/cmd/profile/` run Gno programs through the full
`Machine.Run()` dispatch loop. Parse/preprocess happens once; only
`RunMain()` is measured.

Benchmarks:
- **BenchmarkVM** — synthetic: fib, sieve, structs, maps, closures,
  interfaces, type switches, pointers
- **BenchmarkContract** — token contract: transfers, allowances, balance
  checks (map + method heavy)
- **BenchmarkStdlibStrings** — strings.Builder, Split, Join, Replace,
  strconv.Atoi/Itoa
- **BenchmarkStdlibSort** — sort.IntSlice, StringSlice, custom
  sort.Interface (interface-dispatch heavy)
- **BenchmarkStdlibComplex** — record parsing, filtering, sorting, string
  building with strings/strconv/sort

Each has a `_GasMetered` variant with `GasMeter` enabled.

```bash
# Before/after comparison:
go test -bench=. -benchtime=10s -count=10 ./gnovm/cmd/profile/ > bench.txt
benchstat before.txt after.txt

# Profiling:
go test -bench=BenchmarkStdlibComplex_GasMetered -benchtime=10s \
  -cpuprofile=cpu.prof -memprofile=mem.prof ./gnovm/cmd/profile/
go tool pprof -http=:8080 cpu.prof
```

---

## Baseline (2026-03-16, pre-optimization)

### CPU

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

### 1. CPU Gas Batching (`machine.go`)

**Problem:** `GasMeter.ConsumeGas()` (an interface method call, ~7-10ns)
was called on every single VM op.

**Fix:** Accumulate CPU gas in `Machine.cpuPending` and flush to the
GasMeter when pending exceeds `cpuGasFlushThreshold` (1000). This
reduces GasMeter interface calls ~30x.

**Flush points** (to maintain correctness): `OpHalt`, `GarbageCollect()`,
parser callback, uverse `consumeGas()`. Max overshoot beyond gas limit
is bounded by threshold + one op's max cost (~1424), negligible vs
typical gas limits of millions.

### 2. Block Arena (`machine.go`)

**Problem:** Every function call, loop, and if/switch clause heap-allocated
a `Block` struct. This was 93% of all allocations and drove 41% GC overhead.

**Fix:** Pre-allocate `[]Block` arena (256 slots) per Machine.
`m.newBlock()` bump-allocates from the arena instead of `&Block{}`.
On frame return, arena slots are reclaimed in reverse order (stack
discipline). Falls back to `Allocator.NewBlock()` (heap) if full.

**Key hazard:** `EvalStatic`/`EvalStaticTypeOf` push shared static blocks
(from `StaticBlock.GetBlock()`) onto the block stack and pop them with
`OpPopBlock`. These must NOT be recycled — `recycleBlock()` detects arena
vs non-arena blocks via pointer range check (`blockArenaIdx()`).

### 3. TypedValue Slab Arena (`machine.go`)

**Problem:** Each `NewBlock` call also heap-allocated a `[]TypedValue`
slice for `Block.Values` via `make()`.

**Fix:** Pre-allocate `[]TypedValue` slab (2048 slots) alongside the
Block arena. `newBlock()` bump-allocates Values from the slab. Slab
position rewinds on block recycle if stack discipline holds.

Blocks whose Values grow via `ExpandWith` (switch clauses, if/else)
transparently fall back to heap — the arena slice is abandoned and
Go's append allocates a new backing array.

### 4. popCopyArgs Fast Path (`op_call.go`)

**Problem:** `popCopyArgs()` allocated `make([]TypedValue, numParams)` on
every function call to temporarily hold arguments before copying them
into the block.

**Fix:** For non-variadic calls (the common case), pop args from the
value stack and copy directly into `b.Values`, skipping the intermediate
allocation. Variadic and deferred calls still use `popCopyArgs` (deferred
args are stored in `Defer.Args` and must outlive the frame).

### 5. Eliminate FuncValue.Copy (`types.go`)

**Problem:** `DeclaredType.GetValueAt()` called `FuncValue.Copy()` on every
method dispatch just to lazily fill the `Parent` field. This heap-allocated
a new FuncValue per method call.

**Fix:** Fill `Parent` lazily on the original FuncValue. Safe because the
Machine is single-threaded and Parent is immutable once set (always the
package's file block for non-closure methods).

### 6. OpMethodPrecall — Direct Method Dispatch (`op_call.go`, `op_eval.go`)

**Problem:** Method calls (`obj.Method(args)`) went through:
`OpSelector` → heap-allocate `&BoundMethodValue{}` + `&TypedValue{}` →
push to stack → `OpPrecall` → extract Func+Receiver → discard BMV.
Two heap allocations per method call, consumed one op later.

**Fix:** Add `OpMethodPrecall` that combines selector lookup + precall.
When `doOpEval` sees a `CallExpr` whose Func is a `SelectorExpr` with a
method path, it pushes `OpMethodPrecall` instead of `OpSelector + OpPrecall`.
The new op resolves the method and pushes the call frame directly — zero
heap allocations.

Handles concrete methods (VPValMethod, VPPtrMethod, VPDerefValMethod,
VPDerefPtrMethod) and interface methods (VPInterface, VPDerefInterface).
Stored method values (`f := obj.Method`) still use the old
BoundMethodValue path.

### 7. Interface Method Trail Cache (`types.go`)

**Problem:** Interface method dispatch calls `findEmbeddedFieldType()` at
runtime to resolve a method Name on the concrete type. This walks the type
hierarchy (methods list + embedded fields) on every call, even for the same
(type, method) pair.

**Fix:** Add `DeclaredType.methodMap` — a `map[Name][]ValuePath` that caches
the trail for exported method names. `LookupMethodTrail()` checks the cache
first, falls back to `findEmbeddedFieldType` on miss, and stores the result.
The cache is per-type and permanent (types are immutable after preprocessing),
bounded by the number of exported methods per type (typically 1-10).

Used by `doOpMethodPrecall` (VPInterface path) and `GetPointerToFromTV`
(VPInterface path for non-call selectors).

### 8. Lazy Exception Allocation (`op_binary.go`)

**Problem:** `quoAssign()` and `remAssign()` allocated
`&Exception{Value: typedString("division by zero")}` on **every** `/` and
`%` operation, even when the divisor was non-zero. This was 48% of
contract benchmark allocations.

**Fix:** Only allocate the exception on the error path (actual division
by zero).

---

## Results

### Server (Xeon 8168, optimizations 1-3 only)

| Benchmark | Before | After | Change |
|---|---|---|---|
| VM | 47.4ms | 28.7ms | **-39%** |
| VM_GasMetered | 49.4ms | 28.8ms | **-42%** |
| Contract | 2.08ms | 1.43ms | **-31%** |
| Contract_GasMetered | 2.20ms | 1.44ms | **-34%** |
| **geomean** | | | **-37%** |

### Local (Apple M2, all 8 optimizations)

Approximate cumulative improvement vs pre-optimization baseline:

| Benchmark | Before | After | Change |
|---|---|---|---|
| VM_GasMetered | ~17ms | ~11.5ms | **-32%** |
| Contract_GasMetered | ~740us | ~508us | **-31%** |
| StdlibSort_GasMetered | ~20ms | ~12.3ms | **-39%** |
| StdlibComplex_GasMetered | ~25ms | ~13.2ms | **-47%** |

### Allocation reduction

| Metric | Before | After |
|---|---|---|
| Block allocs (% of total) | 93.4% | ~0% (arena) |
| BoundMethodValue allocs | 39% of contract | ~0% (OpMethodPrecall) |
| Exception allocs (% / %) | 48% of contract | 0% (lazy) |
| findEmbeddedFieldType | per-call walk | cached trail |
| GC CPU overhead | ~41% | ~14% |

---

## Current Profile (post all optimizations)

### CPU (StdlibComplex_GasMetered, M2)

| Function | % | Notes |
|---|---|---|
| scanobject (GC) | 14% | Driven by struct allocs |
| doOpEval | 14% | Expression dispatch |
| madvise (runtime) | 6% | OS memory management |
| doOpExec | 5% | Statement dispatch |
| memclrNoHeapPointers | 4% | Arena/slice zeroing |
| incrCPU | 2% | Gas batching |

No single VM function dominates. Remaining overhead is split between
GC (driven by struct allocs), the op dispatch loop, and OS-level
memory management.

### Memory (StdlibComplex_GasMetered, M2)

| Allocator | % | Notes |
|---|---|---|
| NewStruct | 21% | Struct literal + Copy |
| NewStructFields | 14% | Struct fields slice |
| GetPointerAtIndex | 7% | Array/slice indexing |
| NewListArray | 6% | Slice allocation |
| GetPointerToFromTV | 4% | Non-method selector paths |
| doOpConvert | 4% | Type conversions |

---

## Attempted but Abandoned

### Block Values Slice Recycling (freelist)

Recycled `[]TypedValue` backing arrays via per-Machine freelist. Halved
allocations-per-block from 2 to 1, but Block struct allocation still
dominated. Superseded by the arena approach.

### SelectorFromTV (bypass PointerValue for method dispatch)

Added `SelectorFromTV` that returns TypedValue directly instead of going
through `GetPointerToFromTV` → `PointerValue{TV: &TypedValue{}}` → `Deref()`.
Eliminated one heap allocation per method call but `&BoundMethodValue{}`
dominated. Net ~1% — not worth the code duplication. Superseded by
OpMethodPrecall which eliminates both allocations.

### Struct Allocation Arenas

Added svArena (StructValue) and sfArena (Fields slices) with bump
allocation. Also tried wiring arena allocators into the Allocator via
function pointers so `StructValue.Copy` benefits.

**Dead end.** Without reclaim, arenas fill up quickly and fall back to
heap. Block arenas work because blocks follow stack discipline (reclaimed
at frame pop); structs can escape via return values, container storage,
or realm persistence. Sizing arenas large enough wastes memory.

### Copy-on-Write for StructValue

Explored deferring `StructValue.Copy` until a field is actually written.
**Impractical** with the current architecture: the write path goes through
`PointerValue{TV: &sv.Fields[idx]}` which has no reference back to the
`TypedValue` holding the `*StructValue` pointer. Intercepting writes would
require either a level of indirection (slows all reads) or a back-reference
in PointerValue (increases its size).

---

## Remaining Targets (diminishing returns)

All remaining targets are estimated at 2-3% each:

### Struct allocations (35% of stdlib allocs)

`NewStruct` (21%) + `NewStructFields` (14%) are the largest remaining
allocators. Requires either preprocessor escape analysis or runtime escape
detection — both are significant undertakings. See "Attempted but
Abandoned" section for details on approaches tried.

### Op dispatch overhead

`ifaceeq` (3% CPU), `PopOp` (2.5%), `duffcopy` (2%) are dispatch loop
costs. A function table indexed by op code could eliminate the interface
comparisons in the Run() switch.

### TypedValue.Copy (6% CPU)

Called for every struct/array value pass. Many copies are unnecessary
(source not modified after). Could skip for immutable values or use
copy-on-write (requires architecture changes, see "Attempted but
Abandoned" section).
