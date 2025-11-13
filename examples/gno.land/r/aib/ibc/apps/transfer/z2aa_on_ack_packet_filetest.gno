// PKGPATH: gno.land/r/aib/main
// SEND: 100ugnot
package main

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"encoding/hex"
	"time"

	tmtesting "gno.land/p/aib/ibc/lightclient/tendermint/testing"
	ibctesting "gno.land/p/aib/ibc/testing"
	"gno.land/p/aib/ibc/types"
	"gno.land/p/aib/ics23"
	"gno.land/r/aib/ibc/apps/transfer"
	"gno.land/r/aib/ibc/core"
)

// OnAcknowledgementPacket: success with refund
func main() {
	var (
		chainID        = "chain-id-2"
		trustedHeight  = types.NewHeight(2, 2)
		clientState    = tmtesting.NewClientState(chainID, trustedHeight)
		apphash, _     = hex.DecodeString("5467ca2d19a7b74dbcd20e33bda91932f912d76595872ec856945af96e743b82")
		valset         = tmtesting.GenValset()
		consensusState = tmtesting.GenConsensusState(time.Now(), apphash[:], valset.Hash())
		counterpartyID = "07-tendermint-42"
	)
	clientID := core.CreateClient(cross, clientState, consensusState)
	core.RegisterCounterparty(cross, clientID, [][]byte{[]byte("iavlStoreKey"), []byte("prefix2")}, counterpartyID)

	signer := runtime.OriginCaller()
	payload := transfer.FungibleTokenPacketData{
		Denom:    "ugnot",
		Amount:   "100",
		Sender:   signer.String(),
		Receiver: "atone1user",
	}
	payloadBz := payload.ProtoMarshal()
	sendPacket := types.MsgSendPacket{
		SourceClient:     clientID,
		TimeoutTimestamp: uint64(time.Now().Add(time.Hour).Unix()),
		Payloads: []types.Payload{
			{
				SourcePort:      transfer.PortID,
				DestinationPort: transfer.PortID,
				Encoding:        transfer.EncodingProtobuf,
				Value:           payloadBz,
				Version:         transfer.V1,
			},
		},
	}

	// Send the coins sent to this filetest to the r/aib/ibc/core realm
	bank := banker.NewBanker(banker.BankerTypeRealmSend)
	ibcCoreAddr := chain.PackageAddress("gno.land/r/aib/ibc/core")
	bank.SendCoins(
		runtime.CurrentRealm().Address(), ibcCoreAddr, banker.OriginSend(),
	)
	transferAppAddr := chain.PackageAddress("gno.land/r/aib/ibc/apps/transfer")
	println("r/aib/ibc/core balance before SendPacket:", bank.GetCoins(ibcCoreAddr))
	println("transferAppAddr balance before SendPacket:", bank.GetCoins(transferAppAddr))
	// Send the packet
	sequence := core.SendPacket(cross, sendPacket)
	println("\nr/aib/ibc/core balance after SendPacket:", bank.GetCoins(ibcCoreAddr))
	println("transferAppAddr balance after SendPacket:", bank.GetCoins(transferAppAddr))

	// Acknowledge the packet
	ackPacket := types.MsgAcknowledgement{
		Packet: types.Packet{
			Sequence:          sequence,
			SourceClient:      clientID,
			DestinationClient: counterpartyID,
			TimeoutTimestamp:  sendPacket.TimeoutTimestamp,
			Payloads: []types.Payload{
				{
					SourcePort:      transfer.PortID,
					DestinationPort: transfer.PortID,
					Encoding:        transfer.EncodingProtobuf,
					Value:           payloadBz,
					Version:         transfer.V1,
				},
			},
		},
		Acknowledgement: types.Acknowledgement{
			// This error should trigger a refund
			AppAcknowledgements: [][]byte{types.UniversalErrorAcknowledgement()},
		},
		// Write the proof of acknowledgement written during the RecvPacket of the
		// counterparty client.
		// (generated from https://github.com/tbruyelle/sdktools/blob/31452ec4783bc45db1f84fe43fde93033e6b5576/proof/proof.go#L56)
		// ROOT 5467ca2d19a7b74dbcd20e33bda91932f912d76595872ec856945af96e743b82
		// KEY 7072656669783230372d74656e6465726d696e742d3432030000000000000001 "prefix207-tendermint-42\x03\x00\x00\x00\x00\x00\x00\x00\x01"
		// VAL e2fb30dfbf7abdeaca82d426534d2b3a9d5444dd2a87fa16d38b77ba1a13ced7
		// 0 LEAF PREFIX 000202
		// 	0 OP PREFIX 0204022035f8ea805390e084854f399b42ccdeaea33a1dedc115638ac48d0600637dba1f20
		// 	0 OP SUFFIX
		// 	1 OP PREFIX 04080220
		// 	1 OP SUFFIX 20798e2caa96fdfba376ddeb47999954d2f47e65162264b0536ab5dff7fc0a2e07
		// 	2 OP PREFIX 060c02209af37dd595a0190803b5e05aaef42ae3fad499e4fbe37f7cd31cadff22a9ee7420
		// 	2 OP SUFFIX
		// KEY 6961766c53746f72654b6579 "iavlStoreKey"
		// VAL 9157a500d6c8de1714badf446d17b6b8d3a58b30f854831866d433536fd2eab9
		// 1 LEAF PREFIX 00
		ProofAcked: []ics23.CommitmentProof_Exist{
			{
				// iavl proof
				Exist: ibctesting.NewExistenceProof(ics23.IavlSpec(),
					[]byte("prefix207-tendermint-42\x03\x00\x00\x00\x00\x00\x00\x00\x01"),
					"e2fb30dfbf7abdeaca82d426534d2b3a9d5444dd2a87fa16d38b77ba1a13ced7",
					"000202",
					[][]string{
						{"0204022035f8ea805390e084854f399b42ccdeaea33a1dedc115638ac48d0600637dba1f20", ""},
						{"04080220", "20798e2caa96fdfba376ddeb47999954d2f47e65162264b0536ab5dff7fc0a2e07"},
						{"060c02209af37dd595a0190803b5e05aaef42ae3fad499e4fbe37f7cd31cadff22a9ee7420", ""},
					},
				),
			},
			{
				// rootmulti proof
				Exist: ibctesting.NewExistenceProof(ics23.IavlSpec(),
					[]byte("iavlStoreKey"), "9157a500d6c8de1714badf446d17b6b8d3a58b30f854831866d433536fd2eab9",
					"00", nil),
			},
		},
		ProofHeight: trustedHeight,
	}

	res := core.Acknowledgement(cross, ackPacket)

	println("\nack res:", res)
	println("transferAppAddr balance after Acknowledgement:", bank.GetCoins(transferAppAddr))
}

