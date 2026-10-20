# ADR: Native `int256` and `uint256` types

## Status

Proposed.

## Context

Amounts coming from Ethereum (token balances, `uint256` event fields, IBC
packet data from EVM chains) are 256-bit unsigned integers. Gno's widest
integer is 64 bits, so realms that handle such amounts use
`gno.land/p/onbloc/uint256`, a struct-based library type. That works, but:

- every value is a `*Uint` pointing at a `/p/`-typed struct. Arithmetic
  mutates the receiver in place (`z.Add(x, y)`), so a value held in realm
  state is a mutable object that other code can reach and change, the
  pattern AGENTS.md and the security guide warn about;
- the values carry ownership stamps like any `/p/` object. Cross-realm
  in-place arithmetic on them hit the taint and ownership rules
  (issue #5736, fixed by #5747's type-driven copy stamping), and every
  caller has to reason about pointer aliasing and readonly taint for what is
  conceptually a number;
- each operation is many interpreted Gno operations on four `uint64` limbs,
  and each value costs a struct object in storage.

A primitive type has none of these problems: values are copied by
assignment, cannot be aliased, carry no ownership, and each operator is one
VM operation.

## Decision

Add two typed primitive types, `int256` and `uint256`, defined in the
universe like `int64` and `uint64`. They follow Go's integer rules exactly,
generalised to 256 bits. `int256` is two's complement.

### Representation

- `PrimitiveType`: `Int256Type` and `Uint256Type` are appended after
  `UntypedBigdecType` (bits 21 and 22). `PrimitiveType` values are
  amino-persisted as varints, so they must not renumber existing types.
- `Kind`: `Int256Kind` and `Uint256Kind` are inserted after `Uint64Kind`.
  `Kind` values are never persisted (no amino-registered struct has a `Kind`
  field) and never compared numerically, so the position is free; the
  stringer is regenerated.
- Uverse: `int256` and `uint256` are defined last, so the block indices of
  the existing uverse names, which persisted code refers to, do not move.
- Values do not fit `TypedValue.N` (8 bytes), so they live in `TypedValue.V`
  as `Bits256Value{V uint256.Int}`: the 256 raw bits, by value. The
  signedness comes from `T`. One Value type serves both kinds.
  - `github.com/holiman/uint256` (BSD-3-Clause, the go-ethereum 256-bit
    library) provides the arithmetic. It is a fixed `[4]uint64`, with no
    hidden allocation, and implements the signed operations (`SDiv`,
    `SMod`, `SRsh`, `Slt`) with Go's truncating semantics.
  - A nil `V` is zero. Every operation stores a zero result as nil, so each
    value has one representation and default values cost nothing.
  - Values are immutable: an operation builds a new `Bits256Value`, it never
    writes through an existing one. Copying a `TypedValue` shares the boxed
    array, which is safe because nothing mutates it; `Copy` returns the value
    as is.
- Amino: `Bits256Value` marshals to minimal big-endian bytes (empty for
  zero, at most 32). Unmarshal accepts only that canonical form (no leading
  zero byte), so a value has one encoding in realm state. JSON shows it as
  base64. The type is registered in `package.go` and its genproto2 code
  regenerated.
- Map keys: the 32 big-endian bytes, behind the usual type prefix, so
  `int256(1)` and `uint256(1)` are distinct keys inside `map[any]...`.
- Allocation: a non-zero result charges `allocBits256` (32-byte box plus
  heap overhead); `GetShallowSize` returns the same, and loading an object
  from the store charges for the `Bits256Value`s it holds inline (in
  `internalRefSize`, next to the RefValue slots it already counted), so the
  GC recount agrees with what was allocated.

### Semantics

- `+`, `-`, `*` and unary `-` wrap modulo 2^256.
- `/` and `%` truncate toward zero; the remainder has the sign of the
  dividend. Division by zero panics with Go's runtime error.
  `MinInt256 / -1` wraps to `MinInt256` and `MinInt256 % -1` is 0, as in Go.
- `&`, `|`, `^`, `&^`, unary `^`. `>>` is arithmetic for `int256`, logical
  for `uint256`; `<<` and `>>` by 256 or more give 0 (or -1 for a negative
  `int256` shifted right). The shift count may itself be a 256-bit value: it
  is not converted to `uint` (which would truncate it), and a count that does
  not fit a `uint64` saturates. A negative count panics.
- Comparisons, `++`/`--`, and every compound assignment.
- Conversions:
  - between integer kinds: truncation to the low bits, sign extension from
    signed sources, zero extension from unsigned ones;
  - to floats: rounding to nearest, ties to even (through `math/big`, which
    is software and deterministic); a `uint256` beyond float32's range
    becomes `+Inf`, as in Go;
  - from floats: truncation toward zero. Go leaves out-of-range float to
    integer conversion implementation-defined; Gno defines it, in the spirit
    of softfloat's 64-bit conversions: NaN and ±Inf give 0; for `int256` a
    value outside [-2^255, 2^255) gives 0; for `uint256` a value in
    [-2^255, 0) wraps modulo 2^256 (as `uint64(-1.0)` does) and one outside
    [-2^255, 2^256) gives 0;
  - `string(x)` follows Go's rune rule (out of the code point range gives
    `"�"`); `int256("1")` is refused, as for every integer type.
