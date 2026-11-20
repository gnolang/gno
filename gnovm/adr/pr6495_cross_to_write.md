# ADR pr6495: cross to write

Accepted. Supersedes the owner grant (`mutable`, #6207) and interrealm v2's
borrow rules (#5669). Teaching companion: `docs/resources/gno-interrealm.md`.

## Context

A realm never chose all the code that ran against its stored objects: library
methods borrowed its authority, caller-supplied functions ran as the realm, and
exported non-crossing functions wrote. Only a crossing call into the realm writes.

## The rules

**R1 Cross to write.** A realm's stored state changes only during a crossing
call into it: `func F(cur realm, …)` called `F(cross(cur), …)`. Any other entry
runs read-only. Library (`/p/`) code never writes another realm's objects.

**R2 Only pointers cross.** A struct or array value of a type a realm declares
is created and copied only by that realm's code; others may hold a pointer,
read fields and call methods. Exceptions: the receiver copy of a direct call to
a value-receiver method, and an unnamed struct whose fields are all exported
and which holds no realm's struct or array: it may be passed by value, and the
copy belongs to whoever made it. An error type a realm hands out is a pointer;
`defer x.M()` on another realm's value needs `defer func() { x.M() }()`. A
by-value copy copies every level.

**S3 A function value writes only in the crossing call that made it.** A
non-crossing function value — closure, bound method value, declared function
used as a value — called by value writes only during the crossing call that
made it; every crossing call (the tx entry included) takes a fresh id and each
value records it. Made in read-only code, at init, or loaded from the store:
id 0, which never matches, and the call runs `enterReadOnly`. "By value" is
`ATTR_NAMED_CALLEE != true`; a named declared function or `/p` package
variable is exempt, a realm's own variable holding a function is not. The id
is memory-only, never stored: a saved value comes back 0. It lives in a
per-Allocator side table, not a struct field, so a missed copy cannot carry
write authority and a map miss fails closed; alloc constants and gas are
unchanged. Rejected: S (a fresh value writes anywhere) accepts an escaped
container re-entered from a crossing callback; S2 (write while its maker is
current) accepts a value made in crossing call N run in M≠N; M' (a crossing
closure crosses only from its maker realm) refuses the sanctioned hand-off to
another realm's executor (`r/gov/dao` into `r/sys/validators`) and is dropped.

**S3d: the write is allowed, the save is not.** Inside a read-only stretch
entered by S3, a write to the read-only realm's own object whose ObjectID is
not finalized is allowed and recorded (`ObjectInfo.s3Written`, memory only);
finalize panics if a marked object lands in created/updated. Any read-only
entry that writes a heap item it did not make marks it the same way; the mark
is per object, so a by-value copy of a marked struct stored by the realm's own
crossing code is not refused (`zrealm_s3u_rv3_2`). Stored objects stay refused
at write time. Cost: one flag per S3 write, a finalize scan only when a flag
was set; no amino, hash, state byte or gas constant changes.

**R4 Makers stay home.** When another realm's code hands a function value over
(argument, receiver, result) or copies it, what arrives has no maker.
Publishing a function is never a grant of write power. Not covered: a caller
hands a realm's crossing function a reference (slice, map, pointer, object) to
that realm's own stored functions; your crossing function must not call
function values found in caller-supplied containers.

**R6 One import rule.** Code reached through an interface or a method value is
trusted only if its package is in the calling realm's import tree (its imports
and theirs, standard library included, test files excluded); otherwise it runs
view-only.

Lint (`gno lint`, warnings): an exported function passing or returning its
realm's struct by value, also inside an interface; a crossing function that
calls a function reached through a non-function parameter. Known gaps: values
passed through two or more locals, another function of the realm, a closure
capture. Refusals: `<realm>'s code was entered without crossing and is
read-only (call its crossing function)`; `cannot modify object owned by another
realm from library code (call a function of the owning realm)`; a refused copy
or allocation of another realm's type. An S3-entered stretch says instead:
`<realm>'s code is read-only here: a function value writes only in the crossing
call that made it`. `runtime.FuncPkgPath(fn)` names a func value's package.

## Alternatives rejected

- **Owner grant (`mutable`, #6207).** One grant hands out every method.
- **Stamping function values with their sender.** Lost in containers and variadics; a taint bit or side table was rejected for garbage and gas.
- **Receiver ownership for interface calls** (cannot tell who picked the code); **unsaved objects stay writable** (saving depends on finalization).
- **Implicit crossing on every call.** Shifts the caller identity silently.

## Consequences

- Writers are crossing functions or methods, authenticated by `cur.Previous()`; a crossing method runs as its realm for anyone holding the object, so it must check the caller. Constructors work without crossing; their objects are read-only to callers, and a pointer handed out is a view.
- A callback that writes its own realm is a crossing closure. A realm's iterator or builder needs crossing; hand out a library type value and keep the realm's data in its own or library types.

## Known open

- A library in the realm's import tree is trusted for interface calls; type-assert caller-supplied interface values to the types you expect.
- A function value R made, found through a caller-supplied reference (slice, map, pointer, object) or kept by another realm, acts for R; do not call such values.
- A function value a library stores in a package variable acts for the caller when called in place (`lib.Table[k](x)`); copied to a local first, it has no maker and runs view-only (no test yet).

## Upgrade notes

- Struct errors returned to other realms become pointer errors (`sys/users`).
- Library containers receive realm structs as pointers (`commondao`).
- A constructor whose closures the realm's crossing functions later run must itself be crossing (`gov/dao NewGovDAO`).
- State saved before the change keeps its old function-value references: the rules hold chain-wide only from a fresh genesis or a migration. Gas and storage deposits shift where a saved handle becomes its own object.
