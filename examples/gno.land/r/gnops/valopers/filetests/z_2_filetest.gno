// PKGPATH: gno.land/r/gnops/valopers_test
// SEND: 20000000ugnot

package valopers_test

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/gnops/valopers"
)

var g1user = testutils.TestAddress("g1user")

const (
	validMoniker     = "test-1"
	validDescription = "test-1's description"
	validServerType  = valopers.ServerTypeOnPrem
	validAddress     = address("g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h")
	validPubKey      = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
)

func init() {
	// OriginCaller must equal the operator address (post-genesis squat guard).
	// validAddress here was chosen so that derive(validPubKey) == validAddress,
	// so OperatorAddress == SigningAddress in this fixture.
	testing.SetOriginCaller(validAddress)

	// Register a validator and add the proposal
	valopers.Register(cross, validMoniker, validDescription, validServerType, validAddress, validPubKey)
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
// - Operator Address: g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h
// - Signing Address: g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h
// - Signing PubKey: gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p
// - Server Type: on-prem
//
// [Profile link](/r/demo/profile:u/g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h)
//
