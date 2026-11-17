# ADR: cross to write — a realm's state changes only during a crossing call into it

## Status

Proposed (#6495). Supersedes #6207's owner grant (`mutable`) and the
non-crossing write paths of interrealm v2 borrow rules #1–#3 (#5669).

## Context

A realm's objects could be written by code the realm never chose to run for
it: a library method holding the realm's object borrowed the owner's
authority (#207's `Ownable` takeover), a library function or interface value a
caller handed in ran as the realm, and any exported non-crossing realm
function or method wrote the realm's state for whoever called it. Before
#5669 the last case was refused but a caller's callback ran with the
victim's authority; #5669 fixed the callback and opened the writer.

## Decision

A realm's stored state changes only during a crossing call into that realm
(`fn(cross(cur), …)`, or a crossing closure called with `cross`). Inside it,
the realm's helpers, methods and the library code they name write as usual.

1. Realm code entered without crossing (another realm, a run script, a
   callback) runs read-only. It may read anything, may construct objects (they
   belong to its realm), and may write objects allocated during the current
   read-only stretch, which ends at the next crossing call. It writes no stored
   object, its own or the caller's.
2. A library method borrows its receiver's owner only for an object a library
   package owns; a realm's object is a view to library code.
3. Library code a caller picked runs view-only: a function value, closure,
   method value or method expression called through a value; an interface
   call or method value whose package the realm does not import (directly or
   through its imports); a method value built while view-only. A call that
   names the function (`F(x)`, `pkg.F(x)`, `T.M(x)`) acts for the caller.
4. A realm's closure entered from another realm counts as a call boundary for
   `AssertOriginCall`.

Refusals: `<realm>'s code was entered without crossing and is read-only (call
its crossing function)`; `cannot modify object owned by another realm from
library code (call a function of the owning realm)`. `mutable` is removed.
`runtime.FuncPkgPath(fn)` reports a function value's declaring package.

## Alternatives rejected

- **Owner grant (`mutable`, #6207).** One grant hands out every method of the
  type, including ones that write the owner's other objects.
- **Stamping function values with their sender.** Lost in containers and
  variadics, and a value handed back carries the victim's stamp.
- **Receiver ownership for interface calls.** A struct value is copied under
  the realm and a stored pointer becomes the realm's; it cannot tell who picked
  the code.
- **Unsaved objects stay writable.** When an object is saved depends on
  finalization, which the source does not show.
- **Implicit crossing on every call.** Shifts the caller identity silently
  (the Solidity `msg.sender` problem) and makes every helper a boundary.

## Consequences

- Writers are crossing functions or methods; `cur.Previous()` authenticates
  them, so the `(_ int, rlm realm)` + `IsCurrent()` idiom is not needed.
- Constructors work without crossing; their objects are read-only to callers.
  A pointer handed out is a view; a value is the caller's copy.
- A callback that writes its own realm is a crossing closure.
- Realm-declared types belong to their realm: a realm keeps its own data in
  its own or library types (`impl` got its own vote store).
- Migrated: `gov/dao` (`DAO` methods crossing, `memberstore` crossing writers,
  `SafeExecutor` now admits impl's executions), `sys/users`, `commondao`
  tests, deployment bootstraps, scripts and txtars.
- `authz` handlers that run actions synchronously work; queueing an action in
  one realm and running it from another needs crossing action types.
- A panic in read-only code crosses a realm boundary: `recover` in the caller
  does not catch it, `revive` does.

## Known open

- A library in the realm's import tree is trusted: a caller may pick any of
  its types through the realm's interfaces.
- A stdlib method that writes its argument can be picked through an interface.

## Tests

`zrealm_ctw_*` pin each rule; `zrealm_entry_*` and `zrealm_launder_*` the
library cases; `iface_*`, `assertorigincall` and the governance txtars the
chain path.
