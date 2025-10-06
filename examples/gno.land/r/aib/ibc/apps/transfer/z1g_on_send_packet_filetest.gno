// PKGPATH: gno.land/r/aib/main
// SEND: 100ugnot
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

// OnSendPacket: success
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

	signer := runtime.OriginCaller()
	payload := transfer.FungibleTokenPacketData{
		Denom:    "ugnot",
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
	bank := banker.NewBanker(banker.BankerTypeRealmSend)
	ibcCoreAddr := chain.PackageAddress("gno.land/r/aib/ibc/core")
	bank.SendCoins(
		runtime.CurrentRealm().Address(), ibcCoreAddr, banker.OriginSend(),
	)
	escrowAddr := transfer.GetEscrowAddress(clientID)
	println("r/aib/ibc/core before balance:", bank.GetCoins(ibcCoreAddr))
	println("escrowAddr before balance:", bank.GetCoins(escrowAddr))

	core.SendPacket(cross, packet)

	println("\nr/aib/ibc/core after balance:", bank.GetCoins(ibcCoreAddr))
	println("escrowAddr after balance:", bank.GetCoins(escrowAddr))
}

// Output:
// r/aib/ibc/core before balance: 100ugnot
// escrowAddr before balance:
//
// r/aib/ibc/core after balance:
// escrowAddr after balance: 100ugnot
