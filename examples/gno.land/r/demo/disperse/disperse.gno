package disperse

import (
	"std"

	tokens "gno.land/r/demo/grc20factory"
)

// Get address of Disperse realm
var realmAddr = std.CurrentRealm().Address()

// DisperseUgnot parses receivers and amounts and sends out ugnot
// The function will send out the coins to the addresses and return the leftover coins to the caller
// if there are any to return
func DisperseUgnot(addresses []std.Address, coins std.Coins) {
	coinSent := std.OriginSend()
	caller := std.PreviousRealm().Address()
	banker := std.NewBanker(std.BankerTypeOriginSend)

	if len(addresses) != len(coins) {
		panic(ErrNumAddrValMismatch)
	}

	for _, coin := range coins {
		if coin.Amount <= 0 {
			panic(ErrNegativeCoinAmount)
		}

		if banker.GetCoins(realmAddr).AmountOf(coin.Denom) < coin.Amount {
			panic(ErrMismatchBetweenSentAndParams)
		}
	}

	// Send coins
	for i, _ := range addresses {
		banker.SendCoins(realmAddr, addresses[i], std.NewCoins(coins[i]))
	}

	// Return possible leftover coins
	for _, coin := range coinSent {
		leftoverAmt := banker.GetCoins(realmAddr).AmountOf(coin.Denom)
		if leftoverAmt > 0 {
			send := std.Coins{std.NewCoin(coin.Denom, leftoverAmt)}
			banker.SendCoins(realmAddr, caller, send)
		}
	}
}

// DisperseGRC20 disperses tokens to multiple addresses
// Note that it is necessary to approve the realm to spend the tokens before calling this function
// see the corresponding filetests for examples
func DisperseGRC20(addresses []std.Address, amounts []uint64, symbols []string) {
	caller := std.PreviousRealm().Address()

	if (len(addresses) != len(amounts)) || (len(amounts) != len(symbols)) {
		panic(ErrArgLenAndSentLenMismatch)
	}

	for i := 0; i < len(addresses); i++ {
		tokens.TransferFrom(symbols[i], caller, addresses[i], amounts[i])
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
		coins[i] = std.NewCoin("ugnot", amount)
	}

	DisperseUgnot(parsedAddresses, coins)
}
