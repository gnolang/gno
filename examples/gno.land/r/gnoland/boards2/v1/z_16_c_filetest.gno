package main

import (
	"chain"
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.SetFlaggingThreshold(cross, 1, 0)
}

// Error:
// invalid flagging threshold
