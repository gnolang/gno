# PRxxxx: Exact Untyped Rune Constant Arithmetic

## Context

Go's untyped constant arithmetic is exact. An untyped rune constant has no
int32 bound: `'\U0010ffff' * '\U0010ffff'` is the untyped rune constant
1241243320321, and the range check happens only when the constant takes a
type (an explicit one, or the default type `rune`).

Gno computed every operation on two untyped runes in int32. The machine's
arithmetic (`op_binary.go`, e.g. `mulAssign`) groups `UntypedRuneType` with
`Int32Type` through `GetInt32`/`SetInt32`, so the result wrapped, and
`ConvertUntypedRuneTo` then widened the wrapped value to the destination:

```go
var w int64 = '\U0010ffff' * '\U0010ffff' // Go: 1241243320321, Gno: -2228223
```

go/types accepts such a program (the value fits int64), so it deployed and
computed the wrong value. Related defects found on the same path:

- `'a' + 1`, `1 + 'a'`, `'a' << 1` were typed as untyped int, not untyped
  rune, so `[]rune{'a' + 1}` was rejected (#6339).
- `'a' << 40` failed with "constant overflows"; `-'a'` and `^'a'` failed with
  "unexpected type <untyped> int32 in operation".
- `var f float64 = 'a'` failed: `ConvertUntypedRuneTo` had no float case.
- `var u uint64 = -'a'` wrapped instead of failing (go/types rejects it).

## Decision

### Representation: canonical int32-or-bigint

An `UntypedRuneType` value is held in `N` as an int32 when it fits, exactly
as a rune literal is, and in `V` as a `BigintValue` otherwise
(`TypedValue.GetUntypedRune`/`SetUntypedRune`). The form is canonical, so
every in-range value, and every value already persisted (package-level
constants live in package blocks), keeps its existing layout. Only values
that used to wrap take the new form.

### Folding: evaluate as untyped bigint, keep the rune kind

`evalConst`, the preprocessor's single constant-folding entry point,
recognises a binary operation on two untyped rune constants, a shift of
one, or a unary operation on one (`untypedRuneOperands`) and hands it to
`evalUntypedRuneConst`: each rune operand is replaced by an untyped bigint
of the same value, the expression is evaluated by the machine's existing
bigint arithmetic (with its gas charges and shift cap), and an integer
result becomes an untyped rune again. Comparisons return their untyped bool
unchanged. Doing this inside `evalConst` covers every folding caller, not
only the binary, shift and unary cases of `preprocess`.

The machine's int32 path for `UntypedRuneType` is left in place but no
longer reached by constant folding of rune operands; in debug builds
`GetInt32` panics on a big-form untyped rune so a missed path fails loudly.

### Kind ordering: integer < rune

`PrimitiveType.Specificity` now orders untyped bigint after untyped rune,
matching the spec ("integer, rune, floating-point, complex"). In
`'a' + 1` the untyped int is converted to an untyped rune;
`ConvertUntypedBigintTo` keeps the exact value for that target
(`'a' + 1<<40` is a valid untyped rune). Untyped string keeps its place
before rune, so the only changed diagnostic is `"a" + 1`, which now reads
"cannot use untyped Bigint as StringKind" (as `1 + "a"` already did).

### Conversion: one range check, against the exact value

`ConvertUntypedRuneTo` and `ConvertUntypedBigintTo` share
`convertUntypedIntegerTo`, which range checks the exact value for every
integer target (int32 included), converts to float32/float64, bigint and
bigdec, and clears `N` before writing a narrower result. The "simple
conversion" shortcut in `ConvertUntypedTo` (same kind, retag only) is
skipped for a big-form untyped rune. Messages keep the existing
"rune overflows target kind" / "rune underflows target kind" wording.

Two conversion-call cases in the preprocessor follow: `float64(r)` of an
untyped rune converts the exact value directly, and `string(r)` of a rune
beyond int32 goes through the untyped-int path, yielding "�".

## Alternatives considered

- **Always hold untyped runes as bigint.** Changes the layout of every rune
  literal and of constants already persisted in package blocks; rejected.
- **Make the machine's `UntypedRuneType` arithmetic exact op by op.** Touches
  eleven binary operators, five comparisons, two unary operators and the
  bigint gas helpers, for values that only ever arise while folding;
  folding through the existing bigint path reuses all of that.
- **Fold to untyped bigint and drop the rune kind.** Then
  `var r = '\U0010ffff' * '\U0010ffff'` would silently default to `int`
  instead of failing, and `const c = ...; r := c / '\U0010ffff'` would not
  be a rune.

## Consequences

- Untyped rune constant expressions match Go's values and types.
- Out-of-range defaults (`var r = big`, `println(big)`, `rune(big)`) fail
  at preprocess, as gc does.
- Typed rune arithmetic is unchanged and still wraps at runtime.
- #6339 is resolved; the gocorpus `rune.go` entry carrying
  `KnownIssue compile: #6339` (draft #6340) would flip to `Clean compile`.

## Key files

| File | Role |
|------|------|
| `gnovm/pkg/gnolang/preprocess.go` | `evalConst` hook, `untypedRuneOperands`, `evalUntypedRuneConst`, `float64`/`string` conversion calls |
| `gnovm/pkg/gnolang/values.go` | `GetUntypedRune`, `SetUntypedRune`, `Sign` |
| `gnovm/pkg/gnolang/values_conversions.go` | `ConvertUntypedRuneTo`, `convertUntypedIntegerTo` |
| `gnovm/pkg/gnolang/types.go` | `Specificity` ordering |
| `gnovm/tests/files/rune4.gno`–`rune16.gno` | filetests |