- Indexing, `make` and slicing accept 256-bit integers like any integer
  type; a value outside the int64 range saturates so the bounds check
  rejects it, instead of using its truncated low bits.
- Constants: an untyped constant converts with an exact range check,
  [0, 2^256-1] for `uint256` and [-2^255, 2^255-1] for `int256`
  ("bigint overflows target kind", as for the other kinds). Typed constant
  conversions use the error of `gnovm/adr/pr5269_const_conversion_overflow.md`
  ("constant X overflows Uint256Kind"). Constant folding works.
- `println`/`print`, `values_string_stream`, bounded strings, frame traces,
  gnoweb's state view and the test `fmt` (`%d %v %b %o %O %x %X` with width,
  precision and flags) print the decimal value.

### go/types

Gno type-checks every package with go/types on chain
(`gotypecheck.go`), and go/types knows no 256-bit type. The shims are

```go
type int256 int
type uint256 uintptr
```

declared in go/types' universe scope (`types.Universe`, extended once in
`init`, before any type-check can run), with a `types.Sizes`
(`gnoTypesSizes`) that returns 32 for `uintptr` only.

They are universe types rather than declarations in the injected
`.gnobuiltins` file, as `realm` and `address` are, because a `.gnobuiltins`
declaration lives in the package scope: go/types then reports
"uint256 already declared" for a file that imports a package named
`uint256` (`gno.land/p/onbloc/uint256`), before Gno can give its own error.
A universe object is also the same type in every package, as Gno's
builtins are. The cost is a process-global change to go/types: any Go code
type-checked in the same process also sees the two names (as shadowable
builtins), which no Gno binary does today.
go/types sizes `uint`/`uintptr` through `Config.Sizes` when it checks
constant representability, so it accepts every `uint256` constant up to
2^256-1 and rejects larger or negative ones. Gno has no `uintptr` (it is not
in the uverse), so user code cannot name it: the preprocessor rejects it
("name uintptr not declared"), whatever go/types thinks.

`int`'s size is not changed: go/types computes a sized `int`'s bounds as
`int64(-1)<<(s-1)`, which is wrong for `s > 64` and makes small negative
constants fail. So go/types accepts `int256` constants only in the int64
range.

**Overflow deferral (owner decision).** go/types' numeric-overflow errors
are dropped, and Gno's preprocessor is the only constant-range check, for
every numeric kind. `isDeferredOverflowError` identifies them by error code,
read from `types.Error`'s unexported `go116code` field, which go/types
documents as readable by reflection:

- `NumericOverflow` (44): an untyped constant that does not fit its target
  (assignment, argument, return, operand, index), or a typed constant
  operation whose result does not fit its type;
- `InvalidConversion` (96) whose message is `constant X overflows T`: the
  concise form go/types gives integer-to-integer constant conversions
  (go.dev/issue/63563). Other invalid conversions keep their error, hence
  the narrow message match on top of the code.

The codes are internal to the toolchain (`internal/types/errors`), so
`TestDeferredOverflowErrorCodes` pins both the values and the reflection on
the toolchain in use (Go 1.25), and `TestGoTypesSizes` pins what the shims
accept: a toolchain that changes either fails the tests instead of silently
changing what is dropped. If the field cannot be read, nothing is dropped,
which only rejects more programs.

