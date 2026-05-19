# Cross-Call Migration Guide

Practical learnings from the runtime.{Current,Previous}Realm → `cur`
threading migration, and the subsequent collapse of the bare-`cross`
sentinel into the explicit `cross(rlm)` form. Each entry is a general
rule distilled from a concrete pitfall encountered during migration.

This is a living document. Append new learnings as discovered.

**Migrating a codebase from bare `cross`?** Start with §16 — the
two-step recipe (`cross` → `cross1` → `cross(rlm)`) gets you a
mechanical bulk rename first, then per-site semantic threading.

---

## 1. Crossing-fn vs non-crossing-fn-with-rlm: choosing the form

Two ways to thread `realm` through a function signature:

| Form | Semantics |
|---|---|
| `func F(cur realm, ...)` | **Crossing function.** Called via `F(cross(rlm), ...)`. Each cross-call mints a fresh `cur` in the callee, creates a realm boundary, triggers finalization on return. |
| `func F(_ int, rlm realm, ...)` | **Non-crossing helper** that *takes a realm value*. Called as `F(0, cur, ...)` (plain value pass). No boundary, no finalization, no shift in `runtime.{Current,Previous}Realm()`. |

The `_ int` discriminator in the second form is required: without it,
the parser sees `func F(rlm realm, ...)` and reads it as a crossing
function — which is forbidden in `/p/` packages and changes call-site
semantics.

**Default rule:**

- **Exposed-to-EOA entry points** (MsgCall targets, the `func Foo(cur
  realm, ...)` in a `/r/` realm that humans invoke) → crossing function.
- **Internal / unexposed helpers** (private getters, validators,
  forwarders within the same realm) → `_ int, rlm realm` non-crossing
  helper.

Reason: only the outermost call from an EOA into a realm should shift
`runtime.PreviousRealm()`. Turning a previously-non-crossing helper
into a crossing function changes what `cur.Previous()` returns inside
the body and inside everything it calls. That's a behavioral change,
not a refactor — audit downstream uses of `PreviousRealm()` /
`cur.Previous()` before flipping.

See the `boards2/v1/hub` and `crossingFn`-removal commits for clean
examples of "thread `cur` through, drop the IIFE same-realm-cross,
don't change crossing/non-crossing status".

---

## 2. Historical: `cross(cur)` ≠ bare `cross` for tests using `testing.SetRealm`

> Resolved by the bare-`cross` removal — only `cross(rlm)` remains.
> Preserved here as background context for the test-realm interaction
> in §3 and §11.

When the bare-`cross` sentinel still existed, it minted the new cur
differently from `cross(rlm)`:

- **Bare `cross`** (preprocessor sentinel): runtime's
  `installCrossingCur` took the `argtv.IsUndefined()` path and called
  `m.callingCurOrOrigin()`. If no crossing frame was found, it fell
  through to `buildOriginRealm(m)` which **dynamically read
  `m.Context.OriginCaller`** and constructed a fresh realm with
  `addr=OriginCaller, pkgPath="", prev=truly-nil`.
- **`cross(rlm)`**: takes the `else` branch in `installCrossingCur`
  and uses `*argtv` as prev — the static rlm value passed in.

In test scopes where the pattern was
`testing.SetRealm(NewUserRealm(addr))` followed by an IIFE
same-realm-cross to mint a fresh cur whose prev is the just-set user,
the bare-`cross` form picked up `OriginCaller=addr` from the updated
context. With only `cross(rlm)` available now, tests that depended on
this dynamic-origin behavior must instead either mutate `cur` in place
via SetRealm before the cross-call (§3, §11) or take the caller address
as an explicit parameter.

---

## 3. `testing.SetRealm` mutates `fr.Cur` in place

`testing.SetRealm(NewUserRealm(addr))` doesn't just update
`ctx.CurrentRealm`. It walks the call stack to the first non-testing
frame and **mutates that frame's `fr.Cur` fields in place** (see
`gnovm/tests/stdlibs/testing/context_testing.go` SetContext path).

