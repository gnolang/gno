# cross(rlm): explicit cross-call form

> Historical note: this design originally landed as `cross2(rlm)` while
> a legacy bare-`cross` sentinel was still in use. After the sentinel
> was removed, `cross2` was renamed to `cross`. This document describes
> the resulting canonical form.

## Context

In Gno, a cross-realm call is spelled `f(cross(rlm), args...)` where
`cross` is a uverse-native function that returns its `realm`-typed
argument unchanged after validating that the value is the current
frame's `cur`. The preprocessor recognizes the inner `cross(rlm)` call
at `Args[0]` of a crossing-function call, sets `WithCross` on the outer
`CallExpr`, and leaves the inner call in place. At runtime, the inner
`cross` native runs first, validates, and pushes the realm value back
onto the stack at `Args[0]`'s slot; the outer crossing call's
`installCrossingCur` peeks that value and uses it as the new cur's
`prev`.

The explicit form names the realm value the caller is "crossing from."
The runtime IsCurrent-strict check ensures that value is the current
frame's `Cur` before the new cur is minted — catching stale rlm
(captured in another frame, threaded through a cross-call chain where
the outer frame is no longer topmost, or laundered through a
value-receiver copy).

## Decision

`cross(rlm)` is the canonical and only cross-call form. The argument
must be a bare `NameExpr` (an identifier, not an arbitrary expression).
The identifier must be realm-typed. **No lexical check on where the
realm came from** — the runtime IsCurrent-strict check is the safety.
This allows the threading convention used throughout the codebase:

```gno
// Non-crossing helper takes rlm via (_ int, rlm realm, ...) shape.
// cross(rlm) works here because the runtime check, not a lexical
// rule, validates rlm.
func helper(_ int, rlm realm) {
    callee(cross(rlm))
}

func main(cur realm) {
    helper(0, cur)
}
```

### Syntax

```gno
f(cross(rlm), args...)
```

### Semantics

When `IsCurrent`-strict passes on the resolved rlm:

- A fresh `cur` is minted for the callee's frame.
- The new cur's `prev` is the resolved rlm value.
- The callee's body runs with the new cur in scope.

When `IsCurrent`-strict fails (rlm is stale, came from a sibling frame,
or has been laundered through a value-receiver copy), `cross(rlm)`
panics with `cross: rlm is not the current cur (stale capture or
sibling frame)`.

### Why "strict" IsCurrent

The user-facing `.grealm.IsCurrent` method has a fallback path for the
value-receiver-method-dispatch case (uverse.go): when the outer HIV
pointer is stripped by struct copy, it falls back to matching by
`(addr, pkgPath, prev.HIV)`. This is necessary for the user-facing API
because value-receiver method dispatch is legitimate Gno code, and
`someRealm.IsCurrent()` should give a useful answer even after a
struct copy.

For `cross`, the trust boundary is stricter: the rlm value reaches
`installCrossingCur` via `Block.GetPointerTo` resolving a NameExpr
path, which never goes through value-receiver method dispatch. A
legitimate rlm always retains its HIV. Rejecting the HIV-less fallback
closes a class of attacks where an attacker laundered a cur through a
value-receiver copy and tried to pass the result to `cross`.

### Allowed in tests

`cross(rlm)` follows the existing `crossingAllowed` carveout in
`preprocess.go`: allowed in realm packages AND in any `*_test.gno`
file (regardless of package). Tests in /p/ packages can use
`cross(rlm)` to make their cross-call intent explicit.

## Implementation

The design routes `cross` through Gno's normal eval/native-call
machinery: `cross` is a real runtime function that returns its
argument unchanged (after IsCurrent-strict validation), and the
outer crossing call's `installCrossingCur` peeks the evaluated rlm
on the value stack and uses it as the new cur's prev. **No special
AST plumbing, no path resolution, no wire-format addition.** An
earlier path-stashing design (CrossArgPath as a serialized field
on CallExpr) was abandoned because it couldn't handle closure
captures correctly — closure-captured names live via heap-item
indirection, not the parent-block chain that `Block.GetPointerTo`
walks. The normal eval machinery handles heap-item indirection
transparently.

