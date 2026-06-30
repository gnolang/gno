# PR: BigdecValue — Replace apd.Decimal with big.Rat

## Context

Addresses [#5862](https://github.com/gnolang/gno/issues/5862).

`BigdecValue` is the runtime representation of `UntypedBigdecType`, the type
Gno uses for untyped floating-point constant expressions (e.g. `1.0/3.0`).

It previously used `github.com/cockroachdb/apd/v3` — an arbitrary-precision
decimal library. This caused a correctness divergence from Go:

```go
const c = (1.0 / 3.0) * 3.0
println(c == 1.0) // Go: true — Gno: false (was)
```

`apd` stores `1/3` as `0.3333…` (non-terminating decimal at 1024-digit
precision). Multiplying back by `3` yields `0.9999…`, not `1`. Go's compiler
uses `math/big.Rat` — exact rational arithmetic — where `1/3` is stored as
the fraction `{1, 3}`, and `{1,3} × 3 = {3,3} = 1` exactly.

The Go specification (§ Constants) requires that untyped constants are
represented with arbitrary precision using rational arithmetic. Using
`apd.Decimal` was therefore a spec violation.

## Decision

Replace `BigdecValue.V *apd.Decimal` with `BigdecValue.V *big.Rat` throughout
`gnovm/pkg/gnolang`.

### Why big.Rat and not apd

| Property | `apd.Decimal` | `math/big.Rat` |
|---|---|---|
| `1/3 * 3 == 1` | ✗ (0.999…) | ✓ (exact) |
| Matches Go spec | ✗ | ✓ |
| Standard library | ✗ (external dep) | ✓ |
| Transcendentals | ✓ (not needed) | ✗ (not needed) |
| Memory growth | bounded at 1024 digits | unbounded (GCD-reduced) |

`apd` is correct for financial decimal arithmetic (where `0.1 + 0.2 = 0.3`
exactly). It is wrong for constant-expression semantics, which require rational
arithmetic. `big.Rat`'s denominators can grow with repeated operations, but
a 4096-bit denominator guard is added to reject pathological inputs with a
clear error (matching Go's implementation-defined limit ≥ 256 bits per spec).

### Literal parsing

`big.Rat.SetString` natively accepts both decimal float literals (`"1.5"`,
`"3.14"`, `"1e-3"`) and hex float literals (`"0x1.8p+1"`). The previous
manual hex-float parsing block in `op_eval.go` (≈107 lines of code) is replaced
by a single `r.SetString(x.Value)` call, eliminating significant complexity
and regex-based parsing.

### Denominator guard

After every arithmetic operation on `UntypedBigdecType`, a guard function
`ratGuard` checks `r.Denom().BitLen() > 4096`. If triggered it panics with:

```
constant expression result too large: denominator exceeds 4096 bits
```

This matches Go's observable behavior of rejecting overly large constant
expressions (Go's floor is 256 bits; 4096 is well above it and consistent
with the previous 1024-digit / ~3400-bit apd precision). The guard is applied
after all arithmetic operations in `op_binary.go`: addition, subtraction,
multiplication, and division.

### Serialization

`MarshalAmino` now emits `rat.RatString()` (e.g. `"1/3"`) instead of a
decimal string. `UnmarshalAmino` parses via `big.Rat.SetString`. `BigdecValue`
is an untyped constant type and cannot appear in realm state, so there is no
on-chain migration concern.

### Display

`big.Rat.FloatString(10)` is used for human-readable output (e.g. in error
messages and `println`), giving 10 decimal places of precision. Trailing zeros
are trimmed for cleaner output.

## Key files

| File | Role |
|------|------|
| `gnovm/pkg/gnolang/values.go` | `BigdecValue` struct, `MarshalAmino`/`UnmarshalAmino`, `Copy` |
| `gnovm/pkg/gnolang/op_eval.go` | Float literal → `big.Rat` construction via `SetString` |
| `gnovm/pkg/gnolang/op_binary.go` | Arithmetic ops (`Add`, `Sub`, `Mul`, `Div`) with `ratGuard` |
| `gnovm/pkg/gnolang/values_conversions.go` | `ConvertUntypedBigdecTo`, integer conversion helpers, error messages using `RatString()` |
| `gnovm/pkg/gnolang/bounded_strings.go` | Display rendering using `FloatString(10)` with trimming |
| `gnovm/tests/files/types/bigdec*.gno` | Filetests with updated error message format |

## Consequences

- `(1.0/3.0)*3.0 == 1.0` is now `true` in Gno, matching Go.
- All untyped float constant arithmetic is exact for rational inputs.
- Constant expressions with denominator > 4096 bits are rejected with a clear error.
- `cockroachdb/apd` is no longer imported by `gnovm/pkg/gnolang` (it remains
  in `go.mod` — used by `gno.land/pkg/sdk/vm/convert.go`).
- Amino serialization format for `BigdecValue` changes from decimal string to
  ratio string (`"1/3"` not `"0.3333333333"`). This is safe: `BigdecValue` is
  never persisted to chain state, only used in constant evaluation.
- Error messages in const conversions now use rational form (e.g. `"6/5"`) for
  values that cannot be exactly represented as decimals, improving clarity
  on why a conversion failed.
- Code complexity in `op_eval.go` is reduced by ~70 lines of hex-float parsing
  logic, now handled natively by `big.Rat.SetString`.
