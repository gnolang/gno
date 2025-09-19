package main

import (
	"crypto/sha256"
	"time"

	"gno.land/r/aib/ibc/core"
	"gno.land/r/aib/ibc/core/testing"
)

// RegisterCounterparty twice
func main() {
	var (
		clientState    = testing.GenClientState("chain-id-2")
		apphash        = sha256.Sum256([]byte("apphash"))
		valset         = testing.GenValset()
		consensusState = testing.GenConsensusState(time.Now(), apphash[:], valset.Hash())
	)

	clientID := core.CreateClient(cross, clientState, consensusState)

	// RegisterCounterparty on clientID
	core.RegisterCounterparty(cross, clientID, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "counter-party-id")
	core.RegisterCounterparty(cross, clientID, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "counter-party-id")
}

// Error:
// cannot register counterparty once it is already set
