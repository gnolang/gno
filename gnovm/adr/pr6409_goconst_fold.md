# Constant folding with go/constant

## Context

`evalConst` folded a constant operation on a machine, with runtime
arithmetic: an untyped rune operation wrapped at int32 (#6389, go/types
accepts `'\U0010ffff' * '\U0010ffff'` as 1241243320321); typed integer
and float operations wrapped or went to infinity where gc rejects them
(#6392); unary `-` `^` on an untyped rune were rejected; an untyped
integer converted to float32 was rounded twice (#6394).

## Decision

A binary, unary or shift operation whose operands are constants of
primitive kinds is folded by `foldConst` (`constfold.go`) with
`go/constant`, forked as `internal/goconst` (Go 1.26's `value.go`, with
the Go license): operands cross the boundary as `goconst.Value`
(`BigintValue` → `*big.Int`, `BigdecValue` → `*big.Rat` or 512-bit
`*big.Float`, typed integers and floats exactly), the result comes back
under the operand's Gno type. go/constant has no rune kind, so Gno's
kinds are carried beside the value; the rune/int ranking (#6339) is
unchanged. Literals, names and conversions still go through the machine.

The fork exists for two reasons. Its representation choices (rational
until a component reaches 4096 bits, 512-bit float past that, `%s` of a
value in an error) are consensus inputs, and in the stdlib they follow
the toolchain a node is built with. And every operation takes a `Meter`
and reports its work (`goconst.Work`: representation, bit length,
operator, shift count) before running, so no caller reaches the
arithmetic unpaid; the fork knows nothing about gas, `constMeter` prices
the work with the slopes in `machine.go`.

- An untyped result is exact. An untyped rune past int32 is stored as a
  `BigintValue` under `UntypedRuneType`; an in-range one keeps its int32
  in `N`, so persisted bytes of existing constants are unchanged. The
  range is checked when the constant takes a type: `var w int64 = c` is
  1241243320321, `var r = c` is `constant 1241243320321 overflows rune`.
- A typed result is rounded once to its type (floats) or checked against
  it (integers), with gc's wording: `c * 2 (constant 200 of type int8)
  overflows int8`, `big * big (constant 1e+600 of type float64) overflows
  float64`, `k << 63 (constant 27670116110564327424 of type uint)
  overflows uint` (was `constant overflows`).
- Untyped integer and rune constants convert to float32/float64 through
  `constant.Float32Val`/`Float64Val`: one rounding, from the exact value.
- Unary `-` `+` `^` on an untyped rune fold; `'a' << 40` is a rune.
- The result starts as a copy of the left operand and only the kind's
  bytes of `N` are written, as the machine's in-place fold did, so a
  folded comparison keeps its bytes and the app hash.

Runtime arithmetic (`op_binary.go`, `op_unary.go`) is untouched.

## Metering

`foldConst` charges `OpCPUSlopeConstFold` (600) once per operation on the
preprocess meter, then `constMeter` prices each unit of work reported:
word-sized operands nothing; untyped integers per kilobit as the machine
charged (quadratic for `*` `/` `%`, a shift by its count and width);
rationals a base plus linear and quadratic terms in their digits
(`big.Rat` cross-multiplies and normalizes by a gcd; `OpCPUSlopeConstRat*`,
fitted at 2.5x the M-series times up to 4095-bit components; a typed
float is a rational in go/constant); 512-bit floats a fixed term, doubled
for a division; a string comparison per byte. `TestConstFoldGasCoversTime`
measures
every operation at every size the limits allow and requires the charge
to cover the time, scaled by a reference operation so it holds on any
machine. The four addpkg gas pins and the stdlib `EXACT_GAS` pins do not
move, since those packages fold no constant operation under the tx meter.

## Limits

go/types' limits apply, so every operand of a fold is bounded: an
untyped integer constant (literal or result) is at most 512 bits
(`constant shift overflow`, `constant overflow` for a literal) and a
constant shift count at most 1074 (`invalid shift count 1075`). The
deploy type check already enforced both; the VM now agrees (`1 << 10000`
was accepted). Rationals promote to the 512-bit float form past 4096
bits (`ratOverflowBits`); a float result past big.Float's range is
`constant result is not representable`.

## Behaviour changes vs master

- Untyped rune arithmetic, shifts and unary operators are exact (master
  wrapped; `-'a'`, `'a' << 40`, `var f float64 = 'a'` were rejected); an
  untyped rune out of its target's range is rejected (master wrapped).
- A typed integer operation whose result does not fit the type, and a
  typed float one whose result is infinite, are rejected as gc does.
- `int64(float32(1<<54 + 1<<30 + 1))` is 18014400656965632.
- Shift overflow of a typed constant uses gc's wording.
- `string(c)` of an untyped rune past int32 is `"\uFFFD"`, as gc.
- Untyped integers past 512 bits and shift counts past 1074 are rejected
  with go/types' wording (the VM accepted them).
- A non-constant shift whose untyped rune operand is typed from its
  context (`var x int64 = c << s`) shifts the exact value; the contexts
  #6393 lists are unchanged.

## Upgrade hazard

A package deployed with a wrapped rune constant keeps the old slot until
re-preprocessed, while importers fold from it; migrating persisted
constant slots is a separate decision.
