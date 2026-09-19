# AssertOriginCall: look through /p/ frames when naming a closure's caller

## Status

Proposed in #6218. Follow-up to #6211.

## Context

#6211 made `AssertOriginCall` count call boundaries rather than frames: a func
literal counts only when it runs under a different storage realm than the one
that called it. The walk went from the innermost frame outwards and compared
each frame's `LastRealm` against the realm the frame above it ran in.

A `/p/` package owns no storage, so a call into it is not a boundary and its
frames are skipped. But the walk still let a `/p/` body become the *caller* of
the next frame. For

```go
// gno.land/p/demo/helper
var Run = func(f func()) { f() }

// gno.land/r/wug
func Deposit(cur realm) {
	helper.Run(func() { runtime.AssertOriginCall() })
}
```

the realm's own closure was compared against the frozen `/p/helper` realm,
read as a boundary, and the count reached 3: `invalid non-origin call`. The
only realm on the stack is `wug`; a `/p/` package can be neither a message
target nor a coin recipient, so nothing could have interposed. Master before
#6211 accepted this shape, and the PR's compatibility note did not list it.

## Decision

Walk from the entry frame inwards and track the nearest storage-owning realm
below the frame being decided. A frame's body realm is the next call frame's
`LastRealm` (`m.Realm` for the innermost). A closure counts as a boundary when
its body owns storage and differs from that tracked caller; `/p/` and stdlib
bodies never become the caller, so they are looked through in both directions.

Named functions and frames entered from the message (`LastRealm == nil`) count
as before. A named `/p/` helper (`func RunNamed(f func())`) is still a
boundary, unchanged from master.

"Owns storage" is read off the `PkgID` immutability flag (`IsRealmPkg`), the
same bit the finalizer uses to decide what to persist, instead of matching the
path with a regexp on every frame. Realms are compared by ID, since one path
can be loaded twice.

## Alternatives

- **Skip `/p/` frames but keep the outward walk.** Needs a look-behind for the
  caller on each step; the inward walk gets the same answer with one variable.
- **Treat a `/p/` closure as its declaring package's code.** Rejected in #6211:
  the declaring package is not where the writes land.

## Consequences

- The honest shape passes again; a relay realm routing another realm's closure
  through the same helper is still refused, because the tracked caller is the
  relay and the closure body is the other realm. Both are pinned by
  `gno.land/pkg/integration/testdata/assertorigincall_p_closure_helper.txtar`.
- The test runtime (`gnovm/tests/stdlibs/chain/runtime`) shares the counter,
  so `gno test` moves with the chain. Under `gno test` the `main` package
  carries a realm the flag marks as storage-owning, so it is tracked as a
  caller like any `/r/` realm rather than looked through.
