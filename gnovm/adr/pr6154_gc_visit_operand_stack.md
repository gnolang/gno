# Make GC see every root that retains memory

## Context

`Machine.GarbageCollect()` rebuilds `alloc.bytes` from scratch: `Alloc.Reset()`,
then `Recount` every object reachable from the machine's roots. Anything the
walk cannot reach is dropped from the tally, so a root the walk does not know
about is not a leak — it is a **cap bypass**. `Allocate` asks GC to free
memory, GC reports headroom that does not exist, and the retried allocation
succeeds. Repeat per element and the program grows without bound while
`bytes` stays flat. Past `maxAllocTx` (500 MB) / `maxAllocQuery` (1.5 GB) that
is a Go OOM — a fatal runtime error `recover()` cannot catch: validators
crash, and the chain halts. Both `MsgAddPackage` and the unauthenticated
`qeval` query path reach it.

Three such roots were missing.

**1. The operand stack (`m.Values`).** The roots were `m.Blocks`, `m.Frames`,
`m.Package`, the staging package and `m.Exception` — not `m.Values`.
Intermediate results, call arguments before `OpCall` pops them into a block
and composite-literal elements waiting for `doOp*Lit` all live there, and are
commonly large. Reproduced on the base commit with a 560 MB array literal in
`MsgAddPackage`: the package deployed and failed only at storage deposit, far
past the cap. GC gas was undercharged the same way — `visitCount` omitted
every object rooted there, so `gcVisitGas` billed less than the traversal
actually cost.

**2. Buffers under construction.** Ops allocate a destination buffer as a Go
local and fill it one entry at a time, allocating as they go: `doOpSliceLit`
(`NewListArray` + `PopCopyValues`), `doOpSliceLit2` (keyed copies, then
`defaultTypedValue` for every index left undefined), `doOpStructLit` (both
branches), `popCopyArgs` (the `args` list and the variadic `list`), and the
`default*` fills themselves (`defaultArrayValue`, `defaultStructFields`).
Nothing references the buffer until the op finishes, so a GC part-way through
recounts past every entry already written.

Walking `m.Values` does not cover this, and the aliased case shows why: in
`[]T{a, a, …}` every stack slot holds the *same* `*ArrayValue`, and `vis`
dedups `Object`s, so the walk charges one element however many fresh copies
`PopCopyValues` writes into the buffer. Measured under a 1 MB cap, all with
`Big = [300000]byte` and a single package-level `var a Big`:

| shape | built under a 1 MB cap |
| --- | --- |
| `[]Big{a × 100}` | 30 MB, scaling with the element count |
| `S{a × 10}` (ordered fields) | 3 MB |
| `S{F0: a, …}` (keyed fields) | 3 MB |
| `g(a × 10)` (fixed arity) | 3 MB |
| `f(a...)`, `f(xs ...Big)` | 3 MB |
| `[]Big{0: a, …, 9: a}` (indexed) | 3 MB |

**3. `Block.Blank`.** `_ = expr` assigns through `GetBlankRef`, so the block
pins `expr` for its whole life even though nothing can read it back — but
`(*Block).VisitAssociated` walked only `b.Values` and `b.Parent`. One blank
assignment per recursion level held **200 × 300 KB = 60 MB under a 1 MB cap**,
while the byte-identical program binding the same expression to a name was
correctly refused.

## Decision

**Walk `m.Values`.** For each slot recount the `TypedValue` wrapper
(`Recount(allocTypedValue)`, as `Frame.Visit` already does for the receiver
and defer args) and visit `.V` through the same visitor, so stack-rooted
objects are both counted in `visitCount` (GC gas for the traversal) and
recounted into `alloc.bytes` (cap enforced).

**Charge the linear walks per slot.** `visitCount` only grows for slots that
carry a `.V`, so a stack of primitives (every `.V` nil) is scanned for free:
the walks over `m.Values` and `anchors` are `O(len)` work that `gcVisitGas`
does not see, and `len(m.Values)` is bounded only by gas — `alloc_14` parks
~8192 slots from ~600 bytes of source at depth 128. Measured, the walk is
flat at 58 gas from 1K to 4M slots while wall time rises 4.8 µs → 7.8 ms.
`gcStackScanGas` therefore charges `gcStackScanSlopeGas` (4) per slot
scanned, priced separately from `gcVisitGasTable`: a slot scan is a
sequential pass over a contiguous `[]TypedValue`, ~1.3–3.3 ns/slot, not a
29–700 ns pointer-chasing object visit, so folding it into the visit table
would overcharge by two orders of magnitude. The slope is a rough
developer-machine figure (`BenchmarkGCStackScan`), not a reference-hardware
one; it sits just above the measured maximum and matches
`OpCPUSlopeCopyPrimitive`, the existing slope for a comparable linear pass.

