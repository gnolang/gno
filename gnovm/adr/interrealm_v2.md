In the old Gno Interrealm Specification v1, two realm contexts traveled with
each frame:

  - The crossing realm context (cur / runtime.CurrentRealm()) — the realm
    identity user code observes, used for tracking agency.

  - The storage realm context (m.Realm) — controls write access to persisted
    state.

By default, both propagated unchanged from caller to callee. Two cases shifted them:

  1. Crossing call (fn(cross, ...) into a func(cur realm, ...) crossing
  function): the crossing realm and storage realm both shifted to the callee's
  declaring package.  The `cur realm` of the caller shifted to `cur.Previous()`
  of the callee, and likewise `unsafe.CurrentRealm()` to
  `unsafe.PreviousRealm()`.

  2. Method call on a persisted (real) receiver: only the storage realm shifted
  to the receiver's persistence realm. Unless the method itself is crossing
  (first argument == `cur realm`) the agency did not shift, only the write
  access authority changed to that of the receiver's persistence realm. Thus
  this was called 'borrowing'.
  
Outside these two cases, m.Realm propagated frame-to-frame untouched.

In Gno Interrealm Specification v2 the crossing rule remains the same, but the
receiver-driven borrow is replaced by three new borrow rules It does not matter
where the receiver happens to be persisted. Rather, ALL /r/-declared callables
-- whether a function, method, or closure -- run in the realm context of their
/r/ package in which the callable is declared. This makes static analysis
vastly simpler and prevents attacker /r/-declared logic from exploiting a
victim's realm data.

There are three borrow rules in v2, not one:

  - Borrow rule #1: /r/-declared function/method/closure → borrow to the
    callable's declaring realm (any receiver shape, or top-level function).
    /r/attacker code runs with /r/attacker's authority, not victim's.

Before, /r/realmA types could also end up residing in /r/realmB depending on
where and when things got attached, which was confusing. A registry realm that
stored many objects from many realms could end up owning all of them; calling
methods on them then made the registry's other data vulnerable to direct
modification by any attacker object. Even calling .Title() on a foreign
interface value could be dangerous.

Now, /r/ type values can only be constructed from within the /r/ realm in which
they are declared — foreign realms must call constructor functions, which under
the function borrow rule run in the declaring realm's context. Even if an
external realm copies the values of structs and arrays, the copies still live
under the source's /r/ realm. In other words, all /r/realmA.Objects live in
/r/realmA. Easier to reason about. This is borrow rule #1 of three.

  - Borrow rule #2: (stdlib or /p/-declared) if the receiver of a method is a
    real, foreign-stamped object, borrow to the realm context that was active
    when the receiver was *constructed* (and will be stored).

/p/-declared things are different: they're intended as libraries and meant to
be immutable. In the Gno Interrealm Specification v2, /p/-declared objects get
stamped upon construction with the ObjectInfo.PkgID of the /r/ realm context
that constructed them (before, there was no such construction stamping). And
once ObjectInfo.PkgID is stamped with a /r/ realm, even copies retain that
source /r/ realm.

  - Borrow rule #3: if fv is a closure (FuncLit) declared in a /p/ package,
    borrow to the realm context that was active when the closure was
    *constructed* (and will be stored).

In other words, what storage realm context a /p/ declared closure runs on is
determined by which realm context instantiated it, not where it ends up being
persisted. This is similar to borrow rule #2.

In short, for /r/ declared functions, methods, and closures, the borrowed
storage realm is the realm in which the callable is declared (and constructed).
For /p/-declared methods and closures, the borrowed storage realm is the realm
in which the receiver or closure was constructed.

Statements like foo.Bar = x declared in /r/ — in a function, method, or closure
— can only mutate /r/ types declared in the same realm, or /p/ types
constructed by the same realm. This makes write-access security much easier to
reason about.

Some caveats: 

## Foreign Type Value Caveat 

```
    // PKGPATH: gno.land/r/foreignRealm
    type MyStruct struct {
        Field string
    }

    func (ms *MyStruct) Modify() {
        ms.Field += "_modified"
    }
```

```
    // PKGPATH: gno.land/r/myRealm
    var x foreignRealm.MyStruct     <-- ObjectInfo.PkgID = /r/foreignRealm
    x.Field = "..."                 <-- write fail, /r/myRealm != /r/foreignRealm
```

The zero value is stamped with ObjectInfo.PkgID = /r/foreignRealm upon (default)
construction and is persisted under /r/foreignRealm, even though the slot lives
in /r/myRealm's package block.

```
    // PKGPATH: gno.land/r/myRealm
    var x foreignRealm.MyStruct     <-- ObjectInfo.PkgID = /r/foreignRealm
    x.Modify()                      <-- write fail, /r/myRealm != /r/foreignRealm
```

Field writes and pointer-receiver method mutations are also blocked:
/r/foreignRealm cannot modify `x` because the receiver-pointer's .Base resolves
to /r/myRealm's block.

```
    // PKGPATH: gno.land/r/myRealm
    var x foreignRealm.MyStruct     <-- ObjectInfo.PkgID = /r/foreignRealm
    x = foreignRealm.GlobalMyStruct <-- OK (tags x with N_Readonly taint)
    x.Field = "..."                 <-- write fail, readonly taint
    x = *foreignRealm.NewMyStruct() <-- OK (whole-slot replace)
    x.Field = "..."                 <-- write fail, /r/myRealm != /r/foreignRealm
    x.Modify()                      <-- write fail, /r/myRealm != /r/foreignRealm
```

The slot itself, however, is /r/myRealm's. /r/myRealm can replace the slot
wholesale by assigning the return value of a /r/foreignRealm constructor (e.g.
`x = *foreignRealm.NewT()`).

```
    // PKGPATH: gno.land/r/myRealm
    var x foreignRealm.MyStruct =
        foreignRealm.MyStruct{Field:"..."} <-- fail at checkConstructionTime
```

It cannot choose the contents -- composite literals of foreign types are blocked
at checkConstructionTime.

```
    // PKGPATH: gno.land/r/myRealm
    var x *foreignRealm.MyStruct =
        foreignRealm.NewMyStruct()  <-- ObjectInfo.PkgID = /r/foreignRealm
    x.Modify()                      <-- OK
```

Pointer-typed slots (`var x *foreignRealm.T`) follow the standard cross-realm
pointer model: the pointee lives in /r/foreignRealm's heap, so /r/foreignRealm's
own non-crossing helpers can mutate it through x. This is the canonical idiom,
not a caveat.

## TODO: Note more caveats, special cases, and surprises.

----------------------------------
The following is not yet well organized.

# Interrealm Specification v2 — Realm Authority via Allocation-Time PkgID

Spec for closing the .Title()-attack class across /r/ and /p/ types and
resolving the cross-realm-stored-receiver weirdness, by making
`ObjectInfo.ID.PkgID` reflect the *authority realm* at allocation rather
than the *storage realm* determined at link-time.

Supersedes the cases enumerated in `pr4890_interrealm_cases.md` (v1
"do not storage-cross" semantics). See `gno.land/pkg/integration/testdata/interrealm_v2.txtar`
for the affirmative-coverage test suite.

## Background

The current layered borrow rule in `PushFrameCall` (machine.go:2240+)
fires in two cases:

1. `/r/`-declared callable (top-level function, method, closure) invoked
   from a different realm → soft-borrow `m.Realm` to the declaring realm.
2. Stdlib or `/p/`-declared method on a real foreign receiver →
   soft-borrow `m.Realm` to the receiver's storage realm.

This closes the .Title() attack for `/r/`-declared methods (including on
unreal and primitive receivers) and preserves legitimate library-helper
patterns. Two gaps remain:

- **/p/-attacker via interface, object receiver.** Attacker constructs
  `*/p/attacker.Evil` in attacker's realm context, returns it to victim
  through a cross-call. Under the current rule, method dispatch on Evil
  borrows to the receiver's storage realm. If Evil ended up persisted
  in victim's realm (link-time-determined), the borrow shifts m.Realm
  to victim, attacker's method body runs with victim's authority, and
  mutations to victim's state succeed.

- **External /r/ stored in caller's realm can't modify itself.** Victim
  calls `t := /r/foo.NewThing()`, stores `t` in victim's state. At
  finalize, `Thing.PkgID = victim` (link-time). Later, victim calls
  `t.Method()` — the borrow rule shifts m.Realm to /r/foo (declaring),
  but DidUpdate sees `Thing.PkgID = victim ≠ m.Realm.ID = /r/foo` and
  blocks the write. The method can't mutate its own receiver. Surprising
  and ergonomically wrong.

Both gaps share a root: storage location is allowed to drift away from
the authority semantics implied by the type.

## Design

Set `ObjectInfo.ID.PkgID` at allocation time, based on the type:

- `/r/`-declared type: `PkgID = type's declaring realm`.
- `/p/` or stdlib-declared type: `PkgID = m.Realm.ID` at the moment of
  allocation.

The object **resides in that realm** — its IAVL entry lives in the
authority realm's storage tree, MarkDirty/RefCount changes accrue
against that realm, storage rent attributes there.

`NewTime` is still assigned at finalize from the owning realm's
counter. An object isn't "real" until `NewTime != 0`.

Storage and authority are now the same field. The existing DidUpdate/
MarkDirty machinery continues to work unchanged — every write under a
borrowed `m.Realm = PkgID` is naturally tracked in the realm that owns
the object.

