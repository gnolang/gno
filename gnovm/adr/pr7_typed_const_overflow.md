# Typed constant overflow during preprocessing

## Context

Issue [#6392](https://github.com/gnolang/gno/issues/6392): a typed integer
constant expression that does not fit its type is a compile error in Go, and
`gno lint` / `gno test` already reject it through `go/types`. `gno run` and
any other path that skips `go/types` do not. The preprocessor folds the
expression by running it on a machine in `StagePre`, and the typed arms of
`+`, `-`, `*`, `/`, and unary `-` are the runtime operators. Those wrap.
`const c int8 = 100; println(c * 2)` therefore prints `-56` instead of
failing with `constant 200 overflows int8`.

Shifts are the one operator that checks the folded value, in
`shlCheckOverflow` / `shrCheckOverflow`, and only against the type's maximum.
`const a int8 = -1; println(a << 8)` is the constant `-256`, which overflows
`int8`, but the upper-bound check does not fire and the machine stores `0`.
The same shape with `rune` (`int32` in Gno's universe) prints `0` for
`-1 << 32`.

A non-constant expression is different on purpose. `var a, b uint64 = 8, 40;
println(a - b)` wraps in Go and must keep wrapping here. The differential
fuzzer's lint-run bucket (draft [gnolang/gno#6539](https://github.com/gnolang/gno/pull/6539),
corpus `misc/gnodiff/corpus/const/unsigned_sub.src`) is the constant form of
that subtraction: `go/types` and `gno lint` reject `uint64(8) - uint64(40)`,
and the VM used to evaluate it.

Untyped integer constants already fold through `big.Int` in the machine.
This change does not touch that path, and it does not change conversion
diagnostics (`constant N overflows IntKind`, `bigint overflows target kind`).

## Decision

Immediately before `evalConst`, if every operand is a constant and the
result type is a typed integer, fold `+`, `-`, `*`, `/`, `<<`, `>>`, and
unary `-` with `math/big` and test the mathematical result against the
type's range.

- When the result does not fit, preprocessing panics with gc's wording:
  `expr (constant N of type T) overflows U` for a basic type, and
  `expr (constant N of U type Name) overflows U` for a defined type. `N` is
  the exact value, not the wrapped one. `U` is the underlying type. The
  shift check runs before the preprocessor rewrites the count to
  `uint(...)`, so the message names the source expression, and the panic is
  on the original node so the span is that expression.
- When the result fits, the machine still folds the expression. For a value
  inside the type's range the two's-complement operators, truncating
  division, and Go's own shift (which saturates at the width) produce the
  same bits. Leaving that path in place keeps preprocess gas on the meter
  `evalConst` already charges. The `big.Int` work uses the Go heap, not the
  VM allocator.
- Shift counts above 1074, the bound `go/types` uses, are not typed
  overflows. They stay on the existing invalid-shift / `constant overflows`
  path. Division by zero stays on `AssertCompatible`.
- `%`, bitwise operators, comparisons, floats, and untyped integers are
  unchanged. Bitwise results and remainders of in-range values stay in
  range, so they cannot hide a typed overflow.

`byte` and `rune` are not distinct types in the VM. The universe binds them
to `uint8` and `int32`, so a VM diagnostic says `uint8` or `int32` where gc
says `byte` or `rune`. A defined type does match gc (`of int8 type MyInt`).
The `go/types` diagnostic in a filetest is unchanged.

## Alternatives

Folding every typed integer constant in `big.Int` and skipping the machine
would make the stored value explicit, and would also drop the preprocess gas
those expressions charge today. A gas change can decide whether a
transaction succeeds. The check only replaces the machine when the result is
not representable, which `go/types` already rejects, so accepted packages
keep their values and their gas.

Putting the range check inside `addAssign` and the other typed arms would
see the wrapped result rather than the expression text, and would apply to
every `StagePre` evaluation, not only constant expressions.

Changing `ConvertTo` to gc's type names would rewrite the existing
conversion filetests without fixing the arithmetic bug. Conversion overflow
already panics; those messages stay as they are.

## Consequences

- `gno run` of a typed integer constant that does not fit is now a
  preprocess error, matching `gno lint` and gc for `int`, `int8`..`int64`,
  and `uint`..`uint64`.
- In-range typed constants, untyped constants, and non-constant integer
  arithmetic are unchanged, including wraparound of runtime subtraction,
  multiplication, and negation of the most negative value.
- `overflow6.gno`..`overflow14.gno` already rejected an out-of-range typed
  shift. They now name the value and the type, and the span is the
  expression rather than the enclosing function.
- A later run of the differential fuzzer should move
  `uint64(8) - uint64(40)` out of the lint-run bucket. `byte` and `rune`
  will still differ from gc by those two names.
