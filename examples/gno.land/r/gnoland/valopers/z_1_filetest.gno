// PKGPATH: gno.land/r/gnoland/valopers_test
// SEND: 20000000ugnot

package valopers_test

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/gnoland/valopers"
)

var g1user = testutils.TestAddress("g1user") // g1vuch2um9wf047h6lta047h6lta047h6l2ewm6w

const (
	validMoniker     = "test-1"
	validDescription = "test-1's description"
	validAddress     = std.Address("g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h")
	validPubKey      = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
)

func init() {
	testing.SetOriginCaller(g1user)

	// Register a validator and add the proposal
	valopers.Register(cross, validMoniker, validDescription, validAddress, validPubKey)
}

func main() {
	println(valopers.Render(""))
}

// Error:
// should not happen, method and object not in same realm
