package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner       = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	fee   int64 = 9999999
)

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	cross(boards2.SetReplyFee)(fee)

	// Ensure current reply fee is right
	println(boards2.GetReplyFee() == fee)
}

// Output:
// true
