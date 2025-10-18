package main

import (
	"crypto/sha256"
	"errors"
	"time"

	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/aib/ibc/lightclient/tendermint/testing"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

// validateError implements ClientState or ConsensusState in order to returns
// an error when ValidateBasic() is invoked.
type validateError struct{}

func (validateError) ValidateBasic() error {
	return errors.New("validate error")
}

func (validateError) ClientType() string {
	return lightclient.Tendermint
}

// Assert calls to clientState.ValidateBasic and consensusState.ValidateBasic
// NOTE: We don't want to test ValidateBasic() checks in filetest because it
// has too many cases, hence ValidateBasic is tested in unittests and we only
// ensure that the ValidateBasic() method are properly invoked.
func main() {
	var (
		chainID        = "chain-id-2"
		clientState    = testing.NewClientState(chainID, types.NewHeight(2, 2))
		apphash        = sha256.Sum256([]byte("apphash"))
		valset         = testing.GenValset()
		consensusState = testing.GenConsensusState(time.Now(), apphash[:], valset.Hash())
	)

	println("clientState.ValidateBasic:",
		// NOTE: panics in explicit cross realm can only be caught by revive.
		revive(func() {
			core.CreateClient(cross, validateError{}, consensusState)
		}))
	println("consensusState.ValidateBasic:",
		revive(func() {
			core.CreateClient(cross, clientState, validateError{})
		}))
}

// Output:
// clientState.ValidateBasic: validate error
// consensusState.ValidateBasic: validate error
