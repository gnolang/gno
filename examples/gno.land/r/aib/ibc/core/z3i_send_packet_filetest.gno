package main

import (
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	tmtesting "gno.land/p/aib/ibc/lightclient/tendermint/testing"
	ibctesting "gno.land/p/aib/ibc/testing"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

// SendPacket fails app returns an error
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

	println("next sequence:", core.Render("clients/"+clientID+"/next_sequence_send"))
	// Register app
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/ibcapp"))
	app := &ibctesting.App{}
	app.Calls.OnSendPacketReturn = errors.New("app error")
	sourcePortApp := core.RegisterApp(cross, app)

	packet := types.MsgSendPacket{
		SourceClient:     clientID,
		TimeoutTimestamp: uint64(time.Now().Add(time.Hour).Unix()),
		Payloads: []types.Payload{{
			SourcePort:      sourcePortApp,
			DestinationPort: "destinationPort",
			Encoding:        "application/json",
			Value:           []byte("{}"),
			Version:         "v1",
		}},
	}

	core.SendPacket(cross, packet)
}

// Error:
// send packet failed for app gno.land_r_ibcapp: app error
