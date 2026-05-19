# Cross-Call Migration Guide

Practical learnings from migrating bare `cross` → `cross2(cur)` /
`cross2(rlm)` across the gno tree. Each entry is a general rule
distilled from a concrete pitfall encountered during migration.

This is a living document. Append new learnings as discovered.

---

## 1. Crossing-fn vs non-crossing-fn-with-rlm: choosing the form

Two ways to thread `realm` through a function signature:

| Form | Semantics |
|---|---|
| `func F(cur realm, ...)` | **Crossing function.** Called via `F(cross, ...)` or `F(cross2(rlm), ...)`. Each cross-call mints a fresh `cur` in the callee, creates a realm boundary, triggers finalization on return. |
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

## 2. `cross2(cur)` ≠ bare `cross` for tests using `testing.SetRealm`

Bare `cross` and `cross2(cur)` mint the new cur differently:

- **Bare `cross`** (preprocessor sentinel): runtime's
  `installCrossingCur` takes the `argtv.IsUndefined()` path and calls
  `m.callingCurOrOrigin()`. If no crossing frame is found, it falls
  through to `buildOriginRealm(m)` which **dynamically reads
  `m.Context.OriginCaller`** and constructs a fresh realm with
  `addr=OriginCaller, pkgPath="", prev=truly-nil`.
- **`cross2(rlm)`**: takes the `else` branch in `installCrossingCur`
  and uses `*argtv` as prev — the static rlm value passed in.

Consequence: in test scopes where the pattern is
`testing.SetRealm(NewUserRealm(addr))` followed by an IIFE
same-realm-cross to mint a fresh cur whose prev is the just-set user,
bare `cross` works (it picks up `OriginCaller=addr` from the updated
context), but `cross2(cur)` does **not** (it uses the static outer cur,
whose prev is unchanged).

The IIFE in such tests is doing real work, not noise. If you remove
it during migration, the test will fail in a subtle way — the realm
value reaching the SUT (e.g., `m.NewPost(0, cur, text)`) will have
`Previous() = test framework default user`, not the SetRealm'd user.

**Workaround for migration:** if the function under test takes a
non-crossing `(_ int, rlm realm, ...)` and uses `rlm.Previous()` for
identity, keep the bare-`cross` IIFE for now. Migrating these
specific sites requires either (a) a new "build cur from origin"
primitive, or (b) refactoring the SUT to take `caller address`
explicitly.

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
why `cross2(cur)` inside the test, post-SetRealm, can act as the
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
    SomeCrossingFn(cross2(cur))
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
the call site with `cross2(cur)`.

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

## 9. `defaultXxx` package-level vars + value-copy is a foot-gun

A `/p/`-package init constructs `var defaultThing = Thing{...}`. A
constructor `New() Thing { q := defaultThing; ... }` does a
value-copy. Under foreign-call semantics, the package-level
`defaultThing` is read-only to any caller (it lives in the `/p/`'s
own — frozen — realm); the local `q` inherits the readonly taint via
(7) and any subsequent `q.Field = x` panics.

**Fix:** construct fresh inside the constructor:

```go
q := Thing{Count: DefaultCount}  // owned by caller's realm
```

Don't seed from a package-level default. The package var costs nothing
to remove if the constructor inlines the default values.

See `p/jeronimoalbi/datasource/query.gno` for the corrective.

---

## 10. Removed: dead `crossingFn` shim

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

## 11. Filetest harness strips `_filetest` from filenames

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

## Appendix: open questions

- **How best to mint a fresh cur after `SetRealm(NewUserRealm)`?**
  See (2). A `cross2()`-no-arg form that uses `buildOriginRealm`
  semantics would let us finish the migration without keeping the
  IIFE-bare-cross pattern in test helpers. The alternative — exposing
  the realm constructor primitives at the Gno level — is also viable.
