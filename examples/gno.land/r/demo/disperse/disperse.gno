package disperse

import (
	"std"

	"gno.land/p/demo/ufmt"
	tokens "gno.land/r/demo/grc20factory"
)

// Get address of Disperse realm
var realmAddr = std.CurrentRealm().Addr()

// DisperseUgnot parses receivers and amounts and sends out ugnot
// The function will send out the coins to the addresses and return the leftover coins to the caller
// if there are any to return
func DisperseUgnot(addresses []std.Address, coins std.Coins) {
	coinSent := std.GetOrigSend()
	caller := std.PrevRealm().Addr()
	banker := std.GetBanker(std.BankerTypeOrigSend)

	// Check if realm balance is equal to amount sent
	if banker.GetCoins(realmAddr).AmountOf("ugnot") != coinSent.AmountOf("ugnot") {
		panic(ErrBalanceNotEqToAmountSent)
	}

	if len(addresses) != len(coins) {
		panic(ErrNumAddrValMismatch)
	}

	for _, coin := range coins {
		if coin.Amount <= 0 {
			panic(ErrNegativeCoinAmount)
		}
	}

	var totalAmount int64
	for _, coin := range coins {
		totalAmount += coin.Amount
	}

	// Send coins
	for i, _ := range addresses {
		banker.SendCoins(realmAddr, addresses[i], std.NewCoins(coins[i]))
	}

	// Return possible leftover coins
	for _, coin := range coinSent {
		leftoverAmt := banker.GetCoins(realmAddr).AmountOf(coin.Denom)
		if leftoverAmt > 0 {
			send := std.Coins{{coin.Denom, leftoverAmt}}
			banker.SendCoins(realmAddr, caller, send)
		}
	}
}

// DisperseGRC20 disperses tokens to multiple addresses
func DisperseGRC20(addresses []std.Address, amounts []uint64, symbols []string) {
	caller := std.PrevRealm().Addr()

	if (len(addresses) != len(amounts)) || (len(amounts) != len(symbols)) {
		panic(ErrArgLenAndSentLenMismatch)
	}

	for i := 0; i < len(addresses); i++ {
		// Transfer tokens into the realm
		tokens.TransferFrom(symbols[i], caller, std.CurrentRealm().Addr(), amounts[i])
		// Transfer tokens to the addresses
		tokens.Transfer(symbols[i], addresses[i], uint64(amounts[i]))
	}
}

// DisperseGRC20String receives a string of addresses and a string of tokens
// and parses them to be used in DisperseGRC20
func DisperseGRC20String(addresses string, tokens string) {
	parsedAddresses, err := parseAddresses(addresses)
	if err != nil {
		panic(err)
	}

	parsedAmounts, parsedSymbols, err := parseTokens(tokens)
	if err != nil {
		panic(err)
	}

	DisperseGRC20(parsedAddresses, parsedAmounts, parsedSymbols)
}

// DisperseUgnotString receives a string of addresses and a string of amounts
// and parses them to be used in DisperseUgnot
func DisperseUgnotString(addresses string, amounts string) {
	parsedAddresses, err := parseAddresses(addresses)
	if err != nil {
		panic(err)
	}

	parsedAmounts, err := parseAmounts(amounts)
	if err != nil {
		panic(err)
	}

	coins := make(std.Coins, len(parsedAmounts))
	for i, amount := range parsedAmounts {
		coins[i] = std.Coin{"ugnot", amount}
	}

	DisperseUgnot(parsedAddresses, coins)
}
