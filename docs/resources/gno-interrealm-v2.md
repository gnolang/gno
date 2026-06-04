# Interrealm Specification — v2

This is a standalone specification of Gno's interrealm semantics as
implemented at current HEAD. It supersedes the historical narrative
in [`gno-interrealm.md`](./gno-interrealm.md), which traced the
evolution of the design and remains useful for context. Where this
document and the v1 spec disagree, this document reflects what the
VM actually does. See also [`gnovm/adr/interrealm_v2.md`](../../gnovm/adr/interrealm_v2.md)
for a comparison and migration guide.

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
5. Write guards: readonly provenance, the storage-ownership (PkgID)
   check, conversion guards, and the construction-time check.
6. Panic and recover semantics across realm boundaries.

## 2. Realm-Context and Realm-Storage-Context

Every executing frame in Gno carries two pieces of state:

**Realm-context** — *who is acting*. Surfaced by
`runtime.CurrentRealm()` and `runtime.PreviousRealm()`. Changes only
on explicit `fn(cross, ...)` cross-calls into a crossing function
(one declared as `func fn(cur realm, ...)`).

**Realm-storage-context** (`m.Realm` in VM internals) — *who has
write authority right now*. Determines which realm a mutation
attributes to and which realm pays storage rent for new objects.
Changes on:

- Explicit `cross` cross-calls (matches realm-context after).
- Implicit borrows (described in §4). Borrows do NOT change
  realm-context.

The two diverge whenever a borrow is active. They re-align on the
next cross-call.

### 2.1 Summary Table

