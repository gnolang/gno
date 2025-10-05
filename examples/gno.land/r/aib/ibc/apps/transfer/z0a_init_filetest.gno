package main

import (
	"crypto/sha256"
	"time"

	tmtesting "gno.land/p/aib/ibc/lightclient/tendermint/testing"
	"gno.land/p/aib/ibc/types"

	"gno.land/r/aib/ibc/apps/transfer"
	"gno.land/r/aib/ibc/core"
)

// init: ensure app is registered
func main() {
	var (
		chainID        = "chain-id-2"
		clientState    = tmtesting.NewClientState(chainID, types.NewHeight(2, 2))
		apphash        = sha256.Sum256([]byte("apphash"))
		valset         = tmtesting.GenValset()
		consensusState = tmtesting.GenConsensusState(time.Now(), apphash[:], valset.Hash())
	)
	clientID := core.CreateClient(cross, clientState, consensusState)
	core.RegisterCounterparty(cross, clientID, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "counter-party-id")

	println("portID:", transfer.PortID)
	println("----------- assert render apps")
	println(core.Render("apps"))

	/*
	   	packet := types.MsgSendPacket{
	   		SourceClient:     clientID,
	   		TimeoutTimestamp: uint64(time.Now().Add(time.Hour).Unix()),
	   		Payloads: []types.Payload{
	   			{
	   				SourcePort:      "gno.land_r_aib_ibc_apps_transfer",
	   				DestinationPort: "destinationPort",
	   				Encoding:        "application/json",
	   				Value:           []byte("{}"),
	   				Version:         "v1",
	   			},
	   		},
	   	}

	   core.SendPacket(cross, packet)
	*/
}

// Output:
// portID: gno.land_r_aib_ibc_apps_transfer
// ----------- assert render apps
// ## Registered apps
// - gno.land_r_aib_ibc_apps_transfer
