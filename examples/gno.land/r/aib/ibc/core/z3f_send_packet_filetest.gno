package main

import (
	"crypto/sha256"
	stdtesting "testing"
	"time"

	"gno.land/p/aib/ibc/lightclient/tendermint/testing"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

// SendPacket fails client status Expired
func main() {
	var (
		chainID        = "chain-id-2"
		clientState    = testing.NewClientState(chainID, types.NewHeight(2, 2))
		apphash        = sha256.Sum256([]byte("apphash"))
		valset         = testing.GenValset()
		consensusState = testing.GenConsensusState(time.Now(), apphash[:], valset.Hash())
	)
	clientID := core.CreateClient(cross, clientState, consensusState)
	core.RegisterCounterparty(cross, clientID, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "07-tendermint-2")

	// Change block time to expire the lightclient
	ctx := stdtesting.GetContext()
	ctx.Time = time.Now().Add(clientState.TrustingPeriod)
	stdtesting.SetContext(ctx)

	packet := types.MsgSendPacket{
		SourceClient:     clientID,
		TimeoutTimestamp: uint64(ctx.Time.Add(time.Hour).Unix()),
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
// client (07-tendermint-1) status is Expired
