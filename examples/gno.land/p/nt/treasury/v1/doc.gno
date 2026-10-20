// v1 - Unaudited: This is an initial version that has not yet been formally audited.
// A fully audited version will be published as a subsequent release.
// Use in production at your own risk.
//
// Package treasury provides treasury management for handling coin and GRC20
// token transfers in Gno realms.
//
// v1 reports every Balance and takes every payment amount as a uint256, and
// its GRC20Banker handles gno.land/p/nt/grc20/v1 tokens. Chain coins stay
// int64: CoinsBanker widens them for reporting, and NewCoinsPayment
// range-checks each amount back down to int64, refusing anything outside
// [1, MaxInt64].
package treasury
