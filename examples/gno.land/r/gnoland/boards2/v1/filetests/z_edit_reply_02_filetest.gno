package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.EditReply(cross, 404, 1, 0, "body")
}

// Error:
// board does not exist with ID: 404
