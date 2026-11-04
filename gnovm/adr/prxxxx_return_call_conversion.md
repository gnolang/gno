# ADR: Convert every value that reaches its target with only the name differing

## Context

Gno stores a value's type on the value, so a value assigned across types that
differ only by name has to be converted, or it keeps the old type under the new
static one. `x = y`, `return x`, `g(x)` and `T(x)` get the conversion from
preprocess, and `a, b = f()` from `ATTR_ASSIGN_CONVERT`. These checked
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
- Comma-ok: `doOpAssign`'s existing loop.
- Range: each write to the key or value target.

A program with none of these shapes costs the same gas.

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
  another realm's type through these shapes now aborts, as the explicit
  conversion does, including an honest `return vault.Two()` over stored values of
  a type differing by name.
- `var a, b T = f()` is `gno-fixes#342`'s.
- A value of the wrong type stored before this change keeps its type, and
  `var m T = x` and `*p` still set the static type on it.
