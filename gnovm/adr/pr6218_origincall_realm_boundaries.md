# AssertOriginCall: count realm boundaries, not call frames

## Status

Proposed in #6218. Follow-up to #6211.

## Context

#6211 made `AssertOriginCall` count call boundaries: every named call, plus
every func literal run under another storage realm. Two problems:

- A `/p/` body could become the "caller" of the next frame, so a realm's own
  closure handed to `helper.Run` (`var Run = func(f func()) { f() }` in a
  `/p/` package) was refused. Master before #6211 accepted this shape.
- Whether a frame counted depended on spelling: `func f()` counted, `var f =
  func()` did not, so `helper.RunNamed` and the realm's own `func run(f
  func())` were refused while the closure versions passed, protecting nothing.

## Decision 1: storage ownership from the PkgID bit

`ownsItsStorage(r)` is `r.ID.IsRealmPkg()`, the immutable bit the finalizer
reads, instead of `IsRealmPath || IsEphemeralPath` on the path: one answer, no
regexp per frame (0 allocs). Realms compare by ID; one path can be loaded twice.

Doing so exposed that the bit and the path predicates disagreed:

- `IsRealmPath`/`IsPPackagePath` rejected a `_test` suffix only in the REPO
  segment, so single-segment `gno.land/r/foo_test` read as a realm while its
  PkgID was immutable. They now reject the suffix wherever it lands.
- The bit was a denylist (stdlib, `/p/`, overlays, uverse), so any string on
  no list got the realm bit: synthetic `.dontcare`, the keeper's `""`
  message-entry package, an overlay of a non-realm such as
  `gno.land/r/x_test_test`, or any unvalidated input.

The bit is now an allowlist: storage-owning iff `IsRealmPath ||
IsEphemeralPath`, the predicate the loaders use, so the two agree for every
input by construction and an unvalidated path fails closed (immutable, no
storage) instead of open. Under `debugAssert`, `PkgIDFromPkgPath` panics on a
path no predicate recognizes (`isRecognizedPkgPath`), since every user path is
validated upstream (`MsgCall`, `AddPackage`, `MemPackageType.Validate`) and an
unrecognized one is a missed validation; on a validator it stays quiet and
deterministic. Test fixtures under `gno.vm/t/` and `gno.land/t/` moved to
`gno.land/p/t/` so the repo's own suites are clean under the tag.

`main` (filetests, `gno run`) falls out as immutable, as before via
`IsStdlib`, so a filetest cannot assert an origin call from `package main` and
`std13`-`std17` declare a `// PKGPATH:`. Making it a realm (the local twin of
a `/e/.../run` path: allowlist it, drop its stdlib bit) would also flip borrow
rules, construction-time checks and finalization for every `zrealm_*` filetest
that relies on `main` being transparent, so it is deferred.

`TestPkgIDOwnsStorage` and `TestPkgIDUnrecognizedPath` pin both behaviours.

## Decision 2: count realm crossings, not frames

`NumCallBoundaryFrames(start)` walks from the entry frame inwards. A frame's
body realm is the next call frame's `LastRealm` (`m.Realm` for the innermost).
A frame counts when its body owns storage and differs from the nearest
storage-owning realm below it; the first such frame is the entry. It also
counts when it runs a named function declared in the very realm its body runs
in, above the entry. `/p/` and stdlib bodies never count and never become the
caller, so they are transparent in both directions, and that includes a `/p/`
method borrowed into the realm; a realm's own func literals are transparent
inside it; basic frames (for/range/switch) never count.

The named rule is the one asymmetry kept on purpose. A realm's exported named
function is a value any `maketx run` script can name and hand back to it
(`SetHook(cross, rlm.Withdraw)`); if the realm later invokes stored callbacks
on a guarded path, that function runs with one boundary on the stack and the
check cannot tell it from the realm's own control flow. A func literal inside
the realm is reachable only through code the realm wrote, and a `/p/` body
cannot name realm code at all, so neither gives an outsider that lever. Master
refused every same-realm named call by counting frames; this keeps that
refusal without the frame count, so the contract change versus master is
exactly the regression fix: `/p/` helpers and own closures, however spelled.

The chain requires exactly one boundary: the message entering the realm the
entry check already pins to the message's path. Any other storage-owning
realm on the stack is a second boundary and is refused.

The test runtime shares the walk from the frame after the test function
(`main`/`init.*`: 1, `RunTest`: 3). The runner's call into that function is
the message and the function is the message's own script: its realm is never
an intermediary, even under an `/r/` `PKGPATH`, so a realm filetest can drive
other realms like a user, but only by calling them directly (a wrapping
closure or helper enters the file's realm first and makes it the entry). A
live `testing.SetRealm(NewCodeRealm(p))` below the entry stands in for a code
caller `p`: one more realm unless `p` is the entry realm itself.

## Alternatives

- **Fix only the `/p/` closure case.** Leaves master's frame count, which
  also refused named `/p/` helpers and depended on how many frames a call
  pushed.
- **Treat named and literal same-realm functions alike.** Symmetric and
  simpler, but admits the stored-callback shape above, which master refused;
  the ergonomic gain (`myrlm.A -> myrlm.C`) is not needed for the fix.
- **Treat a `/p/` closure as its declaring package's code.** Rejected in
  #6211: the declaring package is not where the writes land.
- **Keep the path predicates in `ownsItsStorage`.** Two sources of truth;
  they had already drifted on `_test` and synthetic paths.
- **Panic unconditionally on an unrecognized path.** Local tooling runs on
  arbitrary module paths, and a keeper panic is the wrong failure mode for a
  bug that lets a string through; fail closed on chain, loud under `debugAssert`.

## Consequences

- Contract change: a `/p/` helper named or literal, and the realm's own
  closure handed to it, pass (`assertorigincall_p_closure_helper.txtar`).
  A realm's own named function between the entry and the assertion stays
  refused, as on master (`assertorigincall.txtar` case 1, `std13`, `std14`).
- Still refused: any other realm between the message and the assertion,
  however entered (cross-call, non-crossing call, closure handed over,
  re-entry, simulated via `SetRealm`). Pinned in `assertorigincall*.txtar`
  (cases 29-30 are re-entry), `std16`, `std19`, `r/tests/vm/tests_test.gno`.
- No new exposure versus master: every interposing realm has a frame whose
  body runs in its own storage realm and counts; a foreign closure, which
  master's count skipped, counts too. The remaining gap, shared with master,
  is a realm's exported closure variable handed back to it; the value-based
  guard `cur.Previous().IsUserCall()` closes it and is what payment code
  should use. `AssertOriginCall` is not a re-entrancy lock.
- `native.gno`'s doc comment is genesis state: `apphash_crossrealm38_test.go` re-pinned.
