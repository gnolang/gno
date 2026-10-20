# PRxxxx: Exact Untyped Rune Constant Arithmetic

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
computed the wrong value. On the same path:

- `'a' << 40` failed with "constant overflows"; `-'a'` and `^'a'` failed
  with "unexpected type <untyped> int32 in operation";
- `var f float64 = 'a'` failed: `ConvertUntypedRuneTo` had no float case;
- `var u uint64 = 'b' - 'c'` wrapped to 18446744073709551615 (Go rejects
  it);
- an untyped integer constant converted to float32 was rounded twice (to
  53 bits, then to 24), so `float32(1<<54 + 1<<30 + 1)` lost the round-up;
- `string(1<<64)` failed, where Go gives "�".

## Decision

### Fold untyped rune operations exactly, with no new representation

`evalConst`, the preprocessor's constant-folding entry point, recognises an
arithmetic or bitwise operation on two untyped rune constants, a shift of
one, and a unary `+`, `-` or `^` on one (`untypedRuneOperands`), and folds
it with `evalUntypedRuneConst`: each rune operand becomes an untyped bigint
of the same value, and the expression is evaluated with the machine's
existing bigint arithmetic (its gas charges and shift cap included).

- A result that fits in int32 becomes an untyped rune again, laid out
  exactly as the int32 folding laid it out.
- A result beyond int32 stays an **untyped bigint**, Gno's untyped integer
  kind, with its exact value.

Comparisons of two untyped runes are left to the int32 folding: both
operands are within int32, so it is exact, and its bytes are unchanged.

An untyped rune therefore still never holds a value beyond int32. Large
results take the existing untyped-integer paths everywhere (conversions,
defaulting, non-constant expressions), so there is no new runtime
representation and no new non-constant behaviour.

### Conversions

- `ConvertUntypedRuneTo` gains float32/float64 targets (converted like an
  untyped integer constant) and range checks a negative rune converted to
  `uint64`, as it already did for the other unsigned kinds.
- `ConvertUntypedBigintTo` rounds to float once: the `big.Float` gets the
  target precision before `SetInt`.
- `string(x)` of an untyped integer constant outside the int32 range is
  folded to `"�"` first (`foldStringOfLargeIntConst`), since no value
  there is a code point, so it no longer overflows `int` on the way.

### Imported constants: read the recomputed value

A package constant is folded at deploy and copied into the package block,
which is persisted. On every node start the keeper re-preprocesses all
packages (`PreprocessAllFilesAndSaveBlockNodes`), so a package's static
block, from which its own code folds its constants, reflects the running
VM. Importers, however, folded `pkg.C` from the persisted package block. A
package deployed before this fix would keep exporting the wrapped rune to
its importers (`int64(p.C)` = -2228223) while its own code saw the exact
untyped integer.

The preprocessor now resolves an imported declared constant (`GetIsConstAt`)
from the imported package node's static block (`importedConst`), falling
back to the old evaluation when that slot is not set. A package variable is
never read from its static slot, which holds only a placeholder for its
type. The persisted slot is left as is: nothing folds from it any more, but
state exports still show it.

How such a change reaches a live chain is a chain-level decision outside
this PR. A hardfork with genesis replay (`gno.land/adr/pr5511`) re-executes
every deployment and recomputes all persisted slots; an in-place binary
upgrade keeps them, and then `importedConst` keeps importers correct.
Earlier constant-folding fixes (e.g. #5741, signed zero) changed folded
values without migrating persisted slots.

## The divergence

Go keeps a result beyond int32 an untyped *rune*; here it is an untyped
*integer*. The value is exact either way; the kind differs, and with it
whatever depends on the kind:

- defaulting: `var r = '\U0010ffff' * '\U0010ffff'` and
  `println('\U0010ffff' * '\U0010ffff')` default to `int` and print
  1241243320321, where gc rejects them (rune overflow);
  `interface{}('a' << 40)` holds an `int`;
- results derived from such a value, even back within int32, are untyped
  integers: `r := big / '\U0010ffff'` is an `int`, so `[]rune{r}` is
  rejected (gc accepts it);
- in a non-constant expression such a constant acts as an untyped integer
  constant of the same value does, with Gno's existing non-constant
  typing: `(1 << s) == big` compares as `int` (gc rejects it), and
  `(big + ('a' << s)) < 'a'` fails at runtime, as the same expression
  written with an integer literal does on master.

go/types rejects the defaulting cases on chain, so deployments are not
affected by those; the rest belongs to the follow-up for #6339 (the kind
of mixed rune/int constants) together with non-constant shift typing. On
master every one of these programs computed with a wrapped value.

## Alternatives considered

- **Hold large untyped runes in a new big form** (V = `BigintValue`, kind
  rune). Tried in earlier revisions of this PR: Gno's non-constant typing
  is not Go's, so every path where such a value can travel (comparisons,
  switch tags and cases, compound assignment, interface contexts, kind
  promotion) needed a guard, and each guard produced new wrong results or
  rejections of programs master handles. Rejected for this PR.
- **Fix the kind order (#6339)** so `'a' + 1` is an untyped rune. Tried;
  with Gno's non-constant shift typing it produced wrong values
  (`int64((('a'+1)<<s) + (('b'+1)<<s))` gave 0) and new rejections.
  Deferred.
- **Make the machine's `UntypedRuneType` arithmetic exact op by op.**
  Touches every operator for values that only arise while folding; the
  existing bigint path already does it.

## Consequences

- Untyped rune constant arithmetic is exact; conversions of its results are
  exact and range checked; the constant shift and unary cases that failed
  now fold.
- A result within int32 is unchanged, kind and bytes included. Every
  filetest present on master passes with master's golden output, and a
  byte-for-byte dump of every package-block value of the stdlibs and of
  the example packages the test store loads is identical on master and on
  this branch: genesis and the pinned app hash do not change.
- Consensus-breaking for affected code only: a constant whose rune
  arithmetic leaves int32 is now an untyped bigint with its exact value
  (persisted as such), where master stored the wrapped rune; float32
  conversions of untyped integer constants that master double-rounded now
  round once.
- Non-constant expressions keep master's behaviour for every program whose
  constants stay within int32 (`rune16.gno` pins a collection of them).

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
| `gnovm/pkg/gnolang/preprocess.go` | `evalConst` hook, `untypedRuneOperands`, `evalUntypedRuneConst`, `importedConst`, `foldStringOfLargeIntConst` |
| `gnovm/pkg/gnolang/values_conversions.go` | float targets and the `uint64` check in `ConvertUntypedRuneTo`, single rounding in `ConvertUntypedBigintTo` |
| `gnovm/pkg/gnolang/preprocess_import_const_test.go` | layout test; reload tests for imported constants and variables |
| `gnovm/pkg/gnolang/values_conversions_test.go` | conversion unit tests |
| `gnovm/tests/files/rune4.gno`–`rune23.gno`, `switch52.gno` | filetests |