### Method dispatch borrow (layered, extends current HEAD)

The borrow rule in `PushFrameCall` now has three layers — borrow rule #2 was
extended to cover unreal receivers in interrealm v2, and borrow rule #3 was
added as part of the closure-as-capability work to plug the persisted
`/p/`-declared closure case that neither earlier layer caught:

```go
// borrow rule #1 (unchanged): /r/-declared callable in foreign realm →
// borrow to declaring realm. Covers top-level functions, methods on
// any receiver shape, and closures lexically declared in /r/X.
if IsRealmPath(pv.PkgPath) {
    if m.Realm == nil || pv.PkgPath != m.Realm.Path {
        m.setRealm(pv.GetRealm())
    }
    return
}

// borrow rule #2 (changed): stdlib or /p/ method on real foreign receiver →
// borrow to receiver's PkgID. Under the new model PkgID is set at
// allocation and equals the authority realm — same for both real and
// unreal receivers. The borrow now fires for unreal foreign receivers
// too, closing the /p/-attacker via interface gap.
if recv.IsDefined() {
    obj := recv.GetFirstObject(m.Store)
    if obj != nil {
        ownerPkgID := obj.GetObjectInfo().ID.PkgID
        if !ownerPkgID.IsZero() &&
            (m.Realm == nil || ownerPkgID != m.Realm.ID) {
            m.setRealm(m.Store.GetRealmByID(ownerPkgID))
        }
    }
}

// borrow rule #3 (new): /p/-declared closure (FuncLit, not FuncDecl) →
// borrow to the closure's minter: the realm active when doOpFuncLit
// ran. fv.GetObjectInfo().ID.PkgID is set by Alloc.stampPkgID at
// doOpFuncLit using m.Alloc.currentRealmID, so for a closure built
// inside /r/A.init (even via a /p/-package factory) the minter is
// /r/A, not /p/X.
//
// This realizes "closure = capability". A closure carries its
// minter's authority; passing it to another realm cannot grant
// more. /r/M cannot mint a closure that writes /r/V, even if /r/V
// stores and runs it.
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

The key change in borrow rule #2: PkgID is now reliable for unreal receivers
(set at allocation, not at link-time), so the borrow can correctly
target the receiver's authority realm even before finalize. This
closes the /p/-attacker via interface case: attacker's Evil, allocated
in attacker's realm context, carries `PkgID = /r/attacker` from
allocation; method dispatch borrows m.Realm to /r/attacker; mutations
to victim's state fail the PkgID check.

For the legitimate "external /r/ stored in caller's realm" case:
Thing's PkgID = /r/foo (declaring realm) from allocation. borrow rule #1
borrows to /r/foo. Inside Method, t.Field = x writes to Thing whose
PkgID matches m.Realm. Mutation succeeds.

### IsReal() / IsZero() / IsFinalized() semantics

The current definition `!oi.ID.IsZero()` treats any non-zero ObjectID
as real. Under the new model, PkgID is set at allocation (before
finalize), so this would mark objects as real prematurely.

There are three distinct lifecycle states an `ObjectID` can be in
under interrealm v2:

| State | `PkgID` | `NewTime` | Meaning |
|---|---|---|---|
| empty | zero | 0 | not yet allocated, or transient/never-stamped |
| allocated | non-zero | 0 | stamped at allocation, not yet finalized |
| finalized | non-zero | ≥ 1 | real, persisted, has stable identity |

Three predicates, with these intended semantics:

```go
// "totally empty" — no PkgID and no NewTime.
func (oid ObjectID) IsZero() bool {
    return oid.PkgID.IsZero() && oid.NewTime == 0
}

// "has a real identity" — finalized, persisted.
func (oid ObjectID) IsFinalized() bool {
    return oid.NewTime != 0
}

