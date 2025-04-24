# Interrealm Specification

Gno extends Go's type system with a interrealm rules.  These rules can be
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
that begin with the `crossing()` statement are called "crossing" functions".

A crossing function declared in a realm different than the last explicitly
crossed realm *must* be called like `cross(fn)(...)`. That is, functions of
calls that result in explicit realm crossings must be wrapped with `cross()`.

`std.CurrentRealm()` returns the current realm last explicitly crossed to.

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
receiver (and in general anything reacheable is readible).

New unreal objects reachable from the borrowed realm (or current realm if there
was no method call that borrowed) become persisted in the borrowed realm (or
current realm) upon finalization of the foreign object's method (or function).
(When you put an unlabeled photo in someone else's scrap book the photo now
belongs to the other person). In the future we will introduce an `attach()`
function to prevent a new unreal object from being taken.

MsgCall can only call (realm) crossing functions.

MsgRun will run a file's `main()` function in the user's realm and may call 
both crossing functions and non-crossing functions.

A realm package's initialization (including init() calls) execute with current
realm of itself, and it `std.PreviousRealm()` will panic unless the call stack
includes a crossing function called like `cross(fn)(...)`.

### Justifications

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
the caller's current realm. Otherwise two mutable (upgreadeable) realms cannot
export trust unto the chain because functions declared in those two realms can
be upgraded.

Both `crossing()` and `cross(fn)(...)` statements may become special syntax in
future Gno versions.
