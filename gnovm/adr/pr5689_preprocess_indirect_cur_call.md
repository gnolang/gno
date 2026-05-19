# Allow indirect cur-call through a local func variable

## Context

`gnovm/pkg/gnolang/preprocess.go` performs a best-effort same-realm check when it sees a call site of the form `f(cur, ...)` — a non-crossing call of a crossing function. The intent is to reject `f(cur, ...)` at compile time when `f` resolves to a function declared in a different realm, since crossing into an external realm requires the explicit `cross` keyword (or `cross2(rlm)` in the new convention).

The check is implemented around the existing `tryEvalStatic` helper:

```go
ftv, err := tryEvalStatic(store, ctxpn, last, n.Func)
if err == nil {
    // This is fine; e.g. somefunc()(cur,...)
} else if ftv.IsUndefined() {
    // Interface... what can we do?
} else {
    fpp := ftv.GetUnboundFunc().PkgPath
    if fpp != ctxpn.PkgPath {
        panic(fmt.Sprintf("cannot cur-call to external realm function %s.%v from %v",
            fpp, n.Func, ctxpn.PkgPath))
    }
}
```

`tryEvalStatic` evaluates the expression on a throwaway machine. When `n.Func` is a **local variable** holding a function value (e.g. `p := myHandler; p(cur, …)`), the throwaway machine cannot resolve the variable's runtime value — only its type. The internal evaluation panics, the deferred `recover()` returns `err != nil`, and the returned `TypedValue` ends up with `T` populated (the function type) but `V == nil`.

`IsUndefined()` only checks `T == nil`, so the value falls through to the `else` branch, which calls `GetUnboundFunc()` and panics:

```
expected function or bound method but got <nil>
```

This panic surfaces back as a preprocess-time error that points at the call site `p(cur)` with a misleading "got <nil>" diagnostic. The error happens regardless of whether the underlying function is same-realm or cross-realm — i.e., legitimate local-variable indirection is rejected at compile time.

## Decision

Skip the static check when the resolved `TypedValue` has no value bound, and let the runtime enforce the same-realm invariant.

```go
} else if ftv.V == nil {
    // Local variable holding a func value: static eval couldn't
    // bind the value (only the type), so defer the same-realm
    // check to runtime.
} else {
    fpp := ftv.GetUnboundFunc().PkgPath
    ...
}
```

The runtime already enforces the same constraint authoritatively in `machine.go` (around the IsCrossing path in `PushFrameCall`):

```go
if fv.IsCrossing() {
    if m.Realm != pv.Realm {
        panic(fmt.Sprintf(
            "cannot cur-call to external realm function %s.%v from %s", ...))
    }
}
```

The static check was a compile-time best-effort, not the sole line of defence. Six existing filetests already cover the runtime rejection path: `zrealm_crossrealm15b.gno`, `zrealm_crossrealm16b.gno`, `zrealm_crossrealm17b.gno`, `zrealm_crossrealm17c.gno`, `zrealm_crossrealm18b.gno`, `zrealm_crossrealm18c.gno`.

## Alternatives considered

1. **Fail at compile time with a clear message when the static check is inconclusive.** Rejecting `p(cur)` whenever `p` is a local variable would force users to write `p(cross)` or `p(cross2(cur))`. Cleaner discoverability, but breaks legitimate same-realm indirection patterns that work fine at runtime. Discarded — too restrictive.

2. **Auto-rewrite `p(cur)` to `p(cross2(cur))` in the preprocessor when `p` has a crossing function type.** Saves the user a syntactic detour but adds implicit behaviour, and the runtime check is already enough to keep the system safe. Discarded — favours convention over magic.

3. **Make `tryEvalStatic` not partially populate `TypedValue` on panic.** Cleaner from an API standpoint, but invasive (touches every caller of `tryEvalStatic`) and doesn't fix the underlying assumption that the static check can resolve any non-undefined `ftv`. Discarded — broader than the bug.

## Consequences

- Code that previously failed to preprocess with `expected function or bound method but got <nil>` now compiles and runs when the indirection is same-realm.
- Cross-realm indirection through a local variable now panics at **runtime** with `cannot cur-call to external realm function …`, which is the same message the static check used to produce. Externally observable behaviour for legitimate code is unchanged; for cross-realm attempts the diagnostic moves from compile time to run time but the message text is identical.
- No regressions in the existing `TestFiles` suite (~75 filetests).
- Two new regression filetests added:
  - `zrealm_curcall_indirect_local.gno` — same-realm indirection must succeed.
  - `zrealm_curcall_indirect_external.gno` — cross-realm indirection must be rejected at runtime.

## Scope notes

This bug exists on `master` and is independent of PR #5669 (the `cur realm` capability work). It is a preexisting ergonomic defect of the preprocess-time same-realm check. The fix is intentionally minimal — four lines plus regression tests — and does not interact with the crossing-function surface introduced by #5669.
