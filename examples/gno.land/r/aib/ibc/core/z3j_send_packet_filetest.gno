package main

// PKGPATH: gno.land/r/aib/main
// SEND: 500ugnot

import (
	"crypto/sha256"
	"time"

	tmtesting "gno.land/p/aib/ibc/lightclient/tendermint/testing"
	ibctesting "gno.land/p/aib/ibc/testing"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

// SendPacket success with SEND coins
func main() {
	var (
		chainID        = "chain-id-2"
		clientState    = tmtesting.NewClientState(chainID, types.NewHeight(2, 2))
		apphash        = sha256.Sum256([]byte("apphash"))
		valset         = tmtesting.GenValset()
		consensusState = tmtesting.GenConsensusState(time.Now(), apphash[:], valset.Hash())
	)
	clientID := core.CreateClient(cross, clientState, consensusState)
	core.RegisterCounterparty(cross, clientID, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "07-tendermint-2")

	println("next sequence:", core.Render("clients/"+clientID+"/next_sequence_send"))
	// Register  app
	var (
		app       = &ibctesting.App{}
		appPortID = "app1"
	)
	core.RegisterApp(cross, appPortID, app)

	packet := types.MsgSendPacket{
		SourceClient:     clientID,
		TimeoutTimestamp: uint64(time.Now().Add(time.Hour).Unix()),
		Payloads: []types.Payload{
			{
				SourcePort:      appPortID,
				DestinationPort: "destinationPort",
				Encoding:        "application/json",
				Value:           []byte("{}"),
				Version:         "v1",
			},
		},
	}

	sequence := core.SendPacket(cross, packet)

	println("returned sequence:", sequence)
	println("next sequence:", core.Render("clients/"+clientID+"/next_sequence_send"))
	println("----------- app report")
	println(app.Report())
}

// Output:
// next sequence: 1
// returned sequence: 1
// next sequence: 2
// ----------- app report
// OnSendPacket (1)
// - spendable: 500ugnot
// - sourceClient: 07-tendermint-1
// - destinationClient: 07-tendermint-2
// - sequence: 1
// - payload:
//   - sourcePort: app1
//   - destinationPort: destinationPort
//   - version: v1
//   - encoding: application/json
//   - value: {}
//
// OnRecvPacket (0)
// OnTimeoutPacket (0)
// OnAcknowledgementPacket (0)

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
//         "value": "07-tendermint-2"
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
//         "value": "0801120f30372d74656e6465726d696e742d311a0f30372d74656e6465726d696e742d3220e2a1d8cc042a310a0461707031120f64657374696e6174696f6e506f72741a02763122106170706c69636174696f6e2f6a736f6e2a027b7d"
//       }
//     ],
//     "pkg_path": "gno.land/r/aib/ibc/core"
//   }
// ]
