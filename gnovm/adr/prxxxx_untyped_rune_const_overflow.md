# PRxxxx: Reject Untyped Rune Constant Arithmetic That Overflows int32

## Context

Go's untyped constant arithmetic is exact: `'\U0010ffff' * '\U0010ffff'` is
the untyped rune constant 1241243320321, range checked only when it takes a
type.

Gno folded every binary operation on two untyped rune constants with the
machine's int32 arithmetic (`op_binary.go` groups `UntypedRuneType` with
`Int32Type` through `GetInt32`/`SetInt32`), so the result wrapped, and
`ConvertUntypedRuneTo` widened the wrapped value to the destination:

```go
var w int64 = '\U0010ffff' * '\U0010ffff' // Go: 1241243320321, Gno: -2228223
```

go/types accepts such a program (the value fits int64), so it deployed and
silently computed the wrong value. On the same path:

- `var f float64 = 'a'` failed: `ConvertUntypedRuneTo` had no float case;
- `var u uint64 = 'b' - 'c'` wrapped to 18446744073709551615 (Go rejects
  it);
- `string(1<<64)` failed, where Go gives "�".

## Scope

This change makes Gno an exact superset of master for constants: it never
gives a value master did not give, except to replace a wrapped value with
an error. Concretely:

- a binary operation on untyped rune constants whose exact result fits in
  int32 folds to exactly master's result (kind and bytes);
- one whose exact result does not fit is rejected, where master wrapped it;
- programs master rejected stay rejected, except the three conversions
  below, each of which master rejected outright, never folded to a value.

Unary operators on untyped rune constants (`-'a'`, `^'a'`) keep master's
behaviour: the machine does not define them on untyped runes, so they are
rejected ("unexpected type <untyped> int32 in operation"). Enabling them
would let their results into Gno's non-constant shift typing, which is not
Go's (`1 + ((-'a') << s)` would print 4294967200, gc -96).

## Decision

### Fold binary untyped rune operations exactly; reject results beyond int32

`evalConst`, the preprocessor's constant-folding entry point, recognises an
arithmetic or bitwise operation on two untyped rune constants and a shift
of one (`untypedRuneOperands`), and folds it with `evalUntypedRuneConst`:
each rune operand becomes an untyped bigint of the same value, and the
expression is evaluated with the machine's existing bigint arithmetic (its
gas charges and shift cap included).

- A result that fits in int32 is an untyped rune, laid out exactly as the
  int32 folding laid it out (kind and bytes).
- A result beyond int32 is a preprocess error that names the expression as
  written (`constSource`: constants, imported selectors, `iota`) and
  suggests a typed operand:

  ```
  untyped rune constant '\U0010ffff' * '\U0010ffff' overflows int32 (value 1241243320321); convert an operand to a wider integer type, e.g. int64('\U0010ffff')
  ```

- `0 << n` folds to 0 for any count, as the int32 folding did, without the
  bigint shift (whose 10000-bit cap would reject `'\x00' << 10001`).

Comparisons of two untyped runes keep the int32 folding: both operands are
within int32, so it is exact.

So an untyped rune never holds a value beyond int32, rune-only arithmetic
never produces an untyped integer, and no constant changes kind. Nothing
new reaches runtime, and non-constant expressions are unchanged.

The typed workaround the message suggests gives gc's exact value:
`int64('\U0010ffff') * '\U0010ffff'` is 1241243320321.

### Conversions

Each of these was rejected on master, so none changes a value master
computed:

