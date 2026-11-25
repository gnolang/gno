# Deferred call parameter and capture gas

## Context

`doOpCall` charges 53 CPU cycles per parameter and 34 per heap capture.
`doOpReturnCallDefers` copies the same captures and assigns the saved arguments
into the new function block, but previously charged neither slope. Its dispatcher
charges a flat 215 cycles, and `acquireBlock` charges 100 cycles separately.
A capture-heavy deferred closure therefore performed unmetered linear CPU work.

Registration is not identical to an ordinary call: `doOpDefer` already runs
`popCopyArgs`, including argument copies and variadic normalization. Execution
still assigns every normalized parameter slot. Both paths use the resolved
function's unbound `FuncType.Params`, including the receiver, and `fv.Captures`.
Lazy bound methods resolve before these counts are read. Empty defer lists and
nil callables do not reach block creation or the new slope charge. Lazy method
resolution retains its separate per-hop charge.

## Measurement

Measured on Apple M5, darwin/arm64, Go 1.26.2. Each point is the median of three
runs, with 300ms adaptive benchmark iteration counts. Command, before and after:

```sh
go test ./gnovm/pkg/gnolang/ -run '^$' -bench '^BenchmarkOp(ReturnCallDefers|Add_Int)$' -benchtime=300ms -count=3
```

The benchmark source and inputs were identical before and after the production
change: exactly one deferred call per iteration; integer parameter values;
heap-item captures containing integers; P/C in 0, 1, 10, 100, 1000 along each axis.
Go's adaptive `b.N` may differ; the compared unit is one handler invocation with
the same P/C, not total benchmark wall time. Setup and cleanup are excluded by
`bm.SwitchOpCode`. `benchMachine` leaves the CPU GasMeter nil, but records
allocation gas separately. The before/after timing comparison therefore does
not measure gas-meter overhead or establish a speed improvement.

Cold measurements include block allocation; warm measurements recycle a block
with capacity P+C outside the timed region, isolating the copying/binding work
from allocation. The tiny reported warm capture allocation gas is one-time
fixture creation amortized over b.N, not per-invocation allocation. The former
benchmark varied pending defer count with P=C=0 while executing only one defer;
it could not fit either copying slope.

Median handler time in ns (`ns/op(pure)`):

| P | C | Cold before | Cold after | Warm before | Warm after |
|---:|---:|---:|---:|---:|---:|
| 0 | 0 | 212.5 | 209.4 | 112.4 | 112.3 |
| 1 | 0 | 218.1 | 212.8 | 115.0 | 114.1 |
| 10 | 0 | 218.5 | 213.4 | 128.9 | 128.2 |
| 100 | 0 | 590.3 | 578.4 | 285.1 | 298.8 |
| 1000 | 0 | 4187.0 | 4225.0 | 1924.0 | 2148.0 |
| 0 | 1 | 256.3 | 217.2 | 120.2 | 119.1 |
| 0 | 10 | 240.9 | 239.5 | 160.2 | 160.1 |
| 0 | 100 | 846.6 | 843.1 | 574.4 | 610.3 |
| 0 | 1000 | 6548.0 | 6525.0 | 4738.0 | 4828.0 |

OLS fits across all five points on each axis, with an intercept. Reference
normalization uses the existing calibration convention `81 / Add_Int(ns)`;
it is an estimate, not a substitute for measuring on the reference Xeon 8168.
The cold fits below are total time slopes, **not** allocation-subtracted CPU
prices; use warm fits to assess CPU work without double-counting allocation.

| Measurement | Add_Int ns | Param ns/slot | Capture ns/slot | Normalized param | Normalized capture |
|---|---:|---:|---:|---:|---:|
| Before, cold | 42.38 | 3.99 | 6.33 | 7.62 | 12.11 |
| Before, warm | 42.38 | 1.81 | 4.62 | 3.47 | 8.84 |
| After, cold | 39.79 | 4.03 | 6.32 | 8.21 | 12.87 |
| After, warm | 39.79 | 2.04 | 4.71 | 4.15 | 9.59 |

