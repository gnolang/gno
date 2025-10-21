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

// Acknowledgement success
func main() {
	var (
		chainID        = "chain-id-2"
		trustedHeight  = types.NewHeight(2, 2)
		clientState    = tmtesting.NewClientState(chainID, trustedHeight)
		apphash, _     = hex.DecodeString("8a69ba7fc634fdcd3e5b7356d39f226e6893316f90bd3a75be950c5543e621e5")
		valset         = tmtesting.GenValset()
		consensusState = tmtesting.GenConsensusState(time.Now(), apphash[:], valset.Hash())
		counterpartyID = "07-tendermint-42"
	)
	clientID := core.CreateClient(cross, clientState, consensusState)
	core.RegisterCounterparty(cross, clientID, [][]byte{[]byte("iavlStoreKey"), []byte("prefix2")}, counterpartyID)
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
		// Create the proof (generated from https://github.com/tbruyelle/cosmos-sdk/blob/0b8e35b2962476e550918571daaeef7bf7e3ad40/store/rootmulti/proof_test.go#L63)
		// ROOT 8a69ba7fc634fdcd3e5b7356d39f226e6893316f90bd3a75be950c5543e621e5
		// iavlProof:
		// KEY 7072656669783230372d74656e6465726d696e742d3432030000000000000001 "prefix207-tendermint-42\x03\x00\x00\x00\x00\x00\x00\x00\x01"
		// VAL cf49bb81a77249af41ecbe7792d98ddf24b47b491a177ca5a8b1e82e2eaf011e
		// 0 LEAF PREFIX 000202
		// 	0 OP PREFIX 0204022035f8ea805390e084854f399b42ccdeaea33a1dedc115638ac48d0600637dba1f20
		// 	0 OP SUFFIX
		// 	1 OP PREFIX 04080220
		// 	1 OP SUFFIX 20798e2caa96fdfba376ddeb47999954d2f47e65162264b0536ab5dff7fc0a2e07
		// 	2 OP PREFIX 060c02209af37dd595a0190803b5e05aaef42ae3fad499e4fbe37f7cd31cadff22a9ee7420
		// 	2 OP SUFFIX
		// rootmulti proof:
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
	println("\n----------- assert render clients/07-tendermint-1/packet_commitments")
	println(core.Render("clients/" + clientID + "/packet_commitments"))
	println("\n----------- app report")
	println(app.Report())

	// assert NOOP response (res=1) if ack is resubmitted
	res = core.Acknowledgement(cross, ackPacket)

	println("res:", res)
	println("\n----------- app report")
	println(app.Report())
}

// Output:
// res: (2 gno.land/p/aib/ibc/types.ResponseResultType)
//
// ----------- assert render clients/07-tendermint-1/packet_commitments
// []
//
// ----------- app report
// OnSendPacket (1)
// - spendable: 0ugnot
// - sourceClient: 07-tendermint-1
// - destinationClient: 07-tendermint-42
// - sequence: 1
// - payload:
//   - sourcePort: app
//   - destinationPort: destinationPort
//   - version: v1
//   - encoding: application/json
//   - value: {}
//
// OnRecvPacket (0)
// OnTimeoutPacket (0)
// OnAcknowledgementPacket (1)
// - sourceClient: 07-tendermint-1
// - destinationClient: 07-tendermint-42
// - sequence: 1
// - payload:
//   - sourcePort: app
//   - destinationPort: destinationPort
//   - version: v1
//   - encoding: application/json
//   - value: {}
//
//
// res: (1 gno.land/p/aib/ibc/types.ResponseResultType)
//
// ----------- app report
// OnSendPacket (1)
// - spendable: 0ugnot
// - sourceClient: 07-tendermint-1
// - destinationClient: 07-tendermint-42
// - sequence: 1
// - payload:
//   - sourcePort: app
//   - destinationPort: destinationPort
//   - version: v1
//   - encoding: application/json
//   - value: {}
//
// OnRecvPacket (0)
// OnTimeoutPacket (0)
// OnAcknowledgementPacket (1)
// - sourceClient: 07-tendermint-1
// - destinationClient: 07-tendermint-42
// - sequence: 1
// - payload:
//   - sourcePort: app
//   - destinationPort: destinationPort
//   - version: v1
//   - encoding: application/json
//   - value: {}

// Events:
// [
//   {
//     "type": "create_client",
//     "attrs": [
//       {
//         "key": "client_id",
//         "value": "07-tendermint-1"
//       },
//       {
//         "key": "client_type",
//         "value": "07-tendermint"
//       },
//       {
//         "key": "consensus_heights",
//         "value": "2/2"
//       }
//     ],
//     "pkg_path": "gno.land/r/aib/ibc/core"
//   },
//   {
//     "type": "send_packet",
//     "attrs": [
//       {
//         "key": "packet_source_client",
//         "value": "07-tendermint-1"
//       },
//       {
//         "key": "packet_dest_client",
//         "value": "07-tendermint-42"
//       },
//       {
//         "key": "packet_sequence",
//         "value": "1"
//       },
//       {
//         "key": "packet_timeout_timestamp",
//         "value": "1234571490"
//       },
//       {
//         "key": "encoded_packet_hex",
//         "value": "0801120f30372d74656e6465726d696e742d311a1030372d74656e6465726d696e742d343220e2a1d8cc042a300a03617070120f64657374696e6174696f6e506f72741a02763122106170706c69636174696f6e2f6a736f6e2a027b7d"
//       }
//     ],
//     "pkg_path": "gno.land/r/aib/ibc/core"
//   },
//   {
//     "type": "acknowledge_packet",
//     "attrs": [
//       {
//         "key": "packet_source_client",
//         "value": "07-tendermint-1"
//       },
//       {
//         "key": "packet_dest_client",
//         "value": "07-tendermint-42"
//       },
//       {
//         "key": "packet_sequence",
//         "value": "1"
//       },
//       {
//         "key": "packet_timeout_timestamp",
//         "value": "1234571490"
//       },
//       {
//         "key": "encoded_packet_hex",
//         "value": "0801120f30372d74656e6465726d696e742d311a1030372d74656e6465726d696e742d343220e2a1d8cc042a300a03617070120f64657374696e6174696f6e506f72741a02763122106170706c69636174696f6e2f6a736f6e2a027b7d"
//       }
//     ],
//     "pkg_path": "gno.land/r/aib/ibc/core"
//   }
// ]
