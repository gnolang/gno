package main

import (
	"crypto/sha256"
	"time"

	"gno.land/p/aib/ibc/lightclient/tendermint/testing"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

// UpdateClient validatorhash and nextValidatorHash mismatch
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
	core.RegisterCounterparty(cross, clientID, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "07-tendermint-2")

	// Update clientID
	var (
		newHeight     = clientState.LatestHeight.RevisionHeight + 1
		newTimestamp  = consensusState.Timestamp.Add(time.Minute)
		trustedValset = testing.GenValset()
		trustedHeight = clientState.LatestHeight
		msgHeader     = testing.GenMsgHeader(
			chainID, newTimestamp, apphash[:],
			newHeight, trustedHeight, valset, trustedValset,
		)
	)
	// Alter ValidateHash
	msgHeader.ValidatorSet.Validators[0].VotingPower++
	msgHeader.Header.ValidatorsHash = msgHeader.ValidatorSet.Hash()

	core.UpdateClient(cross, clientID, msgHeader)
}

// Error:
// expected old header next validators (4e201092fa289bdef1e42bbaafd4efa1a95c98e5c7e8193e24b5511c6b44023c) to match those from new header (f8ba5dcf7674c3c1082b8309f0ef20c8d0d44a7c35481256065ae99ee6140c9e)
