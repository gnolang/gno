package main

import (
	"chain"
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	name  = "foo123"
)

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	boards2.CreateBoard(cross, name, false)
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.RenameBoard(cross, name, name)
}

// Error:
// board already exists
