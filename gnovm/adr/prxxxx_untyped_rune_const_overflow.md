# PRxxxx: Reject Untyped Rune Constant Arithmetic That Overflows int32

## Context

Go's untyped constant arithmetic is exact: `'\U0010ffff' * '\U0010ffff'` is
the untyped rune constant 1241243320321, range checked only when it takes a
type.

Gno folded every operation on two untyped rune constants with the
machine's int32 arithmetic (`op_binary.go` groups `UntypedRuneType` with
`Int32Type` through `GetInt32`/`SetInt32`), so the result wrapped, and
`ConvertUntypedRuneTo` widened the wrapped value to the destination:

```go
var w int64 = '\U0010ffff' * '\U0010ffff' // Go: 1241243320321, Gno: -2228223
```

go/types accepts such a program (the value fits int64), so it deployed and
silently computed the wrong value. On the same path:

- `-'a'` and `^'a'` failed with "unexpected type <untyped> int32 in
  operation";
- `var f float64 = 'a'` failed: `ConvertUntypedRuneTo` had no float case;
- `var u uint64 = 'b' - 'c'` wrapped to 18446744073709551615 (Go rejects
  it);
- an untyped integer constant converted to float32 was rounded twice (to
  53 bits, then to 24), so `float32(1<<54 + 1<<30 + 1)` lost the round-up;
- `string(1<<64)` failed, where Go gives "�".

## Decision

### Fold untyped rune operations exactly; reject results beyond int32

`evalConst`, the preprocessor's constant-folding entry point, recognises an
arithmetic or bitwise operation on two untyped rune constants, a shift of
one, and a unary `+`, `-` or `^` on one (`untypedRuneOperands`), and folds
it with `evalUntypedRuneConst`: each rune operand becomes an untyped bigint
of the same value, and the expression is evaluated with the machine's
existing bigint arithmetic (its gas charges and shift cap included).

- A result that fits in int32 is an untyped rune, laid out exactly as the
  int32 folding laid it out (kind and bytes).
- A result beyond int32 is a preprocess error:

  ```
  untyped rune constant '\U0010ffff' * '\U0010ffff' overflows int32 (value 1241243320321); convert an operand to a wider integer type, e.g. int64('\U0010ffff')
  ```

Comparisons of two untyped runes keep the int32 folding: both operands
are within int32, so it is exact.

So an untyped rune never holds a value beyond int32, no untyped integer
is produced from rune-only arithmetic, and no constant changes kind.
Results within int32 are exactly master's; the only other change is that
master's wrapped results become errors. Nothing new reaches runtime, and
non-constant expressions are unchanged.

The typed workaround the message suggests gives gc's exact value:
`int64('\U0010ffff') * '\U0010ffff'` is 1241243320321.

### Conversions

- `ConvertUntypedRuneTo` gains float32/float64 targets (converted like an
  untyped integer constant) and range checks a negative rune converted to
  `uint64`, as it already did for the other unsigned kinds.
- `ConvertUntypedBigintTo` rounds to float once: the `big.Float` gets the
  target precision before `SetInt`.
- `string(x)` of an untyped integer constant outside the int32 range is
  folded to `"�"` first (`foldStringOfLargeIntConst`), so it no longer
  overflows `int` on the way.

## The divergence

Programs that gc accepts and that rely on an untyped rune constant beyond
int32, even as an intermediate, are rejected at preprocess:
`var w int64 = '\U0010ffff' * '\U0010ffff'`, `var x int64 = 'a' << 40`,
`r := '\U0010ffff' * '\U0010ffff' / '\U0010ffff'` (gc: the rune 1114111),
`'\U0010ffff' * '\U0010ffff' * '\x00'` (gc: 0). go/types accepts them, so
such a package passes the keeper's type check and fails at Gno's
preprocess with the message above. Master accepted them and computed with
wrapped values. Programs gc rejects stay rejected
(`var r = '\U0010ffff' * '\U0010ffff'`).

