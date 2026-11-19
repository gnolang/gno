# ADR: cross to write — a realm's state changes only during a crossing call into it

## Status

Proposed (#6495). Supersedes #6207's owner grant (`mutable`) and the
non-crossing write paths of interrealm v2 borrow rules #1–#3 (#5669).

## Context

Code a realm never chose to run wrote its objects: a library method borrowed
the owner's authority (#207's `Ownable` takeover), a library function or
interface value a caller handed in ran as the realm, and any exported
non-crossing realm function wrote the realm's state (opened by #5669).

## Decision

A realm's stored state changes only during a crossing call into that realm
(`fn(cross(cur), …)`, or a crossing closure called with `cross`). Inside it,
the realm's helpers, methods and the library code they name write as usual.

1. Realm code entered without crossing (another realm, a run script, a
   callback) runs read-only. It may read anything, may construct objects (they
   belong to its realm), and may write objects allocated during the current
   read-only stretch, which ends at the next crossing call; no stored object.
   The stretch number lives in the object's existing padding.
2. A library method borrows its receiver's owner only for an object a library
   package owns; a realm's object is a view to library code.
3. Library code a caller picked runs view-only: a function value, closure,
   method value or method expression called through a value; an interface
   call or method value whose package the realm does not import (directly or
   through its imports); a method value built while view-only. A call that
   names the function (`F(x)`, `pkg.F(x)`, `T.M(x)`), or reads it from a
   `/p/` or stdlib package-level variable (`lib.Table[k](x)`, fixed after
   init), acts for the caller. A realm's own function or closure reached
   through library code its crossing call named still acts for the realm.
4. A realm's closure entered from another realm is a call boundary for
   `AssertOriginCall`.
5. Every function value has a maker: the realm whose writable code made it
   (none if read-only); a declared function or method expression used as a
   value is a handle made like a closure. Reached through a value, it acts for
   the realm only if that realm made it.
6. A maker stays home: a non-crossing function value copied, passed (argument,
   receiver, deferred argument) or returned by code whose hold realm (current
   realm, or the one current on entry to read-only code) is not its maker
   arrives with none; a stored value is judged by its object id. Saving it is
   refused into its code's realm or for library code; a third realm keeps it.

Refusals: `<realm>'s code was entered without crossing and is read-only (call
its crossing function)`; `cannot modify object owned by another realm from
library code (call a function of the owning realm)`. `mutable` is removed.
`runtime.FuncPkgPath(fn)` reports a function value's declaring package.

## Alternatives rejected

- **Owner grant (`mutable`, #6207).** One grant hands out every method.
- **Stamping function values with their sender.** Lost in containers and
  variadics. The maker is set by the code making it, none in read-only code,
  and dropped when the value leaves (rule 6); tracking where a value came
  from (a taint bit or side table) was rejected for garbage and gas.
- **Receiver ownership for interface calls.** Cannot tell who picked the code.
- **Unsaved objects stay writable.** Saving depends on finalization.
- **Implicit crossing on every call.** Shifts the caller identity silently.

## Consequences

- Writers are crossing functions or methods, authenticated by
  `cur.Previous()`; a crossing method runs as its realm for anyone holding the
  object, so it must check it (gov/dao's impl admits only the proxy).
- Constructors work without crossing; their objects are read-only to callers.
  A pointer handed out is a view; a value is the caller's copy.
- A callback that writes its own realm is a crossing closure.
- A realm's iterator or builder needs crossing; hand out a library type value.
- Realm-declared types belong to their realm: a realm keeps its own data in
  its own or library types (`impl` got its own vote store).
- Migrated: `gov/dao` (`DAO` methods crossing, `memberstore` crossing writers,
  `SafeExecutor` now admits impl's executions), `sys/users`, `commondao`
  tests, deployment bootstraps, scripts and txtars.
- `authz` handlers run actions synchronously; queueing across realms needs
  crossing action types.
- A view-only or read-only frame writes nothing stored, so it is no realm
  boundary: `recover` in the caller catches a panic raised in it (a refused
  write included); only a crossing call is a boundary, caught by `revive`.

## Known open

- A library in the realm's import tree is trusted: a caller may pick any of
  its types through the realm's interfaces.
- A stdlib method writing its argument can be picked through an interface.
- A function value R made, found by R's code through a reference a caller
  supplied (slice, map, pointer, interface) or another realm keeps, acts for
  R: a gate must not call function values found in caller-supplied
  containers unless it would run them for itself (`zrealm_home_known_*`).
- A library's own function read from its package variable acts for the caller
  when the call names it (`ops[k](b)`); copied to a local first it is a value.

## Tests

`zrealm_ctw_*` pin each rule; `zrealm_maker_*`, `zrealm_handle_*` rule 5;
`zrealm_home_*`, `home_*` rule 6; `zrealm_entry_*`, `zrealm_launder_*` the
library cases; `iface_*`, `assertorigincall`, governance txtars the chain.
