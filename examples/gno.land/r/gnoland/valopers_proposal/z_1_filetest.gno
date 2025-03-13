// PKGPATH: gno.land/r/gnoland/valopers_proposal_test
package valopers_proposal_test

// SEND: 100000000ugnot

import (
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/r/gnoland/valopers"
	"gno.land/r/gnoland/valopers_proposal"
	_ "gno.land/r/gov/dao/init" // so that the govdao initializer is executed
	govdao "gno.land/r/gov/dao/v2"
)

const (
	validMoniker     = "test-1"
	validDescription = "test-1's description"
	validAddress     = std.Address("g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h")
	validPubKey      = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
)

func init() {
	// Register a validator and add the proposal
	valopers.Register(validMoniker, validDescription, validAddress, validPubKey)
	valopers_proposal.ProposeNewValidator(validAddress)
}

func main() {
	println(govdao.Render(""))
}

// Output:
// # GovDAO Proposals
//
// ## [Prop #0 - Add valoper test-1 to the valset](/r/gov/dao/v2:0)
//
// **Status: ACTIVE**
//
// **Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm**
//
//
