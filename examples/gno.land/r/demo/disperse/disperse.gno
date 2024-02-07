package disperse

import (
	"std"
)

// SendSingleCoin parses receivers and amounts and sends out a single coin
// Currently sending slices as arguments with `maketx call` is unsupported,
// which is why we parse strings by delimiter here
func SendSingleCoin(addresses string, coins std.Coins) {
	coinSent := std.GetOrigSend() // get Coin value & denom sent with call
	caller := std.GetOrigCaller() // get tx sender

	if len(coinSent) != 1 && len(coins) != 1 {
		panic(errMultipleCoinsSent)
	}

	// Extract denom from Coins
	denom := coinSent[0].Denom

	// Get address of Disperse realm
	realmAddr := std.CurrentRealm().Addr()

	// Get Banker
	banker := std.GetBanker(std.BankerTypeOrigSend)

	parsedAddresses, err := parseAddresses(addresses)
	if err != nil {
		panic(err)
	}

	parsedAmounts, err := parseAmounts(sendAmounts)
	if err != nil {
		panic(err)
	}

	// Check if arg lengths are good
	if len(parsedAddresses) != len(parsedAmounts) {
		panic(errNumAddrValMismatch)
	}

	var totalToSend int64
	for _, val := range parsedAmounts {
		totalToSend += val
	}

	// Check if user sent enough coin
	// Possibly not needed, added for error clarity
	if coinSent.AmountOf(denom) < totalToSend {
		panic(errNotEnoughCoin)
	}

	for i, val := range parsedAmounts {
		send := std.Coins{{denom, val}}
		receiver := parsedAddresses[i]

		// Send coin from realm to receiver address
		banker.SendCoins(realmAddr, receiver, send)
	}

	// Return possible leftover coin
	leftover := banker.GetCoins(realmAddr).AmountOf(denom)
	if leftover > 0 {
		send := std.Coins{{denom, leftover}}
		banker.SendCoins(realmAddr, caller, send)
	}
}
