# Interrealm Specification v3a — Machine-Op Authority, Provenance-Tagged Callables

Sketch for the first step after v2. The substrate v2 built —
allocation-time `PkgID = authority = storage` — is the precondition;
v3a drops the syntactic `cross/non-cross` ceremony, removes the
user-facing `realm` *value* (closing the omarsy class), and reduces
the call-site discipline to a single distinction: **direct vs
indirect**, decided by the *form of the call expression*, not by a
type-level marker.

v3a deliberately does **not** introduce a new function-type qualifier.
That's [v3b](interrealm_v3b.md). v3a is the minimum coherent step:
machine ops + provenance on FuncValue + dispatch-time borrow. Source
code keeps using plain `func(...)` everywhere.

Supersedes the call-site ceremony of v2. Does not change v2's storage
model. Assumes v2 has shipped (PR #5669).

## Background

v2 gave us two correctness properties that v3a relies on:

1. `ObjectInfo.ID.PkgID` is stamped at allocation, equals authority,
   equals storage. The receiver-borrow rule on method dispatch can read
   PkgID and trust it pre-finalize.
2. `/r/`-typed values originate inside their declaring realm
   (construction-time gate). Type forgery via foreign `/r/` types is
   closed at the source.

v2's call-site ergonomics are noisy:

```go
// v2 — ACL helper in /p/nt/ownable/v0/ownable.gno
func (o *Ownable) TransferOwnership(_ int, rlm realm, newOwner address) error {
    if !rlm.IsCurrent() {
        return ErrUnauthorized
    }
    caller := rlm.Previous().Address()
    if !o.OwnedBy(caller) {
        return ErrUnauthorized
    }
    o.owner = newOwner
    return nil
}

// caller in /r/myrealm
func TransferTo(cur realm, newOwner address) {
    o.TransferOwnership(0, cur, newOwner)  // hand cur to ownable
}
```

The `(_ int, rlm realm)` shape is a workaround for "non-crossing helper
that needs caller identity"; `cross(fn)(...)` and `cross2(rlm)` mark
realm transitions explicitly; `rlm.IsCurrent()` rejects forged or stale
realm values. Three call shapes (crossing, non-crossing-with-rlm,
plain), one `cross` sentinel, one `cross2`, and a forgeable interface.

The forgeability is the killer: in v2, `realm` is an *interface* in the
uverse. The `sealed` flag prevents adding methods to it, but does not
prevent user types from satisfying it. A `/p/attacker.FakeRealm` with
hand-written `IsCurrent() { return true }` and
`Previous().Address() { return victim }` defeats every `IsCurrent` +
`rlm.Previous()` gate in the migrated libraries.

## The pivot

Borrowing is redundant for direct calls (the call site literally names
the target — the language can see who's running) and load-bearing for
indirect calls (interface dispatch, function values, closures from
elsewhere — the call site cannot see who's running). v2 applies the same
ceremony to both. v3a separates them by **call-expression form**.

**Direct call:** the static call expression resolves to a known
`FuncDecl`, a `MethodExpr` on a concrete static type, or a method on a
named-type receiver where the receiver's type is fully known at the
call site. The compiler can name the target realm.

**Indirect call:** the static call expression is on a variable, an
interface method, a function field load, a `MapValue` slot, or any
other dynamically-resolved callable. The compiler cannot name the
target realm; provenance must be read off the value at dispatch.

Direct calls do not need a borrow check at dispatch. The dispatch site
statically knows the callable's declaring package; the language inserts
the realm shift automatically and the callee reads caller identity from
the frame above. **No `cur` parameter; no `cross`.**

Indirect calls do need the borrow check. The dispatch site reads the
callable's *provenance realm* (the new `OriginRealm` field on
`FuncValue`/`BoundMethodValue`) and shifts `m.Realm` there for the
duration of the call. **No user-visible `realm` value; the shift is
machine-internal.**

## Design

### Callable provenance

Every `FuncValue` and `BoundMethodValue` carries an `OriginRealm` field
set at construction:

- Function literal / method declaration: `OriginRealm = declaring
  package's realm`.
- Closure: `OriginRealm = declaring lexical scope's realm`.
- Method bound on a defined-type receiver: `OriginRealm` resolves at
  dispatch time from the receiver's `PkgID` (this is already v2
  semantics, just renamed).

Provenance is a property of the *value*. It is set once at
construction and never mutates. Holding a `FuncValue` in a local
variable or stashing it in a map doesn't change its `OriginRealm`. The
distinction that matters at dispatch is the *call-site form* (direct
or indirect), not where the value has travelled.

### Machine-op authority

The uverse `realm` interface is removed from user-visible surface. In
its place, `chain/runtime` exposes the authority queries as machine
ops:

```go
package runtime

func Caller() Realm   // the realm at the nearest cross-realm transition above this frame
func Self() Realm     // the executing frame's authority realm (== m.Realm)
func Origin() Address // the EOA that initiated the tx
func Tx() TxInfo      // metadata about the originating MsgCall

type Realm struct {
    addr    Address
    pkgPath string
}

func (r Realm) Address() Address { return r.addr }
func (r Realm) PkgPath() string  { return r.pkgPath }
func (r Realm) IsCode() bool     { return r.pkgPath != "" }
func (r Realm) IsUser() bool     { return r.pkgPath == "" }
```

`runtime.Realm` is a concrete struct. It is *not* an interface. There
is no method by which a user-defined type can be supplied where a
`Realm` is expected — every `Realm` value originates from a machine op
returning a struct that the VM constructed itself. The forgery class
disappears at the type system.

**`Caller()` semantics — precise.** Starting from the current frame,
walk up the call stack and return the realm of the first frame whose
`m.Realm` differs from the current frame's `m.Realm`. Frames belonging
to the same realm as the callee are transparent — a private helper
inside `/r/me` calling `runtime.Caller()` sees the same answer as if
the public entry point had called it.

Equivalently: `Caller()` returns "who crossed the realm boundary into
me." For a `/r/me` private helper called by another `/r/me` function,
that's whoever crossed into `/r/me` at the top of the contiguous-same-
realm run.

For an indirect call, the borrow has already shifted `m.Realm` to the
callable's origin; the prior frame's `m.Realm` is the dispatcher, so
`Caller()` returns the dispatcher's realm — the correct security
answer to "who is asking me to do this?"

If you need to discover invokers further up (past one or more
cross-realm transitions), the stack walk continues by transition:

```go
runtime.Caller()             // nearest cross-realm transition above this frame
runtime.CallerN(2)           // two cross-realm transitions up
runtime.Origin()             // bottom of stack: the EOA
```

The `n` in `CallerN(n)` counts *transitions*, not raw frames. Same-
realm sequences collapse to one step.

### Direct call dispatch

When the preprocess resolves a call expression to a directly named
function or method, and the declaring package differs from the
caller's, it lowers the call to a *realm-transition call* that:

1. Pushes a frame with `Cur = declaring package's Realm`.
2. Runs the body.
3. Pops; `m.Realm` restored.

The transition is invisible to source code. There is no `cross()`, no
`cross2()`, no `cur realm` parameter. Authors who need caller identity
inside a direct-call body use `runtime.Caller()`.

Direct calls into your own realm don't shift — no transition, no
borrow. (Same as v2.)

### Indirect call dispatch

When the call expression is on a variable, an interface method, a
function-typed field, or a function value loaded from any
non-statically-resolved source, dispatch reads `OriginRealm` off the
`FuncValue`/`BoundMethodValue` and applies the v2 receiver-PkgID
borrow rule: shift `m.Realm` to `OriginRealm` for the duration of the
call.

This is exactly v2's Layer-2 borrow, generalized from "method on real
foreign receiver" to "any indirectly-dispatched callable."

The body sees:
- `runtime.Self()` = `OriginRealm`
- `runtime.Caller()` = the realm that performed the indirect dispatch

### Borrow is automatic

v2's three signature shapes (crossing / non-crossing-borrow / plain)
collapse into one plain signature in v3a. Whether a call "borrows"
caller authority or "crosses" into the callee's realm is decided
**structurally at dispatch**, not declared in the signature.

The rule, restated:

| Call situation | Realm shift at dispatch? | Caller identity via |
|---|---|---|
| Direct call into same realm | no shift | `runtime.Caller()` walks past same-realm frames |
| Direct call into a foreign `/r/` | shift to callee's realm | `runtime.Caller()` returns the caller's realm |
| Method on `/p/`/stdlib type whose `PkgID` = caller's realm | no shift (borrow, same as v2 Layer 2) | `runtime.Caller()` walks past same-realm frames |
| Method on `/p/`/stdlib type whose `PkgID` ≠ caller's realm | shift to receiver's `PkgID` | `runtime.Caller()` returns the caller's realm |
| Indirect call (function value, interface, field load) | shift to `OriginRealm` | `runtime.Caller()` returns the dispatcher's realm |

Concretely, the v2 Ownable case — where `o` is a `/p/`-typed value
whose `PkgID` is the caller's realm — falls into row 3: no shift, the
write to `o.owner` lands in the caller's storage, and the ACL check
uses `runtime.Caller()`. What v2 spelled with `(_ int, rlm realm, ...)`
is just a plain signature in v3a:

```go
// v3a — borrow shape collapses to plain
func (o *Ownable) TransferOwnership(newOwner address) error {
    caller := runtime.Caller().Address()
    if !o.OwnedBy(caller) { return ErrUnauthorized }
    o.owner = newOwner
    return nil
}
```

No `_ int` slot, no `rlm` parameter, no `IsCurrent` check. The borrow
behavior comes from receiver `PkgID` matching `m.Realm`, exactly as
in v2; the **signature ceremony is what goes away**, not the
mechanism.

### Sharp edge: stashing a direct target

```go
// /r/myrealm
f := /r/shop.Purchase
f(10)  // INDIRECT — f is a local variable
```

Storing `/r/shop.Purchase` into a local and calling through the local
is an indirect call by call-form. Dispatch reads `f.OriginRealm =
/r/shop` and borrows; the call works the same as `/r/shop.Purchase(10)`
would have. There is no semantic difference, but the analysis cost is
higher and the call form is less readable.

This is a behavior consistency with v2 (where `f := /r/shop.Purchase`
followed by `cross(f)(10)` is also legal). v3a's gain is that the
`cross` ceremony goes away on the direct call site; the indirect form
"just works" via provenance.

### What goes away

- `func F(cur realm, ...)` — the crossing-function signature.
- `func F(_ int, rlm realm, ...)` — the borrow-helper signature.
- `cross(fn)(args)` — the crossing-call sentinel.
- `cross2(rlm)` — the explicit-cur installer.
- `rlm.IsCurrent()` — there is no rlm value to check.
- `rlm.Previous()` — replaced by `runtime.Caller()`.
- Forgeable `realm` interface — replaced by concrete `runtime.Realm`.

### What stays

- v2's allocation-time PkgID model.
- Construction-time gate on `/r/`-typed values.
- Persistent frozen Realms for `/p/` and stdlib; cross-realm panics
  remain aborts.
- `revive()` for catching cross-realm aborts.
- `banker.NewBanker(...)` — but the capability is now picked up from
  the calling frame via `runtime.Caller()` inside the banker, not
  passed as a parameter. (See Example 5.)
- Plain `func(args) result` function types. No new qualifier in
  v3a; that's v3b.

## Worked syntax: v2 → v3a

### Example 1 — Ownable

**v2:**

```go
package ownable

func (o *Ownable) TransferOwnership(_ int, rlm realm, newOwner address) error {
    if !rlm.IsCurrent() {
        return ErrUnauthorized
    }
    caller := rlm.Previous().Address()
    if !o.OwnedBy(caller) {
        return ErrUnauthorized
    }
    o.owner = newOwner
    return nil
}

// /r/myrealm
func TransferTo(cur realm, newOwner address) {
    o.TransferOwnership(0, cur, newOwner)
}
```

**v3a:**

```go
package ownable

import "chain/runtime"

func (o *Ownable) TransferOwnership(newOwner address) error {
    caller := runtime.Caller().Address()
    if !o.OwnedBy(caller) {
        return ErrUnauthorized
    }
    o.owner = newOwner
    return nil
}

// /r/myrealm
func TransferTo(newOwner address) {
    o.TransferOwnership(newOwner)
}
```

The IsCurrent check is gone (no value to forge). The `_ int` placeholder
is gone (no signature workaround needed). The caller capability passing
is gone (the machine op resolves it). Five lines shorter, no `realm`
in the surface, and the omarsy attack has no foothold.

### Example 2 — Direct cross-realm call into a /r/

**v2:**

```go
// /r/caller
func Buy(cur realm, qty int) {
    /r/shop.Purchase(cross, qty)
}

// /r/shop
func Purchase(cur realm, qty int) {
    buyer := cur.Previous().Address()
    // ...
}
```

**v3a:**

```go
// /r/caller
func Buy(qty int) {
    /r/shop.Purchase(qty)
}

// /r/shop
func Purchase(qty int) {
    buyer := runtime.Caller().Address()
    // ...
}
```

The preprocess statically sees `/r/shop.Purchase` is foreign-realm and
inserts the realm transition. `cross` and `cur` disappear.

### Example 3 — Interface dispatch (indirect)

```go
// /p/orig
type Mutator interface {
    Run(target *Object)
}

func UseMutator(m Mutator, o *Object) {
    m.Run(o)  // indirect — m's concrete type is unknown here
}

// /r/victim
func DoStuff() {
    /p/orig.UseMutator(myMutator, myObject)
}
```

When `m.Run(o)` dispatches, the language reads `m`'s underlying
`BoundMethodValue.OriginRealm`. If `m` was constructed in `/r/attacker`
and passed through, dispatch shifts `m.Realm` to `/r/attacker` before
running `Run`'s body. The body can't write to `o` (whose PkgID is
`/r/victim`) — the v2 write-site readonly check fires. The Class-2
omarsy attack is closed because there's no `realm` value to forge, AND
there's no path by which Run can claim to be victim's authority.

### Example 4 — Callback the caller wants to authorize

Sometimes a /r/ realm wants a /p/ helper to perform an op on its behalf
— and the op must run with /r/'s authority (e.g., emitting an event,
mutating /r/-owned state). In v2 you handed over `cur`. In v3a you hand
over a closure; the closure's `OriginRealm` is automatically the
realm where it was declared.

