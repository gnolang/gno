// PKGPATH: gno.land/r/gnoland/valopers_test
// SEND: 20000000ugnot

package valopers_test

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/gnoland/valopers"
)

var (
	g1user = testutils.TestAddress("g1user")
)

const (
	validMoniker     = "test-1"
	validDescription = "test-1's description"
	validAddress     = std.Address("g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h")
	validPubKey      = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
)

func init() {
	testing.SetOriginCaller(g1user)

	// Register a validator and add the proposal
	cross(valopers.Register)(validMoniker, validDescription, validAddress, validPubKey)
}

func main() {
	// Simulate clicking on the validator
	println(valopers.Render(validAddress.String()))
}

// Output:
// Valoper's details:
// ## test-1
// test-1's description
//
// - Address: g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h
// - PubKey: gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p
//
// [Profile link](/r/demo/profile:u/g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h)
//
