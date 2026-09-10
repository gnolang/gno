# Byte slice to string CPU gas

## Context

`Machine.doOpConvert` charged only `OpCPUConvertStrBytes` (363 CPU gas) for
`[]byte→string`, regardless of length, while `ConvertTo` does O(N) work. The
`[]int32→string` path already charged a per element slope of 8, so only the
byte path underbilled.

The cost depended heavily on the slice's backing. A Data backed slice is one
memcpy. A List backed slice walks N `TypedValue` entries through
`copyListToData`, which measured ~10x slower per byte at N=1,000,000, and costs
40 bytes per element instead of 1. Byte slices were inconsistently backed:
`defaultArrayValue`, `make([]byte,n)`, `[]byte(string)` and both `append` paths
already produced `NewDataArray`, but `doOpSliceLit`, `doOpSliceLit2` and
`Go2GnoValue`'s slice arm produced `NewListArray` unconditionally. So
`[]byte{N:0}` built an N element List backed slice from a one line literal,
and that was the only cheap way to reach the expensive conversion path.

## Decision

Two changes.

**Make every byte slice Data backed.** `doOpSliceLit` and `doOpSliceLit2` now
dispatch `Uint8Kind` to `NewDataArray`, matching what `defaultArrayValue`
already did for arrays; `Go2GnoValue`'s `reflect.Slice` arm does the same,
matching its own `reflect.Array` arm directly above it. With `make`, `append`
and `[]byte(string)` already Data backed, no Gno source can now produce a List
backed byte slice. The `copyListToData` branches in `ConvertTo` and `append`
remain as defensive paths but are unreachable. This also removes a 40x
allocator amplification that existed independently of any conversion:
`[]byte{999999:0}` allocated 40,000,056 bytes where `make([]byte,1000000)`
allocated 1,000,056.

**Charge `OpCPUSlopeConvertBytesStr = 4` per byte** in `doOpConvert`, after the
existing conversion checks and before `ConvertTo`, so an oversized conversion
runs out of gas before doing the work. Match on `baseOf(xv.T).(*SliceType)`
with `Elt.Kind() == Uint8Kind`, covering named byte elements. Charge slice
length, not backing length or capacity. Nil and empty slices pay no slope. The
flat cost and the rune slope are unchanged.

## Coefficient basis

4 is `OpCPUSlopeCopyPrimitive`, which the VM already charges per byte for
exactly this work. Its comment names the `copyDataToList`/`copyListToData`
helpers and raw memcpy, and every `copyListToData` call site in `uverse.go`
charges it. The closest precedent is `append`: both of its byte paths allocate
a fresh `NewDataArray` and then memcpy into it — the same allocate-then-copy
`[]byte→string` performs — and they charge 4 per byte *on top of* the
allocation gas the new array costs. `string(b)` and `append(nil, b...)` now
price identically, which is the point.

It is deliberately above measurement, not fitted to it. On the reference basis
(`ns/op(pure)` minus alloc gas, 1 gas = 1 ns), `BenchmarkOpConvert_BytesToString`
over the standard 1/10/100/1000 series (`-benchtime=2s -count=5`, median, AMD
Ryzen 7 7840U) fits a base of ~161 ns and a per-byte slope of ~0:

| N | ns/op(pure) | alloc gas | net |
| --- | --- | --- | --- |
| 1 | 186.3 | 31.0 | 155.3 |
| 10 | 200.0 | 34.0 | 166.0 |
| 100 | 216.0 | 56.0 | 160.0 |
| 1000 | 394.4 | 246.0 | 148.4 |

That is, allocation gas alone already over-covers a Data backed conversion:
`alloc.NewString` charges ~0.25 gas/byte at N=1000 and ~0.18 at N=1,000,000,
against a memcpy measured at ~0.06 ns/byte. A slope of 0 would be defensible on
these numbers. 4 buys consistency with `copy`/`append` and headroom against
slower reference hardware, at the cost of overcharging a memcpy — the same
overcharge those two builtins already carry.

For reference, the List backed path that this change makes unreachable fitted
~2.4 ns/byte at N=1,000,000 on the same box. Anchoring by ratio to the rune
slope of 8 put its Xeon equivalent near 2.9. Note that such a ratio is not a
portable calibration: the List walk is memory bandwidth bound while
`[]rune→string` is CPU bound on utf8 encoding, so the ratio between them varies
by roughly 2x across hardware. That is why this constant is anchored to an
existing constant for the same operation rather than to a ratio.

## Alternatives considered

* Keep a flat charge. It cannot bound CPU work that grows with N.
* Reuse rune slope 8. Twice the cost of the memcpy the VM already prices at 4.
* Price Data and List separately. Now moot: only Data is reachable. It would
  also make gas depend on backing representation, which `copy` and `append`
  deliberately avoid by charging one slope for both.
* Fit the slope to the Data path alone (0, per the table above). Cheapest for
  realms, and allocation gas would still cover the copy — but it would price
  `string(b)` below `append(nil, b...)` for the identical allocate-and-memcpy.
* Meter inside `ConvertTo`. `ConvertTo` does take an `*Allocator`, which carries
  the same gas meter in production, so this is feasible — but three callers
  (`ConvertUntypedTo`, `convertConst`, `ConvertGetInt`) pass nil and are
  compile time or O(1), and handler metering keeps CPU accounting in one place.
* Leave the backings inconsistent and price the worst case. That keeps the 40x
  allocator amplification, and forces the slope up to cover a List walk that
  only a literal could produce.

## Consequences

Consensus visible, on both counts. A nonempty conversion costs an additional
`4*N` CPU gas (`GasFactorCPU = 1`) — 8x the previous flat charge at N=1000 — and
byte slice literals change allocator cost, persisted representation (`Data`
instead of `List`) and object hashes. A
transaction close to its gas limit can fail earlier. Nodes must use the same
schedule at the same height. This patch does not implement activation or
schedule versioning; it is safe to land unversioned only because the chain has
not launched and there is no state to replay.

`GasUsed` itself is not an app hash input and is absent from `ABCIResult`, whose
`Error`, `Data`, and `Events` feed `LastResultsHash`. Changed execution outcomes
can still change those hashes, so omitting the gas counter is not replay safety.

Two filetests change behaviour rather than just gas:

* `a49.gno` — `println(&c[0])` on a byte slice literal now prints
  `&(65 <databyte> uint8)`, because a pointer into a Data backed array is a
  `DataByteType` pointer. User visible program output.
* `zrealm_listbyte_append0.gno` — the GVM-01 regression it guards now runs
  through the Data dst branch of `append`. The List dst branch it originally
  covered is no longer reachable from Gno source.

## Unresolved scope

* The temporary `make([]byte, N)` in `ConvertTo`'s List branch is not bounded by
  `maxAllocTx`. That branch is now unreachable from Gno source, so the exposure
  is gone in practice, but the allocation is still untracked if it is ever
  reached again.
* Reverse conversion `string→[]byte` has no matching per byte slope. Its copy is
  covered by allocation gas at ~0.18 gas/byte, which is above a memcpy on the
  reference basis, but it is still priced inconsistently with `copy`'s 4.
