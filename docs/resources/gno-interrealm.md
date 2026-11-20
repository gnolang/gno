# Interrealm: the rules

> The overview, with diagrams and the catalogue of interactions, is [the whitepaper](./gno-interrealm-whitepaper.md).
>
> The decision record is
> [`gnovm/adr/pr6495_cross_to_write.md`](../../gnovm/adr/pr6495_cross_to_write.md).
> This page teaches the rules; the mechanism follows them.

A realm's stored state changes only during a crossing call into it. The rules
below say who may write when code changes hands. Each section is a
snippet, one sentence, and the exact error a test pins.

## R1 Cross to write

```go
// /r/bank/bank
var total int

func Deposit(cur realm, n int) { total += n } // writer: crossing
func Total() int               { return total }

// /r/app/app, another realm
bank.Deposit(cross(cur), 5) // writes bank's state
bank.Total()                // reads it
```

A realm's stored state changes only during a crossing call into it; code
entered any other way, and library code touching another realm's objects,
runs read-only.

```
// Error:
// gno.land/r/tests/vm/ctw's code was entered without crossing and is read-only (call its crossing function): Count<~VPBlock(3,0)>
```

## R2 Only pointers cross

```go
// /r/users/users
type User struct{ Name string }

func Get() User { return User{Name: "bob"} }

// /r/app/app
u := users.Get() // refused: a copy of another realm's struct
```

A struct or array value of a realm's declared type is created and copied only
by that realm's code; hold a pointer instead. Exceptions: the receiver copy of
a direct call to a value-receiver method, and an unnamed struct whose fields
are all exported and which holds no realm's struct or array: it may be passed
and returned by value, and the copy belongs to whoever made it. Hand out
errors as pointers, and write `defer func() { x.M() }()` where a direct
`defer x.M()` would copy another realm's value. The type checker refuses a
declaration or copy first (`cannot hold a value of <T> outside its realm (use
a pointer)`); a run-time allocation fails with `cannot allocate <T> in realm
<R>`.

```
// Error:
// cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer)
```

## S3 A function value writes only in the crossing call that made it

```go
// /r/app/app
var total int

func Run(cur realm) {
	f := func() { total++ }
	f()                       // same crossing call: writes
	other.Keep(cross(cur), f)
	other.RunKept(cross(cur)) // later crossing call: read-only
}
```

A non-crossing function value (a closure, a bound method value, a declared
function used as a value) writes only during the crossing call that made it;
made in read-only code, at init, or loaded from the store, its id is 0 and the
call runs read-only. There is then no crossing function to call: make the
callback crossing (`func(cur realm, …)`, called with `cross(cur)`) or rebuild
the value inside the call that runs it.

```
// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only in the crossing call that made it: (const (ref(gno.land/r/tests/vm/cbsaved) package{})).Owner
```

## R4 Makers stay home

```go
// /r/app/app
var sum int

func Run(cur realm) {
	other.ForEach(cross(cur), func(n int) { sum += n }) // handed to another realm: no maker
}
```

When another realm's code hands a function value over (argument, receiver,
result) or copies it, what arrives has no maker, so publishing a function is
never a grant of write power; a function value handed to another realm runs
read-only there. Not covered: a caller hands a realm's crossing function a
reference to that realm's own stored functions; your crossing function must not
call function values found in caller-supplied containers. `gno lint` warns when
a crossing function calls a function reached through a parameter that is not a
function; it does not yet follow values through two or more locals, another
function of the realm, or a closure capture.

```
// Error:
// gno.land/r/ctw_callback's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/ctw_callback) package{})).sum
```

## R6 One import rule

```go
// /r/app/app
var buf = make([]byte, 4)

func Run(cur realm, r io.Reader) {
	r.Read(buf) // view-only unless r's package is in app's import tree
}
```

Code reached through an interface or a method value is trusted only if its
package is in the calling realm's import tree (its imports and theirs, the
standard library included, test files excluded); otherwise it runs view-only.

```
// Error:
// cannot copy to readonly tainted slice
```

## Upgrade notes

Existing code must change in three ways; one more thing shifts:

- A struct error returned to another realm becomes a pointer error; `r/sys/users`
  is the worked example.
- A library container that stores a realm's struct receives a pointer instead
  (`commondao`).
- A constructor whose closures the realm's crossing functions later run must
  itself be crossing (`gov/dao NewGovDAO`).
- State saved before this change keeps its old function-value references: the
  rules hold chain-wide only from a fresh genesis or a migration, and gas and
  storage deposits shift where declared functions are passed as values.

## Introduction
### Design Goals

**Caveat: the rules do not secure applications against arbitrary code
execution. A function or interface value a caller supplies runs read-only in
its realm or view-only when reached through a value or a foreign interface, so
it cannot write your stored state; it still reads what it is handed and can
panic to end the call. Check caller-supplied callbacks;
whitelisting by a DAO is one way.**

P package code behaves the same when copied verbatim into a realm, with one
difference: the copy becomes realm code, so a caller outside the realm runs it
read-only.

Realm crossing with respect to `unsafe.CurrentRealm()` and
`unsafe.PreviousRealm()` must be explicit and warrants type-checking; because
a crossing-function of a realm should be able to call another crossing-function
of the same realm without necessarily crossing (changing the realm-context).
Sometimes the previous realm and current realm must be the same realm, such as
when a realm consumes a service that it offers to external realms and users.

