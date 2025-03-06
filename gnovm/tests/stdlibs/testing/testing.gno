package testing

import (
	"time"

	"github.com/gnolang/gno/gnovm/std"
)

type Context struct {
	CurrentRealm     std.Realm
	OriginCaller     std.Address
	OriginPkgAddress std.Address
	OriginSend       std.Coins
	OriginSpend      std.Coins
	ChainID          string
	Height           int64
	Time             time.Time
}

func testIssueCoins(addr string, denom []string, amt []int64)

func SetOriginCaller(origCaller std.Address) {
	ctx := GetContext()
	ctx.OriginCaller = origCaller
	SetContext(ctx)
}

func SetOriginSend(send std.Coins) {
	ctx := GetContext()
	ctx.OriginSend = send
	SetContext(ctx)
}

func SetOriginSpend(spend std.Coins) {
	ctx := GetContext()
	ctx.OriginSpend = spend
	SetContext(ctx)
}

func SetOriginPkgAddress(addr std.Address) {
	ctx := GetContext()
	ctx.OriginPkgAddress = addr
	SetContext(ctx)
}

func SetHeight(height int64) {
	ctx := GetContext()
	ctx.Height = height
	SetContext(ctx)
}

// SetRealm sets the realm for the current frame.
// After calling SetRealm, calling CurrentRealm() in the test function will yield the value of
// rlm, while if a realm function is called, using PrevRealm() will yield rlm.
func SetRealm(rlm std.Realm) {
	ctx := GetContext()
	ctx.CurrentRealm = rlm
	SetContext(ctx)
}

func SkipHeights(count int64) {
	ctx := GetContext()
	toHeight := ctx.Height + count
	toTime := ctx.Time.Add(time.Duration(count*5) * time.Second)
	SetContext(Context{Height: toHeight, Time: toTime})
}

func IssueCoins(addr std.Address, coins std.Coins) {
	denom, amt := expandNative(coins)
	testIssueCoins(addr.String(), denom, amt)
}

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
