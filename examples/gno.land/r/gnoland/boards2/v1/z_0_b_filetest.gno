package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func init() {
	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.CreateBoard)("")
}

// Error:
// board name is empty
