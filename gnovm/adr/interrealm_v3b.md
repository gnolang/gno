# Interrealm Specification v3b — Type-Level Provenance via `xfunc`

Optional extension on top of [v3a](interrealm_v3a.md). v3a put callable
provenance on values (the `OriginRealm` field) and decided direct vs
indirect dispatch by the *form of the call expression*. v3b promotes
the distinction into the **type system**: a new function-type qualifier
`xfunc` marks "callable from elsewhere" so signatures can declare it
and the preprocess can catch boundary mistakes.

v3a is sufficient for security. v3b is about *clarity and static
checking* — it costs additional language surface and is worth taking
only after v3a has shipped and the call-form analyzer has been
stressed against real example code.

Assumes v3a has shipped.

## Motivation

v3a's runtime model is correct but has three rough edges:

1. **Mistakes aren't caught at preprocess.** A library author writes
   `func Apply(fn func(Item))` and treats `fn` as if it were local —
   reads global state, expects caller authority, etc. The code compiles;
   at runtime, an indirect dispatch borrows realm and the assumptions
   break. There's no signal in the signature that `fn` may arrive from
   elsewhere.

2. **Signatures can't declare intent.** A `/p/` helper might want to
   accept *only* local callables (for performance, or to avoid implicit
   realm shifts) or *only* foreign callables (for capability-style
   APIs). Neither is expressible in v3a.

3. **The round-trip is invisible.** Function `MyFn` declared in
   `/r/me`, exported, imported by `/r/you`, called back as
   `/r/you.RegisterCallback(MyFn)`, then invoked indirectly inside
   `/r/you` — `MyFn`'s `OriginRealm` is still `/r/me`, but a reader of
   `RegisterCallback`'s signature has no way to know that.

v3b adds a single type qualifier to make all three visible at the
type level.

## Design

### The `xfunc` qualifier

Function types gain a parallel form:

```
func(args) result      // local — bound to this realm at declaration
xfunc(args) result     // external — provenance may be from anywhere
```

`xfunc` (read: "external func") is a distinct type, not a subtype of
`func`. Within a single realm, an `xfunc` value is structurally
identical to a `func` value — same underlying `FuncValue`, same
calling convention — but the type system treats them as separate.

**Conversion rules:**

- `func → xfunc`: implicit at any boundary where the value escapes the
  declaring realm. This includes:
  - Being passed as a parameter to a function in a different package.
  - Being returned from a function in a different package.
  - Being stored in a field, slot, or map whose owning object has a
    different `PkgID`.
  - Being assigned to a variable of type `xfunc(...)` (explicit
    request).
