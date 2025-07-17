package main

import (
	"gno.land/p/aib/ibc/lightclient"
	"gno.land/r/aib/ibc/core"
)

func main() {
	// CreateClient
	core.CreateClient(cross, lightclient.Solomachine, nil, nil)
}

// Error:
// unhandled light client type 06-solomachine
