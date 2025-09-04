package main

import (
	"chain"
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = chain.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.RemoveMember(cross, 0, "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn") // Operate on realm DAO instead of individual boards
}

// Error:
// unauthorized
