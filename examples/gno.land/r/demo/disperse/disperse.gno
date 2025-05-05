package disperse

import (
	"std"

	tokens "gno.land/r/demo/grc20factory"
)

// DisperseUgnot parses receivers and amounts and sends out ugnot
// The function will send out the coins to the addresses and return the leftover coins to the caller
// if there are any to return
func DisperseUgnot(addresses []std.Address, coins std.Coins) {
	// CurrentRealm is the one that is calling this function since we didn't crossed
	currentAddr := std.CurrentRealm().Address()
	banker := std.NewBanker(std.BankerTypeOriginSend)

	if len(addresses) != len(coins) {
		panic(ErrNumAddrValMismatch)
	}

	for _, coin := range coins {
		if coin.Amount <= 0 {
			panic(ErrNegativeCoinAmount)
		}

		if banker.GetCoins(currentAddr).AmountOf(coin.Denom) < coin.Amount {
			panic(ErrMismatchBetweenSentAndParams)
		}
	}

	// Send coins
	for i := range addresses {
		banker.SendCoins(currentAddr, addresses[i], std.NewCoins(coins[i]))
	}
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

// DisperseGRC20 disperses tokens to multiple addresses
// Note that it is necessary to approve the realm to spend the tokens before calling this function
// see the corresponding filetests for examples
func DisperseGRC20(addresses []std.Address, amounts []uint64, symbols []string) {
	crossing()

	caller := std.PreviousRealm().Address()

	if (len(addresses) != len(amounts)) || (len(amounts) != len(symbols)) {
		panic(ErrArgLenAndSentLenMismatch)
	}

	for i := 0; i < len(addresses); i++ {
		cross(tokens.TransferFrom)(symbols[i], caller, addresses[i], amounts[i])
	}
}

// DisperseGRC20String receives a string of addresses and a string of tokens
// and parses them to be used in DisperseGRC20
func DisperseGRC20String(addresses string, tokens string) {
	crossing()

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
