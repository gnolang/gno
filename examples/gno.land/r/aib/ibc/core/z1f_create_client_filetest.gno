package main

import (
	"gno.land/p/aib/ibc/lightclient"
	"gno.land/r/aib/ibc/core"
	"gno.land/r/aib/ibc/core/testing"
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
	core.CreateClient(cross, testing.GenClientState("chainid"), SoloMachine{})
}

// Error:
// client type for client state and consensus state do not match
