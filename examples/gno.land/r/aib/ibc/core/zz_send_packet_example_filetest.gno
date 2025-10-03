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

	sequence := core.SendPacket(cross, packet)

	println(sequence)
}

// Error:
// validate client identifier: identifier cannot be blank
