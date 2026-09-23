# AssertOriginCall: count realm boundaries, not call frames

## Status

Proposed in #6218. Follow-up to #6211.

## Context

#6211 made `AssertOriginCall` count call boundaries: every named call, plus
every func literal that ran under a different storage realm than its caller.
The chain required `<= 2`, the test runtime exact totals. Two problems:

- A `/p/` body could become the "caller" of the next frame, so a realm's own
  closure handed to `helper.Run` (`var Run = func(f func()) { f() }` in a
  `/p/` package) was refused. Master before #6211 accepted this shape.
- Whether a frame counted depended on spelling: `func f()` counted, `var f =
  func()` did not. So `helper.RunNamed` and the realm's own `func run(f
  func())` were refused while the closure versions passed. The rule protected
  nothing: a `/p/` package cannot be a message target or hold coins, and a
  realm's own helper is the realm's own code.

The fix is in two parts, each with its own consequence.

## Decision 1: storage ownership from the PkgID bit

`ownsItsStorage(r)` is `r.ID.IsRealmPkg()`, the immutable bit the finalizer
reads to decide what to persist, instead of `IsRealmPath || IsEphemeralPath`
on the frame's path. The origin check and persistence then share one answer
about who owns storage, and there is no regexp per frame (native bench stays
at 0 allocs). Realms compare by ID, since one path can be loaded twice.

Doing so exposed that the bit and the path predicates disagreed:

- `IsRealmPath`/`IsPPackagePath` rejected a `_test` suffix only in the REPO
  segment, so single-segment `gno.land/r/foo_test` read as a realm while its
  PkgID was immutable. They now reject the suffix wherever it lands.
- Synthetic packages (`.dontcare`) carried no immutable bit; they join uverse.
- `main` (filetests, `gno run`) is dot-free, so it matched `IsStdlib` and was
  immutable by accident. It is now named in the immutable set so that is a
  decision. Reclassifying it as a transient program flips borrow rules,
  construction-time checks and finalization for every `zrealm_*` filetest,
  so it is a separate change (`prxxxx_main_transient_program.md`).

`TestPkgIDOwnsStorage` pins the agreement for every path shape. Cost: a
filetest cannot assert an origin call from `package main`; `std13`-`std17`
declare `// PKGPATH: gno.land/r/test`.

## Decision 2: count realm crossings, not frames

`NumCallBoundaryFrames(start)` walks from the entry frame inwards. A frame's
body realm is the next call frame's `LastRealm` (`m.Realm` for the innermost).
A frame counts when its body owns storage and differs from the nearest
storage-owning realm below it; the first such frame is the entry. Named and
literal functions are treated alike; `/p/` and stdlib bodies never count and
never become the caller, so they are transparent in both directions; basic
frames (for/range/switch) never count.

The chain requires exactly one boundary: the message entering the realm the
entry check already pins to the message's path. Any other storage-owning
realm on the stack is a second boundary and is refused.

The test runtime (`gnovm/tests/stdlibs/chain/runtime`) shares the walk from
the frame after the test function (`main`/`init.*`: 1, `RunTest`: 3), which
stands in for the message. The asserting frame may not itself be the entry.

## Alternatives

- **Fix only the `/p/` closure case, keep named frames counting.** Leaves the
  named-vs-literal asymmetry and the same-realm helper refusal in place.
- **Treat a `/p/` closure as its declaring package's code.** Rejected in #6211:
  the declaring package is not where the writes land.
- **Keep the path predicates in `ownsItsStorage`.** Two sources of truth for
  one question; they had already drifted on `_test` and synthetic paths.
- **Test runtime keeps exact frame counts.** Cannot express the new rule, and
  the test function's package is not a realm the chain would ever see.

## Consequences

- Contract change: a realm's own named helper (`myrlm.A -> myrlm.C`), a `/p/`
  helper named or literal, and the realm's own closure handed to either all
  pass. `assertorigincall.txtar` case 1 flips; `std14` case 3 flips.
- Still refused: any other realm between the message and the assertion,
  whether entered by cross-call, by a non-crossing call into its function or
  exported closure, or by handing it a closure to run. Pinned in
  `assertorigincall.txtar`, `assertorigincall_p_closure_helper.txtar`, `std16`.
- No new exposure: every interposing party has a frame whose body runs in
  its own storage realm, and every such frame counts. What became transparent
  was already reachable by writing the helper as a closure.
- The `native.gno` doc comment changed with the same line count; its source
  bytes are genesis state, so `apphash_crossrealm38_test.go` is re-pinned.
