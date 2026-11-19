# ADR: Convert every value that reaches its target with only the name differing

## Context

Gno stores a value's type on the value, so a value assigned across types that
differ only by name has to be converted, or it keeps the old type under the new
static one. `x = y`, `return x`, `g(x)` and `T(x)` get the conversion from
preprocess, and `a, b = f()` from `ATTR_ASSIGN_CONVERT`
(`pr6504_multi_assign_named_conversion.md`). These checked
assignability, or only a count, and converted nothing:

- `return f()` with several results
- `g(f())`, in a call and in a `defer`
- `a, ok = m[k]` and `a, ok = x.(T)`
- `for k, v = range x`

Such a value is one a realm can turn into a weapon. `var m T = x` and `*p` set
the static type on the value with no check, since preprocess sees both sides as
`T`. A slice or map another realm stores became a `/p/` type with a mutator, and
a realm wrote the other realm's state through it. Stored directly, a value of an
unnamed struct type under another realm's `T` panics that realm's methods on it.

## Decision

Convert at each of the four sources, through `convertValue`, so the two
inter-realm checks of `T(x)` apply. Preprocess records the target types on the
statement or call, as `ATTR_ASSIGN_CONVERT` already does for `a, b = f()`:

- `return f()`: a new op, `OpReturnConvert`, converts on the operand stack before
  any return op reads the results, so named results and defers see the new type.
- `g(f())`: `doOpPrecall` and `doOpDefer` convert on the operand stack, before a
  frame can move `m.Realm` to a receiver's realm.
- Comma-ok: `doOpAssign`'s existing loop, the `ok` slot included where it goes
  into a declared bool.
- Range: each write to the key or value target.

A program with none of these shapes costs the same gas, except a range
assignment whose key target is not a name, such as `s[i]` or `s[i].N`: it copies
each array or struct element once more before writing the key. A refused `T(x)`
also changes cost, as below: it now runs the deferred functions, and a recover
there can catch it.

A refusal is now a panic through `m.PanicString`, as the readonly refusals of
`append`, `copy` and `delete` are, so a realm's `recover()` catches it. `T(x)`
shares `convertValue`, so it changes with them. A package-level `var` has no call
on the stack, so nothing recovers a refusal there and the transaction aborts.

The split between conversion refusals and the other readonly refusals was never
a design choice. gnolang/gno#4060, the interrealm spec, raised every interrealm
refusal as a Go `panic(string)`: a write into another realm's object, the
crossing rules, and both conversion checks. Eleven days later b12310b7be, in
gnolang/gno#4228, made `m.Panic` raise a VM exception and moved only the write
refusal onto it. The conversion checks were plain `panic` calls, so that change
did not reach them, and no commit, ADR or doc gives a reason to keep them
apart. The spec section added the same week, "`panic()` and `revive(fn)`" in
`docs/resources/gno-interrealm.md`, draws the line at a realm boundary instead:
a panic is recoverable within one realm and aborts the program once it crosses
into another. gnolang/gno#5167 then added the `append`, `copy` and `delete`
refusals on `m.Panic`, following the write refusal. Conversion refusals now
follow the same rule.

## Alternatives

- Checking at `var m T = x` and `*p`. Both set the type for a reason: `*p` must
  answer `Foo` for a pointer converted from `*Bar`, and a check there refuses a
  deref of such a pointer in another realm. Each new way to read a value is one
  more place to check, and a value of the wrong type still breaks a method when
  stored.
- Checking every argument against its parameter at each call. It covers stored
  values too, but it prices a type comparison into every call.
- Refusing the shapes at preprocess. A node re-preprocesses stored packages at
  boot, so a refusal there could stop it.

## Consequences

- Consensus-affecting. A program moving another realm's stored value or building
  another realm's type through these shapes now panics, as the explicit
  conversion does, including an honest `return vault.Two()` over stored values of
  a type differing by name.
- A `/p/` package reading its own package-level data through these shapes panics
  too, since its caller's realm does not own that data:
  `func Multi() (Names, bool) { return raw() }` fails as `Names(table)` does. A
  deployed `/p/` package cannot be changed, so such a function stays unusable.
- A program that recovers a refusal carries on: a range keeps the writes of the
  iterations before it, and a `T(x)` that used to end the transaction no longer
  does. An unrecovered refusal still reverts the whole transaction, after the
  deferred functions run.
- A range assignment reads and converts its values before writing the key, as Go
  evaluates an assignment's operands first, so a refused value writes nothing from
  its iteration. A key target aliasing the ranged element now reads it as Go does:
  `for s[1], v = range s` gives `v` the 0 iteration 0 wrote to `s[1]`, where it
  read its own key's 1 before; `for s[1].N, v = range s` likewise gives `v.N`
  the 0 iteration 0 wrote, whether or not `v` needs a conversion.
- A comma-ok `ok` going into a declared bool has that type, as in Go, and pays one
  conversion.
- `var a, b T = f()` is converted by the typed multi-value declaration change in the same batch.