### Files

1. **`gnovm/pkg/gnolang/gotypecheck.go`** — `.gnobuiltins.gno`
   per-package shim: `func cross(rlm realm) realm { return rlm }`.

2. **`gnovm/pkg/gnolang/uverse.go`** — `defNative("cross", ...)` with
   generic X param/result. The native body validates
   `realmIsCurrentStrict` on the argument and panics if false;
   otherwise pushes the argument unchanged. The shared
   `realmIsCurrentOnMachine` helper backs both this strict check
   and `.grealm.IsCurrent`.

3. **`gnovm/pkg/gnolang/nodes.go`** — `isLikeWithCross` accepts a
   `*CallExpr` at `Args[0]` whose `Func` is the const-evaluated
   uverse `cross` native, plus `*NameExpr` named `cur` or `.origin`.

4. **`gnovm/pkg/gnolang/preprocess.go`** — Three checks at preprocess:
   - **Inner `cross` TRANS_LEAVE** (uverse-func switch, alongside
     `crossing`, `attach`): validate the call shape is
     `cross(<NameExpr>)` with exactly one bare identifier argument.
     Verify the parent context is `Args[0]` of a crossing-function
     call (`ftype == TRANS_CALL_ARG && index == 0`, parent
     `*CallExpr` whose Func resolves to a crossing `*FuncType`).
     Reject stray usage (`x := cross(rlm)`, `cross` at non-zero arg
     index, `cross` at `Args[0]` of a non-crossing function).
   - **Outer crossing-CallExpr's `Args[0]` validity check**: accept
     `*CallExpr` whose `Func` is the const-evaluated `cross`. Call
     `n.SetWithCross()`. **Leave `Args[0]` in place** — at runtime
     the inner `cross` native runs and pushes the validated rlm
     onto the stack at `Args[0]`'s slot.
   - **Defer rejection**: panic if `n.WithCross && ftype ==
     TRANS_DEFER_CALL`. `defer f(cross(rlm), ...)` would crash at
     runtime because `doOpReturnCallDefers` bypasses `doOpPrecall`
     and never invokes `installCrossingCur`. The panic message
     points at the closure-wrapper workaround.

5. **`gnovm/pkg/gnolang/op_call.go`** — `installCrossingCur` branches
   on whether `m.PeekValue(cx.NumArgs)` is undefined:
   - **Undefined** (compiler-synthesized `.origin` — `Args[0]` is
     `constNil`): mint cur with `prev = m.callingCurOrOrigin()`.
   - **Realm value** (`cross(rlm)` — `Args[0]` is the inner `cross`
     CallExpr that evaluated to rlm): `cross` already validated
     IsCurrent-strict, use `prev = *argtv` directly. No second check
     needed.

