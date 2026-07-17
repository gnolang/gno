# Make `tryEvalStatic`'s error result meaningful

## Context

`tryEvalStatic` (`gnovm/pkg/gnolang/preprocess.go`) evaluates an expression on a throwaway machine and reports whether it could be resolved statically. Its doc comment says "May fail for a variety of reasons", and it returns `(tv TypedValue, err error)`.

It has exactly one caller: the preprocess-time same-realm check for `f(cur, ...)` call sites (a non-crossing call of a crossing function), which rejects a cur-call into a different realm at compile time.

**Two bugs in this pair cancel each other out.**

### Bug 1 — `err` is always non-nil

The deferred recover has no `if r == nil` guard:

```go
defer func() {
    r := recover()
    if e, ok := r.(error); ok {
        err = e
    } else {
        err = fmt.Errorf("recovered panic with: %v", r)
    }
}()
tv = m.EvalStatic(last, x)
```

When evaluation succeeds there is no panic, so `recover()` returns `nil`. `nil` does not type-assert to `error`, so the `else` branch runs and sets `err` to `recovered panic with: <nil>`. The function therefore reports failure on success. `err == nil` is reachable *only* through the `*ConstExpr` early return.

### Bug 2 — the caller's branch is inverted

```go
ftv, err := tryEvalStatic(store, ctxpn, last, n.Func)
if err == nil {
    // This is fine; e.g. somefunc()(cur,...)
} else if ftv.IsUndefined() {
    ...
```

`somefunc()(cur,...)` is a callee that *cannot* be resolved statically — that is the `err != nil` case. The comment describes the opposite branch from the one it sits on. Read literally, this says "if we resolved the callee, skip the check", which would disable the check entirely.

### Why nothing fails today

Bug 1 makes `err` unconditionally non-nil, so the `err == nil` branch is never taken and bug 2 never fires. Control always reaches the `else if` chain, where the real work keys off the *shape* of `tv` (`IsUndefined()`, `GetUnboundFunc() != nil`) rather than `err`. The check is correct — by accident.

Measured on the full `TestFiles` suite: the `err == nil` branch fires **0 times**. It is dead code.

The failure mode is quiet and asymmetric, which is what makes this worth fixing rather than leaving alone:

| Change | `zrealm_crossrealm17b.gno` |
|---|---|
| master (both bugs) | passes |
| fix bug 1 only | **fails** — check silently skipped, error demoted to runtime |
| fix bug 2 only | **fails** — same |
| fix both | passes (full suite green) |

Anyone who fixes the obviously-wrong `recover()` in isolation silently disables the compile-time cross-realm check. The diagnostic doesn't disappear — it moves from preprocess to runtime — so it is easy to miss in review.

## Decision

Fix both, together.

```go
defer func() {
    r := recover()
    if r == nil {
        return // no panic: evaluation succeeded, leave err nil
    }
    ...
}()
```

```go
if err != nil {
    // Couldn't resolve statically; e.g. somefunc()(cur,...) or a
    // local variable. Defer to the runtime check.
} else if ftv.IsUndefined() {
```

`err` now means what its signature says, and the caller's branch matches its comment.

## Consequences

- **No behaviour change in any exercised case.** The `else if` chain already discriminated on `tv`'s shape, and it reaches the same verdict for every input the suite produces. Full `TestFiles` passes.
- One latent hole closes. A `*ConstExpr` callee (`err == nil` via the early return) currently **skips** the realm check; afterwards it is checked, like any other resolved callee. This never occurs in the suite (0 hits), so it is unobservable today, but the post-fix behaviour is the correct one.
- The runtime check in `machine.go` (the `IsCrossing` path in `PushFrameCall`) remains the authoritative enforcement point. The preprocess check stays best-effort.
- `zrealm_crossrealm17b.gno` is the regression guard: it asserts a *preprocess-time* rejection (its expected error carries a `file:line` prefix). If either bug is reintroduced, the check is skipped, the error is produced by `machine.go` without a prefix, and the test fails.
- No new tests. Behaviour is unchanged, so there is nothing new to assert; the existing test already fails on any regression.

## Alternatives considered

1. **Fix only the `recover()` guard.** The obvious local fix, and wrong: it disables the static check. Rejected — this is exactly the trap the ADR documents.

2. **Drop `err` from the signature and return only `tv`.** Honest about how the caller behaves today (it discriminates on `tv`'s shape). But it discards a genuine signal and forces every future caller to re-derive "did this resolve?" from value shape. Rejected — fixes the symptom, keeps the confusion.

3. **Leave it alone; it works.** It works only while both bugs are present. The next person to fix one in isolation ships a silently weakened compile-time check. Rejected.

4. **Add a filetest pinning the `ConstExpr` path.** Would be ideal, but the branch is unreachable from this call site — no Gno source shape was found that makes `n.Func` a `*ConstExpr` here. Rejected as not constructible.

## Scope notes

Found while merging master into #5689 and reviewing the surrounding cur-call check. Independent of that PR, which contributes only regression tests. The `recover()` block dates to #4316 (xform2).
