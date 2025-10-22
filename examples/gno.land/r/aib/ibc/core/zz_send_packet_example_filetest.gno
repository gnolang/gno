package main

import (
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

func main() {
	packet := types.MsgSendPacket{
		SourceClient:     "",        // TODO update
		TimeoutTimestamp: uint64(0), // TODO update
		Payloads: []types.Payload{
			{
				SourcePort:      "",           // TODO update
				DestinationPort: "",           // TODO update
				Encoding:        "",           // TODO update
				Value:           []byte("{}"), // TODO update
				Version:         "",           // TODO update
			},
		},
	}

	// NOTE: if the target app requires funds like the transfer app, the sent
	// coins must be transfered to the `r/aib/ibc/core` realm before calling
	// `core.SendPacket`. Uncomment the following code to do that.
	//
	// Send the coins sent to this realm to the r/aib/ibc/core realm
	// bank := banker.NewBanker(banker.BankerTypeRealmSend)
	// ibcCoreAddr := chain.PackageAddress("gno.land/r/aib/ibc/core")
	// bank.SendCoins(
	// runtime.CurrentRealm().Address(), ibcCoreAddr, banker.OriginSend(),
	// )

	sequence := core.SendPacket(cross, packet)

	println(sequence)
}

// Error:
// validate client identifier: identifier cannot be blank