**Safety argument.** Dropping an error cannot let an invalid program through
only if the preprocessor rejects every constant overflow go/types reported.
Before this change it did not: it computed typed constant expressions with
the type's wrapping arithmetic, so `const c int8 = 100; c * 2` folded to -56
(go/types was the only check). `checkConstFoldOverflow` now recomputes every
folded typed integer `+`, `-`, `*`, `/`, `<<` and unary `-` exactly with
`math/big` from its constant operands and rejects a result outside the type;
`%`, `&`, `|`, `^`, `&^` and `>>` of representable operands are
representable. A typed float result that overflowed to ±Inf is rejected too
(constants have no infinities), since `NumericOverflow` also covers floats.
The other paths already checked: untyped constants (`ConvertUntypedBigintTo`,
`ConvertUntypedBigdecTo`) and typed constant conversions (`ConvertTo`'s
`validate`).

`gnovm/tests/files/constoverflow/` holds one filetest per integer kind
(`int`, `int8` ... `int64`, `uint` ... `uint64`, `int256`, `uint256`) per
context: var declaration (over and under), const declaration, assignment,
compound assignment, function argument, return value, binary operation with
an untyped constant, typed constant folding (`+` and `*`), unary minus and
shift. All 144 are rejected by the preprocessor. For the 120 non-256 ones,
go/types alone (standard sizes) reports an error in every file, with codes
44 or 96 only, so the set that go/types used to reject is exactly the set
the preprocessor now rejects. The 31 existing filetests that showed both a
Gno `Error:` and a go/types `TypeCheckError:` for an overflow lose the
latter; none lost its `Error:`.

A known limitation: go/types reports `var x int256 = 1e30` (a float
literal that is a whole number beyond int64) as `TruncatedFloat`, which is
not dropped, so such constants must be written as integers (`1 << 100`,
`1000...000`). This only rejects valid programs; it does not affect
`uint256`.

### Gas

Each 256-bit operator has an `OpCPU*Int256` constant (`machine.go`), and the
handlers charge it instead of the 64-bit cost (Add, Sub, Mul, Quo, Inc, Dec)
or charge its difference from the 64-bit cost already charged at dispatch
(the other operators, and compound assignments). Conversions to or from a
256-bit kind have their own costs, higher when a float is involved.

The values are ratio-scaled, as `OpCPULazyBoundResolve` is: each 256-bit
handler is benchmarked next to the same handler on `int` in one run
(`bench_int256_test.go`, median of 10 runs of 2e6 operations, full-width
operands, the slowest divisor size and sign for `/` and `%`), and the
existing reference constant is multiplied by the measured ratio, so the
machine-speed factor cancels (`<` and integer conversions were re-measured
after their implementation was made allocation-free). The measurements were taken on an Apple
M-series development machine; they include boxing the result, which alloc
gas also charges, so they err on the high side. They should be re-measured
on the reference hardware with the rest of the table.

| op | int (ns) | 256 (ns) | 64-bit gas | 256-bit gas |
|----|---------:|---------:|-----------:|------------:|
| `+` | 36.7 | 99.8 | 81 | 221 |
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
| `==` | 51.4 | 59.4 | 93 | 108 |
| `<` | 40.0 | 59.8 | 73 | 110 |
| unary `-` | 44.1 | 63.6 | 69 | 100 |
| unary `^` | 43.2 | 63.9 | 69 | 103 |
| convert, integer | 30.2 | 54.7 | 151 | 274 |
| convert, float | 32.5 | 135.9 | 151 | 632 |

`TestGasInt256Ops` pins the difference against `int` per operator, and
`gnovm/tests/files/gas/int256_ops.gno` with `int_ops_baseline.gno` pin the
gas of the same 100-iteration loop on both types (619248 against 471747).

### Native bindings and the rest of the node

- Natives: a Gno `uint256` maps to `gnolang.Uint256` (an alias of
  `uint256.Int`) and `int256` to `gnolang.Int256` (a distinct type over the
  same bits), both by value, in `Gno2GoValue`/`Go2GnoValue` and in
  `misc/genstd`, so stdlib natives can take and return them.
- `gno.land/pkg/sdk/vm/convert.go`: `maketx call` arguments are parsed from
  canonical decimal strings only (no `+`, no leading zero, no `-0`, at most
  78 characters), so each value has one spelling, and checked against the
  type's range.
