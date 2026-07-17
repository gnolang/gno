# Allow indirect cur-call through a local func variable

## Context

`gnovm/pkg/gnolang/preprocess.go` performs a best-effort same-realm check when it sees a call site of the form `f(cur, ...)` — a non-crossing call of a crossing function. The intent is to reject `f(cur, ...)` at compile time when `f` resolves to a function declared in a different realm, since crossing into an external realm requires the explicit `cross(rlm)` form.

The check was originally implemented as:

```go
ftv, err := tryEvalStatic(store, ctxpn, last, n.Func)
if err == nil {
    // This is fine; e.g. somefunc()(cur,...)
} else if ftv.IsUndefined() {
    // Interface... what can we do?
} else {
    fpp := ftv.GetUnboundFunc().PkgPath // <- nil deref
    ...
}
```

`tryEvalStatic` evaluates the expression on a throwaway machine. When `n.Func` is a **local variable** holding a function value (e.g. `p := myHandler; p(cur, …)`), the throwaway machine cannot resolve the variable's runtime value — only its type. The internal evaluation panics, the deferred `recover()` returns `err != nil`, and the returned `TypedValue` ends up with `T` populated (the function type) but `V == nil`.

`IsUndefined()` only checks `T == nil`, so the value fell through to the `else` branch, where `GetUnboundFunc()` returned `nil` and the `.PkgPath` selector panicked. Legitimate same-realm indirection through a local variable was rejected at preprocess time with a misleading diagnostic.

## Decision

**Adopt the fix that landed on `master` in #5722, and reduce this PR to regression tests.**

While this PR was open, #5722 ("handle typed-nil func value in preprocess and vm/qfuncs") fixed the same nil dereference independently, guarding on the *result* of `GetUnboundFunc()` rather than on `ftv.V`:

```go
} else if fv := ftv.GetUnboundFunc(); fv != nil {
    // fv == nil: typed-nil crossing func (e.g. `var f func(cur realm); f(cur)`)
    // or a lazy interface bind (no concrete func until call time);
    // fall through, the runtime check covers both.
    if fv.PkgPath != ctxpn.PkgPath {
        panic(...)
    }
}
```

This PR's original guard was `else if ftv.V == nil { /* skip */ }`. Both avoid the crash for a local func variable, but master's guard is a **strict superset**. `GetUnboundFunc()` returns `nil` in two distinct cases (see its doc comment in `values.go`):

1. `tv.V == nil` — a typed-nil func. Covered by both guards.
2. `tv.V` is a `*BoundMethodValue` whose `Func` is `nil` — a lazy interface bind, where the concrete func does not exist until call time. Covered **only** by master's guard; this PR's `ftv.V == nil` test is false here, so control would still reach `GetUnboundFunc().PkgPath` and nil-deref.

Since master's version fixes everything this PR fixed plus one case it missed, the merge takes master's hunk verbatim and drops this PR's `ftv.V == nil` branch. `preprocess.go` is now byte-identical to master; this PR contributes **no production code**.

The runtime remains the authoritative enforcement point (`machine.go`, the `IsCrossing` path in `PushFrameCall`):

```go
if fv.IsCrossing() {
    if m.Realm != pv.Realm {
        panic(fmt.Sprintf(
            "cannot cur-call to external realm function %s.%v from %s", ...))
    }
}
```

The static check is a compile-time best-effort, not the sole line of defence.

## Alternatives considered

1. **Keep this PR's `ftv.V == nil` guard.** Rejected: it is a subset of master's fix and would leave the lazy-interface-bind case nil-dereferencing. Re-introducing it over master's guard would be a regression.

2. **Close this PR as superseded by #5722.** Reasonable for the code, but discards the regression tests. #5722 added `func31.gno`/`func32.gno`, which cover the *typed-nil* func case (`var f func(cur realm); f(cur)` → runtime nil-call error). Neither covers a **non-nil** func variable, and neither covers the local-variable route into the runtime realm check. Keeping the tests preserves coverage that would otherwise be lost.

3. **Fail at compile time when the static check is inconclusive.** Rejecting `p(cur)` whenever `p` is a local variable would force `p(cross(cur))` at every indirection. Cleaner discoverability, but breaks legitimate same-realm patterns that run fine. Discarded — too restrictive.

4. **Make `tryEvalStatic` not partially populate `TypedValue` on panic.** Cleaner API, but invasive (touches every caller) and broader than the bug. Discarded.

## Consequences

- **No production-code change.** After merging master, `preprocess.go` matches master exactly. This PR is now tests + this ADR.
- Two regression filetests are retained, both passing against master's fix:
  - `zrealm_curcall_indirect_local.gno` — same-realm indirection through a local variable (`p := myHandler; p(cur)`) must succeed. No existing test covers a non-nil func variable in a cur-call.
  - `zrealm_curcall_indirect_external.gno` — cross-realm indirection through a local variable (`p := tests.IncCounter; p(cur)`) must be rejected at **runtime**, with no `file:line` prefix on the message (confirming it comes from `machine.go`, not preprocess).
- These complement, rather than duplicate, the six existing filetests that assert the same runtime message (`zrealm_crossrealm15b/16b/17b/17c/18b/18c.gno`): those reach the runtime check through **interface method dispatch**, not through a plain local func variable.
- Cross-realm indirection via a local variable is diagnosed at run time rather than compile time. The message text is identical, so the observable difference is only *when* it fires.

## Scope notes

The original bug was preexisting on `master` and independent of the `cur realm` capability work in #5669. Following the merge of #5722, this PR's remaining value is regression coverage of the local-variable indirection route.
