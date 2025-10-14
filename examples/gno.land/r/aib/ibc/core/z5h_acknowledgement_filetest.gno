package main

import (
	"crypto/sha256"
	"time"

	tmtesting "gno.land/p/aib/ibc/lightclient/tendermint/testing"
	ibctesting "gno.land/p/aib/ibc/testing"
	"gno.land/p/aib/ibc/types"
	"gno.land/p/aib/ics23"
	"gno.land/r/aib/ibc/core"
)

// Acknowledgement height not found
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
	// Send a packet
	sendPacket := types.MsgSendPacket{
		SourceClient:     clientID,
		TimeoutTimestamp: uint64(time.Now().Add(time.Hour).Unix()),
		Payloads: []types.Payload{{
			SourcePort:      appPortID,
			DestinationPort: "destinationPort",
			Encoding:        "application/json",
			Value:           []byte("{}"),
			Version:         "v1",
		}},
	}
	sequence := core.SendPacket(cross, sendPacket)

	// Acknowledge the packet
	// TODO make send packet commitment bytes and ackPacket.packet commitment bytes match
	ackPacket := types.MsgAcknowledgement{
		Packet: types.Packet{
			Sequence:          sequence,
			SourceClient:      clientID,
			DestinationClient: counterpartyID,
			TimeoutTimestamp:  sendPacket.TimeoutTimestamp,
			Payloads:          sendPacket.Payloads,
		},
		Acknowledgement: types.Acknowledgement{
			AppAcknowledgements: [][]byte{[]byte("ack")},
		},
		ProofAcked:  []ics23.CommitmentProof_Exist{{}},
		ProofHeight: types.NewHeight(2, 1),
	}

	res := core.Acknowledgement(cross, ackPacket)

	println("res:", res)
}

// Error:
// failed packet acknowledgement verification for client (07-tendermint-1): please ensure the proof was constructed against a height that exists on the client
