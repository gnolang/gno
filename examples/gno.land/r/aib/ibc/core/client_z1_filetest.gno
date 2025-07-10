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
	clientID1 := core.CreateClient(lightclient.Tendermint, clientState, consensusState)
	println("CreateClient", clientID1)
	clientID2 := core.CreateClient(lightclient.Tendermint, clientState, consensusState)
	println("CreateClient", clientID2)

	// RegisterCounterparty on clientID1
	core.RegisterCounterparty(cross, clientID1, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "counter-party-id")

	println(core.Render(""))
}

// Output:
// CreateClient 07-tendermint-1
// CreateClient 07-tendermint-2
//
// ## Client 07-tendermint-1
// - Type: 07-tendermint
// - Creator: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// - CounterpartyClientID: counter-party-id
// - CounterpartyMerklePrefix: 'prefix1' 'prefix2'
//
// ## Client 07-tendermint-2
// - Type: 07-tendermint
// - Creator: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// - CounterpartyClientID:
// - CounterpartyMerklePrefix:
