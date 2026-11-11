# Fixed size 256-bit math library

This is a library specialized at replacing the `big.Int` library for math based on 256-bit types.

original repository: [uint256](<https://github.com/holiman/uint256/tree/master>)

> **Prefer the native `uint256` type in new code.** Gno now has native
> `uint256` and `int256` types, with checked arithmetic in `math/overflow`
> (`Addu256`, `Subu256`, ..., and the panicking `...p` variants),
> `strconv.FormatUint256`/`ParseUint256` and `math.MaxUint256`. A native value
> is copied on assignment, cannot be aliased and carries no ownership, and each
> operator is one VM operation, where this library's `*Uint` is a mutable
> object in realm state. This package keeps working for existing code. See
> `gnovm/adr/pr6387_native_int256.md`, and `gno.land/p/nt/grc20/v1` for a token
> built on the native type.
