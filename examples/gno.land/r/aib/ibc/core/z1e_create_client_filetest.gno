package main

import (
	"crypto/sha256"
	"time"

	"gno.land/r/aib/ibc/core"
	"gno.land/r/aib/ibc/core/testing"
)

// CreateClient with Status Expired
func main() {
	var (
		clientState    = testing.GenClientState("chain-id-2")
		apphash        = sha256.Sum256([]byte("apphash"))
		valset         = testing.GenValset()
		consensusState = testing.GenConsensusState(time.Now(), apphash[:], valset.Hash())
	)
	// consensusState.Timestamp is older than now + clientState.TrustingPeriod
	consensusState.Timestamp = time.Now().Add(-clientState.TrustingPeriod)

	core.CreateClient(cross, clientState, consensusState)
}

// Error:
// cannot create client (07-tendermint-1) with status Expired