// Convenience on ObjectInfo, same as IsFinalized.
func (oi *ObjectInfo) GetIsReal() bool {
    return oi.ID.NewTime != 0
}
```

**Redefining `IsZero` to check both fields is a blocking prerequisite
of this plan.** Today (`ownership.go:81-90`) `IsZero` returns
`oid.PkgID.IsZero()` only, with a debug invariant
`PkgID == 0 ↔ NewTime == 0`. interrealm v2 breaks that invariant: an
allocated-but-not-finalized object has `PkgID ≠ 0, NewTime == 0`. If
`IsZero` is not redefined to check both fields, every caller that
means "totally empty" silently flips to mean "no PkgID" (which under
interrealm v2 means "nothing has touched the allocator yet"), and the debug
invariant panics at the first allocator stamping.

Update the existing debug invariant in `IsZero` to reflect the new
lifecycle:

```go
func (oid ObjectID) IsZero() bool {
    if debug {
        // Allocated-but-not-finalized objects: PkgID set, NewTime 0. OK.
        // Finalized-but-no-PkgID is the impossible state.
        if !oid.PkgID.IsZero() && oid.NewTime != 0 {
            // both set: finalized object. OK.
        } else if oid.PkgID.IsZero() && oid.NewTime != 0 {
            panic("invariant: NewTime set but PkgID zero")
        }
    }
    return oid.PkgID.IsZero() && oid.NewTime == 0
}
```

Audit every existing caller of `GetIsReal()`, `IsZero()` on ObjectID,
`GetObjectID().IsZero()`, `GetOwnerID().IsZero()`, and `MustGetObjectID()`,
classify each by intent ("empty" vs. "finalized") and migrate the call
to the right predicate. The audit must be exhaustive — any missed site
causes finalization or refcount bugs.

Grep targets:

```
oid\.IsZero\(\)
\.GetObjectID\(\)\.IsZero\(\)
\.GetOwnerID\(\)\.IsZero\(\)
poid\.IsZero\(\)
tvoid\.IsZero\(\)
recvOID\.IsZero\(\)
\.ID\.IsZero\(\)
GetIsReal\(\)
MustGetObjectID\(\)
```

Exhaustive audit against current HEAD (re-cite line numbers at
implementation time — the table covers every relevant site found by
the greps above):

| File:line | Code | Today's intent | interrealm v2 intent | Migration |
|---|---|---|---|---|
| `realm.go:239` | `!po.GetIsReal()` | real | **finalized** | `IsFinalized()` (or keep `GetIsReal` after its redef) |
| `realm.go:270` | `co.GetIsReal()` (DidUpdate new child) | real | **finalized** | keep `GetIsReal` (redef) |
| `realm.go:286` | `xo.GetIsReal()` (DidUpdate old child rc==0) | real | **finalized** | keep `GetIsReal` (redef) |
| `realm.go:289` | `xo.GetIsReal()` (DidUpdate old child else) | real | **finalized** | keep `GetIsReal` (redef) |
| `realm.go:314` | `!oo.GetOwner().GetIsReal()` (MarkNewReal debug) | real | **finalized** | keep `GetIsReal` (redef) |
| `realm.go:332` | `!oo.GetIsReal() && !oo.GetIsNewReal()` (MarkDirty debug) | real-or-newreal | **finalized**-or-newreal | keep `GetIsReal` (redef) |
| `realm.go:352` | same pattern in MarkNewDeleted | same | same | keep |
| `realm.go:372` | same pattern in MarkNewEscaped | same | same | keep |
| `realm.go:513` | `!oo.GetObjectID().IsZero()` (incRefCreatedDescendants recurse guard) | "already assigned" (PkgID set) | **finalized** ("already stamped with NewTime") | **migrate to `!oo.GetObjectID().IsFinalized()`** — critical: under interrealm v2, newly-allocated objects have non-zero PkgID from allocation, so leaving this as `IsZero` would skip every first visit and break finalization |
| `realm.go:540` | `child.GetIsReal()` | real | **finalized** | keep `GetIsReal` (redef) |
| `realm.go:585` | `oo.GetObjectID().IsZero()` panic (processNewDeletedMarks) | "no ID yet" | **non-finalized** (a marked-deleted object must be finalized) | **migrate to `!oo.GetObjectID().IsFinalized()`** |
| `realm.go:602` | `oo.GetObjectID().IsZero()` panic (decRefDeletedDescendants) | "no ID yet" | **non-finalized** | **migrate to `!oo.GetObjectID().IsFinalized()`** |
| `realm.go:697` | `eo.GetObjectID().IsZero()` (processNewEscapedMarks: passed-from-caller branch) | "needs assignment" | **not finalized** (call incRefCreatedDescendants iff not yet finalized) | **migrate to `!eo.GetObjectID().IsFinalized()`** |
| `realm.go:740` | `!oo.GetOwnerID().IsZero()` (markDirtyAncestors debug: escaped must have no owner) | **owner reference exists** | **empty** | **stays on `IsZero` (after IsZero redef)** — the OwnerID is set via `SetOwner(po)` which copies `po.GetObjectID()`; after `SetOwner(nil)` the OwnerID is `ObjectID{}`. With redefined `IsZero` (both fields zero), this correctly identifies "no owner." |
| `realm.go:837` | `oo.GetObjectID().IsZero()` panic (saveUnsavedObjectRecursively) | "no ID" | **non-finalized** | **migrate to `!oo.GetObjectID().IsFinalized()`** |
| `realm.go:890` | `!oo.GetIsReal()` panic (save-existing branch) | real | **finalized** | keep `GetIsReal` (redef) |
| `realm.go:904` | `oid.IsZero()` panic (saveObject) | "no ID" | **non-finalized** | **migrate to `!oid.IsFinalized()`** + add the PkgID-non-zero invariant check (Finalize-time PkgID-set invariant below) |
| `realm.go:saveObject` (NEW, interrealm v2-added) | `oid.PkgID.IsZero()` panic + PkgID-based sumDiff routing | n/a (new) | **PkgID-non-zero invariant + foreign-realm routing** | new code per §"sumDiff / storage-diff attribution for foreign objects" |
| `realm.go:removeDeletedObjects` (NEW, interrealm v2-added) | `oid.PkgID == rlm.ID` branch | n/a (new) | **foreign-realm routing** | new code per §"sumDiff / storage-diff attribution"; assumes `rlm.deleted` is invariant-finalized at entry (see invariant note below) |
| `realm.go:1782` | `!oid.IsZero()` panic (assignNewObjectID precondition) | "already assigned" | **already finalized** | **migrate to `oid.IsFinalized()` panic** (see API split below) |
| `realm.go:1814` | `!oo.GetIsReal()` panic (toRefValue) | real | **finalized** | keep `GetIsReal` (redef) |
| `realm.go:1830` | `!oo.GetOwnerID().IsZero()` panic (toRefValue: escaped must have no owner) | **owner reference exists** | **empty** | **stays on `IsZero` (after IsZero redef)** |
| `realm.go:1921` | `!poid.IsZero()` (getOwner: lazy-load via store iff owner ref exists) | **owner reference exists** | **empty** | **stays on `IsZero` (after IsZero redef)** |
| `machine.go:2251` | `!recvOID.IsZero()` (PushFrameCall borrow rule #2 guard) | "receiver has been finalized at least to PkgID stage" | **non-empty (any allocated or finalized receiver)** | **stays on `IsZero` (after IsZero redef)** — under interrealm v2 PkgID is set at allocation, so this correctly admits unreal foreign receivers (the desired borrow rule #2 expansion); falls through only for truly transient receivers (off-allocator construction sites that missed PkgID stamping, which the finalize-time invariant catches separately) |
| `machine.go:2550` | `oid.IsZero()` (isExternalRealm: transient/local-var branch) | "transient" | **empty (truly never stamped)** | **stays on `IsZero` (after IsZero redef)**. Note: under interrealm v2, allocator stamps PkgID at allocation for every object that goes through `m.Alloc.*`. Local vars do go through the allocator, so this branch becomes effectively dead for properly-stamped objects. Keep as a defensive fallback and add a comment noting the dead-branch implication. |
| `ownership.go:76` | `!oid.PkgID.IsZero() && oid.NewTime == 1` (`IsPackage` helper) | "is the package self-ObjectID" | unchanged | keep as-is |
| `ownership.go:81-90` | `IsZero()` itself | "PkgID is zero" | "both fields zero" | **redefine per Finding #1** |
| `ownership.go:223` | `oi.ID.IsZero()` panic (MustGetObjectID) | "no ID" | **non-finalized** (must be a valid ID to "must-get") | **migrate to `!oi.ID.IsFinalized()`** — no current call sites for MustGetObjectID (interface + impl only), so safe to tighten |
| `ownership.go:260` | `!oi.OwnerID.IsZero()` (GetIsOwned) | **owner reference exists** | **empty** | **stays on `IsZero` (after IsZero redef)** |
| `ownership.go:264-266` | `GetIsReal()` definition | `!ID.IsZero()` | **`ID.NewTime != 0`** | **redefine per "Redefine" block above** |
| `ownership.go:282` | `oi.GetIsReal()` (DecRefCount debug) | real | **finalized** | keep `GetIsReal` (redef) |
| `ownership.go:481` | `if tvoid.IsZero()` (IsReadonlyBy) | "no associated ObjectID" → not readonly by anyone | **empty** | **stays on `IsZero` (after IsZero redef)** — for an allocated-but-unfinalized object with PkgID set, `IsZero` returns false, so the function proceeds to the `PkgID != rid` check (the correct behavior under interrealm v2) |
| `store.go:628` | `oid.IsZero()` panic (SetObject debug) | "no ID" | **non-finalized** | **migrate to `!oid.IsFinalized()`** |
| `values_export.go:97` | `obj.GetIsReal()` (cycle-handling: emit RefValue if persisted) | real | **finalized** | keep `GetIsReal` (redef) |
| `values_export.go:136` | `oo.GetIsReal()` (persisted → emit RefValue) | real | **finalized** | keep `GetIsReal` (redef) |
| `realm_test.go:27` | `ownerID.IsZero()` (test) | empty | **empty** | stays on `IsZero` |

Summary of action by intent class:

- **15 sites** keep `GetIsReal()` and inherit the redefined `NewTime != 0` semantics. No code edit needed beyond the definition change.
- **8 sites** migrate from `.IsZero()` to `.IsFinalized()` (or `!IsFinalized()`): `realm.go:513`, `585`, `602`, `697`, `837`, `904`, `1782` (assignNewObjectID — see API split), `ownership.go:223`, `store.go:628`.
- **7 sites** stay on `.IsZero()` and inherit the redefined "both fields zero" semantics: `realm.go:740`, `realm.go:1830`, `realm.go:1921`, `machine.go:2251`, `machine.go:2550`, `ownership.go:260`, `ownership.go:481`, `realm_test.go:27`. These are all "owner-reference-exists?" or "transient?" tests where the old single-field check and the new both-fields check return the same result for every state actually used today, but the new check is necessary for correctness under the interrealm v2 allocated-but-unfinalized state.
- **1 site** is a special-form package-self check that uses raw `PkgID.IsZero()` directly and is unaffected: `ownership.go:76`.

### Finalize-time PkgID-set invariant

Add an explicit invariant check at every site that finalizes or
persists an object: `oid.PkgID` MUST be non-zero by the time
finalization runs. A zero PkgID at finalize means an allocation site
was missed by the Phase 2 allocator-stamping plumbing, and the object
slipped through without authority attribution.

Update `saveObject` (`realm.go:902-906`):

```go
func (rlm *Realm) saveObject(store Store, oo Object) {
    oid := oo.GetObjectID()
    if !oid.IsFinalized() {
        panic("unexpected non-finalized object id at save")
    }
    if oid.PkgID.IsZero() {
        // Under interrealm v2, PkgID must be stamped at allocation. A zero
        // PkgID here means an allocation site was missed by the
        // allocator plumbing. Loud-fail rather than silently saving
        // under an unattributed authority.
        panic("invariant violation: zero PkgID at save — allocation site missed allocator stamping")
    }
    // ...
}
```

Apply the same `oid.PkgID.IsZero()` panic to `assignNewObjectID`
and to any other finalize/store-write path that mints a real
object. **No fallback, no log-and-continue, no build tag.** A zero
PkgID at finalize means an off-allocator construction site was
missed in the Phase 2 audit. The panic is the audit mechanism:
running the full test suite under interrealm v2 will surface every missed
site as a loud failure with the file/line of the allocation. Fix
each site by routing through the allocator (or stamping PkgID
inline per the off-allocator construction list above) until the
panic stops firing. This is simpler than instrumenting a debug
flag, more direct than collecting diagnostics, and turns the audit
from "did I find them all?" into a binary "tests pass or they
don't."

### assignNewObjectID API split

Today (`realm.go:1813`):

```go
func (rlm *Realm) assignNewObjectID(oo Object) ObjectID {
    oid := oo.GetObjectID()
    if !oid.IsZero() {
        panic("unexpected non-zero object id")
    }
    noid := rlm.nextObjectID()  // sets both PkgID + NewTime
    oo.SetObjectID(noid)        // overwrites whole struct
    return noid
}
```

Under the new model, PkgID is non-zero at allocation, so the panic
fires immediately. Also, the new model permits a realm's finalize to
process foreign-owned objects that got attached during this realm's
execution (the "myrealm.slice = append(myrealm.slice, yourrealm.Foo)"
pattern) — those need NewTime minted from their **owning** realm's
counter, not from `rlm`. Replace with:

```go
func (rlm *Realm) assignNewObjectID(store Store, oo Object) ObjectID {
    oid := oo.GetObjectID()
    if oid.IsFinalized() {
        panic("unexpected already-finalized object id")
    }
    if oid.PkgID.IsZero() {
        // Zero PkgID at finalize means an allocation site was missed
        // by the allocator-stamping plumbing. Loud-fail so missed
        // sites surface immediately. See "Finalize-time PkgID-set
        // invariant" above.
        panic("invariant violation: zero PkgID at finalize — allocation site missed allocator stamping")
    }
    // Dispatch to the owning realm's counter. For self-owned objects
    // this is `rlm`; for foreign-owned objects it's looked up via
    // the store and tracked in `rlm.touchedForeignRealms` for a
    // single batched record-save at end-of-finalize.
    targetRlm := rlm
    if oid.PkgID != rlm.ID {
        targetRlm = rlm.touchForeignRealm(store, oid.PkgID)
    }
    targetRlm.Time++
    oo.SetNewTime(targetRlm.Time)
    return oo.GetObjectID()
}
```

### Cross-realm finalize and touchedForeignRealms

A single realm's `FinalizeRealmTransaction` may need to mint NewTimes
on foreign-owned objects (the realm executed code that took ownership
of objects allocated under foreign authority). For uniqueness, the
NewTime must come from the **owning** realm's counter, not the
finalizing realm's. To avoid saving the foreign realm's record once
per object, batch the foreign-record saves at end of the finalize
call.

Add to `Realm`:

```go
type Realm struct {
    ID   PkgID
    Path string
    Time uint64
    ...
    // Foreign realms whose Time was advanced during this realm's
    // current FinalizeRealmTransaction call. Populated lazily by
    // touchForeignRealm; drained at end of finalize via a single
    // SetPackageRealm per entry. Reset to nil after each finalize.
    touchedForeignRealms map[PkgID]*Realm
}

