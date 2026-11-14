# ADR: explicit entry — realm state is written only by code acting for that realm

## Status

Accepted.

## Context

A realm stores live references to other realms' objects. A write gate
(`Machine.IsReadonly`) compares the object's owner PkgID with the active storage
realm, but implicit borrow rules let a `/p/` method switch that realm and write
the owner's object. The previous design added a capability flag the owner set on
an object. This ADR replaces it: the authority to write a realm's state belongs
to code, not to an object.

## Decision

A realm's state is written only by code acting for that realm. Code acts for a
realm only if it is written in that realm's source or that source calls it by
name; code that arrived as a value from someone else never does.

What the VM does now (`pushFrameCall`):

1. A function, method or closure declared in an `/r/` realm runs as that realm
   (borrow rule #1, unchanged). Its authority does not depend on who calls it.
2. A `/p/` method on a defined receiver borrows to the receiver's allocating
   realm only when the receiver is owned by a `/p/` package (borrow rule #2).
   A receiver owned by a realm is a view, real or unreal.
3. Borrow rule #3 is gone. A closure not declared in an `/r/` realm, called
   while a realm other than the one it was built under is current, runs
   view-only. A zero stamp counts as another realm.
4. A method value not called by its binder's own code runs view-only, so the
   binder's objects and the caller's cannot be written through it.
5. A top-level `/p/` or stdlib function, or a method expression, called through
   a value (variable, field, argument, package var) by code of another package
   runs view-only. A call that names it (`pkg.F(x)`, `T.M(x)`, marked by the
   preprocessor as `ATTR_NAMED_CALLEE`) or comes from its own package (a
   library's stored default handler) acts for the caller.

A refused write from `/p/` code fails with `cannot modify object owned by another
realm from library code (call a function of the owning realm)`.

The removed `mutable` builtin no longer exists: no builtin, no object flag, no
grant lookup. A realm exposes crossing functions instead, and each crossing
function writes the realm's own variable (`func Set(cur realm, k string, v any)`).

## Alternatives rejected

- **Owner grant (`mutable`, #6207).** One grant hands out every `/p/` method of
  the type, including a method that writes another of the owner's objects (for
  example `CopyTo(dst)`), so the grant is wider than the API the author meant to
  publish.
- **Public-method marker.** A per-method annotation moves the review burden to
  every caller and still cannot say which realm the method acts for.
- **Stamping function values with their sender.** The stamp is lost inside a
  slice, map, struct or variadic argument and when the value is sent from
  view-only code, and a realm that hands a value back stamps it as its own.
- **Parent/home limits.** Restricting where a borrowed call may write handles
  nesting, not identity, and still lets a foreign value act for the owner.

## Consequences

- Realms write through crossing functions; callers call those, and only the
  owner's code performs the write. A returned pointer is a view.
- `grc20reg.Register` uses `runtime.FuncPkgPath` to refuse a token whose
  transfer functions are not declared by the registering realm. A registry no
  longer keeps authority over another realm's object.
- `p/moul/authz`'s `ContractAuthority` no longer guards at-most-once itself; the
  handler realm that stores and runs the action must enforce it.
- The fresh-object corner is documented, not special-cased: a `/p/` object
  a realm function builds and returns carries that realm's stamp from
  allocation, so the caller cannot write it either.

## Known open

- A closure built by a `maketx run` script and stored by a realm is re-stamped
  to that realm, so calling it later runs as the realm.
- A library that calls a function value from its own package accepts it as
  its own, so a realm passing an attacker-chosen function of library L into
  L's callback API lets that function act for the realm.
- A realm's plain non-crossing exported function that writes its own state is
  callable by anyone; such a function is public API by construction.

## Tests

The interrealm filetest corpus pins the view default, the refused foreign
`/p/` write, the method-value refusal, the library-closure refusal and the
fresh-object return. The audit harness keeps its text pattern for the `/p/`
pointer exposure, now described as an exposure rather than a grant.
