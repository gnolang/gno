package main

import (
	"time"

	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

// SendPacket fails client not found
func main() {
	packet := types.MsgSendPacket{
		SourceClient:     "clientID",
		TimeoutTimestamp: uint64(time.Now().Unix()),
		Payloads: []types.Payload{{
			SourcePort:      "transfer",
			DestinationPort: "transfer",
			Encoding:        "application/json",
			Value:           []byte("{}"),
			Version:         "v1",
		}},
	}
	core.SendPacket(cross, packet)
}

// Error:
// client clientID not found