For Go developers familiar with blockchain VMs like the EVM: in Solidity,
calling another contract implicitly shifts `msg.sender`, making it easy to
introduce reentrancy bugs or misattribute the caller. Gno's `cross(rlm)` form
makes every realm-context transition visible in source code and verifiable by
the compiler, eliminating this class of bugs by construction.

Where a real object resides should not matter too much, as it is often
difficult to predict. Thus the realm-context as returned by
`unsafe.PreviousRealm()` and `unsafe.CurrentRealm()` should not change with
non-crossing method calls, and the realm-storage-context should be determined
for non-crossing methods only by the realm-storage of the receiver. The
realm-storage of a receiver should only matter for when elements reside in
external realm-storage and direct dot-selector or index-expression access of
sub-elements are desired of the aforementioned element.

During a crossing call into a realm, a method should be able to modify the
receiver and associated objects of the same realm-storage as that of the
receiver.

A method should be able to create new objects that reside in the same realm by
association in order to maintain storage realm consistency and encapsulation
and reduce fragmentation.

It is difficult to migrate an object from one realm to another even when its
ref-count is 1; such an object may be deep with many descendants of ref-count 1
and so performance is unpredictable.

Code declared in p packages (or declared in "immutable" realm packages) can
help different realms enforce contracts trustlessly, even those that involve
the caller's current realm. Otherwise two mutable (upgradeable) realms cannot
export trust unto the chain because functions declared in those two realms can
be upgraded. This is analogous to Go interfaces but stronger: a Go interface
guarantees method signatures, while a Gno p-package guarantees *behavior* —
the implementation is immutable on-chain, so both parties can trust the logic
without trusting each other.

Both `fn(cross, ...)` and `func fn(cur realm, ...){...}` may become special
syntax in future Gno versions.

```go
// realm /r/alice/alice
package alice

var object any

func SetObject(cur realm, obj any) {
    object = obj
}
```

```go
// package /p/bob/types
package types

type UserProfile struct {
    Name string
    ...
}
```

```go
// realm /r/bob/bob
package bob

import "gno.land/r/alice/alice" // import external realm package
import "gno.land/p/bob/types"   // import external library package

func Register(cur realm, name string) {
    prof := types.UserProfile{Name: name}
    alice.SetObject(cross(cur), prof)
}
```

The Gno language is extended to support a `context.Context`-like argument to
denote the current realm-context of a Gno function. This allows a user realm
function to call itself safely as if it were being called by an external user,
and helps avoid a class of security issues that would otherwise exist.

```go
// realm /r/alice/mail

func SendMail(cur realm, text string) {
    if text == "" {
        // cur.Previous() is preserved for recursive call.
        SendMail(nil, "<empty>")
    }
    caller := cur.Previous()
    if inBlacklist(caller) {
        // cur.Previous() becomes self; message from self to self.
        SendMail(cross(cur), fmt.Sprintf("blacklisted caller %v blocked", caller))
    } else {
        // sendMailPrivate not exposed to external callers.
        sendMailPrivate(text)
    }
}
```

## Object Model and Realm Storage

Unlike other blockchain platforms where developers must manually serialize state
into key-value stores, Gno persists ordinary Go structs and slices
automatically. A Go developer can write familiar data structures — trees,
linked lists, maps of structs — and they survive across transactions with no
explicit marshaling. The runtime handles Merkle-ization and garbage collection
transparently.

Every object in Gno is persisted on disk with additional metadata including the
object ID and an optional OwnerID (if persisted with a ref-count of exactly 1).
The object ID is only set at the end of a realm-transaction during
realm-transaction finalization (more on that later). A GnoVM transaction is
composed of one or many scoped (stacked) realm-transactions.

```go
type ObjectInfo struct {
	ID       ObjectID  // set if real.
	Hash     ValueHash `json:",omitempty"` // zero if dirty.
	OwnerID  ObjectID  `json:",omitempty"` // parent in the ownership tree.
	ModTime  uint64    // time last updated.
	RefCount int       // for persistence. deleted/gc'd if 0.

	// Object has multiple references (refcount > 1) and is persisted separately
	IsEscaped bool `json:",omitempty"` // hash in iavl.
    ...
}
```

When an object is persisted during realm-transaction finalization the object
becomes "real" (as in it is really persisted in the virtual machine state) and
is said to "reside" in the realm; and otherwise is considered "unreal". New
objects instantiated during a transaction are always unreal; and during
finalization such objects are either discarded (transaction-level garbage
collected) or become persisted and real.

Unreal (new) objects that become referenced by a real (persisted) object at
runtime will get their OwnerID set to the parent object's storage realm, but
will not yet have its object ID set before realm-transaction finalization.
Subsequent references at runtime of such an unreal object by real objects
residing in other realms do not override the OwnerID initially set, so during
realm-transaction finalization it ends up residing in the first realm it became
associated with (referenced from). Unreal objects that become persisted but were
never directly referenced by any real object during runtime will only get their
OwnerID set to the realm of the first real ancestor.

Real objects with ref-count of 1 have their hash included in the sole parent
object's serialized byte form, thus an object tree of only ref-count 1
descendants are Merkle-hashed completely.

