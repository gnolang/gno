package main

import (
	"crypto/sha256"
	"testing"
	"time"

	tmtesting "gno.land/p/aib/ibc/lightclient/tendermint/testing"
	ibctesting "gno.land/p/aib/ibc/testing"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

// SendPacket success
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
	// Register 2 apps
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/ibcapp1"))
	app1 := &ibctesting.App{}
	sourcePortApp1 := core.RegisterApp(cross, app1)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/ibcapp2"))
	app2 := &ibctesting.App{}
	sourcePortApp2 := core.RegisterApp(cross, app2)

	packet := types.MsgSendPacket{
		SourceClient:     clientID,
		TimeoutTimestamp: uint64(time.Now().Add(time.Hour).Unix()),
		Payloads: []types.Payload{
			{
				SourcePort:      sourcePortApp1,
				DestinationPort: "destinationPort",
				Encoding:        "application/json",
				Value:           []byte("{}"),
				Version:         "v1",
			},
			{
				SourcePort:      sourcePortApp2,
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
	println("----------- app1 report")
	println(app1.Report())
	println("----------- app2 report")
	println(app2.Report())

	// Send the packet an other time to assert sequence increasing
	sequence = core.SendPacket(cross, packet)

	println("returned sequence:", sequence)
	println("next sequence:", core.Render("clients/"+clientID+"/next_sequence_send"))
	println("----------- assert render clients/07-tendermint-1/packet_commitments")
	println(core.Render("clients/" + clientID + "/packet_commitments"))
}

// Output:
// next sequence: 1
// returned sequence: 1
// next sequence: 2
// ----------- app1 report
// OnSendPacket (1)
// - sourceClient: 07-tendermint-1
// - destinationClient: counter-party-id
// - sequence: 1
// - payload:
//   - sourcePort: gno.land_r_ibcapp1
//   - destinationPort: destinationPort
//   - version: v1
//   - encoding: application/json
//   - value: {}
//
// OnRecvPacket (0)
// OnTimeoutPacket (0)
// OnAcknowledgementPacket (0)
//
// ----------- app2 report
// OnSendPacket (1)
// - sourceClient: 07-tendermint-1
// - destinationClient: counter-party-id
// - sequence: 1
// - payload:
//   - sourcePort: gno.land_r_ibcapp2
//   - destinationPort: destinationPort
//   - version: v1
//   - encoding: application/json
//   - value: {}
//
// OnRecvPacket (0)
// OnTimeoutPacket (0)
// OnAcknowledgementPacket (0)
//
// returned sequence: 2
// next sequence: 3
// ----------- assert render clients/07-tendermint-1/packet_commitments
// ## Client 07-tendermint-1 packet commitments
// - Sequence 0000001
//   - daf71f20b3e8ae8b47f6cead4acd7a5fcb68ce6253d7d14556ed08c522ea125a
// - Sequence 0000002
//   - daf71f20b3e8ae8b47f6cead4acd7a5fcb68ce6253d7d14556ed08c522ea125a

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
//         "value": "counter-party-id"
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
//         "value": "0801120f30372d74656e6465726d696e742d311a10636f756e7465722d70617274792d696420e2a1d8cc042a3f0a12676e6f2e6c616e645f725f69626361707031120f64657374696e6174696f6e506f72741a02763122106170706c69636174696f6e2f6a736f6e2a027b7d2a3f0a12676e6f2e6c616e645f725f69626361707032120f64657374696e6174696f6e506f72741a02763122106170706c69636174696f6e2f6a736f6e2a027b7d"
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
//         "value": "counter-party-id"
//       },
//       {
//         "key": "packet_sequence",
//         "value": "2"
//       },
//       {
//         "key": "packet_timeout_timestamp",
//         "value": "1234571490"
//       },
//       {
//         "key": "encoded_packet_hex",
//         "value": "0802120f30372d74656e6465726d696e742d311a10636f756e7465722d70617274792d696420e2a1d8cc042a3f0a12676e6f2e6c616e645f725f69626361707031120f64657374696e6174696f6e506f72741a02763122106170706c69636174696f6e2f6a736f6e2a027b7d2a3f0a12676e6f2e6c616e645f725f69626361707032120f64657374696e6174696f6e506f72741a02763122106170706c69636174696f6e2f6a736f6e2a027b7d"
//       }
//     ],
//     "pkg_path": "gno.land/r/aib/ibc/core"
//   }
// ]