- Stdlib (all in Gno, no natives):
  - `math.MaxInt256`, `math.MinInt256`, `math.MaxUint256`;
  - `strconv.FormatInt256`, `FormatUint256`, `AppendInt256`,
    `AppendUint256`, `ParseInt256`, `ParseUint256`, which behave like their
    64-bit counterparts without the bit-size argument. Formatting peels off
    the largest power of the base that fits a `uint64`, so a 78-digit number
    costs a handful of 256-bit divisions;
  - `math/overflow`: `Add256`, `Sub256`, `Mul256`, `Div256` and the
    unsigned `Addu256` ... `Divu256`, returning `(result, ok)`, with the
    panicking `...p` variants, generated from the existing template.

## Alternatives considered

1. **Widen `TypedValue.N` to 32 bytes.** Keeps the values out of the heap,
   but grows every `TypedValue` (blocks, arrays, struct fields, the operand
   stack) by 24 bytes for a type most code never uses, and changes the
   persisted layout of every value.
2. **Keep the library type** (`p/onbloc/uint256`) and fix its ergonomics.
   It stays a mutable pointer-shaped object: aliasing, readonly taint and
   ownership stamps remain the caller's problem (#5736), and each operation
   stays a few dozen interpreted operations.
3. **Fork go/types**, or patch it, to know 256-bit types. Exact, but Gno
   would carry a fork of a large toolchain package on chain and track
   upstream by hand. The shims plus the overflow deferral need no fork.
4. **`uint256` only.** Covers token amounts, but signed deltas (balance
   changes, price ticks, `int256` fields in Ethereum ABIs) would need ad hoc
   encodings. The signed type costs little once the unsigned one exists,
   since both share the representation.
5. **Panicking arithmetic** (overflow aborts the transaction). Safer for
   amounts, but not Go's integer semantics, and it would make `int256`
   differ from every other integer type. Checked arithmetic is provided
   explicitly in `math/overflow` instead.

## Consequences

- This is a consensus-breaking VM change: new types, new op semantics and
  gas, a new amino type, new stdlib functions (which also shift the object
  IDs of stdlib objects, see `zrealm_stdlib_ref.gno`), and a preprocessor
  that rejects typed constant overflows go/types used to catch. It needs a
  chain upgrade; it cannot be enabled on a running chain by a code update
  alone.
- `int256` and `uint256` become reserved identifiers. Gno does not let any
  declaration shadow a builtin (#1779), imports included, so code that
  declares those names, or imports a package named `int256` or `uint256`
  without an alias, no longer compiles. In `examples/`, only
  `p/onbloc/int256` was affected: it now imports `p/onbloc/uint256` as
  `u256`, the alias its other users already use. Packages deployed on a
  chain that use those names would fail to load after the upgrade, so the
  upgrade must check deployed code for them.
- Otherwise, every program accepted before is still accepted: the
  preprocessor's new checks reject only programs go/types already
  rejected. Error messages for constant overflow now come from the
  preprocessor only.
- Adding functions to `strconv` grows that package, so code that imports
  it pays slightly more gas to load it (`stdlib_restart_compare.txtar`:
  1985078 to 1992634 for the same call).
- Chain coins (`std.Coin`, the banker) stay `int64`. A realm that holds
  256-bit token balances keeps them in its own state; converting to coins
  needs an explicit range check.
- `p/onbloc/uint256` keeps working; new code can use the native types.

## Key files

| File | Role |
|------|------|
| `gnovm/pkg/gnolang/values_int256.go` | `Bits256Value`, accessors, conversions, arithmetic helpers |
| `gnovm/pkg/gnolang/types.go`, `uverse.go` | the types and kinds; uverse names |
| `gnovm/pkg/gnolang/op_binary.go`, `op_unary.go`, `op_inc_dec.go`, `op_assign.go` | operators |
| `gnovm/pkg/gnolang/values_conversions.go` | conversion dispatch, untyped constants |
| `gnovm/pkg/gnolang/preprocess.go` | `checkConstFoldOverflow`, 256-bit shift counts |
| `gnovm/pkg/gnolang/gotypecheck.go` | universe shims, `gnoTypesSizes`, `isDeferredOverflowError` |
| `gnovm/pkg/gnolang/machine.go` | `OpCPU*Int256`, `incrCPU256` |
| `gnovm/pkg/gnolang/gonative.go`, `misc/genstd/mapping.go` | native bindings |
| `gno.land/pkg/sdk/vm/convert.go` | transaction arguments |
| `gnovm/tests/files/int256/`, `gnovm/tests/files/constoverflow/` | filetests |
| `gno.land/pkg/integration/testdata/int256_args.txtar` | arguments, results, persistence across a restart |
