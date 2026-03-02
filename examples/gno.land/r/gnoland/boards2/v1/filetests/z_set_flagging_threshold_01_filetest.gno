package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address   = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"
	bid   boards.ID = 404
)

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.SetFlaggingThreshold(cross, bid, 1)
}

// Error:
// board does not exist with ID: 404
