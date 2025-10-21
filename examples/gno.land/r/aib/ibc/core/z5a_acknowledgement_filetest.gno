package main

import (
	"encoding/hex"
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
		apphash, _     = hex.DecodeString("7e2766f175b518c898ecd74c51e80bb1d892cd3ba5265af64bb39714417742e7")
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
		// Generated from cosmos-sdk/store/rootmulti/proof_test.go:
		// ROOT 8a69ba7fc634fdcd3e5b7356d39f226e6893316f90bd3a75be950c5543e621e5
		// KEY 7072656669783230372d74656e6465726d696e742d3432030000000000000001 "prefix207-tendermint-42\x03\x00\x00\x00\x00\x00\x00\x00\x01"
		// VAL cf49bb81a77249af41ecbe7792d98ddf24b47b491a177ca5a8b1e82e2eaf011e
		// 0 LEAF PREFIX 000202
		// 	0 OP PREFIX 0204022035f8ea805390e084854f399b42ccdeaea33a1dedc115638ac48d0600637dba1f20
		// 	0 OP SUFFIX
		// 	1 OP PREFIX 04080220
		// 	1 OP SUFFIX 20798e2caa96fdfba376ddeb47999954d2f47e65162264b0536ab5dff7fc0a2e07
		// 	2 OP PREFIX 060c02209af37dd595a0190803b5e05aaef42ae3fad499e4fbe37f7cd31cadff22a9ee7420
		// 	2 OP SUFFIX
		// KEY 6961766c53746f72654b6579 "iavlStoreKey"
		// VAL fba7f41e4afa674eb64ccc33761c255cfb964147ede2893d7f51eb8fb1e4dea1
		// 1 LEAF PREFIX 00
		ProofAcked: []ics23.CommitmentProof_Exist{
			{
				// iavl proof
				Exist: ics23.NewExistenceProof(ics23.IavlSpec(),
					[]byte("prefix207-tendermint-42\x03\x00\x00\x00\x00\x00\x00\x00\x01"),
					"cf49bb81a77249af41ecbe7792d98ddf24b47b491a177ca5a8b1e82e2eaf011e",
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
				Exist: ics23.NewExistenceProof(ics23.IavlSpec(),
					[]byte("iavlStoreKey"), "fba7f41e4afa674eb64ccc33761c255cfb964147ede2893d7f51eb8fb1e4dea1",
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
