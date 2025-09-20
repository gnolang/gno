package main

import (
	"crypto/sha256"
	"time"

	"gno.land/r/aib/ibc/core"
	"gno.land/r/aib/ibc/core/testing"
)

// CreateClient success
func main() {
	var (
		clientState    = testing.GenClientState("chain-id-2")
		apphash        = sha256.Sum256([]byte("apphash"))
		valset         = testing.GenValset()
		consensusState = testing.GenConsensusState(time.Now(), apphash[:], valset.Hash())
	)

	clientID1 := core.CreateClient(cross, clientState, consensusState)
	println("CreateClient", clientID1)
	clientID2 := core.CreateClient(cross, clientState, consensusState)
	println("CreateClient", clientID2)

	// RegisterCounterparty on clientID1
	core.RegisterCounterparty(cross, clientID1, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "counter-party-id")

	println("\n----------- assert render clients")
	println(core.Render("clients"))
	println("----------- assert render clients/unknown")
	println(core.Render("clients/unknown"))
	println("\n----------- assert render clients/07-tendermint-1")
	println(core.Render("clients/" + clientID1))
}

// Output:
// CreateClient 07-tendermint-1
// CreateClient 07-tendermint-2
//
// ----------- assert render clients
// ## Client 07-tendermint-1
// - Type: 07-tendermint
// - Creator: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// - Status: Active
// - CounterpartyClientID: counter-party-id
// - CounterpartyMerklePrefix: 'prefix1' 'prefix2'
// - Client state:
//   - ChainId: chain-id-2
//   - TrustingPeriod: 1h0m0s
//   - UnbondingPeriod: 3h0m0s
//   - LatestHeight: 2/2
// - Latest consensus state:
//   - Timestamp: 2009-02-13 23:31:30 +0000 UTC
//   - Root: 10690182bfe0c0d62af5790f2d2f9459092c46da70ff20802118f2dca8d4948f
//   - NextValidatorsHash: 4e201092fa289bdef1e42bbaafd4efa1a95c98e5c7e8193e24b5511c6b44023c
// ## Client 07-tendermint-2
// - Type: 07-tendermint
// - Creator: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// - Status: Active
// - CounterpartyClientID:
// - CounterpartyMerklePrefix:
// - Client state:
//   - ChainId: chain-id-2
//   - TrustingPeriod: 1h0m0s
//   - UnbondingPeriod: 3h0m0s
//   - LatestHeight: 2/2
// - Latest consensus state:
//   - Timestamp: 2009-02-13 23:31:30 +0000 UTC
//   - Root: 10690182bfe0c0d62af5790f2d2f9459092c46da70ff20802118f2dca8d4948f
//   - NextValidatorsHash: 4e201092fa289bdef1e42bbaafd4efa1a95c98e5c7e8193e24b5511c6b44023c
//
// ----------- assert render clients/unknown
// client unknown not found
//
// ----------- assert render clients/07-tendermint-1
// ## Client 07-tendermint-1
// - Type: 07-tendermint
// - Creator: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// - Status: Active
// - CounterpartyClientID: counter-party-id
// - CounterpartyMerklePrefix: 'prefix1' 'prefix2'
// - Client state:
//   - ChainId: chain-id-2
//   - TrustingPeriod: 1h0m0s
//   - UnbondingPeriod: 3h0m0s
//   - LatestHeight: 2/2
// - Latest consensus state:
//   - Timestamp: 2009-02-13 23:31:30 +0000 UTC
//   - Root: 10690182bfe0c0d62af5790f2d2f9459092c46da70ff20802118f2dca8d4948f
//   - NextValidatorsHash: 4e201092fa289bdef1e42bbaafd4efa1a95c98e5c7e8193e24b5511c6b44023c
