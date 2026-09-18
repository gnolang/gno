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

Add two helpers, `runeStrFromInt64` and `runeStrFromUint64`, that range-check
the value against the valid Unicode code point range (`v < 0 || v > utf8.MaxRune`,
and just `v > utf8.MaxRune` for the unsigned helper) before converting, and
return `string(utf8.RuneError)` when it is out of range. Values within
`[0, utf8.MaxRune]` are handed to Go's native `string(rune)` conversion, which
already maps the surrogate halves in that range to `�`. The helpers are used
at the four 64-bit-capable conversion sites (`IntKind`, `Int64Kind`,
`UintKind`, `Uint64Kind`).

The remaining integer kinds are intentionally unchanged:

- `Int32Kind` converts via `string(tv.GetInt32())` with no narrowing.
- `Int8/Int16/Uint8/Uint16` cannot exceed the int32 range.
- `Uint32Kind` uses `string(rune(tv.GetUint32()))`; values above `MaxInt32`
  reinterpret as negative runes and values in `(0x10FFFF, MaxInt32]` exceed
  the code point range, so both classes already map to `�` — provably
  equivalent to the spec behavior for every uint32 input.

## Alternatives considered

- Deferring the whole conversion to Go with a bare `string(v)` (no `rune`
  cast), which would reproduce the spec behavior — including the
  out-of-range → `�` mapping — for free. Not possible: `go vet`'s
  `stringintconv` check rejects `string(<non-rune integer>)` ("conversion from
  int64 to string yields a string of one rune, not a string of digits"), it
  runs as part of `go test`, and there is no per-line way to dismiss it. The
  code must therefore cast to `rune` first, which is exactly the int32
  truncation this ADR fixes — hence the explicit range check.
- Inlining the range check directly in each of the four case arms: equivalent
  behavior, but duplicates the bound logic at four sites; the shared helpers
  keep it in one place.
- Rejecting out-of-range constant conversions at preprocess time: would
  diverge from Go, which accepts them and produces `"�"`.

## Consequences

- `string(x)` conversions match Go for all integer kinds and all values, in
  both runtime and constant evaluation paths.
- Programs that (incorrectly) relied on the truncating behavior change
  output; such programs were already non-portable Go.
- Covered by `gnovm/tests/files/convert11.gno` (boundary table including
  truncation-aliasing values, in-range values, and in-int32 invalid values).
