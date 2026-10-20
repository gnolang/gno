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

- `'a' << 40` failed with "constant overflows"; `-'a'` and `^'a'` failed with
  "unexpected type <untyped> int32 in operation".
- `var f float64 = 'a'` failed: `ConvertUntypedRuneTo` had no float case.
- `var u uint64 = 'b' - 'c'` wrapped to 18446744073709551615 instead of
  failing (go/types rejects it).

Found while fixing it, on the shared untyped-integer conversion path:

- An integer constant converted to float32 was rounded twice (first to 53
  bits, then to 24), so `float32(1<<54 + 1<<30 + 1)` lost the round-up.
- `string(x)` of an integer constant beyond int64 (`string(1<<64)`) failed,
  where Go gives "�".

## Scope

This change makes **untyped rune constant arithmetic exact**: an operation
whose operands are untyped rune constants (`'\U0010ffff' * '\U0010ffff'`,
`-'a'`, `'a' << 40`, `'a' < 'b'`) is folded without wrapping, and every
conversion of an untyped integer or rune constant range checks the exact
value.

It does **not** change the kind of a mixed operation: an untyped rune and an
untyped int still give an untyped int, as on master. Go gives an untyped
rune (#6339). Fixing that interacts with Gno's non-conformant typing of
non-constant shifts (`'a' << s`): every attempt to change the kind order
produced new wrong results or rejections for non-constant shift expressions
that master handles. Both are left to a follow-up (see "Related, not fixed
here").

For non-constant expressions the rule is: no program that master accepted
and gc also accepts is newly rejected, and none produces a value different
from master's, unless master's value was wrong and the new value is gc's.
The exception is an untyped rune constant beyond int32 next to a
non-constant operand, which is always rejected (see "No big-form untyped
rune at runtime"); in a wider explicit context gc accepts some of these.

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
unchanged. Doing this inside `evalConst` covers every folding caller.

The machine's int32 path for `UntypedRuneType` is left in place; it is no
longer reached by constant folding of rune operands.

### Conversion: one range check, against the exact value

`ConvertUntypedRuneTo` and `ConvertUntypedBigintTo` share
`convertUntypedIntegerTo`, which range checks the exact value for every
integer target (int32 included), converts to float32/float64, bigint and
bigdec, and clears `N` before writing a narrower result. The "simple
conversion" shortcut in `ConvertUntypedTo` (same kind, retag only) is
skipped for a big-form untyped rune. Messages keep the existing
"rune overflows target kind" / "rune underflows target kind" wording. An
untyped bigint converted to an untyped rune is range checked against int32,
as before: only folding produces big-form runes.

Integer to float conversion rounds once, from the exact value: the
`big.Float` gets the target precision before `SetInt`.

Two conversion-call cases in the preprocessor follow. `float64(r)` of an
untyped rune converts the exact value directly. `string(x)` of an untyped
integer or rune constant outside the int32 range is folded to the rune
'�' first (`foldStringOfLargeIntConst`), since no value there is a code
point; this holds for any size and for named string types. Other conversion
targets keep normal defaulting, so `interface{}('a' << 40)` fails, as the
constant cannot default to `rune`.

### No big-form untyped rune at runtime

A big-form untyped rune is exact only while it stays a constant. Next to a
non-constant operand it takes the type of the expression, which Gno does
not always resolve as Go does (non-constant shifts are typed eagerly, and
a mixed comparison defaults to int), and runtime arithmetic reads untyped
runes as int32. Checks at particular call sites kept missing compound
paths (a nested operand, a compound switch tag, a mixed equality that
widens the constant before the check), so there is one rule, applied before
any defaulting, independent of context:

> An untyped rune constant beyond int32 that meets a non-constant operand
> is rejected with "rune overflows target kind".

A constant that is folded from such a rune together with an untyped
integer (`big + 0`, `0 + big`, `big * 1`, and a named constant defined so)
keeps the rune kind, as in Go (`keepBigUntypedRuneKind`): Gno's kind order
would make it an untyped bigint, which the rule would not see. Master
wrapped such a rune, so no result of such an operation was right before.
Runes within int32 keep Gno's kind order (#6339), and in a purely constant
context the value stays exact (`var x int64 = big + 1`).

`rejectBigUntypedRune` applies the rule at the two places where a constant
meets a non-constant operand:

- the start of the `BinaryExpr` `TRANS_LEAVE`, when exactly one operand is
  a `*ConstExpr` (every arithmetic, comparison and shift expression, at any
  nesting depth, passes there before its operands are converted);
- the switch case loop, when the tag is not constant.

Unary expressions cannot combine a constant with a non-constant operand,
and an assignment, argument or interface conversion of a lone constant is
an ordinary constant conversion, range checked as above. Purely constant
expressions are unaffected.

The cost: gc accepts some of these in a wider explicit context and they
are rejected here, e.g. `var y int64 = ('a' << s) + big`,
`var y int64 = big << s`, `var x int64; _ = x + big`. Master accepted them
and computed wrapped values.

Switch tags keep master's typing. Defaulting a non-constant untyped tag,
tried in an earlier revision, converted one operand of a compound tag to
`int` and left the other an untyped bigint, so `switch 1 + (1 << s)`
panicked at runtime; updating the cached type of unary or arithmetic
expressions changed results of programs master accepted. The case-level
check above is enough to keep big runes out. At runtime, unary `-` and
`^` now accept an untyped rune operand (a non-constant shift such as
`-('a' << s)` used as a switch tag), with the int32 semantics the binary
operators already give it; master rejected such tags.

### Guard: debug-only

`GetInt32` panics, in debug builds, on an untyped rune held in its big form.
Made unconditional in production, the check (one `V != nil` test, with the
panic kept out of line so `GetInt32` stays inlinable) costs about 1%: an
isolated int32 `doOpAdd` went from 21.61 ns to 21.80 ns (+0.9%, p=0.003,
n=10), and a 3,000,000-iteration int32 loop under `gno run` from 0.82 s to
0.83-0.84 s. Gas is not affected. Because the cost is measurable on every
int32 operation, the guard stays debug-only. Instead, the guard was made
unconditional temporarily and the gnolang tests (filetests included),
`gnovm/pkg/test`, `gnovm/cmd/gno` and the examples suite were run with it:
it never fired.

### Imported constants: read the recomputed value

A package constant is folded at deploy and its value is copied into the
package block, which is persisted. On every node start the keeper
re-preprocesses all packages (`PreprocessAllFilesAndSaveBlockNodes`), so a
package's static block, from which the package's own code folds its
constants, always reflects the running VM. Importers, however, folded
`pkg.C` by evaluating the selector against the persisted package block. A
package deployed before this fix would keep exporting the wrapped value to
its importers (`int64(p.C)` = -2228223) while its own code saw the exact one.

The preprocessor now resolves an imported **constant** (per the package
node's constant metadata, `GetIsConstAt`) from the imported package node's
static block (`importedConst`), falling back to the old evaluation when that
slot is not set. A package variable is never read from its static slot,
which holds only a placeholder for its type; it keeps the old evaluation
even when its static type is untyped. The persisted slot of a constant is
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
- **Also fix the kind order (#6339).** Tried; it changes the type of every
  mixed rune/int constant, and with Gno's non-constant shift typing that
  produced wrong values (`int64((('a'+1)<<s) + (('b'+1)<<s))` gave 0) and
  new rejections of programs master accepts. Deferred to the non-constant
  shift follow-up.

## Consequences

- Untyped rune constant expressions match Go's values; out-of-range
  defaults (`var r = big`, `println(big)`, `rune(big)`) fail at preprocess,
  as gc does.
- Typed rune arithmetic is unchanged and still wraps at runtime.
- New rejections of programs master accepted: every untyped rune constant
  beyond int32 next to a non-constant operand (master wrapped it; gc
  accepts it only in a wider explicit context).
- Importers see the constant folding of the running VM, not that of the VM
  that deployed the imported package.
- Every filetest present on master passes with master's golden output, and
  a byte-for-byte dump of every package-block value of the stdlibs and of
  the example packages the test store loads is identical on master and on
  this branch, so the pinned app hash is unchanged.
- Consensus-breaking for affected code. Package blocks are amino-encoded
  with all eight bytes of `N`, so a deployment (or a replay) of a package
  holding any of these constants stores different bytes:
  - a result that used to wrap now has its exact value;
  - a comparison of two untyped runes whose left operand does not fit in
    one byte or is negative (`'Ā' == 'Ā'`) has `N` zeroed past
    the boolean; it used to keep the operand's upper bytes
    (`01 01 00 …` → `01 00 00 …`);
  - a negative untyped rune narrowed to `int8` or `int16`
    (`const C int8 = 'a' - 'b'`) has `N` zeroed past the value; it used to
    keep the int32 sign extension (`ff ff ff ff 00 …` → `ff 00 …`).
  The values and types of the last two are unchanged.
  `TestUntypedRuneConstLayout` pins these bytes.

## Related, not fixed here

**Kind of mixed rune/int constants (#6339)** and **context typing of
non-constant shifts.** `'a' + 1` is still an untyped int (Go: untyped rune),
and non-constant shift operands are still not typed from their context.
Divergences that remain, all present on master:

- `r := 'a' + 1; _ = []rune{r}`: gc accepts, Gno rejects
  (`cannot use int as int32`);
- `x := 1 + ('a' << s)`: gc 98 (rune), Gno 1;
- `println(('a' << s) < 100)`: gc true, Gno false;
- `var x interface{} = ('a' << s) + (1 << 40)`: gc rejects, Gno 97;
- `int64(('a' << s) + ('b' << s))` with `s = 32`: gc 837518622720, Gno 0;
- `int64(('a' << s) + (1 << s))`, `var x int64 = ('a' << s) + (1 << s)`: gc
  accepts, Gno rejects (`incompatible types`).

**Typed constant arithmetic** still folds with the runtime (wrapping)
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
| `gnovm/pkg/gnolang/preprocess.go` | `evalConst` hook, `untypedRuneOperands`, `evalUntypedRuneConst`, `keepBigUntypedRuneKind`, `rejectBigUntypedRune` (binary expressions, switch cases), `importedConst`, `foldStringOfLargeIntConst`, `float64`/`string` conversion calls |
| `gnovm/pkg/gnolang/values.go` | `GetUntypedRune`, `SetUntypedRune`, `Sign` |
| `gnovm/pkg/gnolang/values_conversions.go` | `ConvertUntypedRuneTo`, `convertUntypedIntegerTo` |
| `gnovm/pkg/gnolang/op_unary.go` | `-`/`^` on a non-constant untyped rune at runtime |
| `gnovm/pkg/gnolang/preprocess_import_const_test.go` | persisted-layout test, reload tests for imported constants and variables |
| `gnovm/tests/files/rune4.gno`–`rune40.gno`, `switch52.gno` | filetests |
