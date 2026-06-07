# Gno Security Guide for Realm Authors

This guide consolidates the practical security learnings from auditing
cross-realm attack vectors in the Gno VM. It is the long-form companion
to `gno-security.md` (which defines the numbered threat classes) and
assumes the vocabulary of `gno-interrealm.md` (realm-context,
realm-storage-context, borrow rules, `cur realm`, `IsCurrent()`).

The goal: tell a realm author what they must do, and what they must
*not* do, to keep their realm's state safe from external manipulation.

---

## 1. Threat Model

A "victim" realm `/r/V` is one that holds state and exposes APIs.
An "attacker" is any actor — an end user, another realm `/r/A`, or a
pure package `/p/A` — that can call into `/r/V`'s exposed surface.
The attacker's goal is to mutate `/r/V`'s persisted state in ways
`/r/V` did not consent to (forge tokens, change ownership records,
silently flip a flag, etc.).

We assume the attacker:

- Can deploy their own `/r/A` or `/p/A` and import `/r/V`.
- Can call any exported function or method `/r/V` makes available.
- Can hold any pointer `/r/V` returns from its public API.
- Cannot use `reflect`, `unsafe`, goroutines, or any escape hatch
  outside the documented Gno surface (these do not exist in Gno).
- Cannot read `/r/V`'s unexported package-private fields (Go's
  package-scoped identifier rule applies in Gno too).

The threat is **write authority laundering**: causing a write to
`/r/V`-stamped data while `m.Realm` is `/r/V` and the writing code
path is attacker-controlled.

---

## 2. Four Structural Defenses

The VM provides four independent defenses. A realm becomes
exploitable when an API design defeats all of them at once.

### 2.1 Declaration-Site Rule (borrow rule #1 of PushFrameCall)

Any `/r/`-declared callable (function, method, closure) executes its
body with `m.Realm` borrowed to its declaring `/r/`. Symmetric and
unforgeable: attacker code declared in `/r/A` always runs with
`m.Realm = /r/A`, regardless of who called it or with what receiver.

Consequence: an attacker cannot get *their own code* to run with
victim's authority by tricking the victim into calling it.

### 2.2 Storage-Site Rule (borrow rule #2 of PushFrameCall)

A `/p/`-declared method (or stdlib) invoked on an object-bearing
receiver whose `PkgID` differs from the current `m.Realm` borrows
`m.Realm` to the receiver's storage realm for the call duration.

Consequence: generic library helpers (`avl.Tree.Set`, `grc20`
methods) can mutate state living in the caller's realm without
needing per-realm copies — but only state belonging to the receiver's
own realm.

### 2.3 Closure-Capability Rule (borrow rule #3 of PushFrameCall)

A closure (a `FuncLit`, as opposed to a top-level `FuncDecl`) carries
the authority of the realm that created it. The creator is fixed at
the moment the `FuncLit` is evaluated and never changes. Invoking the
closure later borrows `m.Realm` to that creator for the call
duration, regardless of where the closure is currently stored or who
invokes it.

Consequence: a closure cannot gain authority by changing hands. If an
attacker hands `/r/V` a `func()` that writes `/r/V`'s data, `/r/V`
can call it safely — the body still runs under the attacker's
authority, so the write fails readonly.

The legitimate "callback that mutates `/r/V`'s own state" pattern
requires `/r/V` itself to create the closure (e.g. by returning a
`func()` from a `/r/V`-declared crossing function), so that the
closure carries `/r/V`'s authority.

### 2.4 Readonly Taint

Values read from a foreign realm carry an `N_Readonly` taint that
propagates through field access, indexing, slicing, value copies,
interface boxing/unboxing, and conversion. Any write attempt against
a tainted target panics with `cannot directly modify readonly tainted
object`. The taint is sticky — a local copy of a foreign struct is
still tainted.

The readonly check fires uniformly at every write path: `=`, `+=`,
`++`, `*p = v`, `s[i] = v`, `m[k] = v`, `delete(m, k)`,
`copy(dst, src)`, `append(s, v)` (when `s` is foreign-tainted), etc.
The Gno VM has no audited write path that skips this check.

---