// touchForeignRealm resolves the foreign realm by PkgID and caches
// the *Realm pointer for this finalize call. Subsequent
// assignNewObjectID calls for objects in the same foreign realm
// re-use the cached *Realm, so its Time counter advances in-memory
// for all of them in a single finalize.
// touchForeignRealm is a pure lookup + cache. It does NOT advance
// fr.Time. Time advancement happens only in assignNewObjectID's
// own body (targetRlm.Time++) after the lookup returns. Callers
// reach touchForeignRealm via two distinct routes:
//
//   1. assignNewObjectID (minting NewTime for a not-yet-finalized
//      foreign object): the caller advances fr.Time after the
//      lookup.
//   2. saveObject / removeDeletedObjects (routing sumDiff for an
//      already-real foreign object whose refcount changed): the
//      caller only reads fr to accrue sumDiff, never touches
//      fr.Time.
//
// Both routes share the same map, so a single Time counter and a
// single record-save per foreign realm cover all touched objects
// (regardless of which route(s) touched it).
func (rlm *Realm) touchForeignRealm(store Store, pid PkgID) *Realm {
    if rlm.touchedForeignRealms == nil {
        rlm.touchedForeignRealms = make(map[PkgID]*Realm, 1)
    }
    if fr, ok := rlm.touchedForeignRealms[pid]; ok {
        return fr
    }
    fr := store.GetRealmByID(pid)
    if fr == nil {
        panic(fmt.Sprintf(
            "cannot resolve foreign realm %s for cross-realm finalize",
            pid))
    }
    rlm.touchedForeignRealms[pid] = fr
    return fr
}
```

#### Store-level Realm cache (`cacheRealms`)

Today `defaultStore.GetPackageRealm` (store.go:399-425) re-unmarshals
from `baseStore` on every call — there is no store-level `*Realm`
cache. Realms are persisted under a separate IAVL keyspace
(`backendRealmKey`) from objects (`backendObjectKey`), but they are
themselves stable, mutable, per-realm records. Cache them at the
store layer, parallel to `cacheObjects`:

```go
type defaultStore struct {
    ...
    cacheObjects map[ObjectID]Object
    cacheTypes   map[TypeID]Type
    cacheNodes   txlog.Map[Location, BlockNode]
    cacheRealms  map[PkgID]*Realm  // NEW: parallel to cacheObjects
    ...
}
```

Lifecycle (identical to `cacheObjects`):

- Initialized in `defaultStore` constructor (store.go:198-201) and in
  `BeginTransaction` (store.go:233-235) — `make(map[PkgID]*Realm)`.
- Populated by `GetPackageRealm` after loading from baseStore.
- Populated by `SetPackageRealm` after writing to baseStore (keeps
  the cached pointer fresh).
- Populated by `fillPackage` indirectly — it calls `GetPackageRealm`
  which populates the cache, then sets `pv.Realm` to the cached
  pointer. Because both sources go through the cache, `pv.Realm`
  and `cacheRealms[pid]` are guaranteed to be the same pointer.
- Discarded on tx abort, same as `cacheObjects` (the transaction's
  defaultStore is itself thrown away).
- **Evicted in lock-step with `cacheObjects`**: both
  `ClearObjectCache` (store.go:1115-1121) and
  `GarbageCollectObjectCache` (store.go:1123-1133) must also clear
  `cacheRealms`. Specifically:
  - `ClearObjectCache`: add `ds.cacheRealms = make(map[PkgID]*Realm)`
    alongside the existing `ds.cacheObjects = make(...)` reset.
  - `GarbageCollectObjectCache`: when a `*PackageValue` is evicted
    from `cacheObjects` by GC cycle, also `delete(ds.cacheRealms,
    pv.PkgID)` so the Realm pointer doesn't outlive its PackageValue.
    If `pv.PkgID.IsZero()` (legacy load path that bypassed
    `fillPackage`, or PV constructed before the Phase 2 plumbing
    fully landed), fall back to `delete(ds.cacheRealms,
    PkgIDFromPkgPath(pv.PkgPath))`. Without this, the `pv.Realm ==
    cacheRealms[pid]` invariant breaks: next `fillPackage(pv2)` on
    a freshly-loaded PackageValue would set `pv2.Realm` to the
    stale `cacheRealms[pid]` (if not evicted) or to a fresh load
    (if `GetPackageRealm` repopulates), producing two different
    `*Realm` pointers for the same PkgID.

GC-during-finalize invariant: `GarbageCollectObjectCache` is
invoked between blocks (between txs), never inside a
`FinalizeRealmTransaction` call. The cross-realm-finalize machinery
(`touchForeignRealm`, batch-drain) therefore cannot race with GC
eviction. Within one finalize, `rlm.touchedForeignRealms` and
`cacheRealms` hold the same `*Realm` pointers; the post-finalize
`defer rlm.touchedForeignRealms = nil` clears the per-finalize map
but the underlying `*Realm` stays in `cacheRealms` until GC time.

Updated `GetPackageRealm`:

```go
func (ds *defaultStore) GetPackageRealm(pkgPath string) *Realm {
    oid := ObjectIDFromPkgPath(pkgPath)
    if rlm, ok := ds.cacheRealms[oid.PkgID]; ok {
        return rlm
    }
    // ... existing baseStore load + amino.MustUnmarshal ...
    ds.cacheRealms[oid.PkgID] = rlm
    return rlm
}
```

Updated `SetPackageRealm`:

```go
func (ds *defaultStore) SetPackageRealm(rlm *Realm) {
    // ... existing baseStore write ...
    ds.cacheRealms[rlm.ID] = rlm
}
```

Re-introduce `Store.GetRealmByID(pid PkgID) *Realm` on the Store
interface, backed by the cache:

```go
func (ds *defaultStore) GetRealmByID(pid PkgID) *Realm {
    if rlm, ok := ds.cacheRealms[pid]; ok {
        return rlm
    }
    path := pkgPathFromPkgID(ds, pid)
    if path == "" {
        return nil
    }
    return ds.GetPackageRealm(path)  // populates cacheRealms
}
```

This makes `cacheRealms` the single source of truth for in-memory
`*Realm` pointers. `pv.Realm`, `touchedForeignRealms[pid]`, and any
future caller of `GetRealmByID(pid)` all observe the same pointer
during a tx. In-memory mutations (`fr.Time++`, `fr.sumDiff +=
delta`) are visible everywhere. Persistence is unchanged:
`SetPackageRealm(rlm)` is what writes the bumped Time to disk.

Update `FinalizeRealmTransaction` (realm.go:397+) to batch-save
touched foreign realms at the end, with a panic-safe defer-clear
of `touchedForeignRealms` at the top:

```go
func (rlm *Realm) FinalizeRealmTransaction(store Store) {
    // Always clear the per-finalize foreign-realm cache, even if
    // a panic unwinds out of finalize mid-flight. Without this,
    // a long-lived *Realm pointer (cached on the PackageValue or
    // held by Machine) could carry a stale touchedForeignRealms
    // map into a subsequent tx.
    defer func() { rlm.touchedForeignRealms = nil }()

    ...
    rlm.processNewCreatedMarks(store, 0)
    rlm.processNewDeletedMarks(store)
    rlm.processNewEscapedMarks(store, 0)
    // Persist rlm.Time if it advanced.
    if rlm.Time > startTime {
        store.SetPackageRealm(rlm)
    }
    rlm.markDirtyAncestors(store)
    ...
    rlm.saveUnsavedObjects(store)
    rlm.saveNewEscaped(store)
    rlm.removeDeletedObjects(store)

    // Batch-save foreign realms whose Time was advanced or whose
    // objects were saved during this finalize. One SetPackageRealm
    // per touched foreign realm, plus a per-realm sumDiff drain
    // into RealmStorageDiffs so storage rent attributes to the
    // owner realm.
    realmDiffs := store.RealmStorageDiffs()
    for _, fr := range rlm.touchedForeignRealms {
        realmDiffs[fr.Path] += fr.sumDiff
        fr.sumDiff = 0
        store.SetPackageRealm(fr)
    }

    rlm.clearMarks()
    ...
}
```

`store.SetObject(oo)` (realm.go:916) already routes the per-object
save by `oo.GetObjectID().PkgID`, so each foreign object lands in
its owner's IAVL keyspace without any save-path changes. The only
new save action introduced by this section is the foreign-realm-
**record** save at the batch-drain at end-of-finalize.

#### sumDiff / storage-diff attribution for foreign objects

Today `saveObject` (realm.go:902-917) does
`rlm.sumDiff += store.SetObject(oo)` — the per-object size delta is
attributed to the **executing** realm's `sumDiff`. Under interrealm v2's
storage = authority unification, storage rent should accrue to the
**owning** realm. Route the diff to the owner:

```go
func (rlm *Realm) saveObject(store Store, oo Object) {
    oid := oo.GetObjectID()
    if !oid.IsFinalized() {
        panic("unexpected non-finalized object id at save")
    }
    if oid.PkgID.IsZero() {
        panic("invariant violation: zero PkgID at save — allocation site missed allocator stamping")
    }
    // ... escape-index handling ...
    delta := store.SetObject(oo)
    if oid.PkgID == rlm.ID {
        rlm.sumDiff += delta
    } else {
        fr := rlm.touchedForeignRealms[oid.PkgID]
        if fr == nil {
            // Save-path was reached without a prior assignNewObjectID
            // path (e.g., dirty foreign object via pattern (b)).
            // Touch the foreign realm now so its record gets the
            // diff and end-of-finalize SetPackageRealm.
            fr = rlm.touchForeignRealm(store, oid.PkgID)
        }
        fr.sumDiff += delta
    }
}
```

Mirror routing in `removeDeletedObjects` (realm.go:939-943) for
the negative-delta case:

```go
func (rlm *Realm) removeDeletedObjects(store Store) {
    for _, do := range rlm.deleted {
        oid := do.GetObjectID()
        delta := store.DelObject(do)
        if oid.PkgID == rlm.ID {
            rlm.sumDiff -= delta
        } else {
            fr := rlm.touchedForeignRealms[oid.PkgID]
            if fr == nil {
                fr = rlm.touchForeignRealm(store, oid.PkgID)
            }
            fr.sumDiff -= delta
        }
    }
}
```

`store.DelObject(do)` returns a positive `int64` (the bytes freed);
the caller subtracts to record a negative delta. Same touch-then-
accrue pattern as `saveObject`. Without this change, deletes of
foreign objects would mis-attribute the negative delta to the
executing realm, breaking storage=authority symmetry.

**Invariant**: `rlm.deleted` is populated exclusively by
`decRefDeletedDescendants` (realm.go:621), which is only reached
from `processNewDeletedMarks` (realm.go:594) on objects that had
`MarkNewDeleted` called — which requires `GetIsReal() ||
GetIsNewReal()` (realm.go:352). At finalize time, those
new-real objects have already had `assignNewObjectID` run (during
`processNewCreatedMarks`), so their NewTime is set. Therefore every
`do` in `rlm.deleted` satisfies `do.GetObjectID().IsFinalized()`,
and `do.GetObjectID().PkgID` is guaranteed non-zero. The routing
code above doesn't need an explicit guard, but the invariant must
be documented at the function definition site (a comment on
`rlm.deleted`'s population in realm.go).

The end-of-finalize batch-drain
`store.RealmStorageDiffs()[fr.Path] += fr.sumDiff; fr.sumDiff = 0`
runs per touched foreign realm just before (or alongside) the
`SetPackageRealm(fr)` save.

#### Authority implications

Advancing yourrealm.Time during myrealm's finalize does **not**
grant myrealm any write authority over yourrealm's pre-existing
state. The counter is purely an ID-uniqueness device. Mutations of
existing real foreign objects continue to be gated by:

- `DidUpdate`'s `po.PkgID == rlm.ID` invariant (only mutate objects
  you own).
- `PopAsPointer2`'s readonly check + N_Readonly taint propagation.

The cross-realm finalize path only handles **new** objects (foreign
NewTime minting and saving) and existing **dirty** foreign objects
(refcount-driven re-saves like the pattern-(b) case where myrealm
appends a real foreign object to its own slice).

Add to `ObjectInfo` and the Object interface:

```go
func (oi *ObjectInfo) SetNewTime(t uint64) { oi.ID.NewTime = t }
func (oi *ObjectInfo) SetPkgID(p PkgID)    { oi.ID.PkgID = p }
```

`SetObjectID(noid)` stays for deserialization paths (`realm.go:1454+`)
that load both fields together.

#### SetIsDeleted signature simplification

Today `SetIsDeleted(x bool, mt uint64)` (ownership.go:318) takes a
`mt uint64` "deletion timestamp" parameter that the implementation
explicitly ignores (see the in-code comment at ownership.go:319-327).
The single caller passes `rlm.Time` (realm.go:620) which would
otherwise be the executing realm's clock — wrong for foreign-owned
objects being deleted in another realm's finalize.

Simplification: drop the unused parameter. Update the interface
declaration at `ownership.go:121` and the impl at `ownership.go:318`:

```go
// Object interface
SetIsDeleted(bool)

