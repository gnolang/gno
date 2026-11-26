# ADR pr6495: cross to write

Accepted. Supersedes the owner grant (`mutable`, #6207) and interrealm v2's
borrow rules (#5669). Teaching companion: `docs/resources/gno-interrealm.md`.

## Context
A realm never chose all the code that ran against its stored objects: library
methods borrowed its authority, caller-supplied functions ran as the realm, and
exported non-crossing functions wrote. Only a crossing call into the realm writes.

## Decision
**R1 Cross to write.** A realm's stored state changes only during a crossing
call into it: `func F(cur realm, …)` called `F(cross(cur), …)`. Any other entry
runs read-only. Library (`/p/`) code never writes another realm's objects.

**R2 A function value that is not crossing writes only when it is:**
1. the realm's own code called by name, or a local helper (`f := func(){ n++ }; f()`);
2. a literal written in place as a call argument, while that call runs, lent
   only to the realm's own functions or to a `/p/` library function reached by
   name (`lib.Each(xs, func(x int){ sum += x })`), never to another realm's function;
3. a literal called in the expression that writes it (`func(){ n++ }()`).

Anything else runs read-only and a write is refused with `<realm>'s code is
read-only here: a function value writes only in the crossing call that made it`.
The legal form for every other case is a crossing closure, `func(cur realm, …)`
called with `cross`; it is the default for callbacks. The loan (clause 2) ends
when the call returns, when the callee stores or returns the value, and for a
deferred or variadic argument; a callee parameter that is only called or passed
on keeps it. Marks live in the preprocessed code (`ATTR_LOCAL_HELPER`,
`ATTR_NONESC_PARAMS`, `ATTR_NAMED_CALLEE`); nothing is stored, so amino, hashes
and gas constants do not change.

**R2-save.** Inside a read-only stretch, a write to the realm's own object that
is not yet saved is allowed and flagged; finalize refuses to save a flagged
object (`an object written by read-only code would be stored`).

**R2c Only imported library code writes.** Code reached through an interface or
a method value is trusted only if its package is in the calling realm's import
tree (imports and theirs, standard library included, tests excluded).

**R3 Types stay home: only pointers cross.** A struct or array value of a type a
realm declares is created and copied only by that realm's code. Exceptions: the
receiver copy of a direct value-receiver call, and an unnamed struct of exported
fields holding no realm's struct or array. Errors handed out are pointers;
`defer x.M()` on another realm's value needs `defer func() { x.M() }()`.

**R4 The ticket stays in its call.** `cur` is never stored, reassigned,
addressed or forged; banker and `cur.Sub` need the live `cur`.

**Panics.** `recover()` catches a panic unless it leaves a `cross` call; a
library borrow (a call into `/p/` or the standard library at run time) no
longer stops it. Tests expect a panic that leaves a `cross` call with
`revive(func(){ … })`.

Lint (`gno lint`, warnings): an exported function passing or returning its
realm's struct by value; a crossing function calling through a non-function
parameter. Other refusals: `<realm>'s code was entered without crossing and is
read-only (call its crossing function)`; `cannot modify object owned by another
realm from library code (call a function of the owning realm)`.

## Alternatives rejected
- **Copy on hand-out** (a function value writes only in the crossing call that made it, while private to it; a stored or handed-out value becomes a copy without write power): it needed a per-call marker and a deep swap on every attach and realm change, and its edge cases were hard to state to authors.
- **Owner grant (`mutable`, #6207)**: one grant hands out every method. **Stamping function values with their sender**: lost in containers and variadics.
- **Receiver ownership for interface calls** (cannot tell who picked the code); **implicit crossing on every call** (shifts the caller identity silently).

## Consequences
- Writers are crossing functions or methods, authenticated by `cur.Previous()`; a crossing method runs as its realm for anyone holding the object, so it must check the caller.
- A callback that writes its own realm is a crossing closure unless it is a literal lent in place to a library or to the realm's own function.
- Recovered panics: a panic out of a library call is now caught by the caller's `recover()` (`recover_fv_borrow`, `recover_fv_libmade`); one leaving a `cross` call still aborts (`recover_fv_cross`).

## Known open
- A callee not reached by name lends nothing: a local copy (`each := lib.Each; each(xs, f)`) or a function literal kept in a variable (`var Apply = func(f func()) { f() }`) runs `f` read-only (`zrealm_ctw_owncb`).
- Another realm's function gets no loan, even a plain one (`zrealm_t55_other_plain`): give it a crossing callback.
- A method value never writes (`zrealm_chald_lend_lib`); call the method directly.
- `lib.Default.Each(func(int){ count++ })`, a method of a library-owned variable, is refused (`zrealm_chale_plibeach`).
- Swift and Rust put the same line in the function's type; the guide compares the four limits above (`gno-interrealm.md`, "The same line in Swift and Rust").
- A library in the import tree is trusted for interface calls; type-assert caller-supplied interface values.
- A library package variable called in place (`lib.Table[k](x)`) acts for the caller (`zrealm_ctw_owncb`, `zrealm_saved_libtable_copy`).

## Upgrade notes
- These stop writing: a stored callback (store a crossing closure instead); a declared function passed as a value (pass a literal in place, or make it crossing); a method value (call the method, or wrap it in a literal argument); a callback handed to another realm (make it `func(cur realm, …)`, called with `cross`).
- `commondao` v0 `Option` and boards `permissions` v0 `Option` became interfaces (`commondao.WithID(…)` etc. keep their names): a closure made in init runs read-only when `New` calls it.
- Struct errors returned to other realms become pointer errors (`sys/users`); library containers receive realm structs as pointers (`commondao`); `gov/dao NewGovDAO` is crossing.
- Replay: see PR description
