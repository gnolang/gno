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
  costs `OpCPUSlopePreprocessClosureCapture`, 2500 CPU gas. A capture is
  added once per name per closure crossed, so distinct names under deep
  closures did that work for free: 2000 names under 100 closures took about
  eight times as long to preprocess as one name read 2000 times, for 0.4%
  more gas. `addHeapCapture` also reads the captured slot on the declaring
  block instead of climbing from the closure back to it, which removes an
  O(D) walk from every capture and no longer stops at a closure's own later
  declaration of the same name.
- `addHeapCapture` also reports the block slots its define copied, at
  `OpCPUSlopePreprocessCaptureGrowCopy`, 25 CPU gas each. `growBlockValues`
  doubles a block up to 512 slots and then grows it 256 at a time, copying
  every earlier slot each time, so a capture cost more the more captures its
  closure already held: at 60000 names under 8 closures, about twice the
  flat charge.
- The `IsAssignable` check on a plain assignment asks the block the name's
  path resolves to, not the first block holding that name. A local `f`
  declared later in the block let `f = ...` overwrite the package func `f`.
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

`BenchmarkGrowBlockValues` times `growBlockValues` reallocating a full block
of 512 to 65536 slots. `BenchmarkClosureCapture` preprocesses a file reading
`names` distinct parameters under `depth` closures, and the same file reading
the first parameter `names` times. The two have the same bytes and the same
scans, so the time difference over `(names−1)·depth` is the cost of one
capture, garbage collection included. Allocation gas covers none of it. Both
ran on a Ryzen 7 7840U, unpinned, `-count=10`, and use that box's largest
per-element machine factor, 1.52.

| Case | 95% upper | Reference | Copies per capture | Fit |
| --- | ---: | ---: | ---: | ---: |
| grow, 512 slots | 16.3 ns/slot | 24.8 | | 25 |
| capture, 32 names, depth 8 | 902 ns | 1370 | 1.0 | 1345 |
| capture, 2000 names, depth 100 | 1592 ns | 2419 | 3.7 | 2327 |
| capture, 60000 names, depth 8 | 3565 ns | 5418 | 117.3 | 2486 |

A capture's fit is its reference upper bound less its copies at 25 each; the
largest, 2486, rounds up to 2500. The raw output is
`gnovm/cmd/calibrate/closurecapture_bench_7840u_amd64.txt`; a
reference-hardware run should confirm both values.

These runs were unpinned, while the 1.52 factor comes from literal families
timed pinned to one core. Pinning slows the capture benchmark more than the
literal families, because the collector loses its idle cores. A review run
pinned to one core, with that session's own pinned factor, fit 2243 at 2000
names under 100 closures, 2491 on the mean (2712 at most) at 10000 names
under 120 closures, and 3562 at 60000 names under 8 closures. The 60000-name
source is 1,380,159 bytes, past `MaxTxBytes`; 43000 names is the most one
transaction carries at depth 8. The reference run should settle which mode
the constant follows.

## Alternatives considered

- Rewrite or cache the capture search. Rejected because the charge only needs
  to price the existing search. Capture selection, `depth`, `ValuePath`
  updates, `addAttrHeapUse` and the `dbn = flx` sequence stay as they were; an
  earlier rewrite also changed the behavior covered by `heap_alloc_gotoloop9`.
- Change `growBlockValues` to keep doubling past its threshold, so captures
  stop copying their closure's whole block every 256 defines. Rejected
  because `cap(Block.Values)` is what runtime allocation gas charges for
  every block, so the growth policy is part of the gas schedule well beyond
  preprocess. Charging the slots copied prices the same work without moving
  that schedule.
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
- Each capture added costs 2500 CPU gas plus 25 per block slot it copies,
  so a package pays for every name each closure captures. 2000 distinct
  names under 100 closures went from 54.0M to 572.6M CPU gas, against 54.2M
  for one name read 2000 times.
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