// impl
func (oi *ObjectInfo) SetIsDeleted(x bool) {
    oi.isDeleted = x
}
```

Update the single caller at `realm.go:620`:

```go
// Before:
oo.SetIsDeleted(true, rlm.Time)
// After:
oo.SetIsDeleted(true)
```

This also sidesteps the "myrealm's clock stamps yourrealm's
tombstone" semantic discrepancy from the review — no clock is
stamped, the deletion marker is just a boolean.

### Construction-time enforcement

The plan **commits to construction-time enforcement**: the allocator
panics if asked to allocate a /r/foo-typed value when the executing
realm is not /r/foo. This makes the storage=authority invariant
strict — every /r/-typed object originates inside its declaring
realm, full stop. It also collapses the per-allocation `decidePkgID`
to a trivial `return alloc.currentRealmID`, removing the
`PkgIDFromPkgPath(...)` cost from the hot path.

Disruption survey of `examples/`:

- 243 .gno files in `examples/gno.land/r/` import another /r/ realm.
- 16 files contain actual cross-realm /r/-type allocations (all
  simple field-init DTOs: `dao.UpdateRequest{...}`,
  `dao.VoteRequest{...}`, `&memberstore.Member{...}`,
  `validators.ValoperChange{...}`).
- Refactor path: add ~5 constructor functions in 4 realms
  (`r/gov/dao`, `r/gov/dao/v3/memberstore`, `r/sys/validators/v3`,
  `r/gov/dao/v3/impl`); update each call site to use the
  constructor. No `new()`, no `make()`, no nested foreign /r/
  composites. Mechanical.

The DTO-request pattern in `r/gov/dao` (caller builds a Request,
ships via `cross`) is the main idiom affected. Wrapping via
`dao.NewVoteRequest(...)` is a small ergonomic cost in exchange
for the invariant. Constructors run with borrow rule #1 to their
declaring realm, return a /r/-typed value with the correct PkgID
to the caller, who then passes it on.

### Where the construction-time check fires (implementation refinement)

The original spec stated the check fires inside every allocator
constructor (`Alloc.NewStruct` / `NewListArray` / `NewMap` /
`NewHeapItem`). Implementation discovered this is too aggressive:
internal default-init paths legitimately allocate /r/-typed values
outside their declaring realm. Concrete cases:

- **Return-slot init in `doOpCall`** (op_call.go:292): when a /r/foo
  function returns a `/r/bar.T`, the call setup defaults the return
  slot to a zero `bar.T` via `defaultTypedValue` → `defaultStructValue`
  → `Alloc.NewStruct(bar.T, ...)`. This fires from the caller's realm
  context but predates the borrow that would activate `/r/bar`.
- **Var declarations**: `var x bar.T` in /r/foo's code allocates a
  zero value before any /r/bar code runs.
- **Generic copies**: `Copy(alloc, t)` for /r/bar values passing
  through /r/foo code calls `Alloc.NewStruct(bar.T, ...)` with /r/foo
  as currentRealmID, then overwrites PkgID with the source's.

The fix: the check fires only at **user-visible construction sites**
that mint authority — i.e. where the user wrote an expression that
constructs a non-zero value of the type:

- Composite literals: `doOpStructLit`, `doOpArrayLit`, `doOpMapLit`,
  `doOpSliceLit`, `doOpSliceLit2` (`op_expressions.go`).
- The `new()` and `make()` uverse builtins.

`stampPkgID` continues to fire inside every constructor — every
allocator-emitted object gets a PkgID stamp (currentRealmID by
default; overwritten by `Copy` and friends to propagate authority).
Only the **panic gate** moved to user-visible sites; the **stamping
mechanism** remains allocator-level.

Coverage gap: piecewise construction (`var x foreign.T; x.Field =
...`) bypasses the check. Acceptable because (1) composite literal
is the main idiom in Gno code, (2) field writes on real objects
are gated by tainted-readonly checks at the SET op, (3) field
writes on heap (unreal) objects are only visible to the constructing
realm — they can't be persisted into cross-realm storage without
going through borrow rule #1 at the cross-call boundary.

### Type-cached PkgID

`DeclaredType` and `StructType` both carry a `PkgPath` field. Under
the construction-time check (and any future Type→PkgID resolution),
the conversion path `PkgPath → PkgID` is `sync.Map.Load`-backed
(realm.go:96-114) but still costs an atomic lookup, interface
unboxing, and pointer dereference per call. Cache the resolved
PkgID on the Type itself:

```go
type DeclaredType struct {
    PkgPath string
    ...
    pkgID   PkgID  // lazy cache; zero means uncomputed
}

func (dt *DeclaredType) GetPkgID() PkgID {
    if dt.pkgID.IsZero() {
        dt.pkgID = PkgIDFromPkgPath(dt.PkgPath)
    }
    return dt.pkgID
}

type StructType struct {
    PkgPath string
    Fields  []FieldType
    ...
    pkgID   PkgID  // lazy cache; zero means uncomputed
}