```go
// /r/myrealm
func RunBatch(items []Item) {
    /p/batch.Apply(items, applyOne)
}

func applyOne(it Item) { /* mutates /r/myrealm state */ }

// /p/batch
func Apply(items []Item, fn func(Item)) {
    for _, it := range items {
        fn(it)  // indirect dispatch; fn.OriginRealm = /r/myrealm
    }
}
```

`/p/batch.Apply` is a direct call into /p/batch (statically named).
Inside Apply, `fn(it)` is an indirect call: dispatch reads
`fn.OriginRealm = /r/myrealm` (captured at the source declaration of
`applyOne`) and borrows back to /r/myrealm before invoking the body.
`applyOne` runs with /r/myrealm authority. No `cur` plumbing through
every layer.

### Example 5 — Banker

**v2:**

```go
func Withdraw(cur realm, amount int) {
    b := banker.NewBanker(cur, banker.BankerTypeReadonly)
    // ...
}
```

**v3a:**

```go
func Withdraw(amount int) {
    b := banker.NewBanker(banker.BankerTypeReadonly)
    // ...
}
```

`NewBanker` calls `runtime.Caller()` itself to learn whose realm to
bind the banker to. No capability plumbing — capabilities are
implicit-from-frame, not handed around.

If you want to grant a `/p/`-helper the right to bank as you, you pass
it a closure that performs the banking op:

