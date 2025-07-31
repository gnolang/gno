package main

import (
	"crypto/sha256"
	"time"

	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/aib/ibc/lightclient/tendermint"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

func main() {
	// CreateClient
	var (
		clientState = tendermint.ClientState{
			ChainId:                      "chain-id-2",
			TrustLevel:                   tendermint.Fraction{2, 3},
			UnbondingPeriod:              time.Hour * 3,
			TrustingPeriod:               time.Hour,
			MaxClockDrift:                time.Hour,
			LatestHeight:                 types.Height{2, 100},
			ProofSpecs:                   types.GetSDKProofSpecs(),
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: false,
		}
		apphash        = sha256.Sum256([]byte("apphash"))
		valhash        = sha256.Sum256([]byte("validators"))
		consensusState = tendermint.ConsensusState{
			Timestamp:          time.Now(),
			Root:               tendermint.MerkleRoot{Hash: apphash[:]},
			NextValidatorsHash: valhash[:],
		}
	)
	id := core.CreateClient(cross, lightclient.Tendermint, clientState, consensusState)
	println(id)
}

// Output:
// 07-tendermint-1
