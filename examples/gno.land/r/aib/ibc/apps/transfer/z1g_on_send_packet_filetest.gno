package main

import (
	"chain"
	"chain/runtime"
	"crypto/sha256"
	"testing"
	"time"

	tmtesting "gno.land/p/aib/ibc/lightclient/tendermint/testing"
	"gno.land/p/aib/ibc/types"

	"gno.land/r/aib/ibc/apps/transfer"
	"gno.land/r/aib/ibc/core"
)

// OnSendPacket: TODO
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
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", 100)))
	testing.SetRealm(testing.NewUserRealm(signer))

	core.SendPacket(cross, packet)
}

// Error:
// send packet failed for app transfer: packet amount and origin send do not match, 100ugnot != 1ugnot