```go
func Withdraw(amount int) {
    /p/helper.WithRetry(func() {
        b := banker.NewBanker(banker.BankerTypeReadonly)
        b.SendCoins(...)
    })
}
```

The closure's `OriginRealm` is /r/myrealm. When `/p/helper.WithRetry`
invokes it indirectly, the borrow shifts back; `NewBanker` sees
`runtime.Caller() == /r/myrealm` and binds accordingly. The capability
is just a closure — first-class, copyable, no special API.

### Example 6 — Closure escape

```go
// /r/myrealm
func MakeCounter() func() int {
    n := 0
    return func() int {
        n++
        return n
    }
}

// /r/other
import "/r/myrealm"

func Use() {
    c := /r/myrealm.MakeCounter()  // c.OriginRealm = /r/myrealm
    for i := 0; i < 3; i++ {
        c()  // each call shifts m.Realm to /r/myrealm, mutates n
    }
}
```

The closure's `n` is allocated in `/r/myrealm` at counter-construction
time (v2 allocation-time PkgID). Each indirect invocation borrows back
to `/r/myrealm`; the write to `n` succeeds; `/r/other` cannot read or
mutate `n` directly.

Note: in v3a the function type is plain `func() int` in both
signatures. The provenance is on the value, not the type. v3b changes
this so that the type system itself tracks "this is a callable from
elsewhere" — but v3a leaves it implicit.

