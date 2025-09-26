package main

import (
	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/aib/ibc/lightclient/tendermint/testing"
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

type SoloMachine struct{}

func (SoloMachine) ClientType() string {
	return lightclient.Solomachine
}

func (SoloMachine) Validate() error {
	return nil
}

// CreateClient w/ client type mismatch
func main() {
	core.CreateClient(cross, testing.NewClientState("chainid", types.Height{}), SoloMachine{})
}

// Error:
// client type for client state and consensus state do not match
