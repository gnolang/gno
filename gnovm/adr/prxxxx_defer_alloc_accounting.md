# Count pending defers against the allocation limit

## Context

`doOpDefer` built each `Defer` record and its copied arguments as plain Go
memory and never called `m.Alloc`. A loop that only registers defers never
raised the allocator total, so GC never ran. At the 3B block gas limit, such
a loop kept 1.46GB of heap alive against a 500MB `maxAllocTx` (issue #153).

## Decision

After `PushDefer`, `doOpDefer` calls `m.Alloc.AllocateDefer(len(args))`,
which charges `allocDefer + allocTypedValue*args`. `Frame.Visit` also
recounts `allocDefer` for each pending defer, so the charge and the GC
recount agree. The same loop now fails with `allocation limit exceeded` at
about 530MB.

A deferred nil func keeps no arguments, since the call panics when it runs.
`doOpDefer` used to leave them on `m.Values`, where neither the charge nor
the GC saw them, so `defer nilf(<256 ints>)` in a loop kept 1.55GB alive at
the 3B gas limit. It now pops them, and the record is charged with 0 args.

## Alternatives considered

The issue proposed two fixes, and we tried both. We kept the first one,
charging through `m.Alloc`.

- **Make the GC walk count pending defers.** `Frame.Visit` already counts
  them. The walk never runs, because nothing raises the allocator total.
- **Count the bytes without charging gas** (`Allocator.Reserve`). This sets
  the same limit, but it adds an allocator API. It also leaves a real
  allocation unpaid, while every other VM allocation pays allocation gas.

## Consequences

- Each defer statement costs 3–5% more gas. The `gas/defer_*` and
  `alloc_defer_gc` goldens change.
- A transaction whose pending defers exceed `maxAllocTx` now fails with an
  allocation error.
