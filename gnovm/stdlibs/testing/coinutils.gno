package testing

import "std"

// expandNative expands for usage within natively bound functions.
func expandNative(coins std.Coins) (denoms []string, amounts []int64) {
	denoms = make([]string, len(coins))
	amounts = make([]int64, len(coins))
	for i, coin := range coins {
		denoms[i] = coin.Denom
		amounts[i] = coin.Amount
	}

	return denoms, amounts
}

// compactNative compacts coins for usage within natively bound functions.
func compactNative(denoms []string, amounts []int64) std.Coins {
	coins := make(std.Coins, len(denoms))
	for i := range coins {
		coins[i] = std.Coin{Denom: denoms[i], Amount: amounts[i]}
	}
	return coins
}
