# ADR: Respect the Unicode code point range in integer-to-string conversions

## Status

Proposed (AI-assisted fix; found via differential testing against the Go
toolchain).

## Context

Per the Go spec ("Conversions to and from a string type"), converting an
integer value to a string yields the UTF-8 representation of the code point,
and "values outside the range of valid Unicode code points are converted to
`�`".

GnoVM implemented the conversion for the 64-bit-capable kinds as
`string(rune(tv.GetInt64()))` (and likewise for `int`, `uint`, `uint64`) in
`values_conversions.go`. The `rune(...)` wrapper narrows the value to int32
*before* Go's own `string(rune)` range check runs, so an out-of-range 64-bit
value can alias onto a valid code point instead of yielding `�`:

- `string(uint64(0x10001F600))` returned `"😀"` (Go: `"�"`)
- `string(int(-4294967231))` returned `"A"` (Go: `"�"`)
- `string(uint64(0x100000000))` returned `"\x00"` (Go: `"�"`)

The divergence is deterministic but silently wrong, and it affects both the
runtime path and constant evaluation (both flow through `ConvertTo`). Note
that Go accepts out-of-range *constants* in this conversion too —
`string(0x100000041)` compiles in Go and yields `"�"` — so rejecting at
preprocess time would not match Go either; producing `"�"` is correct in
both paths.

## Decision

Add two helpers, `runeStrFromInt64` and `runeStrFromUint64`, that check
whether the value fits in an int32 before converting, and return
`string(utf8.RuneError)` when it does not. Once a value fits in int32, Go's
native `string(rune)` conversion already maps every invalid case (negative
values, surrogate halves, values above 0x10FFFF) to `�`, so only the
truncation hole needs plugging. The helpers are used at the four
64-bit-capable conversion sites (`IntKind`, `Int64Kind`, `UintKind`,
`Uint64Kind`).

The remaining integer kinds are intentionally unchanged:

- `Int32Kind` converts via `string(tv.GetInt32())` with no narrowing.
- `Int8/Int16/Uint8/Uint16` cannot exceed the int32 range.
- `Uint32Kind` uses `string(rune(tv.GetUint32()))`; values above `MaxInt32`
  reinterpret as negative runes and values in `(0x10FFFF, MaxInt32]` exceed
  the code point range, so both classes already map to `�` — provably
  equivalent to the spec behavior for every uint32 input.

## Alternatives considered

- Range-checking against `0..0x10FFFF` directly in each case arm: equivalent
  behavior, but duplicates the bound logic at four sites and re-implements
  what `string(rune)` already does for the in-int32 cases.
- Rejecting out-of-range constant conversions at preprocess time: would
  diverge from Go, which accepts them and produces `"�"`.

## Consequences

- `string(x)` conversions match Go for all integer kinds and all values, in
  both runtime and constant evaluation paths.
- Programs that (incorrectly) relied on the truncating behavior change
  output; such programs were already non-portable Go.
- Covered by `gnovm/tests/files/str_conv_overflow.gno` (boundary table
  including truncation-aliasing values, in-range values, and in-int32 invalid
  values).
