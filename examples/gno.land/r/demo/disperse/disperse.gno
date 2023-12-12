package disperse

import (
	"std"
)

// SendCoin parses receivers and amounts and sends out GNOT
// Currently sending slices as arguments with `maketx call` is unsupported,
// which is why we parse strings by delimiter here
func SendCoin(addresses, sendAmounts string) {
	std.AssertOriginCall()        // assert the call was by a user
	coinSent := std.GetOrigSend() // get Coin value & denom sent with call
	caller := std.GetOrigCaller() // get tx caller

	if len(coinSent) != 1 {
		panic(errMultipleCoinsSent)
	}

	// Extract denom from Coins
	denom := coinSent[0].Denom

	// Get address of Disperse realm
	realmAddr := std.GetOrigPkgAddr()

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

	// Check if user sent enough GNOT
	// Possibly not needed, added for error clarity
	if coinSent.AmountOf(denom) < totalToSend {
		panic(errNotEnoughGnot)
	}

	for i, val := range parsedAmounts {
		send := std.Coins{{denom, val}}
		receiver := parsedAddresses[i]

		// Send coin from realm to receiver address
		banker.SendCoins(realmAddr, receiver, send)
	}

	// Return possible leftover GNOT
	leftover := banker.GetCoins(realmAddr).AmountOf(denom)
	if leftover > 0 {
		send := std.Coins{{denom, leftover}}
		banker.SendCoins(realmAddr, caller, send)
	}
}
