package main

import (
	"crypto/sha256"
	"time"

	"gno.land/r/aib/ibc/core"
	"gno.land/r/aib/ibc/core/testing"
)

// UpdateClient consensusState.NextValidatorHash and trustedValset mismatch
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
		trustedValset = testing.GenValset()
		trustedHeight = clientState.LatestHeight
		msgHeader     = testing.GenMsgHeader(
			chainID, newTimestamp, apphash[:],
			newHeight, trustedHeight, valset, trustedValset,
		)
	)
	// Alter trustedvalset
	trustedValset.Validators[0].VotingPower = 42
	core.UpdateClient(cross, clientID, msgHeader)
}

// Error:
// trusted validators does not hash to latest trusted validators. Expected: "4e201092fa289bdef1e42bbaafd4efa1a95c98e5c7e8193e24b5511c6b44023c", got: "6429cc1fe212aa6abef2565ac6e9abcfd8171f2a7cdce6cf64a0c321d5ea1145"
