# Interrealm Specification

## Introduction

All modern popular programming languages are designed for a single programmer
user.  Programming languages support the importing of program libraries natively
for components of the single user's program, but this does not hold true for
interacting with components of another user's (other) program. Gno is an
extension of the Go language for multi-user programming. Gno allows a massive
number of programmers to iteratively and interactively develop a single shared
program such as Gno.land.

The added dimension of the program domain means the language should be extended
to best express the complexities of programming in the inter-realm (inter-user)
domain. In other words, Go is a restricted subset of the Gno language in the
single-user context. (In this analogy client requests for Go web servers don't
count as they run outside the server program).

### Interrealm Programming Context

Gno.land supports three types of packages:
- **Realms (`/r/`)**: Stateful user applications (smart contracts) that
  maintain persistent state between transactions
- **Pure Packages (`/p/`)**: Stateless libraries that provide reusable
  functionality
- **Ephemeral Packages (`/e/`)**: Temporary code execution with MsgRun
  which allows a custom main() function to be run instead of a single
  function call as with MsgExec.

For an overview of the different package types in Gno (`/p/`, `/r/`, and
`/e/`), see [Anatomy of a Gno Package](../builders/anatomy-of-a-gno-package.md).

Interrealm programming refers to the ability of one realm to call functions
in another realm. This can occur between:
- Regular realms (`/r/`) calling other regular realms via MsgExec and MsgRun.
- Ephemeral realms (`/e/`) calling regular realms via MsgRun (like main.go)

The key concept is that code executing in one realm context can interact with
and call functions in other realms while leveraging the language syntax rules of
Go, enabling complex multi-user interactions while maintaining clear boundaries
and permissions.

### Realm-Context and Realm-Storage-Context

All logic in Gno executes under two contexts that together govern identity and
persistence:

**Realm-context** determines `runtime.CurrentRealm()` and
`runtime.PreviousRealm()`. It controls identity and agency: who is the current
actor and who called them. The realm-context has an associated Gno address from
which native coins can be sent and received. It changes only on explicit
cross-calls (`fn(cross, ...)`).

**Realm-storage-context** determines where new and modified objects are
persisted during realm-transaction finalization. It changes on explicit
cross-calls *and* on implicit borrow-crosses (calling a non-crossing method on
a real object residing in a different realm). The realm-storage-context is not
directly accessible at runtime and has no associated address.

After an explicit cross-call, both contexts refer to the same realm. They
diverge when calling a non-crossing method of a real object residing in a
different realm — the realm-storage-context shifts to the receiver's realm
while the realm-context stays the same.

| Call type | Realm-context changes? | Storage-context changes? | Boundary? | Finalizes? |
|---|---|---|---|---|
| `fn(cross, ...)` to same realm | Yes* | No | Yes | Yes |
| `fn(cross, ...)` to different realm | Yes | Yes | Yes | Yes |
| `fn(nil, ...)` (non-crossing-call) | No | No | No | No |
| Non-crossing method, real receiver in same realm | No | No | No | No |
| Non-crossing method, real receiver in different realm | No | Yes | Yes | Yes |
| Non-crossing method, unreal receiver | No | No | No | No |
| Non-crossing function | No | No | No | No |