## 3. The Safety Hypothesis

A victim realm `/r/V` is safe from external state mutation if **all
three** of the following hold:

### (A) All logic-data types are `/r/`-declared.

Define your data types (`type User struct {...}`, `type Order
struct {...}`) in your own `/r/V` package, not in a shared `/p/`.
Two reasons:

1. `/p/`-attacker code cannot reference `/r/V` types in its
   signatures (the import direction `/p/ → /r/` is forbidden). An
   attacker cannot declare `func (e Evil) Mutate(*v.User)` because
   `v.User` is unreachable to `/p/A`.
2. Any `/r/`-attacker impl of an interface taking `*v.User` runs
   with `m.Realm = /r/A` by borrow rule #1, so its writes hit readonly.

### (B) No `/p/`-type embedded in `/r/V`-data has higher-order methods with concretely-`/p/`-typed callbacks.

The subtle one. If `/r/V` has

```go
type Wrapper struct {
    Inner *somelib.Node
}
```

then attackers reach `Inner` (it's exported), and `somelib.Node` may
have `Iterate(cb func(*Node) bool)` or `Apply(fn func(*Node))`.
Inside `Apply`'s body, `m.Realm` is borrowed to `/r/V` by borrow rule #2 (the
`*Node`'s `PkgID` is `/r/V`, since `/r/V` allocated it). The `Apply`
body invokes `fn`. If `fn` is a top-level `/p/A.Evil` function with
signature `func(*somelib.Node)`, **neither borrow rule fires** —
top-level `/p/`-functions have no `/r/` declaring realm and no
receiver. `m.Realm` stays at `/r/V` for the entire callback. Writes
through the parameter commit under victim authority.

Real-world `/p/`-types with this shape include
`nt/avl/v0/node.Iterate(cb func(*Node) bool)`,
`moul/cow/node.Iterate(cb func(*Node) bool)`,
`onbloc/json/builder.WriteObject(fn func(*NodeBuilder))`.

### (C) Victim does not invoke caller-supplied function/interface values while holding its own authority.

The mirror of (B), viewed from `/r/V`'s API surface. If `/r/V` has

```go
func ApplyHook(fn func(any)) {
    fn(internalState)
}
```

and an attacker passes a `/p/A`-declared `fn` (a top-level
`FuncDecl`, *not* a closure), the same gap applies: `/r/V`'s body
holds `m.Realm = /r/V`, then dispatches to attacker code that doesn't
trigger any borrow.

Closures handed in by an attacker are safe — borrow rule #3 (§2.3) borrows
`m.Realm` back to the attacker for the body, so writes into `/r/V`
fail readonly. The gap in (C) is narrower than it looks: it only
applies to top-level `/p/` `FuncDecl` values, not to arbitrary
`func()` parameters.

Defense in depth: give the callback a parameter type declared in
`/r/V` itself, e.g. `fn func(*v.User)`. `/p/` code can't name
`v.User`, and any `/r/A` implementation of a matching function runs
under `/r/A`'s authority by borrow rule #1.

Empirically verified across 60+ probe filetests:
`gnovm/tests/files/zrealm_launder_rdata_*.gno`.

---

## 4. The Encapsulation Pattern (GRC20 Reference)

`gno.land/p/demo/tokens/grc20` is the canonical example of *safe*
`/p/`-declared data. It violates (A) — `Token`, `PrivateLedger`, and
`fnTeller` are all `/p/`-declared — but compensates with airtight
encapsulation:

| Defense | How |
|---|---|
| All sensitive fields are unexported | `Token.ledger`, `PrivateLedger.balances`, `PrivateLedger.allowances`, `fnTeller.accountFn` all lowercase. Foreign packages cannot access them. |
| No exported method leaks an interior pointer | No `Token` method returns `*PrivateLedger`, `*avl.Tree`, or `*avl.Node`. |
| Authority transitions gated by `rlm.IsCurrent()` | Every `Teller` method checks `rlm.IsCurrent()` before resolving `rlm.Previous().Address()`. |
| Forgery defended by nominal type assertion | `IsCanonicalTeller(t)` checks `_, ok := t.(*fnTeller)`. Embedding wrappers (`type Evil struct { Teller }`) fail this check despite method promotion. |
| `*PrivateLedger`'s unauthenticated mutators isolated by package privacy | `Mint`/`Burn`/etc. have no `rlm` check. They're safe only because no realm exports the `*PrivateLedger` pointer. |

Realm authors using GRC20 must:

1. Store `*PrivateLedger` in a **lowercase** package-level variable.
2. Expose only authenticated entry points (`func Transfer(cur realm,
   to address, amount int64) { userTeller.Transfer(0, cur, to,
   amount) }`).
3. If accepting a `Teller` from external callers, gate with
   `IsCanonicalTeller(t)` before dispatching its methods.
4. Never import `gno.land/r/tests/vm/test20` (its `PrivateLedger` is
   deliberately exported for tests; using it in production = instant
   compromise).

---

## 5. Anti-Patterns and Footguns

### 5.1 Exposing a pointer to mutable state

```go
var users []*User
func Users() []*User { return users }   // attacker gets aliased slice
```

Any pointer (slice header, map, struct pointer) returned by a getter
is mutation-attempt surface. The readonly taint protects you from
direct field writes (`Users()[0].Name = "x"` panics), but if `*User`
has any method with a body that writes its receiver, calling that
method on the returned pointer succeeds — borrow rule #2 borrows `m.Realm`
back to `/r/V`, and the write commits.

**Rule**: getters return either values (copies), unexported method
results, or read-only views. Never a pointer to internal mutable state.

### 5.2 Embedding a `/p/`-type with concrete-callback higher-order methods

The (B)-class vector. Even if your container is `/r/`-declared, if
its embedded `/p/`-type has `Apply(fn func(*T))` or `Iterate(cb
func(*Node) bool)`, attackers can launder via top-level `/p/`-fn
callbacks.

**Rule**: when embedding/fielding a `/p/`-type, audit its method set.
If it has any `func(...) func(*PType)`-shaped method, treat embedding
as **publishing a mutator API** to the world. Either don't embed, or
keep the field unexported AND don't return aliased pointers to it.

### 5.3 Accepting an attacker callback under your own authority

```go
func (v *MyService) ApplyHook(fn func()) {
    // v.state holds /r/V authority; calling fn() runs with /r/V's
    // m.Realm. If fn is a /p/A-declared top-level function, it
    // inherits /r/V authority and can call any /r/V method as
    // "self".
    fn()
}
```

The (C)-class vector. Even `func()` is dangerous — the callback's
body can call back into your own state-mutating methods.

**Rule**: never invoke a caller-supplied function/interface value
while holding your own `m.Realm`. Either:
- Type the callback parameter with one of your own `/r/V`-declared
  types so attackers can't supply a matching `/p/`-callback, OR
- Do not invoke caller callbacks at all; design synchronous APIs.

### 5.4 Trusting an interface value without canonical-type check

```go
func DoBanking(t grc20.Teller) {
    t.Transfer(0, cur, addr, amount)   // who is t? could be Evil{Teller}
}
```

Embedding an interface gives method promotion; a forged
`type Evil struct { grc20.Teller }` passes any seal/marker check via
the embedded methods.

**Rule**: at every public entry point that accepts an interface
implementer from external callers, gate with a canonical-type assert:

```go
if _, ok := t.(*grc20.fnTeller); !ok {
    panic("not a canonical Teller")
}
```

Or use the package-provided predicate (`grc20.IsCanonicalTeller(t)`).

The "unexported marker method" seal pattern is **bypassable via
embedding** — see
`examples/gno.land/p/test/seal/filetests/z_seal_*_filetest.gno` for
four working bypass tests. Sealing remains useful as documentation but
not as an enforced boundary.

### 5.5 `IsUser()` for payment guards

When accepting native coin payment via `banker.OriginSend()`, the
caller guard must be `cur.Previous().IsUserCall()`, **not**
`IsUser()`. `IsUser()` accepts `MsgRun` ephemeral realms, which can
consume the origin-send envelope before calling your function,
bypassing your payment check.

**Rule**: For payment-guarded entry points: `IsUserCall()`. See
[effective-gno.md § Verifying inbound Coin payments](./effective-gno.md#verifying-inbound-coin-payments).

### 5.6 `cur` skipped for caller identity

```go
func DoThing(addr address) {
    log[addr] = ...   // anyone can call this with any address
}
```

The `address` parameter is attacker-controlled. To identify the
actual calling realm, the function must take `cur realm` and derive
the address inside:

```go
func DoThing(cur realm) {
    if !cur.IsCurrent() {
        panic("spoofed realm")
    }
    addr := cur.Previous().Address()
    log[addr] = ...
}
```

This is class **2 (designation-forgery)** from `gno-security.md`.

### 5.7 Stored `realm`-typed values

Storing a `realm` value (whether `cur` or `cur.Previous()`) into a
struct field, map value, package-level variable, or closure capture
panics at attachment time or transaction finalize:
`cannot persist realm value: realm values are ephemeral and tied to
a call frame`.

**Rule**: if you need to remember a caller across transactions, store
the `Address()` or `PkgPath()` (plain strings), not the realm value.

---

## 6. Properties That Make the Boundary Stronger Than Expected

Two empirically-verified properties that strengthen the security
model beyond naive expectations:

### 6.1 Cross-realm panic aborts the transaction

A panic raised inside a realm-borrowed frame **cannot be caught by
`recover()` in any other realm**. `PopFrameAndReturn` walks up frames
through `PopUntilLastReviveFrame` (`op_call.go:530`); a regular
`defer { recover() }` does not stop the unwind. The transaction
aborts entirely.

This means a write that *would* have panicked at the readonly check
takes the entire transaction with it. There is no half-mutated state
to clean up. Attackers cannot recover and retry under a different
guise.

(The `revive(fn)` builtin can catch cross-realm panics in test
contexts; this is the documented exception.)

### 6.2 Readonly taint propagates through value copy

Reading a foreign struct value into a local variable preserves the
`N_Readonly` bit. Writing to the local copy still panics. This is
Go-semantics-divergent (Go would allow the local mutation) but it
closes a class of subtle attacks where attacker might "extract"
victim data into their own context. Even the local copy is sticky.

---

## 7. Properties That Surprised Us (Worth Knowing)

### 7.1 Bound method values carry the receiver's PkgID

`mv := victim.Apply` (bound method value) is a function value that
remembers the receiver. When `mv()` is invoked later — even stored
in attacker state, boxed into an interface, retrieved through
indirection — `PushFrameCall` sees the receiver and borrow rule #2 fires
based on the receiver's `PkgID`.

Method *expressions* (unbound: `(*T).Apply`) do not carry the
receiver stamp. Calling `me(recv, args)` and `recv.Apply(args)`
go through different paths.

**Implication**: if you ever return a bound method value of a
`/p/`-type pointing into your state, an attacker can store and
invoke it later — borrow rule #2 will still borrow to your realm. Don't
return bound method values of `/p/`-types unless you know the
method body is safe under attacker invocation.

### 7.2 Conversion-time panic is not Gno-recoverable

`doOpConvert` Case 1 (foreign-readonly source conversion refused)
panics with raw Go `panic(...)`, which is **not** caught by Gno
`defer { recover() }`. The write-time readonly check at
`machine.go:2555` uses `m.Panic(typedString(...))` and **is**
catchable. This is an implementation inconsistency, not a bug — but
attacker code cannot rely on `recover()` to differentiate failure
modes. (Likely worth normalizing to `m.PanicString` for consistency;
tracked.)

### 7.3 Storage-construction-time check (Phase 2)

Allocating a foreign `/r/`-declared type with a composite literal,
`new()`, or `make()` panics:
`cannot allocate <type> in realm <m.Realm>`. Attackers cannot
fabricate impostor `*v.User` instances and pass them to victim
APIs that expect a user pointer. Construction must go through
constructors declared in the type's home realm (which trigger
borrow rule #1 declaring-borrow on call).

---

## 8. Verification Checklist for Realm Authors

Before deploying a realm:

- [ ] All my logic-data types are declared in this package (`/r/V`),
  not in a shared `/p/`. If using `/p/`-declared types
  (e.g. `grc20.Token`), they're stored in **unexported** package vars.

- [ ] Every exported function/method I expose does one of:
  - Pure read (returns primitives or values, no internal pointers).
  - Takes `cur realm` and authenticates via `cur.IsCurrent()`.
  - Documented intentionally permissive (faucet, public mint).

- [ ] No exported var or function returns a pointer aliasing
  internal mutable state. `grep -E 'func [A-Z].*\*' | grep -v error`
  on my package files is a useful sanity check.

- [ ] Every interface parameter from external callers is gated with
  a canonical-type assert (`IsCanonicalX(t)` or `t.(*ConcreteImpl)`)
  before invoking methods on it.

- [ ] No method I expose takes a `func(*MyPType)` callback (where
  `MyPType` is `/p/`-declared) and invokes it from within. If yes,
  retype the callback to use my own `/r/V`-typed parameter.

- [ ] No exported field is a `/p/`-pointer or embedded `/p/`-type
  whose type has `Iterate(cb func(*T))` / `Apply(fn func(*T))` /
  similar concretely-typed callback methods.

- [ ] Payment-guarded entry points use `cur.Previous().IsUserCall()`,
  not `IsUser()`.

- [ ] No `realm`-typed value is stored in package state, struct
  fields, maps, slices, or closure captures.

- [ ] I have not imported `gno.land/r/tests/vm/test20` (deliberately
  insecure test fixture).

---

## 9. Worked Example: A Secure Counter Realm

```go
// gno.land/r/example/counter
package counter

import "chain"

// /r/-declared data type. (A) satisfied.
type Counter struct {
    value int
    owner address
}

// gCounter is unexported. The only way to reach it is through
// the methods exposed below.
var gCounter *Counter

func init() {
    // m.Realm = /r/example/counter during init; allocation
    // stamps PkgID = /r/example/counter.
    gCounter = &Counter{value: 0, owner: address("")}
}

// Public read. Returns a value, not a pointer.
func Value() int {
    return gCounter.value
}

// Authenticated mutator. cur realm + IsCurrent() check.
func Increment(cur realm) {
    if !cur.IsCurrent() {
        panic("spoofed realm")
    }
    gCounter.value++
}

// Authenticated owner-gated mutator.
func SetOwner(cur realm, newOwner address) {
    if !cur.IsCurrent() {
        panic("spoofed realm")
    }
    if gCounter.owner != "" && cur.Previous().Address() != gCounter.owner {
        panic("not the owner")
    }
    gCounter.owner = newOwner
}

// NO method like:
//   func ApplyHook(fn func(*Counter)) { fn(gCounter) }
// because that violates (C).

// NO method like:
//   func GetCounter() *Counter { return gCounter }
// because that exposes an aliased pointer (the read methods on
// *Counter would borrow rule #2 back, and any mutator method
// on *Counter would let attackers write under our authority).
```

This realm passes the checklist. Attackers can:

- Read `Value()` — returns a copy of the int (no taint, no harm).
- Call `Increment(cur)` — runs under `/r/example/counter` borrow rule #1
  borrow; bumps the value. This is the intended public API.
- Call `SetOwner(cur, ...)` — gated by ownership check.

Attackers cannot:

- Write `gCounter.value` directly (unexported field).
- Get `gCounter` and Apply-launder it (no Apply method, no exported
  pointer).
- Forge a `cur realm` (the `IsCurrent()` check fails).
- Spoof `cur.Previous().Address()` (it's the live crossing frame).

---

## 10. Further Reading

- `gno-security.md` — numbered threat-class taxonomy (Class 1a/1b/2/3/4).
- `gno-interrealm.md` — original interrealm specification.
- `gno-interrealm-v2.md` — updated specification reflecting current HEAD.
- `effective-gno.md` — pattern-level guidance including payment guards.
- `gnovm/tests/files/zrealm_launder_*.gno` — exploit-attempt filetest
  corpus referenced throughout this guide. Each test is annotated
  with the attack mechanism and why it succeeds or fails.
- `examples/gno.land/p/test/seal/filetests/z_seal_*_filetest.gno` —
  the four bypass tests demonstrating why seal is documentation, not
  defense.
