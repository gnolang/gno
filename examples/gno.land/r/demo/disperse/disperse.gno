package disperse

import (
	"chain"
	"chain/banker"
	"chain/runtime"

	tokens "gno.land/r/demo/defi/grc20factory"
)

var realmAddr = runtime.CurrentRealm().Address()

// DisperseUgnot parses receivers and amounts and sends out ugnot
// The function will send out the coins to the addresses and return the leftover coins to the caller
// if there are any to return
func DisperseUgnot(cur realm, addresses []address, coins chain.Coins) {
	coinSent := banker.OriginSend()
	caller := runtime.PreviousRealm().Address()
	banker_ := banker.NewBanker(banker.BankerTypeRealmSend)

	if len(addresses) != len(coins) {
		panic(ErrNumAddrValMismatch)
	}

	for _, coin := range coins {
		if coin.Amount <= 0 {
			panic(ErrNegativeCoinAmount)
		}

		if banker_.GetCoins(realmAddr).AmountOf(coin.Denom) < coin.Amount {
			panic(ErrMismatchBetweenSentAndParams)
		}
	}

	// Send coins
	for i := range addresses {
		banker_.SendCoins(realmAddr, addresses[i], chain.NewCoins(coins[i]))
	}

	// Return possible leftover coins
	for _, coin := range coinSent {
		leftoverAmt := banker_.GetCoins(realmAddr).AmountOf(coin.Denom)
		if leftoverAmt > 0 {
			send := chain.Coins{chain.NewCoin(coin.Denom, leftoverAmt)}
			banker_.SendCoins(realmAddr, caller, send)
		}
	}
}

// DisperseUgnotString receives a string of addresses and a string of amounts
// and parses them to be used in DisperseUgnot
func DisperseUgnotString(cur realm, addresses string, amounts string) {
	parsedAddresses, err := parseAddresses(addresses)
	if err != nil {
		panic(err)
	}

	parsedAmounts, err := parseAmounts(amounts)
	if err != nil {
		panic(err)
	}

	coins := make(chain.Coins, len(parsedAmounts))
	for i, amount := range parsedAmounts {
		coins[i] = chain.NewCoin("ugnot", amount)
	}

	DisperseUgnot(cur, parsedAddresses, coins)
}

// DisperseGRC20 disperses tokens to multiple addresses
// Note that it is necessary to approve the realm to spend the tokens before calling this function
// see the corresponding filetests for examples
func DisperseGRC20(cur realm, addresses []address, amounts []int64, symbols []string) {
	caller := runtime.PreviousRealm().Address()

	if (len(addresses) != len(amounts)) || (len(amounts) != len(symbols)) {
		panic(ErrArgLenAndSentLenMismatch)
	}
	for _, amount := range amounts {
		if amount < 0 {
			panic(ErrInvalidAmount)
		}
	}

	for i := 0; i < len(addresses); i++ {
		tokens.TransferFrom(cross, symbols[i], caller, addresses[i], amounts[i])
	}
}

// DisperseGRC20String receives a string of addresses and a string of tokens
// and parses them to be used in DisperseGRC20
func DisperseGRC20String(cur realm, addresses string, tokens string) {
	parsedAddresses, err := parseAddresses(addresses)
	if err != nil {
		panic(err)
	}

	parsedAmounts, parsedSymbols, err := parseTokens(tokens)
	if err != nil {
		panic(err)
	}

	DisperseGRC20(cur, parsedAddresses, parsedAmounts, parsedSymbols)
}