Concretely, after `SetRealm(NewUserRealm(white))` from inside a test
fn whose first param is `cur realm`:
- `fr.Cur.addr` is overwritten to `white`'s addr
- `fr.Cur.pkgPath` is overwritten to `""`
- `fr.Cur.prev` is set to truly-nil

Because the parameter `cur` and `fr.Cur` share the same underlying
`*HeapItemValue`, reads through `cur` see the mutated values. This is
why `cross(cur)` inside the test, post-SetRealm, can act as the
SetRealm'd user — the mutation propagates through the HIV identity.

**Implication for testing in `_test.gno`:** a test function declared
`func TestXxx(cur realm, t *testing.T)` and called via `gno test` is
the only frame where this mutation is observable. If you wrap part of
the test in a `t.Run(name, func(t *testing.T) { ... })` subtest, the
**closure captures `cur` from the outer scope** but the subtest's own
frame is a new call frame whose `fr.Cur` is unset. SetRealm targets
the deepest non-testing frame — which is now the subtest closure —
where the mutation lands on an empty fr.Cur and silently no-ops.

**The fix (preferred):** declare the subtest closure as a *crossing*
closure so it gets its own `fr.Cur`. `testing.T.Run` already accepts
`func(realm, *testing.T)` as a subtest signature (see
`testing.gno:195`, routes through `tRunner_cur`):

```go
t.Run(tc.name, func(cur realm, t *testing.T) {
    testing.SetRealm(testing.NewUserRealm(tc.caller))
    SomeCrossingFn(cross(cur))
})
```

Now `SetRealm` walks frames, finds the subtest closure's frame
(non-testing, has `fr.Cur` set because the closure is a crossing
fn), and mutates `fr.Cur` in place. The captured `cur` reflects the
mutation because the closure's `cur` and `fr.Cur` share HIV.

**Methods/funcs inside subtests** that previously took `_ int, rlm
realm` (non-crossing helper) need to be reconsidered: if they call
`SetRealm` internally and rely on the mutation reaching `rlm`, they
need to be crossing methods too (`Method(cur realm, ...)`) so that
their frame has its own `fr.Cur`. Bridge non-crossing → crossing at
the call site with `cross(cur)`.

