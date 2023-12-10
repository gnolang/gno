package disperse

import (
	"std"
)

// SendGNOT parses receivers and amounts and sends out GNOT
// Currently sending slices as arguments with `maketx call` is unsupported,
// which is why we parse strings by delimiter here
func SendGNOT(addresses, sendAmounts string) {
	std.AssertOriginCall()         // assert the call was by a user
	coinsSent := std.GetOrigSend() // get GNOT value sent with call
	caller := std.GetOrigCaller()  // get tx caller

	// Get address of Disperse realm
	realmAddr := std.GetOrigPkgAddr()

	// Get Banker
	banker := std.GetBanker(std.BankerTypeOrigSend)

	leftover := banker.GetCoins(realmAddr).AmountOf("ugnot")
	println(leftover)

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
		panic(errNumAddrValMismatch)
	}

	var totalToSend int64
	for _, val := range parsedAmounts {
		totalToSend += val
	}

	// Check if user sent enough GNOT
	// FIXME: possibly not needed, added for error clarity
	if coinsSent.AmountOf("ugnot") < totalToSend {
		panic(errNotEnoughGnot)
	}

	for i, val := range parsedAmounts {
		send := std.Coins{{"ugnot", val}}
		receiver := parsedAddresses[i]

		// Send GNOT from realm to receiver address
		banker.SendCoins(realmAddr, receiver, send)
	}

	// Return possible leftover GNOT
	//
	//if leftover > 0 {
	//	send := std.Coins{{"ugnot", leftover}}
	//	banker.SendCoins(realmAddr, caller, send)
	//}
}
