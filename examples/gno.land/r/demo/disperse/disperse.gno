package disperse

import "std"

// SendGNOT parses receivers and amounts and sends out GNOT
// Currently sending slices as arguments with `maketx call` is unsupported
func SendGNOT(addresses string, sendAmounts string) {
	std.AssertOriginCall()         // assert the call was by a user?
	coinsSent := std.GetOrigSend() // get GNOT value sent with call
	caller := std.GetOrigCaller()  // get tx caller

	parsedAddresses, err := parseAddresses(addresses)
	if err != nil {
		panic(err)
	}

	parsedAmounts, err := parseAmounts(sendAmounts)
	if err != nil {
		panic(err)
	}

	// Check if args are good
	if len(parsedAddresses) != len(parsedAmounts) {
		panic(ErrNumAddrValMismatch)
	}

	var totalToSend int64
	for _, val := range parsedAmounts {
		totalToSend += val
	}

	if coinsSent.AmountOf("ugnot") < totalToSend { // check if user sent enough GNOT// possibly not needed if banker breaks anyway?
		panic(ErrNotEnoughGnot)
	}

	// Get address of Disperse realm
	realmAddr := std.GetOrigPkgAddr()

	// Get Banker
	banker := std.GetBanker(std.BankerTypeOrigSend)

	for i, val := range parsedAmounts {
		send := std.Coins{{"ugnot", val}}
		receiver := parsedAddresses[i]

		// Send GNOT from realm to receiver address
		banker.SendCoins(realmAddr, receiver, send)
	}

	// Return possible leftover GNOT
	leftover := banker.GetCoins(std.GetOrigPkgAddr()).AmountOf("ugnot")

	if leftover > 0 {
		send := std.Coins{{"ugnot", leftover}}
		banker.SendCoins(realmAddr, caller, send)
	}
}
