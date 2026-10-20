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

- `'a' + 1` and `1 + 'a'` were untyped int, not untyped rune, so
  `r := 'a' + 1` defaulted to `int` and `[]rune{r}` was rejected (#6339).
- `'a' << 40` failed with "constant overflows"; `-'a'` and `^'a'` failed with
  "unexpected type <untyped> int32 in operation".
- `var f float64 = 'a'` failed: `ConvertUntypedRuneTo` had no float case.
- `var u uint64 = 'b' - 'c'` wrapped to 18446744073709551615 instead of
  failing (go/types rejects it).

Found while fixing it, on the shared untyped-integer conversion path:

- An integer constant converted to float32 was rounded twice (first to 53
  bits, then to 24), so `float32(1<<54 + 1<<30 + 1)` lost the round-up.
- `string(x)` of an integer constant beyond int64 (`string(1<<64)`) failed,
  where Go gives "\uFFFD".

## Decision

### Representation: canonical int32-or-bigint

An `UntypedRuneType` value is held in `N` as an int32 when it fits, exactly
as a rune literal is, and in `V` as a `BigintValue` otherwise
(`TypedValue.GetUntypedRune`/`SetUntypedRune`). The form is canonical: an
untyped rune value that fits in int32 has the same bytes as before (rune
literals, in-range results, and runes already persisted in package blocks
decode unchanged). Only values that used to wrap take the new form.

This holds for untyped rune values themselves. Some constants derived from
them get different `N` bytes with the same value and type; see
Consequences.

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

Integer to float conversion rounds once, from the exact value: the
`big.Float` gets the target precision before `SetInt`.

Two conversion-call cases in the preprocessor follow. `float64(r)` of an
untyped rune converts the exact value directly. `string(x)` of an untyped
integer or rune constant outside the int32 range is folded to the rune
'\uFFFD' first (`foldStringOfLargeIntConst`), since no value there is a code
point; this holds for any size and for named string types. Other conversion
targets keep normal defaulting, so `interface{}('a' << 40)` still fails, as
the constant cannot default to `rune`.

### Non-constant shifts: one type for the whole expression

A non-constant shift of an untyped constant (`'a' << s`) stays untyped until
context types it. When such an expression meets another untyped operand
with no context (`x := 1 + ('a' << s)`, `('a' << s) < 100`), the
preprocessor used to convert one operand to its default type and the other
to the *untyped* type of the first, or to nothing at all for `<`, `<=`, `>`
and `>=`; mixed untyped kinds on two non-constant operands
(`('a' << s) + (1 << s)`) were rejected. With rune now ranking above int,
that left an untyped rune constant, possibly in big form, in runtime
arithmetic (`('a' << s) + (1 << 40)` printed 97).

Both operands now take one type, `untypedOperandsDefault`: the default type
of the later kind of the two. This applies to an untyped binary expression
without context, to every comparison of two untyped operands (it was only
`==` and `!=`), and to two untyped non-constant operands. The expression's
cached type is updated to match. A constant operand is converted, and so
range checked, before runtime: `('a' << s) + (1 << 40)` is rejected as in
Go. A float shift such as `1<<s + 1.0` now fails as a `float64` shift, as
in Go, instead of a bigdec one.

To check that no big-form untyped rune reaches runtime, the `GetInt32` guard
was made unconditional and the gnolang tests, `gnovm/pkg/test`,
`gnovm/cmd/gno` and the examples suite were run with it: no hit. The guard
is debug-only in the committed code.

### Imported constants: read the recomputed value

A package constant is folded at deploy and its value is copied into the
package block, which is persisted. On every node start the keeper
re-preprocesses all packages (`PreprocessAllFilesAndSaveBlockNodes`), so a
package's static block, from which the package's own code folds its
constants, always reflects the running VM. Importers, however, folded
`pkg.C` by evaluating the selector against the persisted package block. A
package deployed before this fix would keep exporting the wrapped value to
its importers (`int64(p.C)` = -2228223) while its own code saw the exact one.

The preprocessor now resolves an imported constant from the imported
package node's static block (`importedConst`), falling back to the old
evaluation when that slot is not set. Importers and the package itself thus
agree, on fresh deploys and after an upgrade alike. The persisted slot is
left as is: nothing folds from it any more, but state exports still show it.

How such a change reaches a live chain is a chain-level decision outside
this PR. The VM change alters results, so it ships with a coordinated
upgrade. A hardfork with genesis replay (`gno.land/adr/pr5511`) re-executes
every deployment and recomputes all persisted slots. An in-place binary
upgrade keeps them, and then `importedConst` is what keeps importers
correct. Earlier constant-folding fixes (e.g. #5741, signed zero) changed
folded values without any migration of persisted slots.

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
- Importers see the constant folding of the running VM, not that of the VM
  that deployed the imported package.
- Consensus-breaking. Package blocks are amino-encoded with all eight
  bytes of `N`, so a deployment (or a replay) of a package holding any of
  these constants stores different bytes:
  - a result that used to wrap now has its exact value;
  - a mixed rune/int constant (`'a' + 1`) is an untyped rune, not an
    untyped bigint (`N` holds the int32, `V` is nil);
  - a comparison of two untyped runes whose left operand does not fit in
    one byte or is negative (`'\u0100' == '\u0100'`) has `N` zeroed past
    the boolean; it used to keep the operand's upper bytes
    (`01 01 00 …` → `01 00 00 …`);
  - a negative untyped rune narrowed to `int8` or `int16`
    (`const C int8 = 'a' - 'b'`) has `N` zeroed past the value; it used to
    keep the int32 sign extension (`ff ff ff ff 00 …` → `ff 00 …`).
  The values and types of the last two are unchanged.
  `TestUntypedRuneConstLayout` pins these bytes.
- In the genesis set, a byte-for-byte dump of every package-block value of
  the stdlibs (and of the example packages the test store loads) on master
  and on this branch differs in one slot only: `unicode.UpperLower`
  (`MaxRune + 1`), now an untyped rune as in Go. It moves the pinned app
  hash in `gno.land/pkg/sdk/vm/apphash_crossrealm38_test.go`.
- #6339 is resolved; the gocorpus `rune.go` entry carrying
  `KnownIssue compile: #6339` (draft #6340) would flip to `Clean compile`.

## Related, not fixed here

Typed constant arithmetic still folds with the runtime (wrapping)
arithmetic, for every typed integer kind:

```go
const a rune = '\U0010ffff'
const b = a * a // gc: overflows rune; Gno: -2228223
```

go/types rejects it at deployment, so it is a VM divergence, not a
deployable bug. Making typed folding exact with a range check is a
separate change.

## Key files

| File | Role |
|------|------|
| `gnovm/pkg/gnolang/preprocess.go` | `evalConst` hook, `untypedRuneOperands`, `evalUntypedRuneConst`, `untypedOperandsDefault`, `importedConst`, `foldStringOfLargeIntConst`, `float64`/`string` conversion calls |
| `gnovm/pkg/gnolang/values.go` | `GetUntypedRune`, `SetUntypedRune`, `Sign` |
| `gnovm/pkg/gnolang/values_conversions.go` | `ConvertUntypedRuneTo`, `convertUntypedIntegerTo` |
| `gnovm/pkg/gnolang/types.go` | `Specificity` ordering |
| `gnovm/pkg/gnolang/preprocess_import_const_test.go` | persisted-layout test and reload test for imported constants |
| `gnovm/tests/files/rune4.gno`–`rune23.gno` | filetests |
| `gnovm/tests/files/types/shift_f2d.gno`, `shift_f3.gno`, `shift_f3b.gno`, `shift_f4c.gno` | float shift diagnostics now name `float64` |
