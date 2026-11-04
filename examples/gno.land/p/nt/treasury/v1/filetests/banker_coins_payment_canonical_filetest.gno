// PKGPATH: gno.land/r/treasury/main

package main

import (
	"chain"
	"math"

	"gno.land/p/nt/treasury/v1"
)

// foreign satisfies Payment without being one of treasury's own impls: a
// BankerID that names a registered banker, and a String() saying whatever
// its realm likes. No Banker will process it, because Banker.Send
// type-asserts on the concrete type.
type foreign struct{}

func (foreign) BankerID() string { return "Coins" }

func (foreign) String() string { return "42ugnot to g1nobody" }

var destAddr = chain.PackageAddress("gno.land/r/dest/main")

func main(cur realm) {
	// A hand-built set with duplicate denoms is joined, so String() renders
	// the amount the bank keeper will actually debit rather than "5ugnot,10ugnot".
	dup := []treasury.Balance{
		{Denom: "ugnot", Amount: 5},
		{Denom: "ugnot", Amount: 10},
	}
	println("duplicate denoms joined:", treasury.NewCoinsPayment(dup, destAddr).String())

	// An unsorted set is sorted, for the same reason: std.Coins.validate
	// rejects "coins not sorted" at execution, after the vote.
	unsorted := []treasury.Balance{
		{Denom: "zzz", Amount: 1},
		{Denom: "aaa", Amount: 2},
	}
	println("unsorted set sorted:", treasury.NewCoinsPayment(unsorted, destAddr).String())

	// A zero amount is refused at construction, not at execution.
	println("zero amount:", mustPanic(func() {
		treasury.NewCoinsPayment([]treasury.Balance{{Denom: "ugnot", Amount: 0}}, destAddr)
	}))

	// Chain coins are int64: the largest payable amount is MaxInt64 ...
	println("max int64:", treasury.NewCoinsPayment(
		[]treasury.Balance{{Denom: "ugnot", Amount: math.MaxInt64}}, destAddr).String())

	// ... and one more is refused rather than truncated.
	println("above max int64:", mustPanic(func() {
		treasury.NewCoinsPayment([]treasury.Balance{{Denom: "ugnot", Amount: math.MaxInt64 + 1}}, destAddr)
	}))

	// Joining can take a sum out of range although each part is in range.
	println("sum above max int64:", mustPanic(func() {
		treasury.NewCoinsPayment([]treasury.Balance{
			{Denom: "ugnot", Amount: math.MaxInt64},
			{Denom: "ugnot", Amount: 1},
		}, destAddr)
	}))

	// A sum that would wrap uint256 is refused too, not wrapped to a small
	// payable amount.
	println("sum wraps uint256:", mustPanic(func() {
		treasury.NewCoinsPayment([]treasury.Balance{
			{Denom: "ugnot", Amount: math.MaxUint256},
			{Denom: "ugnot", Amount: 2},
		}, destAddr)
	}))

	// The canonical impls are recognised; a foreign one is not.
	println("coins payment canonical:", treasury.IsCanonicalPayment(
		treasury.NewCoinsPayment([]treasury.Balance{{Denom: "ugnot", Amount: 1}}, destAddr)))
	println("grc20 payment canonical:", treasury.IsCanonicalPayment(
		treasury.NewGRC20Payment("TOK", 1, destAddr)))
	println("foreign payment canonical:", treasury.IsCanonicalPayment(foreign{}))
}

func mustPanic(fn func()) (msg string) {
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); ok {
				msg = s
				return
			}
			msg = "rejected"
		}
	}()
	fn()
	return "NOT REJECTED"
}

// Output:
// duplicate denoms joined: 15ugnot to g1fs92ep7z5vtrp26d767ag3zztsd6wqn9rew4ve
// unsorted set sorted: 2aaa,1zzz to g1fs92ep7z5vtrp26d767ag3zztsd6wqn9rew4ve
// zero amount: coins payment amounts must be in [1, MaxInt64], got 0ugnot
// max int64: 9223372036854775807ugnot to g1fs92ep7z5vtrp26d767ag3zztsd6wqn9rew4ve
// above max int64: coins payment amounts must be in [1, MaxInt64], got 9223372036854775808ugnot
// sum above max int64: coins payment amounts must be in [1, MaxInt64], got 9223372036854775808ugnot
// sum wraps uint256: coins payment amount overflows uint256 for denom ugnot
// coins payment canonical: true
// grc20 payment canonical: true
// foreign payment canonical: false
