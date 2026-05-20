# cross2(rlm): explicit cross-call form

## Context

In Gno 0.9, a cross-realm call is spelled `f(cross, args...)` where the
bare identifier `cross` is a preprocessor-recognized sentinel registered
in uverse as a `realm`-typed value. The preprocessor sees `cross` at
`Args[0]` of a call to a crossing function, sets `WithCross` on the
`CallExpr`, and replaces `Args[0]` with `constNil`. At runtime,
`installCrossingCur` mints a fresh `cur` whose prev comes from
`callingCurOrOrigin()` — a frame walk that locates the topmost crossing
frame's `Cur`.

The bare `cross` form leaves the realm that the new cur is "crossing
from" implicit. Reading a call site, you have to know that `cross` here
means "use whatever cur is in scope in the enclosing crossing function."
This becomes confusing once the codebase threads `cur realm` (or `rlm
realm`, in the `(_ int, rlm realm, ...)` migration convention) through
many functions: the author wrote `f(cross, ...)` from inside a function
that has a specific `rlm` parameter, but the source doesn't say which
one.

`cross2(rlm)` is the explicit form. The argument names the realm value
the caller is "crossing from," and the runtime verifies that value is
the current frame's `Cur` via `IsCurrent`-strict before using it as the
prev for the new cur.

## Decision

Add `cross2(rlm)` as a parallel construct to bare `cross`. Both coexist
during the gno 0.9 transition. The transient name (`cross2`, not `cross`
overloaded) sidesteps the Go-typechecker constraint that `cross` is
typed as `var cross realm` and therefore cannot also be a callable.

### Syntax

```gno
// Bare cross — implicit:
f(cross, args...)

// cross2(rlm) — explicit:
f(cross2(rlm), args...)
```

The argument to `cross2` must be a bare `NameExpr` (an identifier, not
an arbitrary expression). The identifier must be realm-typed. **No
lexical check on where the realm came from** — the runtime
IsCurrent-strict check is the safety. This allows the threading
convention used throughout the migration:

```gno
// Non-crossing helper takes rlm via (_ int, rlm realm, ...) shape.
// cross2(rlm) works here because the runtime check, not a lexical
// rule, validates rlm.
func helper(_ int, rlm realm) {
    callee(cross2(rlm))
}

func main(cur realm) {
    helper(0, cur)
}
```

Stale rlm (captured in a different frame, threaded through cross-call
chain where the outer frame is no longer topmost, or laundered through
a value-receiver copy) panics at runtime when `cross2` runs.

### Semantics

When `IsCurrent`-strict passes on the resolved rlm, `cross2(rlm)` is
exactly equivalent to bare `cross`:

- A fresh `cur` is minted for the callee's frame.
- The new cur's `prev` is the resolved rlm value.
- The callee's body runs with the new cur in scope.

When `IsCurrent`-strict fails (rlm is stale, came from a sibling frame,
or has been laundered through a value-receiver copy), `cross2(rlm)`
panics with `cross2: rlm is not the current cur (...)`. Bare `cross`
has no equivalent runtime check — it can't, because it doesn't take an
argument to check.

### Why "strict" IsCurrent

The user-facing `.grealm.IsCurrent` method has a fallback path for the
value-receiver-method-dispatch case (uverse.go:1402-1411): when the
outer HIV pointer is stripped by struct copy, it falls back to
matching by `(addr, pkgPath, prev.HIV)`. This is necessary for the
user-facing API because value-receiver method dispatch is legitimate
Gno code, and `someRealm.IsCurrent()` should give a useful answer
even after a struct copy.

For `cross2`, the trust boundary is stricter: the rlm value reaches
`installCrossingCur` via `Block.GetPointerTo` resolving a NameExpr
path, which never goes through value-receiver method dispatch. A
legitimate rlm always retains its HIV. Rejecting the HIV-less fallback
closes a class of attacks where an attacker laundered a cur through a
value-receiver copy and tried to pass the result to `cross2`.

### Allowed in tests

`cross2(rlm)` follows the existing `crossingAllowed` carveout at
`preprocess.go:4455-4461`: allowed in realm packages AND in any
`*_test.gno` file (regardless of package). Tests in /p/ packages can
use `cross2(rlm)` to make their cross-call intent explicit.

## Implementation

The design routes cross2 through Gno's normal eval/native-call
machinery: cross2 is a real runtime function that returns its
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

### Files changed

1. **`gnovm/pkg/gnolang/gotypecheck.go`** — Add `func cross2(rlm realm)
   realm { return rlm }` shim to the gno 0.9 per-package
   `.gnobuiltins.gno` (alongside `var cross realm`). The 0.0 shim
   intentionally does NOT declare `cross2` — `cross2` is a 0.9-only
   construct.