// Output:
// r/aib/ibc/core balance before SendPacket: 100ugnot
// transferAppAddr balance before SendPacket:
//
// r/aib/ibc/core balance after SendPacket:
// transferAppAddr balance after SendPacket: 100ugnot
//
// ack res: (2 gno.land/p/aib/ibc/types.ResponseResultType)
// transferAppAddr balance after Acknowledgement: 100ugnot

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
//         "value": "0801120f30372d74656e6465726d696e742d311a1030372d74656e6465726d696e742d343220e2a1d8cc042a790a087472616e7366657212087472616e736665721a0769637332302d3122166170706c69636174696f6e2f782d70726f746f6275662a420a0575676e6f7412033130301a28673177796d75343764726872306b7571323039386d3739326c797467746a326e797837377972736d220a61746f6e653175736572"
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
//         "value": "0801120f30372d74656e6465726d696e742d311a1030372d74656e6465726d696e742d343220e2a1d8cc042a790a087472616e7366657212087472616e736665721a0769637332302d3122166170706c69636174696f6e2f782d70726f746f6275662a420a0575676e6f7412033130301a28673177796d75343764726872306b7571323039386d3739326c797467746a326e797837377972736d220a61746f6e653175736572"
//       }
//     ],
//     "pkg_path": "gno.land/r/aib/ibc/core"
//   },
//   {
//     "type": "fungible_token_packet",
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
//       },
//       {
//         "key": "acknowledgement",
//         "value": "{\"response\":{\"result\":\"BQ==\"}}"
//       }
//     ],
//     "pkg_path": "gno.land/r/aib/ibc/apps/transfer"
//   },
//   {
//     "type": "fungible_token_packet",
//     "attrs": [
//       {
//         "key": "success",
//         "value": "\u0005"
//       }
//     ],
//     "pkg_path": "gno.land/r/aib/ibc/apps/transfer"
//   }
// ]
