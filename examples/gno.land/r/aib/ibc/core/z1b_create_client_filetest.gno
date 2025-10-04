package main

import (
	"gno.land/p/aib/ibc/lightclient"
	"gno.land/r/aib/ibc/core"
)

type SoloMachine struct{}

func (SoloMachine) ClientType() string {
	return lightclient.Solomachine
}

func (SoloMachine) ValidateBasic() error {
	return nil
}

// CreateClient w/ unhandled light client type
func main() {
	core.CreateClient(cross, SoloMachine{}, SoloMachine{})
}

// Error:
// unhandled light client type 06-solomachine