## Why this is safe (vs v1's `runtime.PreviousRealm()`)

v3a's `runtime.Caller()` superficially resembles v1's
`runtime.PreviousRealm()` — both are implicit stack walks. v1's was
unsafe; v3a's is safe. The difference is the substrate underneath.

| v1 problem | Why v3a is unaffected |
|---|---|
| `.Title()`/`.String()` attack: a method invoked implicitly from a stdlib helper sees `PreviousRealm()` pointing at the wrong frame. | v2's receiver-borrow rule (preserved in v3a) fires on indirect dispatch, so a value-method invoked through `fmt.*` sits in a frame whose `m.Realm` has already shifted to the receiver's `PkgID`. `Caller()` returns the dispatcher, not the original caller. |
| Stack-walk-as-auth: mid-stack helpers see different answers than entry-point code. | v3a's walk is *defined* as cross-transition-bounded. Same-realm helpers are transparent — a private helper sees the same `Caller()` as the public entry that invoked it. |
| Forgery: a user-defined type satisfies the `realm` interface and returns lies. | v3a's `runtime.Realm` is a concrete struct, not an interface. No user code can construct a `Realm` value; every `Realm` originates from a VM-built machine-op return. Closed at the type system. |
| Implicit cross-realm calls (`/r/other.F(args)` without any marker) skipped the transition. | v3a preprocess statically lowers cross-package direct calls to realm-transition calls. The transition is the same VM event as v2's `cross(...)`; only the source-form spelling differs. |

