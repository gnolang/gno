# Interrealm Specification — v2

> **Start here:** [gno-interrealm.md](./gno-interrealm.md) states the four rules; this document is the mechanism.

> **Superseded for writes.** Borrow rules #1–#3 below are replaced for writes
> by [`pr6495_cross_to_write.md`](../../gnovm/adr/pr6495_cross_to_write.md): a
> realm's stored state changes only during a crossing call into that realm.
> Realm code entered without crossing runs read-only, a `/p/`-owned receiver
> runs as the caller, and library code a caller picked runs view-only.

This is a historical mechanism reference. The four rules in
[`gno-interrealm.md`](./gno-interrealm.md) and the
[decision record](../../gnovm/adr/pr6495_cross_to_write.md) are authoritative;
where this document disagrees with them, they win. See also
[`gnovm/adr/interrealm_v2.md`](../../gnovm/adr/interrealm_v2.md) for a
comparison and migration guide.

For threat classes and defensive patterns, see
[`gno-security.md`](./gno-security.md) and
[`gno-security-guide.md`](./gno-security-guide.md).

## 1. Introduction

Gno is an extension of Go for multi-user programming on a shared,
persistent virtual machine. Where Go assumes a single programmer's
program, Gno is one shared program co-authored by many users — each
deploying their own packages — and the language is extended so the
crossings between users' code are visible in source and verifiable
by the compiler.

There are three package categories:

- **Realm packages (`/r/`)** — stateful per-user (per-realm) code.
  Holds persistent state. Has its own bech32 address that can send
  and receive coins.
- **Pure packages (`/p/`)** — library code. May not import `/r/` or
  `/e/` packages. Has no persisted `Realm` of its own; mutations to
  `/p/`-stamped objects outside of package init are blocked at
  runtime by the /p/-immutability gate (see §3.3).
- **Ephemeral packages (`/e/`)** — single-use `MsgRun` execution
  with a custom `main()`. May import `/r/` and `/p/` packages.

This document defines:

1. The two execution contexts every Gno frame carries
   (realm-context, realm-storage-context).
2. The transitions between them — explicit (`cross`) and implicit
   (the two borrow rules).
3. The captured realm value (`cur realm`) and its runtime invariants.
4. The object model: how storage is attributed (`Storage = Authority`).
5. Write guards: the storage-ownership (PkgID) check, conversion
   guards, and the construction-time check.
6. Panic and recover semantics across realm boundaries.

## 2. Realm-Context and Realm-Storage-Context

Every executing frame in Gno carries two pieces of state:

**Realm-context** — *who is acting*. Surfaced by
`unsafe.CurrentRealm()` and `unsafe.PreviousRealm()`. Changes only
on explicit `fn(cross(cur), ...)` cross-calls into a crossing function
(one declared as `func fn(cur realm, ...)`).

**Realm-storage-context** (`m.Realm` in VM internals) — *who has
write authority right now*. Determines which realm a mutation
attributes to and which realm pays storage rent for new objects.
Changes on:

- Explicit `cross(cur)` cross-calls (matches realm-context after).
- Implicit borrows (described in §4). Borrows do NOT change
  realm-context.

The two diverge whenever a borrow is active. They re-align on the
next cross-call.

### 2.1 Summary Table

