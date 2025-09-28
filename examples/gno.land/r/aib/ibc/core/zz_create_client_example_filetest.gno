package main

import (
	"crypto/sha256"
	"time"

	"gno.land/p/aib/ibc/lightclient/tendermint"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

func main() {
	// CreateClient
	var (
		chainID      = "chain-id-2"  // TODO update
		latestHeight = types.Height{ // TODO update
			RevisionNumber: 2,
			RevisionHeight: 100,
		}
		clientState = tendermint.ClientState{
			ChainID:         chainID,
			TrustLevel:      tendermint.Fraction{2, 3},
			UnbondingPeriod: time.Hour * 3, // TODO update
			TrustingPeriod:  time.Hour,     // TODO update
			MaxClockDrift:   time.Hour,     // TODO update
			LatestHeight:    latestHeight,
			ProofSpecs:      types.GetSDKProofSpecs(),
		}
		apphash        = sha256.Sum256([]byte("apphash"))    // TODO update
		valhash        = sha256.Sum256([]byte("validators")) // TODO update
		timestamp      = time.Now()                          // TODO update
		consensusState = tendermint.ConsensusState{
			Timestamp:          timestamp,
			Root:               tendermint.MerkleRoot{Hash: apphash[:]},
			NextValidatorsHash: valhash[:],
		}
	)
	id := core.CreateClient(cross, clientState, consensusState)
	println(id)
}

// Output:
// 07-tendermint-1
