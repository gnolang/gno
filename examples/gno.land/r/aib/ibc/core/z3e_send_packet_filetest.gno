package main

import (
	"crypto/sha256"
	"time"

	"gno.land/p/aib/ibc/lightclient/tendermint/testing"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

// SendPacket fails timeout after max.
func main() {
	var (
		chainID        = "chain-id-2"
		clientState    = testing.NewClientState(chainID, types.NewHeight(2, 2))
		apphash        = sha256.Sum256([]byte("apphash"))
		valset         = testing.GenValset()
		consensusState = testing.GenConsensusState(time.Now(), apphash[:], valset.Hash())
	)
	clientID := core.CreateClient(cross, clientState, consensusState)
	core.RegisterCounterparty(cross, clientID, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "counter-party-id")

	packet := types.MsgSendPacket{
		SourceClient:     clientID,
		TimeoutTimestamp: uint64(time.Now().Add(time.Hour * 64).Unix()),
		Payloads: []types.Payload{{
			SourcePort:      "transfer",
			DestinationPort: "transfer",
			Encoding:        "application/json",
			Value:           []byte("{}"),
			Version:         "v1",
		}},
	}
	core.SendPacket(cross, packet)
}

// Error:
// timeout is after the max allowed timeout, 1234798290 > 1234654290
