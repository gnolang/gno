// PKGPATH: gno.land/r/aib/main
// SEND: 100uatone
package main

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"crypto/sha256"
	"time"

	tmtesting "gno.land/p/aib/ibc/lightclient/tendermint/testing"
	"gno.land/p/aib/ibc/types"

	"gno.land/r/aib/ibc/apps/transfer"
	"gno.land/r/aib/ibc/core"
)

// OnSendPacket: success with IBC token
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

	bank := banker.NewBanker(banker.BankerTypeRealmIssue)
	// TODO not possible to mint tokens without the realm path prefix
	bank.IssueCoin(runtime.CurrentRealm().Address(), "uatone", 100)

	signer := runtime.OriginCaller()
	payload := transfer.FungibleTokenPacketData{
		Denom:    "transfer/07-tendermint-2/uatone",
		Amount:   "100",
		Sender:   signer.String(),
		Receiver: "atone1user",
	}
	packet := types.MsgSendPacket{
		SourceClient:     clientID,
		TimeoutTimestamp: uint64(time.Now().Add(time.Hour).Unix()),
		Payloads: []types.Payload{
			{
				SourcePort:      transfer.PortID,
				DestinationPort: transfer.PortID,
				Encoding:        transfer.EncodingProtobuf,
				Value:           payload.ProtoMarshal(),
				Version:         transfer.V1,
			},
		},
	}

	// Send the coins sent to this filetest to the r/aib/ibc/core realm
	ibcCoreAddr := chain.PackageAddress("gno.land/r/aib/ibc/core")
	bank.SendCoins(
		runtime.CurrentRealm().Address(), ibcCoreAddr, banker.OriginSend(),
	)
	transferAppAddr := chain.PackageAddress("gno.land/r/aib/ibc/apps/transfer")
	println("r/aib/ibc/core balance before SendPacket:", bank.GetCoins(ibcCoreAddr))
	println("transferAppAddr balance before SendPacket:", bank.GetCoins(transferAppAddr))

	core.SendPacket(cross, packet)

	println("\nr/aib/ibc/core balance after SendPacket:", bank.GetCoins(ibcCoreAddr))
	println("transferAppAddr balance after SendPacket:", bank.GetCoins(transferAppAddr))
}

// Output:
// r/aib/ibc/core balance before SendPacket: 100ugnot
// transferAppAddr balance before SendPacket:
//
// r/aib/ibc/core balance after SendPacket:
// transferAppAddr balance after SendPacket: 100ugnot

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
//         "value": "0801120f30372d74656e6465726d696e742d311a0f30372d74656e6465726d696e742d3220e2a1d8cc042a790a087472616e7366657212087472616e736665721a0769637332302d3122166170706c69636174696f6e2f782d70726f746f6275662a420a0575676e6f7412033130301a28673177796d75343764726872306b7571323039386d3739326c797467746a326e797837377972736d220a61746f6e653175736572"
//       }
//     ],
//     "pkg_path": "gno.land/r/aib/ibc/core"
//   },
//   {
//     "type": "ibc_transfer",
//     "attrs": [
//       {
//         "key": "sender",
//         "value": "g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"
//       },
//       {
//         "key": "receiver",
//         "value": "atone1user"
//       },
//       {
//         "key": "denom",
//         "value": "ugnot"
//       },
//       {
//         "key": "amount",
//         "value": "100"
//       },
//       {
//         "key": "memo",
//         "value": ""
//       }
//     ],
//     "pkg_path": "gno.land/r/aib/ibc/apps/transfer"
//   }
// ]
