# Interrealm Specification

## Introduction

All modern popular programming langauges are designed for a single programmer
user.  Programming languages support the importing of program libraries
natively for components of the single user's program, but this does not hold
true for interacting with components of another user's (other) program. Gno is
an extension of the Go language for multi-user programming. Gno allows a
massive number of programmers to iteratively and interactively develop a single
shared program such as Gno.land.

The added dimension of the program domain means the language should be extended
to best express the complexities of programming in the inter-realm (inter-user)
domain. In other words, Go is a restricted subset of the Gno language in the
single-user context. (In this analogy client requests for Go web servers don't
count as they run outside of the server program).

### Realm Write Access

Objects that are directly or indirectly reachable (referenced) from the realm
package's global variables (and are not already associated with another realm)
are said to reside in the realm (memory space).

**An object can only be mutated if the object resides in the same realm as the
current realm in the Gno Machine's execution context.**

Go's language rules for value access through selector/index expressions are the
same within the same realm, but exposed values through selector/index
expressions are read-only when performed by an external realm; a realm cannot
directly modify another realm's objects.  Thus a Gno package's global variales
even when exposed (e.g. `var MyGlobal int = 1`) is safe from external
manipulation (e.g.  `import "realm"; realm.MyGlobal = 2`). For users to
manipulate them a function or method must be provided.

Realm crossing occurs when a function is called with the Gno `cross(fn)(...)`
syntax.

```go
package main
import "gno.land/r/alice/realm1"

func main() {
    bread := cross(realm1.MakeBread)("flour", "water")
```

(In Linux/Unix operating systems user processes can cross-call into the kernel
by calling special syscall functions, but user processes cannot directly
cross-call into other users' processes. This makes the GnoVM a more complete
multi-user operating system than traditional operating systems)

Besides explicit realm crossing via the `cross(fn)(...)` Gno syntax, implicit
realm crossing occurs when calling a method of a receiver object stored in an
external realm. Implicitly crossing into (borrowing) a receiver object's
storage realm allows the method to directly modify the receiver as well as all
other objects directly reachable from the receiver stored in the same realm as
the receiver. Unlike explicit crosses, implicit crosses do not shift or
otherwise effect the current realm context; `std.CurrentRealm()` does not
change unless a method is called like `cross(receiver.Method)(args...)`.

Realms hold objects in residence and they also have a Gno address to send and
receive coins from. Coins can only be spent from the current realm context.

### Realm Boundaries

A realm boundary is defined as a change in realm in the call frame stack
from one realm to another, whether explicitly crossed with `cross(fn)()`
or implictly borrow-crossed into a different receiver's storage realm.
A realm may cross into itself with an explicit cross-call.

When returning from a realm boundary, all new reachable objects are assigned
object IDs and stored in the current realm, ref-count-zero objects deleted
(full "disk-persistent cycle GC" will come after launch) and any modified
ref-count and Merkle hash root computed. This is called realm finalization.

## Readonly Taint Specification

`otherrealm.Foo` is a direct selector expression so the value is tainted
with the `N_Readonly` attribute.

Same for `externalobject.FieldA` where `externalobject` resides in an external
realm (as compared to the current realm context).

Same for `externalobject[0]`, direct index expressions also taint the resulting
value with the `N_Readonly` attribute. 

The readonly taint follows any subsequently derived values and cannot be
overcome.

The readonly taint also prohibits mutations even if the base object resides in
the current realm. This protects realms against mutating objects it doesn't
intend to (e.g. by an exploit where a realm's own object is passed to the same
realm's mutator function by a malicious third party, where the first object was
not intended to be passed in that way).

Objects returned from functions or methods are not readonly tainted. So if
`func (eo object) GetA() any { return eo.FieldA }` then `externalobject.GetA()`
returns an object that is not tainted. The return object's fields would still
be protected from external realm direct modification, but the return object
could be passed back to the realm for mtuation; or the object may be mutated
through its own methods.

## `cross(fn)()` and `crossing()` Specification

Gno extends Go's type system with interrealm rules. These rules can be
checked during the static type-checking phase (but at the moment they are
partially dependent on runtime checks).

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
   function begins with a `crossing()` statement. The new realm is called the
   "current realm".
 * otherwise follows the same rules as for p packages.

The `crossing()` statement must be the first statement of a function's body.
It is illegal to use anywhere else, and cannot be used in p packages. Functions
that begin with the `crossing()` statement are called "crossing functions".