For an example refactor that threads through methods + an interface,
see `r/morgan/chess/chess_test.gno`'s `testCommandRunner.Run(cur realm,
...)` migration.

**Alternative fallback (if subtest crossing-form is undesired):**
flatten t.Run subtests into the parent test function. Use
`println("=== case ", tc.name)` for readable boundaries.

---

## 4. Cross-realm panics are not catchable by `defer/recover`

Background: a `panic()` raised inside a frame that's borrow-shifted to
a different realm (or crosses any realm boundary on its unwind path)
goes through `PopUntilLastReviveFrame` (`op_call.go:530`). Regular
`defer { recover() }` in the caller does **not** catch — the panic
becomes an unhandled-panic transaction abort. Only `revive(fn)` is
boundary-aware.

For tests, this means the historic pattern:

```go
defer func() {
    if r := recover(); r == nil {
        t.Errorf("expected panic, got none")
    }
}()
tdao.someCrossingThing()
```

is broken if `tdao.someCrossingThing()` panics across a realm
boundary. Replace with:

```go
exc := revive(func() {
    tdao.someCrossingThing()
})
if exc == nil {
    t.Errorf("expected panic, got none")
}
```

Or use `uassert.AbortsWithMessage(t, cur, msg, fn)` / equivalent which
wrap `revive` internally. See `p/samcrew/basedao/basedao_test.gno`
for several converted call sites.

---

## 5. `_test`-suffix PKGPATHs are not realm paths

`IsRealmPath(pkgPath)` (`gnovm/pkg/gnolang/mempackage.go:80-89`)
excludes any `gno.land/r/...` path whose REPO segment ends in `_test`.
A filetest declaring

```
// PKGPATH: gno.land/r/foo/groups_test
package groups_test
```

cannot host crossing functions — the preprocessor rejects them with
*"crossing function (realm first argument) declared in non-realm
package."* This is true even though the path lives under `/r/`.

**Convention:** filetests under `examples/.../filetests/` should use
**unique non-`_test`-suffix realm paths** matching their filename:

```
// PKGPATH: gno.land/r/foo/groups/filetests/z_0_a
package z_0_a
```

This is the pattern used in `r/gnoland/boards2/v1/hub/filetests/`.

---

## 6. Foreign-realm slice mutation panics; defensive copy when sorting

A `[]string` returned from `/r/foo`'s public getter is foreign-readonly
to any caller outside `/r/foo`. Calling `sort.StringSlice(s).Sort()`
on it tries to write back through the slice header and panics with:

```
illegal conversion of readonly or externally stored value
```

(`doOpConvert` Case 1, `op_expressions.go:771`).

**Fix:** copy the slice locally before sorting:

```go
remote := treasury.ListBankerIDs()
local := append([]string(nil), remote...)
sort.StringSlice(local).Sort()
```

Generalizes to any in-place mutation of a foreign-returned composite.

---

## 7. Foreign-readonly taint propagates through value copy

Reading a foreign struct value into a local variable preserves the
`N_Readonly` taint. The local copy is **still readonly**. This is
conservative-safe but Go-semantics-divergent (Go allows mutating the
local copy).

If your code reads a field from a foreign struct and tries to mutate
the local, that's a readonly panic at write time. Solution: extract
primitive values (string, int) instead of struct values; or treat the
read as fully read-only.

---

## 8. Construction-time check refuses cross-realm composite literals

`alloc.checkConstructionTime` (`alloc.go:421`) panics:

```
cannot allocate gno.land/r/foo.SomeStruct in realm gno.land/r/bar
```

when a composite literal of a `/r/`-declared type is evaluated from a
different `/r/` realm. The same applies to `new(/r/foo.T)` and
`make([]/r/foo.T, ...)`.

**Mitigation:** call the type's home-realm constructor function (which
runs under Rule-1 declaring-realm borrow, allocating in the home
realm), not a literal. See `gov/dao/v3`'s `NewVoteRequest` /
`NewUpdateRequest` pattern.

This was the root cause of three call-site fixes during the
`zrealm_tests0` and `realm_govdao` filetest migrations.

---

## 9. Removed: dead `crossingFn` shim

A pattern from before Rule-1 covered closures uniformly:

```go
func crossingFn(fn func()) func() {
    return func() {
        func(realm) { fn() }(cross)
    }
}
```

This added a useless realm boundary (with extra finalization) around
every callback dispatch. Under current borrow rules, Rule 1 already
shifts `m.Realm` to the closure's declaring realm — no extra cross
needed. The shim is pure dead weight; removing it saves one finalize
per call.

See `boards2/v1` `d4b567a64` for the removal.

---

## 10. Filetest harness strips `_filetest` from filenames

The gnovm filetest runner (`gnovm/pkg/test/filetest.go:393`) does
`fname = strings.ReplaceAll(fname, "_filetest", "")` so the synthetic
package built around a single filetest survives MPF mempackage
filtering (which would otherwise exclude `*_filetest.gno`).

Consequence: `IsTestFile(file)` returns false for files loaded by the
filetest harness, even when the original on-disk name had the
`_filetest` suffix. Any lint rule using `IsTestFile` cannot detect
"this file is being run as a filetest." If you need to gate behavior
on filetest mode, thread an explicit option (e.g.
`MachineOptions.LegacyCross`) rather than relying on filename
heuristics.

---

## 11. `SetRealm` inside a non-crossing closure silently no-ops

A natural-looking pattern in test code:

```go
t.Run(name, func(cur realm, t *testing.T) {
    run := func() {
        testing.SetRealm(testing.NewUserRealm(voter))
        tdao.vote(0, cur, ...) // expects rlm.Previous() == voter
    }
    uassert.AbortsWithMessage(t, cur, "...", run)
})
```

**This is broken.** `run` is a non-crossing closure, so its frame has no
`fr.Cur` of its own. When `SetRealm` walks frames to find the deepest
non-testing frame, it lands on `run` first and tries to mutate
`fr.Cur` in place — but `fr.Cur` is empty, so the in-place mutation in
`X_setContext` (see `gnovm/tests/stdlibs/testing/context_testing.go`'s
`if pv, ok := fr.Cur.V.(gno.PointerValue); ok && pv.TV != nil` guard)
silently skips. Only `ctx.RealmFrames[frameIdx]` gets set, which
affects `X_getRealm` reads but not direct realm-value reads through
`fr.Cur`.

The closure-captured `cur` still refers to the *outer* subtest's HIV,
which was never mutated. `cross(cur)` inside `tdao.vote` then uses
the un-mutated cur — `rlm.Previous().Address()` resolves to `""` (the
fresh-origin default), not `voter`.

**Fix:** call `SetRealm` *outside* the non-crossing closure, in the
enclosing crossing scope:

```go
t.Run(name, func(cur realm, t *testing.T) {
    testing.SetRealm(testing.NewUserRealm(voter))  // mutates THIS frame's cur
    run := func() {
        tdao.vote(0, cur, ...)  // captured `cur` now points to mutated HIV
    }
    uassert.AbortsWithMessage(t, cur, "...", run)
})
```

The subtest closure is itself crossing (`func(cur realm, t *testing.T)`),
so it has an `fr.Cur` for SetRealm to target. After mutation, the
closure-captured `cur` reads the mutated values because closure capture
shares the variable (and the variable shares HIV with `fr.Cur`).

See `p/samcrew/basedao/basedao_test.gno` `TestVote` for the corrected
pattern.

---

## 12. `func main(cur realm)` works in `/e/` MsgRun scripts

A `maketx run` script is wrapped into an ephemeral `gno.land/e/<addr>/run`
package and invoked by `RunMain`. Originally only non-crossing
`func main()` with bare `cross` was supported, because `/e/` is not a
realm path and `crossingAllowed` (`preprocess.go:4510`) rejects
crossing-fn declarations in non-realm/non-test packages.

Narrow carve-out now permits a top-level `func main(cur realm)` in
`/e/` packages — only the FuncDecl named `main`, no helper functions
and no function literals. Same `.cur` placeholder dispatch as filetest
crossing-main and `init(cur realm)` in realm packages:

```go
package main
import "gno.land/r/foo"

