package disperse

import (
	"std"
 
	tokens "gno.land/r/demo/grc20factory"
)

// DisperseGnot parses receivers and amounts and sends out gnot
func DisperseGnot(addresses []std.Address, coins std.Coins) {
	coinSent := std.GetOrigSend() // get Coins sent with call
	caller := std.GetOrigCaller() // get tx sender

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

	// Send coins
	for i, _ := range addresses {
		banker.SendCoins(realmAddr, addresses[i], std.Coins{coins[i]})
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

func DisperseToken(addresses []std.Address, coins std.Coins) {
	caller := std.GetOrigCaller() // get tx sender

	var totalAmount uint64
	for _, coin := range coins {
		totalAmount += uint64(coin.Amount)
	}

	tokens.TransferFrom(coins[0].Denom, caller, std.CurrentRealm().Addr(), totalAmount)
	for i, address := range addresses {
		for y, coin := range coins {
			if y == i {
				tokens.Transfer(coin.Denom, address, uint64(coin.Amount))
			}
		}
	}
}

func DisperseTokenString(symbol string, addresses string, amounts string) {
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
		coins[i] = std.Coin{symbol, amount}
	}

	DisperseToken(parsedAddresses, coins)
}

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
