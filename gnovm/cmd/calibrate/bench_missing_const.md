# Calibrating the three missing gas constants (post-#5291 follow-up)

This runbook fits three constants that landed in #5291/#5127/#5154 without
benchmark support. Run these commands on the same reference hardware used for
`bench_output_do_dedicated.txt` (Intel Xeon 8168) so the new slopes share the
`cpuBaseNs = 5.2` baseline.

## Targets

| # | Constant | Source PR | Formula | Current value |
|---|----------|-----------|---------|---------------|
| 1 | `OpCPUSlopeCallReturn` | #5291 gap (ltzMaxwell review on `op_call.go:178`) | `base + slopeP·P + slopeC·C + slopeR·R` | *does not exist yet* |
| 2 | `OpCPUComputeMapKey` (+ per-byte/per-elem slopes) | #5127 | `base + slopeByte·N + slopeElem·M` | flat `= 10` |
| 3 | `OpCPUCmpPerByte` | #5154 | `base + slope·min(len1,len2)` | `= 1` (guess) |

## Step 0 — sanity: confirm the new benchmarks build

```
go test -count=1 -run=^$ -timeout=60s \
  -bench='BenchmarkOpCall_0Params_0Captures_1Results$|BenchmarkOpEql_StringDistinct_8$|BenchmarkOpLss_String_8$|BenchmarkComputeMapKey_NestedArray' \
  -benchtime=1x ./gnovm/pkg/gnolang/
```

All should return in <1s. Dev-laptop numbers are noisy; only the reference box
numbers are used for fitting.

## Step 1 — run benchmarks on reference hardware

On the Xeon 8168 box:

```
cd gnovm
go test -bench='BenchmarkOpCall_' -benchmem -count=5 -timeout=30m \
  -run=^$ ./pkg/gnolang/ > cmd/calibrate/bench_output_call_return.txt

go test -bench='BenchmarkOpEql_StringDistinct_|BenchmarkOpEql_StringDiffFirst_|BenchmarkOpLss_String_' \
  -benchmem -count=5 -timeout=30m -run=^$ ./pkg/gnolang/ \
  > cmd/calibrate/bench_output_cmp_per_byte.txt

go test -bench='BenchmarkComputeMapKey_' -benchmem -count=5 -timeout=30m \
  -run=^$ ./pkg/gnolang/ > cmd/calibrate/bench_output_compute_map_key.txt
```

`-count=5` gives benchstat confidence bands. `-timeout=30m` safety for 16M-byte
and 1M-recursion cases.

## Step 2 — run analysis

`gen_analysis.py` already has entries for all new benchmarks (see `SLOPE_FITS`).
Concatenate the outputs and run it:

```
cat cmd/calibrate/bench_output_call_return.txt \
    cmd/calibrate/bench_output_cmp_per_byte.txt \
    cmd/calibrate/bench_output_compute_map_key.txt \
    cmd/calibrate/bench_output_do_dedicated.txt \
  > cmd/calibrate/bench_output_combined.txt

python3 cmd/calibrate/gen_analysis.py \
  cmd/calibrate/bench_output_combined.txt \
  > cmd/calibrate/op_bench_analysis_new.txt
```

Read the output — each new benchmark set gets a linear fit with R², plus a
suggested integer slope. Cross-check R² > 0.98 before trusting the value.

## Step 3 — decide the constants

### `OpCPUSlopeCallReturn`

Read slope from the *Call (0 params, 0 captures, results)* fit.

Sanity check: the per-result slope should be close to but below
`OpCPUSlopeCallParam = 53`. A result slot is a zero-initialized `TypedValue`
append to the block; a param slot is the same plus an arg-copy from the stack.
Expect ~30–45 gas/result.

### `OpCPUComputeMapKey` + per-N slopes

Three fits produce three coefficients:

- **Base** (`OpCPUComputeMapKey`): intercept of the *ComputeMapKey (bytes)* or
  *(int array)* fit — they should converge on the same base cost. Typically
  30–80 gas.
