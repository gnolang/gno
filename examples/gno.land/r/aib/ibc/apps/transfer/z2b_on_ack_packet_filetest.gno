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

// OnAcknowledgementPacket: cannot pass custom ack error in IBCv2
func main() {
	var (
		chainID        = "chain-id-2"
		trustedHeight  = types.NewHeight(2, 2)
		clientState    = tmtesting.NewClientState(chainID, trustedHeight)
		apphash, _     = hex.DecodeString("21d81e79829ea9f108bf832e1dd3750ca20b334a49084aef98cfd3c03bf5c7a5")
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
	// Send the packet
	sequence := core.SendPacket(cross, sendPacket)
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
			AppAcknowledgements: [][]byte{[]byte(`{"response":{"error":"BQ=="}}`)},
		},
		// Write the proof of acknowledgement written during the RecvPacket of the
		// counterparty client.
		// (generated from https://github.com/tbruyelle/sdktools/blob/31452ec4783bc45db1f84fe43fde93033e6b5576/proof/proof.go#L56)
		// ROOT 21d81e79829ea9f108bf832e1dd3750ca20b334a49084aef98cfd3c03bf5c7a5
		// KEY 7072656669783230372d74656e6465726d696e742d3432030000000000000001 "prefix207-tendermint-42\x03\x00\x00\x00\x00\x00\x00\x00\x01"
		// VAL 3d45dfb58893fff0edd023657cefcdffc23aa5e62a046fa833e38b30454936b5
		// 0 LEAF PREFIX 000202
		// 	0 OP PREFIX 0204022035f8ea805390e084854f399b42ccdeaea33a1dedc115638ac48d0600637dba1f20
		// 	0 OP SUFFIX
		// 	1 OP PREFIX 04080220
		// 	1 OP SUFFIX 20798e2caa96fdfba376ddeb47999954d2f47e65162264b0536ab5dff7fc0a2e07
		// 	2 OP PREFIX 060c02209af37dd595a0190803b5e05aaef42ae3fad499e4fbe37f7cd31cadff22a9ee7420
		// 	2 OP SUFFIX
		// KEY 6961766c53746f72654b6579 "iavlStoreKey"
		// VAL fc717b1db887c1daf906d93d7870ec1c779a4656e600a2656c0a88bd2a7e6ac4
		// 1 LEAF PREFIX 00
		ProofAcked: []ics23.CommitmentProof_Exist{
			{
				// iavl proof
				Exist: ibctesting.NewExistenceProof(ics23.IavlSpec(),
					[]byte("prefix207-tendermint-42\x03\x00\x00\x00\x00\x00\x00\x00\x01"),
					"3d45dfb58893fff0edd023657cefcdffc23aa5e62a046fa833e38b30454936b5",
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
					[]byte("iavlStoreKey"), "fc717b1db887c1daf906d93d7870ec1c779a4656e600a2656c0a88bd2a7e6ac4",
					"00", nil),
			},
		},
		ProofHeight: trustedHeight,
	}

	core.Acknowledgement(cross, ackPacket)
}

// Error:
// acknowledgement packet failed for app transfer: cannot pass in a custom error acknowledgement with IBC v2