func main(cur realm) {
    foo.Bar(cross(cur), "arg")
}
```

Semantic equivalence with bare-cross main: the `.cur` synthetic
invocation does **not** set WithCross on the synthetic CallExpr, so
the main frame is treated identically by the frame walk in `GetRealm`.
`runtime.{Current,Previous}Realm()` and `cur.Previous()` produce the
same values either way.

Touches: `preprocess.go` `crossingAllowed`, `op_call.go`
`doOpEnterCrossing` (matching runtime carve-out), `gno.land/pkg/sdk/vm/
keeper.go` `RunMain` → `RunMainMaybeCrossing`.

---

## 13. `buildOriginRealm` propagates `/e/` pkgPath for MsgRun

`buildOriginRealm` (`uverse.go`) builds the "origin realm" used as the
`prev` of a fresh cur minted by `installCrossingCur`'s bare-cross
fallback and the `.cur` placeholder swap path. Previously it hardcoded
`pkgPath=""`, matching the MsgCall-direct-from-EOA shape — but wrong
for MsgRun, where the calling code lives in `/e/<addr>/run`.

Fix: consult `m.Frames[0].LastPackage.PkgPath` and propagate only when
`IsEphemeralPath` returns true. MsgCall (synthetic main with empty
pkgPath), QueryEval (target /r/), and AddPkg (target /r/) are
unaffected. MsgRun's origin realm now has `pkgPath="/e/<addr>/run"`.

Closes the divergence between `cur.Previous()` and
`runtime.PreviousRealm()` for MsgRun callees — both now agree that
the caller is the `/e/` ephemeral run script, not a bare EOA.
`cur.Previous().IsUserCall()` no longer returns true under MsgRun
(it would for MsgCall direct from EOA), which is the right answer
per `CLAUDE.md`'s payment-guard guidance.

---

## 14. Reverted: CLOSE_LAUNDER preserve-/p/-stamp on Copy

For a window, `StructValue.Copy` and `ArrayValue.Copy` preserved the
source PkgID for `/r/` AND `/p/` (excluding stdlib) — intended to close
a deferred-pointer-copy launder attack where a `/p/`-stamped value
passed by pointer across a cross-call could be adopted into the
caller's realm. That broke legitimate `/p/`-helper-copies-default
patterns (e.g. `gno.land/p/jeronimoalbi/datasource.NewQuery`'s
`q := defaultQuery`).

Reverted in `c71fb7bfc`: gate back to `IsRealmPkg()` — preserve only
`/r/` stamps. `/p/` copies drop the stamp and inherit the caller's
realm via the fresh `NewStruct`/`NewListArray` stamp.

The original close-launder design was sequenced with Phase 3
frozen-realm machinery (`Realm.Frozen` + `DidUpdate`/
`assignNewObjectID` panics on writes to frozen realms). Phase 3 was
reverted at `7f2efb4f7`; without it, preserve-/p/-stamp alone is the
worst of both worlds — breaks legit patterns AND doesn't fully close
laundering at persist time. Until Phase 3 returns, keep both reverted.

The previously-recommended workaround in §9 ("construct fresh inside
the constructor") is no longer load-bearing — `/p/`-helper-copies-
default works again.

---

## 15. `/p/` package state is immutable post-init via DidUpdate gate

`Realm.DidUpdate` panics on writes to real (`NewTime > 0`)
`/p/`-stamped objects in `StageRun` (covers MsgCall, MsgRun,
QueryEval). Stdlib is exempt (legit `fmt.Println` etc. dispatch
through the same code path).

The closed attack: `(&pkg.PInitData).PMethod()` — a `/r/`-realm caller
takes the address of a `/p/`-package-init global, invokes a
`/p/`-method on it. The borrow rule at `PushFrameCall` Layer 2 shifts
`m.Realm` to the receiver's package realm. For `/p/` packages
(`pv.Realm == nil` post Phase-3-revert), the shift sets `m.Realm = nil`.
The write inside the method body fires `Assign2` with `rlm == nil`,
which previously short-circuited DidUpdate. Mutation succeeded silently
in-tx (didn't persist across tx, but did affect later same-tx reads).

The gate, in `realm.go` DidUpdate:

```go
func (rlm *Realm) DidUpdate(m *Machine, po, xo, co Object) {
    if rlm == nil {
        if m != nil && m.Stage == StageRun && po != nil && po.GetIsReal() {
            pid := po.GetObjectID().PkgID
            if pid.IsImmutablePkg() && !pid.IsStdlibPkg() {
                panic(fmt.Sprintf("cannot mutate %s: package is immutable post-init", path))
            }
        }
        return
    }
    // ...
}
```

`Assign2` was updated to always invoke `DidUpdate` (dropping its
rlm-nil short-circuit), and `Assign2`/`DidUpdate`/`GetPointerAtIndex`
all gained `*Machine` first param so the gate can read `m.Stage`.

**Preserved (unaffected):**
- `/r/`-realm internal mutations (m.Realm is /r/, gate's nil-rlm branch
  doesn't fire).
- `/p/`-helper methods on `/r/`-stored receivers (the list.Set /
  avl.Tree pattern): receiver's PkgID is /r/, gate doesn't fire.
- Stdlib mutations (IsStdlibPkg exempt).
- `/p/`-init writes (StageAdd, gate doesn't fire).

**Broken (intentional):**
- `pkg.Singleton.Set(x)` patterns from `/r/` callers — the quirky
  "/p/-singleton-as-scratchpad" usage where the mutation worked in-tx
  but never persisted. `object_pointer_pure.txtar` was pinning this;
  test now asserts the panic.

For `/p/`-helper APIs that previously relied on the singleton pattern,
the migration is to either (a) move the state to a `/r/`-realm the
caller owns, or (b) restructure the API to take a `*T` from the caller
rather than mutating a package singleton.

---

## 16. Two-step migration recipe: bare `cross` → `cross1` → `cross(rlm)`

The gno 0.9 canonical form is `fn(cross(rlm), args...)` where `rlm` is
the in-scope realm value. Codebases originally written against the
bare-`cross` sentinel (`fn(cross, args...)`) migrate in **two
mechanical-then-semantic steps**:

**Step 1 — mechanical: `cross` → `cross1`.**

A pure rename. `cross1` is a uverse-defined legacy sentinel (see
`uverse.go` `def("cross1", undefined)`, `preprocess.go`
`case Name("cross1")`) that lowers to **exactly the same AST shape as
the compiler-synthesized `.origin`**:
- `n.SetWithCross()` (the call is a crossing call)
- `n.Args[0] = constNil(nx)` (Args[0] is nil, not a realm value)

Runtime takes the `Args[0]==nil → callingCurOrOrigin →
buildOriginRealm` path — the same path bare `cross` used to take.
Behavior is preserved, so this step is safe to bulk-apply with sed/awk
across the tree without per-site analysis.

This unblocks deleting the bare-`cross` preprocessor branch from gnovm
(see commit `86b1bd0cf`) while leaving codebases temporarily compilable.

**Step 2 — semantic: `cross1` → `cross(rlm)`.**

This step requires per-call-site judgment: identify which realm value
is in scope (typically `cur` from the enclosing crossing function, but
sometimes a captured realm passed via parameter, or a freshly-minted
`testing.NewUserRealm(...)`), and replace `cross1` with `cross(rlm)`.

Lowering changes: `cross(rlm)` takes the `else` branch in
`installCrossingCur` and uses `*argtv` directly as the new cur's prev,
**bypassing** `callingCurOrOrigin`/`buildOriginRealm`. The new cur's
prev is now a static, statically-validated value rather than something
dynamically reconstructed from `OriginCaller`.

**Why two steps and not one:** the `cross` → `cross(cur)` rewrite can't
be applied mechanically — `cur` isn't always the right realm to thread,
and outside a crossing function `cur` doesn't even exist. Sites that
look obvious (`Foo(cross, x)` inside `func Bar(cur realm)`) usually do
become `Foo(cross(cur), x)`, but sites inside non-crossing helpers, in
test scopes that called `testing.SetRealm`, or in MsgRun `func main()`
without `(cur realm)` all need different fixes. The `cross1`
intermediate lets the mechanical sweep land first (compile-clean,
preserves behavior) and the semantic threading happen at a human pace.

**When you cannot finish Step 2:** leave `cross1` in place. The
sentinel is intentionally preserved in uverse for exactly this case —
it isn't a deprecation that will break, it's a long-tail compatibility
shim. The runtime path (`callingCurOrOrigin → buildOriginRealm`) is
the same path the compiler-synthesized `.origin` lowering uses for
MsgCall chain roots, so `cross1` will keep working as long as that
synthesis does.

See `gnovm/tests/files/zrealm_cross1_legacy.gno` for a minimal example
asserting that `cross1` and `cross(cur)` produce identical
`cur.PkgPath()` from the callee — useful as a regression guard if the
two lowerings ever diverge.

---

## Appendix: open questions

- **How best to mint a fresh cur after `SetRealm(NewUserRealm)`?**
  See (2). The dynamic-origin path in `buildOriginRealm` is reachable
  from user code via `cross1` (the legacy migration sentinel — see
  §16) and from compiler-synthesized `.origin` (MsgCall chain root).
  `cross1` is intentionally available as a long-tail shim for this
  test pattern, so the gap is covered. A named user-facing primitive
  (e.g. `cross.fromOrigin()`) or refactoring SUTs to take caller
  address explicitly would still be cleaner — `cross1` reads as a
  migration tool, not a stable API.
