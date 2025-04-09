# Interrealm Specification

Gno extends Go's type system with a interrealm rules.  These rules can be
checked during the static type-checking phase (but at the moment they are
partially dependent on runtime checks).

All functions in Gno execute under a realm context as determined by the call
(frame) stack. Objects that reside in a realm can only be modified if the realm
context is the same.

A function declared in p packages when called: 

 * Inherits the last realm for package declared functions and closures.
 * Inherits the last realm when a method is called on unreal receiver.
 * Implicitly switches to the realm in which the receiver resides when a method
   is called.

A function declared in a realm package when called:

 * Explicitly switches to the realm in which the function is declared if the
   function begins with a `switchrealm()` statement.
 * Otherwise follows the same rules as for p packages.

The `switchrealm()` statement must be the first statement of a function's body.
It is illegal to use anywhere else, and cannot be used in p packages.

Functions that begin with the `switchrealm()` statement are called "switching"
functions".

A switching function declared in a realm different than the last explicitly
switched realm *must* be called like `withswitch(fn)(...)`.

A switching function declared in the same realm may be called normally OR like
`withswitch(fn)(...)`. 

`std.CurrentRealm()` returns the last realm explicitly switched to.

`std.PreviousRealm()` returns the second last realm explicitly switched to.

The current realm and previous realm as returned by the above functions may
refer to the same realm if a function calls another switching function both
declared in the same realm like `withswitch(fn)(...)`.

The current realm and previous realm as returned by the above functions do not
depend on any implicit switches to the receiver's (storage) realm. That is,
`std.CurrentRealm()` may return a different realm than the current "storage
realm" implicitly switched to.

Calls of methods on receivers stored in realms different than the current realm
must not be called like `withswitch(fn)(...)` if the method is not a switching
function, and vice versa.

A switching method declared in a realm cannot modify the receiver if the object
is stored in a different realm. However not all methods are required to be
switching methods, and switching methods may still read the state of the
receiver (and in general anything reacheable is readible).

MsgCall can only call (realm) switching functions.

MsgRun will run a file's `main()` function in the user's realm and may call 
both switching functions and non-switching functions.

Both `switchrealm()` and `withswitch(fn)(...)` statements may become special
syntax in future Gno versions.

A realm package's initialization (including init() calls) execute with current
realm of itself, and it `std.PreviousRealm()` will panic unless the call stack
includes a switching function called like `withswitch(fn)(...)`.

New unreal objects reachable from the current realm (whether implicitly or
explicitly switched) become persisted in the current realm upon finalization of
(return from) the function or method.

### Justifications

P package code should behave the same even when copied verbatim in a realm
package.

Realm switching with respect to `std.CurrentRealm()` and `std.PreviousRealm()`
is important enough to be explicit and warrants type-checking.

A switching function of a realm should be able to call another switching
function of the same realm without necessarily explicitly switching realms.

Sometimes the previous realm and current realm must be the same realm, such as
when a realm consumes a service that it offers to external realms and users.

A method should be able to modify the receiver and associated objects of the
same storage realm.

A method should be able to create new objects that reside in the same storage
by default in order to maintain storage realm consistency and encapsulation,
and to reduce storage realm fragmentation.

Code declared in p packages (or declared in "immutable" realm packages) can
help different realms enforce contracts trustlessly, even those that involve
the caller's current realm. Otherwise two mutable (upgreadeable) realms cannot
export trust unto the chain because functions declared in those two realms can
be upgraded.
