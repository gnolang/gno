package main

import (
	"crypto/sha256"
	"errors"
	"time"

	"gno.land/p/aib/ibc/lightclient"
	"gno.land/r/aib/ibc/core"
	"gno.land/r/aib/ibc/core/testing"
)

// validateError implements ClientMessage in order to returns an error when
// Validate() is invoked.
type validateError struct{}

func (validateError) Validate() error {
	return errors.New("validate error")
}

func (validateError) ClientType() string {
	return lightclient.Tendermint
}

// UpdateClient assert
// Assert calls to clientMessage.Validate
// NOTE: We don't want to test Validate() checks in filetest because it has too
// many cases, hence Validate is tested in unittests and we only ensure that
// the Validate() method are properly invoked.
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
	println("clientMessage.Validate():",
		// NOTE: panics in explicit cross realm can only be caught by revive.
		revive(func() {
			core.UpdateClient(cross, clientID, validateError{})
		}))
}

// Output:
// clientMessage.Validate(): validate error
