# Meter preprocess closure-stack scans

## Status

Implemented.

## Context

`codaHeapDefinesByUse` determines whether a name use crosses one or more
function-literal boundaries. For each relevant `NameExpr`, it calls
`findFirstClosure`, walks the block stack back to the defining block, records
the nearest capture, sets `dbn = flx`, and repeats for the remaining closures.

With `N` references inside `D` nested function literals, the repeated walks
scan a triangular number of frames per reference. The work is therefore
O(N×D²), while the existing source-size and allocation charges do not account
for the scanned frames. A not-found lookup also walks frames and had the same
unmetered path.

## Decision

Keep the capture algorithm and charge its existing work:

- `findFirstClosure` returns a fourth value, `scanned`, incremented once at the
  start of every loop iteration. This counts closure, non-closure, faux, stop,
  early-return, and final not-found frames uniformly.
- `codaHeapDefinesByUse` obtains the transaction-scoped meter once with
  `preprocessGasMeterOf(store)` and charges
  `scanned × OpCPUSlopePreprocessClosureScan` immediately after every lookup,
  before checking `found`.
- `OpCPUSlopePreprocessClosureScan` is 5 CPU gas per scanned frame.
- `addHeapCapture` reports whether it added a capture, and each one it adds
  costs `OpCPUSlopePreprocessClosureCapture`, 2100 CPU gas. A capture is
  added once per name per closure crossed, so distinct names under deep
  closures did that work for free: 2000 names under 100 closures took about
  eight times as long to preprocess as one name read 2000 times, for 0.4%
  more gas. `addHeapCapture` also reads the captured slot on the declaring
  block instead of climbing from the closure back to it, which removes an
  O(D) walk from every capture and no longer stops at a closure's own later
  declaration of the same name.
- The string branches of `append` and `copy` preprocess their rewritten
  argument with `preprocess1(store, ...)`. Before, they called
  `Preprocess(nil, ...)`, which had no meter, and its coda ran with a stack
  that started at the current block. So closures in that argument escaped
  the charge and could lose their capture. Now the enclosing pass scans them
  with the full stack.
- `preprocess1` and `assertValidAssignLhs` check whether a name is constant
  on the block its path resolves to (`GetIsConstAt`), not on the first block
  holding that name (`GetIsConst`). A name reserved in the current block, by
  a later `const` or `type` or by the left side of the same `:=`, used to
  hide the outer declaration. `limit := limit * 2` over a package-level
  `const limit` panicked at run time, a read of an outer variable before a
  later `type x` read its zero value, and `x = 2` before a later `const x`
  was refused with `cannot assign to const x`. All three now behave as in
  Go.

The benchmark covers depths 1, 8, 32, and 128; closure-only and mixed
closure/non-closure/faux stacks; found and final not-found paths; and the
cumulative `dbn = flx` repetition. It runs without allocations. Calibration
used `-benchtime=200ms -count=10` on Apple M5 and the repository's checked-in
2.1 conversion factor to the reference Xeon, where 1 gas is 1 ns. The raw
output, together with the seven flat ops that re-derive the factor, is
checked in as `gnovm/cmd/calibrate/closurescan_bench_m5_arm64.txt`.

| Representative final case | Local mean ns/frame | 95% upper | Reference upper | Rounded |
| --- | ---: | ---: | ---: | ---: |
| depth 1, non-closure/faux, not found | 1.9999 | 2.0418 | 4.2895 | 5 |
| depth 1, closure-only, cumulative | 1.7603 | 1.7980 | 3.7773 | 4 |
| depth 128, mixed, found | 1.1919 | 1.2650 | 2.6576 | 3 |
| depth 128, non-closure/faux, not found | 1.2286 | 1.3860 | 2.9118 | 3 |

The provisional slope 4 covered the deep cases but fell below the shallow
not-found confidence bound after conversion, so the final value is 5.

`BenchmarkClosureCapture` preprocesses a file reading `names` distinct
parameters under `depth` closures, and the same file reading the first
parameter `names` times. The two have the same bytes and the same scans, so
the time difference over `(names−1)·depth` is the cost of one capture,
garbage collection included. Allocation gas covers none of it. On a Ryzen 7
7840U, unpinned, `-count=10`, the 95% upper bound is 1338 ns at 2000 names
and depth 100, against about 850 ns at the smaller sizes. That box's largest
per-element machine factor, 1.52, gives 2034 reference-ns, rounded up to
2100 because a scalar factor ports allocation-heavy work poorly. A run
pinned to one core, with GC on the measured core, gave 1746 ns. The raw
output is `gnovm/cmd/calibrate/closurecapture_bench_7840u_amd64.txt`; a
reference-hardware run should confirm the value.

## Alternatives considered

- Rewrite or cache the capture search. Rejected because this change is resource
  bounding, not a semantic rewrite. Capture selection, `depth`, `ValuePath`
  updates, `addAttrHeapUse`, `addHeapCapture`, and the `dbn = flx` sequence must
  remain unchanged; an earlier rewrite also changed the behavior covered by
  `heap_alloc_gotoloop9`.
- Charge only successful capture searches. Rejected because not-found searches
  still scan frames; omitting them would retain an unmetered path.
- Add a source-size flat charge. Rejected because equal-size sources can cause
  different triangular scan counts.
- Fetch or store the meter globally. Rejected because the store already owns
  the transaction-scoped preprocess allocator and its meter. A process-global
  meter would not preserve transaction isolation.

## Consequences

- Metered preprocessing now charges exactly 5 CPU gas per frame actually
  visited. For `N` references at closure depth `D`, the added charge follows
  the existing O(N×D²) work instead of changing that work.
- Each capture added costs 2100 CPU gas, so a package pays for every name
  each closure captures. 2000 distinct names under 100 closures went from
  0.18 to 1.31 gas per ns of preprocessing on the calibration box, against
  1.0 for one name read 2000 times.
- If the store, preprocess allocator, or allocator meter is nil,
  `preprocessGasMeterOf` returns nil and `chargeCPUGas` remains a no-op. Normal
  off-chain tests, filetests, and tooling without a preprocess meter retain
  their prior behavior.
- The gas schedule is consensus-visible for deployments and any metered query
  that preprocesses code. Transactions near their gas limit can now stop
  earlier. Replay requires validators to use the same schedule at the same
  activation point; this change does not add activation or schedule versioning.
- The constness fix changes outcomes as well as gas. A transaction that ran
  one of the three shapes above, or deployed a package refused for the third,
  now succeeds where it failed, or reads the outer value where it read zero.
  A shadowed package type, as in `node := &node{...}`, takes the same gas
  and result as before.
- The `gnokey_gasfee` integration fixture increases by 15 gas. Its `Hello`
  function reads package variable `s`, causing one final not-found scan across
  `FuncDecl`, `FileNode`, and `PackageNode`: 3 frames × 5 gas.