So v3a is "v1's ergonomics on v2's correctness substrate." The walk is
implicit again, but the substrate makes it well-defined and unforgeable.

## How this closes omarsy

In v2, the omarsy PoC works because:

1. `FakeRealm` is a user-defined struct satisfying the uverse `realm`
   interface.
2. `Vault.TransferOwnership(0, fake, ...)` accepts `fake` as the `rlm`
   parameter.
3. `fake.IsCurrent()` returns `true` (user-controlled).
4. `fake.Previous().Address()` returns victim's address (also
   user-controlled).
5. The mutator runs, victim's owner is overwritten.

In v3a:

- There is no `rlm` parameter to pass. `TransferOwnership(newOwner)`
  takes no realm value.
- Caller identity is `runtime.Caller()`, which is a machine op returning
  a `Realm` struct *built by the VM*, not constructible by user code.
- The Class-2 attack vector has no surface.

The same logic closes the "user-defined type implements `realm`"
variant of every other migrated ACL gate.

## Identity-as-query, not identity-as-value

The omarsy fix is one instance of a larger architectural insight worth
stating directly: **v2 made caller identity into a first-class
capability value; v3a makes it a query against the live stack.** That
shift removes two attack surfaces at once.

v2's `rlm realm` parameter is a value passed across call boundaries.
Once identity is a value, two classes of attack open up:

