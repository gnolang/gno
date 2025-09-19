package main

import (
	"crypto/sha256"
	stdtesting "testing"
	"time"

	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
	"gno.land/r/aib/ibc/core/testing"
)

// UpdateClient expired client
func main() {
	// CreateClient and RegisterCounterparty
	var (
		chainID        = "chain-id-2"
		clientState    = testing.GenClientState(chainID)
		apphash        = sha256.Sum256([]byte("apphash"))
		valset         = testing.GenValset()
		consensusState = testing.GenConsensusState(time.Now(), apphash[:], valset.Hash())
	)
	clientID := core.CreateClient(cross, clientState, consensusState)
	core.RegisterCounterparty(cross, clientID, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "counter-party-id")

	// Change OriginCaller for the call to RegisterCounterparty
	ctx := stdtesting.GetContext()
	ctx.Time = time.Now().Add(clientState.TrustingPeriod)
	stdtesting.SetContext(ctx)

	// Update clientID
	var (
		newHeight     = types.Height{2, 5}
		newTimestamp  = consensusState.Timestamp.Add(time.Minute)
		trustedValset = valset // no change in valset between the 2 heights
		trustedHeight = clientState.LatestHeight
		msgHeader     = testing.GenMsgHeader(
			chainID, newTimestamp, apphash[:],
			newHeight, trustedHeight, valset, trustedValset,
		)
	)
	core.UpdateClient(cross, clientID, msgHeader)
}

// Error:
// cannot update client (07-tendermint-1) with status Expired
