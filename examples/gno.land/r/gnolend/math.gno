package gnolend

// Virtual amounts for share/asset calculations
// These values are used to prevent share price manipulation when pools are empty or nearly empty
// They create a minimum pool size that exists from the start
// Higher values make manipulation more expensive but also require more gas for calculations
const (
	// VIRTUAL_SHARES represents the minimum number of shares that exist in a pool
	// This creates a buffer that makes share price manipulation more expensive
	VIRTUAL_SHARES int64 = 1000000000 // 1 billion shares

	// VIRTUAL_ASSETS represents the minimum number of assets that exist in a pool
	// This matches VIRTUAL_SHARES to maintain 1:1 initial share ratio
	VIRTUAL_ASSETS int64 = 1000000000 // 1 billion base units
)

// MulDivDown performs multiplication and division while rounding down
func MulDivDown(x, y, denominator int64) int64 {
	if denominator == 0 {
		panic(ErrDivisionByZero)
	}
	return (x * y) / denominator
}

// MulDivUp performs multiplication and division while rounding up
func MulDivUp(x, y, denominator int64) int64 {
	if denominator == 0 {
		panic(ErrDivisionByZero)
	}
	result := (x * y) / denominator
	if (x*y)%denominator > 0 {
		result++
	}
	return result
}

// ToSharesDown converts assets to shares, rounding down (used for supply)
// This is used when a user supplies assets and we need to calculate how many shares they receive
// We round down to ensure the protocol cannot be exploited
func ToSharesDown(assets, totalAssets, totalShares int64) int64 {
	return MulDivDown(
		assets,
		totalShares+VIRTUAL_SHARES,
		totalAssets+VIRTUAL_ASSETS,
	)
}

// ToSharesUp converts assets to shares, rounding up (used for borrow)
// This is used when a user borrows assets and we need to calculate how many shares they need
// We round up to ensure the protocol cannot be exploited
func ToSharesUp(assets, totalAssets, totalShares int64) int64 {
	return MulDivUp(
		assets,
		totalShares+VIRTUAL_SHARES,
		totalAssets+VIRTUAL_ASSETS,
	)
}

// ToAssetsDown converts shares to assets, rounding down (used for withdraw)
// This is used when a user withdraws shares and we need to calculate how many assets they receive
// We round down to ensure the protocol cannot be exploited
func ToAssetsDown(shares, totalAssets, totalShares int64) int64 {
	return MulDivDown(
		shares,
		totalAssets+VIRTUAL_ASSETS,
		totalShares+VIRTUAL_SHARES,
	)
}

// ToAssetsUp converts shares to assets, rounding up (used for borrow)
// This is used when a user specifies shares to borrow and we need to calculate how many assets they'll receive
// We round up to ensure the protocol cannot be exploited
func ToAssetsUp(shares, totalAssets, totalShares int64) int64 {
	return MulDivUp(
		shares,
		totalAssets+VIRTUAL_ASSETS,
		totalShares+VIRTUAL_SHARES,
	)
}
