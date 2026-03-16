package main

import (
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"

var name = strings.Repeat("X", boards2.MaxBoardNameLength+1)

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateBoard(cross, name, false, false)
}

// Error:
// board name is too long, maximum allowed is 50 characters
