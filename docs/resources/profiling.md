# Gno Profiling Guide

This document explains how to enable and consume the VM profiler that mirrors `go tool pprof`
functionality. The new architecture is built on top of the instrumentation layer and works for
unit tests, CLI commands, and any custom tooling that can pass an instrumentation sink.

## Quick Start

Profiling is currently exposed through `gno test`. Flags allow you to choose the profile type,
sampling rate, and output format:

```shell
gno test ./path/to/pkg \
  -profile \
  -profile-type cpu \
  -profile-format toplist \
  -profile-output cpu.out
```

### Key Flags

| Flag | Description |
|------|-------------|
| `-profile` | Enables profiling for the test run. |
| `-profile-type {cpu|memory|gas}` | Choose the metric to record. Default is `cpu`. |
| `-profile-format {text|json|toplist|calltree}` | Output format. Defaults to `text`. |
| `-profile-sample-rate N` | Record every Nth sample (default depends on profile type). Use `1` for very short tests. |
| `-profile-output file` | Write profile data to a file. Defaults to `profile.out`. |
| `-profile-stdout` | Print the report to stdout instead of writing to a file. |
| `-profile-list func` | Produce a line-by-line listing for the matching function. Automatically enables line-level profiling. |

Examples:

```shell
# CPU profile, textual toplist, sample every operation, print to stdout
gno test ./examples/gno.land/p/nt/avl \
  -run TestTreeGetByIndex \
  -profile \
  -profile-format toplist \
  -profile-sample-rate 1 \
  -profile-stdout

# Line-level view for a specific function
gno test ./examples/gno.land/p/nt/avl \
  -run TestTreeGetByIndex \
  -profile \
  -profile-list gno.land/p/nt/avl.GetByIndex \
  -profile-sample-rate 2

# Memory allocations (sample every allocation automatically)
gno test ./examples/... \
  -profile \
  -profile-type memory \
  -profile-output heap.json \
  -profile-format json
```

## Outputs

Depending on the format you choose, the profiler can emit:

* **Text**: Human-readable list similar to `go tool pprof`.
* **Toplist**: Sorted table emphasizing cumulative/flat time with simple ASCII bars.
* **Call tree**: Structured JSON suitable for flame graphs or other visualizations.
* **JSON**: Raw profile data useful for custom analysis scripts.
* **Line list**: When `-profile-list` is set, the output mirrors `go tool pprof list`, showing
  per-line metrics for a specific function, complete with source context fetched from Gno stores.

All formats leverage the same underlying profile samples collected via the instrumentation sink.

For example, the following snippet shows a CPU toplist generated with `-profile-sample-rate 1`:

```text
$ gno test -v . -run TestTreeGetByIndex \
    -profile -profile-stdout \
    -profile-format toplist \
    -profile-sample-rate 1

=== PROFILING RESULTS ===
Top Functions by CPU Cycles
Total cycles: 13985435

Rank   Cumulative   Cum%     Flat         Flat%    Calls    Bar                  Function
----------------------------------------------------------------------------------------------------
1      222651549    1592.02% 222651549    1592.02% 2267     ████████████████████ testing.RunTest
…
23     3823888      27.34  % 262479       1.88   % 15       █████                gno.land/p/nt/avl.GetByIndex
…
```

And the line-level listing produced with `-profile-list gno.land/p/nt/avl.GetByIndex`:

```text
$ gno test -v . -run TestTreeGetByIndex \
    -profile -profile-stdout \
    -profile-format toplist \
    -profile-sample-rate 2 \
    -profile-list gno.land/p/nt/avl.GetByIndex

=== FUNCTION PROFILE ===
ROUTINE ======================== gno.land/p/nt/avl.GetByIndex in tree.gno
…
ROUTINE ======================== gno.land/p/nt/avl.GetByIndex in node.gno
 39287893   39287893 (flat, cum) 83.35% of Total
 9158604    9158604  line 111: if node.height == 0 {
 8516427    8516427  line 123: return node.getRightNode().GetByIndex(index - leftNode.size)
…
```

## How It Works

1. **Instrumentation sink**: The VM emits `SampleContext`, `AllocationEvent`, and `LineSample`
   events whenever the sink declares interest through the `Capabilities` interface.
2. **Profiler adapter**: `profiler.SinkAdapter` implements the instrumentation sink and bridges
   events into the existing profiler data structures.
3. **Test harness**: `pkg/test` constructs a `ProfileConfig`, attaches the profiler sink to every
   `Machine` it creates, and writes results via `DefaultProfileWriter` after the run completes.
4. **CLI flags**: `gno test` exposes the configuration knobs so developers can opt-in on demand.

Internally, the flow mirrors the Go toolchain: sampling occurs at `OpCall` boundaries (CPU/Gas) and
per-allocation (memory), with default rates tuned for most workloads. For tiny tests, lower the
sample rate to `1` to capture meaningful data.

## Tips

* **Short tests**: Use `-profile-sample-rate 1` so the sampler actually fires.
* **Line listings**: `-profile-list func.name` implicitly enables line-level tracking and works best
  with moderate sample rates (e.g., `1` or `2`).
* **Memory/Gas**: These modes require a functioning allocator/gas meter. The test harness wires them
  automatically, but if you embed the profiler elsewhere, ensure you set `Machine.GasMeter` when
  collecting gas profiles.
* **External analysis**: Files written with `-profile-output` can be consumed by the profiler CLI
  (`gnovm/pkg/profiler/cli.go`) or converted to other formats as needed.

## Extending Beyond `gno test`

Any custom VM driver can adopt the profiler by:

1. Creating a `profiler.Profiler` (or custom instrumentation sink).
2. Calling `Machine.StartProfilingWithSink(sink, opts)` before executing code.
3. Invoking `Machine.StopProfiling()` to retrieve the accumulated profile.

This approach keeps the profiler decoupled from VM internals and allows future tooling (e.g.,
transactions, repl sessions) to reuse the same hooks.