Accepting them with gc's values needs untyped runes that hold values beyond
int32. Earlier revisions of this PR tried two ways (a big-form untyped
rune; folding to an untyped integer); both changed kinds or behaviour
along Gno's non-constant typing paths and produced new wrong values. That
belongs to the follow-up for #6339 and non-constant shift typing.

## Already-deployed packages

A package whose rune constants overflow int32 was accepted by master and
persisted with the wrapped value. After this change, preprocessing that
package fails. On every node start the keeper re-preprocesses all stored
packages (`PreprocessAllFilesAndSaveBlockNodes`), without recovering from
a preprocess error, so **a node upgraded in place would fail to start** if
the chain holds such a package (reproduced with a stored mempackage
containing `const C = '\U0010ffff' * '\U0010ffff'`). A hardfork with
genesis replay (`gno.land/adr/pr5511`) would instead fail that deployment
transaction, and everything that depends on it.

Before this ships to a chain, every stored package must be checked with the
new preprocessor. The genesis packages (stdlibs and examples) are clean:
see Consequences. This is a chain-level decision outside this PR.

Because a package either preprocesses with the bytes master persisted or
fails, a persisted constant can never differ from what the running VM
folds, so imported constants need no special handling.

## Alternatives considered

- **Big-form untyped rune** (V = `BigintValue`, kind rune), the exact Go
  semantics. Every path where such a value travels through Gno's
  non-constant typing (comparisons, switch tags and cases, compound
  assignment, interfaces, kind promotion) needed a guard, and each guard
  produced new wrong results or rejections.
- **Fold results beyond int32 to an untyped integer.** Exact values, but
  the kind changes: `r := big / big; println(r << 32)` printed 4294967296
  where gc and master print 0, a new silently wrong value.
- **Fix the kind order (#6339).** Interacts with non-constant shift
  typing; deferred.

## Consequences

- The silent wrap is gone: an untyped rune result beyond int32 is either a
  compile-time error or, with a typed operand, gc's exact value.
- Unary `-`, `^` on untyped rune constants, float targets, the `uint64`
  range check, single float32 rounding and `string()` beyond int64 now match
  gc.
- Every filetest present on master passes with master's golden output, and
  a byte-for-byte dump of every package-block value of the stdlibs and of
  the example packages the test store loads is identical on master and on
  this branch: genesis and the pinned app hash do not change.
- Consensus: a rune-only fold that fits in int32 is now evaluated with
  bigint arithmetic, whose CPU gas grows with operand bit length. Small
  runes cost the same; `const C = ('\U0010ffff' * 'Ā') + '\x00'` costs
  one more gas unit (1738 against 1737 for the whole deployment, measured
  with a metered `RunMemPackage` on master and on this branch). Float32
  conversions master double-rounded now round once.

## Related, not fixed here

- **#6339 and non-constant shift typing** (one follow-up): `'a' + 1` is an
  untyped int; non-constant shifts are not typed from their context; and
  the divergence above.
- **Typed constant arithmetic** still folds with the runtime (wrapping)
  arithmetic: `const a rune = '\U0010ffff'; const b = a * a` gives
  -2228223 (gc: overflows rune). go/types rejects it at deployment.

## Key files

| File | Role |
|------|------|
| `gnovm/pkg/gnolang/preprocess.go` | `evalConst` hook, `untypedRuneOperands`, `evalUntypedRuneConst`, `constSource`, `foldStringOfLargeIntConst` |
| `gnovm/pkg/gnolang/values_conversions.go` | float targets and the `uint64` check in `ConvertUntypedRuneTo`, single rounding in `ConvertUntypedBigintTo` |
| `gnovm/pkg/gnolang/preprocess_rune_const_test.go` | persisted layout of folded rune constants |
| `gnovm/pkg/gnolang/values_conversions_test.go` | conversion unit tests |
| `gnovm/tests/files/rune4.gno`–`rune16.gno`, `switch52.gno` | filetests |
