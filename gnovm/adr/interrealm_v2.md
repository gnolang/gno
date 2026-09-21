# Gno Interrealm Specification v2

The purpose of "Gno Interrealm Specification v2" is to improve the security of
Gno code by (1) making interrealm cross calls more explicit and restrictive
(`cross(cur)`) so as to prevent unexpected unauthorized cross calls from a
victim's realm context when the victim calls an arbitrary (potentially attacker
provided) method or callable, and (2) to make the storage (borrow) realm context
depend not on which storage realm objects get persisted, but rather depend on
where it is constructed (and for /r/ types and callables this is always the
realm where declared). 

## `cur realm` as Authority Token: `cross` --> `cross(cur)`

For case (1) consider the following:

```go
    // PKGPATH: /r/victim
    type Proposal interface {
        Title() string
        ...
    }

    func AddProposal(cur realm, prop Proposal) {
        ...
        title := prop.Title()
        ...
    }
```

`AddProposal(realm, Proposal)` is a crossing function (first argument is `cur
realm`) that is meant to change agency from an external caller realm to
/r/victim. While accepting an interface parameter `prop` and calling its method
appears safe because it takes no arguments besides the caller provided receiver
`prop` and returns a string, in the Gno Interrealm Specification v1 an attacker
could implement `.Title()` as follows:

```go
    // PKGPATH: /r/attacker
    import /r/grc20banker

    type Exploiter struct{...}
    func (_ Exploiter) Title() string {
        // In .SendAllCoins() runtime.Previous().PkgPath() is '/r/victim',
        // so the sender is '/r/victim' implied, to '/r/attacker'.
        grc20banker.SendAllCoins(cross, '/r/attacker')
        return "pwnd"
    }
```

In the previous Gno interrealm spec v1 when an `Exploiter{}` value is passed
to `/r/victim.AddProposal()` the victim loses all the tokens from /r/grc20banker
because `AddProposal()` runs in the crossing & storage realm context of the
victim. From the perspective of `/r/grc20banker.SendAllCoins(cur realm, from, to
string)` it only knows that `runtime.PreviousRealm().PkgPath()` is `/r/victim`
because `/r/victim.AddProposal()` was the last crossing function crossed into.

In the new Gno interrealm spec v2 it is not legal to call a crossing function
with the bare `cross` keyword -- the `cross` keyword is no longer a virtual
uverse/builtin `realm` type, but is a function of type `func(realm) realm`, so
must be called like `cross(cur)`.  Since the attacker's Title() never received
the victim's cur, the attacker can't call `grc20banker.SendAllCoins(cross(?),
...)` -- there's no realm token to pass.

As long as the attacker never receives the `cur realm` capability token from the
victim's realm context it cannot spoof the victim. The attacker cannot even call
`grc20banker.SendAllCoins()` unless it receives a current first `cur realm` or
secondary `rlm realm` parameter either from the initial EOA user transaction, or
from another realm; and the `realm` object in v2 (as in v1) cannot be persisted
either.

## runtime.CurrentRealm() --> unsafe.CurrentRealm()

In v1 `runtime.CurrentRealm()`, `.PreviousRealm()`, `.OriginCaller()` returned
the `stdlib.Realm` type (which is similar but not the same as the builtin
`realm` type) which carried over from the pre-v1 Gno design. Much Gno code was
using it to track caller/callee agency for token transfers, transfer
limitations, and authority and permissions logic.

In v2 these runtime functions have migrated to the `runtime/unsafe` package, so
they are now as follows:

 * runtime.CurrentRealm()  -> unsafe.CurrentRealm()
 * runtime.PreviousRealm() -> unsafe.PreviousRealm()
 * runtime.OriginCaller()  -> unsafe.OriginCaller()
 * runtime.OriginSend()    -> unsafe.OriginSend()

Instead of `unsafe.CurrentRealm().Address()` or `.PreviousRealm().PkgPath()` the
safe way to get these realm context values is to call `cur.Address()` or
`cur.Previous().PkgPath()`. 

```go
    // PKGPATH: /r/grc20banker
    func SendAllCoins(cur realm, to address) {
        // from := unsafe.PreviousRealm().Address() <-- bad
        from := cur.Previous().Address() <-- good
        ...
    }
```

This is less of a problem in the above example, because SendAllCoins is a
crossing function which forces `unsafe.PreviousRealm().Address()` to be
identical to `cur.Previous().Address()`. The problem is that it is too tempting
to write code that uses `.CurrentRealm()` or `.PreviousRealm()` in non-crossing
functions which can be exploited by an attacker (such as in the attacker's 
implementation of `.Title()`), which may call other non-crossing functions.

```go
    // PKGPATH: /r/attacker
    import /r/otherbanker

    type Exploiter struct{...}
    func (ex Exploiter) Title() string {
        // Here .GetBanker() returns an object with methods
        // that are vulnerable, internally using `runtime.CurrentRealm()`.
        otherbanker.GetBanker().SendAllCoins("/r/attacker")
        return "pwnd"
    }
```

