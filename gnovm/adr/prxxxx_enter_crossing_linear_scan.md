# Linear enter-crossing scan with a matching linear charge

## Context

`doOpEnterCrossing` located the first crossing ancestor by repeatedly calling
`PeekCallFrame(i)`. Each lookup restarted at the top of the frame stack, making
frame visits quadratic in call depth: the walk performed `D*(D+1)/2` visits to
reach depth `D`.

The gas schedule matched that shape. `OpCPUSlopeEnterCrossingQuad = 6` charged
`depth^2 * 6 / 10`, i.e. `0.6*D^2` gas for the `D^2/2` visits actually
performed — about 1.2 gas per frame visit, which is what the reference
hardware measured (0.60 ns/visit on M2, scaled by the documented 1.8-1.95x
M-to-Xeon factor).

## Decision

Keep `fr1 = PeekCallFrame(1)` and walk `m.Frames` once from top to bottom with
a cursor, skipping `!fr.IsCall()`. Count the first call frame as step 1 and
inspect both `WithCross` and `DidCrossing`, including on that first frame.
Preserve package validation, the test-file and ephemeral exceptions, the realm
mismatch panic, and `fr1.SetDidCrossing()` on completion. An absent first call
frame still panics.

Because each frame is now visited at most once, replace the quadratic schedule
with a linear one: `OpCPUSlopeEnterCrossing = 3` charged as `depth * slope`,
applied as a single charge at the two accept exits — the same places the old
lump sum was applied. The faux deployer frame costs one extra step, which is
what the walk genuinely spends discovering it.

## Calibration

`BenchmarkOpEnterCrossing_1/10/100/1000` in `bench_ops_test.go` is the
calibration source. It was deleted in #5415 while `gen_analysis.py` and
`plot_fits.py` still referenced it, so the constant could not be refit by the
documented tooling; it is restored here.

Keeping the benchmark name also keeps the before/after comparison checkable
without retaining a copy of the old handler: `ns/op(pure)` for both walks,
one host, one harness, alloc gas 0 throughout.

| depth | before (quadratic) | after (single pass) |
| ---: | ---: | ---: |
| 1 | 634 | 585 |
| 10 | — | 578 |
| 100 | 3463 | 738 |
| 1000 | 267691 | 2030 |

Net of the ~600 ns fixed intercept, a 10x depth increase (100 -> 1000) costs
94x more before and 9.4x more after — the quadratic and linear shapes
respectively. At depth 1000 the walk is 132x cheaper end to end. Depth 1 is
unchanged within noise, which matters because it is the common case.

The "after" column is a clean linear fit of 1.45 ns/frame over that
intercept, consistent at all four depths.

These numbers come from a development host, not the Xeon reference, so the
basis was fixed by measuring the superseded quadratic walk on that same host
with that same benchmark: 0.53-0.56 ns/visit against the 0.60 ns/visit in
`cmd/calibrate/bench_output_m2_arm64.txt`. Same benchmark and same workload
makes that the machine-portable comparison, and it places the host on the M2
basis. Applying the documented 1.8-1.95x M-to-Xeon factor to 1.45 ns/frame
gives 2.6-2.8 ns/frame on the Xeon 8168 reference, rounded up to 3 so the
charge does not undercharge the walk.

Per visit this is dearer than the old schedule's 1.2 gas/visit, which is
expected rather than a discrepancy: the quadratic walk re-read the same few
frames repeatedly out of L1, whereas a single pass streams a distinct
248-byte `Frame` on every step. The total still falls sharply — `3*D` instead
of `0.6*D^2`, so 300 gas instead of 6,000 at depth 100, and 3,000 instead of
600,000 at depth 1000.

Shallow depths move the other way, because `3*D` exceeds `0.6*D^2` while
`D < 5`. The two schedules cross at depth 5, where both charge 15 gas; from
depth 6 up the linear one is cheaper and the gap then widens quadratically.

| depth | 1 | 2 | 3 | 4 | 5 | 6 | 100 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| before | 0 | 2 | 5 | 9 | 15 | 21 | 6000 |
| after | 3 | 6 | 9 | 12 | 15 | 18 | 300 |

Depth 1 is the sharpest of those: `Q(1)` was `1*1*6/10`, which integer
division floored to zero, so the walk a plain `MsgCall` into a crossing
function performs was previously free and now costs 3. Since such a call
accepts at step 1, that is why every pinned `GAS USED` golden covering one
rises by exactly 3: `gc.txtar`, `compute_map_key_restore_gas.txtar`,
`gnokey_gasfee.txtar`, `stdlib_restart_compare.txtar` and
`stdlib_ibc_crypto_determinism.txtar`. The `addpkg`-only pins
(`addpkg_import_testdep_gas.txtar`, `restart_gas.txtar`) are unchanged,
which is the expected control.

## Alternatives considered

- Keep the quadratic charge on the linear walk: prices work the handler no
  longer performs, growing to a ~200x overcharge at depth 1000. The PERF note
  removed from `op_call.go` prescribed the linear charge instead, and
  `bench_ops_test.go` carried a matching TODO to convert the fit.
- Charge incrementally per step (`Q(i)-Q(i-1)` through a running accumulator)
  so a deep walk can be interrupted mid-flight. Measured 2.3x slower than the
  lump sum (1.45 -> 3.31 ns/frame, from `D` extra `incrCPU`/`ConsumeGas`
  round trips), and it buys little: interruption mattered when the walk was
  quadratic, but building `D` call frames already costs ~348 gas each
  (`OpCPUExec` + `OpCPUPrecallFunc` + `OpCPUCall`) against ~3 gas to walk one,
  so an uninterrupted linear walk is bounded at roughly 1% of the gas the
  caller has already spent to create the frames. Incremental charging also
  changes failure accounting, which the lump sum leaves alone.

## Consequences

The ancestor search visits each physical frame at most once, with O(D) work in
the visited physical stack depth and O(1) extra space. The retained initial
lookup of `fr1` adds only a linear prefix scan. Non-call frames do not add gas.

Successful-execution gas changes: deep crossing chains get materially cheaper
(see the figures above), which is the point of the change. This is a
consensus-visible gas change, safe to make now only because the chain has not
launched; after launch it would need a coordinated deployment.

Failure paths keep their pre-existing accounting. The charge lands only at the
accept exits, so a walk that ends in the realm mismatch panic still costs no
gas, and out-of-gas can only occur at the single charge rather than partway
through the walk. `TestDoOpEnterCrossingRealmMismatchChargesNoGas` pins the
mismatch branch, which previously had no coverage anywhere in the tree.
