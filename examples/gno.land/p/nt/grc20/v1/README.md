# `grc20` v1 - GRC20 tokens with `uint256` amounts

`gno.land/p/nt/grc20/v1` is `gno.land/p/nt/grc20/v0` with amounts typed as
the native `uint256` instead of `int64`. Everything else (the `Teller`
interface, `PrivateLedger`, the teller constructors, `IsCanonicalTeller`,
token identity and events) works the same way, so read v0's documentation and
the GRC20 section of `docs/resources/gno-security-guide.md` for the model.

v0 stays as it is for the tokens already built on it. New tokens, and tokens
that bridge 256-bit amounts (e.g. from Ethereum), should use v1. The two
versions are separate types: a v1 `*Token` is not a v0 `*Token`, and each has
its own registry (`r/nt/grc20reg/v0`, `r/nt/grc20reg/v1`).

## Differences from v0

| | v0 | v1 |
|---|---|---|
| Amounts, balances, allowances, total supply | `int64` | `uint256` |
| Supply cap (`ErrMintOverflow`) | `math.MaxInt64` | `math.MaxUint256` |
| Negative amounts | rejected with `ErrInvalidAmount` | not representable; `ErrInvalidAmount` is gone |
| Arithmetic | `overflow.Add64p`/`Sub64p` | `overflow.Addu256`/`Subu256` (checked, never wraps) |
| Zero balances and allowances | a zero mint, transfer or approval stores `0` | never stored: absent means zero (`Approve(..., 0)` removes the entry) |
| `KnownAccounts()` | can count a zero balance left by a zero mint or transfer | accounts with a non-zero balance |
| Event `value` | decimal via `strconv.Itoa` | decimal via `strconv.FormatUint256` |
| Amount formatting | none | `FormatAmount(amount, decimals)`, `Token.FormatAmount` |

## Calling a v1 token

`maketx call` arguments for a `uint256` parameter are canonical decimal
strings (no sign, no leading zeros, at most 78 digits), so
`-args 1000000` works the same as for an `int64` parameter, and values above
2^63 are accepted. Realm code passes `uint256` values; converting from an
`int64` is `uint256(x)` after checking `x >= 0`, and converting down needs an
explicit range check (`x <= math.MaxInt64`).

Chain coins stay `int64`: a realm that wraps coins into a v1 token (such as
`r/gnoland/wugnot`) widens on deposit and range-checks on withdrawal.