In v2 `unsafe.CurrentRealm()` and `unsafe.PreviousRealm()` may still be used for
debugging and diagnostic purposes but the usage of these functions in
determining authority or agency should be considered an explicit bug.

## V2 New Borrow Rules

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
receiver-driven borrow is replaced by three new borrow rules. It does not matter
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
    when the receiver was *constructed* (and will be stored if persisted).

/p/-declared things are different: 

/p/ packages themselves are immutable (no realm state, library semantics), but
/p/-declared types used as values acquire a /r/ stamp (on ObjectInfo.PkgID) at
construction and are then mutable by that /r/ — see borrow rules #2 and #3. In
v1 there was no such construction-time ObjectInfo.PkgID stamping. In v2 the
construction stamp is the constructing realm's PkgID, and copies are stamped
type-driven (stampPkgID's split rule, #5706/#5747): a /r/-declared type keeps
its declared /r/ owner across copies, but a /p/-declared value's copy takes the
copying realm's PkgID — it does NOT retain the source /r/. That is what makes
cross-realm in-place /p/ arithmetic work (#5736).

  - Borrow rule #3: if fv is a closure (FuncLit) declared in a /p/ package,
    borrow to the realm context that was active when the closure was
    *constructed* (and will be stored if persisted).

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

## Some Caveats

### Foreign Type Value Caveat 

```go
    // PKGPATH: gno.land/r/foreignRealm
    type MyStruct struct {
        Field string
    }

    func (ms *MyStruct) Modify() {
        ms.Field += "_modified"
    }
```

```go
    // PKGPATH: gno.land/r/myRealm
    var x foreignRealm.MyStruct     <-- ObjectInfo.PkgID = /r/foreignRealm
    x.Field = "..."                 <-- write fail, /r/myRealm != /r/foreignRealm
```

The zero value is stamped with ObjectInfo.PkgID = /r/foreignRealm upon (default)
construction and is persisted under /r/foreignRealm, even though the slot lives
in /r/myRealm's package block.

```go
    // PKGPATH: gno.land/r/myRealm
    var x foreignRealm.MyStruct     <-- ObjectInfo.PkgID = /r/foreignRealm
    x.Modify()                      <-- write fail, /r/myRealm != /r/foreignRealm
```

Field writes and pointer-receiver method mutations are also blocked:
/r/foreignRealm cannot modify `x` because the receiver-pointer's .Base resolves
to /r/myRealm's block.

```go
    // PKGPATH: gno.land/r/myRealm
    var x foreignRealm.MyStruct     <-- ObjectInfo.PkgID = /r/foreignRealm
    x = foreignRealm.GlobalMyStruct <-- OK (whole-slot replace)
    x.Field = "..."                 <-- write fail, /r/myRealm != /r/foreignRealm
    x = *foreignRealm.NewMyStruct() <-- OK (whole-slot replace)
    x.Field = "..."                 <-- write fail, /r/myRealm != /r/foreignRealm
    x.Modify()                      <-- write fail, /r/myRealm != /r/foreignRealm
```

The slot itself, however, is /r/myRealm's. /r/myRealm can replace the slot
wholesale by assigning the return value of a /r/foreignRealm constructor (e.g.
`x = *foreignRealm.NewT()`).

```go
    // PKGPATH: gno.land/r/myRealm
    var x foreignRealm.MyStruct =
        foreignRealm.MyStruct{Field:"..."} <-- fail at checkConstructionTime
```

It cannot choose the contents -- composite literals of foreign types are blocked
at checkConstructionTime.

```go
    // PKGPATH: gno.land/r/myRealm
    var x *foreignRealm.MyStruct =
        foreignRealm.NewMyStruct()  <-- ObjectInfo.PkgID = /r/foreignRealm
    x.Modify()                      <-- OK
```

Pointer-typed slots (`var x *foreignRealm.T`) follow the standard cross-realm
pointer model: the pointee lives in /r/foreignRealm's heap, so /r/foreignRealm's
own non-crossing helpers can mutate it through x. This is the canonical idiom,
not a caveat.

## `cross` -> `cross(cur)` Migration

In order to facilitate the migration of `cross` to `cross(cur)` the `cross1`
keyword was temporarily provided. **`cross1` has since been removed** — the
name no longer resolves (see `gnovm/tests/files/zrealm_cross1_removed.gno`).
The steps below are retained as a historical record of the migration process.

 * Step 1: rename `cross` to `cross1`. 
 * Step 2: rename `runtime.CurrentRealm()` to `unsafe.CurrentRealm()` etc.
 * Step 3: tests should still pass, all functionality intact.
 * Step 4: replace `unsafe.CurrentRealm().Address()` to `cur.Address()` etc.
   when `cur` is available.
 * Step 5: refactor codebase to thread `cur` as (usually) secondary parameters
   to allow step 6 and step 7.
 * Step 6: replace `cross1` with `cross(cur)`
 * Step 7: replace remaining `unsafe.CurrentRealm().Address()` to
   `cur.Address()` etc.

Step 5 is probably the most difficult part of the migration.

Keep in mind the following:

 * You can pass in `realm` as a secondary parameter to make it non-crossing.

```go
// Send implements Banker.
//
// rlm must be the caller's own captured cur (i.e. the cur of the
// immediate crossing-function caller). Sending with rlm = cur.Previous()
// or any other realm value is rejected: rlm.IsCurrent() asserts pointer
// identity against the topmost crossing frame. Combined with the
// rlm.Address() == gb.owner check, this restricts Send to the owning
// realm acting in its own frame.
//
// `_ int` is the dummy first parameter that forces rlm realm into second
// position, so the method is non-crossing.
func (gb *GRC20Banker) Send(_ int, rlm realm, p Payment) error {
	if !rlm.IsCurrent() {
		return ErrSpoofedRealm
	}
	if rlm.Address() != gb.owner {
		return ErrCurrentRealmIsNotOwner
	}
    ...
}
```

In `GRC20Banker` above the first parameter is `_ int` and the second parameter
`rlm realm`.  The first parameter is ignored, it is only there to make `rlm
realm` the second parameter (as if it were the first parameter it would make
`*GRC20Banker.Send(...)` a crossing method. `_ int, rlm realm` is a common
pattern in /examples now after the v2 refactor.

__NOTE: Notice `rlm.IsCurrent()` is often required as the first assertion check,
otherwise the caller can pass in something like `cur.Previous()` and the
behavior will be different. Crossing functions do not require `cur.IsCurrent()`
because the runtime ensures that it is always true.__

 * You can now pass `realm` to a closure as a name capture (e.g. `func() {
 doSomething(cur) }()`. This was prevented in v1 because closures can also be
 persisted, but the gnovm finalization logic fails when `realm` persistence is
 attempted so it wasn't strictly necessary. Since the `realm` type is more
 ubiquitous this rule has been relaxed, but `realm` still cannot be persisted so
 be careful.

NOTE: this is fine for transient closures used in the same transaction, but if
the closure is assigned to a persisted slot, finalize will fail when it tries to
serialize the captured realm -- `refusePersistRealmHIV` panics with
`errPersistRealm`.

 * Many common functions can optionally take `cur realm` as the first parameter,
 such as `init(cur realm)`, `main(cur realm)` (for filetests and MsgRun),
 `Render()`, `Test...(t *testing.T)`, and even `t.Run("...", func(cur realm, t
 *testing.T) {`.

 * `testing.SetRealm(...)` semantics has changed. In v1, `testing.SetRealm(...)`
 would override the realm context of a frame, and this worked well with
 `runtime.CurrentRealm()` etc because `runtime.CurrentRealm()` was a dynamic
 stack-walking function. On the other hand `cur realm` or `rlm realm` are
 real runtime values and `cur.Address()` or `cur.Previous()` do not inspect the
 stack dynamically.

Now testing.SetRealm(...) sets two things in parallel to keep both observation
paths consistent. (1) It writes a RealmOverride{Addr, PkgPath} into
ctx.RealmFrames[frameIdx] so the legacy stack-walker runtime.CurrentRealm() /
unsafe.CurrentRealm() reports the spoofed identity. (2) It also reaches into the
frame's captured cur value and mutates the underlying .grealm's addr, pkgPath,
and prev fields in place, so subsequent reads of cur.Address(), cur.PkgPath(),
and cur.Previous() reflect the override. Behavior of cur.Previous() depends on
the override shape: a UserRealm override (pkgPath == "") rewrites prev to a
true-nil so cur.Previous() panics like at the EOA-root boundary; a CodeRealm
override (pkgPath != "") rewrites prev to a fresh .grealm whose pkgpath is the
frame's own function's PkgPath -- i.e. the code surrounding the SetRealm call --
so cur.Previous() returns the realm that "called into" the spoofed identity,
matching what the legacy stack walker would surface.

What testing.SetRealm(...) does not touch: m.Realm and m.Alloc.currentRealmID --
the two values that drive borrow-rule decisions and the readonly/PkgID write
gate. A SetRealm-spoofed frame still has the original storage realm; it cannot
fool DidUpdate, the readonly check, or checkConstructionTime. The override is
observation-only — it changes what user code sees via the realm-identity APIs,
not what the VM enforces for cross-realm authority. So a /r/myRealm test that
calls testing.SetRealm(testing.NewCodeRealm("gno.land/r/foreign")) and then
tries to write a foreign-stamped value still panics on the readonly check,
because the underlying authority machinery is independent of the test-frame
override.

Also see [./migration_guide.md](./migration_guide.md) which was auto-generated by Claude
during some of the migration. When using Claude or AI for migration purposes
give both documents as reference along with [gno-interrealm-v2.md](../../docs/resources/gno-interrealm-v2.md).

It also helps to have Claude use the popular 'andrej-karpathy-skills' for more
surgical changes.

## TODO

 * TODO: Note more caveats, special cases, and surprises.
 * TODO: Guide on migrating for new borrow rules.

