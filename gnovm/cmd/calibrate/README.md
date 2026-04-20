# Gas Calibration Benchmarks

Tools and data for calibrating GnoVM's gas constants. Three sets of constants
are calibrated from this directory:

| Constants | Target File | Data Source | Analysis Tool |
|---|---|---|---|
| `allocGasTable` (32 entries) | `alloc.go` | `alloc_bench_test.go` (local) | `gen_alloc_table.py` |
| `OpCPU*` + `OpCPUSlope*` (206 constants) | `machine.go` | `bench_ops_test.go` (pkg/gnolang) | `gen_analysis.py` + `plot_fits.py` |
| `gcVisitGasTable` (25 entries) | `garbage_collector.go` | `bench_gc_test.go` (pkg/gnolang) | Manual (no script yet) |

## Model

The allocation gas model uses:

- **6 exact table entries** (1B–32B): directly from benchmark medians, covering Go's tiny/small allocator tier where costs are flat-ish.
- **Power-law fit** (64B+): `ns = C × size^α`, a straight line in log-log space. The power-law captures the sublinear growth due to Go's `memclr` zero-fill dominating large allocation cost.

At runtime, `allocGas()` uses a 32-entry lookup table indexed by `bits.Len64(size)` with linear interpolation between entries — O(1), ~1.5ns.

## Benchmark Machine

The reference benchmarks (`bench_output_do_dedicated.txt`) were run on:

- **Provider**: DigitalOcean Dedicated Droplet
- **OS**: Linux (goos: linux)
- **Arch**: amd64 (goarch: amd64)
- **CPU**: Intel Xeon Platinum 8168 @ 2.70GHz, 2 cores
- **Go**: 1.24
- **Flags**: `-bench=BenchmarkAlloc -benchmem -count=5 -timeout=30m`
- **Date**: 2026-03-14

Power-law fit result: `ns = 0.47 × size^0.925` (α = 0.925).

Gas unit: 1 gas = 1 nanosecond of wall time on the reference hardware.

## Data Sources

This directory consumes benchmark output from three places:

- **Allocation benchmarks** (`alloc_bench_test.go`): local Go microbenchmarks in this directory. Measures `make([]byte, N)` for N = 1B to 1GB.
- **Opcode benchmarks** (`op_bench_do_dedicated.txt`): produced by running `BenchmarkOp*` tests in `gnovm/pkg/gnolang/bench_ops_test.go`, which use the `gnovm/pkg/benchops` instrumentation framework to report `alloc-gas/op` and `ns/op(pure)` metrics. Consumed by `gen_analysis.py` (gas constants) and `plot_fits.py` (visualization).
- **GC visit benchmarks** (`gc_visit_bench_do_dedicated.txt`): produced by running `BenchmarkGCVisit*` tests in `gnovm/pkg/gnolang/bench_gc_test.go`. Measures per-visit GC traversal cost across heap sizes, capturing CPU cache effects (L1/L2 ~29ns → L3 ~91ns → DRAM ~700ns). Currently consumed manually (no analysis script).

## Usage

### Allocation gas calibration

```bash
# 1. Run allocation benchmarks
cd gnovm/cmd/calibrate
go test -bench=BenchmarkAlloc -benchmem -count=5 -timeout=30m . > bench_output.txt

# 2. Generate allocGasTable and plot
python3 gen_alloc_table.py bench_output.txt --cpu-base-ns 1.0

# 3. Update alloc.go with the printed table (divide milligas by 1000)
```

### Opcode gas calibration

```bash
# 1. Run opcode benchmarks (from repo root)
cd gnovm
go test -run=^$ -bench='BenchmarkOp' -benchtime=2s -count=3 -timeout=60m \
    ./pkg/gnolang/ 2>&1 | tee cmd/calibrate/op_bench.txt

# 2. Generate analysis report with Go constant declarations
cd cmd/calibrate
python3 gen_analysis.py op_bench.txt > op_bench_analysis.txt

# 3. Generate fit visualization
python3 plot_fits.py op_bench.txt op_gas_fits.png

# 4. Copy OpCPU* constants from the report into machine.go
```

### GC visit gas calibration

```bash
# 1. Run GC visit benchmarks (from repo root)
cd gnovm
go test -run=^$ -bench='BenchmarkGCVisit' -benchtime=2s -count=3 -timeout=60m \
    ./pkg/gnolang/ 2>&1 | tee cmd/calibrate/gc_visit_bench.txt

# 2. Manually update gcVisitGasTable in garbage_collector.go
#    by reading the ns/visit column for each heap size tier.
```

## Files

| File | Description |
|------|-------------|
| **Allocation pipeline** | |
| `alloc_bench_test.go` | Go benchmarks: `make([]byte, N)` for N = 1B to 1GB |
| `gen_alloc_table.py` | Parses benchmark output, fits power-law model, generates table + plot |
| `alloc_gas_model.png` | Plot of allocation cost model vs actual benchmarks |
| `bench_output_do_dedicated.txt` | Reference alloc benchmark data (DO Dedicated, amd64, Go 1.24) |
| `bench_output_do_amd64.txt` | Older alloc data from DO Regular (shared vCPU), kept for comparison |
| `bench_output_m2_arm64.txt` | Alloc data from Apple M2, for cross-platform comparison |
| **Opcode pipeline** | |
| `gen_analysis.py` | Parses opcode benchmarks, fits linear/quadratic models, generates OpCPU* constants |
| `plot_fits.py` | Generates multi-panel visualization of opcode gas fits |
| `op_bench_do_dedicated.txt` | Reference opcode benchmark data (DO Dedicated, amd64) |
| `op_bench_analysis.txt` | Generated analysis report with gas constants and fit quality |
| `op_gas_fits.png` | Multi-panel plot of parameterized opcode gas fits |
| `op_gas_formulas.md` | Reference: expected formula shapes for all opcodes |
| `benchops_stats_do_dedicated.csv` | Aggregate opcode statistics (from legacy `cmd/benchops` tool) |
| **GC visit pipeline** | |
| `gc_visit_bench_do_dedicated.txt` | Reference GC visit benchmark data (DO Dedicated, amd64) |
| **Other** | |
| `sha256_bench_test.go` | SHA256 benchmarks at various data sizes |
