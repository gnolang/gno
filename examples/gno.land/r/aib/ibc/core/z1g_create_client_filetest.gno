package main

import (
	"crypto/sha256"
	stdtesting "testing"
	"time"

	"gno.land/p/aib/ibc/lightclient/tendermint/testing"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

// CreateClient and RegisterCounterparty have different signer
func main() {
	var (
		chainID        = "chain-id-2"
		clientState    = testing.NewClientState(chainID, types.NewHeight(2, 2))
		apphash        = sha256.Sum256([]byte("apphash"))
		valset         = testing.GenValset()
		consensusState = testing.GenConsensusState(time.Now(), apphash[:], valset.Hash())
	)

	clientID := core.CreateClient(cross, clientState, consensusState)

	// Change OriginCaller for the call to RegisterCounterparty
	ctx := stdtesting.GetContext()
	ctx.OriginCaller = "g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsx"
	stdtesting.SetContext(ctx)

	// RegisterCounterparty on clientID
	core.RegisterCounterparty(cross, clientID, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "counter-party-id")
}

// Error:
// expected same signer as CreateClient submittor g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm, got g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsx
