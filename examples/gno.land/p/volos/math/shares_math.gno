package math

import (
	u256 "gno.land/p/gnoswap/uint256"
)

// Virtual amounts for share/asset calculations
// These values are used to prevent share price manipulation when pools are empty or nearly empty
// They create a minimum pool size that exists from the start
var (
	// VIRTUAL_SHARES represents the minimum number of shares that exist in a pool
	// This creates a buffer that makes share price manipulation more expensive
	VIRTUAL_SHARES = u256.NewUint(1000000000) // 1 billion shares

	// VIRTUAL_ASSETS represents the minimum number of assets that exist in a pool
	VIRTUAL_ASSETS = u256.NewUint(1) // 1 base units
)

// ToSharesDown converts assets to shares, rounding down (used for supply)
// This is used when a user supplies assets and we need to calculate how many shares they receive
func ToSharesDown(assets, totalAssets, totalShares *u256.Uint) *u256.Uint {
	totalSharesWithVirtual := new(u256.Uint).Add(totalShares, VIRTUAL_SHARES)
	totalAssetsWithVirtual := new(u256.Uint).Add(totalAssets, VIRTUAL_ASSETS)

	return MulDivDown(
		assets,
		totalSharesWithVirtual,
		totalAssetsWithVirtual,
	)
}

// ToSharesUp converts assets to shares, rounding up (used for borrow)
// This is used when a user borrows assets and we need to calculate how many shares they need
func ToSharesUp(assets, totalAssets, totalShares *u256.Uint) *u256.Uint {
	totalSharesWithVirtual := new(u256.Uint).Add(totalShares, VIRTUAL_SHARES)
	totalAssetsWithVirtual := new(u256.Uint).Add(totalAssets, VIRTUAL_ASSETS)

	return MulDivUp(
		assets,
		totalSharesWithVirtual,
		totalAssetsWithVirtual,
	)
}

// ToAssetsDown converts shares to assets, rounding down (used for withdraw)
// This is used when a user withdraws shares and we need to calculate how many assets they receive
func ToAssetsDown(shares, totalAssets, totalShares *u256.Uint) *u256.Uint {
	totalSharesWithVirtual := new(u256.Uint).Add(totalShares, VIRTUAL_SHARES)
	totalAssetsWithVirtual := new(u256.Uint).Add(totalAssets, VIRTUAL_ASSETS)

	return MulDivDown(
		shares,
		totalAssetsWithVirtual,
		totalSharesWithVirtual,
	)
}

// ToAssetsUp converts shares to assets, rounding up (used for borrow)
// This is used when a user specifies shares to borrow and we need to calculate how many assets they'll receive
func ToAssetsUp(shares, totalAssets, totalShares *u256.Uint) *u256.Uint {
	totalSharesWithVirtual := new(u256.Uint).Add(totalShares, VIRTUAL_SHARES)
	totalAssetsWithVirtual := new(u256.Uint).Add(totalAssets, VIRTUAL_ASSETS)

	return MulDivUp(
		shares,
		totalAssetsWithVirtual,
		totalSharesWithVirtual,
	)
}
