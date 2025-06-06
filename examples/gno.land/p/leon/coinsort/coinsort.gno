// Package coinsort provides helpers to sort a slice of std.Coins using the
// classic sort.Sort API (without relying on sort.Slice).
//
// Usage examples:
//
//	coins := banker.GetCoins("g1....")
//
//	// Ascending by balance
//	coinsort.SortByBalance(coins)
//
//	// Custom order – largest balance first
//	coinsort.SortBy(coins, func(a, b std.Coin) bool {
//	    return a.Amount > b.Amount // descending
//	})
//
// Note: when getting std.Coins from the banker, it's sorted by denom by default.
package coinsort

import (
	"sort"
	"std"
)

type ByAmount struct{ std.Coins }

func (b ByAmount) Len() int           { return len(b.Coins) }
func (b ByAmount) Swap(i, j int)      { b.Coins[i], b.Coins[j] = b.Coins[j], b.Coins[i] }
func (b ByAmount) Less(i, j int) bool { return b.Coins[i].Amount < b.Coins[j].Amount }

// SortByBalance sorts c in ascending order by Amount.
//
//	coinsort.SortByBalance(myCoins)
func SortByBalance(c std.Coins) {
	sort.Sort(ByAmount{c})
}

// LessFunc defines the comparison function for SortBy. It must return true if
// 'a' should come before 'b'.

type LessFunc func(a, b std.Coin) bool

// customSorter adapts a LessFunc to sort.Interface so we can keep using
// sort.Sort (rather than sort.Slice).

type customSorter struct {
	coins std.Coins
	less  LessFunc
}

func (cs customSorter) Len() int      { return len(cs.coins) }
func (cs customSorter) Swap(i, j int) { cs.coins[i], cs.coins[j] = cs.coins[j], cs.coins[i] }
func (cs customSorter) Less(i, j int) bool {
	return cs.less(cs.coins[i], cs.coins[j])
}

// SortBy sorts c in place using the provided LessFunc.
//
// Example – descending by Amount:
//
//	coinsort.SortBy(coins, func(a, b std.Coin) bool {
//	    return a.Amount > b.Amount
//	})
func SortBy(c std.Coins, less LessFunc) {
	if less == nil {
		return // nothing to do; keep original order
	}
	sort.Sort(customSorter{coins: c, less: less})
}
