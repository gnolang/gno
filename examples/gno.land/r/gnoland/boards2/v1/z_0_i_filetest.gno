package main

import (
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.CreateBoard(cross, "test123", false)
}

// Error:
// unauthorized