No lexical check on where rlm came from. `cross(rlm)` works with
parameter-rlm, closure-captured rlm, struct-field rlm — any
realm-typed identifier that can be evaluated by the normal eval
machinery. The runtime IsCurrent-strict check (in `cross`'s body)
is the sole safety; if rlm is stale (sibling frame, A→B→A
re-entry, value-receiver-laundered), `cross` panics with a clear
message at the call site.

### Why this design (and not path-stashing)

An earlier implementation stashed the inner NameExpr's resolved
`ValuePath` on a new `CallExpr.CrossArgPath` field (serialized as
proto field 7) and re-resolved it at install time via
`Block.GetPointerTo`. This had two flaws:

1. **Closure-captured names crashed with Go nil-deref.** Closure
   captures live via heap-item indirection, not the parent-block
   chain that `GetPointerTo` walks. A path computed at preprocess
   relative to the closure body didn't resolve at install time.
2. **Wire-format change.** Adding `CrossArgPath` to `CallExpr`
   required a proto field — irreversible once deployed.

The current design (`cross` as a normal native) sidesteps both:
no path resolution at install time (the runtime evaluates the
inner `cross` call through the normal machinery, which handles
heap-item indirection transparently), and no AST shape change
(`cross` is just another uverse function).

## References

- Realm capability-token model and IsCurrent semantics:
  `examples/gno.land/p/test/seal/filetests/z_seal_*_filetest.gno`.
- Filetests verifying behavior:
  - `gnovm/tests/files/zrealm_cross_basic.gno` — happy path.
  - `gnovm/tests/files/zrealm_cross_rlm.gno` — `cross(rlm)` in a
    `(_ int, rlm realm)` non-crossing helper (the threading
    convention; demonstrates the lexical-check relaxation).
  - `gnovm/tests/files/zrealm_cross_crosspkg.gno` — driving a
    cross-realm call into another package.
  - `gnovm/tests/files/zrealm_cross_notname.gno` — preprocess
    rejects non-NameExpr arguments (e.g. function call).
  - `gnovm/tests/files/zrealm_cross_previous.gno` — preprocess
    rejects method-call expressions like `cur.Previous()`.
  - `gnovm/tests/files/zrealm_cross_noarg.gno` — preprocess +
    typecheck reject `cross()` with no args.
  - `gnovm/tests/files/zrealm_cross_extra.gno` — preprocess +
    typecheck reject `cross(rlm, x)` with extra args.
  - `gnovm/tests/files/zrealm_cross_stalerlm.gno` — runtime
    IsCurrent-strict catches a stale realm threaded through a
    cross-call chain where the outer frame is no longer topmost.
  - `gnovm/tests/files/zrealm_cross_iface.gno` — `cross(cur)`
    through interface dispatch (the primary cur-leak surface
    that motivated the whole branch).
  - `gnovm/tests/files/zrealm_cross_ifacestale.gno` — stale rlm
    via interface dispatch — the realistic threat shape — caught
    at runtime by IsCurrent-strict.
  - `gnovm/tests/files/zrealm_cross_closurecap.gno` — closure
    captures a realm-typed local; invocation from the same
    crossing frame succeeds.
  - `gnovm/tests/files/zrealm_cross_closuresib.gno` — closure
    captures rlm in main, is passed via cross-call into another
    crossing function and invoked there; `cross` panics on the
    runtime IsCurrent-strict check because main's frame is no
    longer topmost.
  - `gnovm/tests/files/zrealm_cross_stray.gno` — preprocess
    rejects standalone `x := cross(cur)` (`cross` must appear
    at `Args[0]` of a crossing call).
  - `gnovm/tests/files/zrealm_cross_nonxfn.gno` — preprocess
    rejects `cross(cur)` at `Args[0]` of a non-crossing function.
  - `gnovm/tests/files/zrealm_cross_defer.gno` /
    `zrealm_cross_defercross.gno` — preprocess rejects
    `defer f(cross(rlm), ...)`. The deferred-call dispatch path
    bypasses `installCrossingCur` and would crash at runtime; the
    panic message recommends the closure-wrapper workaround.
  - `gnovm/tests/files/zrealm_cross_deferwrap.gno` — the
    closure-wrapper workaround for deferring a crossing call:
    `defer func() { f(cross(rlm), args) }()`.

## Known follow-ups

- **`fr.Cur` not set on deferred frame**. `doOpReturnCallDefers`
  pushes the deferred callee's frame but doesn't run the
  inherit-from-block code that `op_call.go` does for non-crossing
  calls of crossing functions. Today this is masked because frame
  walks skip `fr.Cur.T == nil` frames and find the deferring
  function's frame underneath (whose Cur happens to match the
  deferred callee's `cur` parameter). A future reader of
  `m.LastFrame().Cur` directly would see nil. ~5-10 LOC fix.

- **`cross(rlm)` inside a closure-wrapper deferred call** — the
  `defer func() { f(cross(rlm)) }()` pattern: rlm is captured by
  the closure at defer time; at defer-fire time rlm is read from
  the closure's captures. If rlm was the deferring function's
  cur (typical case), this works. The runtime IsCurrent-strict
  check is the safety.