When a real or unreal object ends up with a ref-count of 2 or greater during
realm-transaction finalization its OwnerID is set to zero and the object is
considered to have "escaped". When such a real object is persisted with
ref-count of 2 or greater it is forever considered escaped even if its
ref-count in later transactions is reduced to 1. Escaped real objects do not
have their hash included in the parent objects' serialized byte form but
instead are Merkle-ized separately in an iavl tree of escaped object hashes
(keyed by the escaped object's ID) for each realm package. (This is implemented
as a stub but not yet implemented for the initial release of Gno.land.)

**A real object can only be directly mutated through dot-selectors and
index-expressions if the object resides in the same realm as the current
realm-storage-context. Unreal objects can always be directly mutated if their
elements are directly exposed.**

Exposed values accessed through dot-selectors and index-expressions from
external realm logic are tainted read-only. For the full rules see
[Readonly Taint Specification](#readonly-taint-specification).

### Storage = Authority (allocation-time PkgID)

Every object's `ObjectInfo.ID.PkgID` is stamped at *allocation time* to the
realm-storage-context that allocated it. This is the unifying invariant:
the realm that holds storage authority over an object is the same realm
that allocated it — there's no separate "ownership" or "linked-to" concept
divergent from PkgID.

```go
type ObjectID struct {
    PkgID   PkgID  // stamped at allocation (the authoring realm)
    NewTime uint64 // stamped at finalization (zero until persisted)
}
```

Three states an `ObjectID` can be in:

| State        | `PkgID`  | `NewTime`  | Meaning                                  |
|--------------|----------|------------|------------------------------------------|
| empty        | zero     | zero       | Never went through the allocator         |
| allocated    | set      | zero       | In memory; authority known, not persisted|
| finalized    | set      | non-zero   | Real; persisted with a tx-stamped NewTime|

Two consequences follow:

1. **Ownership is known before persistence.** An unreal receiver — for
   example, a value just returned from a foreign realm's constructor —
   carries its authoring realm's PkgID immediately, so a library method
   called on it treats it as a view exactly like a stored object.

2. **Construction-time enforcement.** Composite literals (`/r/foo.T{...}`),
   `new(/r/foo.T)`, and `make([]/r/foo.T, ...)` of a foreign `/r/`-declared
   type panic when invoked outside the declaring realm. Authority cannot
   be forged by constructing values of a realm's types from elsewhere —
   construction must go through a realm-declared constructor function,
   which allocates under its declaring realm even when entered without
   crossing.

   Conversions to a foreign `/r/`-declared type are refused too, with
   `illegal conversion to external realm type`, directly or through one
   pointer level (`(*/r/foo.T)(&local)`), unless its underlying type is a
   primitive, func, or chan type; a nil pointer is let through.

   Holding such a value is refused too: `var t foo.T` in another realm fails
   with `cannot hold a value of foo.T outside its realm (use a pointer)`.
   Hold the `*foo.T` that foo's constructor returns, and never treat a
   non-nil `*foo.T` as authorization; identify the caller through
   `cur.Previous()`.

Storage rent attribution follows PkgID too: when a transaction mutates an
object owned by `/r/A` during a crossing call into `/r/A`, the byte delta
accrues to `/r/A`'s ledger, not the caller's, because `/r/A` is the authoring
(and storage-paying) realm.

## Crossing-Functions and Crossing-Methods

Realm crossing occurs when a crossing function (declared as
`func fn(cur realm, ...){...}`)
is called with the Gno `fn(cross, ...)` syntax.

```go
package main
import "gno.land/r/alice/realm1"

func main() {
    bread := realm1.MakeBread(cross, "flour", "water")
}
```

(In Linux/Unix operating systems user processes can cross-call into the kernel
by calling special syscall functions, but user processes cannot directly
cross-call into other users' processes. This makes the GnoVM a more complete
multi-user operating system than traditional operating systems.)

Only an explicit cross-call lets a realm's stored state change. A non-crossing
call into another realm's code runs read-only, and a library method on a
realm's object treats it as a view; neither shifts the realm-context, so
`unsafe.CurrentRealm()` does not change unless a method is called like
`receiver.Method(cross, args...)`.

Realms hold objects in residence and they also have a Gno address to send and
receive coins from. Coins can only be spent from the current realm context.

### Definition

A crossing-function or crossing-method is that which is declared in a realm and
has as its first argument `cur realm`. The `cur realm` argument must appear as
the first argument of a crossing-function or crossing-method's argument
parameter list. To prevent confusion it is illegal to use anywhere else, and
cannot be used in p packages.

The current realm-context and realm-storage-context changes when a
crossing-function or crossing-method is called with `cross(rlm)` in the first
argument position as in `fn(cross(rlm), ...)`. The `rlm` must be a realm-typed
identifier in scope; the runtime verifies it is the current frame's `cur` before
the cross-call proceeds. Such a call is called a "cross-call" or "crossing-call".

```go
package main
import "gno.land/r/alice/extrealm"

func MyMakeBread(cur realm, ingredients ...any) { ... }

func main(cur realm) {
    MyMakeBread(cross(cur), "flour", "water")    // ok -- cross into self.
    extrealm.MakeBread(cross(cur), "flour", "water") // ok -- cross into extrealm
}
```

When a crossing-function or crossing-method is called with `nil` as the first
argument instead of `cross(rlm)` it is called a non-crossing-call; and no
realm-context nor realm-storage-context change takes place.

```go
package main
import "gno.land/r/alice/extrealm"

func MyMakeBread(cur realm, ingredients ...any) { ... }

func main(cur realm) {
    MyMakeBread(cur, "flour", "water") // ok -- non-crossing.
    extrealm.MakeBread(cur, "flour", "water") // invalid -- external realm function
}
```

To prevent confusion a non-crossing-call of a crossing-function or
crossing-method declared in a realm different than that of the caller's
realm-context and realm-storage-context will result in either a type-check
error; or a runtime error if the crossing-function or crossing-method is
variable.

### Realm-Context Rules

All functions in Gno execute under a realm context as determined by the call
stack. Objects that reside in a realm can only be modified if the realm context
matches.

A function declared in p packages (or in stdlib) when called:

 * runs with the caller's authority for a top-level function called by name,
   and on a receiver the calling realm owns, or a primitive.
 * runs as the caller for a method on a `/p/`-owned receiver; after
   initialization, that package's immutability check reports a write.
 * a method on a realm-owned receiver is a view: its writes to the owner's
   object are refused.
 * a closure, bound method value or declared function used as a value follows
   S3: it acts for the realm whose writable code made it, only in the crossing
   call that made it, and view-only when made in read-only code.
 * runs view-only when reached through an interface or method value whose
   package the current realm does not import, directly or through its imports,
   the standard library included.

A function declared in a realm package (`/r/X`) when called:

 * explicitly crosses to `/r/X` if the function is declared as
   `func fn(cur realm, ...){...}` (with `cur realm` as the first argument) AND
   called with `cross(rlm)` as the first argument. The new realm is called the
   "current realm".
 * otherwise (non-crossing call of a `/r/X`-declared callable from a different
   realm-storage-context) **runs read-only**: it reads anything, constructs
   objects owned by `/r/X`, writes objects allocated during the current
   read-only stretch, and writes no stored object. This applies uniformly:
   top-level functions, methods on any receiver shape, and closures whose
   construction site was `/r/X`. The realm-context does NOT change.

This prevents both directions of abuse. An attacker-supplied value's method,
called by the victim, cannot write the victim's state; and an exported
non-crossing function of the victim, called by the attacker, cannot write the
victim's state either. To write, call the realm's crossing function.

`unsafe.CurrentRealm()` returns the current realm-context that was last
cross-called to. `unsafe.PreviousRealm()` returns the realm-context
cross-called to before the last cross-call. All cross-calls are explicit via
`cross(rlm)` at Args[0], as are non-crossing-calls of crossing-functions and
crossing-methods (which use `nil` instead).

A crossing function declared in the same realm package as the callee may be
called like `fn(cross(cur), ...)` or `fn(cur, ...)`. When called with
`fn(cur, ...)` there is no realm crossing, but when called like
`fn(cross(cur), ...)` there is technically a realm crossing and the current
realm and previous realm returned are the same.

The current realm and previous realm do not depend on any non-crossing call.
In other words `unsafe.CurrentRealm()` may differ from the Machine's active
storage realm (internally `m.Realm`) when a `/p/` method is called on a
`/p/`-owned receiver, and code entered read-only has no storage realm at all.

### Library-Owned Receivers

A `/p/` method on a receiver owned by a `/p/` package runs as the caller, and
that package's post-init immutability check reports any write to its own state.
The realm-context does not change.

A realm's object is a view to a `/p/` method (`avl.Tree.Set`, a GRC20
`Teller` method): it writes the object only while that realm is current, which
means during a crossing call into the realm; called from anywhere else the
write is refused. Realm code entered without crossing runs read-only (R1), so
an `/r/`-declared method called on a foreign object never writes it either.

