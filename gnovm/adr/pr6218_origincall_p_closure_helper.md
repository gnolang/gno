# AssertOriginCall: count realm boundaries, not call frames

## Status

Proposed in #6218. Follow-up to #6211.

## Context

#6211 made `AssertOriginCall` count call boundaries: every named call, plus
every func literal that ran under a different storage realm than its caller.
The walk compared each frame's `LastRealm` against the realm the frame above
it ran in, and let a `/p/` body become the caller of the next frame, so

```go
// gno.land/p/demo/helper
var Run = func(f func()) { f() }

// gno.land/r/wug
func Deposit(cur realm) {
	helper.Run(func() { runtime.AssertOriginCall() })
}
```

was refused: wug's own closure was compared against the frozen `/p/helper`
realm. Master before #6211 accepted this shape.

Fixing only that left a second asymmetry. `helper.RunNamed`, the same helper
as a named function, was still refused, as was wug's own `func run(f func())`.
The documented contract said as much ("even from the same realm or package"),
but the rule protected nothing: a `/p/` package cannot be a message target or
hold coins, and a realm's own helper is the realm's own code. Whether a frame
counted depended on whether the author wrote `func f()` or `var f = func()`.

## Decision

`NumCallBoundaryFrames` counts realm boundaries. Walking from the entry frame
inwards, a frame's body realm is the next call frame's `LastRealm` (`m.Realm`
for the innermost). A frame counts when its body owns storage and differs from
the nearest storage-owning realm below it; the first such frame is the entry.
Named and literal functions are treated alike. `/p/` and stdlib bodies never
count and never become the caller, so they are transparent in both directions.

"Owns storage" is read off the `PkgID` immutability flag (`IsRealmPkg`), the
bit the finalizer uses to decide what to persist, instead of matching the path
with a regexp per frame. Realms are compared by ID, since one path can be
loaded twice.

The chain requires exactly one boundary: the message entering the realm the
entry check already pins to the message's path. Any other storage-owning realm
on the stack is a second boundary and is refused.

The test runtime (`gnovm/tests/stdlibs/chain/runtime`) shares the counter via
`NumCallBoundaryFramesFrom(start)`, with the test function or `main` standing
in for the message: the first storage-owning realm entered after it is the
entry, and the asserting frame may not itself be the entry. A filetest's
`main` package is not storage-owning, so filetests that assert from their own
code declare a `// PKGPATH: gno.land/r/...`.

## Alternatives

- **Fix only the `/p/` closure case, keep named frames counting.** Leaves the
  named-vs-literal asymmetry and the same-realm helper refusal in place.
- **Treat a `/p/` closure as its declaring package's code.** Rejected in #6211:
  the declaring package is not where the writes land.
- **Test runtime keeps exact frame counts.** Cannot express the new rule, and
  the test function's package is not a realm the chain would ever see.

## Consequences

- Passes now: a realm's own named helper, a `/p/` helper named or literal,
  and its own closure handed to either. `assertorigincall.txtar` case 1
  (`myrlm.A -> C`) flips to pass.
- Still refused: any other realm between the message and the assertion,
  whether entered by cross-call, by a non-crossing call into its function or
  exported closure, or by handing it a closure to run. Pinned in
  `assertorigincall.txtar` and `assertorigincall_p_closure_helper.txtar`.
- No new exposure: every interposing party has a frame whose body runs in
  its own storage realm, and every such frame counts. What became transparent
  was already reachable by writing the helper as a closure.