| Call shape | Realm-context | Storage-context | Boundary | Finalizes |
|---|---|---|---|---|
| `fn(cross(cur), ...)` into same realm | shifts† | unchanged | yes | yes |
| `fn(cross(cur), ...)` into different realm | shifts | shifts | yes | yes |
| `fn(cur, ...)` (non-crossing-call of crossing-function), same realm | unchanged | unchanged | no | no |
| Non-crossing call of `/r/X`-declared callable from `/r/Y` | unchanged | shifts to `/r/X` (borrow rule #1) | yes | yes |
| `/p/` method on a `/p/`-owned receiver | unchanged | shifts to receiver's stamp (borrow rule #2) | yes | yes |
| `/p/` method on a realm-owned receiver (a view) | unchanged | unchanged; writes refused | no | no |
| Stdlib/`/p/` method on primitive/nil/unstamped receiver | unchanged | unchanged | no | no |
| Stdlib/`/p/` top-level function | unchanged | unchanged | no | no |

† `unsafe.CurrentRealm()` returns the same realm, but
`unsafe.PreviousRealm()` shifts: the prior current becomes the new
previous.

The "Boundary" and "Finalizes" columns are explained in §6 and §7.

## 3. Object Model

### 3.1 Real vs Unreal

Every object Gno allocates is either **real** (persisted, has a
finalized `ObjectID`) or **unreal** (allocated this transaction, may
or may not become real at finalization).

```go
type ObjectInfo struct {
    ID       ObjectID  // set if real
    Hash     ValueHash
    OwnerID  ObjectID  // parent in ownership tree, if refcount=1
    ModTime  uint64
    RefCount int
    IsEscaped bool      // refcount ≥ 2 at finalization → persisted separately
    ...
}

type ObjectID struct {
    PkgID   PkgID   // stamped at allocation
    NewTime uint64  // stamped at finalization (zero until persisted)
}
```

At realm-transaction finalization (§7), unreal objects reachable
from persisted state become real; unreachable ones are
garbage-collected.

### 3.2 Storage = Authority (PkgID Stamped at Allocation)

Every object's `ObjectID.PkgID` is set to the active
realm-storage-context at the moment the allocator constructed it.
This is the unifying invariant: **the realm that holds authority
over an object is the same realm that allocated it.** There is no
separate "owner" or "linked-from" concept distinct from PkgID.

Three states for an `ObjectID`:

| State        | `PkgID`  | `NewTime`  | Meaning                                   |
|--------------|----------|------------|-------------------------------------------|
| empty        | zero     | zero       | Never went through the allocator          |
| allocated    | set      | zero       | In memory; authority known, not persisted |
| finalized    | set      | non-zero   | Real; persisted with a tx-stamped NewTime |

Two practical consequences:

1. **Borrow rules can fire on unreal receivers.** Because PkgID is
   set at allocation, an unreal value just returned from a foreign
   realm's constructor already carries its allocating realm's PkgID
   — the storage-realm borrow (borrow rule #2) follows immediately.

2. **Construction-time check.** Composite literals, `new()`, and
   `make()` of a foreign `/r/`-declared type panic when invoked
   outside the declaring realm:

   ```
   cannot allocate gno.land/r/v.UserT in realm gno.land/r/a
   ```

   Authority cannot be forged by constructing impostor instances of
   another realm's types. Construction must go through a constructor
   declared in the type's home realm (which triggers borrow rule #1
   declaring-borrow on call, putting `m.Realm` at the home realm
   for the allocation). See `gnovm/pkg/gnolang/alloc.go`
   `checkConstructionTime`.

   Conversions to a foreign `/r/`-declared type are refused too, with
   `illegal conversion to external realm type`, directly or through
   one pointer level, unless its underlying type is a primitive, func,
   or chan type; a nil pointer is let through. See §8.2.

   Holding such a value is still not authority: `var u v.UserT` in
   another realm is allowed (stamped with `v`'s PkgID), and `&u` can be
   passed back to `v`. Never treat a non-nil `*UserT` as authorization.

3. **Copies are type-driven, not source-propagated.**
   `{Array,Struct}Value.Copy` stamps the copy from the *declared type*
   (`getDeclaredPkgID`), mirroring the allocation rule: a `/r/`-declared
   type keeps its declared `/r/` owner, while a `/p/`-declared (or
   unnamed) value's copy takes the copying realm's PkgID. So an in-place
   value-copy of a `/p/`-typed value (e.g. `*z = *x` in `uint256.Set`)
   belongs to the realm doing the copy, not the source's realm — this is
   the #5736 / #5747 fix. See `gnovm/pkg/gnolang/values.go`.

### 3.3 /p/-Immutability

`/p/` packages and the stdlib have no persisted `Realm` of their own.
`pv.GetRealm()` returns nil for them; `IsRealmPath` (in
`gnovm/pkg/gnolang/mempackage.go`) returns true only for `/r/`. Their
package-level state is re-initialized at the start of each
transaction that imports them.

Mutations to real `/p/`-stamped objects from outside their own init
are blocked at runtime. `Realm.DidUpdate` (in
`gnovm/pkg/gnolang/realm.go`) has a branch on `rlm == nil &&
m.Stage == StageRun` that panics with
`"cannot mutate <pkgpath>: package is immutable post-init"` when the
parent object is real and `/p/`-stamped. Stdlib packages are exempted
because legitimate stdlib method dispatch also reaches this path.

This gate is what makes `/p/` package state effectively immutable
post-deployment even though the language permits the syntax of a
mutation. Combined with borrow rule #2 (borrowing `m.Realm` to the
receiver's stamp on `/p/`-method dispatch), it closes the
`/p/`-attacker-via-interface class.

## 4. Borrow Rules

On every function or method call, `PushFrameCall` applies at most
one implicit borrow rule. The two rules are listed below;
implementation lives in `gnovm/pkg/gnolang/machine.go`.

### 4.1 Borrow rule #1 — Declaring-realm borrow (`/r/`-declared callables)

```go
if IsRealmPath(pv.PkgPath) {
    if m.Realm == nil || pv.PkgPath != m.Realm.Path {
        m.setRealm(pv.GetRealm())
    }
    return
}
```

Any function, method, or closure declared in a realm package `/r/X`
runs its body with `m.Realm = /r/X`. This applies uniformly to:

- Top-level `/r/X` functions
- Methods on any receiver shape (real, unreal, primitive, nil)
- Closures whose construction-site package was `/r/X`

The rule is **symmetric and unforgeable**. Calling attacker-declared
code from victim's frame runs that code under attacker's authority,
not victim's — direct field writes to victim-owned state inside the
attacker's body fail the readonly check.

### 4.2 Borrow rule #2 — Storage-realm borrow (stdlib / `/p/` methods)

```go
// a method value not called by its binder's own code runs view-only
if bindPID.IsRealmPkg() && !pOwned && !m.calledBy(bindPID) {
    m.setViewOnly()
}
// a /p/ closure built under another realm runs view-only
if fv.IsClosure && m.Realm != nil && fv.GetObjectInfo().ID.PkgID != m.Realm.ID {
    m.setViewOnly()
}
if pOwned {
    // the receiver is owned by a /p/ package: borrow to its realm
    recvOID := oi.ID
    if m.Realm == nil || recvOID.PkgID != m.Realm.ID {
        objpv := m.Store.GetObject(ObjectIDFromPkgID(recvOID.PkgID)).(*PackageValue)
        m.setRealm(objpv.GetRealm())
    }
}
```

A `/p/`-declared method called on a defined receiver shifts `m.Realm` to
the receiver's allocating realm only when the receiver is owned by a `/p/`
package (`oi.ID.PkgID.IsImmutablePkg()`). A receiver owned by another realm
is a view: `m.Realm` stays the caller's and the method's writes are refused.
This lets library code operate on state the package itself owns:

- `bptree.Set(...)` called by the owning realm runs with that realm's
  authority, so it mutates the realm's bptree.
- A `/p/`-owned receiver borrows to its package, whose post-init
  immutability gate then reports the write.

**Borrow rule #2 does NOT fire when:**

- The receiver is owned by another realm (or any non-`/p/` package): the
  reference is a view, `m.Realm` stays the caller's, and the method's
  writes are refused.
- The receiver is stdlib-stamped: stdlib methods run with the caller's
  authority.
- The receiver has no object identity (`GetFirstObject` returns nil):
  primitive-underlying defined types (`type Mutator int`),
  nil-pointer receivers, nil-valued slice/map/func defined types.
  This is the **no-anchor case** — `m.Realm` inherits the caller's
  value. (See §4.4 below for the attack-class implications.)
- The call is a top-level `/p/` function (no receiver). Named in source
  (`pkg.F(x)`) it runs as the caller; called through a value it runs
  view-only (§4.3).

### 4.3 Library closures — no borrow, view-only when foreign

There is no borrow rule for library closures. A closure declared in an `/r/`
package is handled by borrow rule #1 and runs as that realm wherever it is
called. A `/p/` or stdlib closure borrows nothing: called while the realm
it was built under is current, it runs as that realm; called under any
other realm, it runs view-only, so every write to a stored object is
refused. A zero stamp counts as another realm. A top-level `/p/` or stdlib
function, or a method expression, called through a value (variable, field,
argument, package var) runs view-only, even on the caller's own data. Only an
unexported function called from its own package (a stored default handler)
is exempt, since no other package can pick it. A `/p/` method reached
through an interface or a method value runs view-only too, unless the
current realm imports its package, directly or through its imports.

- A `/p/`-declared factory invoked while `m.Realm = /r/A` (e.g.
  `/p/X.MakeCounter()` called from `/r/A.init`) returns a closure built
  under `/r/A`. Called from `/r/B` it runs view-only, so writes to
  captured `/r/A` state are refused; called from `/r/A` it still writes.
- Attacker `/r/M` cannot build a closure that writes `/r/V`'s data,
  even if `/r/V` accepts the closure and runs it: the closure is not
  declared in `/r/V` and `/r/V` did not call it by name, so it runs
  view-only. `/r/V`'s API can safely accept arbitrary `func()`-valued
  callbacks without being a confused deputy.

### 4.4 The No-Anchor Case

When borrow rule #2 doesn't fire on a `/p/`-method call, the body inherits
the caller's `m.Realm`. This is the no-anchor case: the call has no
borrowed realm of its own. The historic laundering vector through it — a
`/p/` higher-order method such as `Apply` invoking a top-level `/p/`
`FuncDecl` handed in by an attacker — is now refused: rule #2 no longer
borrows `Apply` to `/r/V`, so the callback inherits the attacker's realm
and the readonly gate refuses V's node. The filetests
`gnovm/tests/files/zrealm_launder_rdata_embed_p.gno` and
`zrealm_launder_rdata_ptrfield_p.gno` pin the refusal. See
`gno-security-guide.md` §3(B).

### 4.5 When `m.Realm` is nil

`m.Realm` is nil only in a view-only frame or when no realm is active
(single-user runs). `setViewOnly` sets `m.Realm` to nil and sets the view
flag, so the readonly check refuses every stored object.

A `/p/` or stdlib package is not a nil realm: it carries a frozen realm
derived from its path (`fillPackage`), so borrow rule #2 lands `m.Realm`
on it, and `Realm.DidUpdate` refuses a post-init write with `cannot mutate
<path>: package is immutable post-init`. A stdlib top-level function keeps
the caller's realm.

## 5. Crossing Functions and Crossing-Methods

Realm-context changes occur only through explicit `fn(cross(cur), ...)`
cross-calls into **crossing functions** — functions declared with
`cur realm` as the first parameter:

```go
func MakeBread(cur realm, ingredients ...any) *Bread { ... }
```

The `cur realm` parameter must be the first parameter of the
function. Crossing functions can be declared only in `/r/`
packages, never in `/p/` (would violate the no-state-in-/p/ rule).

### 5.1 Calling a crossing function

Two valid forms:

```go
// (1) Cross-call. Shifts realm-context AND realm-storage-context
// to the callee's declaring realm. Returns via realm boundary,
// finalizing the call.
MakeBread(cross(cur), "flour", "water")

// (2) Non-crossing call. Used inside the same realm. No realm-context
// change, no realm-storage-context change, no boundary, no finalization.
MakeBread(cur, "flour", "water")
```

A non-crossing call from `/r/B` of a crossing function declared in
`/r/A` is rejected — at preprocess if statically detectable,
otherwise at runtime.

### 5.2 The `cur` parameter is a capability token

Inside the body of a crossing function, the `cur realm` parameter
is a typed handle on the realm-context at the moment of the
crossing call. It is **language-enforced**: the runtime mints one
per crossing frame, refuses to persist it, and validates each use.

`realm` is the uverse interface with these methods:

- `Address() address` — bech32 address from the realm's pkgpath.
- `PkgPath() string` — pkgpath, or `""` at chain root.
- `Previous() realm` — the captured realm that was current before
  this crossing.
- `IsCurrent() bool` — true if `cur` is part of the current
  realm-context.
- `IsCode() / IsUser() / IsUserCall() / IsUserRun() / IsEphemeral()` —
  classification by address and pkgpath.
- `String() string` — debug representation.

`IsCurrent()` guards a realm value a caller hands you, named `rlm` by
convention, never your own `cur`, the value handed to you when
crossing. Your `cur` is in the current realm-context from the moment
the call arrives, so `cur.IsCurrent()` is always true and a check on it
refuses nobody. A `rlm` may be in that context too, or left over from
an earlier call, and `rlm.IsCurrent()` is what tells you which. Read no
identity out of a `rlm` until that answers true: a left-over one
answers as readily as a live one, and the caller it names is not the
one calling you. Trusting it is class **2 (designation-forgery)** in
[`gno-security.md`](./gno-security.md).

### 5.3 Realm values are ephemeral

Captured realm values must not survive past the transaction:

- Storing a `realm`-typed value into a top-level realm var, struct
  field, map value, slice/array element, or closure capture causes
  the realm to refuse the operation at attachment time or
  transaction finalize:

  ```
  cannot persist realm value: realm values are ephemeral and tied
  to a call frame
  ```

- A `realm`-typed *parameter* or *return type* in a function
  signature is a static type reference and is allowed — the rule
  applies to the **value**, not the **type**.

To remember a caller across transactions, store `cur.Address()` or
`cur.PkgPath()` (plain strings).

### 5.4 Parity with `unsafe.{Current,Previous}Realm()`

At every comparable position:

- `cur.Address()` and `cur.PkgPath()` agree with
  `unsafe.CurrentRealm()`.
- `cur.Previous().Address()` and `cur.Previous().PkgPath()` agree
  with `unsafe.PreviousRealm()`.

The two APIs differ only in shape: `unsafe.CurrentRealm()` returns
a struct, `cur realm` is the interface. They are **distinct types**
— not assignable to each other — but surface the same identity.

### 5.5 Sub-realm identities — `cur.Sub(subpath)`

A realm can mint a **sub-realm token** for one of its internal actors
(a DAO in a registry, an account in a ledger):

```go
sub := cur.Sub("dao/42")
sub.PkgPath()  // "gno.land/r/nt/commondao/v0#dao/42"  (synthesized)
sub.Address()  // chain.PackageAddress(sub.PkgPath())  (derived)
```

`#` is the sub-realm separator (kept distinct from `:`, which gnoweb
uses in URLs to split a realm path from its render args). `#` is
reserved: no real package path can contain it (rejected at package
validation), so the synthesized form can never collide with a deployed
package, and exact-match pkgpath auth is never silently broadened.

The token is a first-class `realm` value. Cross with it (two-step —
`cross(...)` takes a bare identifier) or hand it to token-style
(`_ int, rlm realm`) APIs:

```go
sub := cur.Sub("dao/42")
target.Foo(cross(sub), ...)   // callee: cur.Previous() = sub identity
teller.Transfer(0, sub, to, amount)
b := banker.NewBanker(banker.BankerTypeRealmSend, sub)
b.SendCoins(sub.Address(), to, coins)  // spend the sub-treasury
```

Semantics:

- `sub.Previous() == cur.Previous()` — the host is **not** inserted as
  a chain step; chain depth matches a non-sub crossing. The host is
  recoverable from the pkgpath prefix (`chain.SplitPkgSubPath`).
- `sub.IsCurrent()` is true while the minting cur is the live topmost
  crossing cur; `cross(sub)` accepts under the same condition. §5.4
  parity holds: `unsafe.{Current,Previous}Realm()` surface the sub
  identity at the same positions `cur`/`cur.Previous()` do.
- Classification: `IsCode()` true; `IsUser()`, `IsUserCall()`,
  `IsUserRun()`, `IsEphemeral()` all false. A sub-identity is
  programmatic, never a user.
- `Subpath()` returns the subpath (`""` for a primary cur) — the
  canonical "am I a sub, of what" accessor, consistent with
  `SplitPkgSubPath`. Prefer it over string-parsing `PkgPath()`.
- Ephemerality (§5.3) applies: sub-tokens cannot be persisted.

Guards — `Sub()` panics unless all hold:

1. subpath matches `segment ("/" segment)*`, where a `segment` is
   `[a-z0-9]` optionally followed by `[a-z0-9_.-]*[a-z0-9]` (lowercase-alnum,
   `/`-separated, `_.-` only inside a segment; no uppercase/whitespace/non-ASCII/`..`),
   and the synthesized `host#subpath` is ≤ 256 bytes. This grammar is
   frozen at introduction — loosening later is safe, tightening would
   strand funds;
2. the receiver's own HIV is the topmost crossing frame's Cur —
   strictly stronger than `IsCurrent()`, so sub-tokens can never be
   `Sub()`d (no `host#a#b`);
3. `m.Realm.Path` equals the receiver's pkgpath — foreign code holding
   a passed-around cur cannot mint in the caller's namespace;
4. the host is not ephemeral (`/e/` run realms cannot mint).

Accepting sub-identities is **opt-in** for callees. Address-keyed auth
works unchanged (sub-addresses are ordinary addresses). PkgPath-keyed
auth must use the anchored idiom — a bare prefix also matches sibling
and subdirectory packages:

```go
p := cur.Previous().PkgPath()
ok := p == host || strings.HasPrefix(p, host+"#")
```

Off-chain and cross-realm derivation without the host's cooperation:
`chain.DerivePkgSubAddr(host, subpath)`; parse a synthesized path with
`chain.SplitPkgSubPath(p) (host, subpath, ok)` — the `#` is the
marker tooling should key on.

Trust note: passing your live `cur` to `/p/` code already delegates
your primary identity; with `Sub`, it also delegates every sub-address
your namespace could ever mint (all are derivable off-chain). Audit
`/p/` imports accordingly.

Design rationale, alternatives, and the full guard analysis:
`gnovm/adr/pr5890_realm_sub.md`.

## 6. Realm Boundaries

A **realm boundary** is a transition point in the call frame stack
where `m.Realm` (or `unsafe.CurrentRealm()`) changes:

- Every explicit `fn(cross(cur), ...)` is a boundary (even when crossing
  into the same realm — the previous-realm-stack shifts).
- Every implicit borrow (borrow rule #1 or borrow rule #2 firing) is a boundary
  when storage-context changes.
- A non-crossing call into the *same* storage-context is not a
  boundary.

The boundary determination is in `op_call.go isRealmBoundary`.
Boundaries control two things:

1. **Realm-transaction finalization** (§7) runs at boundary exit.
2. **Cross-realm panic abort** (§9) is triggered by panics that
   cross a boundary on their unwind path.

## 7. Realm-Transaction Finalization

When returning across a realm boundary, the VM performs
realm-transaction finalization for the realm that was current at
the entry side:

- Newly-reachable unreal objects are assigned `ObjectID.NewTime`
  and persisted under their PkgID (their storage realm by §3.2).
- Objects with zero refcount are garbage-collected.
- Modified objects' Merkle hashes are recomputed.

Finalization does not occur for non-crossing calls within the same
storage-context (which don't cross a boundary).

## 8. Conversion Guards (`convertValue`)

The VM's conversions (`op_expressions.go convertValue`) enforce two
cross-realm invariants. `T(x)` reaches them, and so do the implicit
conversions of `a, b = f()`, `var a, b T = f()`, `return f()`,
`g(f())`, comma-ok and range assignments:

### 8.1 Case 1 — Refuse foreign-readonly source

The guard runs when the source type is not nil, has a mutable base, and the
value is readonly. If the source is the current realm's own declared type, the
conversion is allowed; a `[]byte`/`[]rune` to `string` conversion is allowed
because it does not modify the slice; anything else panics with `illegal
conversion of readonly or externally stored value`.

The immutable-base check skips types whose base is immutable (primitives,
funcs); the readonly check does the rest. Without this check, an attacker could
declare a parallel `/p/`-type with the same struct layout as a
victim-owned `/p/`-value plus a mutator method, convert the victim's
pointer to the parallel type, and invoke the new mutator. Case 1 blocks
the conversion at the source.

The carve-out for `xv.T.PkgPath == m.Realm.Path` allows legitimate
conversion of m.Realm's own declared types.

**Implementation note**: both cases panic with `m.PanicString`,
like the write-time readonly panic, so Gno `defer { recover() }`
catches it inside a call. In a package-level `var` initializer no
call is on the stack, so nothing recovers it and the transaction
aborts. See `zrealm_launder_rdata_conv_recover.gno` and
`zrealm_launder_rdata_conv_recover_pkg.gno`.

### 8.2 Case 2 — Refuse conversion to foreign `/r/`-declared type

The target is unwrapped through one pointer level and refused when it is a
non-immutable-base, `/r/`-declared type from another realm; a nil pointer is
let through. The refusal panics with `illegal conversion to external realm
type`.

A realm cannot forge values of `/r/`-declared types it doesn't
declare. The target is looked through one pointer level, so pointer
conversions are refused too; a nil pointer points at nothing and is
let through. Combined with the construction-time check (§3.2),
this prevents another realm from populating a forged instance through
construction or conversion. It may still hold a zero value it
declared itself (§3.2).

## 9. Panic and Cross-Realm Boundary

`panic()` behaves like Go within a single realm-context. When an
unrecovered panic crosses a realm boundary on its unwind path, the
VM aborts the transaction.

**Empirically verified**: a `defer { recover() }` in the
boundary-crossing caller does **not** catch the panic. The unwind
goes through `PopUntilLastReviveFrame` (op_call.go:530); only
explicit `revive()` frames can catch a cross-boundary panic. Regular
`defer/recover` causes the entire transaction to abort via
`makeUnhandledPanicError`.

This means the readonly check, the construction-time check, and
Case 1 of convertValue are all "transaction-fatal" defenses when
they fire across a realm boundary — there is no half-mutated state
to clean up, and the attacker cannot recover-and-retry under a
different guise.

### 9.1 `revive(fn)` — boundary-aware recover

`revive(fn)` is a Gno builtin that executes `fn` and returns the
exception (if any) that crossed a realm boundary during finalization
of `fn`. It is currently enabled only in test/filetest mode. In a
future release `revive(fn)` will also wrap `fn` in transactional
(cache-wrapped) memory so any mutations are discarded on abort —
effectively giving Gno software transactional memory.

## 10. Method Values

A bound method value `mv := recv.M` is a function value that
remembers its receiver. When invoked later (`mv()`), `PushFrameCall`
sees `recv` and applies borrow rule #1 or borrow rule #2 based on `M`'s declaring
package and `recv`'s PkgID stamp — **at invocation time, not at
binding time**.

Two practical implications:

1. **A bound `/p/` method writes no stored object for a foreign caller.**
   It keeps its binder's authority only while the binder's own code calls
   it and otherwise runs view-only. Verified in
   `zrealm_view_callback.gno`, `zrealm_view_method_value_arg.gno` and
   `zrealm_launder_rdata_mv_stored_bound_mv.gno`.

2. **Method expressions are different.** `me := (*T).M` is an
   *unbound* method value with the receiver as an explicit first
   argument. Calling `me(recv, args)` and `recv.M(args)` go through
   different paths; the unbound form does not anchor on the
   receiver the same way. Verified in
   `zrealm_launder_rdata_mv_method_expr.gno`.

A bound method value of a `/p/`-type over a realm's internal state
publishes nothing by itself; there is no exception. The realm exposes a
crossing function when it wants another realm to write.

## 11. Guidelines

### 11.1 What `/p/` packages may and may not do

- May not import `/r/` or `/e/` packages.
- May not declare crossing functions (`cur realm` parameters
  forbidden).
- May call crossing functions passed in as parameters (rare; usually
  a footgun — see §3 of `gno-security-guide.md`).
- After deployment, `/p/`'s persisted realm is frozen — no state
  changes survive across transactions.

### 11.2 What `/r/` packages should expose

- Public functions intended for `MsgCall` use must be crossing
  functions (`func F(cur realm, ...)`). Non-crossing functions
  cannot be invoked directly via `MsgCall`.
- Utility functions that are common sequences of non-crossing logic
  may be exposed as non-crossing functions; other callers run them
  read-only.
- A function or method that changes the realm's state for another
  caller must be crossing and authenticate with `cur.Previous()`.

### 11.3 Public API checklist

For every exported function or method in your `/r/` realm:

- Does it accept a `rlm` parameter, a realm value the caller fills? If
  yes, call `rlm.IsCurrent()` before reading an identity out of it. Its
  own `cur` needs no check.
- Does it return a pointer that aliases internal mutable state? If
  yes, it is a view to other realms; return a read-only wrapper anyway
  so callers do not mistake it for a writer.
- Does it accept an interface or function-value parameter? If yes,
  gate with canonical-type check (`t.(*MyConcrete)` or an
  `IsCanonicalX` predicate). Embedding-based seal patterns are
  bypassable.
- Does it accept a `func(*MyPType)` callback for any `/p/`-declared
  `MyPType`? If yes, retype to use one of your own `/r/`-declared
  types as the parameter — otherwise `/p/`-attackers can launder.

See `gno-security-guide.md` §8 for the full checklist and worked
examples.

## 12. Message Types

### 12.1 MsgCall

`MsgCall` invokes a single exported crossing function on a target
realm:

```go
// PKGPATH: gno.land/r/test/test
func Public(cur realm) {
    cur.Previous().PkgPath()  // origin user, pkgpath=""
    cur.PkgPath()             // "gno.land/r/test/test"
}
```

`MsgCall` rejects non-crossing functions and `/p/` functions — only
crossing functions of `/r/` packages can be invoked directly. This
prevents accidental "non-crossing" calls that would inherit the
caller's realm-context.

### 12.2 MsgRun

`MsgRun` deploys an ephemeral `/e/g1user/run` package and invokes
its `main()`. Inside `main`, the user is both the previous-realm
(at the chain root) and shares the address with the current
ephemeral realm:

```go
// PKGPATH: gno.land/e/g1user/run
import (
    "chain/runtime/unsafe"
    "gno.land/r/realmA"
)

func main(cur realm) {
    unsafe.PreviousRealm()   // g1user, pkgpath=""
    unsafe.CurrentRealm()    // g1user, pkgpath="gno.land/e/g1user/run"

    realmA.PublicNoncrossing()    // runs inside ephemeral, no boundary
    realmA.PublicCrossing(cross(cur))  // crosses into realmA
}
```

The ephemeral realm's address is derived from the user's address
(via the special `e/<user>/<...>` pattern in `chain.PackageAddress`),
so coins sent to the ephemeral realm flow back to the user.

### 12.3 MsgAddPackage

A new realm's `init()` and global-variable declarations run with:

- `unsafe.PreviousRealm()` = the deployer (only available during
  init — save it if you need it later).
- `unsafe.CurrentRealm()` = the new realm itself.

After init completes, the deployer identity is no longer accessible
through `unsafe.PreviousRealm()`. To remember the deployer,
capture `unsafe.PreviousRealm().Address()` (string) into a
package-level variable during init.

The same flow applies to `/p/` package init, except after init
completes the `/p/`'s realm is frozen.

## 13. Implementation References

- Borrow rules: `gnovm/pkg/gnolang/machine.go` PushFrameCall
- `setRealm` tripwire: `gnovm/pkg/gnolang/machine.go` setRealm
- Construction-time check: `gnovm/pkg/gnolang/alloc.go`
  checkConstructionTime
- Conversion guards: `gnovm/pkg/gnolang/op_expressions.go`
  convertValue (Case 1 and Case 2)
- Readonly check: `gnovm/pkg/gnolang/machine.go` IsReadonly,
  PopAsPointer2
- Cross-realm panic abort:
  `gnovm/pkg/gnolang/op_call.go` doOpReturnCallDefers and
  PopUntilLastReviveFrame
- `unsafe.CurrentRealm()` / `PreviousRealm()`:
  `gnovm/stdlibs/chain/runtime/unsafe/unsafe.gno`
- `cur realm` capability validation:
  `gnovm/pkg/gnolang/uverse.go`, the `IsCurrent()` impl checks
  HIV pointer identity against the topmost live crossing frame.

For the historical evolution of the design (interrealm v1 → v2
phases, the `setRealmAuthorityOnly` mechanism that was explored and
reverted, etc.), see `gno-interrealm.md`.

For threat-class taxonomy and defensive patterns, see
[`gno-security.md`](./gno-security.md) and
[`gno-security-guide.md`](./gno-security-guide.md).