2. **`gnovm/pkg/gnolang/uverse.go`** — Register `cross2` as a generic
   native (`defNative("cross2", ...)` with generic X param/result,
   mirroring `_cross_gno0p0`'s shape). The native body validates
   `realmIsCurrentStrict` on the argument and panics if false;
   otherwise pushes the argument unchanged. Also factor the existing
   `IsCurrent` method body into a shared `realmIsCurrentOnMachine`
   helper and add `realmIsCurrentStrict` (rejects the HIV-less
   fallback path — the strict check is the trust boundary).

3. **`gnovm/pkg/gnolang/nodes.go`** — Extend `isLikeWithCross` to
   accept a `*CallExpr` at `Args[0]` whose `Func` is the const-
   evaluated uverse `cross2` native, in addition to the existing
   `*NameExpr` named `cur`/`cross`.

4. **`gnovm/pkg/gnolang/preprocess.go`** — Three checks at preprocess:
   - **Inner `cross2` TRANS_LEAVE** (uverse-func switch, alongside
     `_cross_gno0p0`, `cross`, `crossing`): validate the call shape
     is `cross2(<NameExpr>)` with exactly one bare identifier
     argument. Verify the parent context is `Args[0]` of a
     crossing-function call (ftype == TRANS_CALL_ARG && index == 0,
     parent `*CallExpr` whose Func resolves to a crossing
     `*FuncType`). Reject stray cross2 usage (`x := cross2(rlm)`,
     cross2 at non-zero arg index, cross2 at Args[0] of a
     non-crossing function).
   - **Outer crossing-CallExpr's `Args[0]` validity check**: accept
     `*CallExpr` whose `Func` is the const-evaluated `cross2`. Call
     `n.SetWithCross()`. **Leave Args[0] in place** — at runtime
     the inner cross2 native runs and pushes the validated rlm
     onto the stack at Args[0]'s slot.
   - **Defer rejection**: at `LEAVE_CALL_EXPR_END_CHECK_CROSSING`,
     panic if `n.WithCross && ftype == TRANS_DEFER_CALL`.
     `defer f(cross, ...)` and `defer f(cross2(rlm), ...)` would
     crash at runtime because `doOpReturnCallDefers` bypasses
     `doOpPrecall` and never invokes `installCrossingCur`. The
     panic message points at the closure-wrapper workaround:
     `defer func() { f(cross, args...) }()`.

5. **`gnovm/pkg/gnolang/op_call.go`** — `installCrossingCur` branches
   on whether `m.PeekValue(cx.NumArgs)` is undefined:
   - **Undefined** (bare cross — Args[0] is `constNil`): mint cur
     with `prev = m.callingCurOrOrigin()`. Existing semantics.
   - **Realm value** (cross2 — Args[0] is the inner cross2 CallExpr
     that evaluated to rlm): cross2 already validated IsCurrent-strict,
     use `prev = *argtv` directly. No second check needed.

No lexical check on where rlm came from. cross2 works with
parameter-rlm, closure-captured rlm, struct-field rlm — any
realm-typed identifier that can be evaluated by the normal eval
machinery. The runtime IsCurrent-strict check (in cross2's body)
is the sole safety; if rlm is stale (sibling frame, A→B→A
re-entry, value-receiver-laundered), cross2 panics with a clear
message at the cross2 call site.

### Why this design (and not path-stashing)

An earlier implementation stashed the inner NameExpr's resolved
`ValuePath` on a new `CallExpr.CrossArgPath` field (serialized as
proto field 7) and re-resolved it at install time via
`Block.GetPointerTo`. This had two flaws:

1. **Closure-captured names crashed with Go nil-deref.** Closure
   captures live via heap-item indirection, not the parent-block
   chain that GetPointerTo walks. A path computed at preprocess
   relative to the closure body didn't resolve at install time.
2. **Wire-format change**. Adding `CrossArgPath` to `CallExpr`
   required a proto field — irreversible once deployed.

The current design (cross2 as a normal native) sidesteps both:
no path resolution at install time (the runtime evaluates the
inner cross2 call through the normal machinery, which handles
heap-item indirection transparently), and no AST shape change
(cross2 is just another uverse function).

## Migration rule

Existing bare `cross` callsites work unchanged. New code MAY use
`cross2(rlm)` for explicit notation. There is no automated transformer
because the migration is not just textual — converting `f(cross, ...)`
to `f(cross2(rlm), ...)` requires identifying the right rlm in scope,
and many existing call sites are in non-crossing helpers that would
need a new `cur` parameter threaded through. That's a per-call
judgment call.

Recommendation: prefer `cross2(rlm)` in new code where `rlm` is
already in scope. Don't refactor existing `cross` to `cross2(rlm)`
unless you're already touching the file for an unrelated reason.

## Future work

When the bare `cross` form is eventually removed, `cross2` will be
renamed back to `cross`. That rename is a textual change plus an AST
attribute (`CrossArgPath`) that may or may not be kept under the new
name. Cost is one commit, no semantic change. The transient name
"cross2" signals to readers that this is the migration form, not the
final shape.

## References

- Bare `cross` introduction: `adr/pr4264_lint_transpile.md`.
- The realm capability-token model and IsCurrent semantics:
  `examples/gno.land/p/test/seal/filetests/z_seal_*_filetest.gno` and
  the project's runtime-to-cur migration notes.
- Filetests verifying cross2 behavior:
  - `gnovm/tests/files/zrealm_cross2_basic.gno` — happy path; bare
    `cross` and `cross2(cur)` produce identical output.
  - `gnovm/tests/files/zrealm_cross2_rlm.gno` — cross2 in a
    `(_ int, rlm realm)` non-crossing helper (the threading
    convention; demonstrates the lexical-check relaxation).
  - `gnovm/tests/files/zrealm_cross2_crosspkg.gno` — cross2 driving
    a cross-realm call into another package; output identical to
    bare cross.
  - `gnovm/tests/files/zrealm_cross2_notname.gno` — preprocess
    rejects non-NameExpr arguments (e.g. function call).
  - `gnovm/tests/files/zrealm_cross2_previous.gno` — preprocess
    rejects method-call expressions like `cur.Previous()`.
  - `gnovm/tests/files/zrealm_cross2_noarg.gno` — preprocess +
    typecheck reject `cross2()` with no args.
  - `gnovm/tests/files/zrealm_cross2_extra.gno` — preprocess +
    typecheck reject `cross2(rlm, x)` with extra args.
  - `gnovm/tests/files/zrealm_cross2_stalerlm.gno` — runtime
    IsCurrent-strict catches a stale realm threaded through a
    cross-call chain where the outer frame is no longer topmost.
  - `gnovm/tests/files/zrealm_cross2_iface.gno` — cross2(cur)
    through interface dispatch (the primary cur-leak surface
    that motivated the whole branch).
  - `gnovm/tests/files/zrealm_cross2_ifacestale.gno` — stale rlm
    via interface dispatch — the realistic threat shape — caught
    at runtime by IsCurrent-strict.
  - `gnovm/tests/files/zrealm_cross2_closurecap.gno` — closure
    captures a realm-typed local; the closure's invocation from
    the same crossing frame succeeds. (Previously crashed under
    the path-stashing design.)
  - `gnovm/tests/files/zrealm_cross2_closuresib.gno` — closure
    captures rlm in main, is passed via cross-call into another
    crossing function and invoked there; cross2 panics on the
    runtime IsCurrent-strict check because main's frame is no
    longer topmost.
  - `gnovm/tests/files/zrealm_cross2_stray.gno` — preprocess
    rejects standalone `x := cross2(cur)` (cross2 must appear at
    Args[0] of a crossing call).
  - `gnovm/tests/files/zrealm_cross2_nonxfn.gno` — preprocess
    rejects `cross2(cur)` at Args[0] of a non-crossing function.
  - `gnovm/tests/files/zrealm_cross2_defer.gno` /
    `zrealm_cross2_defercross.gno` — preprocess rejects
    `defer f(cross2(rlm), ...)` and `defer f(cross, ...)`. The
    deferred-call dispatch path bypasses `installCrossingCur`
    and would crash at runtime; the panic message recommends
    the closure-wrapper workaround.
  - `gnovm/tests/files/zrealm_cross2_deferwrap.gno` — the
    closure-wrapper workaround for deferring a crossing call:
    `defer func() { f(cross, args) }()`.

## Known follow-ups (out of scope for this implementation)

- **`fr.Cur` not set on deferred frame**. `doOpReturnCallDefers`
  pushes the deferred callee's frame but doesn't run the
  inherit-from-block code that op_call.go does for non-crossing
  calls of crossing functions. Today this is masked because
  frame walks skip `fr.Cur.T == nil` frames and find the
  deferring function's frame underneath (whose Cur happens to
  match the deferred callee's `cur` parameter). A future-code
  reader of `m.LastFrame().Cur` directly would see nil. ~5-10
  LOC fix; deferred to a follow-up.

- **cross2(rlm) inside a closure-wrapper deferred call** — the
  `defer func() { f(cross2(rlm)) }()` pattern: rlm is captured
  by the closure at defer time; at defer-fire time rlm is read
  from the closure's captures. If rlm was the deferring
  function's cur (typical case), this works. The runtime
  IsCurrent-strict check is the safety.