1. **Forgery.** The value has a type. If that type is implementable by
   user code, attackers can construct lying instances. v2's `realm`
   uverse interface is implementable; `FakeRealm{}` lies in
   `IsCurrent()` and `Previous().Address()`. (The omarsy class.)
2. **Staleness.** A captured legitimate value can be stashed and
   replayed later. v2 defends with `rlm.IsCurrent()` — a live-stack
   check — but the check itself runs through the (potentially forged)
   interface dispatch. The defense is only as trustworthy as the
   value carrying it.

Every defense is a new attack surface. Make `realm` sealed → attacker
satisfies it from outside. Add `IsCurrent` → attacker overrides it.
Add provenance bits → attacker forges the bits. The fundamental
problem is that identity has been extracted into a value whose
integrity must then be defended.

v3a doesn't defend the value — it removes the value. `runtime.Caller()`
is a machine op: the VM looks at its own call stack at the moment of
the query, walks to the nearest cross-realm transition, and constructs
a `Realm` struct directly from that frame's recorded realm. There is:

- **No parameter slot** through which a user value can reach this code.
  The VM produces the answer; no attacker code touches the wire.
- **No interface to satisfy.** `runtime.Realm` is a concrete struct.
  The type system has no path for a user-defined type to appear where
  a `Realm` is expected.
- **No snapshot to replay.** Every query reads the live stack at the
  moment of the call. The `Realm` struct that's returned is just data
  (`{addr, pkgPath}`); storing it doesn't carry authority — it's
  equivalent to storing an address. The authority *is* the act of
  walking the live stack.

Forgery and staleness aren't independent threats that need separate
defenses; they are both consequences of "identity is a value." Remove
the value, and both classes vanish together.

This is the same architectural move as v2's allocation-time PkgID:
v1 inferred `PkgID` from where an object had drifted to at link-time
(a value derived after the fact); v2 stamps it at the source (a fact
the VM owns). Both moves trade an attackable derivation for a
non-attackable origin. v3a applies the move to *caller identity*; v2
already applied it to *object authority*.

## Migration sketch

**Prerequisite:** PR #5669 (v2) must merge first. v3a's substrate
dependency requires v2's allocation-time `PkgID` and receiver-borrow
rule. Don't start Phase A while v2 is still in review.

### Phase A — runtime additions + generalized anchor (non-breaking)

Single PR. Lands the machine-op identity model alongside v2's surface
(both coexist during transition). Closes omarsy and the
primitive-recv gap (Attack-H same-pkg case) without forcing migration.

- Land `chain/runtime.{Caller, Self, Origin, CallerN}` as machine ops.
- Land `runtime.Realm` concrete struct alongside (not replacing)
  the v2 uverse interface during the transition.
- Land `OriginRealm` field on `FuncValue` and `BoundMethodValue`.
- Land the call-form analyzer at preprocess: classify each `CallExpr`
  as direct or indirect; emit the appropriate dispatch op.
- **Generalize the primitive-recv anchor:** drop the
  `isPrimitiveRecvWithForeignPPtrParam`'s same-pkg restriction (the
  `pdt.PkgPath != recvPkgPath` check in v2 commit 9d560263d). Anchor
  fires uniformly for any primitive/nil receiver whose declaring pkg
  differs from current `m.Realm`. Closes the same-/p/-pkg variant of
  Attack H that v2's narrow patch missed.

New test corpus (filetests + unit tests):
- `runtime.Caller()` walk semantics — direct same-realm transparency,
  cross-realm transitions, `CallerN(n)` transition-counting.
- `runtime.Realm` non-implementability — user type satisfying its
  method set is rejected (omarsy regression).
- `OriginRealm` lifecycle — set at construction for function
  literals, closures, method values; preserved across pass/return/
  store/load.
- Call-form analyzer correctness — bare name, selector, method value,
  method expression, parenthesized, generic instantiation,
  embedded-method promoted access, function-value-in-slice/map slot.
- Indirect-dispatch borrow generalization — function value, interface
  dispatch, function-typed field load, map slot all trigger correct
  authority shift.
- Generalized anchor — same-pkg primitive-recv + foreign-/p/-ptr-param
  attack closed; legitimate primitive-recv methods (no args, primitive
  args only) unaffected.

