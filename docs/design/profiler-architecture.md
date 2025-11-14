# Gno Profiler Architecture

## Background

The current profiling prototype lives directly inside the VM (`gnovm/pkg/gnolang`) and the allocator. The VM owns a concrete `*profiler.Profiler` (`gnovm/pkg/gnolang/machine.go:42-66`), emits CPU samples from `Run()` (`gnovm/pkg/gnolang/machine.go:1270-1353`), and exposes helper methods that simply forward to the profiler (`gnovm/pkg/gnolang/machine_profiling.go:90-128`). Memory accounting is hard-wired by embedding a `*Machine` in `Allocator` (`gnovm/pkg/gnolang/alloc.go:11-182`) and calling back into `Machine.RecordAllocation`.

This structure enables the initial feature set but introduces several problems:

- **Allocator ↔ Machine cycle** – `Allocator` now retains a concrete VM pointer, complicating reuse, forking, and tests.
- **Duplicated accounting logic** – `allocateWithType` re-implements `Allocate` and can double-count when GC kicks in (`gnovm/pkg/gnolang/alloc.go:157-181`), creating correctness and maintenance risk.
- **Hot-path instrumentation** – `Run()` inspects expressions/statements for every single opcode to support line profiling (`gnovm/pkg/gnolang/machine.go:1286-1323`) even when profiling is disabled, and every call instruction hits `RecordProfileSample`, which grabs locks and adapters even if no sample is taken.
- **Profiler knows too much about VM internals** – `profiler.RecordSample` pulls `Frames`, `Cycles`, and `Gas` through adapters, effectively binding the profiler to specific VM data structures.

The question is whether we can design a profiler that remains largely independent from `Machine`/`Allocator` while still delivering CPU, gas, memory, and line-level views. The answer is yes: introduce a narrow instrumentation layer that publishes VM events, and let the profiler subscribe to that layer.

## Goals

- Keep the profiler and any future tooling decoupled from VM internals.
- Avoid allocator ↔ machine references; the allocator should only report events.
- Ensure instrumentation overhead is practically zero when profiling is disabled.
- Support current features (CPU/gas samples, heap allocations, line listings) and keep the door open for future ones (goroutines, I/O).

Non-goals: redesigning the VM execution model, adding sampling-based GC, or matching `go tool pprof` output bit-for-bit in this iteration.

## Proposed Architecture

### 1. Instrumentation Hub

Create a lightweight `instrumentation` package that defines the events the VM can emit:

```go
// instrumentation/events.go
type SampleContext struct {
    Frames []FrameSnapshot
    Cycles int64
    GasUsed int64
}

type FrameSnapshot struct {
    FuncName string
    File     string
    Line     int
    Column   int
    PkgPath  string
}

type AllocationEvent struct {
    Bytes     int64
    Objects   int64
    Kind      string
    Stack     []FrameSnapshot
}

type LineSample struct {
    Func string
    File string
    Line int
    Cycles int64
}

type Sink interface {
    OnSample(*SampleContext)
    OnAllocation(*AllocationEvent)
    OnLineSample(*LineSample)
}
```

`Machine` would only hold a nullable `Sink` (or a small array of sinks) inside `MachineOptions`, e.g. `MachineOptions.Instrumentation instrumentation.Sink`. When profiling is off, the sink is `nil` and the VM performs zero extra work aside from a single nil check.

### 2. Sample Capture

Instead of letting the profiler introspect machine state, the VM captures the minimal snapshot once per sample:

1. `Machine` keeps a cheap `shouldSample()` helper that increments a local counter and returns true every Nth `OpCall`.
2. When `shouldSample()` is true and a sink exists, `Machine` builds a `SampleContext`:
   - Convert the current frames into `FrameSnapshot` structs (reuse scratch buffers from a `sync.Pool` to keep allocations low).
   - Read `Cycles`, `GasMeter.GasConsumed()`, and any other counters.
3. Emit the context to the sink: `m.instrumentation.OnSample(ctx)`.

The profiler no longer needs adapters or direct access to `Machine`; it simply consumes snapshots.

### 3. Allocation Tracing

`Allocator` is modified to depend on a tiny interface instead of a VM pointer:

```go
type AllocationSink interface {
    OnAllocation(event *AllocationEvent)
}

func (alloc *Allocator) SetAllocationSink(s AllocationSink)
```

`NewMachineWithOptions` wires the machine’s sink into both the VM and the allocator. Allocation methods go back to calling the single `Allocate()` helper. After `Allocate()` succeeds, it checks `alloc.sink != nil` and emits an `AllocationEvent` (re-using the stack snapshots captured from the most recent CPU sample or capturing on demand when memory profiling is enabled).

This removes the `allocateWithType` duplication and eliminates the double-accounting bug while keeping the profiler independent from the allocator itself.

### 4. Line-Level Sampling

Line profiling becomes another optional event:

- When line profiling is requested, the profiler sets a flag on its sink implementation.
- Inside the VM loop, we move the existing line-detection logic behind a guard:

```go
if m.instrumentation != nil && m.instrumentation.WantsLineSamples() {
    if ls, ok := m.captureLineSample(); ok {
        m.instrumentation.OnLineSample(ls)
    }
}
```

`captureLineSample` reuses current helper logic but executes only when a sink explicitly opts in, eliminating overhead for the default case.

### 5. Profiler Implementation

`gnovm/pkg/profiler` no longer imports `gnolang`. Instead, it implements `instrumentation.Sink`. Each callback simply feeds the internal profile builder:

- `OnSample` updates CPU/gas statistics and stores stack samples.
- `OnAllocation` records heap allocations (sampling based on the profiler’s `SampleRate`).
- `OnLineSample` updates the line table used by `-profile-list`.

Because the profiler now receives fully-populated snapshots, it becomes usable in other contexts (e.g., profiling replay traces) and test code can inject mock sinks without spinning up a full VM.

### 6. CLI and Test Integration

`gnovm/pkg/test` already passes profiling options around. With the new structure:

1. Tests create a profiler via `profiler.New()`.
2. They request an instrumentation sink from it (`sink := profiler.NewSink()`).
3. When creating machines/stores, they pass `MachineOptions.Instrumentation = sink`.
4. After execution they call `profiler.Stop()` and write the collected output.

No VM-facing code needs to import the profiler package; it only speaks in terms of the instrumentation interfaces.

## Transition Plan

1. **Introduce instrumentation scaffolding** – define the `Sink`, event structs, and helpers under a new package.
2. **Refactor allocator** – remove `machine *Machine`, re-use `Allocate`, and emit events via `AllocationSink`.
3. **Refactor VM run loop** – replace direct profiler calls with instrumentation hooks guarded by capability checks.
4. **Port profiler implementation** – adapt the existing profiler to consume events, keeping all CLI/test APIs intact.
5. **Adjust tests/CLI** – wire sinks via `MachineOptions`; ensure profiling still functions end-to-end.
6. **Follow-up tuning** – profile the profiler to confirm overhead stays beneath existing thresholds and optimize buffer pools if needed.

## Open Questions & Risks

- **Stack snapshot cost** – capturing full stacks every sample requires allocations unless carefully pooled. Need benchmarks to pick suitable pool sizes and sampling rates.
- **Allocation attribution** – deciding whether to reuse the last captured stack or to capture synchronously on allocation affects accuracy vs. performance.
- **Future event types** – goroutine or I/O profiling may need additional events; the instrumentation package should remain generic enough to grow without leaking VM details.

Despite these questions, the instrumentation-first design achieves the key objective: profiling functionality that is largely independent from `Machine` and `Allocator`, with clean seams for future profilers or observability tools.