- `xfunc → func`: **disallowed**. There is no down-cast. Once a
  callable has been marked external, it stays external for the
  remainder of its life. (Same idea as v2's N\_Readonly bit on data.)

The "is this an escape?" check uses v2's allocation-time PkgID. The
preprocess can resolve every escape statically because it knows the
declaring package of every name.

### Method values follow the same rule

A bound-method value `obj.Method` has type `func(args) result` if
`obj` is from the local realm and `Method` is locally declared.
Otherwise it has type `xfunc(args) result`. This makes v3a's open
question #2 (method-value-vs-direct-method-call drift) visible at the
type level:

```go
// /r/me
f := /r/me.LocalThing.DoIt       // typed: func()
g := /r/you.ForeignThing.DoIt    // typed: xfunc()
```

The reader can tell at a glance which calls cross a realm.

### Dispatch is unchanged

v3a's dispatch rule still applies. The preprocess still classifies each
`CallExpr` as direct or indirect by call-form. The type qualifier
doesn't change runtime behavior; it changes what the preprocess
*accepts*.

In particular:
- Calling an `xfunc` value is always indirect dispatch.
- Calling a `func` value (still possible, when the value has not yet
  crossed any boundary) is indirect dispatch by call-form, but the
  borrow is a no-op because `OriginRealm == m.Realm`.
- Calling a directly-named function is direct dispatch (statically
  resolvable to a `FuncDecl`).

### What the type system catches

```go
// /p/batch
func Apply(items []Item, fn func(Item)) {  // expects local callable
    for _, it := range items {
        fn(it)
    }
}

// /r/myrealm
import "/p/batch"

func RunBatch() {
    /p/batch.Apply(items, applyOne)  // ERROR: applyOne escapes /r/myrealm
                                      //        on parameter boundary →
                                      //        promotes to xfunc(Item),
                                      //        cannot pass as func(Item)
}
```

The fix is in the library:

```go
// /p/batch — fixed
func Apply(items []Item, fn xfunc(Item)) {
    for _, it := range items {
        fn(it)  // indirect, borrows to fn.OriginRealm
    }
}
```

Now the signature *declares* that Apply accepts callables from
anywhere, and the type checker enforces it everywhere `Apply` is
called.

### What the type system protects against

```go
// /p/attacker
type EvilOp struct{}

func (EvilOp) Do(target *victim.Thing) {
    target.Field = "owned"
}

// /p/orig (the library)
type Doer interface {
    Do(target *victim.Thing)
}

func Run(d Doer, t *victim.Thing) {
    d.Do(t)  // indirect dispatch — d.OriginRealm = /p/attacker
             // borrow shifts m.Realm; write to t fails the v2 readonly
             // check.
}
```

This is already closed by v2 + v3a. v3b adds a *static* hint: the
method on `EvilOp` is `xfunc`-typed at any boundary where it crosses
realms, making "this is being dispatched as a foreign callable" visible
to the audit reader.

## Worked syntax: v3a → v3b

### Example 1 — Apply with type-declared expectation

**v3a:**

```go
// /p/batch — accepts anything, behavior at runtime depends on provenance
func Apply(items []Item, fn func(Item)) {
    for _, it := range items {
        fn(it)
    }
}

// /r/myrealm
/p/batch.Apply(items, applyOne)  // works; applyOne borrows back at dispatch
/p/batch.Apply(items, otherRealmFn)  // also works; borrows to otherRealm
```

**v3b:**

```go
// /p/batch — declares "I accept callables from anywhere"
func Apply(items []Item, fn xfunc(Item)) {
    for _, it := range items {
        fn(it)
    }
}

// /r/myrealm
/p/batch.Apply(items, applyOne)  // applyOne implicit-promoted to xfunc
```

The signature now reads as documentation: "Apply does not assume fn
runs with caller authority — it borrows to fn's origin."

### Example 2 — local-only helper

```go
// /r/myrealm
func eachItem(items []Item, fn func(Item)) {  // declares: local only
    for _, it := range items {
        fn(it)  // safe — fn cannot have crossed any boundary to get here
    }
}

func RunLocal() {
    eachItem(items, applyOne)  // OK — same realm
}

// /r/other
import "/r/myrealm"
func Bad() {
    /r/myrealm.eachItem(items, applyOne)  // ERROR: applyOne would escape;
                                           // can't pass as func, would need
                                           // xfunc; eachItem only accepts func
}
```

`eachItem` is now a *realm-private* iteration helper at the type level.
Other realms simply cannot call it with a callable from elsewhere.

### Example 3 — visible round-trip

```go
// /r/me
func MyHandler(req Request) Response { /* ... */ }

// /r/you
import "/r/me"

func Register() {
    /r/you.Hub.Register(/r/me.MyHandler)
    // MyHandler escapes /r/me on parameter boundary → xfunc
}

// /r/you continues
type Hub struct {
    handlers []xfunc(Request) Response  // store slot typed xfunc
}

func (h *Hub) Register(fn xfunc(Request) Response) {
    h.handlers = append(h.handlers, fn)
}

func (h *Hub) Dispatch(req Request) {
    for _, fn := range h.handlers {
        fn(req)  // each fn borrows to its OriginRealm
    }
}
```

The `xfunc` typing on `handlers` makes the audit trivial: every
handler in this slice runs with foreign authority. A new developer
reading `Hub.Dispatch` immediately sees that `fn` is not running with
the dispatcher's authority.

### Example 4 — capability binding

A `/p/` capability library can declare exactly what it produces:

```go
// /p/capability
func Bind(fn xfunc(int) int) xfunc(int) int {
    // wraps fn with logging, retry, etc.
    return func(x int) int {
        log("calling")
        return fn(x)
    }
}
```

The return type is `xfunc` — the wrapped value retains the original
`OriginRealm` (the closure captures `fn`'s identity). Callers know
from the signature that invoking the returned value will trigger an
indirect dispatch.

## What v3b adds and costs

**Adds:**
- Static catch of "library expected local callable, got foreign one."
- Signature-level declaration of "I accept callables from anywhere"
  vs "I only accept local callables."
- Visible type-level marker for the round-trip ("here is a slot
  holding foreign callables").

**Costs:**
- One new type qualifier in the language. Two flavors of every
  function type.
- Implicit promotion at boundaries — a tiny type-inference rule, but
  one that needs careful spec wording so error messages are
  understandable.
- Persistence schema must record the qualifier (already needed for
  v3a's `OriginRealm`, so the extra bit is cheap).
- Examples migration: every `/p/` library that accepts callbacks needs
  its signatures audited and possibly retyped to `xfunc`.

The cost is mostly one-time. Once the qualifier is in the language and
the standard libraries are audited, downstream code mostly Just Works
because the implicit `func → xfunc` promotion handles common cases.

## Migration sketch (assumes v3a has shipped)

Phase A — language addition, non-breaking:
- Add `xfunc` type qualifier to the parser and the type system.
- Define the implicit `func → xfunc` conversion at the relevant
  boundaries (parameter, return, field-store, escape via map/slice).
- Treat existing `func` types as accepting either flavor at call sites
  for backwards compatibility during the transition. (Reverse-direction
  flow `xfunc → func` is the new restriction; tighten it later.)

Phase B — standard library migration:
- Audit every callback-taking signature in `stdlibs/` and the
  canonical `/p/` libraries (`p/nt/...`, `p/moul/authz`, etc.). Retype
  callback parameters to `xfunc` where the library expects foreign
  callables.
- Leave realm-private helpers as `func`.

Phase C — strict mode:
- Disallow passing a `func`-typed value into a parameter that's been
  retyped to `xfunc` *without* an implicit promote — i.e., enforce
  the type system strictly. Sources that haven't been audited will
  break loudly.
- Disallow `xfunc → func` conversions entirely.

Phase A is non-breaking. Phase B is moderate-scope but smaller than
v2's migration (only callback-taking signatures). Phase C is breaking
and is the one that delivers the static-catch property.

## Open questions

1. **Where to draw the "boundary" line.** Parameter and return
   boundaries are obvious. Field stores: is `Map[k] = fn` a boundary
   if the map is in a different realm? Yes (the map's allocation-time
   PkgID is the right discriminator). Local variables in the same
   realm: not a boundary, even though `var x func(...) = fn`
   syntactically resembles an assignment. The spec needs to enumerate
   precisely.

2. **Generic functions and `xfunc`.** `func[T any](f T) T`
   instantiated with `T = func(int) int` vs `T = xfunc(int) int` —
   probably two distinct instantiations. Acceptable cost given how
   rare generic-function-of-function-type is.

3. **Interaction with reflection / type-switching.** `switch v := x.(type)
   { case func(int): ...; case xfunc(int): ... }`. Both arms exist
   in the type system; both are reachable. Implementation should
   ensure the type-switch test matches strictly.

4. **Persisted `xfunc` values across upgrades.** A realm upgrades; an
   `xfunc` slot now has a stale `OriginRealm` pointer (the source
   realm's package was redeployed). Same problem as persisted
   closures in v2; v3b doesn't introduce it but inherits it.

5. **Naming.** `xfunc` is the working title. Alternatives: `extfunc`,
   `foreign`, `capability` (overloaded), `closure` (too narrow). The
   `xchan`/`xchan<-` precedent from various languages suggests `x` as
   a prefix is at least precedented.

6. **Compatibility with v3a-only code.** v3a-era code uses plain
   `func` for all callback parameters. v3b's Phase A keeps that
   working. Phase C tightens. There's no flag-day; libraries can opt
   into `xfunc` typing on their own schedule.

## Relation to v3a

v3b is a strict extension. It does not change v3a's runtime model:
provenance is still on the value, dispatch is still decided by
call-expression form, the borrow rule at indirect dispatch is
unchanged. v3b only adds a type-level marker that:

- Prevents `xfunc → func` conversions.
- Forces `func → xfunc` at boundaries.
- Lets signatures declare which they accept.

A v3a-only deployment is fully functional and secure. A v3b deployment
is the same, plus preprocess-time signal on a class of mistakes that
v3a catches only via "the borrow happened when you didn't expect it."

## When to take v3b

After v3a has shipped and stabilized. Defer until:

1. The call-form analyzer in v3a has been exercised against the full
   `examples/` corpus and the classification rules are settled.
2. There's a concrete corpus of "library expected local, got foreign"
   bugs that v3a's runtime model exposes — bugs that v3b would catch
   statically. Without that evidence, the type-system cost may not be
   justified.
3. The language team is willing to introduce a new type qualifier.
   This is a real spec addition with parser, typecheck, and persisted-
   schema implications, not a runtime tweak.