func (st *StructType) GetPkgID() PkgID {
    if st.pkgID.IsZero() {
        st.pkgID = PkgIDFromPkgPath(st.PkgPath)
    }
    return st.pkgID
}
```

Both fields are unexported, so amino serialization skips them. Types
are long-lived and shared across the VM, so the cache populates once
and stays warm. Other Type implementations (ArrayType, SliceType,
MapType, FuncType, ChanType) have no meaningful PkgPath; their
`GetPkgID` would return `PkgID{}` and is unnecessary.

Add a helper for the construction-time check that walks
Pointer/Declared/Struct wrappers and reads the cached PkgID:

```go
// Returns the declaring-realm PkgID for /r/-declared named types,
// or PkgID{} for anonymous composites and non-/r/ types.
func getDeclaredPkgID(t Type) PkgID {
    for {
        switch tt := t.(type) {
        case *PointerType:
            t = tt.Elt
            continue
        case *DeclaredType:
            return tt.GetPkgID()
        case *StructType:
            return tt.GetPkgID()
        default:
            return PkgID{}
        }
    }
}
```

Add a `PkgID.IsRealmPkg()` predicate using the existing flag bits:

```go
// IsRealmPkg returns true for /r/-declared packages (not stdlib,
// not /p/, not internal). Reads the IsImmutablePkg flag bit.
func (pid PkgID) IsRealmPkg() bool {
    return !pid.IsZero() && !pid.IsImmutablePkg()
}
```

### Allocator API and PkgID assignment

```go
func (alloc *Allocator) checkConstructionTime(t Type) {
    pkgID := getDeclaredPkgID(t)
    if !pkgID.IsRealmPkg() {
        return  // anonymous, primitive, /p/, stdlib — no restriction
    }
    if pkgID != alloc.currentRealmID {
        panic(fmt.Sprintf(
            "cannot allocate %s in realm %s",
            t.String(), alloc.currentRealmPath))
    }
}
```

`checkConstructionTime` is **not** called from generic allocator
constructors. Per the refinement above, it fires only at user-
visible construction sites — composite-literal handlers in
`op_expressions.go` and the `new()`/`make()` uverse builtins.
Generic constructors keep the `t Type` parameter for API
consistency (and so future Type-driven behavior can hook in) but
no longer enforce the construction-time check internally.

Every allocator constructor calls `alloc.stampPkgID(&obj.ObjectInfo)`
unconditionally — `obj.ID.PkgID = alloc.currentRealmID` for every
new Object. `Copy` and `CopyForReceiver` then overwrite that stamp
with the source's PkgID to propagate authority across copies; the
per-allocation `decidePkgID` helper from the original spec is gone.

Thread `Type` through allocator constructors:

- `NewStruct(t Type, fields []TypedValue) *StructValue`
- `NewListArray(t Type, n int) *ArrayValue`
- `NewListArray2(t Type, l, c int) *ArrayValue`
- `NewDataArray(t Type, n int) *ArrayValue`
- `NewMap(t Type, size int) *MapValue`
- `NewBlock(source BlockNode, parent *Block) *Block` — Block has no
  user type; uses currentRealmID directly (Block belongs to executing
  package's realm).
- `NewHeapItem(t Type, tv TypedValue) *HeapItemValue`
- `NewPackageValue(pn *PackageNode) *PackageValue` — uses the package's
  own PkgPath (a PackageValue belongs to its own realm).

For off-allocator construction sites (raw `&XxxValue{}` literals),
each needs to set `ID.PkgID` after construction or be refactored to
use the allocator. Exhaustive enumeration (current HEAD line numbers;
re-cite at implementation time). `values_export.go` is excluded —
its constructors are client-facing API marshaling, not chain state.

Inside `alloc.go` (constructor-internal inner literals — stamp on
the inner object, not just the outer one returned by the
constructor):

- `alloc.go:464` `&ArrayValue{}` as `SliceValue.Base` inside
  `NewSliceFromList` — PkgID = currentRealmID. (Slice and its base
  array share the same realm — they're allocated as one unit. The
  outer SliceValue isn't itself an Object, but the Base ArrayValue
  is.)
- `alloc.go:481` `&ArrayValue{}` as `SliceValue.Base` inside
  `NewSliceFromData` — PkgID = currentRealmID. Same reasoning.
- `alloc.go:512` `&MapValue{}` inside `NewMap` — covered by adding
  `t Type` to `NewMap`; stamp here.
- `alloc.go:521-522` `&PackageValue{}` + inner `&Block{}` inside
  `NewPackageValue` — PkgID = `pn.GetPkgID()` for both. Package and
  its top-level block share the package's authority. See "Cached
  PackageValue.PkgID" below — the PackageNode caches the PkgID once
  at construction.
- `alloc.go:547` `&HeapItemValue{}` inside `NewHeapItem` — covered
  by adding `t Type` to `NewHeapItem`.

In `op_expressions.go` / `op_exec.go` (executor literals — have
`*Machine` or `*Allocator` in scope):

- `op_expressions.go:708` `doOpFuncLit` `&FuncValue{}` — closures:
  PkgID = `m.Alloc.currentRealmID`. FuncType has no declaring-realm
  semantics; the closure's identity belongs to wherever it was
  evaluated.
- `op_exec.go:133` `&HeapItemValue{}` for-loop init — PkgID =
  `m.Alloc.currentRealmID`.

In `values.go` (mixed — some have `*Allocator`, some don't):

- `values.go:515` `FuncValue.Copy` — takes `*Allocator`. Fresh
  FuncValue preserves PkgID from source: `cp.ID.PkgID = fv.ID.PkgID`
  (the copy is a re-binding, not a re-creation in a new realm).
  Do **not** stamp from `alloc.currentRealmID`.
- `values.go:1716` `DefineToBlock` `&HeapItemValue{Value: other}` —
  **API change required.** This method on `*TypedValue` has no
  `*Machine` or `*Allocator` access. Two options:
  - (a) Add `currentRealmID PkgID` parameter to `DefineToBlock`,
    push through every call site (currently called from blocks/
    machine.go); stamp `hiv.ID.PkgID = currentRealmID`.
  - (b) Refactor `DefineToBlock` to take `*Allocator`; read
    `alloc.currentRealmID` inside. Cleaner but touches more sites.
  - Pick (b) for consistency with the rest of the plumbing.
- `values.go:1893` `&BoundMethodValue{}` (ptr-method binding) —
  PkgID = currentRealmID. The bound receiver carries its own PkgID
  independently; the BoundMethodValue wrapper belongs to the realm
  doing the binding.
- `values.go:1930` `&BoundMethodValue{}` (value-method binding) —
  same as 1893.
- `values.go:2377` `&Block{}` in package-init `NewBlock` body —
  `alloc.NewBlock` already takes `*Allocator`; stamp PkgID =
  `alloc.currentRealmID` on the returned Block. (Note: this literal
  is inside the global `NewBlock` function called from
  `Allocator.NewBlock` at `alloc.go:535-538`. Stamping at this site
  covers both call paths.)
- `values.go:2501` `&HeapItemValue{}` in `GetPointerToMaybeHeapDefine`
  — PkgID = currentRealmID. This method has `*Allocator` in scope
  through the `alloc` parameter.

In `nodes.go` / `preprocess.go` / `uverse.go` (preprocess-time and
runtime-boot literals — not chain-critical, but must still carry a
correct PkgID for the borrow rule to be coherent):

- `nodes.go:1362-1363` `&PackageValue{}` + inner `&Block{}` in
  `PrepareNewValues` — PkgID = `pv.PkgID` (the cached field; see
  "Cached PackageValue.PkgID" below). Preprocess-time; the package's
  own realm.
- `preprocess.go:4035-4036, 4201-4205` `&PackageValue{}` + inner
  `&Block{}` in package definition / init paths — PkgID =
  `pn.GetPkgID()` (cached on PackageNode).
- `preprocess.go:5580` `&FuncValue{}` inside `TryDefineMethod` —
  PkgID = `pn.GetPkgID()`. The `PkgPath` field is already set; the
  new step is also stamping `ID.PkgID`.
- `preprocess.go:5607` `&FuncValue{}` for top-level function
  definition — same.
- `uverse.go:198` `&StructValue{}` in `newRealmHIVPointer` (the
  Realm-handle struct) — PkgID = the runtime realm being constructed
  for. Inspect the call site context: this constructs the per-tx
  Realm representation, so PkgID = the active package's PkgID at
  bootstrap (read from `m.Realm.ID`).
- `uverse.go:205` `&HeapItemValue{}` wrapping the above — same
  PkgID as 198.
- `uverse.go:505` `&PackageValue{}` (uverse stub package itself) —
  PkgID = the cached uverse PkgID. Compute once at boot, cache on
  the uverse PackageValue.

In `realm.go` deserialization paths (`fillType` etc., ~lines
1454-1575) — `&ArrayValue{}`, `&StructValue{}`, `&FuncValue{}`,
`&BoundMethodValue{}`, `&MapValue{}`, `&PackageValue{}`, `&Block{}`,
`&HeapItemValue{}` constructed from `RefValue` — **no stamping
needed**: PkgID comes from the stored ObjectID that the deserializer
is materializing. The `ObjectInfo` is copied in from the persisted
form; the inner-literal constructors are wrapped by setting the
recovered `ObjectInfo` immediately after construction.

### ObjectInfo.Copy() and PkgID propagation

There are two distinct "copy" paths through `*Value` types and the
plan needs each to handle PkgID with explicit intent:

**Path A — `ObjectInfo.Copy()` for serialization** (`ownership.go:182`).
Today this preserves the full `ID` (PkgID + NewTime) and is called
only from `copyValueWithRefs` (`realm.go:1419`) during the persist-
to-store marshaling pass. This is **correct under interrealm v2** — the
persisted identity must round-trip; PkgID is the authority realm
and survives serialization unchanged. No change needed.

**Path B — `*Value.Copy(alloc)` for in-memory copies**
(`values.go:324` ArrayValue.Copy, `values.go:450` StructValue.Copy).
These allocate a fresh object via `alloc.NewListArray` /
`alloc.NewStruct`, which under interrealm v2 stamps PkgID =
`alloc.currentRealmID` on the fresh object. The source's
ObjectInfo is **not** inherited. This matches the design intent:
copying a value into the current realm makes the copy live in the
current realm.

The construction-time check must wire through Path B for it to be
sound. `StructValue.Copy(alloc)` calls `alloc.NewStruct(fields)` —
under interrealm v2 this needs a `Type` parameter for the
construction-time check. Add `t Type` parameter to `Copy`: `(sv *StructValue)
Copy(alloc *Allocator, t Type)`. Threads the type from the caller
(who has `tv.T` in scope).

`FuncValue.Copy(alloc)` is **not** Path B — although it takes an
`*Allocator`, it uses `alloc.AllocateFunc()` only for size
accounting and constructs the FuncValue as a literal (values.go:513),
bypassing `alloc.NewStruct`/`NewListArray` and therefore the
construction-time check. It is correctly handled in the off-allocator
list above: the copy preserves source PkgID (`cp.ID.PkgID =
fv.ID.PkgID`) because a closure copy is a re-binding, not a
re-creation.

#### Receiver-copy exception (`CopyForReceiver`)

Value-method dispatch is a Path-B exception that must preserve
source PkgID. In `VPValMethod` (values.go:1885 area), the receiver
TypedValue is copied via `dtv.Copy(alloc)` *before* PushFrameCall
fires. Under the generic Path B rule, the copy would get stamped
`PkgID = alloc.currentRealmID` — laundering the source's authority
into the caller's realm. This re-opens the /p/-attacker-via-
interface gap for value-method dispatch:

1. /r/attacker defines `type Evil struct { Bar *Victim }` (Evil is
   /p/-typed or /r/attacker-typed satisfying victim's interface)
   with value-method `func (e Evil) M() { e.Bar.field = x }`.
2. Attacker pre-plants `evil.Bar = victimRef` under attacker
   authority (DidUpdate passes: po=evil, po.PkgID=attacker,
   m.Realm=attacker).
3. Attacker passes Evil to victim (cross-call return).
4. Victim calls `iface.M()`. Under generic Path B, copy.PkgID =
   victim. borrow rule #2 reads PkgID == m.Realm.ID → no borrow. M runs
   as victim. `e.Bar.field = x` → po=victimObj (PkgID=victim),
   m.Realm=victim → DidUpdate passes. **Attack succeeds.**

To close this, value-method receiver copying must preserve source
PkgID **and bypass the construction-time check**. The naive approach
of "Copy then re-stamp" doesn't work because `Copy(alloc, t)` invokes
`alloc.NewStruct(t, fields)` which runs `checkConstructionTime(t)` —
for a cross-realm `/r/foo.Thing` receiver, the check sees
`pkgID=/r/foo, currentRealmID=victim` and **panics before** the
re-stamp runs. This would also break the legitimate `External /r/
stored in caller's realm, method mutates self` case (victim does
`t := /r/foo.NewThing(); t.ValueRead()` → VPValMethod copies t →
construction-time check panics).

