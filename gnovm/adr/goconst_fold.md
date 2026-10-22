# Constant folding with go/constant

## Context

`evalConst` folded a constant operation by running it on a machine, with
runtime arithmetic. An untyped rune operation was computed in int32 and
wrapped (#6389: `'\U0010ffff' * '\U0010ffff'` gave -2228223, go/types
accepts the program with 1241243320321); typed integer and float
operations wrapped or went to infinity where gc rejects them (#6392);
unary `-` `^` on an untyped rune were rejected; an untyped integer
converted to float32 was rounded twice (#6394:
`int64(float32(1<<54 + 1<<30 + 1))`).

## Decision

A binary, unary or shift operation whose operands are constants of
primitive kinds is folded by `foldConst` (`constfold.go`) with Go's own
`go/constant`, used untouched: operands cross the boundary as
`constant.Value` (`BigintValue` → `*big.Int`, `BigdecValue` → `*big.Rat`
or 512-bit `*big.Float`, typed integers and floats exactly), the result
comes back under the operand's Gno type. go/constant has no rune kind, so
Gno's kinds are carried beside the value; the rune/int ranking (#6339) is
unchanged. Literals, names and conversions still go through the machine.

- An untyped result is exact. An untyped rune past int32 is stored as a
  `BigintValue` under `UntypedRuneType`; an in-range one keeps its int32
  in `N`, so persisted bytes of existing constants are unchanged. The
  range is checked when the constant takes a type: `var w int64 = c` is
  1241243320321, `var r = c` is `constant 1241243320321 overflows rune`,
  `var u uint64 = 'b' - 'c'` is `constant -1 overflows uint64`.
- A typed result is rounded once to its type (floats) or checked against
  it (integers), with gc's wording: `c * 2 (constant 200 of type int8)
  overflows int8`, `big * big (constant 1e+600 of type float64) overflows
  float64`, `k << 63 (constant 27670116110564327424 of type uint)
  overflows uint` (was `constant overflows`).
- Untyped integer and rune constants convert to float32/float64 through
  `constant.Float32Val`/`Float64Val`: one rounding, from the exact value.
- Unary `-` `+` `^` on an untyped rune fold (`-'a'` is -97); `'a' << 40`
  is the untyped rune 106652627894272.
- The result `TypedValue` starts as a copy of the left operand and only
  the kind's bytes of `N` are written, as the machine's in-place fold did,
  so a folded comparison keeps the same bytes and the app hash.

Runtime arithmetic (`op_binary.go`, `op_unary.go`) is untouched.

## Metering

`foldConst` charges `OpCPUSlopeConstFold` (600) once per operation on the
preprocess meter (a 512-bit integer division is ~180 ns on an M-series
laptop, `BenchmarkConstFold`), plus the operand-size terms the machine
charged for the same operation: untyped integers per kilobit, quadratic
for `*` `/` `%`, a shift by its amount and the operand's width, bigdecs
per hundred digits (`OpCPUSlopeBigInt*`, `OpCPUSlopeBigDec*`). A string
comparison adds `OpCPUSlopeBytesCmp` per byte; a concatenation is paid by
its allocation, as before. The old path charged VM opcodes; the four
addpkg gas pins and the stdlib `EXACT_GAS` pins do not move, since those
packages fold no constant operation under the tx meter.

## Limits

A shift count is bounded by `maxBigintShift` (10000) for every operand
kind, as the machine bounded untyped shifts; rationals promote to the
512-bit float form past 4096 bits (`ratOverflowBits`); a float result
past big.Float's range is `constant result is not representable`. An
untyped integer result has no width cap, as before.

## Behaviour changes vs master

- Untyped rune arithmetic, shifts and unary operators are exact (values
  change where master wrapped; `-'a'`, `^'a'`, `'a' << 40`, `var f
  float64 = 'a'` were rejected).
- A negative untyped rune to an unsigned type, and any untyped rune out
  of the target's range, is rejected (master wrapped `uint64`).
- A typed integer constant operation whose result does not fit the type,
  and a typed float one whose result is infinite, are rejected (master
  wrapped or gave ±Inf).
- `int64(float32(1<<54 + 1<<30 + 1))` is 18014400656965632.
- Shift overflow of a typed constant uses gc's wording.

Non-constant shift context typing (#6393) is out of scope.

## Upgrade hazard

The value of an already-persisted constant may change at an upgrade: a
package deployed with a wrapped rune constant keeps the old slot until
re-preprocessed, while importers fold from the persisted slot. A
migration of persisted constant slots is a separate decision.
