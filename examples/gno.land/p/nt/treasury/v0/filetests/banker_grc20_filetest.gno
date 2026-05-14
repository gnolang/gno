// PKGPATH: gno.land/r/treasury/main

package main

import (
	"chain"
	"strings"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/treasury/v0"
)

const amount = int64(1000)

func createToken(name string, toMint address) *grc20.Token {
	// Create the token.
	symbol := strings.ToUpper(name)
	token, ledger := grc20.NewToken(name, symbol, 0)

	// Mint the requested amount.
	ledger.Mint(toMint, amount)

	return token
}

func main() {
	// Define addresses for the sender (owner) and destination.
	ownerAddr := chain.PackageAddress("gno.land/r/treasury/main")
	destAddr := chain.PackageAddress("gno.land/r/dest/main")

	// Try to create a GRC20Banker using a nil lister.
	gbanker, err := treasury.NewGRC20Banker(nil)
	if err == nil {
		panic("expected error when creating GRC20Banker with nil lister")
	}

	// Define a list of token and the associated lister.
	tokens := []*grc20.Token{
		createToken("TestToken0", ownerAddr),
		createToken("TestToken1", ownerAddr),
		createToken("TestToken2", ownerAddr),
	}

	grc20Lister := func() map[string]*grc20.Token {
		tokensMap := make(map[string]*grc20.Token, len(tokens))

		for _, token := range tokens {
			tokensMap[token.GetSymbol()] = token
		}

		return tokensMap
	}

	// Create a GRC20Banker.
	gbanker, err = treasury.NewGRC20Banker(grc20Lister)
	if err != nil {
		panic("failed to create GRC20Banker: " + err.Error())
	}

	println("GRC20Banker ID:", gbanker.ID())
	println("GRC20Banker Address:", gbanker.Address())

	// Check if the GRC20Banker address matches the owner address.
	if gbanker.Address() != ownerAddr.String() {
		panic("GRC20Banker address does not match current realm address")
	}

	// Check the balances of the GRC20Banker.
	println("GRC20Banker Balances count:", len(gbanker.Balances()))
	for _, balance := range gbanker.Balances() {
		if balance.Amount != amount {
			panic("GRC20Banker balance does not match expected amount")
		}
	}

	// Send a valid payment.
	token := tokens[len(tokens)-1]
	validPayment := treasury.NewGRC20Payment(
		token.GetSymbol(),
		100,
		destAddr,
	)
	err = gbanker.Send(validPayment)
	println("Valid payment error:", err)
	if err != nil {
		panic("failed to send valid payment: " + err.Error())
	}

	println("Owner balance:", token.BalanceOf(ownerAddr))
	println("Dest balance:", token.BalanceOf(destAddr))

	// Send an unknown token payment.
	unknownPayment := treasury.NewGRC20Payment(
		"unknown",
		100,
		destAddr,
	)
	err = gbanker.Send(unknownPayment)
	println("Unknown token payment error:", err)
	if err == nil {
		panic("expected error for unknown token, but got none")
	}

	// Send an unsufficient funds payment.
	unsufficientPayment := treasury.NewGRC20Payment(
		tokens[0].GetSymbol(),
		amount+1,
		destAddr,
	)
	err = gbanker.Send(unsufficientPayment)
	println("Unsufficient funds payment error:", err)
	if err == nil {
		panic("expected error for insufficient funds, but got none")
	}

	// Send a payment with an invalid type.
	invalidPaymentType := treasury.NewCoinsPayment(chain.Coins{}, destAddr)
	err = gbanker.Send(invalidPaymentType)
	println("Invalid payment type error:", err)
	if err == nil {
		panic("expected error for invalid payment type, but got none")
	}
}

// Output:
// GRC20Banker ID: GRC20
// GRC20Banker Address: g1ynsdz5zaxhn9gnqtr6t40m5k4fueeutq7xy224
// GRC20Banker Balances count: 3
// Valid payment error: undefined
// Owner balance: 900
// Dest balance: 100
// Unknown token payment error: GRC20 token not found: unknown
// Unsufficient funds payment error: insufficient balance
// Invalid payment type error: invalid payment type
