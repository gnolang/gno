package main

import (
	"crypto/sha256"
	stdtesting "testing"
	"time"

	"gno.land/r/aib/ibc/core"
	"gno.land/r/aib/ibc/core/testing"
)

// UpdateClient on expired lightclient
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

	// Change block time to expire the lightclient
	ctx := stdtesting.GetContext()
	ctx.Time = time.Now().Add(clientState.TrustingPeriod)
	stdtesting.SetContext(ctx)

	// Update clientID
	var (
		newHeight     = clientState.LatestHeight.RevisionHeight + 3
		newTimestamp  = consensusState.Timestamp.Add(time.Minute)
		trustedValset = valset
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