### Crossing-Method Semantics

When a method is a crossing-method called as
`receiver.Method(cross, args...)`, both the realm-context and
realm-storage-context change to that of the realm package in which the
type/method is declared (which is not necessarily the same as where the
receiver is stored). The call writes the receiver only if the receiver belongs
to that realm, which is where it was allocated (see Storage = Authority); a
receiver owned by another realm is a view. As mentioned previously a
non-crossing-call of a crossing-method will fail during type-checking or at
runtime if the receiver belongs to an external realm-storage.

Calls of methods on receivers residing in realms different from the current
realm must *not* be called like `fn(cross, ...)` if the method is not a
crossing function itself, and vice versa. Implicit storage changes are not a
realm crossing.

A crossing method declared in a realm cannot modify the receiver if the object
belongs to a different realm. However not all methods are required to be
crossing methods, and crossing methods may still read the state of the receiver
(and in general anything reachable is readable).

```go
// Type and methods declared in /r/alice/tokens
package tokens

type Token struct {
    Owner   string
    Balance int
}

func New() *Token { return &Token{Balance: 100} }

// Non-crossing method on /r/-declared type: writes only inside a crossing
// call into /r/alice/tokens, on a token alice allocated. Called from another
// realm it runs read-only.
func (t *Token) Debit(amount int) {
    t.Balance -= amount
}

// Crossing method: realm-context and realm-storage-context shift to
// /r/alice/tokens. The token is alice's object, so the write lands.
func (t *Token) Transfer(cur realm, to string, amount int) {
    t.Balance -= amount
}
```