These results do **not** reproduce 53/34. In particular, deferred parameter
binding omits `popCopyArgs`, so its CPU slope is structurally lower than an
ordinary call. For context, same-machine ordinary-call cold measurements
(200ms, three runs) gave median P=0/1000 times of 134.4/9319 ns and C=1/1000
of 138.2/6564 ns; the corresponding deferred cold times above are consistent
with cheaper parameter binding and comparable capture copying. Different
micro-operations need not scale uniformly between Apple M5 and Xeon.

## Decision

Charge `OpCPUSlopeCallParam * len(ft.Params) +
OpCPUSlopeCallCapture * len(fv.Captures)` immediately after resolving `ft` and
before `PushFrameCall`. Reuse 53/34 as a conservative shared call schedule,
not as a claim that defer-specific measured slopes equal those constants.
This bounds the previously omitted linear work before allocating the callee
frame or copying captures, including during panic unwinding.

`doOpCall` already charged the same slopes, but only after `acquireBlock` and
`PushBlock`, so an out-of-gas there still paid for the block first. The charge
moves ahead of `acquireBlock`, giving both paths the same order: charge, then
allocate and bind. The amount charged by `doOpCall` is unchanged.

The handler adds `100 + 53*P + 34*C` CPU cycles for a plain deferred function;
the dispatcher base, lazy resolution, allocation and body execution remain
separate. Compared with the old schedule, the added CPU gas for a successfully
entered deferred call is `(53*P + 34*C) * GasFactorCPU`.

## Alternatives considered

- Charge only captures: leaves linear parameter assignment unaccounted for.
- Add defer-specific slopes from this M5 run: more precise locally, but changes
  the schedule using non-reference hardware and adds constants before a full
  reference calibration. Defer this optimization to a Xeon calibration.
- Move the charge to registration: charges calls never executed and cannot use
  the eventual function resolved by lazy method dispatch.
- Share one call-setup helper between `doOpCall` and `doOpReturnCallDefers`:
  a larger refactor than the accounting fix needs. The two paths only share the
  charge's position (before `acquireBlock`), which is aligned directly instead.

## Consequences

This is a consensus-visible gas schedule change. Old and new binaries consume
different gas for identical transactions; transactions near their gas limit may
now fail. Validators must activate the change at a coordinated consensus upgrade
height, not through an uncoordinated rolling deployment. This patch does not
select that height or add activation gating.

53/34 conservatively overprices the measured defer CPU work, especially parameters.
A follow-up on reference Xeon hardware can decide whether separate deferred-call
prices are warranted. These microbenchmarks establish linear accounting for the
fixture shapes, not a universal wall-time bound for every possible argument type.

## Validation

Two filetests fix total gas through Gno source. `gas/defer_capture_gas.gno`
runs 10 deferred calls to a closure taking two parameters and capturing eight
integer variables (golden 68681). `gas/defer_method_gas.gno` runs 10 deferred
calls to a value method with two parameters, so the receiver counts as a third
parameter (golden 42420).

`TestDeferGas` keeps the three guarantees a total-gas golden cannot express:
exhaustion before `PushFrameCall` and capture copying (unchanged frame/block
counts), the same ordering during `doOpPanic2` unwinding, and `OutOfGasError`
escaping the Gno exception recovery in `Run`. Its four cases are mixed P=2/C=4
and capture-only P=0/C=4, each on normal return and panic unwind.
`TestCallGas` checks that `doOpCall` runs out of gas before `acquireBlock`.

- Reverse check, removing only the charge in `doOpReturnCallDefers`:
  `defer_capture_gas` consumes 64901 instead of 68681, a difference of
  `3780 = 10 calls * (2*53 + 8*34)`; `defer_method_gas` consumes 40830 instead
  of 42420, `1590 = 10 calls * 3*53`. Both goldens fail, and all four
  `TestDeferGas` cases fail their pre-frame assertion. With the previous
  `op_call.go`, both `TestCallGas` cases fail as well.
- Full `go test ./gnovm/pkg/gnolang/ -run Files -test.short` passes apart from
  the type-check diagnostic goldens that differ under Go 1.26; use the Go
  version declared in `go.mod`.
- SDK `-run Gas`, integration `-run TestTestdata` and the examples suite pass. No
  integration gas goldens changed, including `gnokey_gasfee.txtar` (its hello
  function has no Gno defer).