\* `runtime.CurrentRealm()` returns the same realm, but `runtime.PreviousRealm()`
shifts — what was current becomes previous. See [Realm Boundaries](#realm-boundaries)
for definitions of boundary and finalization.

### Design Goals

**Caveat: The interrealm specification does not secure applications against
arbitrary code execution. It is important for realm logic (and even p package
logic) to ensure that arbitrary (variable) functions (and similarly arbitrary
interface methods) are not provided by malicious callers; such arbitrary
functions and methods whether crossing (or non-crossing) will inherit the
previous realm (or both current and previous realms) and could abuse these
realm-contexts.** It does not make sense for any realm user to cross-call an
arbitrary function or method as it loses agency while being marked as the
responsible caller by the callee's runtime previous realm. This problem is
worse when calling a non-crossing function or method. It can be reasonable when
such variable functions or interface values are restricted in other ways such
as by whitelisting by a DAO upon careful inspection of every such variable
function or interface value (both its type declaration as well as its state).

P package code should behave the same even when copied verbatim in a realm
package; and likewise non-crossing code should behave the same when copied
verbatim from one realm to another. Otherwise there will be lots of security
related bugs from user error.

Realm crossing with respect to `runtime.CurrentRealm()` and
`runtime.PreviousRealm()` must be explicit and warrants type-checking; because
a crossing-function of a realm should be able to call another crossing-function
of the same realm without necessarily crossing (changing the realm-context).
Sometimes the previous realm and current realm must be the same realm, such as
when a realm consumes a service that it offers to external realms and users.

For Go developers familiar with blockchain VMs like the EVM: in Solidity,
calling another contract implicitly shifts `msg.sender`, making it easy to
introduce reentrancy bugs or misattribute the caller. Gno's `cross` keyword
makes every realm-context transition visible in source code and verifiable by
the compiler, eliminating this class of bugs by construction.

Where a real object resides should not matter too much, as it is often
difficult to predict. Thus the realm-context as returned by
`runtime.PreviousRealm()` and `runtime.CurrentRealm()` should not change with
non-crossing method calls, and the realm-storage-context should be determined
for non-crossing methods only by the realm-storage of the receiver. The
realm-storage of a receiver should only matter for when elements reside in
external realm-storage and direct dot-selector or index-expression access of
sub-elements are desired of the aforementioned element.

A method should be able to modify the receiver and associated objects of the
same realm-storage as that of the receiver.

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
    alice.SetObject(cross, prof)
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
        // runtime.PreviousRealm() is preserved for recursive call.
        SendMail(nil, "<empty>")
    }
    caller := runtime.PreviousRealm()
    if inBlacklist(caller) {
        // runtime.PreviousRealm() becomes self; message from self to self.
        SendMail(cross, fmt.Sprintf("blacklisted caller %v blocked", caller))
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

Besides explicit realm crossing via the `fn(cross, ...)` Gno syntax, implicit
realm crossing occurs when calling a method of a receiver object stored in an
external realm. Implicitly crossing into (borrowing) a receiver object's
storage realm allows the method to directly modify the receiver as well as all
other objects directly reachable from the receiver stored in the same realm as
the receiver. Unlike explicit crosses, implicit crosses do not shift or
otherwise affect the current realm context; `runtime.CurrentRealm()` does not
change unless a method is called like `receiver.Method(cross, args...)`.

Realms hold objects in residence and they also have a Gno address to send and
receive coins from. Coins can only be spent from the current realm context.

### Definition

A crossing-function or crossing-method is that which is declared in a realm and
has as its first argument `cur realm`. The `cur realm` argument must appear as
the first argument of a crossing-function or crossing-method's argument
parameter list. To prevent confusion it is illegal to use anywhere else, and
cannot be used in p packages.

The current realm-context and realm-storage-context changes when a
crossing-function or crossing-method is called with the `cross` keyword in the
first argument as in `fn(cross, ...)`. Such a call is called a "cross-call" or
"crossing-call".

```go
package main
import "gno.land/r/alice/extrealm"

func MyMakeBread(cur realm, ingredients ...any) { ... }

func main(cur realm) {
    MyMakeBread(cross, "flour", "water") // ok -- cross into self.
    extrealm.MakeBread(cross, "flour", "water") // ok -- cross into extrealm
}
```

When a crossing-function or crossing-method is called with `nil` as the first
argument instead of `cross` it is called a non-crossing-call; and no
realm-context nor realm-storage-context change takes place.

```go
package main
import "gno.land/r/alice/extrealm"

func MyMakeBread(cur realm, ingredients ...any) { ... }

func main(cur realm) {
    MyMakeBread(nil, "flour", "water") // ok -- non-crossing.
    extrealm.MakeBread(nil, "flour", "water") // invalid -- external realm function
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

A function declared in p packages when called:

 * inherits the last realm for package declared functions and closures.
 * inherits the last realm when a method is called on unreal receiver.
 * implicitly crosses to the receiver's resident realm when a method of the
   receiver is called. The receiver's realm is also called the "borrow realm".

A function declared in a realm package when called:

 * explicitly crosses to the realm in which the function is declared if the
   function is declared as `func fn(cur realm, ...){...}` (with `cur realm` as the
   first argument). The new realm is called the "current realm".
 * otherwise follows the same rules as for p packages.

`runtime.CurrentRealm()` returns the current realm-context that was last
cross-called to. `runtime.PreviousRealm()` returns the realm-context
cross-called to before the last cross-call. All cross-calls are explicit with
the `cross` keyword, as well as non-crossing-calls of crossing-functions and
crossing-methods with `nil` instead of `cross`.

A crossing function declared in the same realm package as the callee may be
called like `fn(cross, ...)` or `fn(cur, ...)`. When called with `fn(cur, ...)`
there will be no realm crossing, but when called like `fn(cross, ...)` there is
technically a realm crossing and the current realm and previous realm returned
are the same.

The current realm and previous realm do not depend on any implicit crossing to
the receiver's borrowed/storage realm even if the borrowed realm is the last
realm of the call stack. In other words `runtime.CurrentRealm()` may differ
from the Machine's active storage realm (internally `m.Realm`, i.e. the borrow
realm) when a method is called on a receiver residing in a foreign realm.

### Implicit Realm-Storage Borrowing

Besides (explicit) realm-context changes via the `fn(cross, ...)` cross-call
syntax, implicit realm-storage-context changes occur when calling a
non-crossing method of a receiver object residing in different realm-storage.
The realm-storage-context "borrows" to the receiver's realm, which allows the
method to directly modify its receiver and any objects reachable from it that
reside in the same realm-storage. The realm-context does not change (so
`runtime.CurrentRealm()` and `runtime.PreviousRealm()` are unaffected; the
agency of the caller remains the same). However, objects reachable from the
receiver but residing in a *different* realm-storage cannot be modified — the
`DidUpdate()` guard in the runtime enforces that only objects belonging to the
borrowed realm can be mutated.

This borrowing behavior allows non-crossing methods of receivers to behave the
same whether declared in a realm package or p package: p package code copied
over to a realm package, or realm package code copied over to another realm,
have the exact same behavior. Crossing methods of a realm package would still
behave differently when copied over to another realm, as crossing-methods
always change the realm-context and realm-storage-context to the declared
realm.

### Crossing-Method Semantics

When a method is a crossing-method called as
`receiver.Method(cross, args...)`, both the realm-context and
realm-storage-context change to that of the realm package in which the
type/method is declared (which is not necessarily the same as where the
receiver resides). Such a crossing method-call cannot directly modify the
real receiver if it happens to reside in an external realm that differs from
where the type and methods are declared; but it can modify any unreal receiver
or unreal reachable objects. As mentioned previously a non-crossing-call of a
crossing-method will fail during type-checking or at runtime if the receiver
resides in an external realm-storage.

Calls of methods on receivers residing in realms different from the current
realm must *not* be called like `fn(cross, ...)` if the method is not a
crossing function itself, and vice versa. Or it could be said that implicit
crossing is not real realm crossing. (When you sign a document with someone
else's pen it is still your signature; signature:pen :: current:borrowed)

A crossing method declared in a realm cannot modify the receiver if the object
resides in a different realm. However not all methods are required to be
crossing methods, and crossing methods may still read the state of the receiver
(and in general anything reachable is readable).

```go
// Type and methods declared in /r/alice/tokens
package tokens

type Token struct {
    Owner   string
    Balance int
}

// Non-crossing method: borrows receiver's storage realm.
func (t *Token) Debit(amount int) {
    t.Balance -= amount
}

// Crossing method: realm-context and realm-storage-context
// shift to /r/alice/tokens (where the type is declared).
func (t *Token) Transfer(cur realm, to string, amount int) {
    t.Balance -= amount // FAILS if receiver resides outside /r/alice/tokens
}
```

```go
// /r/bob/bob stores a Token
package bob

import "gno.land/r/alice/tokens"

var myToken *tokens.Token // persisted in /r/bob/bob

func Spend(_ realm) {
    // Non-crossing call: storage-context borrows to /r/bob/bob
    // (receiver's realm). Debit() CAN modify myToken.
    myToken.Debit(10) // ok: borrowed storage matches receiver

    // Crossing call: storage-context shifts to /r/alice/tokens
    // (where the type is declared), not /r/bob/bob (where myToken resides).
    // Transfer() CANNOT directly modify myToken.
    myToken.Transfer(cross, "g1...", 10) // fails: receiver in external realm
}
```

New unreal objects reachable from the borrowed realm (or current realm if there
was no method call that borrowed) become persisted in the borrowed realm (or
current realm) upon finalization of the foreign object's method (or function).
(When you put an unlabeled photo in someone else's scrapbook the photo now
belongs to the other person.) In the future the `attach()` function will
prevent a new unreal object from being taken.

For how crossing rules apply to MsgCall, MsgRun, and package initialization,
see [Message Types and Testing](#message-types-and-testing).

## Realm Boundaries

A realm boundary is defined as a change in realm in the call frame stack
from one realm to another, whether explicitly crossed with `fn(cross, ...)`
or implicitly borrow-crossed into a different receiver's storage realm.
A realm may cross into itself with an explicit cross-call.

When a crossing-function or crossing-method is cross-called it shifts the
"current" runtime realm-context to the "previous" runtime realm-context such
that `runtime.PreviousRealm()` returns what used to be returned with
`runtime.CurrentRealm()` before the realm boundary. The current
realm-storage-context is always set to that of realm-context after
cross-calling.

For which call types create boundaries, see the
[summary table](#realm-context-and-realm-storage-context) above. Every
crossing-call creates a realm boundary even when there is no resulting change
in realm-context or realm-storage-context. A non-crossing-call of a
crossing-function or crossing-method (`fn(nil, ...)`) never creates a realm
boundary.

## Realm-Transaction Finalization

Realm-transaction finalization occurs when returning from a realm
boundary. When returning from a cross-call (with `cross`) realm-transaction
finalization will occur even with no change of realm-context or
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
realm, a function (or method) can be declared in the external realm which
references and modifies the aforementioned untainted object directly (by a name
declared outside of the scope of said function or method). Or, the function can
take in as argument an untainted real object returned by another function.

Besides protecting against writing by direct access, the readonly taint also
helps prevent a class of security issue where a realm may be tricked into
modifying something that it otherwise would not want to modify. Since the
readonly taint prohibits mutations even from logic declared in the same realm,
it protects realms against mutating its own object that it doesn't intend to:
such as when a realm's real object is passed as an argument to a mutator
function where the object happens to match the type of the argument.

Objects returned from functions or methods are not readonly tainted. So if
`func (eo object) GetA() any { return eo.FieldA }` then `externalobject.GetA()`
returns an object that is not tainted assuming eo.FieldA was not otherwise
tainted. While the parent object `eo` is still protected from direct
modification by external realm logic, the returned object from `GetA()` can be
passed as an argument to logic declared in the residing realm of `eo.FieldA`
for direct mutation.

Whether or not an object is readonly tainted it can always be mutated by a
method declared on the receiver.

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

Generally you want your methods to be non-crossing. Because they should work
for everyone. They are functions that are pre-bound to an object, and that
object is like a quasi-realm in itself, that could possibly reside and migrate
to other realms. This is consistent with any p code copied over to r realms;
none of those methods would be crossing, and behavior would be the same; stored
in any realm, mostly non-crossing methods that anyone can call. Why is a
quasi-realm self-encapsulated object in need to modify the realm in which it is
declared, by crossing? That's intrusive, but sometimes desired.

You can always cross-call a method from a non-crossing method if you need it.

## Message Types and Testing

### MsgCall

MsgCall may only call crossing functions. This is to prevent potential
confusion for non-sophisticated users. Non-crossing calls of non-crossing
functions of other realms is still possible with MsgRun.

```go
// PKGPATH: gno.land/r/test/test

func Public(_ realm) {

    // Returns (
    //     addr:<origin_caller>,
    //     pkgpath:""
    // ) == testing.NewUserRealm(origin_caller)
    runtime.PreviousRealm()

    // Returns (
    //     addr:chain.PackageAddress("gno.land/r/test/test"),
    //     pkgpath:"gno.land/r/test/test"
    // ) == testing.NewCodeRealm("gno.land/r/test/test")
    runtime.CurrentRealm()

    // Call a crossing function of same realm with crossing
    AnotherPublic(cross)

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

import "gno.land/r/realmA"

func main() {
    // Before main() is called there is an implicit
    // crossing from UserRealm(g1user) to
    // CodeRealm(gno.land/e/g1user/run).

    // Returns (
    //     addr:g1user,
    //     pkgpath:""
    // ) == testing.NewUserRealm(g1user)
    runtime.PreviousRealm()

    // Returns (
    //     addr:g1user,
    //     pkgpath:"gno.land/e/g1user/run"
    // ) == testing.NewCodeRealm("gno.land/e/g1user/run")
    runtime.CurrentRealm()

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
current realm-context of itself. `runtime.PreviousRealm()` refers to the
package deployer both in global var decls and inside `init()` functions. After
that the package deployer is no longer provided, so packages need to remember
the deployer in the initialization phase if needed.

```go
// PKGPATH: gno.land/r/test/test

func init() {
    // Returns (
    //     addr:<origin_deployer>,
    //     pkgpath:""
    // ) == testing.NewUserRealm(origin_deployer)
    // Inside init() and global var decls
    // are the only time runtime.PreviousRealm()
    // returns the deployer of the package.
    // Save it here or lose it forever.
    runtime.PreviousRealm()

    // Returns (
    //     addr:chain.PackageAddress("gno.land/r/test/test"),
    //     pkgpath:"gno.land/r/test/test"
    // ) == testing.NewCodeRealm("gno.land/r/test/test")
    runtime.CurrentRealm()
}

// Same as in init().
var _ = runtime.PreviousRealm()
```

```go
// PKGPATH: gno.land/e/g1user/run

func init() {
    // Returns (
    //     addr:g1user,
    //     pkgpath:""
    // ) == testing.NewUserRealm(g1user)
    runtime.PreviousRealm()

    // Returns (
    //     addr:g1user,
    //     pkgpath:"gno.land/e/g1user/run"
    // ) == testing.NewCodeRealm("gno.land/e/g1user/run")
    runtime.CurrentRealm()
}
```

The same applies for pure package (`/p/`) initialization. During initialization
and tests, `runtime.CurrentRealm()` can return a package path that starts with
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
    "chain/runtime"
    "testing"
)

func TestFoo(t *testing.T) {
    // At first OriginCaller is not set.

    // Override the OriginCaller.
    testing.SetRealm(testing.NewUserRealm("g1user"))

    // Identical behavior:
    testing.SetOriginCaller("g1user")

    // This panics now: seeking beyond the overridden origin frame:
    // runtime.PreviousRealm()

    // Simulate g1user cross-calling Public().
    // Produce a new frame to override
    func() {
        testing.SetRealm(testing.NewCodeRealm("gno.land/r/user/myrealm"))

        runtime.PreviousRealm() // "g1user", ""
        runtime.CurrentRealm()  // bech32(hash("gno.land/r/user/myrealm")), "gno.land/r/user/myrealm"

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
    "chain/runtime"
    "testing"

    "gno.land/r/user/myrealm"
)

func init() {
    // XXX Frame not found, there is no deployer for filetests.
    runtime.PreviousRealm()

    // Returns (
    //     addr:chain.PackageAddress("gno.land/r/test/test")
    //     pkgpath:"gno.land/r/test/test"
    // ) == testing.NewCodeRealm("gno.land/r/test/test")
    runtime.CurrentRealm()
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
    runtime.PreviousRealm()

    // Returns (
    //     addr:g1user,
    //     pkgpath:"gno.land/r/test/test"
    // ) == testing.NewCodeRealm("gno.land/r/test/test")
    runtime.CurrentRealm()

    // gno.land/r/test/test cross-calling
    // gno.land/r/user/myrealm:
    myrealm.Public(cross, ...)
}

// Output:
// XXX
```

## Implementation

Implementation for `runtime.CurrentRealm()` and `runtime.PreviousRealm()` are
defined in `gnovm/stdlibs/chain/runtime/native.gno` and related files in the
directory, while overrides for testing are defined in
`gnovm/tests/stdlibs/testing/context_testing.gno`. All stdlibs functions are
available unless overridden by the latter.

## Proposed Changes

`testing.SetOriginCaller()` may be deprecated in favor of
`testing.SetRealm(testing.NewOriginRealm(user))`.

`testing.NewCodeRealm(path)` may be renamed to
`testing.NewPackageRealm(path)`.