Scale: ~150 lines VM code + ~30 filetests. Non-breaking; existing v2
surface continues to work.

### Phase B — example migration (bulk work, non-breaking)

Library-by-library migration to the new API. v2 surface still works,
so individual /p/ and /r/ packages can migrate independently.

- Replace `(_ int, rlm realm, ...)` shapes with bare signatures +
  `runtime.Caller()`.
- Replace `cross(fn)(args)` with `fn(args)`. The preprocess infers
  realm-transition from the static target's pkgpath.
- Replace `cross2(rlm)` patterns with closure-passing (Example 4 / 5).

Scale comparable to v2's PR #5669 migration. No new tests needed
beyond what migration regressions surface.

### Phase C — surface removal (breaking)

Drop the v2 surface entirely. Gates on Phase B reaching every realm
in `examples/` and every `/p/` / stdlib library.

- Drop the uverse `realm` interface, `gConcreteRealmType`, `cross`,
  `cross2`. Remove the v2 capability-passing call shapes from the
  parser.
- Once removed, omarsy's PoC fails to parse.
- Delete the dead VM paths: `installCrossingCur`, `WithCross` /
  `DidCrossing` frame fields, `Frame.Cur`.

Breaking change. Requires deprecation cycle.

### Adjacent tracks (orthogonal to v3a phases)

Independent work that proceeds in parallel without blocking v3a:

- **Safe reference contracts.** jaekwon's library-discipline pivot:
  ship `/p/grc20`, `/r/wugnot`, `/r/grcfactory`, `/r/grcreg`,
  `/r/atomicswap` as canonical templates for capability-bearing /p/
  APIs. Each demonstrates the rule "interface methods at trust
  boundaries take only primitive or /r/-declared args; never
  /p/-declared mutable refs." If the rule holds across these contracts,
  it's an existence proof that /p/ can host assets safely without
  language changes.
- **Lint rule** encoding the rule above: scan interface declarations,
  flag method specs that take /p/-declared mutable references as
  parameters. Catches downstream code that re-opens the exploit
  surface even when the underlying libraries are safe.
- **Security-discipline rules** for library and realm authors —
  five short rules covering interface contract design, capability
  handling, allowlist gating, top-level-fn-over-method preference,
  and primitive-recv attack-shape avoidance. Documented in
  [`interrealm_discipline.md`](interrealm_discipline.md). Rules 1, 4,
  5 are lintable; rule 2 is a security-guide convention; rule 3 is
  a per-callsite review pattern.