A crossing function declared in a realm different than the last explicitly
crossed realm *must* be called like `cross(fn)(...)`. That is, functions of
calls that result in explicit realm crossings must be wrapped with `cross()`.

`std.CurrentRealm()` returns the current realm that was last explicitly crossed
to.

`std.PreviousRealm()` returns the realm explicitly crossed to before that.

A crossing function declared in the same realm package as the callee may be
called normally OR like `cross(fn)(...)`. When called normally there will be no
realm crossing, but when called like `cross(fn)(...)` there is technically a
realm crossing and the current realm and previous realm returned are the same.

The current realm and previous realm do not depend on any implicit crossing to
the receiver's borrowed/storage realm even if the borrowed realm is the last
realm of the call stack equal to `m.Realm`. In other words `std.CurrentRealm()`
may be different than `m.Realm` (the borrow realm) when a receiver is called on
a foreign object.

Calls of methods on receivers residing in realms different than the current
realm must not be called like `cross(fn)(...)` if the method is not a
crossing function itself, and vice versa. Or it could be said that implicit
crossing is not real realm crossing. (When you sign a document with someone
else's pen it is still your signature; signature:pen :: current:borrowed.

A crossing method declared in a realm cannot modify the receiver if the object
resides in a different realm. However not all methods are required to be
crossing methods, and crossing methods may still read the state of the
receiver (and in general anything reachable is readable).

New unreal objects reachable from the borrowed realm (or current realm if there
was no method call that borrowed) become persisted in the borrowed realm (or
current realm) upon finalization of the foreign object's method (or function).
(When you put an unlabeled photo in someone else's scrapbook the photo now
belongs to the other person). In the future we will introduce an `attach()`
function to prevent a new unreal object from being taken.

MsgCall can only call (realm) crossing functions.

MsgRun will run a file's `main()` function in the user's realm and may call 
both crossing functions and non-crossing functions.

A realm package's initialization (including `init()` calls) execute with current
realm of itself, and it `std.PreviousRealm()` will panic unless the call stack
includes a crossing function called like `cross(fn)(...)`.

### `cross` and `crossing` Design Goals

P package code should behave the same even when copied verbatim in a realm
package.

Realm crossing with respect to `std.CurrentRealm()` and `std.PreviousRealm()`
is important enough to be explicit and warrants type-checking.

A crossing function of a realm should be able to call another crossing function
of the same realm without necessarily explicitly crossing realms.

Sometimes the previous realm and current realm must be the same realm, such as
when a realm consumes a service that it offers to external realms and users.

A method should be able to modify the receiver and associated objects of the
same borrowed realm.

A method should be able to create new objects that reside in the same realm by
default in order to maintain storage realm consistency and encapsulation and
reduce fragmentation.

In the future an object may be migrated from one realm to another when it loses
all references in one realm and gains references in another. The behavior of
the object should not change after migration because this type of migration is
implicit and generally not obvious without more language features.

Code declared in p packages (or declared in "immutable" realm packages) can
help different realms enforce contracts trustlessly, even those that involve
the caller's current realm. Otherwise two mutable (upgradeable) realms cannot
export trust unto the chain because functions declared in those two realms can
be upgraded.

Both `crossing()` and `cross(fn)(...)` statements may become special syntax in
future Gno versions.

## `attach()`

## `panic()` and `revive(fn)`

`panic()` behaves the same within the same realm boundary, but when a panic
crosses a realm boundary (as defined in [Realm
Finalization](#realm-finalization)) the Machine aborts the program. This is
because in a multi-user environment it isn't safe to let the caller recover
from realm panics that often leave the state in an invalid state.

This would be sufficient, but we also want to write our tests to be able
to detect such aborts and make assertions. For this reason Gno provides
the `revive(fn)` builtin.

```go
abort := revive(func() {
    cross(func() {
        crossing()
        panic("cross-realm panic")
    })
})
abort == "cross-realm panic"
```

`revive(fn)` will execute 'fn' and return the exception that crossed
a realm finalization boundary.

This is only enabled in testing mode (for now), behavior is only partially
implemented. In the future `revive(fn)` will be available for non-testing code,
and the behavior will change such that `fn()` is run in transactional
(cache-wrapped) memory context and any mutations discarded if and only if there
was an abort.

TL;DR: `revive(fn)` is Gno's builtin for STM (software transactional memory).

## Application

P package code cannot contain crossing functions, nor use `crossing()`. P
package code also cannot import R realm packages. But code can call named
crossing functions e.g. those passed in as parameters.

You must declare a public realm function to be `crossing()` if it is intended to
be called by end users, because users cannot MsgCall non-crossing functions
(for safety/consistency) or p package functions (there's no point).

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
quasi-realm self-encapsulated Object in need to modify the realm in which it is
declared, by crossing? That's intrusive, but sometimes desired.

You can always cross-call a method from a non-crossing method if you need it.

Implementation for `std.CurrentRealm()` and `std.PreviousRealm()` are defined
in `stdlibs/std/native.gno/go` and related files in the directory, while
overrides for testing are defined in `testing/stdlibs/std/std.gno/go`. All
stdlibs functions are available unless overridden by the latter.

`std.CurrentRealm()` shifts to `std.PreviousRealm()` if and only if a function
is called like `cross(fn)(...)`.

### MsgCall

MsgCall may only call crossing functions. This is to prevent potential
confusion for non-sophisticated users. Non-crossing calls of non-crossing
functions of other realms is still possible with MsgRun.

```go
// PKGPATH: gno.land/r/test/test

func Public() {
    crossing()

    // Returns (
    //     addr:<origin_caller>,
    //     pkgpath:""
    // ) == std.NewUserRealm(origin_caller)
    std.PreviousRealm()

    // Returns (
    //     addr:<derived_from "gno.land/r/test/test">,
    //     pkgpath:"gno.land/r/test/test"
    // ) == std.NewCodeRealm("gno.land/r/test/test")
    std.CurrentRealm()

    // Already in gno.land/r/test/test realm,
    // no need to cross unless the intent
    // is to call AnotherPublic() as a consumer
    // in which case cross(AnotherPublic)() needed.
    AnotherPublic()
}

func AnotherPublic() {
    crossing()
    ...
}
```

### MsgRun

```go
// PKGPATH: gno.land/r/g1user/run

import "gno.land/r/realmA"

func main() {
    // There is assumed to be in "frame -1"
    // a crossing from UserRealm(g1user) to
    // CodeRealm(gno.land/r/g1user/run) before
    // main() is called, so crossing() here
    // is redundant.
    // crossing()

    // Returns (
    //     addr:g1user,
    //     pkgpath:""
    // ) == std.NewUserRealm(g1user)
    std.PreviousRealm()

    // Returns (
    //     addr:g1user,
    //     pkgpath:"gno.land/r/g1user/run"
    // ) == std.NewCodeRealm("gno.land/r/g1user/run")
    std.CurrentRealm()

    realmA.PublicNoncrossing()
    cross(realmA.PublicCrossing)()
}
```

Notice in `gnovm/pkg/gnolang/misc.go`, the following:

```go
// For keeping record of package & realm coins.
// If you need the bech32 address it is faster to call DerivePkgBech32Addr().
func DerivePkgCryptoAddr(pkgPath string) crypto.Address {
	b32addr, ok := IsGnoRunPath(pkgPath)
	if ok {
		addr, err := crypto.AddressFromBech32(b32addr)
		if err != nil {
			panic("invalid bech32 address in run path: " + pkgPath)
		}
		return addr
	}
	// NOTE: must not collide with pubkey addrs.
	return crypto.AddressFromPreimage([]byte("pkgPath:" + pkgPath))
}

func DerivePkgBech32Addr(pkgPath string) crypto.Bech32Address {
	b32addr, ok := IsGnoRunPath(pkgPath)
	if ok {
		return crypto.Bech32Address(b32addr)
	}
	// NOTE: must not collide with pubkey addrs.
	return crypto.AddressFromPreimage([]byte("pkgPath:" + pkgPath)).Bech32()
}
```

These function names are distinct from what is available in Gno
from `stdlibs/std/crypto.gno`:

```go
// Returns a crypto hash derived pkgPath, unless pkgPath is a MsgRun run path,
// in which case the address is extracted from the path.
func DerivePkgAddr(pkgPath string) Address {
	addr := derivePkgAddr(pkgPath) <-- calls gno.DerivePkgBech32Addr()
	return Address(addr)
}
```

1. `std.DerivePkgAddr("gno.land/r/name123/realm")` - bech32 from hash(path)
2. `std.DerivePkgAddr("gno.land/r/g1user/run")` - bech32 substring "g1user"

Therefore in the MsgRun file's `init()` function the previous realm and current
realm have different pkgpaths (the origin caller always has empty pkgpath) but
the address is the same.

### MsgAddPackage

During MsgAddPackage `std.PreviousRealm()` refers to the package deployer both
in global var decls as well as inside `init()` functions. After that the
package deployer is no longer provided, so packages need to remember the
deployer in the initialization phase if needed.

```go
// PKGPATH: gno.land/r/test/test

func init() {
    // Returns (
    //     addr:<origin_deployer>,
    //     pkgpath:""
    // ) == std.NewUserRealm(origin_deployer)
    // Inside init() and global var decls
    // are the only time std.PreviousRealm()
    // returns the deployer of the package.
    // Save it here or lose it forever.
    std.PreviousRealm()

    // Returns (
    //     addr:<origin_deployer>,
    //     pkgpath:"gno.land/r/test/test"
    // ) == std.NewCodeRealm("gno.land/r/test/test")
    std.CurrentRealm()
}

// Same as in init().
var _ = std.PreviousRealm()
```

```go
// PKGPATH: gno.land/r/g1user/run

func init() {
    // Returns (
    //     addr:g1user,
    //     pkgpath:""
    // ) == std.NewUserRealm(g1user)
    std.PreviousRealm()

    // Returns (
    //     addr:g1user,
    //     pkgpath:"gno.land/r/g1user/run"
    // ) == std.NewCodeRealm("gno.land/r/g1user/run")
    std.CurrentRealm()
}
```

The same applies for p package initialization. Initialization and tests are the
only times that `std.CurrentRealm()` will return a p package path that starts
with "/p/" instead of "/r/". The package is technically still mutable during
initialization.

### Testing overrides with stdlibs/testing

The `gnovm/tests/stdlibs/testing/context_testing.gno` file provides functions
for overriding frame details from Gno test code.

`testing.SetRealm(std.NewUserRealm("g1user"))` is identical to
`testing.SetOriginCaller("g1user")`. Both will override the Gno frame to make it
appear as if the current frame is the end user signing with a hardware signer.
Both will also set `ExecContext.OriginCaller` to that user. One of these will
become deprecated.

#### Gno test cases with `_test.gno` like `TestFoo(t *testing.T)`

```go
// PKGPATH: gno.land/r/user/myrealm
package myrealm

import (
    "std"
    "stdlibs/testing"
)

func TestFoo(t *testing.T) {
    // At first OriginCaller is not set.

    // Override the OriginCaller.
    testing.SetRealm(std.NewUserRealm("g1user"))

    // Identical behavior:
    testing.SetOriginCaller("g1user")

    // This panics now: seeking beyond the overridden origin frame:
    // std.PreviousRealm()

    // Simulate g1user cross-calling Public().
    // Produce a new frame to override
    func() {
        testing.SetRealm(std.SetCodeRealm("gno.land/r/user/myrealm"))

        std.PreviousRealm() // "g1user", ""
        std.CurrentRealm()  // bech32(hash("gno.land/r/user/myrealm")), "gno.land/r/user/myrealm"

        Public(...) // already in "gno.land/r/user/myrealm"
    }()

    // The following is identical to the above,
    // but not possible in p packages which
    // cannot import realms.
    cross(Public)(...)
}
```

#### Gno filetest cases with `_filetest.gno`

```go
// PKGPATH: gno.land/r/test/test
package test

import (
    "std"
    "stdlibs/testing"

    "gno.land/r/user/myrealm"
)

func init() {
    // XXX Frame not found, there is no deployer for filetests.
    std.PreviousRealm()

    // Returns (
    //     addr:std.DerivePkgAddr("gno.land/r/test/test")
    //     pkgpath:"gno.land/r/test/test"
    // ) == std.NewCodeRealm("gno.land/r/test/test")
    std.CurrentRealm()
}

func main() {
    // There is assumed to be in "frame -1"
    // a crossing from UserRealm(g1user) to
    // CodeRealm(gno.land/r/test/test) before
    // main() is called, so crossing() here
    // is redundant.
    // crossing()

    // Returns (
    //     addr:g1user,
    //     pkgpath:""
    // ) == std.NewUserRealm(g1user)
    std.PreviousRealm()

    // Returns (
    //     addr:g1user,
    //     pkgpath:"gno.land/r/test/test"
    // ) == std.NewCodeRealm("gno.land/r/test/test")
    std.CurrentRealm()

    // gno.land/r/test/test cross-calling
    // gno.land/r/user/myrealm:
    cross(myrealm.Public)(...)
}

// Output:
// XXX
```

## Future Work

`std.SetOriginCaller()` should maybe be deprecated in favor of
`std.SetRealm(std.NewUserRealm(user))` renamed to
`std.SetRealm(std.NewOriginRealm(user))`.

`std.SetRealm(std.NewCodeRealm(path))` renamed to
`std.SetRealm(std.NewPackageRealm(path))`.
