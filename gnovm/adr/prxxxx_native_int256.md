# ADR: Native `int256` and `uint256` types

## Status

Proposed. Ships with a new chain (see "Deployment").

## Context

Amounts coming from Ethereum (token balances, `uint256` event fields, IBC
packet data from EVM chains) are 256-bit unsigned integers. Gno's widest
integer is 64 bits, so realms that handle such amounts use the library
`gno.land/p/onbloc/uint256/v0` (and `p/onbloc/int256/v0`). Its `Uint` is a
struct of four `uint64` limbs; the constructors return `*Uint`, and the
arithmetic is a mutating pointer-receiver API (`z.Add(x, y)` writes `z`).
That works, but:

- a value reached through a pointer is a mutable `/p/`-typed object, the
  pattern AGENTS.md and the security guide warn about, and cross-realm
  in-place arithmetic hit the taint and ownership rules (issue #5736, fixed
  by #5747's type-driven copy stamping); callers have to reason about
  aliasing for what is conceptually a number;
- each operation is many interpreted Gno operations, and there are no
  operators (`z.Add(x, y)` instead of `x + y`);
- each stored value is a struct object.

A primitive type has none of these problems: values are copied by
assignment, cannot be aliased, carry no ownership, have operators, and each
operator is one VM operation.

## Decision

Add two typed primitive types, `int256` and `uint256`, defined in the
universe like `int64` and `uint64`. Run-time semantics follow Go's integer
rules generalised to 256 bits. `int256` is two's complement. One language
restriction applies to `int256` constants (see "go/types").

### Representation

- `PrimitiveType`: `Int256Type` (1<<21) and `Uint256Type` (1<<22) are
  appended after `UntypedBigdecType`. `PrimitiveType` values are
  amino-persisted, so existing ones must not be renumbered.
- `Kind`: `Int256Kind` and `Uint256Kind` sit after `Uint64Kind`. `Kind`
  values are never persisted or compared numerically; the stringer is
  regenerated.
- Uverse: `int256` and `uint256` are defined last, so the block indices of
  the existing uverse names do not move.
- Values do not fit `TypedValue.N` (8 bytes), so they live in `TypedValue.V`
  as `Int256Value` or `Uint256Value`, each `struct{ V uint256.Int }`: the
  256 bits by value (`github.com/holiman/uint256`, BSD-3-Clause, the
  go-ethereum library, which provides Go's truncating signed operations).
  Two Value types, rather than one, let the amino encoding carry the
  signedness.
- Zero is a nil `V`. Every operation stores a zero result as nil, and a boxed
  zero is neither produced nor accepted: `MarshalAmino` refuses it and
  `UnmarshalAmino` rejects it. Amino can still build one without calling
  `UnmarshalAmino`, from a JSON `null` payload or an absent binary payload,
  so `checkCanonical256` checks the enclosing value, recursively through
  inline slots, interfaces and composites, and the store runs it on every
  object it loads. Zero therefore has exactly one encoding
  (`TestBits256CanonicalZero`, `TestBits256CanonicalZeroJSONNull`,
  `TestBits256Amino`).
- Values are immutable: an operation builds a new value and never writes
  through an existing one. Copying a `TypedValue` shares the boxed array,
  which is safe because nothing mutates it.

### Amino and JSON

The amino representation is a canonical decimal string, in binary and JSON
alike: digits with no leading zero, `-` for a negative `int256`, nothing
else. The decoder rejects any other spelling, and values out of range. In
JSON, as returned by `qeval_json` and `qobject_json`:

```json
{"T":{"@type":"/gno.PrimitiveType","value":"2097152"},"V":{"@type":"/gno.Int256Value","value":"128"}}
{"T":{"@type":"/gno.PrimitiveType","value":"2097152"},"V":{"@type":"/gno.Int256Value","value":"-128"}}
{"T":{"@type":"/gno.PrimitiveType","value":"4194304"},"V":{"@type":"/gno.Uint256Value","value":"115792089237316195423570985008687907853269984665640564039457584007913129639935"}}
{"T":{"@type":"/gno.PrimitiveType","value":"4194304"}}
```

The last line is zero: no `V`, and no `N` either. A decimal string avoids the
pitfalls of bytes (padding, a signed reading of a minimal payload that turns
`int256(128)` into -128) and of JSON numbers (precision loss in JavaScript).
It also stores typical amounts compactly (a 10^24 balance is 25 bytes).

**Client contract change:** every other fixed-width, typed numeric
primitive stores its value in `N`, and a client reads it there (the untyped
bigint and bigdec constant types already use `V`). For these two types `N` is always empty and the value is
in `V`, as for strings. A client that reads only `N` for numeric primitives
shows 0; such clients must handle the two new value types. Pinned by
`TestBits256AminoJSON` and `int256_args.txtar`.

### Semantics

- `+`, `-`, `*` and unary `-` wrap modulo 2^256.
- `/` and `%` truncate toward zero; the remainder has the sign of the
  dividend. Division by zero panics with `runtime error: division by zero`
  (Gno's text for every integer type; Go's is `integer divide by zero`).
  `MinInt256 / -1` wraps to `MinInt256` and `MinInt256 % -1` is 0, as in Go.
- `&`, `|`, `^`, `&^`, unary `^`. `>>` is arithmetic for `int256`, logical
  for `uint256`; shifting by 256 or more gives 0 (or -1 for a negative
  `int256` shifted right). The count may itself be a 256-bit value; it is
  not converted to `uint` (that would truncate it), and a count that does
  not fit a `uint64` saturates. A negative count panics.
- Comparisons, `++`/`--`, and every compound assignment.
- Map keys: a key is encoded as its type ID prefix followed by the 32-byte
  big-endian two's complement bits (fixed width, zero included). Because the
  type ID is part of the key, `int256(1)` and `uint256(1)`, or `int256(-1)`
  and `uint256(MaxUint256)`, which have the same bits, are distinct keys in a
  `map[any]V`.
- Conversions:
  - between integer kinds: truncation to the low bits, sign extension from
    signed sources, zero extension from unsigned ones;
  - to floats: rounding to nearest, ties to even, through `math/big`
    (software, deterministic); a `uint256` beyond float32's range becomes
    `+Inf`;
  - from floats: truncation toward zero. Go leaves out-of-range float to
    integer conversion implementation-defined; Gno defines it in the spirit
    of softfloat's 64-bit conversions: NaN and ±Inf give 0; for `int256` a
    value outside [-2^255, 2^255) gives 0; for `uint256` a value in
    [-2^255, 0) wraps modulo 2^256 (as `uint64(-1.0)` does), and one outside
    [-2^255, 2^256) gives 0;
  - `string(x)` follows Go's rune rule (out of range gives `"�"`);
    `int256("1")` is refused.
- Indexes, slice expressions, and slice lengths and capacities in `make`
  accept 256-bit integers like any integer type. A value outside the int64
  range saturates (to -1 or `MaxInt64`) so that the bounds check rejects it.
  A size hint in `make(map[K]V, n)` is accepted and ignored, as for every
  integer type. A 3-index slice always checks its max,
  also when it is -1; this also fixes a pre-existing case where a max of -1
  was taken for "no max".
- Constant errors in the preprocessor (on chain, go/types reports these
  first):
  - an untyped integer constant out of range: `bigint overflows target
    kind` or `bigint underflows target kind`; an untyped float constant
    goes through the same check once it is a whole number. A fractional
    one gives `cannot convert (const (1.5 <untyped> bigdec)) to integer
    type` in an explicit conversion such as `uint256(1.5)`, and `cannot
    convert untyped bigdec to integer -- 1.5 not an exact integer` when it
    is converted implicitly (assignment, argument, operand);
  - a negative untyped rune converted to `uint256`: `rune underflows target
    kind`;
  - a typed integer constant converted out of range: `constant X overflows
    <Target>Kind` (pr5269), e.g. `constant -1 overflows Uint256Kind`;
  - a typed float constant that is not a whole number in range, converted
    to a 256-bit kind: `cannot convert constant of type <Source>Kind to
    <Target>Kind`;
  - a 256-bit constant converted to `float32` beyond its range: the
    preprocessor folds it to `+Inf` (the conversion runs with run-time
    semantics), and go/types rejects the program (`cannot convert c
    (constant ... of uintptr type uint256) to type float32`).
- Printing: `println`/`print`, `values_string_stream`, bounded strings and
  frame traces print the signed or unsigned decimal value. gnoweb's state
  view does too; when it reaches a value without its static type it uses
  the value type (`Int256Value` or `Uint256Value`), which carries the
  signedness. The test `fmt` supports `%v` and `%d` (decimal), `%b`, `%o`,
  `%O`, `%x`, `%X` (with width, precision and the `+ - # 0` and space
  flags), and `%#v` (hex with `0x` for `uint256`, decimal for `int256`, as
  for the 64-bit integers). `%c`, `%q` and `%U` are refused. A 256-bit
  value is accepted as a `*` width or precision argument when it fits an
  `int` and is within the formatter's limits, as for any integer: above
  a magnitude of 1,000,000 it is `%!(BADWIDTH)` or `%!(BADPREC)`; a negative
  width left-justifies and a negative precision
  is refused.

### go/types

Gno type-checks every package with go/types (`gotypecheck.go`) in every
transaction that adds or runs code, and go/types knows no 256-bit type. The
shims, declared in the central `gnobuiltins/gno0p9` package next to `realm`
and `address` so that the types are identical across packages, are

```go
type int256 int
type uint256 uintptr
```

with `GoTypesSizes`, a `types.Sizes` that returns 32 for `uintptr` only.
go/types sizes `uint`/`uintptr` through `Config.Sizes` when it checks
constant representability, so it accepts exactly the `uint256` constants in
[0, 2^256-1]. Gno has no `uintptr` (it is not in the uverse), so user code
cannot name it: the preprocessor rejects it. Nothing mutates go/types'
process-global universe scope.

`int`'s size is not changed: go/types computes a sized `int`'s bounds as
`int64(-1)<<(s-1)`, which is wrong for `s > 64`. **Language restriction: an
`int256` constant must fit in an int64.** go/types rejects, with its usual
overflow errors, a typed `int256` constant outside the int64 range, an
untyped constant assigned to `int256` outside it, a float constant
converted to `int256` outside it, and typed `int256` constant arithmetic
whose result leaves it:

```go
const c int256 = 1 << 100 // rejected
const f float64 = 1 << 100
var x = int256(f)         // rejected
const k int256 = 1 << 62
var y = k * 4             // rejected
```

Wider `int256` values come from run-time arithmetic or from converting a
non-constant value, typically a `uint256`:

```go
var bits uint256 = 1 << 255 // uint256 constants have the full range
min := int256(bits)         // -2^255
var one int256 = 1
wide := one << 100          // run-time shift
```

For the same reason `math.MaxInt256` and `math.MinInt256` are untyped
constants that cannot be used where an `int256` constant is expected; the
stdlib computes its int256 limits as variables (`strconv`, `math/overflow`).
Pinned by `TestGoTypesSizes` and the `err_restrict_*` filetests.

go/types' errors are kept as they are: no diagnostic is filtered. An
earlier draft dropped go/types' numeric-overflow errors to lift the
restriction, leaving range checks to the preprocessor. That was unsound: an
error invalidates the operand inside go/types, and the checks that depend
on it are then skipped, so dropping the diagnostic dropped those checks too
(for instance `var u uint64 = 'a' - 'b'` and `var r = '\U0010ffff' *
'\U0010ffff'` were accepted). It also read an unexported go/types field by
reflection, which made acceptance depend on the toolchain build. Both
programs are regression tests (`rune_overflow0.gno`, `rune_overflow1.gno`),
rejected by go/types.

With go/types' errors kept, the preprocessor's constant folding is
unchanged for every existing kind: no constant check was added there, so
query paths that preprocess without go/types keep their current behaviour.

**Tooling.** `gno lint`, `gno test` and the node all type-check through
`TypeCheckMemPackage`, so they use the same shims and sizes. An external
checker, such as an editor's language server, must use the shims from
`gnoBuiltinsGno0p9` and `gnolang.GoTypesSizes`, exported for that purpose.
`gno fmt` needs no change: the names are ordinary identifiers to the parser.

### Gas

Each 256-bit operator has an `OpCPU*Int256` constant (`machine.go`). Every
handler charges the operator's 64-bit cost as for any integer, then
`incrCPU256` adds the difference. Unary `+` is a no-op and costs the same as
for any type. A compound assignment is charged its flat `OpCPU*Assign` cost
plus the operator's surcharge: `x += y` on `uint256` costs
`OpCPUAddAssign` + (221 − 81), that is 97 + 140. A conversion between a 256-bit kind and
another kind is charged `OpCPUConvertInt256` or, when a float is involved,
`OpCPUConvertInt256Float`, instead of `OpCPUConvertNumeric`; retyping within
the same kind keeps the ordinary cost.

The values are ratio-scaled, as `OpCPULazyBoundResolve` is: each 256-bit
handler is benchmarked next to the same handler on `int` in one run
(`bench_int256_test.go`, median of 10 runs of 2e6 operations, full-width
operands, the slowest divisor size and sign for `/` and `%`), and the
existing reference constant is multiplied by the measured ratio, so the
machine-speed factor cancels. The measurements were taken on an Apple
M-series development machine.

| op | int (ns) | 256 (ns) | 64-bit gas | 256-bit gas |
|----|---------:|---------:|-----------:|------------:|
| `+`, `++`, `--` | 36.7 | 99.8 | 81 | 221 |
| `-` | 41.2 | 100.0 | 70 | 170 |
| `*` | 40.7 | 95.6 | 71 | 167 |
| `/` | 51.6 | 144.0 | 138 | 386 |
| `%` | 60.2 | 152.6 | 142 | 360 |
| `&` | 41.4 | 82.2 | 71 | 141 |
| `\|` | 41.9 | 93.4 | 71 | 159 |
| `^` | 39.2 | 94.2 | 71 | 171 |
| `&^` | 41.0 | 95.0 | 71 | 165 |
| `<<` | 40.5 | 83.4 | 80 | 165 |
| `>>` | 39.8 | 82.4 | 79 | 164 |
| `==` (`!=`: 83 + 15) | 51.4 | 59.4 | 93 | 108 |
| `<` (`<=`, `>`, `>=`: 72 → 110) | 40.0 | 59.8 | 73 | 110 |
| unary `-` | 44.1 | 63.6 | 69 | 100 |
| unary `^` | 43.2 | 63.9 | 69 | 103 |
| convert, integer | 30.2 | 54.7 | 151 | 274 |
| convert, float | 32.5 | 135.9 | 151 | 632 |

Allocation is accounted for exactly as for strings, the other immutable
value held in `V` and shared by its copies:

- a box is charged once, when it is built: at run time when an operation,
  conversion or native binding builds a non-zero value (`SetBits256` on the
  machine's allocator), and during preprocessing when a constant conversion
  builds one (`chargeConstBits256`, on the transaction's preprocessing
  allocator, which the keeper installs; folded constant arithmetic runs on a
  machine that uses the same allocator). String literals are charged on that
  allocator when they are evaluated, so the two match there too
  (`TestBits256ConstPreprocessCharge` covers integer, rune, decimal, typed
  and 256-to-256 constant conversions). A conversion between values of the
  same kind keeps the existing box;
- copying a value into another slot shares the box and charges nothing,
  whatever the path (assignment, `TypedValue.Copy`, a method value formed
  through an interface, a saved method's receiver copied into its call
  block, a named result filled on return), and so do the temporary aliases
  on the operand stack;
- the GC recount visits every slot and counts one box per non-zero value,
  as it counts one string per slot;
- a reload from the store charges the object's shallow size, which, as for
  strings, does not include the inline values.

So the recount can exceed what a copy-heavy program was charged, as it can
for strings; the GC's limit check bounds both the same way.
`TestBits256AccountingMatchesStrings` checks, for one value copied into many
slots, retained interface method values, a saved method called repeatedly
and a retained named result, that a 256-bit value is allocated and recounted
in exactly as many slots as a string in the same program.

The equality surcharge is charged by the 256-bit comparison itself (in
`isEql`), so comparisons of arrays, structs and interface fields pay it per
256-bit element compared, as do switch cases; `!=` costs `OpCPUNeq` plus the
same surcharge. `TestGasInt256Ops` pins the difference from `int` for
`+ - * / % | == <`, unary `-`, `<<` and the zero-result case, and
`TestGasInt256Equality` pins equality in arrays (including an early
mismatch), structs, interface fields, switches and `!=`.
`gnovm/tests/files/gas/int256_ops.gno` and `int_ops_baseline.gno` pin the
gas of the same 100-iteration loop on both types.

**Follow-up before activation (tracked):** re-measure the table on the
reference hardware used for the rest of the gas table (Xeon 8168), with
worst-case operands for every operator: full-width values, every divisor
size and sign combination for `/` and `%`, float conversions at the range
limits. The table is accepted when every 256-bit constant is at least the
reference median of its worst case, and is then frozen for the launch.

### Native bindings and the rest of the node

- Natives: a Gno `uint256` maps to `gnolang.Uint256` (an alias of
  `uint256.Int`) and `int256` to `gnolang.Int256` (a distinct type over the
  same bits), by value, in `Gno2GoValue`/`Go2GnoValue` and in
  `misc/genstd`.
- `gno.land/pkg/sdk/vm/convert.go`: `maketx call` arguments are parsed from
  canonical decimal strings only (no `+`, no leading zero, no `-0`, at most
  78 characters) and range-checked.
- Stdlib, all in Gno:
  - `math.MaxInt256`, `math.MinInt256`, `math.MaxUint256` (untyped; see the
    restriction above);
  - `strconv.FormatInt256`, `FormatUint256`, `AppendInt256`,
    `AppendUint256`, `ParseInt256`, `ParseUint256`, like their 64-bit
    counterparts without the bit-size argument, including their error
    precedence: syntax and unsigned overflow are checked digit by digit, left
    to right, and the first failure is returned; then underscores (base 0
    only); then, for `ParseInt256`, the signed range;
  - `math/overflow`: `Add256`, `Sub256`, `Mul256`, `Div256` and the
    unsigned `Addu256` ... `Divu256`, returning `(result, ok)`, with the
    panicking `...p` variants. On overflow they return the wrapped result
    and false; `Div` returns 0 and false for a zero divisor.
- Existing GRC20 interfaces are unchanged; a uint256-based token standard is
  separate work (`p/nt/grc20/v1`, on the `feat/grc20-uint256` branch).

### Production helpers

Which helpers outside the VM support the new types:

- `strconv`, `math` and `math/overflow` (above): supported.
- The test `fmt`: supported (above); it is not available on chain.
- `p/nt/ufmt`: support is being added on the `feat/grc20-uint256` branch,
  with `p/nt/grc20/v1`; not part of this change.
- `p/onbloc/json`: exposes numbers as `float64`, which is lossy beyond 2^53
  for these types. The lossless route is a decimal string, written and read
  with `strconv.FormatUint256`/`ParseUint256` (and the `Int256` pair).
- Follow-up (tracked): native int256/uint256 number support in
  `p/onbloc/json`, or a documented string convention for JSON APIs.

### Reserved identifiers

Gno does not let any declaration shadow a builtin identifier (#1779),
imports included. `int256` and `uint256` are therefore reserved: code that
declares them, or imports a package named `int256` or `uint256` without an
alias, does not compile. In `examples/`, `p/onbloc/int256/v0` imported
`p/onbloc/uint256/v0` under its default name; it now uses the alias `u256`,
as the library's other users do.

## Deployment

This change ships with a new chain or genesis, as `gnoland-1` does.
Existing chains are not upgraded in place, and there is no migration
machinery. The change alters the VM (types, operators, gas, new amino
types), adds stdlib code (which changes the object IDs of stdlib objects,
see `zrealm_stdlib_ref.gno`), and reserves two identifiers, so state and
history produced by an older binary are not compatible with it. Historical
replay of existing chains with this binary is out of scope.

Before launch, the genesis packages are scanned for the reserved names: a
declaration of `int256` or `uint256`, or an unaliased import of a package
with either name. The scan is part of the genesis checks run by the team
preparing the genesis, and a conflict blocks the launch until the package is
fixed.

## Known issues

- Untyped rune constant arithmetic is computed in int32 and wraps before the
  result is converted, so `var w int64 = '\U0010ffff' * '\U0010ffff'` is
  -2228223 instead of 1241243320321. This is an existing bug on master
  (`mulAssign` groups `UntypedRuneType` with `Int32Type`), unchanged here; it
  affects `int256` targets the same way.

## Alternatives considered

1. **Widen `TypedValue.N` to 32 bytes.** Keeps values off the heap, but
   grows every `TypedValue` (blocks, arrays, struct fields, the operand
   stack) by 24 bytes for a type most code never uses, and changes the
   persisted layout of every value.
2. **A library type with an immutable value API**, possibly with native
   acceleration. Returning fresh values instead of mutating receivers removes
   the aliasing problem, needs no checker shims, reserves no identifiers, has
   no constant restriction, and natives would make each operation cheap. It
   still has no operators (`a.Add(b)` everywhere, no `<`, no `+=`), still
   stores each value as a struct object with its ownership metadata. The
   primitive gives arithmetic its usual syntax and cost; that is the
   decision.
3. **A pinned, bounded patch to go/types** that knows 256-bit sizes, which
   would lift the `int256` constant restriction. It is the exact solution,
   but go/types would become a vendored, patched dependency on the consensus
   path, rebased on every toolchain upgrade and reviewed as consensus code
   each time. The restriction costs a documented workaround instead. The
   patch remains an option if the restriction proves costly.
4. **Dropping go/types' overflow errors**, leaving range checks to the
   preprocessor, to lift the restriction without a patch. Rejected as
   unsound (see "go/types").
5. **`uint256` only.** Covers token amounts, but signed deltas would need ad
   hoc encodings; the signed type costs little given the shared
   representation.
6. **Panicking arithmetic.** Safer for amounts, but not Go's integer
   semantics; checked arithmetic is provided in `math/overflow` instead.

## Consequences

- A new chain is required (see "Deployment").
- `int256` and `uint256` are reserved identifiers.
- `int256` constants are limited to the int64 range.
- Clients that decode `qeval_json`/`qobject_json` values must read `V` for
  the two new types.
- Adding code to `strconv` grows that package, so code that imports it pays
  slightly more gas to load it (`stdlib_restart_compare.txtar`: 1985078 to
  1993542 for the same call).
- Chain coins (`std.Coin`, the banker) stay `int64`. A realm that holds
  256-bit balances keeps them in its own state; converting to coins needs an
  explicit range check.

## Key files

| File | Role |
|------|------|
| `gnovm/pkg/gnolang/values_int256.go` | value types, accessors, conversions, arithmetic helpers |
| `gnovm/pkg/gnolang/types.go`, `uverse.go` | the types and kinds; uverse names |
| `gnovm/pkg/gnolang/op_binary.go`, `op_unary.go`, `op_inc_dec.go`, `op_assign.go` | operators |
| `gnovm/pkg/gnolang/values_conversions.go` | conversion dispatch, untyped constants |
| `gnovm/pkg/gnolang/gotypecheck.go` | shims, `GoTypesSizes` |
| `gnovm/pkg/gnolang/machine.go` | `OpCPU*Int256`, `incrCPU256` |
| `gnovm/pkg/gnolang/gonative.go`, `misc/genstd/mapping.go` | native bindings |
| `gno.land/pkg/sdk/vm/convert.go` | transaction arguments |
| `gnovm/tests/files/int256/` | filetests |
| `gno.land/pkg/integration/testdata/int256_args.txtar` | arguments, results, JSON, persistence across a restart |