**Add `Allocator.anchors`, a root set for buffers under construction.**
`PushAnchor(tvs)` marks a `[]TypedValue` as live before the first write;
`PopAnchor` releases it once the object owning it is reachable from another
root. `GarbageCollect` walks the anchors alongside the machine's roots. The
rule for callers is "no allocation between the pop and the moment the owner
is rooted", which is easy to satisfy because GC only ever runs from
`Allocator.Allocate` — a stretch of code that allocates nothing cannot GC.

An anchor is either a `[]TypedValue` buffer (`PushAnchor`, the common case)
or a whole object (`PushAnchorValue`), for a destination whose entries are
not a flat slice. `doOpMapLit` needs the second form: it fills a `MapValue`'s
linked list, and with the operands all aliasing one value (`{0: a, 1: a, …}`)
the popped sources are a single object while the copies grow N times — ten
300 KB entries built 3 MB under a 1 MB cap, 100 entries 30 MB. GC visits an
object anchor through the ordinary visitor, so everything already written
into it is counted. It is anchored *after* the operands are popped, on
purpose: pushing the map onto the operand stack instead would write over
`kvs[0]`, which `PopValues` leaves in place below the new top.

The anchor list lives on the `Allocator` rather than the `Machine` so
`defaultArrayValue` and `defaultStructFields`, which take only an
`*Allocator`, can protect their own fills for every caller at once.
`Machine.runOnce` truncates back to its entry depth whenever it catches a
panic, so an op that dies mid-fill cannot leave its buffer anchored — without
that, a recovered panic inside a loop would grow the list per iteration. Under
`-tags debugAssert` an op that returns normally with unbalanced anchors
panics; the whole filetest suite is clean.

**Visit `b.Blank`.** Chosen over removing the retention (nothing reads `_`
back, so the store exists only to give `Assign2` a `*TypedValue` to write
through) because `Assign2` calls `rlm.DidUpdate(m, pv.Base, oo1, oo2)` with
the block as base: today the block genuinely holds the value, so the realm
bookkeeping is truthful. Dropping the reference without a matching
un-reference would make it lie. Counting the retention is correct and is the
same mechanism as the other two roots; removing it is a separate change that
has to reason about `DidUpdate`, and about `Blank` being serialized
(`pb3_gen.go`) and type-checked for publicity (`realm.go`).

## Alternatives considered

- **Charge operand-stack slots at push/pop** so `bytes` is exact at all times
  and GC needs no stack walk. Touches the interpreter's hottest path and
  needs every pop/reslice site covered symmetrically; a missed site leaks
  `bytes` upward. The GC walk is a strictly smaller change and is the same
  mechanism the other roots use.
- **Skip the slot wrapper, visit `.V` only.** Symmetric with runtime (which
  never charges the wrapper) but inconsistent with `Frame.Visit` /
  `Exception.Visit`, which do recount it. Kept the existing convention here
  and left the reconciliation to a follow-up (see below).
- **Special-case strings at stack slots** (skip their bytes to avoid the
  alias over-count described below). Rejected: it reopens the under-count for
  strings reachable only from the stack — a `[]string` literal of large
  elements is exactly the OOM shape this change closes — and it makes the
  tally depend on where an alias happens to sit.
- **Copy elements in place instead of anchoring**, writing each copy back
  over its own operand-stack slot and popping only at the end. Needs no new
  state and unwinds for free, but it cannot cover the fills whose
  destination is larger than the source: `doOpSliceLit2`'s default fill is
  sized by the index range, not the element count, and `defaultArrayValue` /
  `defaultStructFields` have no operands at all. Anchors cover every site
  with one mechanism.

## Consequences

- Every root that pins memory is now counted, so the cap is enforced against
  composite literals, call arguments and blank assignments. All six shapes in
  the table above are refused with `allocation limit exceeded`, and so is the
  560 MB `MsgAddPackage` literal, before the storage-deposit step.
- Consensus-visible on three axes, all measured against `develop`:
  - GC visit gas from the extra visits the new roots contribute:
    `gas/nested_alloc` 8559690224 → 8559693211 (+2987).
  - The per-slot scan charge: `gas/nested_alloc` → 8559693635 (a further
    +424 over 54 cycles), `gas/compute_map_key_big_bytes` +4 (one slot,
    one cycle), `gas/compute_map_key_big_struct` +36 (nine slots).
    Programs with shallow operand stacks pay almost nothing; the charge is
    proportional to the depth actually scanned.
  - The effective allocation ceiling: each recounted slot costs 40 B at
    every GC, so the 10 pre-existing `alloc_*` filetests move by 40 B per
    live slot — +80 where two are live (`alloc_0` 8094 → 8174), +120 in
    `alloc_defer_gc` (10216 → 10336) where three are. `gas/slice_alloc`
    `n` is lowered by 3 elements to stay under the cap in isolation.
- Must land with the rest of the series, not alone. The anchors themselves
  moved no golden: they only matter when a GC lands inside a fill, which no
  existing fixture does.
