# Drop vestigial Defer.Parent; recycle defer-origin blocks

## Context

#5813 introduced the per-machine runtime block pool (see
`pr5813_block_pool.md`), but conservatively excluded "defer-site" blocks: a
`Defer` recorded the block it was registered in as `Defer.Parent`, and
`doOpDefer` marked that block `notRecyclable`, so the VM-GC kept counting it
live until the defer ran.

That exclusion is unnecessary. A deferred call resolves its scope from
`FuncValue.GetParent` + copied `Captures` at execution — closures set
`Parent: nil` and copy heap-item captures; non-closure defers resolve to the
long-lived file/package block; arguments are copied at the defer statement —
never from the block the defer was registered in. A defer-origin block is
therefore provably dead once popped, and `Defer.Parent` was vestigial: its
only readers were the VM-GC re-count (`Frame.Visit`) and the `releaseBlock`
`debugAssert`.

## Decision

- Drop the `Defer.Parent` field (`frame.go`) and the `Block.notRecyclable`
  flag plus `setNotRecyclable`/`isNotRecyclable` (`values.go`).
- `doOpDefer` (`op_call.go`) no longer captures the last block or sets the flag.
- `Frame.Visit` (`garbage_collector.go`) no longer visits `dfr.Parent`, so a
  popped defer-origin block — and any uncaptured locals it held — is freed at
  the next GC.
- Replace the `releaseBlock` `debugAssert` (which scanned every frame's pending
  `Defer.Parent` list) with a reference-path-agnostic check in
  `GarbageCollect`: the recount must never reach a pooled block (a pooled block
  is zeroed on release, so `LastGCCycle == 0`, while the visitor stamps the
  current cycle on everything it reaches).

## Alternatives considered

- Keep `Defer.Parent` but visit it only for the GC re-count: rejected — it
  pins dead blocks (and their uncaptured locals) in the live set for no reason.
- Drop `Defer.Parent` but keep the `notRecyclable` flag: rejected — with no
  `Defer.Parent` reader, the flag guards nothing.

## Consequences

- Defer-origin blocks now recycle through the pool once popped; their
  uncaptured locals are reclaimed at the next GC (pinned by `alloc_defer_gc.gno`).
- `sizeof(Block)` and `_allocBlock` stay 536: `poisoned` (added by #5813) still
  occupies the tail padding word that `notRecyclable` used, so this PR does not
  shrink `Block` and moves no gas golden.
- No existing gas/alloc/txtar goldens change. New regression tests:
  `gnovm/tests/files/defer_block_recycle.gno` (named-return mutation,
  per-iteration churn, `continue`, nested defer) and
  `gno.land/pkg/integration/testdata/defer_realm_recycle.txtar` (deferred
  realm-state write survives restart).

## Verification

```sh
go test ./gnovm/pkg/gnolang/ -run 'Files$/defer_block_recycle|Files$/alloc_defer_gc' -test.short -count=1
go test -tags debugAssert ./gnovm/pkg/gnolang/ -run 'Files$/(defer_block_recycle|alloc_defer_gc)' -count=1
go test ./gno.land/pkg/sdk/vm/ -run Gas -count=1
go test ./gno.land/pkg/integration/ -run TestTestdata -count=1
```
