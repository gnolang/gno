# PR5678: math/big stdlib (Int subset)

## Context

Gno today has no user-facing arbitrary-precision integer type:

- `UntypedBigintType` / `UntypedBigdecType` exist only at preprocess time
  for constant folding (e.g. `1 << 200` evaluated in a const context).
  They are never exposed to user code and constants must collapse to a
  sized type (`int`, `int64`, `uint64`, etc.) before runtime.
- The runtime has full machinery — `BigintValue`, `BigdecValue`,
  `BigintKind`/`BigdecKind` in `op_binary.go` — but no `.gno` program
  can reach it.
- The ecosystem has filled the gap with fixed-width libraries
  (`p/onbloc/int256`, ~2.3k lines of hand-written `[]uint64` arithmetic).
  Arbitrary precision is unavailable.
- PR #5646 wired CPU gas metering for `BigintKind`/`BigdecKind` operator
  helpers (`isEql`, `isLss`, etc.). The ADR for that PR called out that
  the metering covers an unreachable path. Either we expose the type
  (the metering becomes live) or we should revisit whether the metering
  is worth keeping. This PR takes the first option for `Int`.

## Decision

Add `math/big` as a stdlib, starting with `Int` only. The `.gno` API
mirrors Go's `math/big.Int` so existing Go code ports with minimal
change.

Wire format: an `Int` carries `(neg bool, abs []byte)` where `abs` is the
big-endian unsigned magnitude with no leading zeros (empty for zero).
Methods that do heavy lifting (`Add`, `Sub`, `Mul`, `QuoRem`, `DivMod`,
`SetString`, `Text`) call into native Go-side functions via the existing
`genstd` bridge; the Go side decodes to `*big.Int`, computes, re-encodes.
Trivial methods (`Set`, `SetInt64`, `Sign`, `Neg`, `Abs`, `Cmp`, `Bytes`,
`SetBytes`, `BitLen`, etc.) are pure Gno.

## Scope

