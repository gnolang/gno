// PKGPATH: gno.land/r/treasury/main

package main

import (
	"chain"

	"gno.land/p/nt/treasury/v0"
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
	dup := chain.Coins{
		chain.Coin{Denom: "ugnot", Amount: 5},
		chain.Coin{Denom: "ugnot", Amount: 10},
	}
	println("duplicate denoms joined:", treasury.NewCoinsPayment(dup, destAddr).String())

	// An unsorted set is sorted, for the same reason: std.Coins.validate
	// rejects "coins not sorted" at execution, after the vote.
	unsorted := chain.Coins{
		chain.Coin{Denom: "zzz", Amount: 1},
		chain.Coin{Denom: "aaa", Amount: 2},
	}
	println("unsorted set sorted:", treasury.NewCoinsPayment(unsorted, destAddr).String())

	// A non-positive amount is refused at construction, not at execution.
	println("negative amount:", mustPanic(func() {
		treasury.NewCoinsPayment(chain.Coins{chain.Coin{Denom: "ugnot", Amount: -5}}, destAddr)
	}))

	// Joining can produce a non-positive amount the caller never wrote.
	println("cancels to negative:", mustPanic(func() {
		treasury.NewCoinsPayment(chain.Coins{
			chain.Coin{Denom: "ugnot", Amount: 5},
			chain.Coin{Denom: "ugnot", Amount: -10},
		}, destAddr)
	}))

	// The canonical impls are recognised; a foreign one is not.
	println("coins payment canonical:", treasury.IsCanonicalPayment(
		treasury.NewCoinsPayment(chain.NewCoins(chain.NewCoin("ugnot", 1)), destAddr)))
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
// negative amount: coins payment amounts must be positive, got -5ugnot
// cancels to negative: coins payment amounts must be positive, got -5ugnot
// coins payment canonical: true
// grc20 payment canonical: true
// foreign payment canonical: false
