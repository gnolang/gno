// PKGPATH: gno.land/r/treasury/main

package main

import (
	"chain"
	"chain/banker"
	"testing"

	"gno.land/p/nt/treasury/v0"
)

func main() {
	// Define addresses for the sender (owner) and destination.
	ownerAddr := chain.PackageAddress("gno.land/r/treasury/main")
	destAddr := chain.PackageAddress("gno.land/r/dest/main")

	// Create a CoinsBanker.
	banker_ := banker.NewBanker(banker.BankerTypeRealmSend)
	cbanker, err := treasury.NewCoinsBanker(banker_)
	if err != nil {
		panic("failed to create CoinsBanker: " + err.Error())
	}

	println("CoinsBanker ID:", cbanker.ID())
	println("CoinsBanker Address:", cbanker.Address())

	// Check if the CoinsBanker address matches the owner address.
	if cbanker.Address() != ownerAddr.String() {
		panic("CoinsBanker address does not match current realm address")
	}

	println("CoinsBanker Balances count:", len(cbanker.Balances()))

	// Issue some coins to the owner address.
	testing.IssueCoins(ownerAddr, chain.NewCoins(chain.NewCoin("ugnot", 42)))

	println("CoinsBanker Balances count:", len(cbanker.Balances()))
	println("Ugnot balance:", cbanker.Balances()[0].Amount)

	// Send a valid payment.
	validPayment := treasury.NewCoinsPayment(
		chain.NewCoins(chain.NewCoin("ugnot", 10)),
		destAddr,
	)
	err = cbanker.Send(validPayment)
	println("Valid payment error:", err)
	if err != nil {
		panic("failed to send valid payment: " + err.Error())
	}

	println("Ugnot balance:", cbanker.Balances()[0].Amount)

	// Send a payment with an invalid type.
	invalidPaymentType := treasury.NewGRC20Payment("", 0, destAddr)
	err = cbanker.Send(invalidPaymentType)
	println("Invalid payment type error:", err)
	if err == nil {
		panic("expected error for invalid payment type, but got none")
	}

	// Issue another coin to the owner address to test the Balances method.
	testing.IssueCoins(ownerAddr, chain.NewCoins(chain.NewCoin("anothercoin", 1337)))

	println("CoinsBanker Balances count:", len(cbanker.Balances()))
}

// Output:
// CoinsBanker ID: Coins
// CoinsBanker Address: g1ynsdz5zaxhn9gnqtr6t40m5k4fueeutq7xy224
// CoinsBanker Balances count: 0
// CoinsBanker Balances count: 1
// Ugnot balance: 42
// Valid payment error: undefined
// Ugnot balance: 32
// Invalid payment type error: invalid payment type
// CoinsBanker Balances count: 2
