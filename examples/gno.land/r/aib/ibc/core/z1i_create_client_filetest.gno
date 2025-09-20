package main

import (
	"crypto/sha256"
	"errors"
	"time"

	"gno.land/p/aib/ibc/lightclient"
	"gno.land/r/aib/ibc/core"
	"gno.land/r/aib/ibc/core/testing"
)

// validateError implements ClientState or ConsensusState in order to returns
// an error when Validate() is invoked.
type validateError struct{}

func (validateError) Validate() error {
	return errors.New("validate error")
}

func (validateError) ClientType() string {
	return lightclient.Tendermint
}

// Assert calls to clientState.Validate and consensusState.Validate
// NOTE: We don't want to test Validate() checks in filetest because it has too
// many cases, hence Validate is tested in unittests and we only ensure that
// the Validate() method are properly invoked.
func main() {
	var (
		clientState    = testing.GenClientState("chain-id-2")
		apphash        = sha256.Sum256([]byte("apphash"))
		valset         = testing.GenValset()
		consensusState = testing.GenConsensusState(time.Now(), apphash[:], valset.Hash())
	)

	println("clientState.Validate:",
		// NOTE: panics in explicit cross realm can only be caught by revive.
		revive(func() {
			core.CreateClient(cross, validateError{}, consensusState)
		}))
	println("consensusState.Validate:",
		revive(func() {
			core.CreateClient(cross, clientState, validateError{})
		}))
}

// Output:
// clientState.Validate: validate error
// consensusState.Validate: validate error
