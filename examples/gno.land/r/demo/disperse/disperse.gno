package disperse

import (
	"std"

	tokens "gno.land/r/demo/grc20factory"
)

// DisperseGnot parses receivers and amounts and sends out gnot
func DisperseGnot(addresses []std.Address, coins std.Coins) {
	coinSent := std.GetOrigSend() // get Coins sent with call
	caller := std.PrevRealm().Addr() // get tx sender

	checkCoinAmount(coins)
	checkNumAddrValMatch(addresses, coins)

	// Get address of Disperse realm
	realmAddr := std.CurrentRealm().Addr()

	// Get Banker
	banker := std.GetBanker(std.BankerTypeOrigSend)

	var totalAmount int64
	for _, coin := range coins {
		totalAmount += coin.Amount
	}

	// Check if total amount of coins sent is bigger or equal to total amount of coins to be sent
	if totalAmount > coinSent.AmountOf("ugnot") {
		panic(ErrWrongAmount)
	}

	// Check if the realm has zero ugnot
	if banker.GetCoins(realmAddr).AmountOf("ugnot") < coinSent.AmountOf("ugnot") {
		panic(ErrBalanceNotEqToAmountSent)
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

// DisperseToken disperses tokens to multiple addresses
func DisperseToken(addresses []std.Address, coins std.Coins) {
	caller := std.PrevRealm().Addr() // get tx sender

	checkCoinAmount(coins)
	checkNumAddrValMatch(addresses, coins)

	// Transfer tokens into the realm
	for _, coin := range coins {
		tokens.TransferFrom(coin.Denom, caller, std.CurrentRealm().Addr(), uint64(coin.Amount))
	}

	// Disperse tokens
	for i := 0; i < len(addresses); i++ {
		tokens.Transfer(coins[i].Denom, addresses[i], uint64(coins[i].Amount))
	}
}

// DisperseTokenString receives a string of addresses and a string of tokens
// and parses them to be used in DisperseToken
func DisperseTokenString(addresses string, tokens string) {
	parsedAddresses, err := parseAddresses(addresses)
	if err != nil {
		panic(err)
	}

	parseTokens, err := parseTokens(tokens)
	if err != nil {
		panic(err)
	}

	DisperseToken(parsedAddresses, parseTokens)
}

// DisperseGnotString receives a string of addresses and a string of amounts
// and parses them to be used in DisperseGnot
func DisperseGnotString(addresses string, amounts string) {
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

	DisperseGnot(parsedAddresses, coins)
}

// Check if number of addresses and coins match
func checkNumAddrValMatch(addresses []std.Address, coins std.Coins) {
	if len(addresses) != len(coins) {
		panic(ErrNumAddrValMismatch)
	}
}

// Check if caller correctly sent positive amount of coins
func checkCoinAmount(coins std.Coins) {
	for _, coin := range coins {
		if coin.Amount <= 0 {
			panic(ErrNegativeCoinAmount)
		}
	}
}