- Hot-path cost is one slice append and truncate per composite literal and
  per call (the backing array is retained across pops, so it stops
  allocating after warm-up). `BenchmarkOpCall_10Params_0Captures` and
  `BenchmarkOpCallWarm` show no regression above run-to-run noise, and
  `alloc-gas/op` is unchanged.
- Tests: `TestGarbageCollect_VisitsOperandStack` (before/after bytes with a
  large operand-stack-only value; the gas assertion reads the gas
  `GarbageCollect` itself consumes on two machines that differ only in
  whether the value is parked, so it fails if the `m.Values` walk is
  removed), filetests `alloc_13_operand_stack.gno` (four 300 KB elements
  under a 1 MB cap), `alloc_14_operand_stack_nil_v.gno` (primitive slots),
  `alloc_15_slicelit_alias.gno`, `alloc_16_blank.gno`,
  `alloc_17_structlit.gno`, `alloc_18_structlit_keyed.gno`,
  `alloc_19_call_args.gno`, `alloc_20_varg.gno`, `alloc_21_slicelit2.gno`
  (one per root closed above — each one passes on the merge base and is
  refused here), and txtar `alloc_operand_stack` (the 560 MB array literal,
  asserted refused before deposit).

### Known gaps, left to a follow-up

All deterministic and all fail-safe: the cap trips earlier, never later.

- **Slot wrapper over-count.** The 40 B `TypedValue` wrapper is recounted per
  slot but never charged by runtime. Bounded: each live slot costs metered
  work (a source byte for literal width, a charged `Block` for recursion
  depth). Kept on purpose: for primitive slots (`V == nil`) the wrapper is
  the only thing the cap sees, so dropping it would hide the stack's own
  footprint. The follow-up reconciles runtime and GC by charging the
  operand stack at runtime, by capacity rather than per push: `PushValue`
  grows the backing array itself (VM-owned doubling, so the charge does not
  depend on Go's `append` policy) and charges `allocTypedValue` × the added
  capacity at that moment, and GC recounts `cap(m.Values)` instead of
  walking slots. Pops change `len`, not `cap`, and Go never shrinks the backing
  array, so there is nothing to refund on the pop side and no reslice site
  to cover; the charge fires O(log n) times and tracks the retained memory
  exactly.
- **String alias over-count.** `vis` dedups only `Object`s; a `StringValue`
  is not one, so a string aliased into N slots is charged N × its length.
  Not specific to this root — blocks, arrays, structs and frames alias-count
  strings the same way today. The follow-up dedups string backings inside
  `vis`, once per GC cycle, so every root gets the same answer.
- **Anchored container headers.** While a buffer is anchored the walk
  recounts its slots but not the header of the object that will own them
  (`allocArray`, `allocStruct`, `allocSlice`). A constant per literal, and
  only until the op roots the object.
- **Other roots not walked:** `m.Realm`'s `created/updated/deleted/escaped`
  lists, `Frame.LastRealm`, `Frame.Cur`. Not amplification vectors (aliased
  into visited slots, or realm objects excluded from GC by design);
  root-completeness items.
- **Assignment destinations that are not block slots.** `doOpAssign` pops
  both its RHS values and its LHS operands, and `resolvePointer` derives
  `pv.Base` for an `IndexExpr` / `SelectorExpr` / `StarExpr` /
  `CompositeLitExpr` LHS from those popped operands. For a `NameExpr` LHS
  the destination is a rooted block slot and the miss is one in-flight copy,
  but for the other forms the destination is as invisible as the operands, so
  the miss is N. Measured: ten `f()[0]` destinations taking ten copies of one
  aliased 300 KB value build 3 MB under a 1 MB cap (60 MB at 200 terms). An
  anchor does not fit here — the destinations are not a buffer this op
  allocates — so closing it needs either an object anchor per destination or
  the runtime-charging follow-up above, which removes the dependence on
  GC seeing in-flight copies at all.
- **`doOpArrayLit`'s anchor is defence in depth, not a closed vector.**
  `defaultArrayValue` allocates the destination at full size before the copy
  loop, and for a value element type a copy is the same size as the default
  it replaces, so the exposure is ~2x rather than N x. Attempts to build a
  discriminating case failed: `[N]any{a, ...}` and `[N]string{s, ...}` both
  materialise their copies on the operand stack before `doOpArrayLit` runs,
  where the `m.Values` walk already counts them. Removing the anchor moves
  `gas/nested_alloc` and nothing else — no allocation-cap fixture covers it.
- **Stale operand tails.** `PopValue` only re-slices; `Release()` clears the
  first 512 backing slots, so slots ≥ 512 stay pinned for Go's GC while
  invisible to the allocator.

Two comments (`store.go`, `keeper.go`) had cited the `m.Values` blind spot as
the reason the preprocess allocator keeps `collect = nil`. The actual reason
is that `preprocessAlloc` is shared by every preprocess sub-Machine in the tx,
so a GC callback bound to one machine would walk only that machine's roots;
the comments — and the test that pins the invariant — now say so.

Ordering: this fix lands here, then upstream; the follow-ups are normal
public PRs based on it.