```go
// /r/bob/bob stores a token alice made
package bob

import "gno.land/r/alice/tokens"

var myToken *tokens.Token // alice's object, stored in /r/bob/bob

func Spend(cur realm) {
    myToken = tokens.New() // alice's constructor allocates it

    // Crossing call: storage-context shifts to /r/alice/tokens and
    // Transfer writes myToken, which belongs to alice.
    myToken.Transfer(cross(cur), "g1...", 10) // writes the token
}

func DebitFails() {
    // Non-crossing call on alice's method: alice's code runs read-only.
    myToken.Debit(10) // fails: alice's code entered without crossing
}
```

A method declared in an `/r/` package never writes another realm's object, and
writes its own realm's objects only during a crossing call into that realm.

If you need a `/r/A`-declared method to mutate state owned by `/r/B`, the
caller in `/r/B` must own the mutation: pass the data into the method as an
argument and let the method return the new value, or expose a method on
`/r/B`'s own type that delegates internally. The pattern of "method on
my-type, mutate via a receiver owned by another realm" is intentionally not
supported, because it would allow foreign-realm code to run with the
storage realm's authority.

An object belongs to the realm whose code allocated it, wherever it is later
stored; new unreal objects reachable from the current realm become persisted
in it upon finalization. In the future the `attach()` function will prevent a
new unreal object from being taken.

