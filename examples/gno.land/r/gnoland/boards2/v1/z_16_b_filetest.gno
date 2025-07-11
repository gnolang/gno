package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner                 = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	bid   boards2.BoardID = 404
)

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	boards2.SetFlaggingThreshold(cross, bid, 1)
}

// Error:
// board does not exist with ID: 404
