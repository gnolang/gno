package main

import (
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.CreateBoard(cross, "", false)
}

// Error:
// board name is empty