- **Per-byte slope**: slope from *(string)* and *(bytes)* fits — should match
  each other to within 10% (same inner code path in `values.go`). Expect
  ~0.1–0.3 gas/byte; use the `slope/N` divisor pattern (e.g. `/10` or `/100`)
  to preserve resolution since integer slopes round aggressively.
- **Per-element slope**: slope from *(int array)* and *(struct)* fits. Should
  also match each other (same inner recursion). Expect ~20–30 gas/element.

The *nested array* fit should match the per-element slope multiplied by 1
(one inner element per level). If it's meaningfully higher, the base cost
is being undercharged — bump the base.

### `OpCPUCmpPerByte`

Read from the *Eql (string, distinct)* and *Lss (string)* fits. Both should
give the same slope (same memcmp inner loop). On Xeon 8168 expect ~0.3 gas/byte.
Use the `slope / N` divisor pattern:

```go
// gas = len * OpCPUSlopeCmpByte / 10  (fit: ~0.3 ns/byte → slope=3, divisor=10)
m.incrCPU(int64(len(ls)) * OpCPUSlopeCmpByte / 10)
```

Then rename the constant to `OpCPUSlopeCmpByte` and update the `overflow.Mulp`
call sites in `op_binary.go` accordingly.

The *Eql (string, DiffFirst)* benchmark is a **validation** case, not a fit
input — its cost should be flat and roughly equal to the intercept of the
*distinct* fit (one byte examined, length mismatch or first-byte mismatch
terminates).

## Step 4 — update code + golden values

1. Apply the new constants in `machine.go`.
2. Charge sites:
   - `op_call.go:178`: add `+ OpCPUSlopeCallReturn*int64(len(ft.Results))`.
   - `values.go` `ComputeMapKey`: replace flat `store.ConsumeGas(...)` with
     base + per-byte/per-element charges at the actual append/recurse sites.
     (Separate design question: migrate to `*Machine` / `m.incrCPU` — see plan
     `humming-floating-moore.md`. Not required for the benchmark PR.)
   - `op_binary.go`: replace `OpCPUCmpPerByte` with `OpCPUSlopeCmpByte/N`
     divisor pattern.
3. Regenerate golden gas files:
   ```
   go test -run='TestFiles/gas' -update ./gnovm/pkg/gnolang/
   ```
4. Run full suite to confirm no gas-metered test regressed:
   ```
   go test -run=TestFiles -test.short ./gnovm/pkg/gnolang/
   go test -run=TestTestdata ./gno.land/pkg/integration/
   ```
5. Update `op_gas_formulas.md`:
   - Call row: add `R = results` to the list of parameters.
   - New row for `ComputeMapKey` (under a new "Map key" section).
   - Comparison row: string cmp has `base + slope·min(N1,N2)`.

## Step 5 — adversarial validation

For each constant, write a regression `.gno` file under `gnovm/tests/files/gas/`
that exercises the worst case:

- `call_many_returns.gno` — function returning 100+ results, confirm charged
  gas grows linearly.
- `compute_map_key_deep_nest.gno` — map key is `[1][1]...[1]int` nested 32
  levels, confirm gas grows linearly with depth.
- `string_eql_1mb.gno` — compare two 1 MB distinct strings, confirm charged
  gas ≈ `1M × slope / divisor + base`.

Include them in the PR to lock the fitted behavior against future regressions.

## Notes

- All three constants are consensus-breaking. Land behind the same upgrade
  gate as #5291 / #5127 / #5154.
- If per-byte and per-element slopes fit cleanly, consider whether
  `OpCPUComputeMapKey` should use `/100` or `/1024` divisor, same as BigInt
  slopes — sub-integer resolution matters when typical keys are 8–64 bytes.
- Benchmark runs on laptops give wildly different absolute numbers but
  *ratios* between lengths are stable — that's enough to sanity-check the
  slope even before the reference-hardware run.
