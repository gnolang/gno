# ADR-xxxx: Cycle guard in `fillValueTV`

## Status

Proposed. Bug fix, no consensus-visible behavior change for programs that
previously terminated.

## Context

`fillValueTV` (`gnovm/pkg/gnolang/values.go`) resolves lazily-loaded values:
`RefValue` -> the stored object, and a `PointerValue` -> its base plus the `TV`
it points at. To resolve a pointer it descends into the slot the pointer names —
a heap item's value directly, or an array element / struct field / block slot
via `GetElementPointer` / `GetPointerToInt`, which fill that slot in turn.

The value graph it walks can be cyclic, and nothing tracked what had already
been visited. Storing a pointer in the slot it points at is ordinary Gno, and
closes a cycle in one line:

```go
var x any; x = &x       // block slot (heap item)
s.a = &s.a              // struct field
arr[0] = &arr[0]        // array element
```

Reading the value back recursed until the Go stack overflowed. `fatal error:
stack overflow` is not recoverable, so this aborted the process rather than the
transaction: a realm holding such a value crashes the node of every validator
that replays the call, and — since the crash is on the load path — of any node
serving a `vm/qeval` read of it. A chain halt reachable by anyone who can
publish a package, and a `gno run`/`gno test` crash off-chain. Mutual cycles
(`a = &b; b = &a`) and longer ones behave the same.

## Decision

Thread a visited set through the walk, keyed on the **slot** (`*TypedValue`)
being filled rather than on any one container type, so that one key covers every
shape a cycle can close through: heap item value, struct field, array element,
block slot.

- `fillValueTV` keeps its signature and delegates to `fillValueTVGuarded`.
- The three index helpers reached from inside the walk —
  `(*ArrayValue).GetElementPointer`, `(*StructValue).GetPointerToInt`,
  `(*Block).GetPointerToInt` — gain unexported variants carrying the set. They
  are call sites *inside* the recursion: without threading, each one restarts
  the walk with an empty set and the guard never fires.
- Stopping at a revisited slot loses no work: to be revisited it must already be
  in memory, so a cycle can hold no unresolved `RefValue`, and re-descending
  would only redo what the first visit did.
- The set is not built until the walk passes `fillGuardDepth`, 32. Every fill the
  VM evaluates on the hot path is shallower than that and pays one compare; a
  cycle recurses to that depth and is caught within one more turn around itself.
- `(*TypedValue).IsReadonlyBy` gets the same treatment, keyed on the heap item.
  It is a second walker over this graph, and today the `fillValueTV` crash hides
  it: a self-referential value dies in the fill before anything reaches it.
  Guarding the fill alone moves the halt rather than removing it, because writing
  through such a pointer descends from a heap item into its own value. Its single
  descent is the heap-item one, every other arm reading an `ObjectID` and
  stopping, so there the heap item is a sufficient key.

## Alternatives considered

- **Keying the set on `*HeapItemValue`** (the first shape of this fix). Rejected:
  run as realm round trips it overflows on all three slot shapes, `var x any;
  x = &x` included, because a cycle closes through the field, element or block
  slot the pointer names rather than through a heap item.
- **A path stack instead of a visited set**, scanned linearly. Rejected: it
  would be `O(n^2)` on the deep-but-acyclic chains that are legal today, where
  the set is `O(n)` and additionally skips shared subgraphs already filled.
- **Depth limit** (as `MaxValueRenderDepth` does for rendering), meaning a depth
  that aborts the walk. Rejected: chains of escaped pointer-holding locals of
  arbitrary depth are legal and must keep resolving, so a constant low enough to
  protect the stack rejects working programs. `fillGuardDepth` is not this: it
  changes when the set is built and never stops the walk, so no depth becomes
  illegal.
- **A "being filled" flag on the object**, a `bool` on `HeapItemValue` rather
  than a set. Cheapest of all at runtime, and rejected for three reasons. It
  grows `HeapItemValue` from 192 to 200 bytes, and `_allocHeapItemValue` is a
  constant the VM charges allocation against, so every escaped local in every
  deployed realm would be billed differently: a consensus change inside a crash
  fix. A flag on the heap item is also the rejected key above, guarding none of
  the field and element cycles, and per-object flags on `StructValue` and
  `ArrayValue` would be coarser than the slot they need to track, skipping a
  legitimate fill of a sibling field reached later in the same walk. Lastly
  `fillValueTV` can panic partway (store read, out of gas), leaving the flag set
  unless every path clears it. The same flag on `IsReadonlyBy` avoids the first
  two objections, since that walker descends only through a heap item, but it
  buys less there than the map costs.
- **Making the fill lazy** (the `XXX but why, isn't lazy better?` already in
  the function). A real simplification, but it changes when every `RefValue` in
  the graph is resolved, and when the store reads behind it are gas-charged;
  out of scope for a crash fix.

## Consequences

- Programs that terminated before behave identically; the guard only fires
  where the old code could not return.
- Cost on the hot path: `canRefillTV` gates the bookkeeping on whether filling
  this slot can reach it again — only a pointer into a heap item or the store,
  and a heap item holding one of those, can. Anything else (an escaped local
  holding an int, a string, a struct) ends the descent. `seen` does not escape,
  so it costs a map init rather than a heap allocation: `allocs/op` is 0 before
  and after. A closure hammering an escaped `int` counter (2M calls, n=20, Apple
  M4 Pro, go1.26.4) costs about 1.7% more than before; without the gate it costs
  about 20% more.

  Deferring the set past `fillGuardDepth` keeps the pointer shapes off the map
  too: on `(*Block).GetPointerToInt`, `p := &x` measures 3.29ns before and 4.30ns
  after, and an escaped local holding a pointer 6.46ns before and 8.53ns after.
  A 30M-iteration loop reading a pointer-valued local twice per iteration runs
  7.28s before and 7.68s after, 5.5% (median of 3). Without the depth deferral
  the same loop runs 7.95s, 9.2%.
- The sibling walkers checked here handle cycles already:
  `copyValueWithRefs` / `getChildObjects` / `fillTypesOfValue` stop at object
  boundaries, `incRefCreatedDescendants` and `decRefDeletedDescendants` have
  explicit recurse guards, the GC stamps `LastGCCycle`, and the string renderers
  cap depth. Two stop for weaker reasons and are worth naming rather than
  counting as guarded: `saveUnsavedObjectRecursively` terminates because a cycle
  drives the refcount above 1 and the escape branch then skips the object, and
  `Value.DeepFill` has no guard at all, reached today only by generated native
  bindings whose parameters are all primitives.
- The walk is still bounded by the Go stack rather than by a depth limit, so an
  acyclic chain of about a million pointer-holding slots aborts the process as it
  did before. Gas meters the transaction paths and not the query paths, so that
  bound is worth revisiting on its own; it is not what this change is for.
- Tests: `gnovm/tests/files/heap_self_pointer.gno` covers the in-memory cycles
  (self, mutual, struct field, array element, and a 100-deep acyclic chain that
  must still resolve); `gno.land/pkg/integration/testdata/selfref_pointer.txtar`
  stores all three slot shapes in a realm and reads each back in a later
  transaction, covering the store round trip and asserting the cycle survives
  it, and writes through a cycle in a single call for the `IsReadonlyBy` descent.
  Reverting `values.go` crashes both files; reverting `ownership.go` crashes the
  write-through step alone. In memory the filetest's `field` and `elem` cases
  terminate without the fix, so it is the txtar that pins those two shapes.