- `ConvertUntypedRuneTo` gains float32/float64 targets
  (`var f float64 = 'a'` failed with "cannot convert untyped rune type to
  Float64Kind"; explicit `float64('a')` already worked, by another path);
- it range checks a negative rune converted to `uint64`, as it already did
  for the other unsigned kinds (master wrapped; Go rejects);
- `string(x)` of an untyped integer constant beyond the int64 range is
  folded to `"�"` first (`foldStringOfLargeIntConst`); master rejected
  it ("bigint overflows target kind"). Values within int64 keep their path.

The `uint64` check is the one that turns an accepted program into an
error; it belongs with the overflow rejections below.

## The divergence

Programs that gc accepts and that rely on an untyped rune constant beyond
int32, even as an intermediate, are rejected at preprocess. go/types
accepts them, so such a package passes the keeper's type check and fails
at Gno's preprocess with the message above.

- Master accepted the arithmetic ones and computed with wrapped values:
  `var w int64 = '\U0010ffff' * '\U0010ffff'` (-2228223),
  `r := '\U0010ffff' * '\U0010ffff' / '\U0010ffff'` (gc: the rune 1114111;
  master -2), `'\U0010ffff' * '\U0010ffff' * '\x00'` (0, by accident).
- Master already rejected overflowing shifts, with "constant overflows":
  `var x int64 = 'a' << 40`. Only the message changes.

Programs gc rejects stay rejected (`var r = '\U0010ffff' * '\U0010ffff'`).

Accepting them with gc's values needs untyped runes that hold values beyond
int32. Earlier revisions of this PR tried two ways (a big-form untyped
rune; folding to an untyped integer); both changed kinds or behaviour
along Gno's non-constant typing paths and produced new wrong values. That
belongs to the follow-up for #6339 and non-constant shift typing.

## Already-deployed packages

A package whose rune constants overflow int32 (or that converts a negative
rune constant to `uint64`) was accepted by master and persisted with the
wrapped value. After this change, preprocessing that package fails. On
every node start the keeper re-preprocesses all stored packages
(`PreprocessAllFilesAndSaveBlockNodes`), without recovering from a
preprocess error, so **a node upgraded in place would fail to start** if
the chain holds such a package (reproduced with a stored mempackage
containing `const C = '\U0010ffff' * '\U0010ffff'`). A hardfork with
genesis replay (`gno.land/adr/pr5511`) would instead fail that deployment
transaction, and everything that depends on it.

Before this ships to a chain, every stored package must be checked with the
new preprocessor. The genesis packages (stdlibs and examples) are clean:
see Consequences. This is a chain-level decision outside this PR.

Otherwise, a stored package preprocesses to exactly the constant values
and bytes master persisted: every fold that master completed, this change
completes identically (or rejects), and every conversion that changed was
rejected on master. So a persisted constant cannot differ from what the
running VM folds, and imported constants, which importers read from the
persisted package block, need no special handling.

## Alternatives considered

- **Big-form untyped rune** (V = `BigintValue`, kind rune), the exact Go
  semantics. Every path where such a value travels through Gno's
  non-constant typing (comparisons, switch tags and cases, compound
  assignment, interfaces, kind promotion) needed a guard, and each guard
  produced new wrong results or rejections.
- **Fold results beyond int32 to an untyped integer.** Exact values, but
  the kind changes: `r := big / big; println(r << 32)` printed 4294967296
  where gc and master print 0, a new silently wrong value.
- **Also fold unary operators.** Turns master's rejection of `-'a'` into a
  value that then meets non-constant shift typing (see Scope).
- **Round untyped integers to float32 once.** A real defect (master rounds
  to 53 bits, then to 24), but it changes values master computed and
  persisted, so importers would read stale constants after an upgrade.
  Split out to its own issue.
- **Fix the kind order (#6339).** Interacts with non-constant shift typing;
  deferred.

## Consequences

- The silent wrap is gone: a binary untyped rune result beyond int32 is
  either a compile-time error or, with a typed operand, gc's exact value.
- Float targets for untyped runes, the `uint64` range check and `string()`
  beyond int64 now match gc.
- Every filetest present on master passes with master's golden output, and
  a byte-for-byte dump of every package-block value of the stdlibs and of
  the example packages the test store loads is identical on master and on
  this branch: genesis and the pinned app hash do not change.
- Gas: a rune-only fold that fits in int32 is now evaluated with bigint
  arithmetic, whose CPU gas grows with operand bit length. Small runes cost
  the same; `const C = ('\U0010ffff' * 'Ā') + '\x00'` costs one more
  gas unit (1738 against 1737 for the whole deployment, measured with a
  metered `RunMemPackage` on master and on this branch). Imported
  constants cost the same as on master.

## Related, not fixed here

- **#6339 and non-constant shift typing** (one follow-up): `'a' + 1` is an
  untyped int; non-constant shifts are not typed from their context;
  unary operators on untyped rune constants; and the divergence above.
- **Double rounding to float32** of untyped integer constants, with its
  upgrade hazard (own issue).
- **Typed constant arithmetic** still folds with the runtime (wrapping)
  arithmetic: `const a rune = '\U0010ffff'; const b = a * a` gives
  -2228223 (gc: overflows rune). go/types rejects it at deployment.

## Key files

| File | Role |
|------|------|
| `gnovm/pkg/gnolang/preprocess.go` | `evalConst` hook, `untypedRuneOperands`, `evalUntypedRuneConst`, `constSource`, `foldStringOfLargeIntConst` |
| `gnovm/pkg/gnolang/values_conversions.go` | float targets and the `uint64` check in `ConvertUntypedRuneTo` |
| `gnovm/pkg/gnolang/preprocess_rune_const_test.go` | persisted layout of folded rune constants |
| `gnovm/pkg/gnolang/values_conversions_test.go` | conversion unit test |
| `gnovm/tests/files/rune4.gno`–`rune22.gno`, `switch52.gno` | filetests |
