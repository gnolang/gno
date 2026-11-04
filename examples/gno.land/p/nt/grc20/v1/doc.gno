// Package grc20 implements GRC20 fungible tokens whose amounts are native
// uint256 values.
//
// It has the API and the security model of gno.land/p/nt/grc20/v0
// (Teller, PrivateLedger, CallerTeller/RealmTeller/ReadonlyTeller,
// IsCanonicalTeller), with every amount, balance, allowance and the total
// supply typed uint256 instead of int64. All arithmetic is checked with
// math/overflow: wrapping is the language default for uint256, and a token
// ledger must never wrap. Mint refuses (ErrMintOverflow) any amount that
// would take the total supply past math.MaxUint256.
//
// Chain coins (chain.Coin, chain/banker) stay int64; code that converts
// between them and grc20 amounts must range-check the conversion.
package grc20