- **Marker model (conditional).** A `borrow` keyword on methods +
  interface specs (described as v3b's relative), letting interface
  designers declare whether borrow-capability impls are admitted.
  **Deferred — not committed to a phase.** Lands only if the
  library-discipline + lint approach proves insufficient in the
  field. See "Marker model: conditional future work" below.

## Open questions

1. **`runtime.CallerN` and the stack-walk surface.** The v2 model
   tries hard to avoid stack-walking-as-auth (the original
   `runtime.PreviousRealm()` bug class). `CallerN` is a controlled
   stack walk bounded by realm transitions — same shape, different
   spelling. Whether to expose `CallerN` at all is a security-product
   call: the use cases are narrow (most code wants `Caller()` or
   `Origin()`).

2. **Direct calls through method values.** `f := obj.Method` then
   `f()`. The expression `obj.Method` evaluates a bound-method value;
   the subsequent `f()` is indirect by call-form. This is a behavior
   change from v2 where method-value vs direct method call were
   roughly equivalent. Documented as a sharp edge; v3b's type
   qualifier makes it visible.

3. **Persistence of FuncValue across tx boundaries.** v2 already
   persists closures; in v3a the persisted form must include the
   `OriginRealm` field. Schema change, plus a migration story for any
   existing persisted closure data.

4. **Detection accuracy of the call-form analyzer.** Pathological cases
   (call expression on a parenthesized identifier, on an embedded
   method via promoted access, on a generic function instantiation)
   need a precise rule so direct/indirect classification is stable and
   reviewable. The Go spec's definition of "method value" vs "method
   expression" is the prior art; v3a should mirror it.

5. **Realm-internal indirect calls.** When you call through a function
   value inside your own realm, the borrow is a no-op (`OriginRealm ==
   m.Realm`). Worth ensuring no spurious frame-finalize fires on these
   — see v2 commit 9d560263d's `AuthOnlyShift` for the precedent.

## Marker model: conditional future work

A language-level extension that has come up in design discussion:
introduce a `borrow` keyword on method declarations and interface method
specs, defaulting unmarked methods/interfaces to "anchor to declaring
realm." Interface designers could refuse borrow-capability impls by
declining the marker on the interface, closing the .Title()-class
exploit at the type-system level.

```go
// hypothetical syntax
type Voter interface {
    Vote(p *Proposal)        // unmarked → only anchored impls satisfy
}

type Hasher interface {
    borrow Hash() []byte     // explicitly admits borrow impls
}

type EvilWeight uint32
func (e EvilWeight) borrow Vote(p *Proposal) { ... }
// COMPILE ERROR: borrow-marked method can't satisfy non-borrow Voter
```

**Status:** not committed to a v3a phase. Deferred pending evidence
that library-discipline + lint is insufficient.

### Why defer rather than commit

1. **Phase A + generalized anchor already closes most of the attack
   surface.** omarsy is gone, the primitive-recv same-pkg gap closes,
   and v2's Layer-2 + readonly check defends mutations through object
   receivers. The residual is narrower than the marker would address.
2. **The library-discipline pivot is genuinely promising.** If the
   safe-reference contract set (grc20, wugnot, etc.) holds the rule
   "primitive or /r/-declared args only at trust boundaries," that's
   strong evidence the language doesn't need the marker.
3. **The marker is a real language addition.** New keyword, type-system
   change, interface satisfaction rule. Worth taking only if evidence
   demands it.
4. **It remains additive.** Phase A lands cleanly without the marker.
   If later evidence shows discipline isn't holding, the marker can
   land as a follow-on PR (~100 marker annotations + ~150 lines VM
   changes) without disrupting the v3a phases.

### What would trigger landing the marker

- The lint rule shows recurring violations in third-party `/p/`
  libraries that authors are unwilling or unable to refactor.
- A concrete exploit lands in production traceable to the gap that the
  marker would close.
- The interface-design discussion in the community shifts toward
  wanting opt-in borrow contracts as a first-class language feature.

### Where to land it if triggered

Between Phase A and Phase B of v3a, as a separate PR:

| PR | Content |
|---|---|
| 1 | v3a Phase A + generalized anchor |
| **(triggered)** | **Marker model** (`borrow` keyword + interface satisfaction + stdlib annotations) |
| 2 | v3a Phase B (migration) |
| 3 | v3a Phase C (surface removal) |

The marker would inform how Phase B's migration tags `/p/` methods
that legitimately need caller authority (avl, slices, etc. wouldn't
need it because their receivers are structs — Layer-2 handles them;
only primitive/nil-receiver methods that legitimately need borrow
would carry the keyword).

## Relation to v2

v3a does not change v2's storage model, allocation-time PkgID, or
construction-time gate. Those remain the substrate. v3a changes the
*surface*: who passes what at the call site, and what the runtime sees
as authority. The internal `setRealm` / `setRealmAuthorityOnly`
machinery is largely unchanged — v3a just routes a wider set of
callable shapes through the same borrow rule, and removes the
user-visible knobs that v2 needed to make the cross/non-cross
distinction explicit.

## Where v3a stops

v3a deliberately does not introduce a type-level distinction between
"function declared here" and "function from elsewhere." Provenance is
on the *value*, classification of direct/indirect is on the *call
expression form*. This works at runtime, but it does not:

- Catch direct-vs-indirect mistakes at preprocess time.
- Let signatures *declare* "I expect a callable from elsewhere"
  vs "I expect a local callable."
- Make the round-trip "I exported a func, you imported it back" visible
  in types.

Those are [v3b](interrealm_v3b.md) territory.
