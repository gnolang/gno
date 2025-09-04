package main

import (
	"chain"
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var bid boards2.BoardID

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false)
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.DeleteReply(cross, bid, 404, 1)
}

// Error:
// thread does not exist with ID: 404
