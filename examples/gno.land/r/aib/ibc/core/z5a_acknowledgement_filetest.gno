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

// Acknowledgement TODO
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
	ackPacket := types.MsgAcknowledgement{
		Packet: types.Packet{
			Sequence:          sequence,
			SourceClient:      clientID,
			DestinationClient: counterpartyID,
			TimeoutTimestamp:  sendPacket.TimeoutTimestamp,
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
		//	ROOT 7e2766f175b518c898ecd74c51e80bb1d892cd3ba5265af64bb39714417742e7
		//  0 LEAF PREFIX 000202
		//		0 OP PREFIX 0204022035f8ea805390e084854f399b42ccdeaea33a1dedc115638ac48d0600637dba1f20
		//		0 OP SUFFIX
		//		1 OP PREFIX 04080220
		//		1 OP SUFFIX 20798e2caa96fdfba376ddeb47999954d2f47e65162264b0536ab5dff7fc0a2e07
		//	  2 OP PREFIX 060c02209af37dd595a0190803b5e05aaef42ae3fad499e4fbe37f7cd31cadff22a9ee7420
		//	  2 OP SUFFIX
		//  1 LEAF PREFIX 00
		ProofAcked: []ics23.CommitmentProof_Exist{
			{
				// iavl proof
				Exist: ics23.NewExistenceProof(ics23.IavlSpec,
					[]byte("prefix207-tendermint-42"), "cf49bb81a77249af41ecbe7792d98ddf24b47b491a177ca5a8b1e82e2eaf011e",
					"000202",
					[]string{
						"0204022035f8ea805390e084854f399b42ccdeaea33a1dedc115638ac48d0600637dba1f20",
						"04080220",
						"060c02209af37dd595a0190803b5e05aaef42ae3fad499e4fbe37f7cd31cadff22a9ee7420",
					},
					[]string{
						"",
						"20798e2caa96fdfba376ddeb47999954d2f47e65162264b0536ab5dff7fc0a2e07",
						"",
					},
				),
			},
			{
				// rootmulti proof
				Exist: ics23.NewExistenceProof(ics23.IavlSpec,
					[]byte("iavlStoreKey"), "84c09ad209aabb70c74c97d0d2ca94fb826ab18c89b98e88e6abef2af82fe58c",
					"00", nil, nil),
			},
		},
		ProofHeight: trustedHeight,
	}

	res := core.Acknowledgement(cross, ackPacket)

	println("res:", res)
}

// Error:
// TODO