Included in this PR (subset of Go's `math/big.Int`):

- Construction / conversion: `NewInt`, `Set`, `SetInt64`, `SetUint64`,
  `Int64`, `Uint64`, `IsInt64`, `IsUint64`, `Bytes`, `SetBytes`,
  `SetString`, `String`, `Text`
- Sign / magnitude: `Sign`, `Neg`, `Abs`, `BitLen`
- Comparison: `Cmp`, `CmpAbs`
- Arithmetic: `Add`, `Sub`, `Mul`, `Quo`, `Rem`, `QuoRem`, `Div`, `Mod`,
  `DivMod`

Explicitly deferred (see "Follow-ups"):

- `big.Float`, `big.Rat`
- Bit operations (`Lsh`, `Rsh`, `And`, `Or`, `Xor`, `AndNot`, `Bit`,
  `SetBit`)
- Advanced ops (`Exp`, `ModInverse`, `GCD`, `Sqrt`, `ProbablyPrime`)
- Formatting helpers (`Format`, `Append`, `FillBytes`)
- Encoding helpers (`MarshalJSON`, `GobEncode`, `MarshalText`)

## Alternatives Considered

1. **Pure-Gno reimplementation of `math/big`** (Karatsuba on `[]uint64`,
   etc.). Honest but enormous — Go's upstream is roughly 5k LOC for `Int`
   alone, plus `nat`. Native bridge gets us 95% of the value at 5% of the
   surface; the API surface stays portable so a pure-Gno backing can
   replace the native one later without changing call sites.

2. **Hook `*big.Int` directly into the VM as a first-class `BigintKind`
   type**, so `==`, `<`, etc. dispatch through the existing operator
   handlers in `op_binary.go` (and PR #5646's metering becomes live).
   This is the right long-term answer but requires preprocess changes to
   accept a user-facing type with `BigintKind`, plus careful spec work
   for operator/literal semantics. Out of scope here. See "Open
   Question."

3. **Don't add the type; revert PR #5646.** Coherent — if no `.gno`
   program can hold a runtime bigint, gas-charging the operator helpers
   is dead defensive code. Rejected because real ecosystem demand exists
   (see `p/onbloc/int256`).

## Implementation Notes

### Wire format

`(neg bool, abs []byte)` chosen over a single signed-magnitude `[]byte`
or a length-prefixed encoding because:

- Two clean primitive params with no parser. `genstd` handles `bool` and
  `[]byte` directly with no custom type plumbing.
- Mirrors `*big.Int`'s internal representation closely. `toBig` /
  `fromBig` are trivial.
- Empty `abs` is canonical zero. `neg=true, abs=nil` is rejected at the
  Gno setter so the `Sign() == 0` check is just `len(abs) == 0`.

### Aliasing

Go's `math/big` allows the receiver to alias any argument
(`z.Add(z, y)`, `z.Mul(z, z)`, etc.). Native calls preserve this because
arguments are read out of the call block into Go locals before the X_
function runs; the result `(neg, abs)` is fresh and assigned to the
receiver only after the call returns. Tested in `TestAliasing`.

### Gas

Seven new native functions (`add`, `sub`, `mul`, `quoRem`, `divMod`,
`setString`, `text`) are registered in
`gnovm/stdlibs/native_gas.go` with **placeholder** calibration values.
They are conservative (over-charge rather than under-charge), but real
calibration must come from a benchmark run in
`gnovm/cmd/calibrate/native_bench_test.go` before any consensus-relevant
deployment. The placeholders are clearly marked `TODO: calibrate` in
the table comment and on each row.

Single-slope (`SlopeIdx=1` on the first `[]byte` operand) under-counts
for binary ops when `|b| > |a|`. A two-dimensional fit (Slope2 on `b`)
is in scope for the calibration follow-up — the schema already supports
it; only the placeholder rows here use the single-slope form.

`setString` is the exception: its first parameter is the input `string`,
not a sign `bool`, so its row uses `SlopeIdx=0` (sloping on the length
of `s`) rather than `SlopeIdx=1`.

### Note on PR #5646

This PR does **not** make PR #5646's `BigintKind` operator metering
reachable. `*big.Int` is a regular Gno struct here, not a
`BigintKind`-typed value at the VM level — so `==`, `<`, etc. on a
`*big.Int` go through ordinary struct/pointer comparison, not through
`op_binary.go:isEql`/`isLss`. The metering in PR #5646 remains
unreachable from `.gno` source.

Making it reachable is a separate, larger piece of work: it requires
either (a) hooking `math/big.Int` into the VM as a `BigintKind`-typed
type with operator support (preprocess + type-checker changes), or
(b) adding a `bigint` primitive keyword. Alternative 2 above. The two
PRs are decoupled — this one stands on its own utility for users today,
and the operator path can come later without invalidating the API
shipped here.

## Consequences

### Positive

- Gno gets arbitrary-precision integers in the stdlib, matching Go's
  API. Ecosystem libraries (`p/onbloc/int256`, future u256 wrappers)
  can wrap or migrate.
- `genstd` bridge surface stays small (7 native functions, all using
  primitive Gno types).
- Determinism is inherited from Go's `math/big` (deterministic across
  platforms by construction).

### Negative

- Per-op overhead is non-trivial: every arithmetic call goes through
  `Gno2GoValue` (param decode) → `SetBytes` (build `*big.Int`) → op →
  `Bytes` (encode result) → `Go2GnoValue` (push). For tight loops over
  small integers, this is meaningfully slower than fixed-width
  alternatives like `int64` or `p/onbloc/int256`.
- Placeholder gas values are not safe for mainnet. A calibration pass is
  a hard prerequisite before any production rollout.

## Open Question

Should we follow up with **(2)** from "Alternatives" — exposing
`big.Int` as a true `BigintKind` value at the VM level so `==`, `<`,
`+`, etc. dispatch through `op_binary.go`?

Arguments for: makes the type natural to use (`a + b` instead of
`new(big.Int).Add(a, b)`), makes PR #5646's metering live, retires the
dead-code question for the `Bigint`/`Bigdec` runtime paths.

Arguments against: breaks Go API parity (Go's `*big.Int` does *not*
support operators), opens a non-trivial spec discussion (literal
syntax, type inference rules, how `bigint` interacts with `int`/`int64`
in mixed expressions), and adds VM/preprocess surface for a feature
that the method-based API already covers.

I think this is worth a separate ADR after the method-based stdlib has
seen real use. Filing as a tracking issue rather than embedding the
decision here.

## Follow-ups

- Real gas calibration: add `BenchmarkNative_MathBig_*` benches to
  `gnovm/cmd/calibrate/native_bench_test.go` (sweeping input lengths
  for add/sub/mul/divMod, plus base/length sweeps for setString/text),
  re-fit, replace placeholder rows in `native_gas.go`.
- Bit operations: `Lsh`, `Rsh`, `And`, `Or`, `Xor`, `AndNot`, `Bit`,
  `SetBit`. Add one native per op; same wire format.
- `Exp`, `ModInverse`, `GCD`, `Sqrt`, `ProbablyPrime`.
- `big.Float`, `big.Rat`. Larger surface; defer until `Int` has shipped
  and the calibration model is validated.
