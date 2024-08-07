package disperse

import (
	"std"

	"gno.land/p/demo/ufmt"
	tokens "gno.land/r/demo/grc20factory"
)

// Get address of Disperse realm
var realmAddr = std.CurrentRealm().Addr()

// DisperseUgnot parses receivers and amounts and sends out ugnot
func DisperseUgnot(addresses []std.Address, coins std.Coins) {
	coinSent := std.GetOrigSend()
	caller := std.PrevRealm().Addr()
	banker := std.GetBanker(std.BankerTypeOrigSend)

	// Check if realm balance is equal to amount sent
	if banker.GetCoins(realmAddr).AmountOf("ugnot") != coinSent.AmountOf("ugnot") {
		panic(ErrBalanceNotEqToAmountSent)
	}

	checkCoinAmount(addresses, coins)

	var totalAmount int64
	for _, coin := range coins {
		totalAmount += coin.Amount
	}

	// Check if total amount of coins sent is bigger or equal to total amount of coins to be sent
	if totalAmount > coinSent.AmountOf("ugnot") {
		panic(ErrWrongAmount)
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
func DisperseGRC20(addresses []std.Address, coins std.Coins) {
	caller := std.PrevRealm().Addr() // get tx sender

	checkCoinAmount(addresses, coins)

	// Transfer tokens into the realm
	for _, coin := range coins {
		tokens.TransferFrom(coin.Denom, caller, std.CurrentRealm().Addr(), uint64(coin.Amount))
	}

	// Disperse tokens
	for i := 0; i < len(addresses); i++ {
		tokens.Transfer(coins[i].Denom, addresses[i], uint64(coins[i].Amount))
	}
}

// DisperseGRC20String receives a string of addresses and a string of tokens
// and parses them to be used in DisperseGRC20
func DisperseGRC20String(addresses string, tokens string) {
	parsedAddresses, err := parseAddresses(addresses)
	if err != nil {
		panic(err)
	}

	parseTokens, err := parseTokens(tokens)
	if err != nil {
		panic(err)
	}

	DisperseGRC20(parsedAddresses, parseTokens)
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


// Check if caller correctly sent coins
func checkCoinAmount(addresses []std.Address, coins std.Coins) {
	if len(addresses) != len(coins) {
		panic(ErrNumAddrValMismatch)
	}

	for _, coin := range coins {
		if coin.Amount <= 0 {
			panic(ErrNegativeCoinAmount)
		}
	}
}