package main

import (
	"crypto/sha256"
	"time"

	tmtesting "gno.land/p/aib/ibc/lightclient/tendermint/testing"
	ibctesting "gno.land/p/aib/ibc/testing"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

// Acknowledgement no commitment found returns noop (res=1)
func main() {
	var (
		chainID        = "chain-id-2"
		trustedHeight  = types.NewHeight(2, 2)
		clientState    = tmtesting.NewClientState(chainID, trustedHeight)
		apphash        = sha256.Sum256([]byte("apphash"))
		valset         = tmtesting.GenValset()
		consensusState = tmtesting.GenConsensusState(time.Now(), apphash[:], valset.Hash())
		counterpartyID = "07-tendermint-42"
	)
	clientID := core.CreateClient(cross, clientState, consensusState)
	core.RegisterCounterparty(cross, clientID, [][]byte{[]byte("prefix1"), []byte("prefix2")}, counterpartyID)
	// Register app
	var (
		app       = &ibctesting.App{}
		appPortID = "app"
	)
	core.RegisterApp(cross, appPortID, app)

	// Acknowledge the packet
	ackPacket := types.MsgAcknowledgement{
		Packet: types.Packet{
			Sequence:          42,
			SourceClient:      clientID,
			DestinationClient: counterpartyID,
			TimeoutTimestamp:  1,
			Payloads: []types.Payload{{
				SourcePort:      appPortID,
				DestinationPort: "destinationPort",
				Encoding:        "application/json",
				Value:           []byte("{}"),
				Version:         "v1",
			}},
		},
		Acknowledgement: types.Acknowledgement{
			AppAcknowledgements: [][]byte{[]byte("ack")},
		},
		ProofAcked:  []byte("proof"),
		ProofHeight: trustedHeight,
	}

	res := core.Acknowledgement(cross, ackPacket)

	println("res:", res)
}

// Output:
// res: (1 gno.land/p/aib/ibc/types.ResponseResultType)
