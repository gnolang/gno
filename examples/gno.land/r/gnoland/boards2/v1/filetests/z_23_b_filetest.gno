package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/test"))

	// Calling as a non Boards2 sub realm should fail
	_ = boards2.GetRealmPermissions()
}

// Error:
// forbidden
