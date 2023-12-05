package disperse

import "std"

func SendGNOT(addresses []std.Address, sendValues []int64) {
	std.AssertOriginCall()        // assert the call was by a user?
	sent := std.GetOrigSend()     // get GNOT value sent with call
	caller := std.GetOrigCaller() // get tx caller

	var totalToSend int64
	for _, val := range sendValues {
		totalToSend += val
	}

	if sent.AmountOf("ugnot") < totalToSend { // check if user sent enough GNOT// possibly not needed if banker breaks anyway?
		panic(ErrNotEnoughGnot)
	}

	// Check if args are good
	if len(addresses) != len(sendValues) {
		panic(ErrNumAddrValMismatch)
	}

	// Get address of Disperse realm
	pkgaddr := std.GetOrigPkgAddr()

	// Get Banker
	banker := std.GetBanker(std.BankerTypeOrigSend)

	for i, val := range sendValues {
		send := std.Coins{{"ugnot", val}}

		receiver := addresses[i]
		if !receiver.IsValid() {
			panic(ErrInvalidAddress)
		}

		// Send GNOT from realm to receiver address
		banker.SendCoins(pkgaddr, receiver, send)
	}

	// Return possible leftover GNOT
	leftover := banker.GetCoins(std.GetOrigPkgAddr()).AmountOf("ugnot")

	if leftover > 0 {
		send := std.Coins{{"ugnot", leftover}}
		banker.SendCoins(pkgaddr, caller, send)
	}
}