The fix is to swap `alloc.currentRealmID` to the source's PkgID for
the duration of the receiver copy. Construction-time checks at every
depth then pass naturally because `currentRealmID` matches the
type's declaring realm at every nested field. The fresh
`*StructValue` keeps source PkgID, and nested foreign-struct fields
copy through their normal allocator path (no special unchecked
variant needed).

```go
// CopyForReceiver makes a transient receiver copy that preserves
// the source's authority PkgID (and NewTime). The only call site is
// VPValMethod at values.go:1885.
//
// The copy is transient (never finalized, never persisted, no
// DidUpdate fires on it via realm.go:239's !GetIsReal() skip), so
// it does not "mint foreign authority" — it routes to existing
// authority for dispatch. To prevent the generic construction-time
// check from panicking on cross-realm receivers (and on their
// nested foreign struct fields), temporarily swap
// alloc.currentRealmID to match the source's declaring realm for
// the duration of the copy.
func (sv *StructValue) CopyForReceiver(alloc *Allocator) *StructValue {
    saved := alloc.currentRealmID
    alloc.currentRealmID = sv.ObjectInfo.ID.PkgID
    defer func() { alloc.currentRealmID = saved }()

    fields := alloc.NewStructFields(len(sv.Fields))
    for i, field := range sv.Fields {
        fields[i] = field.Copy(alloc)
    }
    cp := alloc.NewStruct(sv.GetType(), fields)
    // NewStruct stamped cp.ID.PkgID = alloc.currentRealmID =
    // source PkgID, which matches what we want. NewTime is still 0
    // (cp is unreal); copy it too so cp.GetIsReal() reflects the
    // source's finalization state.
    cp.ObjectInfo.ID.NewTime = sv.ObjectInfo.ID.NewTime
    return cp
}

// Similar variant for *ArrayValue.
```

Update the **single** call site in VPValMethod (values.go:1885 area)
to use `CopyForReceiver(alloc)` instead of `Copy(alloc, t)`. The
other ~13 generic Copy call sites (general-purpose in-memory copies
via `unrefCopy`, package-init paths, etc.) remain on the generic
checked-stamping rule.

Receiver-copy is the only legitimate exception to the construction-time check.
`*MapValue`, `*FuncValue`, `*SliceValue` cannot be value-method
receivers (per Go semantics — maps are reference types, funcs use
pointer-style dispatch), so the variant is needed only for
`*StructValue` and `*ArrayValue`.

Trace under the fix:

- **Attacker attack** (the gap being closed): copy preserves
  PkgID=attacker. borrow rule #2 reads PkgID != m.Realm.ID → borrows
  m.Realm to attacker. Inside M, `e.Bar.field = x` — PopAsPointer2's
  IsReadonly check fires (e.Bar's underlying has PkgID=victim,
  m.Realm=attacker, so IsReadonly returns true). **Panic. Attack
  closed.**

- **Legitimate cross-realm value-method dispatch**: victim does
  `t := /r/foo.NewThing(); t.ValueRead()`. VPValMethod calls
  CopyForReceiver(alloc) — no construction-time check, copy retains PkgID=/r/foo.
  borrow rule #1 fires (`ValueRead` is /r/foo-declared) so m.Realm shifts
  to /r/foo. Inside ValueRead, the body operates on the receiver
  copy (transient, PkgID=/r/foo matches m.Realm.ID). Reads succeed;
  writes to the copy itself don't trip DidUpdate (copy is unreal).
  Writes to reachable real /r/foo state succeed (PkgID matches).
  **Works correctly.**

The receiver copy retains source PkgID even though it's a fresh
*StructValue. This is sound because the copy is transient (never
finalized, never persisted, no DidUpdate fires on it because
`!GetIsReal()` skips at realm.go:239), so the PkgID is purely
metadata for authority dispatch — not a claim of ownership over
new state.

Allocator gains a `currentRealmID` field synced from `m.Realm` by the
`Machine.setRealm(r)` helper. Every `m.Realm = X` assignment in
machine.go (~10 sites) routes through `m.setRealm(X)` so the allocator
stays in sync.

### Cached PackageValue.PkgID and PackageNode.PkgID

The cold-path off-allocator sites above (`nodes.go:1362`,
`preprocess.go:4035`, `preprocess.go:5580`, etc.) each have a
`*PackageValue` or `*PackageNode` in scope. Rather than re-resolving
`PkgIDFromPkgPath(pv.PkgPath)` at every site, cache the PkgID once at
construction:

```go
type PackageValue struct {
    ObjectInfo
    Block      Value
    PkgName    Name
    PkgPath    string
    PkgID      PkgID  `json:"-"`  // re-derived on load from PkgPath
    ...
}

type PackageNode struct {
    ...
    PkgPath string
    pkgID   PkgID  // lazy cache, unexported, not serialized
}

func (pn *PackageNode) GetPkgID() PkgID {
    if pn.pkgID.IsZero() {
        pn.pkgID = PkgIDFromPkgPath(pn.PkgPath)
    }
    return pn.pkgID
}
```

**The `json:"-"` tag on `PackageValue.PkgID` is required.** Amino's
encoder reads field tags via `field.Tag.Get("json")`
(`tm2/pkg/amino/codec.go:777-787`) and skips fields tagged `json:"-"`.
Without the tag, an exported `PkgID` field would be serialized into
the wire format, breaking the "re-derived on load" claim and
introducing a redundant on-chain field. Precedent for `json:"-"` on
existing exported PackageValue fields: `values.go:804`
(`Realm *Realm \`json:"-"\``).

The PackageValue's `PkgID` is set in every constructor path:
- `alloc.go:518` `NewPackageValue(pn)`: `pv.PkgID = pn.GetPkgID()`.
- `nodes.go:1362`, `preprocess.go:4035`, `preprocess.go:4201`,
  `uverse.go:505`: each sets `pv.PkgID` immediately after the
  literal.
