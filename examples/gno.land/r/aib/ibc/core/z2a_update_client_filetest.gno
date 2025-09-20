package main

import (
	"crypto/sha256"
	"time"

	"gno.land/r/aib/ibc/core"
	"gno.land/r/aib/ibc/core/testing"
)

// UpdateClient success
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

	// Update clientID
	var (
		newHeight     = uint64(5)
		newTimestamp  = consensusState.Timestamp.Add(time.Minute)
		trustedValset = valset
		trustedHeight = clientState.LatestHeight
		msgHeader     = testing.GenMsgHeader(
			chainID, newTimestamp, apphash[:],
			newHeight, trustedHeight, valset, trustedValset,
		)
	)
	core.UpdateClient(cross, clientID, msgHeader)

	println("----------- assert render clients")
	println(core.Render("clients"))
}

// Output:
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
//   - LatestHeight: 2/5
// - Latest consensus state:
//   - Timestamp: 2009-02-13 23:32:30 +0000 UTC
//   - Root: 10690182bfe0c0d62af5790f2d2f9459092c46da70ff20802118f2dca8d4948f
//   - NextValidatorsHash: 4e201092fa289bdef1e42bbaafd4efa1a95c98e5c7e8193e24b5511c6b44023c
