// PKGPATH: gno.land/r/gnoland/valopers_proposal_test
package valopers_proposal_test

// SEND: 20000000ugnot

import (
	"std"

	"gno.land/r/gnoland/valopers"
	"gno.land/r/gnoland/valopers_proposal"
	_ "gno.land/r/gov/dao/init" // so that the govdao initializer is executed
)

const (
	validMoniker     = "test-1"
	validDescription = "test-1's description"
	validAddress     = std.Address("g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h")
	otherAddress     = std.Address("g1juz2yxmdsa6audkp6ep9vfv80c8p5u76e03vvh")
	validPubKey      = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
)

func init() {
}

func main() {
	// Register a validator
	valopers.Register(validMoniker, validDescription, validAddress, validPubKey)
	// Try to make a proposal for a non-existing validator
	defer func() {
		r := recover()
		println("r: ", r)
	}()
	valopers_proposal.ProposeNewValidator(otherAddress)
}

// Output:
// r:  valoper does not exist
