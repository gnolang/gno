package main

import (
	"crypto/sha256"
	"time"

	"gno.land/p/aib/ibc/lightclient/tendermint/testing"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

// UpdateClient revision number mismatch
func main() {
	// CreateClient and RegisterCounterparty
	var (
		chainID        = "chain-id-2"
		clientState    = testing.NewClientState(chainID, types.NewHeight(2, 2))
		apphash        = sha256.Sum256([]byte("apphash"))
		valset         = testing.GenValset()
		consensusState = testing.GenConsensusState(time.Now(), apphash[:], valset.Hash())
	)
	clientID := core.CreateClient(cross, clientState, consensusState)
	core.RegisterCounterparty(cross, clientID, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "counter-party-id")

	// Update clientID
	var (
		newHeight     = clientState.LatestHeight.RevisionHeight + 1
		newTimestamp  = consensusState.Timestamp.Add(time.Minute)
		trustedValset = valset
		trustedHeight = clientState.LatestHeight
		msgHeader     = testing.GenMsgHeader(
			chainID, newTimestamp, apphash[:],
			newHeight, trustedHeight, valset, trustedValset,
		)
	)
	// Alter chainID to change the revision number
	msgHeader.Header.ChainID = "chain-id-3"

	core.UpdateClient(cross, clientID, msgHeader)
}

// Error:
// header height revision 3 does not match trusted header revision 2