For how crossing rules apply to MsgCall, MsgRun, and package initialization,
see [Message Types and Testing](#message-types-and-testing).

## Realm Boundaries

A realm boundary is a crossing call, `fn(cross, ...)`: the realm-context
changes from one realm to another. A realm may cross into itself. Entering
another realm's code read-only changes the running realm but writes nothing
stored; a method on a `/p/`-owned receiver and picked library code are not a
realm crossing.

When a crossing-function or crossing-method is cross-called it shifts the
"current" runtime realm-context to the "previous" runtime realm-context such
that `unsafe.PreviousRealm()` returns what used to be returned with
`unsafe.CurrentRealm()` before the realm boundary. The current
realm-storage-context is always set to that of realm-context after
cross-calling.

Every crossing-call creates a realm boundary even when there is no resulting
change in realm-context or realm-storage-context. A non-crossing-call of a
crossing-function or crossing-method (`fn(cur, ...)`) never creates a realm
boundary.

## Captured Realm Values (`cur realm`)

A crossing-function's first parameter `cur realm` is a captured realm value:
a typed handle on the realm-context at the moment of the crossing call.
`realm` is a uverse interface with the following methods:

  - `Address() address` — bech32 address derived from the realm's pkgpath
    (or the EOA address at the chain root).
  - `PkgPath() string` — the realm's pkgpath, or `""` at the chain root.
  - `Previous() realm` — the captured realm that was current before this
    one. At the chain root this returns a non-nil origin realm whose
    `PkgPath() == ""`; calling `Previous()` past the origin panics with
    the same "frame not found" message `unsafe.PreviousRealm()` uses
    for the same walk-end.
  - `IsCode() bool`, `IsUser() bool`, `IsUserCall() bool`,
    `IsUserRun() bool`, `IsEphemeral() bool` — classification methods
    that mirror their counterparts on `chain/runtime.Realm`. Derived
    purely from `Address()` + `PkgPath()`, so they return the same
    values as the equivalent runtime calls at the corresponding
    position.
  - `String() string` — debug-friendly representation.

Parity with `unsafe.{Current,Previous}Realm()` at every comparable
position:

  - `cur.Address()` and `cur.PkgPath()` agree with `unsafe.CurrentRealm()`.
  - `cur.Previous().Address()` and `cur.Previous().PkgPath()` agree with
    `unsafe.PreviousRealm()`.

The two APIs differ only in shape: `unsafe.CurrentRealm()` and
`unsafe.PreviousRealm()` return a `runtime.Realm` **struct** (defined in
`chain/runtime`), while `cur realm` is the uverse **interface**. They are
**distinct types** — not assignable to each other — that happen to surface
the same addr+pkgpath pair. The struct form is the legacy ergonomic API;
the interface form is the chain-aware handle that crossing-functions
receive directly.

### Realm values are never persisted

Captured realm values are ephemeral and tied to a call frame's chain. They
must not survive past the transaction:

  - Storing a `realm`-typed value (whether `cur` itself or any `Previous()`
    result) into a top-level realm var, a struct field, a map value, a
    slice/array element, or a closure capture causes the realm to refuse
    the operation at the point of attachment or at transaction finalize
    with: `cannot persist realm value: realm values are ephemeral and
    tied to a call frame`.
  - A `realm`-typed *parameter* or *return type* in a function signature is
    a static type reference and is allowed — this rule applies to the
    *value*, not the *type*. A function `func F(r realm) realm { ... }`
    is persistable; assigning its result into a persistent slot is not.

Code that needs to remember a caller realm across transactions should
store its `Address()` or `PkgPath()` (plain strings) — not the realm value
itself.

## Realm-Transaction Finalization

Realm-transaction finalization occurs when returning from a realm
boundary. When returning from a cross-call (via `cross(rlm)`)
realm-transaction finalization will occur even with no change of
realm-context or
realm-storage-context. Realm-transaction finalization does NOT occur when
returning from a non-crossing-call of a method of an unreal receiver or a real
receiver that resides in the same realm-storage-context as that of the caller.

During realm-transaction finalization all new reachable objects are assigned
object IDs and stored in the current realm-storage-context; and ref-count-zero
objects deleted (full "disk-persistent cycle GC" will come after launch); and
any modified ref-counts and new Merkle hash root computed.

## Readonly Taint Specification

Go's language rules for value access through dot-selectors & index-expressions
are the same within the same realm, but exposed values through dot-selector &
index-expressions are tainted read-only when performed by an external realm.

The readonly taint prevents the direct modification of real objects by any
logic, even from logic declared in the same realm as that of the object's
storage-realm.

A realm cannot directly modify another realm's objects without calling a
function that gives permission for the modification to occur.

For example `externalrealm.Foo` is a dot-selector expression on an external
object (package) so the value is tainted with the `N_Readonly` attribute.

The same is true for `externalobject.FieldA` where `externalobject` resides in
an external realm.

The same is true for `externalobject[0]`: direct index expressions also taint
the resulting value with the `N_Readonly` attribute.

The same is true for `externalobject.FieldA.FieldB[0]`: the readonly taint
persists for any subsequent direct access, so even if FieldA or FieldB resided in
the caller's own realm-context or realm-storage the result is tainted readonly.

A Gno package's global variables even when exposed (e.g. `package realm1; var
MyGlobal int = 1`) are safe from external manipulation (e.g. `import
"xxx/realm1"; realm1.MyGlobal = 2`) by the readonly taint when accessed
directly by dot-selector or index-expression from external realm logic; and
also by a separate `DidUpdate()` guard when accessed by other means such as by
return value of a function and the return value is real and external.

A function or method's arguments and return values retain and pass through any
readonly taint from caller to callee. Even if a realm's function (or method)
returns an untainted real object, the runtime guard in `DidUpdate()` prevents
it from being modified by an external realm-storage-context.

For a realm (user) to manipulate an untainted object residing in an external
realm, a crossing function (or crossing method) can be declared in the external
realm which references and modifies the aforementioned untainted object
directly (by a name declared outside of the scope of said function or method).
Or, the function can take in as argument an untainted real object returned by
another function.

Besides protecting against writing by direct access, the readonly taint also
helps prevent a class of security issue where a realm may be tricked into
modifying something that it otherwise would not want to modify. Since the
readonly taint prohibits mutations even from logic declared in the same realm,
it protects realms against mutating its own object that it doesn't intend to:
such as when a realm's real object is passed as an argument to a mutator
function where the object happens to match the type of the argument.

Objects returned from functions or methods carry no extra taint, but a
`/p/` method called on a foreign-owned object runs with the caller's authority,
so a write to the owner's object is refused; see "Views" below.

Whether or not an object is readonly tainted it can be mutated by the realm
that owns it during a crossing call into that realm. Called from elsewhere
without crossing, the owner's methods run read-only and `/p/` methods treat
the object as a view; the owner exposes a crossing function instead.

```go
// /r/alice

var blacklist []string

func GetBlacklist() []string {
    return blacklist
}

func FilterList(cur realm, testlist []string) { // blanks out blacklist items from testlist
    for i, item := range testlist {
        if contains(blacklist, item) {
            testlist[i] = ""
        }
    }
}
```

This is a toy example, but you can see that the intent of `FilterList()` is to
modify an externally provided slice; yet if you call `alice.FilterList(cross,
alice.GetBlacklist())` you can trick alice into modifying its own blacklist--the
result is that alice.BlackList becomes full of blank values.

With the readonly taint `var Blacklist []string` solves the problem for you;
that is, /r/bob cannot successfully call `alice.FilterList(cross,
alice.Blacklist)` because `alice.Blacklist` is readonly tainted for bob.

The problem remains if alice implements `func GetBlacklist() []string { return
Blacklist }` since then /r/bob can call `alice.FilterList(cross,
alice.GetBlacklist())` and the argument is not readonly tainted.

Future versions of Gno may also expose a new modifier keyword `readonly` to
allow for return values of functions to be tainted as readonly. Then with `func
GetBlacklist() readonly []string` the return value would be readonly tainted
for both bob and alice.

### Views

A reference that crosses a realm boundary, as a return value or as an
argument, is a **view**. Direct writes through it are refused by the readonly
taint, and a `/p/` method called on it runs with the *caller's*
storage-context, so its writes to the owner's object are refused as well.
Only a receiver owned by a `/p/` package runs as the caller for writes to its
own package's state; a realm's stored object is always a view. There is no
exception that changes this.

A realm that wants another realm to write exposes crossing functions that name
its own variables:

```go
// realm X
var users = avl.NewTree()

func Users() *avl.Tree                { return users } // view: foreign writes refused
func ReadUsers() *rotree.ReadOnlyTree { return rotree.Wrap(users, nil) }
func SetUser(cur realm, k string, v any) { users.Set(k, v) } // X writes, as X

// realm A
X.Users().Set("k", v)              // refused: X's tree, A's authority
X.SetUser(cross(cur), "k", v)      // commits to X's storage
X.ReadUsers().Set("k", v)          // does not compile: ReadOnlyTree has no Set
```

An argument keeps the same rule: writing through a caller-supplied reference
writes whichever of your own objects the caller named, so name your own
variable instead. A closure declared in an `/r/` realm runs read-only when
another realm calls it; to let a callback write its realm, make it a crossing
closure (`func(cur realm) {...}`) called with `cross`. A `/p/` or stdlib
closure called through a value by another realm's code runs view-only. A
method value of a `/p/` or stdlib method writes no stored object, receiver or
argument when reached from code whose realm does not import its package; when
that package is in the calling realm's import tree it acts for the caller
(R6). A `/p/`-owned receiver runs as the caller, and the package's post-init
immutability check reports a write to its own state.

Fresh-object corner: an object a realm function builds and returns carries
that realm's stamp from allocation, so the caller cannot write it either,
exactly like a stored object.

## `panic()` and `revive(fn)`

`panic()` behaves the same within the same realm boundary, but when a panic
crosses a realm boundary (as defined in [Realm Boundaries](#realm-boundaries))
the Machine aborts the program. This is because in a multi-user environment it
isn't safe to let the caller recover from realm panics that often leave the
state in an invalid state.

This would be sufficient, but we also want to write our tests to be able
to detect such aborts and make assertions. For this reason Gno provides
the `revive(fn)` builtin.

```go
abort := revive(func() {
    fn := func(_ realm) {
        panic("cross-realm panic")
    }
    fn(cross)
})
abort == "cross-realm panic"
```

`revive(fn)` will execute 'fn' and return the exception that crossed a realm
boundary during finalization.

This is only enabled in testing mode (for now), behavior is only partially
implemented. In the future `revive(fn)` will be available for non-testing code,
and the behavior will change such that `fn()` is run in transactional
(cache-wrapped) memory context and any mutations discarded if and only if there
was an abort.

TL;DR: `revive(fn)` is Gno's builtin for STM (software transactional memory).

## `attach()`

In future releases of Gno the `attach()` function can be used to associate
unreal objects to the current realm-storage-context before being passed into
a function declared in an external realm package, or into a method of a real
receiver residing in an external realm-context.

## `safely(cb func())`

In future releases of Gno the `safely(cb func())` function may be used to clear
the current and previous realm-context as well as any realm-storage-context
such that no matter what `cb func()` does the caller does not yield agency to
the callee.

For now this can be simulated by implementing an (immutable non-upgradeable)
realm crossing-function that cross-calls into itself once more before calling
the callback function.

## Guidelines

P package code cannot contain crossing functions. P package code also cannot
import R realm packages. But code can call named crossing functions e.g.
those passed in as parameters.

You must declare a public realm function to be a crossing function if it is
intended to be called by end users, because users cannot MsgCall non-crossing
functions (for safety/consistency) or p package functions (there's no point).

Utility functions that are a common sequence of non-crossing logic can be
offered in realm packages as non-crossing functions. These can also import and
use other realm utility non-crossing functions; whereas p packages cannot
import realm packages at all. And convenience/utility functions that are being
staged before publishing as permanent p code should also reside in upgradeable
realms.

Exported non-crossing functions and methods of a realm are read-only for other
callers: use them for getters, constructors, and helpers your own crossing
functions call. Anything another realm or a user calls to change the realm's
state must be crossing, and authenticates the caller with `cur.Previous()`.

You can always cross-call a method from a non-crossing method if you need it.

## Message Types and Testing

### MsgCall

MsgCall may only call crossing functions. This is to prevent potential
confusion for non-sophisticated users. Non-crossing calls of non-crossing
functions of other realms is still possible with MsgRun.

```go
// PKGPATH: gno.land/r/test/test

func Public(cur realm) {

    // cur.Previous() is (
    //     addr:<origin_caller>,
    //     pkgpath:""
    // ) == testing.NewUserRealm(origin_caller)
    cur.Previous()

    // cur is (
    //     addr:chain.PackageAddress("gno.land/r/test/test"),
    //     pkgpath:"gno.land/r/test/test"
    // ) == testing.NewCodeRealm("gno.land/r/test/test")
    cur.PkgPath()

    // Call a crossing function of same realm with crossing
    AnotherPublic(cross(cur))

    // Call a crossing function of same realm without crossing
    AnotherPublic(cur)
}

func AnotherPublic(_ realm) {
    ...
}
```

### MsgRun

```go
// PKGPATH: gno.land/e/g1user/run

import (
    "chain/runtime/unsafe"

    "gno.land/r/realmA"
)

func main() {
    // Before main() is called there is an implicit
    // crossing from UserRealm(g1user) to
    // CodeRealm(gno.land/e/g1user/run).

    // Returns (
    //     addr:g1user,
    //     pkgpath:""
    // ) == testing.NewUserRealm(g1user)
    unsafe.PreviousRealm()

    // Returns (
    //     addr:g1user,
    //     pkgpath:"gno.land/e/g1user/run"
    // ) == testing.NewCodeRealm("gno.land/e/g1user/run")
    unsafe.CurrentRealm()

    realmA.PublicNoncrossing()
    realmA.PublicCrossing(cross)
}
```

Realm addresses are derived from package paths via `chain.PackageAddress()`
(defined in `gnovm/stdlibs/chain/address.gno`):

- `chain.PackageAddress("gno.land/r/name123/realm")` — bech32 from hash(path)
- `chain.PackageAddress("gno.land/e/g1user/run")` — bech32 substring "g1user"

Therefore in the MsgRun file's `init()` function the previous realm and current
realm have different pkgpaths (the origin caller always has empty pkgpath) but
the address is the same.

### MsgAddPackage

A realm package's initialization (including `init()` calls) executes with
current realm-context of itself. `unsafe.PreviousRealm()` refers to the
package deployer both in global var decls and inside `init()` functions. After
that the package deployer is no longer provided, so packages need to remember
the deployer in the initialization phase if needed.

```go
// PKGPATH: gno.land/r/test/test

import "chain/runtime/unsafe"

func init() {
    // Returns (
    //     addr:<origin_deployer>,
    //     pkgpath:""
    // ) == testing.NewUserRealm(origin_deployer)
    // Inside init() and global var decls
    // are the only time unsafe.PreviousRealm()
    // returns the deployer of the package.
    // Save it here or lose it forever.
    unsafe.PreviousRealm()

    // Returns (
    //     addr:chain.PackageAddress("gno.land/r/test/test"),
    //     pkgpath:"gno.land/r/test/test"
    // ) == testing.NewCodeRealm("gno.land/r/test/test")
    unsafe.CurrentRealm()
}

// Same as in init().
var _ = unsafe.PreviousRealm()
```

```go
// PKGPATH: gno.land/e/g1user/run

import "chain/runtime/unsafe"

func init() {
    // Returns (
    //     addr:g1user,
    //     pkgpath:""
    // ) == testing.NewUserRealm(g1user)
    unsafe.PreviousRealm()

    // Returns (
    //     addr:g1user,
    //     pkgpath:"gno.land/e/g1user/run"
    // ) == testing.NewCodeRealm("gno.land/e/g1user/run")
    unsafe.CurrentRealm()
}
```

The same applies for pure package (`/p/`) initialization. During initialization
and tests, `unsafe.CurrentRealm()` can return a package path that starts with
"/p/". This is because the package is technically still mutable during its
initialization phase. After initialization, pure packages become immutable and
cannot maintain state.

### Testing overrides with stdlibs/testing

The `gnovm/tests/stdlibs/testing/context_testing.gno` file provides functions
for overriding frame details from Gno test code.

`testing.SetRealm(testing.NewUserRealm("g1user"))` is identical to
`testing.SetOriginCaller("g1user")`. Both will override the Gno frame to make it
appear as if the current frame is the end user signing with a hardware signer.
Both will also set `ExecContext.OriginCaller` to that user. One of these will
become deprecated.

#### Gno test cases with `_test.gno` like `TestFoo(t *testing.T)`

```go
// PKGPATH: gno.land/r/user/myrealm
package myrealm

import (
    "chain/runtime/unsafe"
    "testing"
)

func TestFoo(t *testing.T) {
    // At first OriginCaller is not set.

    // Override the OriginCaller.
    testing.SetRealm(testing.NewUserRealm("g1user"))

    // Identical behavior:
    testing.SetOriginCaller("g1user")

    // This panics now: seeking beyond the overridden origin frame:
    // unsafe.PreviousRealm()

    // Simulate g1user cross-calling Public().
    // Produce a new frame to override
    func() {
        testing.SetRealm(testing.NewCodeRealm("gno.land/r/user/myrealm"))

        unsafe.PreviousRealm() // "g1user", ""
        unsafe.CurrentRealm()  // bech32(hash("gno.land/r/user/myrealm")), "gno.land/r/user/myrealm"

        Public(...) // already in "gno.land/r/user/myrealm"
    }()

    // The following is identical to the above,
    // but not possible in p packages which
    // cannot import realms.
    Public(cross, ...)
}
```

#### Gno filetest cases with `_filetest.gno`

```go
// PKGPATH: gno.land/r/test/test
package test

import (
    "chain/runtime/unsafe"
    "testing"

    "gno.land/r/user/myrealm"
)

func init() {
    // XXX Frame not found, there is no deployer for filetests.
    unsafe.PreviousRealm()

    // Returns (
    //     addr:chain.PackageAddress("gno.land/r/test/test")
    //     pkgpath:"gno.land/r/test/test"
    // ) == testing.NewCodeRealm("gno.land/r/test/test")
    unsafe.CurrentRealm()
}

func main() {
    // There is assumed to be in "frame -1"
    // a crossing from UserRealm(g1user) to
    // CodeRealm("gno.land/r/test/test") before
    // main() is called, so crossing() here
    // is redundant.

    // Returns (
    //     addr:g1user,
    //     pkgpath:""
    // ) == testing.NewUserRealm(g1user)
    unsafe.PreviousRealm()

    // Returns (
    //     addr:g1user,
    //     pkgpath:"gno.land/r/test/test"
    // ) == testing.NewCodeRealm("gno.land/r/test/test")
    unsafe.CurrentRealm()

    // gno.land/r/test/test cross-calling
    // gno.land/r/user/myrealm:
    myrealm.Public(cross, ...)
}

// Output:
// XXX
```

## Implementation

Implementation for `unsafe.CurrentRealm()` and `unsafe.PreviousRealm()` are
defined in `gnovm/stdlibs/chain/runtime/unsafe/unsafe.gno` and related files in
the directory, while overrides for testing are defined in
`gnovm/tests/stdlibs/testing/context_testing.gno`. All stdlibs functions are
available unless overridden by the latter.

## Proposed Changes

`testing.SetOriginCaller()` may be deprecated in favor of
`testing.SetRealm(testing.NewOriginRealm(user))`.

`testing.NewCodeRealm(path)` may be renamed to
`testing.NewPackageRealm(path)`.
