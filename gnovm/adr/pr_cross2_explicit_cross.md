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
an arbitrary expression). It must resolve to the `cur realm` parameter
of an enclosing crossing function, and it cannot be passed as a closure
capture. These are the same lexical rules bare `cur` follows at
non-crossing call sites (`g(cur, args...)`).

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
`preprocess.go:4385-4398`: allowed in realm packages AND in any
`*_test.gno` file (regardless of package). Tests in /p/ packages can
use `cross2(rlm)` to make their cross-call intent explicit.

## Implementation

### Files changed

1. **`gnovm/pkg/gnolang/nodes.go`** — Add `CrossArgPath ValuePath` field
   to `CallExpr`. Relax `isLikeWithCross` to return true when
   `CrossArgPath.Type != 0` (the cross2 indicator, independent of
   `Args[0]`).

2. **`gnovm/pkg/gnolang/gnolang.proto`** — Add proto field 7
   (`ValuePath cross_arg_path`) to `CallExpr`.

3. **`gnovm/pkg/gnolang/pb3_gen.go`** — Regenerated via
   `misc/genproto2` (auto-generated; do not hand-edit).

4. **`gnovm/pkg/gnolang/gotypecheck.go`** — Add `func cross2(rlm realm)
   realm { return rlm }` shim to the gno 0.9 per-package
   `.gnobuiltins.gno` (alongside `var cross realm`). The 0.0 shim
   intentionally does NOT declare `cross2` — `cross2` is a 0.9-only
   construct.

5. **`gnovm/pkg/gnolang/uverse.go`** — Register `cross2` as a generic
   native (`defNative("cross2", ...)` with generic X param/result,
   mirroring `_cross_gno0p0`'s shape; body panics because the
   preprocessor should always rewrite the outer call away). Also
   factor the existing `IsCurrent` method body into a shared
   `realmIsCurrentOnMachine` helper and add `realmIsCurrentStrict`
   (rejects the HIV-less fallback path).

6. **`gnovm/pkg/gnolang/preprocess.go`** —
   - Inner `cross2` TRANS_LEAVE (uverse-func switch, alongside
     `_cross_gno0p0`, `cross`, `crossing`): validate the call shape is
     `cross2(<NameExpr>)` with exactly one bare identifier argument.
   - Outer crossing-CallExpr's `Args[0]` validity check: accept
     `*CallExpr` whose `Func` is the const-evaluated `cross2`.
     Validate the inner NameExpr resolves to a `cur realm` parameter
     of an enclosing crossing function (same rules as bare `cur`
     at non-crossing call sites: enclosing-function check,
     closure-capture rejection). Stash the resolved `ValuePath` on
     the outer CallExpr's `CrossArgPath`, replace `Args[0]` with
     `constNil`, and call `SetWithCross()`.

7. **`gnovm/pkg/gnolang/op_call.go`** — `installCrossingCur` branches
   on `cx.CrossArgPath.Type`:
   - Zero (bare cross): existing path — `prev = m.callingCurOrOrigin()`.
   - Non-zero (cross2): resolve via `m.LastBlock().GetPointerTo(m.Store,
     cx.CrossArgPath)`, run `realmIsCurrentStrict`, panic if false,
     `prev = *ptr.TV`.

### Why CrossArgPath is a serialized struct field, not an attribute

ASTs are persisted in chain state preprocessed. Attributes
(`Attributes.data`) are marked `// not persisted` — they don't survive
cold-loads of a realm. `WithCross` is a struct field (proto field 6)
and survives. If we'd stashed `CrossArgPath` as an attribute, then
after a realm cold-load the runtime would see `WithCross=true` with no
`CrossArgPath` info, silently fall back to the bare-cross path
(`callingCurOrOrigin`), and lose the IsCurrent-strict check — exactly
the security property the explicit form adds.

So `CrossArgPath` must be persisted alongside `WithCross`. This is a
wire-format change (proto field 7) and is reversible only as a future
proto change; the change is small (one `ValuePath` per CallExpr, and
only non-zero for the cross2 form).

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
  - `gnovm/tests/files/zrealm_cross2_basic.gno` — happy path.
  - `gnovm/tests/files/zrealm_cross2_notname.gno` — preprocess-time
    rejection of non-NameExpr arguments.
  - `gnovm/tests/files/zrealm_cross2_closurecap.gno` — preprocess-time
    rejection of closure-captured rlm.