- `realm.go` deserialization path: amino skips the tagged field, so
  it's the zero value after unmarshal. `fillPackage`
  (`store.go:544`) sets `pv.PkgID = PkgIDFromPkgPath(pv.PkgPath)`
  (one-time cost per package load) right after the unmarshal — same
  pattern `fillPackage` uses for `pv.Realm` today.

`Realm.ID` (realm.go:155) is already the realm's PkgID — that's
the existing pattern. PackageValue gains the same convenience.

### Authorization checks

`DidUpdate`'s existing PkgID check is unchanged:

```go
if po.GetObjectID().PkgID != rlm.ID {
    panic("invariant violation...")
}
```

This naturally enforces the authority rule under the new model (PkgID
is the authority realm).

`Machine.isExternalRealm` also unchanged in its core comparison but
its meaning shifts: foreign now means "not my authority," not
"physically stored elsewhere." Update the comment.

### Readonly relaxation (narrow: Case 1 only)

Drop Case 1 (source-side readonly check) in `doOpConvert` — and **only**
Case 1. The rest of the N_Readonly taint machinery must be preserved.

```go
// REMOVE the entire Case 1 block (op_expressions.go ~741-752):
if xv.T != nil && !xv.T.IsImmutable() && m.IsReadonly(&xv) {
    if xvdt, ok := xv.T.(*DeclaredType); ok &&
        xvdt.PkgPath == m.Realm.Path {
        // Except allow if xv.T is m.Realm.
    } else {
        panic("illegal conversion of readonly or externally stored value")
    }
}
```

Why this is safe: `ConvertTo` (values_conversions.go:16) operates in
place on `*TypedValue` via `tv.T = ...` / `tv.SetInt(...)` / etc. It
never touches `tv.N`, so the `N_Readonly` bit survives conversion.
Any later write through the converted value is still caught at the
write site (PopAsPointer2, append/copy/delete in uverse.go). Case 1
was redundant defense-in-depth that fired on legitimate read patterns
(borrowed-method bodies type-asserting / converting caller arguments).

Keep Case 2 (target-side type forgery) unchanged — it prevents
constructing values of foreign /r/-declared types from outside.

**Do not** also drop the `m.IsReadonly` checks at append/copy/delete
sites in `uverse.go` (~lines 697, 905, 941, 994), and **do not** drop
the `SetReadonly` propagation in read ops in `op_expressions.go`
(doOpSelector, doOpIndex, doOpStar, doOpRef, doOpSlice). The
N_Readonly taint is what blocks the round-trip attack:

```go
// /r/myrealm
var GlobalBytes []byte
func MutateBytes(b []byte) { b[0] = 0xff }

// /r/attacker
import "myrealm"
func Attack() {
    myrealm.MutateBytes(myrealm.GlobalBytes)
}
```

Trace under interrealm v2 *with* readonly machinery preserved:

1. Attacker reads `myrealm.GlobalBytes`. m.Realm = /r/attacker;
   GlobalBytes underlying slice's PkgID = /r/myrealm. doOpSelector
   computes `ro = m.IsReadonly(base)` — base's PkgID ≠ m.Realm.ID, so
   `ro = true`. Result tv gets `N_Readonly` bit set.
2. Attacker calls `myrealm.MutateBytes(taintedSlice)`. Arg copy
   preserves `N_Readonly`. borrow rule #1 shifts m.Realm to /r/myrealm.
3. Inside MutateBytes, `b[0] = 0xff` runs PopAsPointer2 IndexExpr.
   `ro = m.IsReadonly(xv)` returns true via the N_Readonly bit check
   in `IsReadonlyBy` (ownership.go:424+), independent of the PkgID
   comparison.
4. Write panics: "cannot modify readonly object."

Trace under interrealm v2 *without* readonly machinery (the version that was
considered and rejected):

1. Attacker reads `myrealm.GlobalBytes`. No taint propagation. Result
   tv has no N_Readonly bit.
2. Attacker calls `myrealm.MutateBytes(slice)`. borrow rule #1 shifts
   m.Realm to /r/myrealm.
3. `b[0] = 0xff`. DidUpdate sees `slice.underlying.PkgID = /r/myrealm`
   and `rlm.ID = /r/myrealm` — passes. Write succeeds. Attack
   completes.

So: interrealm v2's borrow rule alone does not prevent the round-trip case
when the attacker passes the victim's own data back to a victim
method. The N_Readonly bit is the load-bearing piece of that defense
and must be kept. The borrow rule + N_Readonly taint together close
the .Title()-attack class.

## Compatibility

interrealm v2 is a **chain-breaking change**. It assumes a fresh
genesis. No migration path is provided for existing on-chain state:
interrealm v2 changes `ObjectID.PkgID` semantics from "storage realm
determined at link time" to "authority realm stamped at allocation
time," and existing persisted objects carry the old semantics in
their stored ObjectIDs. Re-interpreting that state under the new
model would mis-attribute authority and break the storage=authority
invariant. A new chain is required.

## What this delivers

| Scenario | Before | After |
|---|---|---|
| /r/-attacker via interface | closed | closed |
| /r/-attacker via direct call | closed | closed |
| /r/-attacker via top-level function call | closed | closed |
| /p/-attacker via interface (object recv, attacker-constructed) | open | **closed** |
| /p/-attacker via interface, **value-method receiver** (pre-planted victim ref) | open | **closed** (CopyForReceiver preserves source PkgID) |
| External /r/ stored in caller's realm, method mutates self | broken | **works** |
| /r/-method body type-asserts/converts caller args | broken | **works** |
| /p/-helper reads caller args | works | works |
| Round-trip: `victim.MutateBytes(victim.GlobalBytes)` from attacker | closed (N_Readonly taint) | closed (N_Readonly taint) |
| Type forgery via foreign /r/ declared type | closed | closed |
| Cross-realm object attachment: myrealm holds yourrealm-allocated objects, both get finalized correctly in one tx | partial (panic-prone) | **works** (batched foreign-realm record save in myrealm.FinalizeRealmTransaction) |

The /p/-factory laundering case (victim imports /p/factory which
transitively imports /p/attacker; factory's MakeEvil returns Evil with
allocation-context PkgID = victim) remains open by design. This is
the importer's responsibility — importing a /p/ package is an explicit
trust decision, and the security guide's canonical-impl-allowlist
pattern is the recommended defense for cases where authors want
runtime verification.

## Implementation notes

- **Phase 2 depends on Phase 1**, not the other way around.
  - Phase 1 alone (without Phase 2) is safe: `IsZero` is redefined
    to check both fields, `GetIsReal` is redefined to `NewTime != 0`,
    and `IsFinalized` is added. Because Phase 2 hasn't shipped yet,
    no object is in the "allocated but not finalized" state
    (`PkgID set, NewTime == 0`) — every existing object has either
    both fields zero or both fields set, so the redefined predicates
    return the same answer as today's. Tests pass.
  - Phase 2 alone (without Phase 1) is **broken**: the allocator
    starts stamping PkgID at construction time, producing the
    `PkgID set, NewTime == 0` state. The old `GetIsReal` returns
    true on these objects (`!ID.IsZero()` was true once PkgID was
    set), so newly-allocated objects are prematurely "real" — they
    pass the DidUpdate early-return, get treated as persisted, and
    finalization breaks.
  - Land Phase 1 first; Phase 2 can follow in the same PR or a
    follow-on. Do not land Phase 2 alone.

- The off-allocator construction sites need careful audit. Missing
  any of them means that allocation path produces objects with
  PkgID=zero, which falls into the legacy / pre-stamp branch of
  `assignNewObjectID` and would silently work under the existing
  realm's authority — defeating the new model for those objects.

- Storage griefing (anyone can grow /r/foo's storage by calling its
  constructor) is a known consequence of the unified
  storage=authority model. /r/foo authors must rate-limit or gate
  construction through their own logic. Not addressed at the runtime
  level.

- /r/-typed allocation outside the declaring realm is handled by
  construction-time enforcement (see the section of that name above).
  The allocator panics if asked to allocate a /r/foo-typed value
  when `currentRealmID != /r/foo`. This forces every /r/-typed value
  to originate inside its declaring realm via that realm's
  constructor function. The `decidePkgID` per-allocation helper
  collapses to `return alloc.currentRealmID` because the check has
  already enforced equality.

## Future considerations (not part of this implementation)

The following are notes for future work, not part of this plan:

- **Restrict cross-realm allocations at preprocess time** — promote
  the runtime construction-time check to a static (preprocess-time)
  rejection of composite literals / `make` / `new` expressions whose
  type is declared in a foreign /r/ realm. Better error messages and
  catches the violation before allocation. Not addressed here; the
  runtime enforcement is sufficient for safety and the
  16-file refactor in `examples/` is the same either way.

- **Restrict primitive methods** — at preprocess time, reject method
  declarations on primitive-kind types (`type Foo int; func (f Foo)
  Method(...)`) that take pointer-typed or interface-typed arguments.
  This would close the .Title()-attack class against primitive
  receivers by removing the dangerous signature shape at the language
  level. The survey of existing examples found zero current usages of
  this pattern, so the restriction is forward-only. Not addressed
  here; the current threat model and the canonical-impl-allowlist
  recommendation for interface boundaries are considered sufficient
  for the primitive case.

- **Restrict nil receivers** - :unreal? same with primitve?
