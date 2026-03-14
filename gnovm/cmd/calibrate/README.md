# Gas Calibration Benchmarks

Tools and data for calibrating GnoVM's allocation gas model (`allocGasTable` in `alloc.go`).

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

`cpuBaseNs = 5.2` — the weighted-average ns per gas unit, computed from `benchops` on the same hardware. Used to convert benchmark ns to gas units: `gas = ns / cpuBaseNs`.

## Usage

### 1. Run benchmarks on target hardware

```bash
cd gnovm/cmd/calibrate
go test -bench=BenchmarkAlloc -benchmem -count=5 -timeout=30m . > bench_output.txt
```

### 2. Compute cpuBaseNs

Run `benchops` on the same machine (see `gnovm/cmd/benchops/`). Then compute the weighted average:

```
cpuBaseNs = sum(ns × count) / sum(gas × count)
```

where `ns` and `count` come from `results_stats.csv`, and `gas` comes from the `OpCPU*` constants in `machine.go`.

### 3. Generate table

```bash
python3 gen_alloc_table.py bench_output.txt --cpu-base-ns 5.2
```

This prints:
- The power-law fit parameters (C, α)
- A Go `allocGasTable` declaration (in milligas — divide by 1000 for gas units)
- An accuracy report (actual vs model for every benchmark size)
- A plot (`alloc_gas_model.png`)

### 4. Update alloc.go

Convert the printed milligas table to gas (divide by 1000) and replace the `allocGasTable` in `gnovm/pkg/gnolang/alloc.go`.

## Files

| File | Description |
|------|-------------|
| `alloc_bench_test.go` | Go benchmarks: `make([]byte, N)` for N = 1B to 1GB |
| `gen_alloc_table.py` | Parses benchmark output, fits model, generates table + plot |
| `bench_output_do_dedicated.txt` | Reference benchmark data (DO Dedicated, amd64, Go 1.24) |
| `bench_output_do_amd64.txt` | Older data from DO Regular (shared vCPU), kept for comparison |