| Call shape | Realm-context | Storage-context | Boundary | Finalizes |
|---|---|---|---|---|
| `fn(cross, ...)` into same realm | shifts† | unchanged | yes | yes |
| `fn(cross, ...)` into different realm | shifts | shifts | yes | yes |
| `fn(cur, ...)` (non-crossing-call of crossing-function), same realm | unchanged | unchanged | no | no |
| Non-crossing call of `/r/X`-declared callable from `/r/Y` | unchanged | shifts to `/r/X` (borrow rule #1) | yes | yes |
| Stdlib/`/p/` method on real foreign-stamped receiver | unchanged | shifts to receiver's stamp (borrow rule #2) | yes | yes |
| Stdlib/`/p/` method on primitive/nil/unstamped receiver | unchanged | unchanged | no | no |
| Stdlib/`/p/` top-level function | unchanged | unchanged | no | no |

† `runtime.CurrentRealm()` returns the same realm, but
`runtime.PreviousRealm()` shifts: the prior current becomes the new
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

### 3.4 Readonly Provenance

Values read from foreign realm storage via selector, index, slice,
deref, or address-of carry readonly provenance. Mutation sites check
that provenance in addition to the target object's PkgID ownership.

Array and struct value copies are special. `ArrayValue.Copy` and
`StructValue.Copy` still use type-driven PkgID stamping (§3.2), so
transitively value-only `/p/` types such as `uint256.Uint` become
writable local copies. However, copies that may retain references
keep readonly provenance. For example, `[1][]byte` is not
transitively value-only: the copied array wrapper is fresh, but the
slice header can still alias foreign backing storage.

## 4. Borrow Rules

On every function or method call, `PushFrameCall` applies at most
one implicit borrow rule. The three rules are listed below;
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
if recv.IsDefined() {
    obj := recv.GetFirstObject(m.Store)
    if obj != nil {
        recvOID := obj.GetObjectInfo().ID
        if !recvOID.IsZero() &&
            (m.Realm == nil || recvOID.PkgID != m.Realm.ID) {
            recvPkgOID := ObjectIDFromPkgID(recvOID.PkgID)
            objpv := m.Store.GetObject(recvPkgOID).(*PackageValue)
            m.setRealm(objpv.GetRealm())
        }
    }
}
```

A non-`/r/`-declared method (stdlib or `/p/`) called on a defined
receiver whose `PkgID` differs from `m.Realm.ID` shifts `m.Realm` to
the receiver's allocating realm for the call duration. This lets
generic library helpers operate on caller-owned state:

- `bptree.Set(...)` mutates the bptree even though `bptree` lives
  in `/p/`.
- A `*grc20.fnTeller` method mutates `grc20`-typed ledger data
  whose underlying `*StructValue` is stamped with the realm that
  called `NewToken`.

**Borrow rule #2 does NOT fire when:**

- The receiver has no object identity (`GetFirstObject` returns nil):
  primitive-underlying defined types (`type Mutator int`),
  nil-pointer receivers, nil-valued slice/map/func defined types.
  This is the **no-anchor case** — `m.Realm` inherits the caller's
  value. (See §4.4 below for the attack-class implications.)
- The call is a top-level `/p/` function (no receiver).

### 4.3 Borrow rule #3 — Closure-capability borrow (`/p/`-declared closures)

```go
if fv.IsClosure {
    pid := fv.GetObjectInfo().ID.PkgID
    if !pid.IsZero() && (m.Realm == nil || pid != m.Realm.ID) {
        if pobj := m.Store.GetObject(ObjectIDFromPkgID(pid)); pobj != nil {
            if objpv, ok := pobj.(*PackageValue); ok {
                m.setRealm(objpv.GetRealm())
            }
        }
    }
}
```

When a `FuncLit` evaluates, the resulting closure remembers the
realm that created it. Later, no matter who invokes the closure or
where it was stored, borrow rule #3 sets `m.Realm` to that creator realm
for the call. (See the borrow rule #3 code block in
`gnovm/adr/interrealm_v2.md` for how the creator is recorded.)

This is what **"closure = capability"** means in practice: a closure
carries its creator's authority, and nothing — not storage, not who
calls it — can give it more.

- A `/p/`-declared factory invoked while `m.Realm = /r/A` (e.g.
  `/p/X.MakeCounter()` called from `/r/A.init`) returns a closure
  owned by `/r/A`. Subsequent invocations from `/r/B` still run
  under `/r/A`'s authority, so writes to captured `/r/A` state
  commit normally.
- Attacker `/r/M` cannot build a closure that writes `/r/V`'s data,
  even if `/r/V` accepts the closure and runs it. The closure's
  creator is `/r/M`, so its body runs under `/r/M`'s authority and
  any write into `/r/V` hits the readonly check. `/r/V`'s API can
  safely accept arbitrary `func()`-valued callbacks without being a
  confused deputy.

If the closure's source file lives in `/r/X`, borrow rule #1 has already
borrowed to `/r/X` and borrow rule #3 is a no-op. borrow rule #3 only matters when
the closure was written in `/p/` (or in code with no realm of its
own).

### 4.4 The No-Anchor Case

When borrow rule #2 doesn't fire on a `/p/`-method call, the body inherits
the caller's `m.Realm`. If the caller was *already* borrowed to a
victim realm (e.g. inside a different borrow rule #2ed `/p/`-method
body that dispatches a `/p/`-callback), the no-anchor body runs
under the victim's authority. This is the open laundering vector
documented as the **Apply class**: a `/p/`-method that invokes a
concretely-`/p/`-typed callback lets a top-level `/p/`-attacker
function inherit victim authority and write through the callback's
parameter.

See `gno-security-guide.md` §3(B) for full discussion. The
filetests `gnovm/tests/files/zrealm_launder_rdata_embed_p.gno`,
`_ptrfield_p.gno`, `_valfield_p.gno` demonstrate the attack succeeds
when victim's `/r/`-data embeds or fields a `/p/`-type with such a
higher-order method.

### 4.5 When `m.Realm` is nil

`m.Realm` is nil in two cases:

- During `/p/`-receiver method dispatch: borrow rule #2 borrows `m.Realm` to
  `pv.GetRealm()` of the receiver's stamping package, which is nil
  for `/p/` and stdlib. `m.Realm` stays nil for the duration of the
  method body and is restored on frame pop via `fr.LastRealm`.
- During stdlib top-level function calls (same mechanism — no
  declaring realm to shift to).

Both cases are intentional. The `/p/`-immutability gate in
`Realm.DidUpdate` (§3.3) fires when `rlm == nil && m.Stage ==
StageRun` and the object being written is real and `/p/`-stamped —
catching writes that would otherwise slip through unattributed.

`m.Realm` is non-nil during all other execution: `/r/`-method
dispatch (borrow rule #2 borrows to the receiver's `/r/`), declaring-realm
borrow on borrow rule #1, closure capture-realm on borrow rule #3, and the
top-level frame of a transaction (one of `/r/` or `/e/`).

## 5. Crossing Functions and Crossing-Methods

Realm-context changes occur only through explicit `fn(cross, ...)`
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
- `IsCurrent() bool` — **true only when this `cur` matches the
  topmost live crossing frame's HIV pointer identity.** Stored or
  stale realm values return false.
- `IsCode() / IsUser() / IsUserCall() / IsUserRun() / IsEphemeral()` —
  classification by address and pkgpath.
- `String() string` — debug representation.

`IsCurrent()` is the authentication primitive. Any public entry
point that uses `cur` to derive caller identity (e.g.
`cur.Previous().Address()`) **must** check `cur.IsCurrent()` first.
Without that check, a stale or attacker-supplied realm value's
`Address()` and `PkgPath()` still resolve numerically — they just
no longer refer to the live caller. This is class **2
(designation-forgery)** in `gno-security.md`.

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

### 5.4 Parity with `runtime.{Current,Previous}Realm()`

At every comparable position:

- `cur.Address()` and `cur.PkgPath()` agree with
  `runtime.CurrentRealm()`.
- `cur.Previous().Address()` and `cur.Previous().PkgPath()` agree
  with `runtime.PreviousRealm()`.

The two APIs differ only in shape: `runtime.CurrentRealm()` returns
a struct, `cur realm` is the interface. They are **distinct types**
— not assignable to each other — but surface the same identity.

## 6. Realm Boundaries

A **realm boundary** is a transition point in the call frame stack
where `m.Realm` (or `runtime.CurrentRealm()`) changes:

- Every explicit `fn(cross, ...)` is a boundary (even when crossing
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

## 8. Conversion Guards (`doOpConvert`)

The VM's conversion operator (`op_expressions.go doOpConvert`)
enforces two cross-realm invariants:

### 8.1 Case 1 — Refuse foreign-readonly source

```go
if xv.T != nil && !xv.T.IsImmutable() && m.IsReadonly(&xv) {
    if xvdt, ok := xv.T.(*DeclaredType); ok &&
        xvdt.PkgPath == m.Realm.Path {
        // allow: converting m.Realm's own declared type
    } else {
        panic("illegal conversion of readonly or externally stored value")
    }
}
```

Without this, an attacker could declare a parallel `/p/`-type with
the same struct layout as a victim-owned `/p/`-value plus a mutator
method, convert the victim's pointer to the parallel type, and
invoke the new mutator — borrow rule #2 would route `m.Realm` to victim's
realm for the duration of the `/p/`-method, so the write would
succeed under victim authority. Case 1 blocks the conversion at the
source.

The carve-out for `xv.T.PkgPath == m.Realm.Path` allows legitimate
conversion of m.Realm's own declared types.

**Implementation note**: Case 1 panics with raw Go `panic(...)`
rather than `m.Panic(...)`, which means it is **not catchable by
Gno `defer { recover() }`**. This is an implementation inconsistency
with the write-time readonly panic (which uses `m.Panic` and is
catchable). Future cleanup may normalize this; in the meantime,
realm code cannot recover from conversion panics. See
`zrealm_launder_rdata_conv_iface_box.gno` for tests confirming the
recoverability difference.

### 8.2 Case 2 — Refuse conversion to foreign `/r/`-declared type

```go
if tdt, ok := t.(*DeclaredType); ok && !tdt.IsImmutable() && m.Realm != nil {
    if IsRealmPath(tdt.PkgPath) && tdt.PkgPath != m.Realm.Path {
        panic("illegal conversion to external realm type")
    }
}
```

A realm cannot forge values of `/r/`-declared types it doesn't
declare. Combined with the construction-time check (§3.2), this
ensures every real instance of a `/r/`-declared type traces back to
its home realm's allocator.

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
Case 1 of doOpConvert are all "transaction-fatal" defenses when
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

1. **Storing a bound method value isn't a safety boundary.** A
   `/p/`-method bound to a victim-stamped receiver, stored anywhere,
   still borrow rule #2 borrows to victim when invoked. Verified in
   `zrealm_launder_rdata_mv_stored_bound_mv.gno` and
   `_attacker_stored_mv.gno`.

2. **Method expressions are different.** `me := (*T).M` is an
   *unbound* method value with the receiver as an explicit first
   argument. Calling `me(recv, args)` and `recv.M(args)` go through
   different paths; the unbound form does not anchor on the
   receiver the same way. Verified in
   `zrealm_launder_rdata_mv_method_expr.gno`.

Realm authors should treat bound method values of `/p/`-types over
their internal state as **publishing the underlying method to any
holder** — equivalent to returning a setter closure.

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
  may be exposed as non-crossing functions.
- Methods should generally be non-crossing — they describe behavior
  on data and should work uniformly regardless of where the data
  resides.

### 11.3 Public API checklist

For every exported function or method in your `/r/` realm:

- Does it take `cur realm`? If yes, does it check `cur.IsCurrent()`
  before using `cur.Previous()`, `cur.Address()`, or `cur.PkgPath()`?
- Does it return a pointer that aliases internal mutable state? If
  yes, expect attackers to invoke any method on the returned pointer
  type that borrow rule #2 borrows back to you.
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
    runtime.PreviousRealm()  // origin user, pkgpath=""
    runtime.CurrentRealm()   // /r/test/test
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
import "gno.land/r/realmA"

func main() {
    runtime.PreviousRealm()   // g1user, pkgpath=""
    runtime.CurrentRealm()    // g1user, pkgpath="gno.land/e/g1user/run"

    realmA.PublicNoncrossing()    // runs inside ephemeral, no boundary
    realmA.PublicCrossing(cross)  // crosses into realmA
}
```

The ephemeral realm's address is derived from the user's address
(via the special `e/<user>/<...>` pattern in `chain.PackageAddress`),
so coins sent to the ephemeral realm flow back to the user.

### 12.3 MsgAddPackage

A new realm's `init()` and global-variable declarations run with:

- `runtime.PreviousRealm()` = the deployer (only available during
  init — save it if you need it later).
- `runtime.CurrentRealm()` = the new realm itself.

After init completes, the deployer identity is no longer accessible
through `runtime.PreviousRealm()`. To remember the deployer,
capture `runtime.PreviousRealm().Address()` (string) into a
package-level variable during init.

The same flow applies to `/p/` package init, except after init
completes the `/p/`'s realm is frozen.

## 13. Implementation References

- Borrow rules: `gnovm/pkg/gnolang/machine.go` PushFrameCall
- `setRealm` tripwire: `gnovm/pkg/gnolang/machine.go` setRealm
- Construction-time check: `gnovm/pkg/gnolang/alloc.go`
  checkConstructionTime
- Conversion guards: `gnovm/pkg/gnolang/op_expressions.go`
  doOpConvert (Case 1 and Case 2)
- Readonly check: `gnovm/pkg/gnolang/machine.go` IsReadonly,
  PopAsPointer2
- Cross-realm panic abort:
  `gnovm/pkg/gnolang/op_call.go` doOpReturnCallDefers and
  PopUntilLastReviveFrame
- `runtime.CurrentRealm()` / `PreviousRealm()`:
  `gnovm/stdlibs/chain/runtime/native.gno`
- `cur realm` capability validation:
  `gnovm/stdlibs/uverse_realm.gno`, the `IsCurrent()` impl checks
  HIV pointer identity against the topmost live crossing frame.

For the historical evolution of the design (interrealm v1 → v2
phases, the `setRealmAuthorityOnly` mechanism that was explored and
reverted, etc.), see `gno-interrealm.md`.

For threat-class taxonomy and defensive patterns, see
[`gno-security.md`](./gno-security.md) and
[`gno-security-guide.md`](./gno-security-guide.md).
