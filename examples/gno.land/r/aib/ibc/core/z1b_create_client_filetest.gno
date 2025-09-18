package main

import (
	"gno.land/p/aib/ibc/lightclient"
	"gno.land/r/aib/ibc/core"
)

// CreateClient w/ unhandled light client type
func main() {
	core.CreateClient(cross, lightclient.Solomachine, nil, nil)
}

// Error:
// unhandled light client type 06-solomachine
